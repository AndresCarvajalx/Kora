package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"kora/backend/internal/database"
	"kora/backend/internal/model"
)

var ErrUsuarioNoEncontrado = errors.New("usuario no encontrado")

// ObtenerUsuarioPorEmail busca un usuario por su email
func ObtenerUsuarioPorEmail(ctx context.Context, email string) (*model.Usuario, error) {
	query := `
		SELECT id, nombre, email, password_hash, rol, creado_en
		FROM usuarios
		WHERE email = $1
	`

	var usuario model.Usuario
	err := database.DB.QueryRow(ctx, query, email).Scan(
		&usuario.ID,
		&usuario.Nombre,
		&usuario.Email,
		&usuario.PasswordHash,
		&usuario.Rol,
		&usuario.CreadoEn,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUsuarioNoEncontrado
		}
		return nil, err
	}

	return &usuario, nil
}

// ObtenerUsuarioPorID busca un usuario por su ID
func ObtenerUsuarioPorID(ctx context.Context, id int) (*model.Usuario, error) {
	query := `
		SELECT id, nombre, email, password_hash, rol, creado_en
		FROM usuarios
		WHERE id = $1
	`

	var usuario model.Usuario
	err := database.DB.QueryRow(ctx, query, id).Scan(
		&usuario.ID,
		&usuario.Nombre,
		&usuario.Email,
		&usuario.PasswordHash,
		&usuario.Rol,
		&usuario.CreadoEn,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUsuarioNoEncontrado
		}
		return nil, err
	}

	return &usuario, nil
}
