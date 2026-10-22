package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/angellsan94/dogspots-api/internal/auth"
	"github.com/google/uuid"
)

func TestRateLimiterFallsBackToClientIP(t *testing.T) {
	calls := 0
	handler := NewRateLimiter(1, time.Minute, false)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
	}))

	for requestNumber := 1; requestNumber <= 2; requestNumber++ {
		request := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		request.RemoteAddr = "203.0.113.7:1234"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if requestNumber == 1 && response.Code != http.StatusCreated {
			t.Fatalf("first request: expected 201, got %d", response.Code)
		}
		if requestNumber == 2 && response.Code != http.StatusTooManyRequests {
			t.Fatalf("second request: expected 429, got %d", response.Code)
		}
	}
	if calls != 1 {
		t.Fatalf("expected handler to run once, got %d", calls)
	}
}

func TestRateLimiterKeysByForwardedClientIP(t *testing.T) {
	handler := NewRateLimiter(1, time.Minute, true)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	send := func(ip string) int {
		request := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		request.Header.Set("X-Real-IP", ip)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response.Code
	}

	if code := send("198.51.100.1"); code != http.StatusCreated {
		t.Fatalf("first client IP: expected 201, got %d", code)
	}
	if code := send("198.51.100.1"); code != http.StatusTooManyRequests {
		t.Fatalf("same client IP: expected 429, got %d", code)
	}
	if code := send("198.51.100.2"); code != http.StatusCreated {
		t.Fatalf("different client IP: expected 201, got %d", code)
	}
}

func TestRateLimiterIgnoresSpoofedIPWhenProxyUntrusted(t *testing.T) {
	handler := NewRateLimiter(1, time.Minute, false)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	send := func(realIP string) int {
		request := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		request.RemoteAddr = "203.0.113.7:1234"
		request.Header.Set("X-Real-IP", realIP)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response.Code
	}

	if code := send("198.51.100.1"); code != http.StatusCreated {
		t.Fatalf("first request: expected 201, got %d", code)
	}
	if code := send("198.51.100.2"); code != http.StatusTooManyRequests {
		t.Fatalf("spoofed X-Real-IP must not bypass the limit: expected 429, got %d", code)
	}
}

func TestRateLimiterLimitsAuthenticatedUser(t *testing.T) {
	manager := auth.NewTokenManager("rate-limit-test-secret", time.Minute)
	userID := uuid.New()
	token, err := manager.GenerateAccessToken(userID)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	handler := auth.RequireAuth(manager, NewRateLimiter(2, time.Minute, false)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
