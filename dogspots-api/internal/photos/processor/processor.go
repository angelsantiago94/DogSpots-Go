package processor

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

var ErrFileTooLarge = fmt.Errorf("file too large")
var ErrUnsupportedFormat = fmt.Errorf("unsupported image format")

const maxImagePixels = 16_000_000

// Dos decodificaciones simultáneas como máximo para acotar la memoria.
var processSlots = make(chan struct{}, 2)

// ImageInfo contiene metadatos de la imagen procesada.
type ImageInfo struct {
	Width  int
	Height int
	Format string
}

// ProcessImage valida la imagen, elimina EXIF, la redimensiona y genera una miniatura.
// Devuelve el buffer con la imagen original procesada, el buffer con la miniatura, y la info.
func ProcessImage(input io.Reader, maxSize int) (original []byte, thumb []byte, info ImageInfo, err error) {
	processSlots <- struct{}{}
	defer func() { <-processSlots }()

	data, err := io.ReadAll(io.LimitReader(input, int64(maxSize)+1))
	if err != nil {
		return nil, nil, ImageInfo{}, fmt.Errorf("failed to read input: %w", err)
	}

	if len(data) > maxSize {
		return nil, nil, ImageInfo{}, ErrFileTooLarge
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, nil, ImageInfo{}, fmt.Errorf("%w: %v", ErrUnsupportedFormat, err)
	}
	format = strings.ToLower(format)
	if format != "jpeg" && format != "png" && format != "webp" {
		return nil, nil, ImageInfo{}, fmt.Errorf("%w: %s", ErrUnsupportedFormat, format)
	}
	if config.Width <= 0 || config.Height <= 0 || int64(config.Width) > maxImagePixels/int64(config.Height) {
		return nil, nil, ImageInfo{}, fmt.Errorf("%w: invalid image dimensions", ErrUnsupportedFormat)
	}

	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, nil, ImageInfo{}, fmt.Errorf("%w: %v", ErrUnsupportedFormat, err)
	}
	info = ImageInfo{Width: img.Bounds().Dx(), Height: img.Bounds().Dy(), Format: "jpeg"}

	if img.Bounds().Dx() > 2048 || img.Bounds().Dy() > 2048 {
		img = imaging.Fit(img, 2048, 2048, imaging.Lanczos)
	}
	info.Width, info.Height = img.Bounds().Dx(), img.Bounds().Dy()
	thumbImage := imaging.Fit(img, 400, 400, imaging.Lanczos)
	encoder := func(target io.Writer, source image.Image) error {
		background := imaging.New(source.Bounds().Dx(), source.Bounds().Dy(), color.White)
		return jpeg.Encode(target, imaging.Overlay(background, source, image.Pt(0, 0), 1), &jpeg.Options{Quality: 85})
	}
	var originalBuffer, thumbBuffer bytes.Buffer
	if err := encoder(&originalBuffer, img); err != nil {
		return nil, nil, ImageInfo{}, fmt.Errorf("failed to encode image: %w", err)
	}
	if err := encoder(&thumbBuffer, thumbImage); err != nil {
		return nil, nil, ImageInfo{}, fmt.Errorf("failed to encode thumbnail: %w", err)
	}
	return originalBuffer.Bytes(), thumbBuffer.Bytes(), info, nil
}
