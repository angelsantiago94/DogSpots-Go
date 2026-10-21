package reports

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrPlaceNotFound   = errors.New("place not found")
	ErrCommentNotFound = errors.New("comment not found")
	ErrInvalidReason   = errors.New("invalid report reason")
)

type Repository interface {
	CreateForPlace(ctx context.Context, userID, placeID uuid.UUID, reason string, details *string) error
	CreateForComment(ctx context.Context, userID, commentID uuid.UUID, reason string, details *string) error
	List(ctx context.Context, limit int) ([]ReportListItem, error)
}
