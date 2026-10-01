package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"kora/backend/internal/model"
	"kora/backend/internal/repository"
	"kora/backend/internal/service"
)

var errTurnoNoEncontrado = repository.ErrTurnoNoEncontrado

// prefijoAdminTurnos es la raíz de las rutas de administración de turnos
const prefijoAdminTurnos = "/api/admin/turnos/"

// RouterAdminTurnos enruta las subrutas de turnos del admin
func RouterAdminTurnos(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	sufijos := []struct {
		sufijo  string
		handler http.HandlerFunc
	}{
		{"/llamar", LlamarTurnoHandler},
		{"/atender", AtenderTurnoHandler},
		{"/estado", CambiarEstadoTurnoHandler},
		{"/cancelar", CancelarTurnoHandler},
		{"/servicio", CambiarServicioTurnoHandler},
	}

	for _, ruta := range sufijos {
		if len(path) > len(prefijoAdminTurnos) && strings.HasSuffix(path, ruta.sufijo) {
			ruta.handler(w, r)
			return
		}
	}

	if path == "/api/admin/turnos" {
		ListarTurnosHandler(w, r)
		return
	}

	ErrorResponse(w, http.StatusNotFound, "ruta no encontrada")
}

// AuthMiddleware valida el token de autenticación para endpoints de admin
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			ErrorResponse(w, http.StatusUnauthorized, "token de autorización requerido")
			return
		}

		// Extraer token de "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			ErrorResponse(w, http.StatusUnauthorized, "formato de autorización inválido")
			return
		}

		if _, valido := service.ValidarToken(parts[1]); !valido {
			ErrorResponse(w, http.StatusUnauthorized, "token inválido o expirado")
			return
		}

		next(w, r)
	}
}

// LoginHandler maneja POST /api/admin/login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	token, err := service.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrCredencialesInvalidas) {
			ErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}
		ErrorInterno(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, model.LoginResponse{Token: token})
}

// ListarTurnosHandler maneja GET /api/admin/turnos?estado=&servicio_id=&hoy=
func ListarTurnosHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	query := r.URL.Query()
	servicioID, _ := strconv.Atoi(query.Get("servicio_id"))
	soloHoy, _ := strconv.ParseBool(query.Get("hoy"))
	limite, _ := strconv.Atoi(query.Get("limite"))

	turnos, err := service.ListarTurnos(r.Context(), model.FiltroTurnos{
		Estado:     strings.TrimSpace(query.Get("estado")),
		ServicioID: servicioID,
		Contacto:   strings.TrimSpace(query.Get("contacto")),
		Numero:     strings.TrimSpace(query.Get("numero")),
		SoloHoy:    soloHoy,
		Limite:     limite,
	})
	if err != nil {
		if errors.Is(err, service.ErrEstadoInvalido) {
			ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		ErrorInterno(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, turnos)
}

// ObtenerTurnosPendientesHandler maneja GET /api/admin/turnos/pendientes
func ObtenerTurnosPendientesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	turnos, err := service.ObtenerTurnosPendientes(r.Context())
	if err != nil {
		ErrorInterno(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, turnos)
}

// CambiarEstadoTurnoHandler maneja PATCH /api/admin/turnos/:id/estado
func CambiarEstadoTurnoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id, err := idDesdeRuta(r.URL.Path, "/estado")
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var req model.CambiarEstadoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	turno, err := service.CambiarEstadoTurno(r.Context(), id, req.Estado)
	if err != nil {
		esValidacion := errors.Is(err, service.ErrEstadoInvalido) ||
			errors.Is(err, service.ErrTransicionInvalida) ||
			errors.Is(err, service.ErrMotivoInvalido)
		switch {
		case esValidacion:
			ErrorResponse(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, errTurnoNoEncontrado):
			ErrorResponse(w, http.StatusNotFound, "turno no encontrado")
		default:
			ErrorInterno(w, err)
		}
		return
	}

	SuccessResponse(w, http.StatusOK, turno)
}

