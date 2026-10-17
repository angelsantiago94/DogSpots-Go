package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/angellsan94/dogspots-api/internal/config"
	"github.com/angellsan94/dogspots-api/internal/errors"
	"github.com/angellsan94/dogspots-api/internal/middleware"
)

func main() {
	// 1. Cargar Configuración (B-1.2)
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error cargando configuración: %v\n", err)
		os.Exit(1)
	}

	// 2. Inicializar Logger (B-1.4)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("iniciando DogSpots API",
		slog.String("port", cfg.Port),
		slog.String("db_path", cfg.DBPath),
	)

	// 3. Configurar Router y Handlers (B-1.3)
	mux := http.NewServeMux()

	// Endpoint de salud
	mux.HandleFunc("GET /api/v1/healthz", healthzHandler)

	// 4. Aplicar Middlewares (B-1.4)
	// El orden es importante: Recovery -> Logger -> CORS -> Mux
	var handler http.Handler = mux
	handler = middleware.CORS(cfg.AllowedOrigins)(handler)
	handler = middleware.Logger(logger)(handler)
	handler = middleware.Recovery(logger)(handler)

	// 5. Configurar el Servidor HTTP
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 6. Ejecutar servidor en una goroutine para permitir el Graceful Shutdown
	go func() {
		logger.Info("servidor escuchando", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("error al iniciar el servidor", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// 7. Esperar señales de terminación para Graceful Shutdown (B-1.3)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	sig := <-stop
	logger.Info("recibida señal de terminación", slog.String("signal", sig.String()))

	// Crear un contexto con timeout para el apagado
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("error durante el apagado ordenado", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("servidor apagado correctamente")
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	// Responder con un JSON simple
	res := map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		// Si falla la codificación, usamos el helper de errores RFC 7807
		errors.NewProblem(http.StatusInternalServerError, "Internal Server Error", "Failed to encode health response", "encoding_error").Write(w)
	}
}
