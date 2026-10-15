package model

import (
	"strings"
	"time"
)

// Turno representa un turno en el sistema
type Turno struct {
	ID                 int        `json:"id" db:"id"`
	NumeroTurno        string     `json:"numero_turno,omitempty" db:"numero_turno"`
	ServicioID         *int       `json:"servicio_id,omitempty" db:"servicio_id"`
	ServicioCodigo     string     `json:"servicio_codigo,omitempty" db:"servicio_codigo"`
	ServicioNombre     string     `json:"servicio_nombre,omitempty" db:"servicio_nombre"`
	TipoAtencionID     *int       `json:"tipo_atencion_id,omitempty" db:"tipo_atencion_id"`
	TipoAtencionNombre string     `json:"tipo_atencion_nombre,omitempty" db:"tipo_atencion_nombre"`
	ClienteContacto    string     `json:"cliente_contacto" db:"cliente_contacto"`
	ClienteNombre      string     `json:"cliente_nombre,omitempty" db:"cliente_nombre"`
	ClienteDocumento   string     `json:"cliente_documento,omitempty" db:"cliente_documento"`
	FechaHora          time.Time  `json:"fecha_hora" db:"fecha_hora"`
	ExpiraEn           time.Time  `json:"expira_en" db:"expira_en"`
	Estado             string     `json:"estado" db:"estado"`
	OrdenCola          *int       `json:"orden_cola,omitempty" db:"orden_cola"`
	ColaFecha          *time.Time `json:"-" db:"cola_fecha"`
	LlamadoEn          *time.Time `json:"llamado_en,omitempty" db:"llamado_en"`
	AtendidoEn         *time.Time `json:"atendido_en,omitempty" db:"atendido_en"`
	CanceladoEn        *time.Time `json:"cancelado_en,omitempty" db:"cancelado_en"`
	MotivoCancelacion  string     `json:"motivo_cancelacion,omitempty" db:"motivo_cancelacion"`
	CreadoEn           time.Time  `json:"creado_en" db:"creado_en"`
}

// TurnoPublico es la vista pública del turno sin datos personales (Ley 1581 de 2012).
// Excluye cliente_nombre, cliente_contacto y cliente_documento.
type TurnoPublico struct {
	ID                 int        `json:"id,omitempty"`
	NumeroTurno        string     `json:"numero_turno"`
	ServicioID         *int       `json:"servicio_id,omitempty"`
	ServicioCodigo     string     `json:"servicio_codigo,omitempty"`
	ServicioNombre     string     `json:"servicio_nombre,omitempty"`
	TipoAtencionID     *int       `json:"tipo_atencion_id,omitempty"`
	TipoAtencionNombre string     `json:"tipo_atencion_nombre,omitempty"`
	FechaHora          time.Time  `json:"fecha_hora,omitempty"`
	ExpiraEn           time.Time  `json:"expira_en,omitempty"`
	Estado             string     `json:"estado"`
	OrdenCola          *int       `json:"orden_cola,omitempty"`
	LlamadoEn          *time.Time `json:"llamado_en,omitempty"`
	AtendidoEn         *time.Time `json:"atendido_en,omitempty"`
	CanceladoEn        *time.Time `json:"cancelado_en,omitempty"`
	MotivoCancelacion  string     `json:"motivo_cancelacion,omitempty"`
	CreadoEn           time.Time  `json:"creado_en,omitempty"`
}

// TicketTurno agrega al turno la información de cola que ve el cliente:
// posición, cuántas personas hay antes y el tiempo estimado de espera.
type TicketTurno struct {
	Turno
	PersonasAntes     int `json:"personas_antes"`
	Posicion          int `json:"posicion"`
	TotalEnCola       int `json:"total_en_cola"`
	TiempoEstimadoMin int `json:"tiempo_estimado_min"`
}

// TicketPublico es la vista pública del ticket sin datos personales.
type TicketPublico struct {
	TurnoPublico
	PersonasAntes     int `json:"personas_antes"`
	Posicion          int `json:"posicion"`
	TotalEnCola       int `json:"total_en_cola"`
	TiempoEstimadoMin int `json:"tiempo_estimado_min"`
}

// AConsultaPublica extrae los campos públicos y anónimos de un Turno
func (t *Turno) AConsultaPublica() *TurnoPublico {
	if t == nil {
		return nil
	}
	return &TurnoPublico{
		ID:                 t.ID,
		NumeroTurno:        t.NumeroTurno,
		ServicioID:         t.ServicioID,
		ServicioCodigo:     t.ServicioCodigo,
		ServicioNombre:     t.ServicioNombre,
		TipoAtencionID:     t.TipoAtencionID,
		TipoAtencionNombre: t.TipoAtencionNombre,
		FechaHora:          t.FechaHora,
		ExpiraEn:           t.ExpiraEn,
		Estado:             t.Estado,
		OrdenCola:          t.OrdenCola,
		LlamadoEn:          t.LlamadoEn,
		AtendidoEn:         t.AtendidoEn,
		CanceladoEn:        t.CanceladoEn,
		MotivoCancelacion:  t.MotivoCancelacion,
		CreadoEn:           t.CreadoEn,
	}
}

// ATicketPublico convierte un TicketTurno en su vista pública sin datos personales
func (tt *TicketTurno) ATicketPublico() *TicketPublico {
	if tt == nil {
		return nil
	}
	pub := tt.Turno.AConsultaPublica()
	return &TicketPublico{
		TurnoPublico:      *pub,
		PersonasAntes:     tt.PersonasAntes,
		Posicion:          tt.Posicion,
		TotalEnCola:       tt.TotalEnCola,
		TiempoEstimadoMin: tt.TiempoEstimadoMin,
	}
}

// NormalizarContacto elimina espacios y guiones para comparar teléfonos/contactos
func NormalizarContacto(c string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(c), "-", ""), " ", "")
}

// CrearTurnoRequest representa la petición para crear un turno
type CrearTurnoRequest struct {
	Contacto       string `json:"contacto"`
	Nombre         string `json:"nombre,omitempty"`
	Documento      string `json:"documento,omitempty"`
	ServicioID     int    `json:"servicio_id"`
	TipoAtencionID int    `json:"tipo_atencion_id,omitempty"`
	FechaHora      string `json:"fecha_hora,omitempty"`
}

// ConsultarTurnoRequest representa la petición para consultar un turno.
// El query puede ser el número de turno (CAR-014) o el contacto.
type ConsultarTurnoRequest struct {
	Query    string `json:"query"`
	Contacto string `json:"contacto,omitempty"`
}

// CambiarEstadoRequest representa la petición para cambiar el estado
type CambiarEstadoRequest struct {
	Estado string `json:"estado"`
}

// CancelarTurnoRequest representa la petición de cancelación de un turno
type CancelarTurnoRequest struct {
	Motivo string `json:"motivo,omitempty"`
}

// CambiarServicioRequest representa el traslado de un turno a otro servicio
type CambiarServicioRequest struct {
	ServicioID     int    `json:"servicio_id"`
	TipoAtencionID int    `json:"tipo_atencion_id,omitempty"`
	Motivo         string `json:"motivo,omitempty"`
}

// Estados válidos de turno
const (
	EstadoPendiente  = "PENDIENTE"
	EstadoEnAtencion = "EN_ATENCION"
	EstadoAtendido   = "ATENDIDO"
	EstadoCancelado  = "CANCELADO"
	EstadoVencido    = "VENCIDO"
)
