package places

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/angellsan94/dogspots-api/internal/auth"
	apierrors "github.com/angellsan94/dogspots-api/internal/errors"
	"github.com/google/uuid"
)

type Handler struct {
	repo         Repository
	tokenManager *auth.TokenManager
}

func NewHandler(repo Repository, tokenManager *auth.TokenManager) *Handler {
	return &Handler{repo: repo, tokenManager: tokenManager}
}

func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.repo.ListTags(r.Context())
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": tags})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized")
		return
	}
	var request CreatePlaceRequest
	if err := decodeJSON(r, &request); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error")
		return
	}
	request.Title = strings.TrimSpace(request.Title)
	request.Description = strings.TrimSpace(request.Description)
	request.Category = strings.TrimSpace(request.Category)
	request.Tags = normalizeTags(request.Tags)
	if err := validateCreate(request); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error")
		return
	}
	if err := h.repo.ValidateTags(r.Context(), request.Tags); err != nil {
		writePlaceError(w, err)
		return
	}
	place, err := h.repo.Create(r.Context(), userID, request)
	if err != nil {
		writePlaceError(w, err)
		return
	}
	w.Header().Set("Location", "/api/v1/places/"+place.ID.String())
	writeJSON(w, http.StatusCreated, place)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "placeId")
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid place identifier", "invalid_place_id")
		return
	}
	place, err := h.repo.Get(r.Context(), id, h.optionalUserID(r))
	if err != nil {
		writePlaceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, place)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized")
		return
	}
	id, err := parsePathID(r, "placeId")
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid place identifier", "invalid_place_id")
		return
	}
	var request UpdatePlaceRequest
	if err := decodeJSON(r, &request); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error")
		return
	}
	if err := validateUpdate(&request); err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", err.Error(), "validation_error")
		return
	}
	if request.Tags != nil {
		*request.Tags = normalizeTags(*request.Tags)
		if len(*request.Tags) > 10 {
			writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", "At most 10 tags are allowed", "validation_error")
			return
		}
		if err := h.repo.ValidateTags(r.Context(), *request.Tags); err != nil {
			writePlaceError(w, err)
			return
		}
	}
	place, err := h.repo.Update(r.Context(), id, userID, request)
	if err != nil {
		writePlaceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, place)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized")
		return
	}
	id, err := parsePathID(r, "placeId")
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid place identifier", "invalid_place_id")
		return
	}
	if err := h.repo.Delete(r.Context(), id, userID); err != nil {
		writePlaceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	options, err := parseListOptions(r)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", err.Error(), "invalid_query")
		return
	}
	if err := h.repo.ValidateTags(r.Context(), options.Tags); err != nil {
		if errors.Is(err, ErrInvalidTags) {
			writeProblem(w, http.StatusBadRequest, "Bad Request", "One or more tags are not in the catalog", "invalid_tags")
		} else {
			writePlaceError(w, err)
		}
		return
	}
	result, err := h.repo.List(r.Context(), options)
	if err != nil {
		writePlaceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized")
		return
	}
	options, err := parsePagingOptions(r)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", err.Error(), "invalid_query")
		return
	}
	options.AuthorID = &userID
	result, err := h.repo.List(r.Context(), options)
	if err != nil {
		writePlaceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) optionalUserID(r *http.Request) *uuid.UUID {
	if h.tokenManager == nil {
		return nil
	}
	parts := strings.SplitN(strings.TrimSpace(r.Header.Get("Authorization")), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil
	}
	claims, err := h.tokenManager.ValidateToken(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil
	}
	return &claims.UserID
}

