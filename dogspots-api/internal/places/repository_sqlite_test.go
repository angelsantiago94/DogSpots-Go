package places

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/angellsan94/dogspots-api/db/migrations"
	"github.com/angellsan94/dogspots-api/internal/auth"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func newTestRepository(t *testing.T) (*sql.DB, *SQLiteRepository, uuid.UUID) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if err := migrations.Up(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	userID := uuid.New()
	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		userID.String(), "owner@example.test", "hash", "Owner", createdAt, createdAt); err != nil {
		t.Fatal(err)
	}
	repo, err := NewSQLiteRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	return db, repo, userID
}

func TestCreateDetailAndOwnership(t *testing.T) {
	db, repo, ownerID := newTestRepository(t)
	ctx := context.Background()
	created, err := repo.Create(ctx, ownerID, CreatePlaceRequest{
		Title: "Parque canino", Lat: 40.4, Lng: -3.7, Category: "dog_park", Tags: []string{"agua", "sombra"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Author.DisplayName != "Owner" || created.Author.PlacesCount != 1 {
		t.Fatalf("unexpected author: %#v", created.Author)
	}
	if len(created.Tags) != 2 || len(created.Photos) != 0 {
		t.Fatalf("unexpected place relations: %#v", created)
	}
	if _, err := repo.Update(ctx, created.ID, uuid.New(), UpdatePlaceRequest{}); err != ErrForbidden {
		t.Fatalf("expected forbidden update, got %v", err)
	}
	if err := repo.Delete(ctx, created.ID, uuid.New()); err != ErrForbidden {
		t.Fatalf("expected forbidden delete, got %v", err)
	}
	if _, err := repo.Update(ctx, created.ID, ownerID, UpdatePlaceRequest{Title: stringPointer("Parque actualizado")}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, created.ID, ownerID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, created.ID, nil); err != ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
	var remaining int
	if err := db.QueryRow(`SELECT COUNT(*) FROM places`).Scan(&remaining); err != nil || remaining != 0 {
		t.Fatalf("unexpected remaining places: %d, %v", remaining, err)
	}
}

func TestDevelopmentSeedIsIdempotent(t *testing.T) {
	db, repo, _ := newTestRepository(t)
	ctx := context.Background()
	for range 2 {
		if err := SeedDevelopmentData(ctx, db); err != nil {
			t.Fatalf("SeedDevelopmentData() error = %v", err)
		}
	}
	places, err := repo.List(ctx, ListOptions{Sort: "recent", Limit: 100})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(places.Items) != len(developmentPlaces) {
		t.Fatalf("seeded %d places, want %d", len(places.Items), len(developmentPlaces))
	}
	for _, seeded := range developmentPlaces {
		place, err := repo.Get(ctx, uuid.MustParse(seeded.id), nil)
		if err != nil {
			t.Fatalf("Get(%q) error = %v", seeded.title, err)
		}
		if place.Title != seeded.title || place.Author.DisplayName != "DogSpots Demo" {
			t.Errorf("unexpected seeded place: %#v", place)
		}
	}
}

func TestListFiltersRadiusAndCursor(t *testing.T) {
	_, repo, ownerID := newTestRepository(t)
	ctx := context.Background()
	near, err := repo.Create(ctx, ownerID, CreatePlaceRequest{Title: "Agua cerca", Lat: 0, Lng: 0, Category: "park", Tags: []string{"agua"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.Create(ctx, ownerID, CreatePlaceRequest{Title: "Agua también", Lat: 0.02, Lng: 0, Category: "park", Tags: []string{"agua"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.Create(ctx, ownerID, CreatePlaceRequest{Title: "Lejos", Lat: 10, Lng: 10, Category: "beach", Tags: []string{"sombra"}})
	if err != nil {
		t.Fatal(err)
	}
	otherUser := uuid.New()
	createUser(t, repo, otherUser, "rater@example.test")
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := repo.db.Exec(`INSERT INTO ratings (id, user_id, place_id, rating, created_at) VALUES (?, ?, ?, ?, ?)`, uuid.New().String(), otherUser.String(), near.ID.String(), 5, now); err != nil {
		t.Fatal(err)
	}
	options := ListOptions{Sort: "recent", Limit: 1, RadiusM: 5000, Tags: []string{"agua"}}
	first, err := repo.List(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Items) != 1 || first.NextCursor == nil {
		t.Fatalf("expected first page and cursor, got %#v", first)
	}
	options.Cursor = *first.NextCursor
	second, err := repo.List(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Items) != 1 || second.Items[0].ID == first.Items[0].ID {
		t.Fatalf("cursor returned duplicate or missing row: %#v", second)
	}
	options = ListOptions{Sort: "recent", Limit: 10, Lat: floatPointer(0), Lng: floatPointer(0), RadiusM: 5000, MinRating: floatPointer(4)}
	filtered, err := repo.List(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered.Items) != 1 || filtered.Items[0].ID != near.ID || filtered.Items[0].DistanceM == nil {
		t.Fatalf("unexpected radius/rating results: %#v", filtered)
	}
	options = ListOptions{Sort: "rating", Limit: 10}
	rated, err := repo.List(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(rated.Items) != 3 || rated.Items[0].ID != near.ID {
		t.Fatalf("rating sort did not put the highest-rated place first: %#v", rated.Items)
	}
	options = ListOptions{Sort: "distance", Limit: 10, Lat: floatPointer(0), Lng: floatPointer(0), RadiusM: 5000}
	closest, err := repo.List(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(closest.Items) != 2 || closest.Items[0].ID != near.ID {
		t.Fatalf("distance sort did not put the closest place first: %#v", closest.Items)
	}
	options = ListOptions{Sort: "recent", Limit: 10, MinLat: floatPointer(-1), MaxLat: floatPointer(1), MinLng: floatPointer(-1), MaxLng: floatPointer(1)}
	boxed, err := repo.List(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(boxed.Items) != 2 {
		t.Fatalf("expected two bbox results, got %d", len(boxed.Items))
	}
}

func TestPlaceHandlersValidateInputAndEnforceAuthor(t *testing.T) {
	_, repo, _ := newTestRepository(t)
	manager := auth.NewTokenManager("test-secret", time.Minute)
	handler := NewHandler(repo, manager)
	mux := http.NewServeMux()
	mux.Handle("POST /places", auth.RequireAuth(manager, http.HandlerFunc(handler.Create)))
	mux.Handle("PATCH /places/{placeId}", auth.RequireAuth(manager, http.HandlerFunc(handler.Update)))
	mux.HandleFunc("GET /places/{placeId}", handler.Get)
	ownerID := uuid.New()
	otherID := uuid.New()
	createUser(t, repo, ownerID, "creator@example.test")
	createUser(t, repo, otherID, "other@example.test")
	ownerToken, _ := manager.GenerateAccessToken(ownerID)
	otherToken, _ := manager.GenerateAccessToken(otherID)

	invalid := httptest.NewRequest(http.MethodPost, "/places", strings.NewReader(`{"title":"Bad","lat":91,"lng":0,"category":"park"}`))
	invalid.Header.Set("Authorization", "Bearer "+ownerToken)
	invalidResponse := httptest.NewRecorder()
	mux.ServeHTTP(invalidResponse, invalid)
	if invalidResponse.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", invalidResponse.Code, invalidResponse.Body.String())
	}

	request := httptest.NewRequest(http.MethodPost, "/places", strings.NewReader(`{"title":"Buen parque","lat":40,"lng":-3,"category":"park"}`))
	request.Header.Set("Authorization", "Bearer "+ownerToken)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", response.Code, response.Body.String())
	}
	var place Place
	if err := json.Unmarshal(response.Body.Bytes(), &place); err != nil {
		t.Fatal(err)
	}

	patch := httptest.NewRequest(http.MethodPatch, "/places/"+place.ID.String(), strings.NewReader(`{"title":"No autorizado"}`))
	patch.Header.Set("Authorization", "Bearer "+otherToken)
	patchResponse := httptest.NewRecorder()
	mux.ServeHTTP(patchResponse, patch)
	if patchResponse.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", patchResponse.Code, patchResponse.Body.String())
	}

	detail := httptest.NewRequest(http.MethodGet, "/places/"+place.ID.String(), nil)
	detailResponse := httptest.NewRecorder()
	mux.ServeHTTP(detailResponse, detail)
	if detailResponse.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", detailResponse.Code, detailResponse.Body.String())
	}
}

func createUser(t *testing.T, repo *SQLiteRepository, id uuid.UUID, email string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := repo.db.Exec(`INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id.String(), email, "hash", "User", now, now); err != nil {
		t.Fatal(err)
	}
}

func stringPointer(value string) *string  { return &value }
func floatPointer(value float64) *float64 { return &value }
