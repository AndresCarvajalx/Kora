package database

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate aplica, en orden, las migraciones pendientes del directorio
// migrations/. Cada archivo se ejecuta una sola vez y queda registrado
// en la tabla schema_migrations.
func Migrate(ctx context.Context) error {
	const crearTablaMigraciones = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			aplicada_en TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`

	if _, err := DB.Exec(ctx, crearTablaMigraciones); err != nil {
		return fmt.Errorf("error creando schema_migrations: %w", err)
	}

	aplicadas, err := migracionesAplicadas(ctx)
	if err != nil {
		return err
	}

	nombres, err := nombresMigraciones()
	if err != nil {
		return err
	}

	for _, nombre := range nombres {
		version := strings.TrimSuffix(strings.TrimPrefix(nombre, "migrations/"), ".sql")

		if aplicadas[version] {
			continue
		}

		contenido, err := migrationsFS.ReadFile(nombre)
		if err != nil {
			return fmt.Errorf("error leyendo migración %s: %w", version, err)
		}

		if _, err := DB.Exec(ctx, string(contenido)); err != nil {
			return fmt.Errorf("error aplicando migración %s: %w", version, err)
		}

		if _, err := DB.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING`,
			version,
		); err != nil {
			return fmt.Errorf("error registrando migración %s: %w", version, err)
		}
	}

	return nil
}

// migracionesAplicadas devuelve el conjunto de versiones ya ejecutadas
func migracionesAplicadas(ctx context.Context) (map[string]bool, error) {
	filas, err := DB.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("error consultando schema_migrations: %w", err)
	}
	defer filas.Close()

	aplicadas := make(map[string]bool)
	for filas.Next() {
		var version string
		if err := filas.Scan(&version); err != nil {
			return nil, err
		}
		aplicadas[version] = true
	}

	return aplicadas, filas.Err()
}

// nombresMigraciones lista los archivos .sql ordenados por versión
func nombresMigraciones() ([]string, error) {
	entradas, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("error listando migraciones: %w", err)
	}

	nombres := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		if !entrada.IsDir() && strings.HasSuffix(entrada.Name(), ".sql") {
			nombres = append(nombres, "migrations/"+entrada.Name())
		}
	}

	sort.Strings(nombres)
	return nombres, nil
}
