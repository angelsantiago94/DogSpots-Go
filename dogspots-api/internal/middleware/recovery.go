package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/angellsan94/dogspots-api/internal/errors"
)

// Recovery captura los pánicos para evitar que el servidor se caiga
// y devuelve una respuesta de error estandarizada.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Recuperar el stack trace para debugging
					stack := debug.Stack()

					logger.Error("panic recovered",
						slog.Any("error", err),
						slog.String("stack", string(stack)),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					)

					// Devolver error RFC 7807
					problem := errors.NewProblem(
						http.StatusInternalServerError,
						"Internal Server Error",
						"An unexpected error occurred. The server team has been notified.",
						"panic_recovered",
					)
					problem.Write(w)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
