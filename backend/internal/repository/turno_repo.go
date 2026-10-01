package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"kora/backend/internal/database"
	"kora/backend/internal/model"
)

var ErrTurnoNoEncontrado = errors.New("turno no encontrado")

// columnasTurno es la lista de proyección compartida por todas las consultas
const columnasTurno = `
	t.id, t.numero_turno, t.servicio_id,
	COALESCE(s.codigo, ''), COALESCE(s.nombre, ''),
	t.tipo_atencion_id, COALESCE(ta.nombre, ''),
	t.cliente_contacto, COALESCE(t.cliente_nombre, ''), COALESCE(t.cliente_documento, ''),
	t.fecha_hora, t.expira_en, t.estado, t.orden_cola, t.cola_fecha,
	t.llamado_en, t.atendido_en, t.cancelado_en,
	COALESCE(t.motivo_cancelacion, ''), t.creado_en`

// joinsTurno une el catálogo de servicios para traer nombre y código
const joinsTurno = `
	FROM turnos t
	LEFT JOIN servicios s ON s.id = t.servicio_id
	LEFT JOIN tipos_atencion ta ON ta.id = t.tipo_atencion_id`

// escaneador abstrae pgx.Row y pgx.Rows
type escaneador interface {
	Scan(dest ...interface{}) error
}

// escanearTurno mapea una fila al modelo de turno
func escanearTurno(f escaneador) (*model.Turno, error) {
	var turno model.Turno

	err := f.Scan(
		&turno.ID,
		&turno.NumeroTurno,
		&turno.ServicioID,
		&turno.ServicioCodigo,
		&turno.ServicioNombre,
		&turno.TipoAtencionID,
		&turno.TipoAtencionNombre,
		&turno.ClienteContacto,
		&turno.ClienteNombre,
		&turno.ClienteDocumento,
		&turno.FechaHora,
		&turno.ExpiraEn,
		&turno.Estado,
		&turno.OrdenCola,
		&turno.ColaFecha,
		&turno.LlamadoEn,
		&turno.AtendidoEn,
		&turno.CanceladoEn,
		&turno.MotivoCancelacion,
		&turno.CreadoEn,
	)
	if err != nil {
		return nil, err
	}

	return &turno, nil
}

// NuevoTurno son los datos necesarios para insertar un turno
type NuevoTurno struct {
	Contacto       string
	Nombre         string
	Documento      string
	ServicioID     int
	ServicioCodigo string
	TipoAtencionID *int
	FechaHora      time.Time
	ExpiraEn       time.Time
	ColaFecha      time.Time
}

