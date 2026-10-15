package model

import "time"

// TurnoTablero representa la vista pública de un turno en el tablero de llamados
// y flujo SSE, excluyendo datos personales sensibles como contacto/teléfono y documento (Ley 1581 de 2012).
type TurnoTablero struct {
	ID                 int        `json:"id,omitempty"`
	NumeroTurno        string     `json:"numero_turno"`
	ServicioCodigo     string     `json:"servicio_codigo,omitempty"`
	ServicioNombre     string     `json:"servicio_nombre,omitempty"`
	TipoAtencionNombre string     `json:"tipo_atencion_nombre,omitempty"`
	ClienteNombre      string     `json:"cliente_nombre,omitempty"`
	Estado             string     `json:"estado"`
	LlamadoEn          *time.Time `json:"llamado_en,omitempty"`
	AtendidoEn         *time.Time `json:"atendido_en,omitempty"`
}

// Tablero es el estado en vivo que muestran los paneles de llamados
type Tablero struct {
	ActualizadoEn  time.Time       `json:"actualizado_en"`
	Pendientes     int             `json:"pendientes"`
	EnAtencion     []*TurnoTablero `json:"en_atencion"`
	Historial      []*TurnoTablero `json:"historial"`
	UltimoServicio string          `json:"ultimo_servicio,omitempty"`
}

// ATurnoTablero convierte un Turno en la representación pública de pantalla de llamados
func (t *Turno) ATurnoTablero() *TurnoTablero {
	if t == nil {
		return nil
	}
	return &TurnoTablero{
		ID:                 t.ID,
		NumeroTurno:        t.NumeroTurno,
		ServicioCodigo:     t.ServicioCodigo,
		ServicioNombre:     t.ServicioNombre,
		TipoAtencionNombre: t.TipoAtencionNombre,
		ClienteNombre:      t.ClienteNombre,
		Estado:             t.Estado,
		LlamadoEn:          t.LlamadoEn,
		AtendidoEn:         t.AtendidoEn,
	}
}

// FiltroTurnos agrupa los criterios opcionales del listado de turnos
type FiltroTurnos struct {
	Estado     string
	ServicioID int
	Contacto   string
	Numero     string
	SoloHoy    bool
	Limite     int
}
