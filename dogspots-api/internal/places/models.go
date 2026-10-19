package places

import (
	"time"

	"github.com/google/uuid"
)

var categories = map[string]struct{}{
	"park": {}, "beach": {}, "trail": {}, "dog_park": {},
	"cafe": {}, "restaurant": {}, "other": {},
}

type Tag struct {
	Slug  string `json:"slug"`
	Label string `json:"label"`
}

type Photo struct {
	ID         uuid.UUID `json:"id"`
	URL        string    `json:"url"`
	ThumbURL   string    `json:"thumb_url"`
	Caption    *string   `json:"caption"`
	Width      *int      `json:"width"`
	Height     *int      `json:"height"`
	Position   *int      `json:"position"`
	UploadedBy uuid.UUID `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type PublicUser struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	PlacesCount int       `json:"places_count,omitempty"`
}

type Place struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description,omitempty"`
	Lat          float64    `json:"lat"`
	Lng          float64    `json:"lng"`
	Category     string     `json:"category"`
	Tags         []Tag      `json:"tags"`
	Photos       []Photo    `json:"photos"`
	RatingAvg    *float64   `json:"rating_avg"`
	RatingCount  int        `json:"rating_count"`
	MyRating     *int       `json:"my_rating"`
	CommentCount int        `json:"comment_count"`
	DistanceM    *int       `json:"distance_m,omitempty"`
	Author       PublicUser `json:"author"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type PlaceSummary struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Lat           float64   `json:"lat"`
	Lng           float64   `json:"lng"`
	Category      string    `json:"category"`
	CoverThumbURL *string   `json:"cover_thumb_url"`
	RatingAvg     *float64  `json:"rating_avg"`
	RatingCount   int       `json:"rating_count"`
	DistanceM     *int      `json:"distance_m,omitempty"`
	CreatedAt     time.Time `json:"-"`
}

type PlaceList struct {
	Items      []PlaceSummary `json:"items"`
	NextCursor *string        `json:"next_cursor"`
}

type CreatePlaceRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Lat         float64  `json:"lat"`
	Lng         float64  `json:"lng"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}

type UpdatePlaceRequest struct {
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Lat         *float64  `json:"lat"`
	Lng         *float64  `json:"lng"`
	Category    *string   `json:"category"`
	Tags        *[]string `json:"tags"`
}

type ListOptions struct {
	MinLat    *float64
	MinLng    *float64
	MaxLat    *float64
	MaxLng    *float64
	Lat       *float64
	Lng       *float64
	RadiusM   int
	Category  string
	Tags      []string
	MinRating *float64
	Query     string
	Sort      string
	Limit     int
	Cursor    string
	AuthorID  *uuid.UUID
}

func validCategory(category string) bool {
	_, ok := categories[category]
	return ok
}
