package config

import "testing"

func TestLoadReadsComposeStoragePaths(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("DB_PATH", "/data/db/dogspots.db")
	t.Setenv("MEDIA_DIR", "/data/media")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DBPath != "/data/db/dogspots.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/data/db/dogspots.db")
	}
	if cfg.MediaDir != "/data/media" {
		t.Errorf("MediaDir = %q, want %q", cfg.MediaDir, "/data/media")
	}
}
