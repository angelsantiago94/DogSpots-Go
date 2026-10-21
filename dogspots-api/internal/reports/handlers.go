package reports

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/angellsan94/dogspots-api/internal/auth"
	apierrors "github.com/angellsan94/dogspots-api/internal/errors"
	"github.com/google/uuid"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) CreatePlaceReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized")
		return
	}
	placeID, err := uuid.Parse(r.PathValue("placeId"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid place identifier", "invalid_place_id")
		return
	}
	reason, details, ok := decodeRequest(w, r)
	if !ok {
		return
	}
	if err := h.repo.CreateForPlace(r.Context(), userID, placeID, reason, details); err != nil {
		writeCreateError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateCommentReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized")
		return
	}
	commentID, err := uuid.Parse(r.PathValue("commentId"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid comment identifier", "invalid_comment_id")
		return
	}
	reason, details, ok := decodeRequest(w, r)
	if !ok {
		return
	}
	if err := h.repo.CreateForComment(r.Context(), userID, commentID, reason, details); err != nil {
		writeCreateError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeRequest(w http.ResponseWriter, r *http.Request) (string, *string, bool) {
	var payload struct {
		Reason  string  `json:"reason"`
		Details *string `json:"details"`
	}
	if r.Body == nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", "A request body is required", "validation_error")
		return "", nil, false
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error")
		return "", nil, false
	}
	payload.Reason = strings.TrimSpace(payload.Reason)
	if !ValidReason(payload.Reason) {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", "reason must be one of spam, inappropriate, wrong_location, dangerous or other", "validation_error")
		return "", nil, false
	}
	var details *string
	if payload.Details != nil {
		trimmed := strings.TrimSpace(*payload.Details)
		if utf8.RuneCountInString(trimmed) > 500 {
			writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", "details cannot exceed 500 characters", "validation_error")
			return "", nil, false
		}
		if trimmed != "" {
			details = &trimmed
		}
	}
	return payload.Reason, details, true
}

func writeCreateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrPlaceNotFound):
		writeProblem(w, http.StatusNotFound, "Not Found", "Place not found", "place_not_found")
	case errors.Is(err, ErrCommentNotFound):
		writeProblem(w, http.StatusNotFound, "Not Found", "Comment not found", "comment_not_found")
	case errors.Is(err, ErrInvalidReason):
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", "reason must be one of spam, inappropriate, wrong_location, dangerous or other", "validation_error")
	default:
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error")
	}
}

func writeProblem(w http.ResponseWriter, status int, title, detail, code string) {
	apierrors.NewProblem(status, title, detail, code).Write(w)
}
