package auth

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/angellsan94/dogspots-api/db/migrations"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func newAuthTestRepository(t *testing.T) *SQLiteRepository {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if err := migrations.Up(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	repo, err := NewSQLiteRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	return repo
}

func TestRotateRefreshTokenAllowsOnlyOneRotation(t *testing.T) {
	repo := newAuthTestRepository(t)
	ctx := context.Background()
	now := time.Now().UTC()
	userID, familyID, oldID := uuid.New(), uuid.New(), uuid.New()
	if _, err := repo.CreateUser(ctx, UserCreateParams{
		ID: userID, Email: "rotator@example.test", PasswordHash: "hash", Name: "Rotator",
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateRefreshToken(ctx, RefreshTokenParams{
		ID: oldID, UserID: userID, TokenHash: "old", FamilyID: familyID,
		ExpiresAt: now.Add(time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	const attempts = 8
	results := make(chan bool, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rotated, err := repo.RotateRefreshToken(ctx, oldID, RefreshTokenParams{
				ID: uuid.New(), UserID: userID, TokenHash: fmt.Sprintf("new-%d", i), FamilyID: familyID,
				ExpiresAt: now.Add(time.Hour), CreatedAt: now,
			})
			if err != nil {
				t.Errorf("RotateRefreshToken() error = %v", err)
			}
			results <- rotated
		}(i)
	}
	wg.Wait()
	close(results)

	rotations := 0
	for rotated := range results {
		if rotated {
			rotations++
		}
	}
	if rotations != 1 {
		t.Fatalf("successful rotations = %d, want 1", rotations)
	}
}
