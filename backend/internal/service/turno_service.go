package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5/pgconn"

	"kora/backend/internal/model"
	"kora/backend/internal/realtime"
	"kora/backend/internal/repository"
)

var (
	ErrContactoRequerido     = errors.New("el número de contacto es requerido")
	ErrNombreRequerido       = errors.New("el nombre es requerido")
	ErrServicioRequerido     = errors.New("debe seleccionar un servicio médico")
	ErrServicioInvalido      = errors.New("el servicio médico no existe o está inactivo")
	ErrTipoAtencionInvalido  = errors.New("el tipo de atención no corresponde al servicio seleccionado")
	ErrConsultaRequerida     = errors.New("ingrese su número de turno o su número de contacto")
	ErrTurnoNoEncontrado     = errors.New("no se encontró ningún turno con esos datos")
	ErrFechaHoraInvalida     = errors.New("formato de fecha_hora inválido, use RFC3339")
	ErrYaTieneTurnoActivo    = errors.New("ya tiene un turno activo en este servicio")
	ErrServicioDestinoIgual  = errors.New("el turno ya pertenece a ese servicio y tipo de atención")
	ErrEstadoInvalido        = errors.New("estado inválido")
	ErrTransicionInvalida    = errors.New("transición de estado no permitida")
	ErrCredencialesInvalidas = errors.New("credenciales inválidas")
	ErrMotivoInvalido        = errors.New("el motivo no puede superar los 255 caracteres")
)

// reCodigoTurno detecta si el texto buscado parece un número de turno (CAR-014).
// Exige letras iniciales para no confundirlo con un número de contacto.
var reCodigoTurno = regexp.MustCompile(`^[A-Z]{2,4}-?\d{1,5}$`)

// ResultadoConsulta es la respuesta normalizada de la consulta de turno
type ResultadoConsulta struct {
	TurnoActivo bool                `json:"turno_activo"`
	Criterio    string              `json:"criterio,omitempty"`
	Mensaje     string              `json:"mensaje,omitempty"`
	Turno       *model.TicketTurno  `json:"turno,omitempty"`
	Turnos      []model.TicketTurno `json:"turnos,omitempty"`
}

// ListarServicios devuelve el catálogo de servicios con sus tipos de atención
func ListarServicios(ctx context.Context) ([]model.Servicio, error) {
	return repository.ListarServicios(ctx, true)
}

// CrearTurno crea un nuevo turno para un cliente dentro del servicio elegido
func CrearTurno(ctx context.Context, req *model.CrearTurnoRequest) (*model.TicketTurno, error) {
	contacto := strings.TrimSpace(req.Contacto)
	nombre := strings.TrimSpace(req.Nombre)
	documento := strings.TrimSpace(req.Documento)

	if contacto == "" {
		return nil, ErrContactoRequerido
	}
	if len(contacto) < 4 {
		return nil, fmt.Errorf("%w (mínimo 4 caracteres)", ErrContactoRequerido)
	}
	if nombre == "" {
		return nil, ErrNombreRequerido
	}
	if req.ServicioID <= 0 {
		return nil, ErrServicioRequerido
	}

	servicio, err := repository.ObtenerServicio(ctx, req.ServicioID)
	if err != nil || !servicio.Activo {
		return nil, ErrServicioInvalido
	}

	var tipoAtencionID *int
	if req.TipoAtencionID > 0 {
		tipo, err := repository.ObtenerTipoAtencion(ctx, req.TipoAtencionID)
		if err != nil || tipo.ServicioID != servicio.ID {
			return nil, ErrTipoAtencionInvalido
		}
		tipoAtencionID = &tipo.ID
	}

	fechaHora := time.Now()
	if req.FechaHora != "" {
		fechaHora, err = time.Parse(time.RFC3339, req.FechaHora)
		if err != nil {
			return nil, ErrFechaHoraInvalida
		}
	}

	// Si ya tiene un turno activo en este servicio, se libera si ya venció
	turnoExistente, err := repository.ObtenerTurnoActivoPorContactoServicio(ctx, contacto, servicio.ID)
	if err != nil {
		return nil, err
	}
	if turnoExistente != nil {
		if turnoExistente.Estado == model.EstadoPendiente && time.Now().After(turnoExistente.ExpiraEn) {
			if _, err := repository.ActualizarEstadoTurno(ctx, turnoExistente.ID, model.EstadoVencido); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("%w (turno %s)", ErrYaTieneTurnoActivo, turnoExistente.NumeroTurno)
		}
	}

	turno, err := repository.CrearTurno(ctx, repository.NuevoTurno{
		Contacto:       contacto,
		Nombre:         nombre,
		Documento:      documento,
		ServicioID:     servicio.ID,
		ServicioCodigo: servicio.Codigo,
		TipoAtencionID: tipoAtencionID,
		FechaHora:      fechaHora,
		ExpiraEn:       fechaHora.Add(expiracionTurno()),
		ColaFecha:      fechaHora,
	})
	if err != nil {
		// Índice único de turno activo violated (carrera entre peticiones)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrYaTieneTurnoActivo
		}
		return nil, err
	}

	ticket, err := construirTicket(ctx, turno)
	if err != nil {
		return nil, err
	}

	realtime.Default().Publicar(realtime.EventoTurnoCreado, ticket)

	return ticket, nil
}

// ConsultarTurno busca un turno por número de turno o por contacto.
// Si el texto parece un número (CAR-014) busca por código, si no por contacto.
func ConsultarTurno(ctx context.Context, consulta string) (*ResultadoConsulta, error) {
	texto := strings.TrimSpace(consulta)
	if texto == "" {
		return nil, ErrConsultaRequerida
	}

	// Se ignoran mayúsculas y espacios para decidir si es código o contacto
	compacto := strings.ToUpper(strings.ReplaceAll(texto, " ", ""))

	if reCodigoTurno.MatchString(compacto) {
		return consultarPorNumero(ctx, texto)
	}

	return consultarPorContacto(ctx, texto)
}

