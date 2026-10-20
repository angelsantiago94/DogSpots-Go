package storage

import (
	"net/http"
	"path/filepath"
)

func MediaHandler(mediaDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		filename := r.PathValue("filename")
		extension := filepath.Ext(filename)
		if filepath.Base(filename) != filename || (extension != ".jpg" && extension != ".jpeg" && extension != ".png") {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(mediaDir, filename))
	})
}
