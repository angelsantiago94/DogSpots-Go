package reports

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/angellsan94/dogspots-api/db/migrations"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := migrations.Up(context.Background(), database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return database
}

func insertUser(t *testing.T, database *sql.DB, email string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	now := time.Now().UTC()
	if _, err := database.Exec(`INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?, ?, 'hash', 'User', ?, ?)`,
		id.String(), email, now, now); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func insertPlace(t *testing.T, database *sql.DB, authorID uuid.UUID, title string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	now := time.Now().UTC()
	if _, err := database.Exec(`INSERT INTO places (id, author_id, title, lat, lng, category, created_at, updated_at) VALUES (?, ?, ?, 40.4, -3.7, 'park', ?, ?)`,
		id.String(), authorID.String(), title, now, now); err != nil {
		t.Fatalf("seed place: %v", err)
	}
	return id
}

func TestCreateReportsIsIdempotent(t *testing.T) {
	database := newTestDB(t)
	repo := NewSQLiteRepository(database)
	author := insertUser(t, database, "author@example.com")
	reporter := insertUser(t, database, "reporter@example.com")
	placeID := insertPlace(t, database, author, "Parque del Oeste")

	details := "Hay cristales rotos"
	if err := repo.CreateForPlace(context.Background(), reporter, placeID, ReasonDangerous, &details); err != nil {
		t.Fatalf("create place report: %v", err)
	}
	if err := repo.CreateForPlace(context.Background(), reporter, placeID, ReasonOther, nil); err != nil {
		t.Fatalf("duplicate place report should be ignored: %v", err)
	}

	commentID := uuid.New()
	now := time.Now().UTC()
	if _, err := database.Exec(`INSERT INTO comments (id, place_id, author_id, body, created_at, updated_at) VALUES (?, ?, ?, 'Comentario', ?, ?)`,
		commentID.String(), placeID.String(), author.String(), now, now); err != nil {
		t.Fatalf("seed comment: %v", err)
	}
	if err := repo.CreateForComment(context.Background(), reporter, commentID, ReasonSpam, nil); err != nil {
		t.Fatalf("create comment report: %v", err)
	}
	if err := repo.CreateForComment(context.Background(), reporter, commentID, ReasonSpam, nil); err != nil {
		t.Fatalf("duplicate comment report should be ignored: %v", err)
	}

	var total int
	if err := database.QueryRow(`SELECT COUNT(*) FROM reports`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Fatalf("report count = %d, want 2 (one per target)", total)
	}

	items, err := repo.List(context.Background(), 10)
	if err != nil {
		t.Fatalf("list reports: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("listed %d reports, want 2", len(items))
	}
	placeReports, commentReports := 0, 0
	for _, item := range items {
		if item.ReporterEmail != "reporter@example.com" {
			t.Fatalf("reporter email = %q", item.ReporterEmail)
		}
		if item.PlaceID != nil {
			placeReports++
			if item.PlaceTitle != "Parque del Oeste" || item.Details == nil || *item.Details != details {
				t.Fatalf("unexpected place report: %#v", item)
			}
		}
		if item.CommentID != nil {
			commentReports++
			if item.CommentBody != "Comentario" {
				t.Fatalf("unexpected comment report: %#v", item)
			}
		}
	}
	if placeReports != 1 || commentReports != 1 {
		t.Fatalf("place reports = %d, comment reports = %d", placeReports, commentReports)
	}
}

func TestCreateReportValidatesTargetAndReason(t *testing.T) {
	database := newTestDB(t)
	repo := NewSQLiteRepository(database)
	userID := insertUser(t, database, "user@example.com")
	placeID := insertPlace(t, database, userID, "Parque")

	if err := repo.CreateForPlace(context.Background(), userID, uuid.New(), ReasonSpam, nil); !errors.Is(err, ErrPlaceNotFound) {
		t.Fatalf("missing place error = %v, want ErrPlaceNotFound", err)
	}
	if err := repo.CreateForComment(context.Background(), userID, uuid.New(), ReasonSpam, nil); !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("missing comment error = %v, want ErrCommentNotFound", err)
	}
	if err := repo.CreateForPlace(context.Background(), userID, placeID, "boring", nil); !errors.Is(err, ErrInvalidReason) {
		t.Fatalf("invalid reason error = %v, want ErrInvalidReason", err)
	}

	commentID := uuid.New()
	now := time.Now().UTC()
	if _, err := database.Exec(`INSERT INTO comments (id, place_id, author_id, body, deleted, created_at, updated_at) VALUES (?, ?, ?, NULL, TRUE, ?, ?)`,
		commentID.String(), placeID.String(), userID.String(), now, now); err != nil {
		t.Fatalf("seed deleted comment: %v", err)
	}
	if err := repo.CreateForComment(context.Background(), userID, commentID, ReasonSpam, nil); !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("deleted comment error = %v, want ErrCommentNotFound", err)
	}
}
