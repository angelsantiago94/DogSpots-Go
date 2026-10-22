package config

import (
	"os"
	"testing"
	"time"
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

func TestLoadTrustProxy(t *testing.T) {
	cases := map[string]bool{
		"true":  true,
		"1":     true,
		"":      false,
		"yes":   false,
		"false": false,
	}
	for value, want := range cases {
		t.Run("value="+value, func(t *testing.T) {
			t.Setenv("JWT_SECRET", "test-secret")
			t.Setenv("TRUST_PROXY", value)
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.TrustProxy != want {
				t.Errorf("TrustProxy = %v, want %v for TRUST_PROXY=%q", cfg.TrustProxy, want, value)
			}
		})
	}

	t.Run("unset", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "test-secret")
		value, exists := os.LookupEnv("TRUST_PROXY")
		if err := os.Unsetenv("TRUST_PROXY"); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if exists {
				_ = os.Setenv("TRUST_PROXY", value)
			} else {
				_ = os.Unsetenv("TRUST_PROXY")
			}
		})
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.TrustProxy {
			t.Error("TrustProxy = true, want false when TRUST_PROXY is unset")
		}
	})
}

func TestLoadTokenTTLDefaults(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	for _, key := range []string{"ACCESS_TOKEN_TTL", "REFRESH_TOKEN_TTL"} {
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
	if cfg.AccessTokenTTL != 15*time.Minute || cfg.RefreshTokenTTL != 720*time.Hour {
		t.Errorf("token TTLs = %s/%s, want 15m/720h", cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	}
}

func TestLoadTokenTTLsFromEnv(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ACCESS_TOKEN_TTL", "30m")
	t.Setenv("REFRESH_TOKEN_TTL", "48h")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AccessTokenTTL != 30*time.Minute || cfg.RefreshTokenTTL != 48*time.Hour {
		t.Errorf("token TTLs = %s/%s, want 30m/48h", cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	}

	t.Setenv("REFRESH_TOKEN_TTL", "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject invalid REFRESH_TOKEN_TTL")
	}
}