// LlamarTurnoHandler maneja POST /api/admin/turnos/:id/llamar
func LlamarTurnoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id, err := idDesdeRuta(r.URL.Path, "/llamar")
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	turno, err := service.LlamarTurno(r.Context(), id)
	if err != nil {
		esValidacion := errors.Is(err, service.ErrTransicionInvalida) ||
			errors.Is(err, service.ErrEstadoInvalido)
		switch {
		case esValidacion:
			ErrorResponse(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, errTurnoNoEncontrado):
			ErrorResponse(w, http.StatusNotFound, "turno no encontrado")
		default:
			ErrorInterno(w, err)
		}
		return
	}

	SuccessResponse(w, http.StatusOK, turno)
}

// AtenderTurnoHandler maneja POST /api/admin/turnos/:id/atender
func AtenderTurnoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id, err := idDesdeRuta(r.URL.Path, "/atender")
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	turno, err := service.AtenderTurno(r.Context(), id)
	if err != nil {
		esValidacion := errors.Is(err, service.ErrTransicionInvalida) ||
			errors.Is(err, service.ErrEstadoInvalido)
		switch {
		case esValidacion:
			ErrorResponse(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, errTurnoNoEncontrado):
			ErrorResponse(w, http.StatusNotFound, "turno no encontrado")
		default:
			ErrorInterno(w, err)
		}
		return
	}

	SuccessResponse(w, http.StatusOK, turno)
}

// CancelarTurnoHandler maneja POST /api/admin/turnos/:id/cancelar
func CancelarTurnoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id, err := idDesdeRuta(r.URL.Path, "/cancelar")
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var req model.CancelarTurnoRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req) // el cuerpo es opcional
	}

	turno, err := service.CancelarTurno(r.Context(), id, req.Motivo)
	if err != nil {
		esValidacion := errors.Is(err, service.ErrTransicionInvalida) ||
			errors.Is(err, service.ErrMotivoInvalido)
		switch {
		case esValidacion:
			ErrorResponse(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, errTurnoNoEncontrado):
			ErrorResponse(w, http.StatusNotFound, "turno no encontrado")
		default:
			ErrorInterno(w, err)
		}
		return
	}

	SuccessResponse(w, http.StatusOK, turno)
}

// CambiarServicioTurnoHandler maneja PATCH /api/admin/turnos/:id/servicio
func CambiarServicioTurnoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id, err := idDesdeRuta(r.URL.Path, "/servicio")
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var req model.CambiarServicioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	turno, err := service.CambiarServicioTurno(r.Context(), id, &req)
	if err != nil {
		esValidacion := errors.Is(err, service.ErrServicioRequerido) ||
			errors.Is(err, service.ErrServicioInvalido) ||
			errors.Is(err, service.ErrTipoAtencionInvalido) ||
			errors.Is(err, service.ErrServicioDestinoIgual) ||
			errors.Is(err, service.ErrMotivoInvalido) ||
			errors.Is(err, service.ErrTransicionInvalida)
		switch {
		case esValidacion:
			ErrorResponse(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrYaTieneTurnoActivo):
			ErrorResponse(w, http.StatusConflict, err.Error())
		case errors.Is(err, errTurnoNoEncontrado):
			ErrorResponse(w, http.StatusNotFound, "turno no encontrado")
		default:
			ErrorInterno(w, err)
		}
		return
	}

	SuccessResponse(w, http.StatusOK, turno)
}

// MarcarTurnosVencidosHandler maneja POST /api/admin/turnos/vencer
func MarcarTurnosVencidosHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	count, err := service.MarcarTurnosVencidos(r.Context())
	if err != nil {
		ErrorInterno(w, err)
		return
	}

	if count > 0 {
		service.PublicarSnapshot(r.Context())
	}

	SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"turnos_vencidos": count,
	})
}

// idDesdeRuta extrae el ID del turno de la ruta /api/admin/turnos/:id/{sufijo}
func idDesdeRuta(path, sufijo string) (int, error) {
	idStr := strings.TrimPrefix(path, prefijoAdminTurnos)
	idStr = strings.TrimSuffix(idStr, sufijo)

	id, err := strconv.Atoi(strings.Trim(idStr, "/"))
	if err != nil || id <= 0 {
		return 0, errors.New("ID de turno inválido")
	}

	return id, nil
}
