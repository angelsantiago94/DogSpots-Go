package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeUserRepository struct {
	users         map[string]User
	byID          map[uuid.UUID]User
	refreshByID   map[uuid.UUID]RefreshToken
	refreshByHash map[string]RefreshToken
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		users:         make(map[string]User),
		byID:          make(map[uuid.UUID]User),
		refreshByID:   make(map[uuid.UUID]RefreshToken),
		refreshByHash: make(map[string]RefreshToken),
	}
}

func TestArgon2HasherUsesConfiguredParameters(t *testing.T) {
	hasher := NewArgon2HasherWithParams(Argon2Params{Memory: 8 * 1024, Iterations: 1, Parallelism: 1})
	hash, err := hasher.HashPassword("a-secure-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if !strings.Contains(hash, "$m=8192,t=1,p=1$") {
		t.Fatalf("hash does not contain configured Argon2 parameters: %q", hash)
	}
	if err := hasher.VerifyPassword("a-secure-password", hash); err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
}

func (r *fakeUserRepository) CreateUser(_ context.Context, user UserCreateParams) (User, error) {
	if _, exists := r.users[user.Email]; exists {
		return User{}, ErrUserAlreadyExists
	}
	stored := User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
	r.users[user.Email] = stored
	r.byID[user.ID] = stored
	return stored, nil
}

func (r *fakeUserRepository) GetUserByEmail(_ context.Context, email string) (User, error) {
	user, ok := r.users[email]
	if !ok {
		return User{}, errors.New("not found")
	}
	return user, nil
}

func (r *fakeUserRepository) GetUserByID(_ context.Context, id uuid.UUID) (User, error) {
	user, ok := r.byID[id]
	if !ok {
		return User{}, errors.New("not found")
	}
	return user, nil
}

func (r *fakeUserRepository) UpdateUser(_ context.Context, id uuid.UUID, name string) (User, error) {
	user, ok := r.byID[id]
	if !ok {
		return User{}, errors.New("not found")
	}
	user.Name = name
	r.byID[id] = user
	for email, stored := range r.users {
		if stored.ID == id {
			stored.Name = name
			r.users[email] = stored
			break
		}
	}
	return user, nil
}

func (r *fakeUserRepository) CreateRefreshToken(_ context.Context, token RefreshTokenParams) (RefreshToken, error) {
	stored := RefreshToken{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		FamilyID:  token.FamilyID,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
		IsRevoked: false,
	}
	r.refreshByID[stored.ID] = stored
	r.refreshByHash[stored.TokenHash] = stored
	return stored, nil
}

func (r *fakeUserRepository) GetRefreshToken(_ context.Context, id uuid.UUID) (RefreshToken, error) {
	user, ok := r.refreshByID[id]
	if !ok {
		return RefreshToken{}, errors.New("not found")
	}
	return user, nil
}

func (r *fakeUserRepository) GetRefreshTokenByHash(_ context.Context, tokenHash string) (RefreshToken, error) {
	user, ok := r.refreshByHash[tokenHash]
	if !ok {
		return RefreshToken{}, errors.New("not found")
	}
	return user, nil
}

func (r *fakeUserRepository) ListRefreshTokensByFamily(_ context.Context, familyID uuid.UUID) ([]RefreshToken, error) {
	var out []RefreshToken
	for _, token := range r.refreshByID {
		if token.FamilyID == familyID {
			out = append(out, token)
		}
	}
	return out, nil
}

func (r *fakeUserRepository) RevokeRefreshToken(_ context.Context, id uuid.UUID) error {
	token, ok := r.refreshByID[id]
	if !ok {
		return errors.New("not found")
	}
	token.IsRevoked = true
	r.refreshByID[id] = token
	r.refreshByHash[token.TokenHash] = token
	return nil
}

func (r *fakeUserRepository) RevokeAllTokensByFamily(_ context.Context, familyID uuid.UUID) error {
	for id, token := range r.refreshByID {
		if token.FamilyID == familyID {
			token.IsRevoked = true
			r.refreshByID[id] = token
			r.refreshByHash[token.TokenHash] = token
		}
	}
	return nil
}

func TestRegisterCreatesUserAndSession(t *testing.T) {
	repo := newFakeUserRepository()
	service := NewAuthService(repo, NewArgon2Hasher(), NewTokenManager("test-secret", time.Minute), time.Minute, time.Hour)

	resp, err := service.Register(context.Background(), RegisterRequest{
		Email:    "test@example.com",
		Password: "s3cur3passw0rd",
		Name:     "Alice",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatal("Register() should return both tokens")
	}
	if resp.UserID == uuid.Nil {
		t.Fatal("Register() should return user ID")
	}
	if repo.users["test@example.com"].Name != "Alice" {
		t.Fatal("Register() should create user record")
	}
}

func TestLoginRejectsInvalidCredentials(t *testing.T) {
	repo := newFakeUserRepository()
	service := NewAuthService(repo, NewArgon2Hasher(), NewTokenManager("test-secret", time.Minute), time.Minute, time.Hour)

	_, _ = service.Register(context.Background(), RegisterRequest{
		Email:    "test@example.com",
		Password: "s3cur3passw0rd",
		Name:     "Alice",
	})

	_, err := service.Login(context.Background(), LoginRequest{Email: "test@example.com", Password: "wrong-password"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestRefreshRotatesTokenAndRejectsReplay(t *testing.T) {
	repo := newFakeUserRepository()
	service := NewAuthService(repo, NewArgon2Hasher(), NewTokenManager("test-secret", time.Minute), time.Minute, time.Hour)

	resp, err := service.Register(context.Background(), RegisterRequest{
		Email:    "test@example.com",
		Password: "s3cur3passw0rd",
		Name:     "Alice",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	rotated, err := service.Refresh(context.Background(), RefreshRequest{RefreshToken: resp.RefreshToken})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if rotated.RefreshToken == resp.RefreshToken {
		t.Fatal("Refresh() should rotate the token")
	}

	_, err = service.Refresh(context.Background(), RefreshRequest{RefreshToken: resp.RefreshToken})
	if !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("Refresh replay error = %v, want %v", err, ErrTokenRevoked)
	}
}

func TestLogoutRevokesRefreshToken(t *testing.T) {
	repo := newFakeUserRepository()
	service := NewAuthService(repo, NewArgon2Hasher(), NewTokenManager("test-secret", time.Minute), time.Minute, time.Hour)

	resp, err := service.Register(context.Background(), RegisterRequest{
		Email:    "test@example.com",
		Password: "s3cur3passw0rd",
		Name:     "Alice",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	token, err := repo.GetRefreshTokenByHash(context.Background(), hashRefreshToken(resp.RefreshToken))
	if err != nil {
		t.Fatalf("GetRefreshTokenByHash() error = %v", err)
	}

	if err := service.Logout(context.Background(), resp.UserID, token.ID); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	if token, err := repo.GetRefreshToken(context.Background(), token.ID); err == nil && !token.IsRevoked {
		t.Fatal("Logout() should revoke the refresh token")
	}
}
