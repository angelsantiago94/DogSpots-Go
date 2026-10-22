package comments

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/angellsan94/dogspots-api/db"
	"github.com/angellsan94/dogspots-api/internal/places"
	"github.com/google/uuid"
)

type SQLiteRepository struct {
	database *sql.DB
	queries  *db.Queries
}

func NewSQLiteRepository(database *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		database: database,
		queries:  db.New(database),
	}
}

func (r *SQLiteRepository) Create(ctx context.Context, placeID, authorID uuid.UUID, body string, parentID *uuid.UUID) (Comment, error) {
	return r.CreateComment(ctx, placeID, authorID, body, parentID, nil)
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id uuid.UUID) (Comment, error) {
	return r.GetComment(ctx, id)
}

func (r *SQLiteRepository) ListByPlace(ctx context.Context, placeID uuid.UUID, sort string, limit int, cursor string) (CommentList, error) {
	var exists int
	if err := r.database.QueryRowContext(ctx, `SELECT 1 FROM places WHERE id = ?`, placeID.String()).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CommentList{}, ErrPlaceNotFound
		}
		return CommentList{}, fmt.Errorf("check place before listing comments: %w", err)
	}
	if sort == "" {
		sort = "recent"
	}
	if sort != "recent" && sort != "oldest" {
		return CommentList{}, errors.New("sort must be recent or oldest")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var after *commentCursor
	if cursor != "" {
		parsed, err := decodeCommentCursor(cursor, sort)
		if err != nil {
			return CommentList{}, err
		}
		after = &parsed
	}

	order := "DESC"
	comparison := "<"
	if sort == "oldest" {
		order = "ASC"
		comparison = ">"
	}
	query := `SELECT c.id, c.place_id, c.parent_id, c.reply_to, c.author_id, c.body,
		c.deleted, c.edited, c.created_at, c.updated_at, c.reply_count,
		u.name, COALESCE(reply_user.name, '')
		FROM comments c
		JOIN users u ON c.author_id = u.id
		LEFT JOIN users reply_user ON c.reply_to = reply_user.id
		WHERE c.place_id = ? AND c.parent_id IS NULL`
	args := []any{placeID.String()}
	if after != nil {
		query += ` AND (julianday(c.created_at) ` + comparison + ` julianday(?) OR
			(julianday(c.created_at) = julianday(?) AND c.id ` + comparison + ` ?))`
		args = append(args, after.CreatedAt, after.CreatedAt, after.ID)
	}
	query += ` ORDER BY julianday(c.created_at) ` + order + `, c.id ` + order + ` LIMIT ?`
	args = append(args, limit+1)
	rows, err := r.database.QueryContext(ctx, query, args...)
	if err != nil {
		return CommentList{}, fmt.Errorf("list comments by place: %w", err)
	}
	defer rows.Close()
	items := make([]Comment, 0, limit+1)
	for rows.Next() {
		var stored db.Comment
		var authorName, replyToName string
		if err := rows.Scan(&stored.ID, &stored.PlaceID, &stored.ParentID, &stored.ReplyTo,
			&stored.AuthorID, &stored.Body, &stored.Deleted, &stored.Edited, &stored.CreatedAt,
			&stored.UpdatedAt, &stored.ReplyCount, &authorName, &replyToName); err != nil {
			return CommentList{}, fmt.Errorf("scan comments by place: %w", err)
		}
		comment, err := fromDBComment(stored)
		if err != nil {
			return CommentList{}, err
		}
		if comment.Author != nil {
			comment.Author.DisplayName = authorName
		}
		if comment.ReplyTo != nil {
			comment.ReplyTo.DisplayName = replyToName
		}
		items = append(items, comment)
	}
	if err := rows.Err(); err != nil {
		return CommentList{}, fmt.Errorf("iterate comments by place: %w", err)
	}
	list := CommentList{Items: items}
	if len(items) > limit {
		last := items[limit-1]
		items = items[:limit]
		list.Items = items
		encoded, err := encodeCommentCursor(commentCursor{CreatedAt: last.CreatedAt.Format(time.RFC3339Nano), ID: last.ID.String(), Sort: sort})
		if err != nil {
			return CommentList{}, err
		}
		list.NextCursor = &encoded
	}
	return list, nil
}

