package middleware

import (
	"net/http"

	apierrors "github.com/angellsan94/dogspots-api/internal/errors"
)

const DefaultMaxBodySize int64 = 11 << 20

func MaxBodySize(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > limit {
				apierrors.NewProblem(http.StatusRequestEntityTooLarge, "Payload Too Large", "The request body exceeds the allowed size", "request_body_too_large").Write(w)
				return
			}

			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}
