package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"kora/backend/internal/model"
	"kora/backend/internal/repository"
)

// almacén de tokens en memoria (en producción usar Redis o BD)
var (
	tokensMu sync.RWMutex
	tokens   = make(map[string]tokenInfo)
)

type tokenInfo struct {
	UsuarioID int
	ExpiraEn  time.Time
}

// Login autentica al personal administrativo
func Login(ctx context.Context, req *model.LoginRequest) (string, error) {
	usuario, err := repository.ObtenerUsuarioPorEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUsuarioNoEncontrado) {
			return "", ErrCredencialesInvalidas
		}
		return "", err
	}

	// Verificar que sea admin
	if usuario.Rol != "ADMIN" {
		return "", ErrCredencialesInvalidas
	}

	// Verificar contraseña con bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(usuario.PasswordHash), []byte(req.Password)); err != nil {
		return "", ErrCredencialesInvalidas
	}

	token := uuid.New().String()

	// Obtener duración del token
	expiryHours := 8
	if v := os.Getenv("TOKEN_EXPIRY_HOURS"); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			expiryHours = h
		}
	}

	tokensMu.Lock()
	tokens[token] = tokenInfo{
		UsuarioID: usuario.ID,
		ExpiraEn:  time.Now().Add(time.Duration(expiryHours) * time.Hour),
	}
	tokensMu.Unlock()

	return token, nil
}

// ValidarToken verifica si un token es válido y no ha expirado
func ValidarToken(token string) (int, bool) {
	tokensMu.RLock()
	info, existe := tokens[token]
	tokensMu.RUnlock()

	if !existe {
		return 0, false
	}

	if time.Now().After(info.ExpiraEn) {
		tokensMu.Lock()
		delete(tokens, token)
		tokensMu.Unlock()
		return 0, false
	}

	return info.UsuarioID, true
}

// ListarTurnos lista turnos filtrados por estado, servicio o día
func ListarTurnos(ctx context.Context, filtro model.FiltroTurnos) ([]model.Turno, error) {
	if filtro.Estado != "" && !esEstadoValido(filtro.Estado) {
		return nil, ErrEstadoInvalido
	}

	return repository.ListarTurnos(ctx, filtro)
}

// ObtenerTurnosPendientes obtiene todos los turnos pendientes
func ObtenerTurnosPendientes(ctx context.Context) ([]model.Turno, error) {
	return repository.ListarTurnos(ctx, model.FiltroTurnos{Estado: model.EstadoPendiente})
}

// CambiarEstadoTurno cambia el estado de un turno con validación de transiciones
func CambiarEstadoTurno(ctx context.Context, id int, nuevoEstado string) (*model.Turno, error) {
	if !esEstadoValido(nuevoEstado) {
		return nil, ErrEstadoInvalido
	}

	turno, err := repository.ObtenerTurnoPorID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !esTransicionValida(turno.Estado, nuevoEstado) {
		return nil, fmt.Errorf("%w: no se puede cambiar de %s a %s", ErrTransicionInvalida, turno.Estado, nuevoEstado)
	}

	actualizado, err := repository.ActualizarEstadoTurno(ctx, id, nuevoEstado)
	if err != nil {
		return nil, err
	}

	PublicarTablero(ctx, eventoSegunEstado(nuevoEstado), actualizado)

	return actualizado, nil
}

// LlamarTurno pasa un turno de PENDIENTE a EN_ATENCION
func LlamarTurno(ctx context.Context, id int) (*model.Turno, error) {
	turno, err := repository.ObtenerTurnoPorID(ctx, id)
	if err != nil {
		return nil, err
	}

	if turno.Estado == model.EstadoEnAtencion {
		return turno, nil // idempotente
	}

	return CambiarEstadoTurno(ctx, id, model.EstadoEnAtencion)
}

// AtenderTurno marca un turno como ATENDIDO
func AtenderTurno(ctx context.Context, id int) (*model.Turno, error) {
	turno, err := repository.ObtenerTurnoPorID(ctx, id)
	if err != nil {
		return nil, err
	}

	if turno.Estado == model.EstadoAtendido {
		return turno, nil // idempotente
	}

	return CambiarEstadoTurno(ctx, id, model.EstadoAtendido)
}

