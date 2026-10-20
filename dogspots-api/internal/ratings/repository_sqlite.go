package ratings

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

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Upsert(ctx context.Context, placeID, userID uuid.UUID, value int) (Summary, error) {
	if value < 1 || value > 5 {
		return Summary{}, errors.New("rating must be between 1 and 5")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Summary{}, fmt.Errorf("begin upsert rating: %w", err)
	}
	defer tx.Rollback()
	if err := ensurePlaceExists(ctx, tx, placeID); err != nil {
		return Summary{}, err
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO ratings (id, user_id, place_id, rating, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id, place_id) DO UPDATE SET
			rating = EXCLUDED.rating,
			updated_at = EXCLUDED.updated_at
	`, uuid.NewString(), userID.String(), placeID.String(), value, now, now); err != nil {
		return Summary{}, fmt.Errorf("upsert rating: %w", err)
	}
	if err := syncPlaceStats(ctx, tx, placeID); err != nil {
		return Summary{}, err
	}
	summary, err := readSummary(ctx, tx, placeID, &userID)
	if err != nil {
		return Summary{}, err
	}
	if err := tx.Commit(); err != nil {
		return Summary{}, fmt.Errorf("commit upsert rating: %w", err)
	}
	return summary, nil
}

func (r *SQLiteRepository) Delete(ctx context.Context, placeID, userID uuid.UUID) (Summary, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Summary{}, fmt.Errorf("begin delete rating: %w", err)
	}
	defer tx.Rollback()
	if err := ensurePlaceExists(ctx, tx, placeID); err != nil {
		return Summary{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM ratings WHERE user_id = ? AND place_id = ?`, userID.String(), placeID.String()); err != nil {
		return Summary{}, fmt.Errorf("delete rating: %w", err)
	}
	if err := syncPlaceStats(ctx, tx, placeID); err != nil {
		return Summary{}, err
	}
	summary, err := readSummary(ctx, tx, placeID, &userID)
	if err != nil {
		return Summary{}, err
	}
	if err := tx.Commit(); err != nil {
		return Summary{}, fmt.Errorf("commit delete rating: %w", err)
	}
	return summary, nil
}

func (r *SQLiteRepository) Get(ctx context.Context, placeID, userID uuid.UUID) (Rating, error) {
	var item Rating
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, place_id, rating, created_at, updated_at FROM ratings WHERE user_id = ? AND place_id = ?`, userID.String(), placeID.String())
	var id, userIDStr, placeIDStr string
	var createdAt, updatedAt time.Time
	var err error
	if err = row.Scan(&id, &userIDStr, &placeIDStr, &item.Rating, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Rating{}, err
		}
		return Rating{}, fmt.Errorf("get rating: %w", err)
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return Rating{}, fmt.Errorf("parse rating id: %w", err)
	}
	item.ID = parsedID
	parsedUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		return Rating{}, fmt.Errorf("parse rating user id: %w", err)
	}
	item.UserID = parsedUserID
	parsedPlaceID, err := uuid.Parse(placeIDStr)
	if err != nil {
		return Rating{}, fmt.Errorf("parse rating place id: %w", err)
	}
	item.PlaceID = parsedPlaceID
	item.CreatedAt = createdAt
	item.UpdatedAt = updatedAt
	return item, nil
}

func (r *SQLiteRepository) Summary(ctx context.Context, placeID uuid.UUID, viewerID *uuid.UUID) (Summary, error) {
	if err := r.placeExists(ctx, placeID); err != nil {
		return Summary{}, err
	}
	return readSummary(ctx, r.db, placeID, viewerID)
}

type ratingQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func readSummary(ctx context.Context, queryer ratingQueryer, placeID uuid.UUID, viewerID *uuid.UUID) (Summary, error) {
	var avg sql.NullFloat64
	var count int
	var myRating sql.NullInt64
	var userArg any
	if viewerID == nil {
		userArg = nil
	} else {
		userArg = viewerID.String()
	}
	row := queryer.QueryRowContext(ctx, `
		SELECT
			(SELECT AVG(rating) FROM ratings WHERE place_id = ?),
			(SELECT COUNT(*) FROM ratings WHERE place_id = ?),
			(SELECT rating FROM ratings WHERE place_id = ? AND user_id = ?)
	`, placeID.String(), placeID.String(), placeID.String(), userArg)
	if err := row.Scan(&avg, &count, &myRating); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Summary{}, fmt.Errorf("missing place summary: %w", err)
		}
		return Summary{}, fmt.Errorf("scan summary: %w", err)
	}

	summary := Summary{RatingCount: count}
	if avg.Valid {
		v := avg.Float64
		summary.RatingAvg = &v
	}
	if myRating.Valid {
		v := int(myRating.Int64)
		summary.MyRating = &v
	}
	return summary, nil
}

func syncPlaceStats(ctx context.Context, tx *sql.Tx, placeID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE places
		SET rating_avg = (SELECT AVG(rating) FROM ratings WHERE place_id = ?),
			rating_count = (SELECT COUNT(*) FROM ratings WHERE place_id = ?)
		WHERE id = ?
	`, placeID.String(), placeID.String(), placeID.String())
	return err
}

func ensurePlaceExists(ctx context.Context, queryer ratingQueryer, placeID uuid.UUID) error {
	var exists int
	if err := queryer.QueryRowContext(ctx, `SELECT 1 FROM places WHERE id = ?`, placeID.String()).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("place not found")
		}
		return fmt.Errorf("check place: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) placeExists(ctx context.Context, placeID uuid.UUID) error {
	var exists int
	if err := r.db.QueryRowContext(ctx, `SELECT 1 FROM places WHERE id = ?`, placeID.String()).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("place not found")
		}
		return err
	}
	return nil
}
