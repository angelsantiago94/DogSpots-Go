package photos

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

func TestSQLiteRepositoryEnforcesPhotoLimitAndSupportsDelete(t *testing.T) {
	ctx := context.Background()
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if err := migrations.Up(ctx, database); err != nil {
		t.Fatal(err)
	}
	userID, placeID := uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := database.Exec(`INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		userID.String(), "photo-owner@example.test", "hash", "Photo Owner", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO places (id, author_id, title, lat, lng, category) VALUES (?, ?, ?, ?, ?, ?)`,
		placeID.String(), userID.String(), "Photo test", 0, 0, "park"); err != nil {
		t.Fatal(err)
	}
	repository := NewSQLiteRepository(database)
	for index := 0; index < 2; index++ {
		photo := Photo{ID: uuid.New(), PlaceID: placeID, URL: "/media/photo.jpg", ThumbURL: "/media/thumb.jpg", UploadedBy: userID}
		created, err := repository.Create(ctx, photo, 2)
		if err != nil {
			t.Fatalf("Create() #%d error = %v", index+1, err)
		}
		if created.Position == nil || *created.Position != index {
			t.Fatalf("position = %v, want %d", created.Position, index)
		}
		if _, err := repository.GetByID(ctx, photo.ID); err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
	}
	if _, err := repository.Create(ctx, Photo{ID: uuid.New(), PlaceID: placeID, URL: "/media/extra.jpg", ThumbURL: "/media/extra-thumb.jpg", UploadedBy: userID}, 2); !errors.Is(err, ErrPhotoLimitReached) {
		t.Fatalf("Create() over limit error = %v", err)
	}
	photos, err := repository.ListByPlaceID(ctx, placeID)
	if err != nil || len(photos) != 2 {
		t.Fatalf("ListByPlaceID() returned %d photos, error = %v", len(photos), err)
	}
	if err := repository.DeleteByID(ctx, photos[0].ID); err != nil {
		t.Fatalf("DeleteByID() error = %v", err)
	}
	count, err := repository.CountByPlaceID(ctx, placeID)
	if err != nil || count != 1 {
		t.Fatalf("CountByPlaceID() = %d, error = %v", count, err)
	}
}
