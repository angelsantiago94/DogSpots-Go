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

	dbmigrations "github.com/angellsan94/dogspots-api/db/migrations"
	"github.com/angellsan94/dogspots-api/internal/auth"
	"github.com/angellsan94/dogspots-api/internal/comments"
	"github.com/angellsan94/dogspots-api/internal/config"
	apierrors "github.com/angellsan94/dogspots-api/internal/errors"
	"github.com/angellsan94/dogspots-api/internal/middleware"
	"github.com/angellsan94/dogspots-api/internal/photos"
	"github.com/angellsan94/dogspots-api/internal/places"
	"github.com/angellsan94/dogspots-api/internal/ratings"
	"github.com/angellsan94/dogspots-api/internal/reports"
	"github.com/angellsan94/dogspots-api/internal/storage"
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
	if err := dbmigrations.Up(context.Background(), db); err != nil {
		logger.Error("error al ejecutar migraciones", slog.Any("error", err))
		os.Exit(1)
	}
	if cfg.AppEnv == "development" {
		if err := places.SeedDevelopmentData(context.Background(), db); err != nil {
			logger.Error("error al cargar datos de desarrollo", slog.Any("error", err))
			os.Exit(1)
		}
	}

	repo, err := auth.NewSQLiteRepository(db)
	if err != nil {
		logger.Error("error al inicializar el repositorio de auth", slog.Any("error", err))
		os.Exit(1)
	}

	tokenManager := auth.NewTokenManager(cfg.JWTSecret, 15*time.Minute)
	authService := auth.NewAuthService(repo, auth.NewArgon2HasherWithParams(cfg.Argon2Params), tokenManager, 15*time.Minute, 7*24*time.Hour)
	handler := auth.NewHandler(authService, repo)
	placesRepo, err := places.NewSQLiteRepository(db)
	if err != nil {
		logger.Error("error al inicializar el repositorio de lugares", slog.Any("error", err))
		os.Exit(1)
	}
	placesHandler := places.NewHandler(placesRepo, tokenManager)
	photoRepo := photos.NewSQLiteRepository(db)
	fileStorage, err := storage.NewLocalStorage(cfg.MediaDir)
	if err != nil {
		logger.Error("error al inicializar el almacenamiento de fotos", slog.Any("error", err))
		os.Exit(1)
	}
	photosHandler := photos.NewHandler(photoRepo, placesRepo, fileStorage)
	placesHandler.SetPhotoCleanup(photosHandler.PreparePlaceFilesCleanup)
	commentsRepo := comments.NewSQLiteRepository(db)
	commentsHandler := comments.NewHandler(commentsRepo)
	authRateLimit := middleware.NewRateLimiter(10, time.Minute)
	commentCreateRateLimit := middleware.NewRateLimiter(5, time.Minute)
	placeCreateRateLimit := middleware.NewRateLimiter(10, time.Minute)
	photoUploadRateLimit := middleware.NewRateLimiter(20, time.Minute)
	ratingsRepo := ratings.NewSQLiteRepository(db)
	ratingsHandler := ratings.NewHandler(ratingsRepo)
	reportsRepo := reports.NewSQLiteRepository(db)
	reportsHandler := reports.NewHandler(reportsRepo)
	accountHandler := auth.NewAccountHandler(repo, fileStorage)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/healthz", healthzHandler)
	mux.Handle("POST /api/v1/auth/register", authRateLimit(http.HandlerFunc(handler.Register)))
	mux.Handle("POST /api/v1/auth/login", authRateLimit(http.HandlerFunc(handler.Login)))
	mux.Handle("POST /api/v1/auth/refresh", authRateLimit(http.HandlerFunc(handler.Refresh)))
	mux.HandleFunc("POST /api/v1/auth/logout", handler.Logout)
	mux.Handle("GET /api/v1/users/me", auth.RequireAuth(tokenManager, http.HandlerFunc(handler.GetMe)))
	mux.Handle("PATCH /api/v1/users/me", auth.RequireAuth(tokenManager, http.HandlerFunc(handler.UpdateMe)))
	mux.Handle("DELETE /api/v1/users/me", auth.RequireAuth(tokenManager, http.HandlerFunc(accountHandler.DeleteMe)))
	mux.HandleFunc("GET /api/v1/users/{userId}", handler.GetUserByID)
	mux.HandleFunc("GET /api/v1/tags", placesHandler.ListTags)
	mux.HandleFunc("GET /api/v1/places", placesHandler.List)
	mux.Handle("POST /api/v1/places", auth.RequireAuth(tokenManager, placeCreateRateLimit(http.HandlerFunc(placesHandler.Create))))
	mux.HandleFunc("GET /api/v1/places/{placeId}", placesHandler.Get)
	mux.Handle("PATCH /api/v1/places/{placeId}", auth.RequireAuth(tokenManager, http.HandlerFunc(placesHandler.Update)))
	mux.Handle("DELETE /api/v1/places/{placeId}", auth.RequireAuth(tokenManager, http.HandlerFunc(placesHandler.Delete)))
	mux.Handle("POST /api/v1/places/{placeId}/photos", auth.RequireAuth(tokenManager, photoUploadRateLimit(http.HandlerFunc(photosHandler.Upload))))
	mux.Handle("DELETE /api/v1/places/{placeId}/photos/{photoId}", auth.RequireAuth(tokenManager, http.HandlerFunc(photosHandler.Delete)))
	mux.Handle("POST /api/v1/places/{placeId}/comments", auth.RequireAuth(tokenManager, commentCreateRateLimit(http.HandlerFunc(commentsHandler.Create))))
	mux.HandleFunc("GET /api/v1/places/{placeId}/comments", commentsHandler.List)
	mux.HandleFunc("GET /api/v1/comments/{commentId}/replies", commentsHandler.Replies)
	mux.Handle("PATCH /api/v1/comments/{commentId}", auth.RequireAuth(tokenManager, http.HandlerFunc(commentsHandler.Update)))
	mux.Handle("DELETE /api/v1/comments/{commentId}", auth.RequireAuth(tokenManager, http.HandlerFunc(commentsHandler.Delete)))
	mux.Handle("PUT /api/v1/places/{placeId}/rating", auth.RequireAuth(tokenManager, http.HandlerFunc(ratingsHandler.Put)))
	mux.Handle("DELETE /api/v1/places/{placeId}/rating", auth.RequireAuth(tokenManager, http.HandlerFunc(ratingsHandler.Delete)))
	mux.Handle("POST /api/v1/places/{placeId}/reports", auth.RequireAuth(tokenManager, http.HandlerFunc(reportsHandler.CreatePlaceReport)))
	mux.Handle("POST /api/v1/comments/{commentId}/reports", auth.RequireAuth(tokenManager, http.HandlerFunc(reportsHandler.CreateCommentReport)))
	mux.Handle("GET /api/v1/users/me/places", auth.RequireAuth(tokenManager, http.HandlerFunc(placesHandler.ListMine)))
	mux.Handle("GET /media/{filename}", storage.MediaHandler(cfg.MediaDir))

	var app http.Handler = mux
	app = middleware.MaxBodySize(middleware.DefaultMaxBodySize)(app)
	app = middleware.CORS(cfg.AllowedOrigins)(app)
	app = middleware.Logger(logger)(app)
	app = middleware.Recovery(logger)(app)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      app,
		ReadTimeout:  60 * time.Second,
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
