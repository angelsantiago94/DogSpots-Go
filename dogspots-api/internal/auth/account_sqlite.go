package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DeleteAccountData borra los datos personales del usuario y anonimiza la cuenta.
// Devuelve las rutas de los ficheros que hay que eliminar del almacenamiento.
func (r *SQLiteRepository) DeleteAccountData(ctx context.Context, userID uuid.UUID) ([]string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin delete account: %w", err)
	}
	defer tx.Rollback()

	urls, err := collectAccountPhotoFiles(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM places WHERE author_id = ?`, userID.String()); err != nil {
		return nil, fmt.Errorf("delete account places: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM photos WHERE uploaded_by = ?`, userID.String()); err != nil {
		return nil, fmt.Errorf("delete account photos: %w", err)
	}

	ratedPlaces, err := collectRatingPlaces(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM ratings WHERE user_id = ?`, userID.String()); err != nil {
		return nil, fmt.Errorf("delete account ratings: %w", err)
	}
	for _, placeID := range ratedPlaces {
		if _, err := tx.ExecContext(ctx, `
			UPDATE places
			SET rating_avg = (SELECT AVG(rating) FROM ratings WHERE place_id = ?),
				rating_count = (SELECT COUNT(*) FROM ratings WHERE place_id = ?)
			WHERE id = ?
		`, placeID, placeID, placeID); err != nil {
			return nil, fmt.Errorf("refresh place rating stats: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE comments SET deleted = TRUE, body = NULL, edited = FALSE, updated_at = ?
		WHERE author_id = ? AND deleted = FALSE
	`, time.Now().UTC().Format(time.RFC3339Nano), userID.String()); err != nil {
		return nil, fmt.Errorf("anonymize account comments: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE places
		SET comment_count = (SELECT COUNT(*) FROM comments c WHERE c.place_id = places.id AND c.deleted = FALSE)
		WHERE id IN (SELECT DISTINCT place_id FROM comments WHERE author_id = ?)
	`, userID.String()); err != nil {
		return nil, fmt.Errorf("refresh place comment counts: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE comments
		SET reply_count = (SELECT COUNT(*) FROM comments c WHERE c.parent_id = comments.id AND c.deleted = FALSE)
		WHERE id IN (SELECT DISTINCT parent_id FROM comments WHERE author_id = ? AND parent_id IS NOT NULL)
	`, userID.String()); err != nil {
		return nil, fmt.Errorf("refresh comment reply counts: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM reports WHERE user_id = ?`, userID.String()); err != nil {
		return nil, fmt.Errorf("delete account reports: %w", err)
	}

	// La fila del usuario se conserva para no romper el historial de comentarios de terceros.
	if _, err := tx.ExecContext(ctx, `
		UPDATE users SET email = ?, name = 'Usuario eliminado', password_hash = '!', updated_at = ?
		WHERE id = ?
	`, deletedAccountEmail(userID), time.Now().UTC().Format(time.RFC3339Nano), userID.String()); err != nil {
		return nil, fmt.Errorf("anonymize account: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE refresh_tokens SET is_revoked = 1 WHERE user_id = ?`, userID.String()); err != nil {
		return nil, fmt.Errorf("revoke account sessions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit delete account: %w", err)
	}
	return urls, nil
}

func deletedAccountEmail(userID uuid.UUID) string {
	return "deleted+" + userID.String() + "@deleted.invalid"
}

func collectAccountPhotoFiles(ctx context.Context, tx *sql.Tx, userID uuid.UUID) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT url, thumb_url FROM photos
		WHERE uploaded_by = ? OR place_id IN (SELECT id FROM places WHERE author_id = ?)
	`, userID.String(), userID.String())
	if err != nil {
		return nil, fmt.Errorf("list account photo files: %w", err)
	}
	defer rows.Close()

	urls := make([]string, 0)
	for rows.Next() {
		var url, thumbURL string
		if err := rows.Scan(&url, &thumbURL); err != nil {
			return nil, fmt.Errorf("scan account photo file: %w", err)
		}
		urls = append(urls, url, thumbURL)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate account photo files: %w", err)
	}
	return urls, nil
}

func collectRatingPlaces(ctx context.Context, tx *sql.Tx, userID uuid.UUID) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT place_id FROM ratings WHERE user_id = ?`, userID.String())
	if err != nil {
		return nil, fmt.Errorf("list account rating places: %w", err)
	}
	defer rows.Close()

	places := make([]string, 0)
	for rows.Next() {
		var placeID string
		if err := rows.Scan(&placeID); err != nil {
			return nil, fmt.Errorf("scan account rating place: %w", err)
		}
		places = append(places, placeID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate account rating places: %w", err)
	}
	return places, nil
}