// CancelarTurno cancela un turno con un motivo opcional
func CancelarTurno(ctx context.Context, id int, motivo string) (*model.Turno, error) {
	if len(motivo) > 255 {
		return nil, ErrMotivoInvalido
	}

	turno, err := repository.ObtenerTurnoPorID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Solo se pueden cancelar turnos PENDIENTE o EN_ATENCION
	if turno.Estado != model.EstadoPendiente && turno.Estado != model.EstadoEnAtencion {
		return nil, fmt.Errorf("%w: no se puede cancelar un turno en estado %s", ErrTransicionInvalida, turno.Estado)
	}

	cancelado, err := repository.CancelarTurno(ctx, id, motivo)
	if err != nil {
		return nil, err
	}

	PublicarTablero(ctx, eventoSegunEstado(model.EstadoCancelado), cancelado)

	return cancelado, nil
}

// CambiarServicioTurno re-encola un turno PENDIENTE en otro servicio
func CambiarServicioTurno(ctx context.Context, id int, req *model.CambiarServicioRequest) (*model.Turno, error) {
	if req.ServicioID <= 0 {
		return nil, ErrServicioRequerido
	}
	if len(req.Motivo) > 255 {
		return nil, ErrMotivoInvalido
	}

	servicio, err := repository.ObtenerServicio(ctx, req.ServicioID)
	if err != nil || !servicio.Activo {
		return nil, ErrServicioInvalido
	}

	turno, err := repository.ObtenerTurnoPorID(ctx, id)
	if err != nil {
		return nil, err
	}

	if turno.Estado != model.EstadoPendiente {
		return nil, fmt.Errorf("%w: solo se pueden re-encolar turnos PENDIENTE (estado actual %s)", ErrTransicionInvalida, turno.Estado)
	}

	var tipoAtencionID *int
	if req.TipoAtencionID > 0 {
		tipo, err := repository.ObtenerTipoAtencion(ctx, req.TipoAtencionID)
		if err != nil || tipo.ServicioID != servicio.ID {
			return nil, ErrTipoAtencionInvalido
		}
		tipoAtencionID = &tipo.ID
	}

	// Validar que realmente cambie de cola
	mismoServicio := turno.ServicioID != nil && *turno.ServicioID == servicio.ID
	mismoTipo := (turno.TipoAtencionID == nil && tipoAtencionID == nil) ||
		(turno.TipoAtencionID != nil && tipoAtencionID != nil && *turno.TipoAtencionID == *tipoAtencionID)
	if mismoServicio && mismoTipo {
		return nil, ErrServicioDestinoIgual
	}

	// Si ya tiene turno activo en el servicio destino, no se permite
	otro, err := repository.ObtenerTurnoActivoPorContactoServicio(ctx, turno.ClienteContacto, servicio.ID)
	if err != nil {
		return nil, err
	}
	if otro != nil && otro.ID != turno.ID {
		return nil, fmt.Errorf("%w (turno %s en %s)", ErrYaTieneTurnoActivo, otro.NumeroTurno, servicio.Nombre)
	}

	actualizado, err := repository.ReencolarTurno(ctx, id, servicio.ID, tipoAtencionID, servicio.Codigo)
	if err != nil {
		return nil, err
	}

	PublicarTablero(ctx, "servicio_cambiado", actualizado)

	return actualizado, nil
}

// MarcarTurnosVencidos ejecuta la limpieza de turnos vencidos
func MarcarTurnosVencidos(ctx context.Context) (int64, error) {
	return repository.MarcarTurnosVencidos(ctx)
}

// esEstadoValido verifica si un estado es válido
func esEstadoValido(estado string) bool {
	switch estado {
	case model.EstadoPendiente, model.EstadoEnAtencion, model.EstadoAtendido,
		model.EstadoCancelado, model.EstadoVencido:
		return true
	}
	return false
}

// esTransicionValida verifica si una transición de estado es permitida
func esTransicionValida(actual, nuevo string) bool {
	switch actual {
	case model.EstadoPendiente:
		// De PENDIENTE puede ir a EN_ATENCION, CANCELADO o VENCIDO
		return nuevo == model.EstadoEnAtencion || nuevo == model.EstadoCancelado || nuevo == model.EstadoVencido
	case model.EstadoEnAtencion:
		// De EN_ATENCION puede ir a ATENDIDO o CANCELADO
		return nuevo == model.EstadoAtendido || nuevo == model.EstadoCancelado
	case model.EstadoAtendido, model.EstadoCancelado, model.EstadoVencido:
		// Estados finales, no permiten transiciones
		return false
	}
	return false
}

// eventoSegunEstado mapea un estado al evento de tiempo real correspondiente
func eventoSegunEstado(estado string) string {
	switch estado {
	case model.EstadoEnAtencion:
		return "turno_llamado"
	case model.EstadoAtendido:
		return "turno_atendido"
	case model.EstadoCancelado, model.EstadoVencido:
		return "turno_cancelado"
	default:
		return "turno_actualizado"
	}
}
