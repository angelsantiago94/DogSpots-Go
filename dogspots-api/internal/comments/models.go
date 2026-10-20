package comments

import (
	"errors"
	"time"

	"github.com/angellsan94/dogspots-api/internal/places"
	"github.com/google/uuid"
)

var (
	ErrCommentNotFound  = errors.New("comment not found")
	ErrCommentForbidden = errors.New("comment belongs to another user")
	ErrPlaceNotFound    = errors.New("place not found")
)

type Comment struct {
	ID         uuid.UUID          `json:"id"`
	PlaceID    uuid.UUID          `json:"place_id"`
	AuthorID   uuid.UUID          `json:"-"`
	Author     *places.PublicUser `json:"author"`
	Body       *string            `json:"body"`
	ParentID   *uuid.UUID         `json:"parent_id"`
	ReplyTo    *places.PublicUser `json:"reply_to"`
	Edited     bool               `json:"edited"`
	Deleted    bool               `json:"deleted"`
	ReplyCount int                `json:"reply_count"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type CommentWithAuthor struct {
	Comment
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
}

type CommentList struct {
	Items      []Comment `json:"items"`
	NextCursor *string   `json:"next_cursor"`
}
