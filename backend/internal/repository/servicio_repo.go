package repository

import (
	"context"

	"kora/backend/internal/database"
	"kora/backend/internal/model"
)

// ListarServicios devuelve el catálogo de servicios activos con sus
// tipos de atención ordenados.
func ListarServicios(ctx context.Context, soloActivos bool) ([]model.Servicio, error) {
	query := `
		SELECT id, codigo, nombre, COALESCE(descripcion, ''), duracion_minutos, activo, orden
		FROM servicios
	`

	if soloActivos {
		query += ` WHERE activo = TRUE`
	}

	query += ` ORDER BY orden ASC, nombre ASC`

	filas, err := database.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	servicios := []model.Servicio{}
	for filas.Next() {
		var s model.Servicio
		if err := filas.Scan(&s.ID, &s.Codigo, &s.Nombre, &s.Descripcion, &s.DuracionMinutos, &s.Activo, &s.Orden); err != nil {
			filas.Close()
			return nil, err
		}
		s.TiposAtencion = []model.TipoAtencion{}
		servicios = append(servicios, s)
	}

	if err := filas.Err(); err != nil {
		filas.Close()
		return nil, err
	}
	filas.Close()

	if len(servicios) == 0 {
		return servicios, nil
	}

	tipos, err := ListarTiposAtencion(ctx)
	if err != nil {
		return nil, err
	}

	porServicio := make(map[int][]model.TipoAtencion, len(servicios))
	for _, tipo := range tipos {
		porServicio[tipo.ServicioID] = append(porServicio[tipo.ServicioID], tipo)
	}

	for i := range servicios {
		if lista, ok := porServicio[servicios[i].ID]; ok {
			servicios[i].TiposAtencion = lista
		}
	}

	return servicios, nil
}

// ListarTiposAtencion devuelve todos los tipos de atención
func ListarTiposAtencion(ctx context.Context) ([]model.TipoAtencion, error) {
	query := `
		SELECT id, servicio_id, nombre, orden, activo
		FROM tipos_atencion
		WHERE activo = TRUE
		ORDER BY servicio_id ASC, orden ASC, nombre ASC
	`

	filas, err := database.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	tipos := []model.TipoAtencion{}
	for filas.Next() {
		var t model.TipoAtencion
		if err := filas.Scan(&t.ID, &t.ServicioID, &t.Nombre, &t.Orden, &t.Activo); err != nil {
			return nil, err
		}
		tipos = append(tipos, t)
	}

	return tipos, filas.Err()
}

// ObtenerServicio busca un servicio por su ID
func ObtenerServicio(ctx context.Context, id int) (*model.Servicio, error) {
	query := `
		SELECT id, codigo, nombre, COALESCE(descripcion, ''), duracion_minutos, activo, orden
		FROM servicios
		WHERE id = $1
	`

	var s model.Servicio
	err := database.DB.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Codigo, &s.Nombre, &s.Descripcion, &s.DuracionMinutos, &s.Activo, &s.Orden,
	)
	if err != nil {
		return nil, err
	}

	tipos, err := ListarTiposAtencion(ctx)
	if err != nil {
		return nil, err
	}
	for _, t := range tipos {
		if t.ServicioID == s.ID {
			s.TiposAtencion = append(s.TiposAtencion, t)
		}
	}

	return &s, nil
}

// ObtenerTipoAtencion busca un tipo de atención por su ID
func ObtenerTipoAtencion(ctx context.Context, id int) (*model.TipoAtencion, error) {
	query := `SELECT id, servicio_id, nombre, orden, activo FROM tipos_atencion WHERE id = $1`

	var t model.TipoAtencion
	err := database.DB.QueryRow(ctx, query, id).Scan(&t.ID, &t.ServicioID, &t.Nombre, &t.Orden, &t.Activo)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