func parseListOptions(r *http.Request) (ListOptions, error) {
	options, err := parsePagingOptions(r)
	if err != nil {
		return ListOptions{}, err
	}
	query := r.URL.Query()
	var errParse error
	options.MinLat, errParse = optionalFloat(query.Get("min_lat"))
	if errParse != nil {
		return ListOptions{}, fmt.Errorf("min_lat must be a number")
	}
	options.MinLng, errParse = optionalFloat(query.Get("min_lng"))
	if errParse != nil {
		return ListOptions{}, fmt.Errorf("min_lng must be a number")
	}
	options.MaxLat, errParse = optionalFloat(query.Get("max_lat"))
	if errParse != nil {
		return ListOptions{}, fmt.Errorf("max_lat must be a number")
	}
	options.MaxLng, errParse = optionalFloat(query.Get("max_lng"))
	if errParse != nil {
		return ListOptions{}, fmt.Errorf("max_lng must be a number")
	}
	bboxCount := 0
	for _, value := range []*float64{options.MinLat, options.MinLng, options.MaxLat, options.MaxLng} {
		if value != nil {
			bboxCount++
		}
	}
	if bboxCount != 0 && bboxCount != 4 {
		return ListOptions{}, fmt.Errorf("bbox requires min_lat, min_lng, max_lat and max_lng")
	}
	if bboxCount == 4 {
		if *options.MinLat < -90 || *options.MaxLat > 90 || *options.MinLat > *options.MaxLat || *options.MinLng < -180 || *options.MinLng > 180 || *options.MaxLng < -180 || *options.MaxLng > 180 {
			return ListOptions{}, fmt.Errorf("bbox coordinates are out of range")
		}
	}
	options.Lat, errParse = optionalFloat(query.Get("lat"))
	if errParse != nil {
		return ListOptions{}, fmt.Errorf("lat must be a number")
	}
	options.Lng, errParse = optionalFloat(query.Get("lng"))
	if errParse != nil {
		return ListOptions{}, fmt.Errorf("lng must be a number")
	}
	if (options.Lat == nil) != (options.Lng == nil) {
		return ListOptions{}, fmt.Errorf("lat and lng must be provided together")
	}
	if options.Lat != nil && bboxCount != 0 {
		return ListOptions{}, fmt.Errorf("bbox and radius filters are mutually exclusive")
	}
	if options.Lat != nil && (*options.Lat < -90 || *options.Lat > 90 || *options.Lng < -180 || *options.Lng > 180) {
		return ListOptions{}, fmt.Errorf("coordinates are out of range")
	}
	options.RadiusM = 5000
	if value := query.Get("radius_m"); value != "" {
		options.RadiusM, errParse = strconv.Atoi(value)
		if errParse != nil || options.RadiusM < 1 || options.RadiusM > 50000 {
			return ListOptions{}, fmt.Errorf("radius_m must be between 1 and 50000")
		}
	}
	if options.Lat == nil && query.Has("radius_m") {
		return ListOptions{}, fmt.Errorf("radius_m requires lat and lng")
	}
	options.Category = query.Get("category")
	if options.Category != "" && !validCategory(options.Category) {
		return ListOptions{}, fmt.Errorf("category is not supported")
	}
	if tags := query.Get("tags"); tags != "" {
		options.Tags = normalizeTags(strings.Split(tags, ","))
	}
	if len(options.Tags) > 10 {
		return ListOptions{}, fmt.Errorf("at most 10 tags can be filtered")
	}
	if value := query.Get("min_rating"); value != "" {
		rating, parseErr := strconv.ParseFloat(value, 64)
		if parseErr != nil || rating < 1 || rating > 5 || math.IsNaN(rating) {
			return ListOptions{}, fmt.Errorf("min_rating must be between 1 and 5")
		}
		options.MinRating = &rating
	}
	options.Query = strings.TrimSpace(query.Get("q"))
	if utf8.RuneCountInString(options.Query) > 100 {
		return ListOptions{}, fmt.Errorf("q cannot exceed 100 characters")
	}
	if value := query.Get("sort"); value != "" {
		options.Sort = value
	}
	if options.Sort != "recent" && options.Sort != "rating" && options.Sort != "distance" {
		return ListOptions{}, fmt.Errorf("sort must be recent, rating or distance")
	}
	if options.Sort == "distance" && options.Lat == nil {
		return ListOptions{}, fmt.Errorf("sort=distance requires lat, lng and radius_m")
	}
	return options, nil
}

