package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// Logger middleware añade un Request ID a la petición y loguea los detalles de la misma.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := uuid.New().String()
			w.Header().Set("X-Request-ID", requestID)

			// Añadir request ID al contexto para que otros componentes lo usen
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			// Registrar la petición entrante
			logger.Info("incoming request",
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)

			// Usar un wrapper para poder capturar el status code en el log de salida
			rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r.WithContext(ctx))

			// Loguear la finalización de la petición
			logger.Info("request completed",
				slog.String("request_id", requestID),
				slog.Int("status", rw.statusCode),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)
		})
	}
}

// responseWriterWrapper captura el status code de la respuesta
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(p []byte) (int, error) {
	return rw.ResponseWriter.Write(p)
}
