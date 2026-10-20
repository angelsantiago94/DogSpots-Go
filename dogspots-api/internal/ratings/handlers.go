package ratings

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/angellsan94/dogspots-api/internal/auth"
	apierrors "github.com/angellsan94/dogspots-api/internal/errors"
	"github.com/google/uuid"
)

type Handler struct {
	repo Repository
}

type Repository interface {
	Upsert(ctx context.Context, placeID, userID uuid.UUID, value int) (Summary, error)
	Delete(ctx context.Context, placeID, userID uuid.UUID) (Summary, error)
	Summary(ctx context.Context, placeID uuid.UUID, viewerID *uuid.UUID) (Summary, error)
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized").Write(w)
		return
	}
	placeID, err := parsePathID(r, "placeId")
	if err != nil {
		apierrors.NewProblem(http.StatusBadRequest, "Bad Request", "Invalid place identifier", "invalid_place_id").Write(w)
		return
	}
	var req UpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error").Write(w)
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", "rating must be between 1 and 5", "validation_error").Write(w)
		return
	}
	result, err := h.repo.Upsert(r.Context(), placeID, userID, req.Rating)
	if err != nil {
		if strings.Contains(err.Error(), "place not found") {
			apierrors.NewProblem(http.StatusNotFound, "Not Found", "Place not found", "place_not_found").Write(w)
			return
		}
		apierrors.NewProblem(http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error").Write(w)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized").Write(w)
		return
	}
	placeID, err := parsePathID(r, "placeId")
	if err != nil {
		apierrors.NewProblem(http.StatusBadRequest, "Bad Request", "Invalid place identifier", "invalid_place_id").Write(w)
		return
	}
	result, err := h.repo.Delete(r.Context(), placeID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "place not found") {
			apierrors.NewProblem(http.StatusNotFound, "Not Found", "Place not found", "place_not_found").Write(w)
			return
		}
		apierrors.NewProblem(http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error").Write(w)
		return
	}
	writeJSON(w, http.StatusOK, result)
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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func parsePathID(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(key))
}
