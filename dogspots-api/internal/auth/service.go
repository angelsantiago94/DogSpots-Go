package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuthService defines the business logic for authentication.
type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	Refresh(ctx context.Context, req RefreshRequest) (*AuthResponse, error)
	Logout(ctx context.Context, userID uuid.UUID, refreshTokenID uuid.UUID) error
}

// authService is the implementation of AuthService.
type authService struct {
	repo         UserRepository
	hasher       PasswordHasher
	tokenManager *TokenManager
	accessTTL    time.Duration
	refreshTTL   time.Duration
	dummyHash    string
}

func NewAuthService(repo UserRepository, hasher PasswordHasher, tm *TokenManager, accessTTL, refreshTTL time.Duration) AuthService {
	// Hash desechable para igualar el coste de login cuando el email no existe.
	dummyHash, _ := hasher.HashPassword("dummy-password-for-timing")
	return &authService{
		repo:         repo,
		hasher:       hasher,
		tokenManager: tm,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
		dummyHash:    dummyHash,
	}
}

func validateRegisterRequest(req RegisterRequest) error {
	email := strings.TrimSpace(req.Email)
	name := strings.TrimSpace(req.Name)
	if email == "" || !strings.Contains(email, "@") {
		return fmt.Errorf("email is required")
	}
	if len(req.Password) < 10 {
		return fmt.Errorf("password must be at least 10 characters")
	}
	if name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetUserByEmail(ctx, strings.TrimSpace(req.Email)); err == nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID := uuid.New()
	now := time.Now().UTC()
	user, err := s.repo.CreateUser(ctx, UserCreateParams{
		ID:           userID,
		Email:        strings.TrimSpace(req.Email),
		PasswordHash: hashedPassword,
		Name:         strings.TrimSpace(req.Name),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.generateAuthResponse(ctx, user.ID, uuid.New())
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, strings.TrimSpace(req.Email))
	if err != nil {
		// Mismo coste que un login válido para no filtrar qué emails existen.
		_ = s.hasher.VerifyPassword(req.Password, s.dummyHash)
		return nil, ErrInvalidCredentials
	}

	if err := s.hasher.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateAuthResponse(ctx, user.ID, uuid.New())
}

func (s *authService) Refresh(ctx context.Context, req RefreshRequest) (*AuthResponse, error) {
	if strings.TrimSpace(req.RefreshToken) == "" {
		return nil, ErrTokenInvalid
	}

	tokenHash := hashRefreshToken(req.RefreshToken)
	stored, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	if stored.IsRevoked {
		if revokeErr := s.repo.RevokeAllTokensByFamily(ctx, stored.FamilyID); revokeErr != nil {
			return nil, revokeErr
		}
		return nil, ErrTokenRevoked
	}
	if stored.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrTokenExpired
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(stored.UserID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	newRefreshTokenHash := hashRefreshToken(newRefreshToken)
	newTokenID := uuid.New()
	newExpiresAt := time.Now().UTC().Add(s.refreshTTL)

	rotated, err := s.repo.RotateRefreshToken(ctx, stored.ID, RefreshTokenParams{
		ID:        newTokenID,
		UserID:    stored.UserID,
		TokenHash: newRefreshTokenHash,
		FamilyID:  stored.FamilyID,
		ExpiresAt: newExpiresAt,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("rotating refresh token: %w", err)
	}
	if !rotated {
		// Otro request rotó este token primero: se trata como reutilización.
		if revokeErr := s.repo.RevokeAllTokensByFamily(ctx, stored.FamilyID); revokeErr != nil {
			return nil, revokeErr
		}
		return nil, ErrTokenRevoked
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		UserID:       stored.UserID,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *authService) generateAuthResponse(ctx context.Context, userID uuid.UUID, familyID uuid.UUID) (*AuthResponse, error) {
	accessToken, err := s.tokenManager.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refreshTokenStr, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	if _, err = s.repo.CreateRefreshToken(ctx, RefreshTokenParams{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hashRefreshToken(refreshTokenStr),
		FamilyID:  familyID,
		ExpiresAt: time.Now().UTC().Add(s.refreshTTL),
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return nil, fmt.Errorf("failed to persist refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		UserID:       userID,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID uuid.UUID, refreshTokenID uuid.UUID) error {
	token, err := s.repo.GetRefreshToken(ctx, refreshTokenID)
	if err != nil {
		return ErrTokenInvalid
	}
	if token.UserID != userID {
		return ErrTokenInvalid
	}
	return s.repo.RevokeRefreshToken(ctx, refreshTokenID)
}
