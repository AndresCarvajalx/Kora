package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB es el pool de conexiones a PostgreSQL
var DB *pgxpool.Pool

// InitDB inicializa el pool de conexiones
func InitDB() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL no está definida")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("error parseando DATABASE_URL: %w", err)
	}

	// Configuración del pool
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	DB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("error creando pool de conexiones: %w", err)
	}

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := DB.Ping(ctx); err != nil {
		return fmt.Errorf("error conectando a la base de datos: %w", err)
	}

	return nil
}

// CloseDB cierra el pool de conexiones
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
