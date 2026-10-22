package photos

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/angellsan94/dogspots-api/internal/auth"
	apierrors "github.com/angellsan94/dogspots-api/internal/errors"
	"github.com/angellsan94/dogspots-api/internal/photos/processor"
	"github.com/angellsan94/dogspots-api/internal/places"
	"github.com/angellsan94/dogspots-api/internal/storage"
	"github.com/google/uuid"
)

const (
	maxPhotoSize  = 10 << 20
	maxPhotos     = 30
	maxCaptionLen = 200
)

type placeReader interface {
	Get(context.Context, uuid.UUID, *uuid.UUID) (places.Place, error)
}

type Handler struct {
	photos  Repository
	places  placeReader
	storage storage.Storage
}

func NewHandler(photoRepo Repository, placeRepo placeReader, fileStorage storage.Storage) *Handler {
	return &Handler{photos: photoRepo, places: placeRepo, storage: fileStorage}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writePhotoProblem(w, http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized")
		return
	}
	placeID, err := parsePhotoID(r, "placeId")
	if err != nil {
		writePhotoProblem(w, http.StatusBadRequest, "Bad Request", "Invalid place identifier", "invalid_place_id")
		return
	}
	if _, err := h.places.Get(r.Context(), placeID, nil); err != nil {
		writePhotoPlaceError(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPhotoSize+(1<<20))
	if err := r.ParseMultipartForm(maxPhotoSize); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writePhotoProblem(w, http.StatusRequestEntityTooLarge, "Payload Too Large", "The uploaded file exceeds 10 MB", "file_too_large")
		} else {
			writePhotoProblem(w, http.StatusUnprocessableEntity, "Validation Error", "A multipart file field named 'file' is required", "validation_error")
		}
		return
	}
	defer r.MultipartForm.RemoveAll()
	file, header, err := r.FormFile("file")
	if err != nil {
		writePhotoProblem(w, http.StatusUnprocessableEntity, "Validation Error", "A multipart file field named 'file' is required", "validation_error")
		return
	}
	defer file.Close()
	caption := strings.TrimSpace(r.FormValue("caption"))
	if utf8.RuneCountInString(caption) > maxCaptionLen {
		writePhotoProblem(w, http.StatusUnprocessableEntity, "Validation Error", "Caption cannot exceed 200 characters", "validation_error")
		return
	}
	original, thumbnail, info, err := processor.ProcessImage(file, maxPhotoSize)
	if err != nil {
		if errors.Is(err, processor.ErrFileTooLarge) {
			writePhotoProblem(w, http.StatusRequestEntityTooLarge, "Payload Too Large", "The uploaded file exceeds 10 MB", "file_too_large")
		} else if errors.Is(err, processor.ErrUnsupportedFormat) {
			writePhotoProblem(w, http.StatusUnsupportedMediaType, "Unsupported Media Type", "Only JPEG, PNG and WebP images are supported", "unsupported_media_type")
		} else {
			writePhotoProblem(w, http.StatusUnprocessableEntity, "Validation Error", "The image could not be processed", "invalid_image")
		}
		return
	}
	if header.Size > maxPhotoSize {
		writePhotoProblem(w, http.StatusRequestEntityTooLarge, "Payload Too Large", "The uploaded file exceeds 10 MB", "file_too_large")
		return
	}

	photoID := uuid.New()
	extension := ".jpg"
	url, err := h.storage.Save(r.Context(), photoID.String()+extension, bytes.NewReader(original))
	if err != nil {
		writePhotoProblem(w, http.StatusInternalServerError, "Internal Server Error", "Failed to store the uploaded image", "storage_error")
		return
	}
	thumbURL, err := h.storage.Save(r.Context(), photoID.String()+"-thumb"+extension, bytes.NewReader(thumbnail))
	if err != nil {
		_ = h.storage.Delete(context.Background(), url)
		writePhotoProblem(w, http.StatusInternalServerError, "Internal Server Error", "Failed to store the image thumbnail", "storage_error")
		return
	}
	photo := Photo{ID: photoID, PlaceID: placeID, URL: url, ThumbURL: thumbURL, UploadedBy: userID, Width: &info.Width, Height: &info.Height}
	if caption != "" {
		photo.Caption = &caption
	}
	photo, err = h.photos.Create(r.Context(), photo, maxPhotos)
	if err != nil {
		_ = h.storage.Delete(context.Background(), url)
		_ = h.storage.Delete(context.Background(), thumbURL)
		if errors.Is(err, ErrPhotoLimitReached) {
			writePhotoProblem(w, http.StatusUnprocessableEntity, "Validation Error", "A place can have at most 30 photos", "photo_limit_reached")
		} else {
			writePhotoProblem(w, http.StatusInternalServerError, "Internal Server Error", "Failed to save photo metadata", "database_error")
		}
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/api/v1/places/%s/photos/%s", placeID, photoID))
	writePhotoJSON(w, http.StatusCreated, photo)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writePhotoProblem(w, http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized")
		return
	}
	placeID, err := parsePhotoID(r, "placeId")
	if err != nil {
		writePhotoProblem(w, http.StatusBadRequest, "Bad Request", "Invalid place identifier", "invalid_place_id")
		return
	}
	photoID, err := parsePhotoID(r, "photoId")
	if err != nil {
		writePhotoProblem(w, http.StatusBadRequest, "Bad Request", "Invalid photo identifier", "invalid_photo_id")
		return
	}
	photo, err := h.photos.GetByID(r.Context(), photoID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && photo.PlaceID != placeID) {
		writePhotoProblem(w, http.StatusNotFound, "Not Found", "Photo not found", "photo_not_found")
		return
	}
	if err != nil {
		writePhotoProblem(w, http.StatusInternalServerError, "Internal Server Error", "Failed to load photo", "database_error")
		return
	}
	place, err := h.places.Get(r.Context(), placeID, nil)
	if err != nil {
		writePhotoPlaceError(w, err)
		return
	}
	if photo.UploadedBy != userID && place.Author.ID != userID {
		writePhotoProblem(w, http.StatusForbidden, "Forbidden", "Only the uploader or place owner can delete this photo", "forbidden")
		return
	}
	if err := h.photos.DeleteByID(r.Context(), photoID); err != nil {
		writePhotoProblem(w, http.StatusInternalServerError, "Internal Server Error", "Failed to delete photo metadata", "database_error")
		return
	}
	if err := h.storage.Delete(r.Context(), photo.URL); err != nil {
		writePhotoProblem(w, http.StatusInternalServerError, "Internal Server Error", "Photo was removed but its file could not be deleted", "storage_error")
		return
	}
	if err := h.storage.Delete(r.Context(), photo.ThumbURL); err != nil {
		writePhotoProblem(w, http.StatusInternalServerError, "Internal Server Error", "Photo was removed but its thumbnail could not be deleted", "storage_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PreparePlaceFilesCleanup(ctx context.Context, placeID uuid.UUID) (func(context.Context) error, error) {
	photos, err := h.photos.ListByPlaceID(ctx, placeID)
	if err != nil {
		return nil, err
	}
	return func(cleanupContext context.Context) error {
		for _, photo := range photos {
			if err := h.storage.Delete(cleanupContext, photo.URL); err != nil {
				return err
			}
			if err := h.storage.Delete(cleanupContext, photo.ThumbURL); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

func parsePhotoID(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(name))
}

func writePhotoJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writePhotoProblem(w http.ResponseWriter, status int, title, detail, code string) {
	apierrors.NewProblem(status, title, detail, code).Write(w)
}

func writePhotoPlaceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, places.ErrNotFound):
		writePhotoProblem(w, http.StatusNotFound, "Not Found", "Place not found", "place_not_found")
	case errors.Is(err, places.ErrForbidden):
		writePhotoProblem(w, http.StatusForbidden, "Forbidden", "The place belongs to another user", "forbidden")
	default:
		writePhotoProblem(w, http.StatusInternalServerError, "Internal Server Error", "Failed to load place", "internal_error")
	}
}