// ObtenerTicketPorNumero devuelve el ticket asociado a un número de turno.
// Acepta las mismas variantes que la consulta: "CAR-014", "car14", "CAR-14".
func ObtenerTicketPorNumero(ctx context.Context, numero string) (*model.TicketTurno, error) {
	for _, candidato := range candidatosNumeroTurno(numero) {
		turno, err := repository.ObtenerTurnoPorNumero(ctx, candidato)
		if err != nil {
			if errors.Is(err, repository.ErrTurnoNoEncontrado) {
				continue
			}
			return nil, err
		}

		return construirTicket(ctx, turno)
	}

	return nil, ErrTurnoNoEncontrado
}

// consultarPorNumero busca por código de turno
func consultarPorNumero(ctx context.Context, texto string) (*ResultadoConsulta, error) {
	for _, candidato := range candidatosNumeroTurno(texto) {
		turno, err := repository.ObtenerTurnoPorNumero(ctx, candidato)
		if err != nil {
			if errors.Is(err, repository.ErrTurnoNoEncontrado) {
				continue
			}
			return nil, err
		}

		if turno.Estado == model.EstadoPendiente && time.Now().After(turno.ExpiraEn) {
			vencido, err := repository.ActualizarEstadoTurno(ctx, turno.ID, model.EstadoVencido)
			if err != nil {
				return nil, err
			}
			turno = vencido
		}

		ticket, err := construirTicket(ctx, turno)
		if err != nil {
			return nil, err
		}

		return &ResultadoConsulta{
			TurnoActivo: esEstadoActivo(turno.Estado),
			Criterio:    "numero",
			Turno:       ticket,
		}, nil
	}

	return &ResultadoConsulta{
		TurnoActivo: false,
		Criterio:    "numero",
		Mensaje:     "no encontramos un turno con ese número",
	}, nil
}

// candidatosNumeroTurno genera las variantes de búsqueda a partir de lo que
// escribió el usuario. Tolera minúsculas, espacios y falta del guion o de los
// ceros: "car 14", "CAR14" y "CAR-14" terminan encontrando "CAR-014".
func candidatosNumeroTurno(texto string) []string {
	base := strings.ToUpper(strings.ReplaceAll(texto, " ", ""))
	candidatos := []string{base}

	pos := strings.IndexFunc(base, unicode.IsDigit)
	if pos <= 0 {
		return candidatos // no tiene forma de código de servicio
	}

	prefijo := strings.TrimRight(base[:pos], "-")
	cifras := strings.TrimLeft(base[pos:], "-")
	if cifras == "" || prefijo == "" {
		return candidatos
	}

	// Los números se emiten con al menos tres dígitos
	for _, largo := range []int{3, 4} {
		if len(cifras) > largo {
			continue
		}
		relleno := strings.Repeat("0", largo-len(cifras))
		candidatos = append(candidatos, prefijo+"-"+relleno+cifras)
	}

	return candidatos
}

// consultarPorContacto busca todos los turnos activos de un contacto
func consultarPorContacto(ctx context.Context, contacto string) (*ResultadoConsulta, error) {
	turnos, err := repository.ObtenerTurnosActivosPorContacto(ctx, contacto)
	if err != nil {
		return nil, err
	}

	tickets := make([]model.TicketTurno, 0, len(turnos))
	for i := range turnos {
		if turnos[i].Estado == model.EstadoPendiente && time.Now().After(turnos[i].ExpiraEn) {
			vencido, err := repository.ActualizarEstadoTurno(ctx, turnos[i].ID, model.EstadoVencido)
			if err != nil {
				return nil, err
			}
			turnos[i] = *vencido
		}
		if !esEstadoActivo(turnos[i].Estado) {
			continue
		}

		ticket, err := construirTicket(ctx, &turnos[i])
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, *ticket)
	}

	if len(tickets) == 0 {
		return &ResultadoConsulta{
			TurnoActivo: false,
			Criterio:    "contacto",
			Mensaje:     "no tiene un turno activo",
		}, nil
	}

	// Si tiene varios servicios, el más reciente es el principal
	resultado := &ResultadoConsulta{
		TurnoActivo: true,
		Criterio:    "contacto",
		Turno:       &tickets[0],
	}
	if len(tickets) > 1 {
		resultado.Turnos = tickets
	}

	return resultado, nil
}

// construirTicket completa el turno con los datos de cola
func construirTicket(ctx context.Context, turno *model.Turno) (*model.TicketTurno, error) {
	personasAntes, total, err := repository.ContarCola(ctx, turno)
	if err != nil {
		return nil, err
	}

	ticket := &model.TicketTurno{Turno: *turno, PersonasAntes: personasAntes, TotalEnCola: total}

	if turno.Estado == model.EstadoPendiente {
		duracion := 15
		if turno.ServicioID != nil {
			if minutos, err := repository.DuracionServicio(ctx, *turno.ServicioID); err == nil {
				duracion = minutos
			}
		}
		ticket.Posicion = personasAntes + 1
		ticket.TiempoEstimadoMin = personasAntes * duracion
	}

	return ticket, nil
}

// esEstadoActivo indica si el turno sigue vigente
func esEstadoActivo(estado string) bool {
	return estado == model.EstadoPendiente || estado == model.EstadoEnAtencion
}

// expiracionTurno devuelve las horas de vigencia configuradas para un turno
func expiracionTurno() time.Duration {
	horas := 5
	if v := os.Getenv("TURNO_EXPIRY_HOURS"); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			horas = h
		}
	}
	return time.Duration(horas) * time.Hour
}
