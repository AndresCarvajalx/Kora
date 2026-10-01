package model

// Servicio representa un servicio médico ofrecido por el centro.
// El código es el prefijo del número de turno (ej. CAR-014).
type Servicio struct {
	ID              int            `json:"id" db:"id"`
	Codigo          string         `json:"codigo" db:"codigo"`
	Nombre          string         `json:"nombre" db:"nombre"`
	Descripcion     string         `json:"descripcion,omitempty" db:"descripcion"`
	DuracionMinutos int            `json:"duracion_minutos" db:"duracion_minutos"`
	Activo          bool           `json:"activo" db:"activo"`
	Orden           int            `json:"orden" db:"orden"`
	TiposAtencion   []TipoAtencion `json:"tipos_atencion,omitempty" db:"-"`
}

// TipoAtencion representa un tipo de atención disponible en un servicio
type TipoAtencion struct {
	ID         int    `json:"id" db:"id"`
	ServicioID int    `json:"servicio_id" db:"servicio_id"`
	Nombre     string `json:"nombre" db:"nombre"`
	Orden      int    `json:"orden" db:"orden"`
	Activo     bool   `json:"activo" db:"activo"`
}
