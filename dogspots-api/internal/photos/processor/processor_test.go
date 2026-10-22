package processor

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"
)

func TestProcessImageFormatsAndThumbnail(t *testing.T) {
	for _, format := range []string{"jpeg", "png"} {
		t.Run(format, func(t *testing.T) {
			input := encodeTestImage(t, format)
			original, thumbnail, info, err := ProcessImage(bytes.NewReader(input), 1<<20)
			if err != nil {
				t.Fatalf("ProcessImage() error = %v", err)
			}
			if info.Format != "jpeg" || info.Width != 1600 || info.Height != 800 {
				t.Fatalf("image info = %+v", info)
			}
			_, originalFormat, err := image.DecodeConfig(bytes.NewReader(original))
			if err != nil || originalFormat != "jpeg" {
				t.Fatalf("processed image format = %q, error = %v", originalFormat, err)
			}
			thumbnailConfig, _, err := image.DecodeConfig(bytes.NewReader(thumbnail))
			if err != nil {
				t.Fatalf("decode thumbnail: %v", err)
			}
			if thumbnailConfig.Width > 400 || thumbnailConfig.Height > 400 {
				t.Fatalf("thumbnail dimensions = %dx%d", thumbnailConfig.Width, thumbnailConfig.Height)
			}
		})
	}
}

func TestProcessImageWebPIsReencodedAsJPEG(t *testing.T) {
	input := []byte("RIFF\x00\x00\x00\x00WEBP")
	_, _, _, err := ProcessImage(bytes.NewReader(input), 1<<20)
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("ProcessImage() error = %v, want unsupported format", err)
	}
}

func TestProcessImageRejectsOversizeAndInvalidInput(t *testing.T) {
	if _, _, _, err := ProcessImage(strings.NewReader("0123456789"), 5); !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("oversize error = %v", err)
	}
	if _, _, _, err := ProcessImage(strings.NewReader("not an image"), 100); !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("invalid image error = %v", err)
	}
}

func TestProcessImageRejectsTooManyPixels(t *testing.T) {
	input := pngHeader(5000, 4000)
	_, _, _, err := ProcessImage(bytes.NewReader(input), 1<<20)
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("ProcessImage() error = %v, want unsupported format", err)
	}
}

func pngHeader(width, height uint32) []byte {
	data := make([]byte, 13)
	binary.BigEndian.PutUint32(data[0:4], width)
	binary.BigEndian.PutUint32(data[4:8], height)
	data[8] = 8 // profundidad de bits
	data[9] = 2 // color RGB

	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(data)))
	var checksum [4]byte
	binary.BigEndian.PutUint32(checksum[:], crc32.ChecksumIEEE(append([]byte("IHDR"), data...)))

	header := []byte("\x89PNG\r\n\x1a\n")
	header = append(header, length[:]...)
	header = append(header, "IHDR"...)
	header = append(header, data...)
	return append(header, checksum[:]...)
}

func TestProcessImageRemovesEXIFSegment(t *testing.T) {
	jpegData := encodeTestImage(t, "jpeg")
	segment := []byte{0xff, 0xe1, 0, 14, 'E', 'x', 'i', 'f', 0, 0, 1, 2, 3, 4}
	withEXIF := append(append(append([]byte{}, jpegData[:2]...), segment...), jpegData[2:]...)
	processed, _, _, err := ProcessImage(bytes.NewReader(withEXIF), 1<<20)
	if err != nil {
		t.Fatalf("ProcessImage() error = %v", err)
	}
	if bytes.Contains(processed, []byte("Exif\x00\x00")) {
		t.Fatal("processed JPEG still contains EXIF metadata")
	}
}

func encodeTestImage(t *testing.T, format string) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1600, 800))
	for y := 0; y < 800; y++ {
		for x := 0; x < 1600; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 80, A: 255})
		}
	}
	var output bytes.Buffer
	var err error
	if format == "png" {
		err = png.Encode(&output, img)
	} else {
		err = jpeg.Encode(&output, img, &jpeg.Options{Quality: 90})
	}
	if err != nil {
		t.Fatalf("encode test image: %v", err)
	}
	return output.Bytes()
}
