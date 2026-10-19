package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// LocalStorage implementa la interfaz Storage utilizando el sistema de archivos local.
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage crea una nueva instancia de LocalStorage.
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	// Asegurar que el directorio base existe
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create media directory: %w", err)
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

// Save guarda un stream de datos con un nombre aleatorio para evitar colisiones.
func (s *LocalStorage) Save(ctx context.Context, filename string, content io.Reader) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", fmt.Errorf("unsupported media extension: %s", ext)
	}
	newFilename := uuid.New().String() + ext
	targetPath := filepath.Join(s.baseDir, newFilename)
	temporary, err := os.CreateTemp(s.baseDir, ".upload-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := io.Copy(temporary, content); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("failed to copy content: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("failed to sync media file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("failed to close media file: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := os.Rename(temporaryPath, targetPath); err != nil {
		return "", fmt.Errorf("failed to save media file: %w", err)
	}
	url := fmt.Sprintf("/media/%s", newFilename)
	return url, nil
}

// Delete elimina el archivo de la ruta especificada.
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	// El path viene como "/media/..." o similar, necesitamos la parte del nombre del archivo
	filename := filepath.Base(path)
	targetPath := filepath.Join(s.baseDir, filename)

	if err := os.Remove(targetPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Si no existe, no hay error
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