// CrearTurno inserta un nuevo turno asignándole número y posición de cola
// dentro de una misma transacción.
func CrearTurno(ctx context.Context, nuevo NuevoTurno) (*model.Turno, error) {
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op si ya se hizo commit

	orden, err := siguienteNumeroTurno(ctx, tx, nuevo.ServicioID, nuevo.ColaFecha)
	if err != nil {
		return nil, err
	}

	numero := fmt.Sprintf("%s-%03d", nuevo.ServicioCodigo, orden)

	query := `
		INSERT INTO turnos (
			cliente_contacto, cliente_nombre, cliente_documento,
			servicio_id, tipo_atencion_id, numero_turno, orden_cola, cola_fecha,
			fecha_hora, expira_en, estado
		)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	var id int
	err = tx.QueryRow(ctx, query,
		nuevo.Contacto,
		nuevo.Nombre,
		nuevo.Documento,
		nuevo.ServicioID,
		nuevo.TipoAtencionID,
		numero,
		orden,
		nuevo.ColaFecha,
		nuevo.FechaHora,
		nuevo.ExpiraEn,
		model.EstadoPendiente,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return ObtenerTurnoPorID(ctx, id)
}

// siguienteNumeroTurno reserva el siguiente consecutivo del día para el
// servicio indicado. La operación es atómica (INSERT ... ON CONFLICT).
func siguienteNumeroTurno(ctx context.Context, tx pgx.Tx, servicioID int, fecha time.Time) (int, error) {
	query := `
		INSERT INTO turno_secuencia (servicio_id, fecha, ultimo)
		VALUES ($1, $2, 1)
		ON CONFLICT (servicio_id, fecha)
		DO UPDATE SET ultimo = turno_secuencia.ultimo + 1
		RETURNING ultimo`

	var ultimo int
	if err := tx.QueryRow(ctx, query, servicioID, fecha).Scan(&ultimo); err != nil {
		return 0, err
	}

	return ultimo, nil
}

// ObtenerTurnoPorID busca un turno por su ID
func ObtenerTurnoPorID(ctx context.Context, id int) (*model.Turno, error) {
	query := `SELECT ` + columnasTurno + joinsTurno + ` WHERE t.id = $1`

	turno, err := escanearTurno(database.DB.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTurnoNoEncontrado
		}
		return nil, err
	}

	return turno, nil
}

// ObtenerTurnoPorNumero busca un turno por su número visible (CAR-014)
func ObtenerTurnoPorNumero(ctx context.Context, numero string) (*model.Turno, error) {
	query := `SELECT ` + columnasTurno + joinsTurno + `
		WHERE UPPER(t.numero_turno) = UPPER($1)`

	turno, err := escanearTurno(database.DB.QueryRow(ctx, query, strings.TrimSpace(numero)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTurnoNoEncontrado
		}
		return nil, err
	}

	return turno, nil
}

// ObtenerTurnosActivosPorContacto lista los turnos PENDIENTE/EN_ATENCION
// de un contacto, del más reciente al más antiguo.
func ObtenerTurnosActivosPorContacto(ctx context.Context, contacto string) ([]model.Turno, error) {
	query := `SELECT ` + columnasTurno + joinsTurno + `
		WHERE t.cliente_contacto = $1
		  AND t.estado IN ($2, $3)
		ORDER BY t.creado_en DESC`

	filas, err := database.DB.Query(ctx, query, contacto, model.EstadoPendiente, model.EstadoEnAtencion)
	if err != nil {
		return nil, err
	}

	turnos, err := escanearListaTurnos(filas)
	if err != nil {
		return nil, err
	}

	return turnos, nil
}

// ObtenerTurnoActivoPorContactoServicio busca el turno activo de un
// contacto en un servicio concreto.
func ObtenerTurnoActivoPorContactoServicio(ctx context.Context, contacto string, servicioID int) (*model.Turno, error) {
	query := `SELECT ` + columnasTurno + joinsTurno + `
		WHERE t.cliente_contacto = $1
		  AND t.servicio_id = $2
		  AND t.estado IN ($3, $4)
		ORDER BY t.creado_en DESC
		LIMIT 1`

	turno, err := escanearTurno(database.DB.QueryRow(ctx, query,
		contacto, servicioID, model.EstadoPendiente, model.EstadoEnAtencion))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // no hay turno activo
		}
		return nil, err
	}

	return turno, nil
}

// ListarTurnos lista turnos aplicando los filtros opcionales
func ListarTurnos(ctx context.Context, filtro model.FiltroTurnos) ([]model.Turno, error) {
	query := `SELECT ` + columnasTurno + joinsTurno + ` WHERE 1 = 1`
	args := []interface{}{}

	if filtro.Estado != "" {
		args = append(args, filtro.Estado)
		query += fmt.Sprintf(" AND t.estado = $%d", len(args))
	}

	if filtro.ServicioID > 0 {
		args = append(args, filtro.ServicioID)
		query += fmt.Sprintf(" AND t.servicio_id = $%d", len(args))
	}

	if filtro.Contacto != "" {
		args = append(args, filtro.Contacto)
		query += fmt.Sprintf(" AND t.cliente_contacto = $%d", len(args))
	}

	if filtro.Numero != "" {
		args = append(args, strings.ToUpper(filtro.Numero))
		query += fmt.Sprintf(" AND UPPER(t.numero_turno) = $%d", len(args))
	}

	if filtro.SoloHoy {
		query += " AND t.creado_en::date = CURRENT_DATE"
	}

	query += " ORDER BY COALESCE(t.orden_cola, t.id) ASC, t.creado_en ASC"

	if filtro.Limite > 0 {
		args = append(args, filtro.Limite)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}

	filas, err := database.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	turnos, err := escanearListaTurnos(filas)
	if err != nil {
		return nil, err
	}

	return turnos, nil
}

// ContarCola devuelve cuántas personas están antes del turno dentro de su
// cola (servicio + tipo de atención + día) y cuántas quedan en total.
func ContarCola(ctx context.Context, turno *model.Turno) (personasAntes int, total int, err error) {
	if turno.ServicioID == nil || turno.OrdenCola == nil {
		return 0, 0, nil
	}

	query := `
		SELECT
			COUNT(*) FILTER (
				WHERE estado = $4 AND orden_cola < $3
			),
			COUNT(*) FILTER (WHERE estado = $4)
		FROM turnos
		WHERE servicio_id = $1
		  AND tipo_atencion_id IS NOT DISTINCT FROM $2
		  AND cola_fecha = $5`

	args := []interface{}{*turno.ServicioID, turno.TipoAtencionID, *turno.OrdenCola, model.EstadoPendiente, *turno.ColaFecha}

	err = database.DB.QueryRow(ctx, query, args...).Scan(&personasAntes, &total)
	return personasAntes, total, err
}

// DuracionServicio devuelve la duración estimada de atención de un servicio
func DuracionServicio(ctx context.Context, servicioID int) (int, error) {
	query := `SELECT duracion_minutos FROM servicios WHERE id = $1`

	var minutos int
	if err := database.DB.QueryRow(ctx, query, servicioID).Scan(&minutos); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrTurnoNoEncontrado
		}
		return 0, err
	}

	return minutos, nil
}

// ActualizarEstadoTurno actualiza el estado y su marca de tiempo asociada
func ActualizarEstadoTurno(ctx context.Context, id int, nuevoEstado string) (*model.Turno, error) {
	// El nombre de columna sale de una lista cerrada, nunca de la entrada
	marca := ""
	switch nuevoEstado {
	case model.EstadoEnAtencion:
		marca = ", llamado_en = NOW()"
	case model.EstadoAtendido:
		marca = ", atendido_en = NOW()"
	case model.EstadoCancelado:
		marca = ", cancelado_en = NOW()"
	}

	query := fmt.Sprintf(`UPDATE turnos SET estado = $1%s WHERE id = $2 RETURNING id`, marca)

	var idActual int
	if err := database.DB.QueryRow(ctx, query, nuevoEstado, id).Scan(&idActual); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTurnoNoEncontrado
		}
		return nil, err
	}

	return ObtenerTurnoPorID(ctx, idActual)
}

// CancelarTurno registra la cancelación con su motivo
func CancelarTurno(ctx context.Context, id int, motivo string) (*model.Turno, error) {
	query := `
		UPDATE turnos
		SET estado = $1, cancelado_en = NOW(), motivo_cancelacion = NULLIF($2, '')
		WHERE id = $3
		RETURNING id`

	var idActual int
	if err := database.DB.QueryRow(ctx, query, model.EstadoCancelado, motivo, id).Scan(&idActual); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTurnoNoEncontrado
		}
		return nil, err
	}

	return ObtenerTurnoPorID(ctx, idActual)
}

// ReencolarTurno mueve un turno PENDIENTE a la cola de otro servicio,
// asignándole un número y posición nuevos.
func ReencolarTurno(ctx context.Context, id, servicioID int, tipoAtencionID *int, codigoServicio string) (*model.Turno, error) {
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op si ya se hizo commit

	fecha := time.Now()

	orden, err := siguienteNumeroTurno(ctx, tx, servicioID, fecha)
	if err != nil {
		return nil, err
	}

	numero := fmt.Sprintf("%s-%03d", codigoServicio, orden)

	query := `
		UPDATE turnos
		SET servicio_id = $1,
			tipo_atencion_id = $2,
			numero_turno = $3,
			orden_cola = $4,
			cola_fecha = $5
		WHERE id = $6
		RETURNING id`

	var idActual int
	err = tx.QueryRow(ctx, query, servicioID, tipoAtencionID, numero, orden, fecha, id).Scan(&idActual)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTurnoNoEncontrado
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return ObtenerTurnoPorID(ctx, idActual)
}

// MarcarTurnosVencidos marca como VENCIDO todos los turnos PENDIENTE
// cuya expiración ya pasó.
func MarcarTurnosVencidos(ctx context.Context) (int64, error) {
	query := `
		UPDATE turnos
		SET estado = $1
		WHERE estado = $2 AND expira_en < NOW()`

	tag, err := database.DB.Exec(ctx, query, model.EstadoVencido, model.EstadoPendiente)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

// ListarEnAtencion devuelve los turnos que están siendo atendidos ahora
func ListarEnAtencion(ctx context.Context) ([]model.Turno, error) {
	query := `SELECT ` + columnasTurno + joinsTurno + `
		WHERE t.estado = $1
		ORDER BY t.llamado_en DESC NULLS LAST, t.id DESC`

	filas, err := database.DB.Query(ctx, query, model.EstadoEnAtencion)
	if err != nil {
		return nil, err
	}

	return escanearListaTurnos(filas)
}

// ListarUltimosLlamados devuelve el historial reciente de llamadas
func ListarUltimosLlamados(ctx context.Context, limite int) ([]model.Turno, error) {
	if limite <= 0 {
		limite = 10
	}

	query := `SELECT ` + columnasTurno + joinsTurno + `
		WHERE t.llamado_en IS NOT NULL
		ORDER BY t.llamado_en DESC
		LIMIT $1`

	filas, err := database.DB.Query(ctx, query, limite)
	if err != nil {
		return nil, err
	}

	return escanearListaTurnos(filas)
}

// ContarPendientes devuelve el total de turnos pendientes
func ContarPendientes(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM turnos WHERE estado = $1`

	var total int
	err := database.DB.QueryRow(ctx, query, model.EstadoPendiente).Scan(&total)
	return total, err
}

// escanearListaTurnos recorre un pgx.Rows y devuelve los turnos encontrados
func escanearListaTurnos(filas pgx.Rows) ([]model.Turno, error) {
	defer filas.Close()

	turnos := []model.Turno{}
	for filas.Next() {
		turno, err := escanearTurno(filas)
		if err != nil {
			return nil, err
		}
		turnos = append(turnos, *turno)
	}

	return turnos, filas.Err()
}
