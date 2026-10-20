package middleware

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggerReturnsRequestID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestID, ok := r.Context().Value(RequestIDKey).(string); !ok || requestID == "" {
			t.Fatal("request ID missing from context")
		}
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Header().Get("X-Request-ID") == "" {
		t.Fatal("X-Request-ID response header is missing")
	}
}

func TestMaxBodySizeRejectsKnownOversizeBody(t *testing.T) {
	called := false
	handler := MaxBodySize(3)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("four")))

	if called {
		t.Fatal("handler ran for an oversized request")
	}
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
	if got := response.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("content type = %q, want application/problem+json", got)
	}
}

func TestMaxBodySizeCapsUnknownLengthBody(t *testing.T) {
	handler := MaxBodySize(3)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		var maxBytesErr *http.MaxBytesError
		if !errors.As(err, &maxBytesErr) {
			t.Errorf("body read error = %v, want *http.MaxBytesError", err)
			return
		}
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))
	request := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader("four")))
	request.ContentLength = -1
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
}
