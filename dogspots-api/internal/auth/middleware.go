package auth

import (
	"context"
	"net/http"
	"strings"

	apierrors "github.com/angellsan94/dogspots-api/internal/errors"
	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

func RequireAuth(tokenManager *TokenManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := strings.TrimSpace(r.Header.Get("Authorization"))
		if authorization == "" {
			apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "Missing bearer token", "unauthorized").Write(w)
			return
		}

		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "Invalid bearer token format", "unauthorized").Write(w)
			return
		}

		claims, err := tokenManager.ValidateToken(parts[1])
		if err != nil {
			apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "The access token is invalid or expired", "unauthorized").Write(w)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}
