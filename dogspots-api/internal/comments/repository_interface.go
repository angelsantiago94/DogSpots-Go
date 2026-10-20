package comments

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Comment CRUD
	Create(ctx context.Context, placeID, authorID uuid.UUID, body string, parentID *uuid.UUID) (Comment, error)
	CreateComment(ctx context.Context, placeID, authorID uuid.UUID, body string, parentID, replyTo *uuid.UUID) (Comment, error)
	GetByID(ctx context.Context, id uuid.UUID) (Comment, error)
	GetComment(ctx context.Context, id uuid.UUID) (Comment, error)
	ListByPlace(ctx context.Context, placeID uuid.UUID, sort string, limit int, cursor string) (CommentList, error)
	ListReplies(ctx context.Context, parentID uuid.UUID, limit int, cursor string) (CommentList, error)
	GetCommentsByPlace(ctx context.Context, placeID uuid.UUID) ([]CommentWithAuthor, error)
	GetReplies(ctx context.Context, parentID uuid.UUID) ([]CommentWithAuthor, error)
	Update(ctx context.Context, id, authorID uuid.UUID, body string) (Comment, error)
	UpdateComment(ctx context.Context, id, authorID uuid.UUID, body string) error
	Delete(ctx context.Context, id, authorID uuid.UUID) error
	DeleteCommentLogical(ctx context.Context, id, authorID uuid.UUID) error
	DeleteCommentPhysical(ctx context.Context, id uuid.UUID) error

	// Denormalization updates
	IncrementCommentCount(ctx context.Context, placeID uuid.UUID) error
	DecrementCommentCount(ctx context.Context, placeID uuid.UUID) error
	IncrementReplyCount(ctx context.Context, commentID uuid.UUID) error
	DecrementReplyCount(ctx context.Context, commentID uuid.UUID) error
}
