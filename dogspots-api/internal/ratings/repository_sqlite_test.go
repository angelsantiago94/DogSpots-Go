package ratings

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func TestUpsertAndDeleteRating(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE users (
			id UUID PRIMARY KEY,
			email TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE places (
			id UUID PRIMARY KEY,
			author_id UUID NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			lat REAL NOT NULL,
			lng REAL NOT NULL,
			category TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			comment_count INT DEFAULT 0,
			rating_avg REAL DEFAULT 0,
			rating_count INT DEFAULT 0
		);
		CREATE TABLE ratings (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			place_id UUID NOT NULL,
			rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (user_id, place_id)
		);
	`); err != nil {
		t.Fatalf("schema setup: %v", err)
	}

	userID := uuid.New()
	placeID := uuid.New()
	if _, err := db.Exec(`INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`, userID.String(), "user@example.com", "hash", "User", time.Now().UTC(), time.Now().UTC()); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO places (id, author_id, title, description, lat, lng, category, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, placeID.String(), userID.String(), "Park", "Desc", 40.0, -3.0, "park", time.Now().UTC(), time.Now().UTC()); err != nil {
		t.Fatalf("seed place: %v", err)
	}

	repo := NewSQLiteRepository(db)

	summary, err := repo.Upsert(context.Background(), placeID, userID, 5)
	if err != nil {
		t.Fatalf("upsert rating: %v", err)
	}
	if summary.RatingCount != 1 || summary.RatingAvg == nil || *summary.RatingAvg != 5 || summary.MyRating == nil || *summary.MyRating != 5 {
		t.Fatalf("unexpected first summary: %#v", summary)
	}
	var storedAverage sql.NullFloat64
	var storedCount int
	if err := db.QueryRow(`SELECT rating_avg, rating_count FROM places WHERE id = ?`, placeID.String()).Scan(&storedAverage, &storedCount); err != nil {
		t.Fatal(err)
	}
	if !storedAverage.Valid || storedAverage.Float64 != 5 || storedCount != 1 {
		t.Fatalf("unexpected persisted first aggregates: avg=%#v count=%d", storedAverage, storedCount)
	}

	summary, err = repo.Upsert(context.Background(), placeID, userID, 3)
	if err != nil {
		t.Fatalf("replace rating: %v", err)
	}
	if summary.RatingCount != 1 || summary.RatingAvg == nil || *summary.RatingAvg != 3 || summary.MyRating == nil || *summary.MyRating != 3 {
		t.Fatalf("unexpected updated summary: %#v", summary)
	}
	if err := db.QueryRow(`SELECT rating_avg, rating_count FROM places WHERE id = ?`, placeID.String()).Scan(&storedAverage, &storedCount); err != nil {
		t.Fatal(err)
	}
	if !storedAverage.Valid || storedAverage.Float64 != 3 || storedCount != 1 {
		t.Fatalf("unexpected persisted updated aggregates: avg=%#v count=%d", storedAverage, storedCount)
	}

	summary, err = repo.Delete(context.Background(), placeID, userID)
	if err != nil {
		t.Fatalf("delete rating: %v", err)
	}
	if summary.RatingCount != 0 || summary.RatingAvg != nil || summary.MyRating != nil {
		t.Fatalf("unexpected deleted summary: %#v", summary)
	}
	if err := db.QueryRow(`SELECT rating_avg, rating_count FROM places WHERE id = ?`, placeID.String()).Scan(&storedAverage, &storedCount); err != nil {
		t.Fatal(err)
	}
	if storedAverage.Valid || storedCount != 0 {
		t.Fatalf("unexpected persisted deleted aggregates: avg=%#v count=%d", storedAverage, storedCount)
	}
}
