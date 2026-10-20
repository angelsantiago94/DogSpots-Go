package comments

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		Body             string  `json:"body"`
		ReplyToCommentID *string `json:"reply_to_comment_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error").Write(w)
		return
	}
	body := strings.TrimSpace(req.Body)
	if body == "" {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", "Comment body is required", "validation_error").Write(w)
		return
	}
	if utf8.RuneCountInString(body) > 1000 {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", "Comment body exceeds 1000 characters", "validation_error").Write(w)
		return
	}

	var parentID *uuid.UUID
	var replyTo *uuid.UUID
	if req.ReplyToCommentID != nil && strings.TrimSpace(*req.ReplyToCommentID) != "" {
		commentID, err := uuid.Parse(strings.TrimSpace(*req.ReplyToCommentID))
		if err != nil {
			apierrors.NewProblem(http.StatusBadRequest, "Bad Request", "Invalid comment identifier", "invalid_comment_id").Write(w)
			return
		}
		target, err := h.repo.GetByID(r.Context(), commentID)
		if err != nil {
			apierrors.NewProblem(http.StatusNotFound, "Not Found", "Comment not found", "comment_not_found").Write(w)
			return
		}
		if target.PlaceID != placeID {
			apierrors.NewProblem(http.StatusBadRequest, "Bad Request", "Comment does not belong to the place", "invalid_comment_place").Write(w)
			return
		}
		if target.ParentID != nil {
			parentID = target.ParentID
		} else {
			parentID = &target.ID
		}
		replyTo = &target.AuthorID
	}
	comment, err := h.repo.CreateComment(r.Context(), placeID, userID, body, parentID, replyTo)
	if err != nil {
		switch {
		case errors.Is(err, ErrPlaceNotFound), errors.Is(err, ErrCommentNotFound):
			apierrors.NewProblem(http.StatusNotFound, "Not Found", "Place or reply target not found", "not_found").Write(w)
		case strings.Contains(err.Error(), "does not belong to the place"):
			apierrors.NewProblem(http.StatusBadRequest, "Bad Request", err.Error(), "invalid_comment_reply").Write(w)
		default:
			apierrors.NewProblem(http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error").Write(w)
			return
		}
		return
	}
	writeJSON(w, http.StatusCreated, comment)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	placeID, err := parsePathID(r, "placeId")
	if err != nil {
		apierrors.NewProblem(http.StatusBadRequest, "Bad Request", "Invalid place identifier", "invalid_place_id").Write(w)
		return
	}
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "recent"
	}
	if sort != "recent" && sort != "oldest" {
		apierrors.NewProblem(http.StatusBadRequest, "Bad Request", "sort must be recent or oldest", "invalid_query").Write(w)
		return
	}
	limit, cursor, err := parseCommentPage(r)
	if err != nil {
		apierrors.NewProblem(http.StatusBadRequest, "Bad Request", err.Error(), "invalid_query").Write(w)
		return
	}
	items, err := h.repo.ListByPlace(r.Context(), placeID, sort, limit, cursor)
	if err != nil {
		if errors.Is(err, ErrPlaceNotFound) {
			apierrors.NewProblem(http.StatusNotFound, "Not Found", "Place not found", "place_not_found").Write(w)
			return
		}
		if strings.Contains(err.Error(), "cursor") || strings.Contains(err.Error(), "sort") {
			apierrors.NewProblem(http.StatusBadRequest, "Bad Request", err.Error(), "invalid_query").Write(w)
			return
		}
		apierrors.NewProblem(http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error").Write(w)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) Replies(w http.ResponseWriter, r *http.Request) {
	commentID, err := parsePathID(r, "commentId")
	if err != nil {
		apierrors.NewProblem(http.StatusBadRequest, "Bad Request", "Invalid comment identifier", "invalid_comment_id").Write(w)
		return
	}
	root, err := h.repo.GetByID(r.Context(), commentID)
	if err != nil {
		apierrors.NewProblem(http.StatusNotFound, "Not Found", "Comment not found", "comment_not_found").Write(w)
		return
	}
	if root.ParentID != nil {
		apierrors.NewProblem(http.StatusNotFound, "Not Found", "Replies are available for root comments only", "comment_not_found").Write(w)
		return
	}
	limit, cursor, err := parseCommentPage(r)
	if err != nil {
		apierrors.NewProblem(http.StatusBadRequest, "Bad Request", err.Error(), "invalid_query").Write(w)
		return
	}
	items, err := h.repo.ListReplies(r.Context(), commentID, limit, cursor)
	if err != nil {
		apierrors.NewProblem(http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error").Write(w)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized").Write(w)
		return
	}
	commentID, err := parsePathID(r, "commentId")
	if err != nil {
		apierrors.NewProblem(http.StatusBadRequest, "Bad Request", "Invalid comment identifier", "invalid_comment_id").Write(w)
		return
	}
	var req struct {
		Body string `json:"body"`
	}
	if err := decodeJSON(r, &req); err != nil {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error").Write(w)
		return
	}
	body := strings.TrimSpace(req.Body)
	if body == "" {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", "Comment body is required", "validation_error").Write(w)
		return
	}
	if utf8.RuneCountInString(body) > 1000 {
		apierrors.NewProblem(http.StatusUnprocessableEntity, "Validation Error", "Comment body exceeds 1000 characters", "validation_error").Write(w)
		return
	}
	comment, err := h.repo.Update(r.Context(), commentID, userID, body)
	if err != nil {
		if errors.Is(err, ErrCommentForbidden) {
			apierrors.NewProblem(http.StatusForbidden, "Forbidden", "You are not allowed to edit this comment", "forbidden").Write(w)
			return
		}
		if errors.Is(err, ErrCommentNotFound) {
			apierrors.NewProblem(http.StatusNotFound, "Not Found", "Comment not found", "comment_not_found").Write(w)
			return
		}
		apierrors.NewProblem(http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error").Write(w)
		return
	}
	writeJSON(w, http.StatusOK, comment)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized").Write(w)
		return
	}
	commentID, err := parsePathID(r, "commentId")
	if err != nil {
		apierrors.NewProblem(http.StatusBadRequest, "Bad Request", "Invalid comment identifier", "invalid_comment_id").Write(w)
		return
	}
	if err := h.repo.Delete(r.Context(), commentID, userID); err != nil {
		if errors.Is(err, ErrCommentForbidden) {
			apierrors.NewProblem(http.StatusForbidden, "Forbidden", "You are not allowed to delete this comment", "forbidden").Write(w)
			return
		}
		if errors.Is(err, ErrCommentNotFound) {
			apierrors.NewProblem(http.StatusNotFound, "Not Found", "Comment not found", "comment_not_found").Write(w)
			return
		}
		apierrors.NewProblem(http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error").Write(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseCommentPage(r *http.Request) (int, string, error) {
	limit := 20
	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			return 0, "", errors.New("limit must be between 1 and 100")
		}
		limit = parsed
	}
	return limit, r.URL.Query().Get("cursor"), nil
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
