package reports

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(database *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: database}
}

func (r *SQLiteRepository) CreateForPlace(ctx context.Context, userID, placeID uuid.UUID, reason string, details *string) error {
	var exists int
	if err := r.db.QueryRowContext(ctx, `SELECT 1 FROM places WHERE id = ?`, placeID.String()).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPlaceNotFound
		}
		return fmt.Errorf("check reported place: %w", err)
	}
	return r.create(ctx, userID, &placeID, nil, reason, details)
}

func (r *SQLiteRepository) CreateForComment(ctx context.Context, userID, commentID uuid.UUID, reason string, details *string) error {
	var deleted bool
	if err := r.db.QueryRowContext(ctx, `SELECT deleted FROM comments WHERE id = ?`, commentID.String()).Scan(&deleted); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCommentNotFound
		}
		return fmt.Errorf("check reported comment: %w", err)
	}
	if deleted {
		return ErrCommentNotFound
	}
	return r.create(ctx, userID, nil, &commentID, reason, details)
}

func (r *SQLiteRepository) create(ctx context.Context, userID uuid.UUID, placeID, commentID *uuid.UUID, reason string, details *string) error {
	if !ValidReason(reason) {
		return ErrInvalidReason
	}
	var placeArg, commentArg, detailsArg any
	if placeID != nil {
		placeArg = placeID.String()
	}
	if commentID != nil {
		commentArg = commentID.String()
	}
	if details != nil {
		detailsArg = *details
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO reports (id, user_id, place_id, comment_id, reason, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING
	`, uuid.NewString(), userID.String(), placeArg, commentArg, reason, detailsArg, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) List(ctx context.Context, limit int) ([]ReportListItem, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.user_id, r.place_id, r.comment_id, r.reason, r.details, r.created_at,
			u.email, COALESCE(p.title, ''), COALESCE(c.body, '')
		FROM reports r
		JOIN users u ON u.id = r.user_id
		LEFT JOIN places p ON p.id = r.place_id
		LEFT JOIN comments c ON c.id = r.comment_id
		ORDER BY julianday(r.created_at) DESC, r.id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	items := make([]ReportListItem, 0)
	for rows.Next() {
		var item ReportListItem
		var id, userID, reason, email, placeTitle, commentBody string
		var placeID, commentID, details sql.NullString
		var createdAt any
		if err := rows.Scan(&id, &userID, &placeID, &commentID, &reason, &details, &createdAt, &email, &placeTitle, &commentBody); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		if item.ID, err = uuid.Parse(id); err != nil {
			return nil, fmt.Errorf("parse report id: %w", err)
		}
		if item.UserID, err = uuid.Parse(userID); err != nil {
			return nil, fmt.Errorf("parse report user id: %w", err)
		}
		if placeID.Valid {
			parsed, err := uuid.Parse(placeID.String)
			if err != nil {
				return nil, fmt.Errorf("parse report place id: %w", err)
			}
			item.PlaceID = &parsed
		}
		if commentID.Valid {
			parsed, err := uuid.Parse(commentID.String)
			if err != nil {
				return nil, fmt.Errorf("parse report comment id: %w", err)
			}
			item.CommentID = &parsed
		}
		if item.CreatedAt, err = parseReportTime(createdAt); err != nil {
			return nil, err
		}
		item.Reason = reason
		if details.Valid {
			item.Details = &details.String
		}
		item.ReporterEmail = email
		item.PlaceTitle = placeTitle
		item.CommentBody = commentBody
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reports: %w", err)
	}
	return items, nil
}

func parseReportTime(value any) (time.Time, error) {
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC(), nil
	case string:
		for _, layout := range []string{
			time.RFC3339Nano,
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05-07:00",
			"2006-01-02 15:04:05.999999999 -0700 MST",
			"2006-01-02 15:04:05 -0700 MST",
			"2006-01-02 15:04:05",
		} {
			if parsed, err := time.Parse(layout, typed); err == nil {
				return parsed.UTC(), nil
			}
		}
		return time.Time{}, fmt.Errorf("unsupported report time %q", typed)
	case []byte:
		return parseReportTime(string(typed))
	default:
		return time.Time{}, fmt.Errorf("unsupported report time type %T", value)
	}
}
