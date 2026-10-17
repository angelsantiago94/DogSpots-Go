package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port           string
	DBPath         string
	JWTSecret      string
	MediaDir       string
	AllowedOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		DBPath:         getEnv("DB_PATH", "dogspots.db"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		MediaDir:       getEnv("MEDIA_DIR", "photos"),
		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "*"), ","),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
