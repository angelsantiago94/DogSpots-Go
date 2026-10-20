package ratings

import (
	"time"

	"github.com/google/uuid"
)

type Summary struct {
	RatingAvg   *float64 `json:"rating_avg"`
	RatingCount int      `json:"rating_count"`
	MyRating    *int     `json:"my_rating"`
}

type Rating struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	PlaceID   uuid.UUID `json:"place_id"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateRequest struct {
	Rating int `json:"rating"`
}
