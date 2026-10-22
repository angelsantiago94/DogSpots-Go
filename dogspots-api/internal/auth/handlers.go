package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	apierrors "github.com/angellsan94/dogspots-api/internal/errors"
	"github.com/google/uuid"
)

type Handler struct {
	service AuthService
	repo    UserRepository
}

func NewHandler(service AuthService, repo UserRepository) *Handler {
	return &Handler{service: service, repo: repo}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error").Write(w)
		return
	}

	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		writeAuthProblem(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error").Write(w)
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		writeAuthProblem(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error").Write(w)
		return
	}

	resp, err := h.service.Refresh(r.Context(), req)
	if err != nil {
		writeAuthProblem(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error").Write(w)
		return
	}

	if strings.TrimSpace(req.RefreshToken) == "" {
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "Refresh token is required", "unauthorized").Write(w)
		return
	}

	tokenHash := hashRefreshToken(req.RefreshToken)
	token, err := h.repo.GetRefreshTokenByHash(r.Context(), tokenHash)
	if err != nil {
		writeAuthProblem(w, ErrTokenInvalid)
		return
	}

	if err := h.service.Logout(r.Context(), token.UserID, token.ID); err != nil {
		writeAuthProblem(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized").Write(w)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		apierrors.NewProblem(http.StatusNotFound, "Not Found", "User not found", "user_not_found").Write(w)
		return
	}

	writeUserResponse(w, user)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized").Write(w)
		return
	}

	var payload struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error").Write(w)
		return
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", "Name is required", "validation_error").Write(w)
		return
	}

	user, err := h.repo.UpdateUser(r.Context(), userID, name)
	if err != nil {
		apierrors.NewProblem(http.StatusNotFound, "Not Found", "User not found", "user_not_found").Write(w)
		return
	}

	writeUserResponse(w, user)
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		apierrors.NewProblem(http.StatusBadRequest, "Bad Request", "Invalid user identifier", "invalid_user_id").Write(w)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		apierrors.NewProblem(http.StatusNotFound, "Not Found", "User not found", "user_not_found").Write(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(PublicUserResponse{ID: user.ID, Name: user.Name})
}

func decodeJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return err
	}
	return nil
}

func writeUserResponse(w http.ResponseWriter, user User) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	})
}

func writeAuthProblem(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUserAlreadyExists):
		apierrors.NewProblem(http.StatusConflict, "Conflict", "A user with that email already exists", "user_already_exists").Write(w)
	case errors.Is(err, ErrInvalidCredentials):
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "Invalid email or password", "invalid_credentials").Write(w)
	case errors.Is(err, ErrTokenExpired):
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "The refresh token has expired", "token_expired").Write(w)
	case errors.Is(err, ErrTokenRevoked):
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "The refresh token has been revoked", "token_revoked").Write(w)
	case errors.Is(err, ErrTokenInvalid):
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "The refresh token is invalid", "invalid_refresh_token").Write(w)
	default:
		apierrors.NewProblem(http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error").Write(w)
	}
}
