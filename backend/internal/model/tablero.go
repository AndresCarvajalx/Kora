package model

import "time"

// Tablero es el estado en vivo que muestran los paneles de llamados
type Tablero struct {
	ActualizadoEn  time.Time `json:"actualizado_en"`
	Pendientes     int       `json:"pendientes"`
	EnAtencion     []*Turno  `json:"en_atencion"`
	Historial      []*Turno  `json:"historial"`
	UltimoServicio string    `json:"ultimo_servicio,omitempty"`
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
