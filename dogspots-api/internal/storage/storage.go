package storage

import (
	"context"
	"io"
)

// Storage define la interfaz para el almacenamiento de archivos.
// Se utiliza para abstraer el destino físico (disco local, S3, R2, etc.)
type Storage interface {
	// Save guarda un stream de datos en la ruta especificada y devuelve la URL pública y el nombre del archivo.
	Save(ctx context.Context, filename string, content io.Reader) (url string, err error)
	// Delete elimina un archivo de la ruta especificada.
	Delete(ctx context.Context, path string) error
}
