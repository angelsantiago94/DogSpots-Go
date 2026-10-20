package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/angellsan94/dogspots-api/internal/auth"
	"github.com/google/uuid"
)

func TestRateLimiterLimitsAuthenticatedUser(t *testing.T) {
	manager := auth.NewTokenManager("rate-limit-test-secret", time.Minute)
	userID := uuid.New()
	token, err := manager.GenerateAccessToken(userID)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	handler := auth.RequireAuth(manager, NewRateLimiter(2, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
	})))

	for requestNumber := 1; requestNumber <= 3; requestNumber++ {
		request := httptest.NewRequest(http.MethodPost, "/comments", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if requestNumber <= 2 && response.Code != http.StatusCreated {
			t.Fatalf("request %d: expected 201, got %d", requestNumber, response.Code)
		}
		if requestNumber == 3 {
			if response.Code != http.StatusTooManyRequests {
				t.Fatalf("expected 429, got %d", response.Code)
			}
			if response.Header().Get("Retry-After") == "" {
				t.Fatal("expected Retry-After header")
			}
		}
	}
	if calls != 2 {
		t.Fatalf("expected handler to run twice, got %d", calls)
	}
	otherUserToken, err := manager.GenerateAccessToken(uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/comments", nil)
	request.Header.Set("Authorization", "Bearer "+otherUserToken)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected a separate user's request to pass, got %d", response.Code)
	}
}
