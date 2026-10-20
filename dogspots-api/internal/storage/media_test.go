package storage

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestMediaHandlerServesFileWithLongCacheHeaders(t *testing.T) {
	mediaDir := t.TempDir()
	const filename = "sample.jpg"
	const contents = "image data"
	if err := os.WriteFile(filepath.Join(mediaDir, filename), []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /media/{filename}", MediaHandler(mediaDir))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/media/"+filename, nil))

	if response.Code != http.StatusOK {
		t.Fatalf("GET media status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != contents {
		t.Fatalf("GET media body = %q, want %q", response.Body.String(), contents)
	}
	if got := response.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Errorf("Cache-Control = %q", got)
	}
	if got := response.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
}