func (r *SQLiteRepository) ListReplies(ctx context.Context, parentID uuid.UUID, limit int, cursor string) (CommentList, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var after *commentCursor
	if cursor != "" {
		parsed, err := decodeCommentCursor(cursor, "oldest")
		if err != nil {
			return CommentList{}, err
		}
		after = &parsed
	}
	query := `SELECT c.id, c.place_id, c.parent_id, c.reply_to, c.author_id, c.body,
		c.deleted, c.edited, c.created_at, c.updated_at, c.reply_count,
		u.name, COALESCE(reply_user.name, '')
		FROM comments c
		JOIN users u ON c.author_id = u.id
		LEFT JOIN users reply_user ON c.reply_to = reply_user.id
		WHERE c.parent_id = ?`
	args := []any{parentID.String()}
	if after != nil {
		query += ` AND (julianday(c.created_at) > julianday(?) OR
			(julianday(c.created_at) = julianday(?) AND c.id > ?))`
		args = append(args, after.CreatedAt, after.CreatedAt, after.ID)
	}
	query += ` ORDER BY julianday(c.created_at) ASC, c.id ASC LIMIT ?`
	args = append(args, limit+1)
	rows, err := r.database.QueryContext(ctx, query, args...)
	if err != nil {
		return CommentList{}, fmt.Errorf("list comment replies: %w", err)
	}
	defer rows.Close()
	items := make([]Comment, 0, limit+1)
	for rows.Next() {
		var stored db.Comment
		var authorName, replyToName string
		if err := rows.Scan(&stored.ID, &stored.PlaceID, &stored.ParentID, &stored.ReplyTo,
			&stored.AuthorID, &stored.Body, &stored.Deleted, &stored.Edited, &stored.CreatedAt,
			&stored.UpdatedAt, &stored.ReplyCount, &authorName, &replyToName); err != nil {
			return CommentList{}, fmt.Errorf("scan comment replies: %w", err)
		}
		comment, err := fromDBComment(stored)
		if err != nil {
			return CommentList{}, err
		}
		if comment.Author != nil {
			comment.Author.DisplayName = authorName
		}
		if comment.ReplyTo != nil {
			comment.ReplyTo.DisplayName = replyToName
		}
		items = append(items, comment)
	}
	if err := rows.Err(); err != nil {
		return CommentList{}, fmt.Errorf("iterate comment replies: %w", err)
	}
	list := CommentList{Items: items}
	if len(items) > limit {
		last := items[limit-1]
		list.Items = items[:limit]
		encoded, err := encodeCommentCursor(commentCursor{CreatedAt: last.CreatedAt.Format(time.RFC3339Nano), ID: last.ID.String(), Sort: "oldest"})
		if err != nil {
			return CommentList{}, err
		}
		list.NextCursor = &encoded
	}
	return list, nil
}

type commentCursor struct {
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
	Sort      string `json:"sort"`
}

func encodeCommentCursor(cursor commentCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("encode comment cursor: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeCommentCursor(value, sort string) (commentCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return commentCursor{}, errors.New("invalid comment cursor")
	}
	var cursor commentCursor
	if err := json.Unmarshal(payload, &cursor); err != nil || cursor.Sort != sort || cursor.ID == "" || cursor.CreatedAt == "" {
		return commentCursor{}, errors.New("invalid comment cursor")
	}
	if _, err := uuid.Parse(cursor.ID); err != nil {
		return commentCursor{}, errors.New("invalid comment cursor")
	}
	if _, err := time.Parse(time.RFC3339Nano, cursor.CreatedAt); err != nil {
		return commentCursor{}, errors.New("invalid comment cursor")
	}
	return cursor, nil
}

func (r *SQLiteRepository) Update(ctx context.Context, id, authorID uuid.UUID, body string) (Comment, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return Comment{}, errors.New("comment body cannot be empty")
	}
	if utf8.RuneCountInString(trimmed) > 2000 {
		return Comment{}, errors.New("comment body cannot exceed 2000 characters")
	}
	result, err := r.database.ExecContext(ctx, `UPDATE comments SET body = ?, edited = TRUE, updated_at = ? WHERE id = ? AND author_id = ? AND deleted = FALSE`,
		trimmed, time.Now().UTC(), id.String(), authorID.String())
	if err != nil {
		return Comment{}, fmt.Errorf("update comment: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return Comment{}, fmt.Errorf("check updated comment: %w", err)
	}
	if updated == 0 {
		stored, getErr := r.GetComment(ctx, id)
		if errors.Is(getErr, sql.ErrNoRows) {
			return Comment{}, ErrCommentNotFound
		}
		if getErr != nil {
			return Comment{}, getErr
		}
		if stored.AuthorID != authorID || stored.Deleted {
			return Comment{}, ErrCommentForbidden
		}
		return Comment{}, ErrCommentNotFound
	}
	return r.GetComment(ctx, id)
}

