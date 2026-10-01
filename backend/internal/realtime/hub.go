// Package realtime implementa un bus de eventos en memoria para
// transmitir los cambios de turnos a los paneles conectados por SSE.
package realtime

import (
	"sync"
)

// Tipos de evento publicados por el sistema
const (
	EventoSnapshot         = "snapshot"
	EventoTurnoCreado      = "turno_creado"
	EventoTurnoLlamado     = "turno_llamado"
	EventoTurnoAtendido    = "turno_atendido"
	EventoTurnoCancelado   = "turno_cancelado"
	EventoServicioCambiado = "servicio_cambiado"
	EventoTurnoVencido     = "turno_vencido"
)

// Evento representa un mensaje para los suscriptores
type Evento struct {
	Tipo  string      `json:"tipo"`
	Datos interface{} `json:"datos,omitempty"`
}

// Hub mantiene el registro de suscriptores conectados
type Hub struct {
	mu   sync.RWMutex
	subs map[chan Evento]struct{}
}

var defaultHub = NewHub()

// Default devuelve el hub global de la aplicación
func Default() *Hub {
	return defaultHub
}

// NewHub crea un hub vacío
func NewHub() *Hub {
	return &Hub{subs: make(map[chan Evento]struct{})}
}

// Suscribir registra un nuevo suscriptor y devuelve el canal de eventos
// junto con la función para cerrar la suscripción.
func (h *Hub) Suscribir() (<-chan Evento, func()) {
	canal := make(chan Evento, 16)

	h.mu.Lock()
	h.subs[canal] = struct{}{}
	h.mu.Unlock()

	var cerrar sync.Once
	cancelar := func() {
		cerrar.Do(func() {
			h.mu.Lock()
			delete(h.subs, canal)
			h.mu.Unlock()
			close(canal)
		})
	}

	return canal, cancelar
}

// Publicar envía el evento a todos los suscriptores activos.
// Si un suscriptor está saturado se descarta el evento (slow client).
func (h *Hub) Publicar(tipo string, datos interface{}) {
	evento := Evento{Tipo: tipo, Datos: datos}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for canal := range h.subs {
		select {
		case canal <- evento:
		default:
		}
	}
}

// Cantidad de suscriptores conectados (diagnóstico)
func (h *Hub) CantidadSuscriptores() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs)
}
