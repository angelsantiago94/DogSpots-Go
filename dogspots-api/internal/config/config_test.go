package config

import (
	"os"
	"testing"
)

func TestLoadReadsComposeStoragePaths(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("DB_PATH", "/data/db/dogspots.db")
	t.Setenv("MEDIA_DIR", "/data/media")
	t.Setenv("ARGON2_MEMORY_KIB", "65536")
	t.Setenv("ARGON2_ITERATIONS", "3")
	t.Setenv("ARGON2_PARALLELISM", "2")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DBPath != "/data/db/dogspots.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/data/db/dogspots.db")
	}
	if cfg.AppEnv != "development" {
		t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, "development")
	}
	if cfg.MediaDir != "/data/media" {
		t.Errorf("MediaDir = %q, want %q", cfg.MediaDir, "/data/media")
	}
	if cfg.Argon2Params.Memory != 65536 || cfg.Argon2Params.Iterations != 3 || cfg.Argon2Params.Parallelism != 2 {
		t.Errorf("Argon2Params = %+v, want memory=65536 iterations=3 parallelism=2", cfg.Argon2Params)
	}
}

func TestLoadArgon2DefaultsAndValidation(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	for _, key := range []string{"ARGON2_MEMORY_KIB", "ARGON2_ITERATIONS", "ARGON2_PARALLELISM"} {
		value, exists := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if exists {
				_ = os.Setenv(key, value)
			} else {
				_ = os.Unsetenv(key)
			}
		})
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Argon2Params.Memory != 64*1024 || cfg.Argon2Params.Iterations != 3 || cfg.Argon2Params.Parallelism != 2 {
		t.Errorf("unexpected Argon2 defaults: %+v", cfg.Argon2Params)
	}
	t.Setenv("ARGON2_MEMORY_KIB", "not-a-number")
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject invalid Argon2 memory")
	}
}
