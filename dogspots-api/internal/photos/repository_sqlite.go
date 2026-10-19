package photos

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/angellsan94/dogspots-api/db"
	"github.com/google/uuid"
)

// SQLiteRepository implementa la interfaz Repository usando sqlc.
type SQLiteRepository struct {
	database *sql.DB
	queries  *db.Queries
}

// NewSQLiteRepository crea una nueva instancia de SQLiteRepository.
func NewSQLiteRepository(database *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		database: database,
		queries:  db.New(database),
	}
}

// Create guarda una nueva foto asociada a un lugar.
func (r *SQLiteRepository) Create(ctx context.Context, photo Photo, maxPerPlace int) (Photo, error) {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return Photo{}, fmt.Errorf("begin photo transaction: %w", err)
	}
	defer tx.Rollback()
	queries := db.New(tx)
	count, err := queries.CountByPlaceID(ctx, photo.PlaceID)
	if err != nil {
		return Photo{}, fmt.Errorf("count photos by place id: %w", err)
	}
	if int(count) >= maxPerPlace {
		return Photo{}, ErrPhotoLimitReached
	}
	photo.Position = intPointer(int(count))
	if err := queries.CreatePhoto(ctx, createPhotoParams(photo)); err != nil {
		return Photo{}, fmt.Errorf("failed to create photo: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Photo{}, fmt.Errorf("commit photo transaction: %w", err)
	}
	photo.CreatedAt = time.Now().UTC()
	return photo, nil
}

func createPhotoParams(photo Photo) db.CreatePhotoParams {
	params := db.CreatePhotoParams{
		ID:         photo.ID,
		PlaceID:    photo.PlaceID,
		Url:        photo.URL,
		ThumbUrl:   photo.ThumbURL,
		UploadedBy: photo.UploadedBy,
	}

	// Mapeo de campos opcionales
	if photo.Caption != nil {
		params.Caption = sql.NullString{String: *photo.Caption, Valid: true}
	}
	if photo.Width != nil {
		params.Width = sql.NullInt64{Int64: int64(*photo.Width), Valid: true}
	}
	if photo.Height != nil {
		params.Height = sql.NullInt64{Int64: int64(*photo.Height), Valid: true}
	}
	if photo.Position != nil {
		params.Position = sql.NullInt64{Int64: int64(*photo.Position), Valid: true}
	}

	return params
}

func intPointer(value int) *int {
	return &value
}

func (r *SQLiteRepository) GetByID(ctx context.Context, photoID uuid.UUID) (Photo, error) {
	stored, err := r.queries.GetPhoto(ctx, photoID)
	if err != nil {
		if err == sql.ErrNoRows {
			return Photo{}, sql.ErrNoRows
		}
		return Photo{}, fmt.Errorf("get photo: %w", err)
	}
	return fromDBPhoto(stored)
}

func (r *SQLiteRepository) ListByPlaceID(ctx context.Context, placeID uuid.UUID) ([]Photo, error) {
	stored, err := r.queries.ListByPlaceID(ctx, placeID)
	if err != nil {
		return nil, fmt.Errorf("list photos by place id: %w", err)
	}
	photos := make([]Photo, 0, len(stored))
	for _, item := range stored {
		photo, err := fromDBPhoto(item)
		if err != nil {
			return nil, err
		}
		photos = append(photos, photo)
	}
	return photos, nil
}

func fromDBPhoto(stored db.Photo) (Photo, error) {
	photoID, err := uuid.Parse(fmt.Sprint(stored.ID))
	if err != nil {
		return Photo{}, fmt.Errorf("parse photo id: %w", err)
	}
	placeID, err := uuid.Parse(fmt.Sprint(stored.PlaceID))
	if err != nil {
		return Photo{}, fmt.Errorf("parse place id: %w", err)
	}
	uploadedBy, err := uuid.Parse(fmt.Sprint(stored.UploadedBy))
	if err != nil {
		return Photo{}, fmt.Errorf("parse uploader id: %w", err)
	}
	createdAt, err := parsePhotoTime(stored.CreatedAt)
	if err != nil {
		return Photo{}, err
	}
	photo := Photo{ID: photoID, PlaceID: placeID, URL: stored.Url, ThumbURL: stored.ThumbUrl, UploadedBy: uploadedBy, CreatedAt: createdAt}
	if stored.Caption.Valid {
		photo.Caption = &stored.Caption.String
	}
	if stored.Width.Valid {
		value := int(stored.Width.Int64)
		photo.Width = &value
	}
	if stored.Height.Valid {
		value := int(stored.Height.Int64)
		photo.Height = &value
	}
	if stored.Position.Valid {
		value := int(stored.Position.Int64)
		photo.Position = &value
	}
	return photo, nil
}

func parsePhotoTime(value any) (time.Time, error) {
	if parsed, ok := value.(time.Time); ok {
		return parsed.UTC(), nil
	}
	text := fmt.Sprint(value)
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02 15:04:05-07:00"} {
		if parsed, err := time.Parse(layout, text); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("parse photo timestamp %q", text)
}

// CountByPlaceID cuenta cuántas fotos tiene un lugar.
func (r *SQLiteRepository) CountByPlaceID(ctx context.Context, placeID uuid.UUID) (int, error) {
	count, err := r.queries.CountByPlaceID(ctx, placeID)
	if err != nil {
		return 0, fmt.Errorf("failed to count photos by place id: %w", err)
	}
	return int(count), nil
}

// DeleteByPlaceID elimina todas las fotos de un lugar.
func (r *SQLiteRepository) DeleteByPlaceID(ctx context.Context, placeID uuid.UUID) error {
	err := r.queries.DeleteByPlaceID(ctx, placeID)
	if err != nil {
		return fmt.Errorf("failed to delete photos by place id: %w", err)
	}
	return nil
}

// DeleteByID elimina una foto específica.
func (r *SQLiteRepository) DeleteByID(ctx context.Context, photoID uuid.UUID) error {
	err := r.queries.DeleteByID(ctx, photoID)
	if err != nil {
		return fmt.Errorf("failed to delete photo by id: %w", err)
	}
	return nil
}
