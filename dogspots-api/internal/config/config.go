package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/angellsan94/dogspots-api/internal/auth"
)

type Config struct {
	AppEnv          string
	Port            string
	DBPath          string
	JWTSecret       string
	MediaDir        string
	AllowedOrigins  []string
	TrustProxy      bool
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Argon2Params    auth.Argon2Params
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		DBPath:         getEnv("DB_PATH", "dogspots.db"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		MediaDir:       getEnv("MEDIA_DIR", "photos"),
		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "*"), ","),
	}
	accessTTL, err := getDurationEnv("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	refreshTTL, err := getDurationEnv("REFRESH_TOKEN_TTL", 720*time.Hour)
	if err != nil {
		return nil, err
	}
	trustProxy := strings.ToLower(os.Getenv("TRUST_PROXY"))
	cfg.TrustProxy = trustProxy == "true" || trustProxy == "1"
	cfg.AccessTokenTTL = accessTTL
	cfg.RefreshTokenTTL = refreshTTL

	argon2Memory, err := getUint32Env("ARGON2_MEMORY_KIB", 64*1024)
	if err != nil {
		return nil, err
	}
	argon2Iterations, err := getUint32Env("ARGON2_ITERATIONS", 3)
	if err != nil {
		return nil, err
	}
	argon2Parallelism, err := getUint32Env("ARGON2_PARALLELISM", 2)
	if err != nil {
		return nil, err
	}
	if argon2Memory < 8 || argon2Memory > 1024*1024 {
		return nil, fmt.Errorf("ARGON2_MEMORY_KIB must be between 8 and 1048576")
	}
	if argon2Iterations < 1 || argon2Iterations > 20 {
		return nil, fmt.Errorf("ARGON2_ITERATIONS must be between 1 and 20")
	}
	if argon2Parallelism < 1 || argon2Parallelism > 16 || argon2Memory < 8*argon2Parallelism {
		return nil, fmt.Errorf("ARGON2_PARALLELISM must be between 1 and 16 and no greater than ARGON2_MEMORY_KIB / 8")
	}
	cfg.Argon2Params = auth.Argon2Params{
		Memory:      argon2Memory,
		Iterations:  argon2Iterations,
		Parallelism: uint8(argon2Parallelism),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set")
	}

	return cfg, nil
}

func getDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration (for example 15m or 720h)", key)
	}
	return parsed, nil
}

func getUint32Env(key string, fallback uint32) (uint32, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return uint32(parsed), nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
