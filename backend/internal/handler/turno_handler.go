package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"kora/backend/internal/model"
	"kora/backend/internal/service"
)

// ListarServiciosHandler maneja GET /api/servicios
func ListarServiciosHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	servicios, err := service.ListarServicios(r.Context())
	if err != nil {
		ErrorInterno(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, servicios)
}

// CrearTurnoHandler maneja POST /api/turnos
func CrearTurnoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	var req model.CrearTurnoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	ticket, err := service.CrearTurno(r.Context(), &req)
	if err != nil {
		esValidacion := errors.Is(err, service.ErrContactoRequerido) ||
			errors.Is(err, service.ErrNombreRequerido) ||
			errors.Is(err, service.ErrServicioRequerido) ||
			errors.Is(err, service.ErrServicioInvalido) ||
			errors.Is(err, service.ErrTipoAtencionInvalido) ||
			errors.Is(err, service.ErrFechaHoraInvalida)

		switch {
		case esValidacion:
			ErrorResponse(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrYaTieneTurnoActivo):
			ErrorResponse(w, http.StatusConflict, err.Error())
		default:
			ErrorInterno(w, err)
		}
		return
	}

	SuccessResponse(w, http.StatusCreated, ticket)
}

// ConsultarTurnoHandler maneja POST /api/turnos/consultar
// Acepta el número de turno (CAR-014) o el número de contacto.
func ConsultarTurnoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	// Se acepta {query} (número de turno o contacto) y {contacto}
	var req struct {
		Query    string `json:"query"`
		Contacto string `json:"contacto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	consulta := req.Query
	if consulta == "" {
		consulta = req.Contacto
	}

	resultado, err := service.ConsultarTurno(r.Context(), consulta)
	if err != nil {
		if errors.Is(err, service.ErrConsultaRequerida) {
			ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		ErrorInterno(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, resultado)
}

// ObtenerTurnoHandler maneja GET /api/turnos/{numero}
func ObtenerTurnoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	numero := strings.TrimPrefix(r.URL.Path, "/api/turnos/")
	numero = strings.Trim(numero, "/")
	if numero == "" {
		ErrorResponse(w, http.StatusBadRequest, "número de turno requerido")
		return
	}

	ticket, err := service.ObtenerTicketPorNumero(r.Context(), numero)
	if err != nil {
		if errors.Is(err, service.ErrTurnoNoEncontrado) {
			ErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		ErrorInterno(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, ticket)
}