func parsePagingOptions(r *http.Request) (ListOptions, error) {
	options := ListOptions{Sort: "recent", Limit: 20, RadiusM: 5000}
	query := r.URL.Query()
	if value := query.Get("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 100 {
			return ListOptions{}, fmt.Errorf("limit must be between 1 and 100")
		}
		options.Limit = limit
	}
	options.Cursor = query.Get("cursor")
	return options, nil
}

func validateCreate(request CreatePlaceRequest) error {
	length := utf8.RuneCountInString(request.Title)
	if length < 3 || length > 100 {
		return fmt.Errorf("title must contain between 3 and 100 characters")
	}
	if utf8.RuneCountInString(request.Description) > 2000 {
		return fmt.Errorf("description cannot exceed 2000 characters")
	}
	if !validCoordinates(request.Lat, request.Lng) {
		return fmt.Errorf("coordinates are out of range")
	}
	if !validCategory(request.Category) {
		return fmt.Errorf("category is not supported")
	}
	if len(request.Tags) > 10 {
		return fmt.Errorf("at most 10 tags are allowed")
	}
	return nil
}

func validateUpdate(request *UpdatePlaceRequest) error {
	if request.Title == nil && request.Description == nil && request.Lat == nil && request.Lng == nil && request.Category == nil && request.Tags == nil {
		return fmt.Errorf("at least one field must be provided")
	}
	if request.Title != nil {
		*request.Title = strings.TrimSpace(*request.Title)
		length := utf8.RuneCountInString(*request.Title)
		if length < 3 || length > 100 {
			return fmt.Errorf("title must contain between 3 and 100 characters")
		}
	}
	if request.Category != nil {
		*request.Category = strings.TrimSpace(*request.Category)
	}
	if request.Description != nil && utf8.RuneCountInString(*request.Description) > 2000 {
		return fmt.Errorf("description cannot exceed 2000 characters")
	}
	if request.Lat != nil && (*request.Lat < -90 || *request.Lat > 90 || math.IsNaN(*request.Lat)) {
		return fmt.Errorf("lat must be between -90 and 90")
	}
	if request.Lng != nil && (*request.Lng < -180 || *request.Lng > 180 || math.IsNaN(*request.Lng)) {
		return fmt.Errorf("lng must be between -180 and 180")
	}
	if request.Category != nil && !validCategory(*request.Category) {
		return fmt.Errorf("category is not supported")
	}
	return nil
}

func validCoordinates(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180 && !math.IsNaN(lat) && !math.IsNaN(lng) && !math.IsInf(lat, 0) && !math.IsInf(lng, 0)
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}

func optionalFloat(value string) (*float64, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return nil, errors.New("invalid number")
	}
	return &parsed, nil
}

func parsePathID(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(name))
}

func decodeJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func writePlaceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeProblem(w, http.StatusNotFound, "Not Found", "Place not found", "place_not_found")
	case errors.Is(err, ErrForbidden):
		writeProblem(w, http.StatusForbidden, "Forbidden", "You do not own this place", "forbidden")
	case errors.Is(err, ErrInvalidTags):
		writeProblem(w, http.StatusUnprocessableEntity, "Validation Error", "One or more tags are not in the catalog", "invalid_tags")
	case errors.Is(err, ErrInvalidCursor):
		writeProblem(w, http.StatusBadRequest, "Bad Request", "The cursor is invalid for this query", "invalid_cursor")
	default:
		writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error")
	}
}

func writeProblem(w http.ResponseWriter, status int, title, detail, code string) {
	apierrors.NewProblem(status, title, detail, code).Write(w)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
