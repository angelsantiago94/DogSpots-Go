package migrations

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestUpAdoptsExistingSQLiteDatabase(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)

	if _, err := database.Exec(`CREATE TABLE tags (
		id TEXT PRIMARY KEY,
		slug TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO tags (id, slug, name, created_at) VALUES ('existing', 'agua', 'Agua anterior', CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := Up(ctx, database); err != nil {
		t.Fatalf("first Up() error = %v", err)
	}
	if err := Up(ctx, database); err != nil {
		t.Fatalf("second Up() error = %v", err)
	}

	var tagName string
	if err := database.QueryRow(`SELECT name FROM tags WHERE slug = 'agua'`).Scan(&tagName); err != nil {
		t.Fatal(err)
	}
	if tagName != "Agua anterior" {
		t.Fatalf("existing tag name = %q, want preserved value", tagName)
	}
	var tableCount int
	if err := database.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'reports'`).Scan(&tableCount); err != nil {
		t.Fatal(err)
	}
	if tableCount != 1 {
		t.Fatal("reports table was not created by migrations")
	}
}
