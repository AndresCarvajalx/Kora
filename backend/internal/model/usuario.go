package model

import "time"

// Usuario representa un usuario del sistema (cliente o admin)
type Usuario struct {
	ID           int       `json:"id" db:"id"`
	Nombre       string    `json:"nombre" db:"nombre"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Rol          string    `json:"rol" db:"rol"`
	CreadoEn     time.Time `json:"creado_en" db:"creado_en"`
}

// LoginRequest representa la petición de login del admin
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse representa la respuesta del login
type LoginResponse struct {
	Token string `json:"token"`
}
