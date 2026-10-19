package photos

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrPhotoLimitReached = errors.New("photo limit reached")

// Repository define las operaciones de base de datos para las fotos.
type Repository interface {
	// Create guarda una nueva foto asociada a un lugar.
	Create(ctx context.Context, photo Photo, maxPerPlace int) (Photo, error)
	GetByID(ctx context.Context, photoID uuid.UUID) (Photo, error)
	ListByPlaceID(ctx context.Context, placeID uuid.UUID) ([]Photo, error)
	// CountByPlaceID cuenta cuántas fotos tiene un lugar.
	CountByPlaceID(ctx context.Context, placeID uuid.UUID) (int, error)
	// DeleteByPlaceID elimina todas las fotos de un lugar (usado al borrar un lugar).
	DeleteByPlaceID(ctx context.Context, placeID uuid.UUID) error
	// DeleteByID elimina una foto específica.
	DeleteByID(ctx context.Context, photoID uuid.UUID) error
}

// Photo representa el modelo de datos de una foto.
// Es similar al de internal/places pero encapsulado aquí para el dominio de fotos.
type Photo struct {
	ID         uuid.UUID `json:"id"`
	PlaceID    uuid.UUID `json:"place_id"`
	URL        string    `json:"url"`
	ThumbURL   string    `json:"thumb_url"`
	Caption    *string   `json:"caption"`
	Width      *int      `json:"width"`
	Height     *int      `json:"height"`
	Position   *int      `json:"position"`
	UploadedBy uuid.UUID `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}
