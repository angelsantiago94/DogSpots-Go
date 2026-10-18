package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/angellsan94/dogspots-api/internal/auth"
	"github.com/angellsan94/dogspots-api/internal/config"
	apierrors "github.com/angellsan94/dogspots-api/internal/errors"
	"github.com/angellsan94/dogspots-api/internal/middleware"
	_ "modernc.org/sqlite"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error cargando configuración: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("iniciando DogSpots API",
		slog.String("port", cfg.Port),
		slog.String("db_path", cfg.DBPath),
	)

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		logger.Error("error al abrir la base de datos", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		logger.Error("error al conectar con la base de datos", slog.Any("error", err))
		os.Exit(1)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	repo, err := auth.NewSQLiteRepository(db)
	if err != nil {
		logger.Error("error al inicializar el repositorio de auth", slog.Any("error", err))
		os.Exit(1)
	}

	tokenManager := auth.NewTokenManager(cfg.JWTSecret, 15*time.Minute)
	authService := auth.NewAuthService(repo, auth.NewArgon2Hasher(), tokenManager, 15*time.Minute, 7*24*time.Hour)
	handler := auth.NewHandler(authService, repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/healthz", healthzHandler)
	mux.HandleFunc("POST /api/v1/auth/register", handler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", handler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", handler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", handler.Logout)
	mux.Handle("GET /api/v1/users/me", auth.RequireAuth(tokenManager, http.HandlerFunc(handler.GetMe)))
	mux.Handle("PATCH /api/v1/users/me", auth.RequireAuth(tokenManager, http.HandlerFunc(handler.UpdateMe)))
	mux.HandleFunc("GET /api/v1/users/{userId}", handler.GetUserByID)

	var app http.Handler = mux
	app = middleware.CORS(cfg.AllowedOrigins)(app)
	app = middleware.Logger(logger)(app)
	app = middleware.Recovery(logger)(app)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      app,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("servidor escuchando", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("error al iniciar el servidor", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("error durante el apagado ordenado", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("servidor apagado correctamente")
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	res := map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		apierrors.NewProblem(http.StatusInternalServerError, "Internal Server Error", "Failed to encode health response", "encoding_error").Write(w)
	}
}
