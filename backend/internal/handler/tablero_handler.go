package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"kora/backend/internal/realtime"
	"kora/backend/internal/service"
)

// TableroHandler maneja GET /api/tablero
// Devuelve el estado actual de llamados. Es el respaldo del flujo SSE.
func TableroHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorResponse(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	tablero, err := service.ObtenerTablero(r.Context())
	if err != nil {
		ErrorInterno(w, err)
		return
	}

	SuccessResponse(w, http.StatusOK, tablero)
}

// EventosHandler maneja GET /api/eventos
// Flujo Server-Sent Events con el estado de los llamados en tiempo real.
func EventosHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		ErrorResponse(w, http.StatusInternalServerError, "streaming no soportado")
		return
	}

	cabeceras := w.Header()
	cabeceras.Set("Content-Type", "text/event-stream")
	cabeceras.Set("Cache-Control", "no-cache")
	cabeceras.Set("Connection", "keep-alive")
	cabeceras.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	enviar := func(tipo string, datos interface{}) bool {
		payload, err := json.Marshal(datos)
		if err != nil {
			return true
		}
		_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", tipo, payload)
		if err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	// Reintento rápido en el cliente y primer snapshot con el estado actual
	if _, err := fmt.Fprint(w, "retry: 3000\n\n"); err != nil {
		return
	}
	flusher.Flush()

	if tablero, err := service.ObtenerTablero(r.Context()); err == nil {
		if !enviar(realtime.EventoSnapshot, tablero) {
			return
		}
	}

	canal, cancelar := realtime.Default().Suscribir()
	defer cancelar()

	// Ping periódico para mantener viva la conexión
	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-ping.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()

		case evento, ok := <-canal:
			if !ok {
				return
			}
			if !enviar(evento.Tipo, evento.Datos) {
				return
			}
		}
	}
}