func (r *SQLiteRepository) Delete(ctx context.Context, id, authorID uuid.UUID) error {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete comment: %w", err)
	}
	defer tx.Rollback()
	queries := db.New(tx)
	comment, err := queries.GetComment(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCommentNotFound
	}
	if err != nil {
		return fmt.Errorf("load comment before delete: %w", err)
	}
	storedAuthor, err := toUUID(comment.AuthorID)
	if err != nil {
		return fmt.Errorf("parse comment author: %w", err)
	}
	if storedAuthor != authorID {
		return ErrCommentForbidden
	}
	if comment.Deleted {
		// Borrar dos veces no debe volver a decrementar los contadores.
		return ErrCommentNotFound
	}
	var replyCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM comments WHERE parent_id = ?`, id.String()).Scan(&replyCount); err != nil {
		return fmt.Errorf("count comment replies: %w", err)
	}
	if replyCount > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE comments SET deleted = TRUE, body = NULL, updated_at = ? WHERE id = ?`, time.Now().UTC(), id.String()); err != nil {
			return fmt.Errorf("logically delete comment: %w", err)
		}
	} else {
		if err := queries.DeleteCommentPhysical(ctx, id); err != nil {
			return fmt.Errorf("physically delete comment: %w", err)
		}
	}
	parentID, err := toUUIDPtr(comment.ParentID)
	if err != nil {
		return fmt.Errorf("parse deleted comment parent: %w", err)
	}
	if parentID != nil {
		if err := queries.DecrementReplyCount(ctx, *parentID); err != nil {
			return fmt.Errorf("decrement parent reply count: %w", err)
		}
	}
	placeID, err := toUUID(comment.PlaceID)
	if err != nil {
		return fmt.Errorf("parse deleted comment place: %w", err)
	}
	if err := queries.DecrementCommentCount(ctx, placeID); err != nil {
		return fmt.Errorf("decrement place comment count: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete comment: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) CreateComment(ctx context.Context, placeID, authorID uuid.UUID, body string, parentID, replyTo *uuid.UUID) (Comment, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return Comment{}, errors.New("comment body cannot be empty")
	}
	if utf8.RuneCountInString(trimmed) > 2000 {
		return Comment{}, errors.New("comment body cannot exceed 2000 characters")
	}

	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return Comment{}, fmt.Errorf("begin create comment: %w", err)
	}
	defer tx.Rollback()
	queries := db.New(tx)
	var placeExists int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM places WHERE id = ?`, placeID.String()).Scan(&placeExists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Comment{}, ErrPlaceNotFound
		}
		return Comment{}, fmt.Errorf("check place before creating comment: %w", err)
	}
	effectiveParentID := parentID
	effectiveReplyTo := replyTo
	if effectiveParentID != nil {
		parentComment, err := queries.GetComment(ctx, effectiveParentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Comment{}, ErrCommentNotFound
			}
			return Comment{}, fmt.Errorf("load parent comment: %w", err)
		}
		parentPlaceID, err := toUUID(parentComment.PlaceID)
		if err != nil {
			return Comment{}, fmt.Errorf("parse parent place id: %w", err)
		}
		if parentPlaceID != placeID {
			return Comment{}, errors.New("parent comment does not belong to the place")
		}
		parentParentID, err := toUUIDPtr(parentComment.ParentID)
		if err != nil {
			return Comment{}, fmt.Errorf("parse parent parent id: %w", err)
		}
		if parentParentID != nil {
			effectiveParentID = parentParentID
		} else {
			effectiveParentID = parentID
		}
		if effectiveReplyTo == nil {
			repliedTo, err := toUUID(parentComment.AuthorID)
			if err != nil {
				return Comment{}, fmt.Errorf("parse parent author: %w", err)
			}
			effectiveReplyTo = &repliedTo
		}
	}

	commentID := uuid.New()
	now := time.Now().UTC()
	created, err := queries.CreateComment(ctx, db.CreateCommentParams{
		ID:        commentID,
		PlaceID:   placeID,
		AuthorID:  authorID,
		Body:      sql.NullString{String: trimmed, Valid: true},
		ParentID:  toInterface(effectiveParentID),
		ReplyTo:   toInterface(effectiveReplyTo),
		Edited:    false,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return Comment{}, fmt.Errorf("create comment: %w", err)
	}
	if effectiveParentID != nil {
		if err := queries.IncrementReplyCount(ctx, effectiveParentID); err != nil {
			return Comment{}, fmt.Errorf("increment reply count: %w", err)
		}
	}
	if err := queries.IncrementCommentCount(ctx, placeID); err != nil {
		return Comment{}, fmt.Errorf("increment comment count: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Comment{}, fmt.Errorf("commit create comment: %w", err)
	}

	return fromDBComment(created)
}

func (r *SQLiteRepository) GetComment(ctx context.Context, id uuid.UUID) (Comment, error) {
	stored, err := r.queries.GetComment(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Comment{}, sql.ErrNoRows
		}
		return Comment{}, fmt.Errorf("get comment: %w", err)
	}
	return fromDBComment(stored)
}

func (r *SQLiteRepository) GetCommentsByPlace(ctx context.Context, placeID uuid.UUID) ([]CommentWithAuthor, error) {
	rows, err := r.queries.GetCommentsByPlace(ctx, placeID)
	if err != nil {
		return nil, fmt.Errorf("list comments by place: %w", err)
	}
	items := make([]CommentWithAuthor, 0, len(rows))
	for _, row := range rows {
		comment, err := fromDBCommentRow(row)
		if err != nil {
			return nil, err
		}
		items = append(items, CommentWithAuthor{Comment: comment, AuthorName: row.AuthorName, AuthorEmail: row.AuthorEmail})
	}
	return items, nil
}

func (r *SQLiteRepository) GetReplies(ctx context.Context, parentID uuid.UUID) ([]CommentWithAuthor, error) {
	rows, err := r.queries.GetReplies(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("list replies: %w", err)
	}
	items := make([]CommentWithAuthor, 0, len(rows))
	for _, row := range rows {
		comment, err := fromDBCommentRow(row)
		if err != nil {
			return nil, err
		}
		items = append(items, CommentWithAuthor{Comment: comment, AuthorName: row.AuthorName, AuthorEmail: row.AuthorEmail})
	}
	return items, nil
}

func (r *SQLiteRepository) UpdateComment(ctx context.Context, id, authorID uuid.UUID, body string) error {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return errors.New("comment body cannot be empty")
	}
	if utf8.RuneCountInString(trimmed) > 2000 {
		return errors.New("comment body cannot exceed 2000 characters")
	}
	updatedAt := time.Now().UTC()
	if err := r.queries.UpdateComment(ctx, db.UpdateCommentParams{
		Body:      sql.NullString{String: trimmed, Valid: true},
		UpdatedAt: updatedAt,
		ID:        id,
		AuthorID:  authorID,
	}); err != nil {
		return fmt.Errorf("update comment: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) DeleteCommentLogical(ctx context.Context, id, authorID uuid.UUID) error {
	comment, err := r.queries.GetComment(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return fmt.Errorf("load comment before logical delete: %w", err)
	}
	if comment.AuthorID == nil {
		return errors.New("comment author missing")
	}
	if authorID != mustUUID(comment.AuthorID) {
		return errors.New("comment does not belong to the author")
	}
	updatedAt := time.Now().UTC()
	if _, err := r.database.ExecContext(ctx, `UPDATE comments SET deleted = TRUE, body = NULL, updated_at = ? WHERE id = ? AND author_id = ?`, updatedAt, id, authorID); err != nil {
		return fmt.Errorf("logical delete comment: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) DeleteCommentPhysical(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.DeleteCommentPhysical(ctx, id); err != nil {
		return fmt.Errorf("physical delete comment: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) IncrementCommentCount(ctx context.Context, placeID uuid.UUID) error {
	return r.queries.IncrementCommentCount(ctx, placeID)
}

func (r *SQLiteRepository) DecrementCommentCount(ctx context.Context, placeID uuid.UUID) error {
	return r.queries.DecrementCommentCount(ctx, placeID)
}

func (r *SQLiteRepository) IncrementReplyCount(ctx context.Context, commentID uuid.UUID) error {
	return r.queries.IncrementReplyCount(ctx, commentID)
}

func (r *SQLiteRepository) DecrementReplyCount(ctx context.Context, commentID uuid.UUID) error {
	return r.queries.DecrementReplyCount(ctx, commentID)
}

func fromDBComment(stored db.Comment) (Comment, error) {
	id, err := toUUID(stored.ID)
	if err != nil {
		return Comment{}, fmt.Errorf("parse comment id: %w", err)
	}
	placeID, err := toUUID(stored.PlaceID)
	if err != nil {
		return Comment{}, fmt.Errorf("parse comment place id: %w", err)
	}
	authorID, err := toUUID(stored.AuthorID)
	if err != nil {
		return Comment{}, fmt.Errorf("parse comment author id: %w", err)
	}
	parentID, err := toUUIDPtr(stored.ParentID)
	if err != nil {
		return Comment{}, fmt.Errorf("parse parent id: %w", err)
	}
	replyTo, err := toUUIDPtr(stored.ReplyTo)
	if err != nil {
		return Comment{}, fmt.Errorf("parse reply_to: %w", err)
	}
	createdAt, err := toTime(stored.CreatedAt)
	if err != nil {
		return Comment{}, fmt.Errorf("parse created_at: %w", err)
	}
	updatedAt, err := toTime(stored.UpdatedAt)
	if err != nil {
		return Comment{}, fmt.Errorf("parse updated_at: %w", err)
	}
	replyCount := 0
	if stored.ReplyCount.Valid {
		replyCount = int(stored.ReplyCount.Int64)
	}
	var body *string
	if stored.Body.Valid {
		value := stored.Body.String
		body = &value
	}
	var publicAuthor *places.PublicUser
	if !stored.Deleted {
		publicAuthor = &places.PublicUser{ID: authorID}
	}
	var publicReply *places.PublicUser
	if replyTo != nil {
		replyID := *replyTo
		publicReply = &places.PublicUser{ID: replyID}
	}
	return Comment{
		ID:         id,
		PlaceID:    placeID,
		AuthorID:   authorID,
		Author:     publicAuthor,
		Body:       body,
		ParentID:   parentID,
		ReplyTo:    publicReply,
		Edited:     stored.Edited,
		Deleted:    stored.Deleted,
		ReplyCount: replyCount,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func fromDBCommentRow(row interface{}) (Comment, error) {
	switch value := row.(type) {
	case db.GetCommentsByPlaceRow:
		comment, err := fromDBComment(db.Comment{
			ID:         value.ID,
			PlaceID:    value.PlaceID,
			ParentID:   value.ParentID,
			ReplyTo:    value.ReplyTo,
			AuthorID:   value.AuthorID,
			Body:       value.Body,
			Deleted:    value.Deleted,
			Edited:     value.Edited,
			CreatedAt:  value.CreatedAt,
			UpdatedAt:  value.UpdatedAt,
			ReplyCount: value.ReplyCount,
		})
		if err != nil {
			return Comment{}, err
		}
		if comment.Author != nil {
			comment.Author.DisplayName = value.AuthorName
		}
		if comment.ReplyTo != nil {
			comment.ReplyTo.DisplayName = ""
		}
		return comment, nil
	case db.GetRepliesRow:
		comment, err := fromDBComment(db.Comment{
			ID:         value.ID,
			PlaceID:    value.PlaceID,
			ParentID:   value.ParentID,
			ReplyTo:    value.ReplyTo,
			AuthorID:   value.AuthorID,
			Body:       value.Body,
			Deleted:    value.Deleted,
			Edited:     value.Edited,
			CreatedAt:  value.CreatedAt,
			UpdatedAt:  value.UpdatedAt,
			ReplyCount: value.ReplyCount,
		})
		if err != nil {
			return Comment{}, err
		}
		if comment.Author != nil {
			comment.Author.DisplayName = value.AuthorName
		}
		if comment.ReplyTo != nil {
			comment.ReplyTo.DisplayName = ""
		}
		return comment, nil
	default:
		return Comment{}, fmt.Errorf("unsupported comment row type %T", row)
	}
}

func toUUID(value any) (uuid.UUID, error) {
	if value == nil {
		return uuid.Nil, errors.New("null uuid")
	}
	switch typed := value.(type) {
	case uuid.UUID:
		return typed, nil
	case string:
		return uuid.Parse(typed)
	case []byte:
		return uuid.ParseBytes(typed)
	default:
		return uuid.Nil, fmt.Errorf("unsupported uuid value type %T", value)
	}
}

func toUUIDPtr(value any) (*uuid.UUID, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := toUUID(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func toTime(value any) (time.Time, error) {
	if value == nil {
		return time.Time{}, errors.New("null time")
	}
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC(), nil
	case string:
		if parsed, err := time.Parse(time.RFC3339Nano, typed); err == nil {
			return parsed.UTC(), nil
		}
		if parsed, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", typed); err == nil {
			return parsed.UTC(), nil
		}
		if parsed, err := time.Parse("2006-01-02 15:04:05 -0700 MST", typed); err == nil {
			return parsed.UTC(), nil
		}
		if parsed, err := time.Parse("2006-01-02 15:04:05", typed); err == nil {
			return parsed.UTC(), nil
		}
		return time.Time{}, fmt.Errorf("unsupported time value %q", typed)
	case []byte:
		return toTime(string(typed))
	default:
		return time.Time{}, fmt.Errorf("unsupported time value type %T", value)
	}
}

func mustUUID(value any) uuid.UUID {
	u, err := toUUID(value)
	if err != nil {
		panic(err)
	}
	return u
}

func toInterface(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}
