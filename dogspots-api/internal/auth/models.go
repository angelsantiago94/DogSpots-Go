package auth

import (
	"github.com/google/uuid"
)

// RegisterRequest es el payload para crear un nuevo usuario.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest es el payload para iniciar sesión.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest es el payload para renovar los tokens.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse es la respuesta tras un login o registro exitoso.
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	UserID       uuid.UUID `json:"user_id"`
	ExpiresIn    int64     `json:"expires_in"` // Segundos hasta la expiración del access token
}

// UserResponse es la representación de un usuario para la API.
type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}
