package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserRepository define las operaciones de base de datos necesarias para la autenticación.
type UserRepository interface {
	// Métodos de Usuario
	CreateUser(ctx context.Context, user UserCreateParams) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, name string) (User, error)

	// Métodos de Token
	CreateRefreshToken(ctx context.Context, token RefreshTokenParams) (RefreshToken, error)
	RotateRefreshToken(ctx context.Context, id uuid.UUID, token RefreshTokenParams) (bool, error)
	GetRefreshToken(ctx context.Context, id uuid.UUID) (RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshToken, error)
	ListRefreshTokensByFamily(ctx context.Context, familyID uuid.UUID) ([]RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
	RevokeAllTokensByFamily(ctx context.Context, familyID uuid.UUID) error
}

// User y RefreshToken son tipos que representan las filas de la base de datos.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	FamilyID  uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
	IsRevoked bool
}

// UserCreateParams es el parámetro para crear un usuario.
type UserCreateParams struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RefreshTokenParams es el parámetro para crear un refresh token.
type RefreshTokenParams struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	FamilyID  uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
}
