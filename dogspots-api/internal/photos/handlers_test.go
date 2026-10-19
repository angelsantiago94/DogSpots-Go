package photos

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/angellsan94/dogspots-api/internal/auth"
	"github.com/angellsan94/dogspots-api/internal/places"
	"github.com/google/uuid"
)

type handlerPhotoRepository struct {
	created Photo
	limit   int
}

func (r *handlerPhotoRepository) Create(_ context.Context, photo Photo, limit int) (Photo, error) {
	r.created = photo
	r.limit = limit
	return photo, nil
}

func (*handlerPhotoRepository) GetByID(context.Context, uuid.UUID) (Photo, error) {
	return Photo{}, nil
}

func (*handlerPhotoRepository) ListByPlaceID(context.Context, uuid.UUID) ([]Photo, error) {
	return nil, nil
}

func (*handlerPhotoRepository) CountByPlaceID(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (*handlerPhotoRepository) DeleteByPlaceID(context.Context, uuid.UUID) error {
	return nil
}

func (*handlerPhotoRepository) DeleteByID(context.Context, uuid.UUID) error {
	return nil
}

type handlerPlaceReader struct {
	place places.Place
}

func (r handlerPlaceReader) Get(context.Context, uuid.UUID, *uuid.UUID) (places.Place, error) {
	return r.place, nil
}

type handlerStorage struct {
	saved []string
}

func (s *handlerStorage) Save(_ context.Context, filename string, _ io.Reader) (string, error) {
	s.saved = append(s.saved, filename)
	return "/media/" + filename, nil
}

func (*handlerStorage) Delete(context.Context, string) error {
	return nil
}

func TestUploadStoresProcessedImageForPlaceOwner(t *testing.T) {
	userID, placeID := uuid.New(), uuid.New()
	repository := &handlerPhotoRepository{}
	fileStorage := &handlerStorage{}
	handler := NewHandler(repository, handlerPlaceReader{place: places.Place{Author: places.PublicUser{ID: userID}}}, fileStorage)
	tokenManager := auth.NewTokenManager("test-secret", time.Minute)
	token, err := tokenManager.GenerateAccessToken(userID)
	if err != nil {
		t.Fatal(err)
	}
	imageData := makeJPEG(t)
	var body bytes.Buffer
	multipartWriter := multipart.NewWriter(&body)
	file, err := multipartWriter.CreateFormFile("file", "uploaded.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(imageData); err != nil {
		t.Fatal(err)
	}
	if err := multipartWriter.WriteField("caption", "  Un paseo  "); err != nil {
		t.Fatal(err)
	}
	if err := multipartWriter.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/places/"+placeID.String()+"/photos", &body)
	request.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+token)
	request.SetPathValue("placeId", placeID.String())
	response := httptest.NewRecorder()
	auth.RequireAuth(tokenManager, http.HandlerFunc(handler.Upload)).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, response = %s", response.Code, response.Body.String())
	}
	if repository.limit != maxPhotos || repository.created.PlaceID != placeID || repository.created.UploadedBy != userID {
		t.Fatalf("persisted photo = %+v, limit = %d", repository.created, repository.limit)
	}
	if repository.created.Caption == nil || *repository.created.Caption != "Un paseo" {
		t.Fatalf("caption = %v", repository.created.Caption)
	}
	if len(fileStorage.saved) != 2 || !strings.HasSuffix(fileStorage.saved[0], ".jpg") || !strings.Contains(fileStorage.saved[1], "-thumb.jpg") {
		t.Fatalf("stored files = %v", fileStorage.saved)
	}
}

func TestUploadRejectsNonOwner(t *testing.T) {
	userID, placeID := uuid.New(), uuid.New()
	repository := &handlerPhotoRepository{}
	fileStorage := &handlerStorage{}
	handler := NewHandler(repository, handlerPlaceReader{place: places.Place{Author: places.PublicUser{ID: uuid.New()}}}, fileStorage)
	tokenManager := auth.NewTokenManager("test-secret", time.Minute)
	token, err := tokenManager.GenerateAccessToken(userID)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	request.Header.Set("Authorization", "Bearer "+token)
	request.SetPathValue("placeId", placeID.String())
	response := httptest.NewRecorder()
	auth.RequireAuth(tokenManager, http.HandlerFunc(handler.Upload)).ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || len(fileStorage.saved) != 0 {
		t.Fatalf("status = %d, saved files = %v", response.Code, fileStorage.saved)
	}
}

func makeJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 32, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 120, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return encoded.Bytes()
}
