package auth

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/angellsan94/dogspots-api/db/migrations"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type recordingStorage struct {
	deleted []string
}

func (*recordingStorage) Save(context.Context, string, io.Reader) (string, error) {
	return "", errors.New("not used")
}

func (s *recordingStorage) Delete(_ context.Context, path string) error {
	s.deleted = append(s.deleted, path)
	return nil
}

type accountFixture struct {
	db       *sql.DB
	userA    uuid.UUID
	userB    uuid.UUID
	placeOwn uuid.UUID
	placeOth uuid.UUID
	fileA    string
	fileATh  string
	fileA2   string
	fileA2Th string
	fileB    string
	fileBTh  string
}

func newAccountFixture(t *testing.T) accountFixture {
	t.Helper()
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := migrations.Up(context.Background(), database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	f := accountFixture{db: database, userA: uuid.New(), userB: uuid.New()}
	now := time.Now().UTC()

	f.userA = insertAccountUser(t, database, f.userA, "a@example.com", now)
	f.userB = insertAccountUser(t, database, f.userB, "b@example.com", now)
	f.placeOwn = insertAccountPlace(t, database, f.userA, "Plaza de A", now)
	f.placeOth = insertAccountPlace(t, database, f.userB, "Parque de B", now)

	f.fileA, f.fileATh = "/media/a.jpg", "/media/a-thumb.jpg"
	f.fileA2, f.fileA2Th = "/media/a2.jpg", "/media/a2-thumb.jpg"
	f.fileB, f.fileBTh = "/media/b.jpg", "/media/b-thumb.jpg"
	insertAccountPhoto(t, database, f.placeOwn, f.userA, f.fileA, f.fileATh, now)
	insertAccountPhoto(t, database, f.placeOth, f.userA, f.fileA2, f.fileA2Th, now)
	insertAccountPhoto(t, database, f.placeOth, f.userB, f.fileB, f.fileBTh, now)

	commentBOnOwn := insertAccountComment(t, database, f.placeOwn, f.userB, "Comentario de B en el lugar de A", now)
	commentByA := insertAccountComment(t, database, f.placeOth, f.userA, "Comentario de A en el lugar de B", now)
	insertAccountCommentReply(t, database, f.placeOth, commentByA, f.userB, "Respuesta de B", now)
	insertAccountRating(t, database, f.placeOwn, f.userB, 3, now)
	insertAccountRating(t, database, f.placeOth, f.userA, 5, now)
	insertAccountRating(t, database, f.placeOth, f.userB, 4, now)
	insertAccountReport(t, database, f.userA, &f.placeOth, nil, now)
	insertAccountReport(t, database, f.userB, nil, &commentByA, now)
	insertAccountReport(t, database, f.userB, nil, &commentBOnOwn, now)
	insertAccountToken(t, database, f.userA, now)
	insertAccountToken(t, database, f.userB, now)

	if _, err := database.Exec(`UPDATE places SET comment_count = 2, rating_avg = 4.5, rating_count = 2 WHERE id = ?`, f.placeOth.String()); err != nil {
		t.Fatalf("seed place aggregates: %v", err)
	}
	if _, err := database.Exec(`UPDATE comments SET reply_count = 1 WHERE id = ?`, commentByA.String()); err != nil {
		t.Fatalf("seed reply count: %v", err)
	}
	return f
}

func insertAccountUser(t *testing.T, database *sql.DB, id uuid.UUID, email string, now time.Time) uuid.UUID {
	t.Helper()
	if _, err := database.Exec(`INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?, ?, 'hash', 'User', ?, ?)`,
		id.String(), email, now, now); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func insertAccountPlace(t *testing.T, database *sql.DB, authorID uuid.UUID, title string, now time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := database.Exec(`INSERT INTO places (id, author_id, title, lat, lng, category, created_at, updated_at) VALUES (?, ?, ?, 40.4, -3.7, 'park', ?, ?)`,
		id.String(), authorID.String(), title, now, now); err != nil {
		t.Fatalf("seed place: %v", err)
	}
	return id
}

func insertAccountPhoto(t *testing.T, database *sql.DB, placeID, uploadedBy uuid.UUID, url, thumbURL string, now time.Time) {
	t.Helper()
	if _, err := database.Exec(`INSERT INTO photos (id, place_id, url, thumb_url, uploaded_by, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), placeID.String(), url, thumbURL, uploadedBy.String(), now); err != nil {
		t.Fatalf("seed photo: %v", err)
	}
}

func insertAccountComment(t *testing.T, database *sql.DB, placeID, authorID uuid.UUID, body string, now time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := database.Exec(`INSERT INTO comments (id, place_id, author_id, body, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id.String(), placeID.String(), authorID.String(), body, now, now); err != nil {
		t.Fatalf("seed comment: %v", err)
	}
	return id
}

func insertAccountCommentReply(t *testing.T, database *sql.DB, placeID, parentID, authorID uuid.UUID, body string, now time.Time) {
	t.Helper()
	if _, err := database.Exec(`INSERT INTO comments (id, place_id, parent_id, author_id, body, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), placeID.String(), parentID.String(), authorID.String(), body, now, now); err != nil {
		t.Fatalf("seed reply: %v", err)
	}
}

func insertAccountRating(t *testing.T, database *sql.DB, placeID, userID uuid.UUID, value int, now time.Time) {
	t.Helper()
	if _, err := database.Exec(`INSERT INTO ratings (id, user_id, place_id, rating, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), userID.String(), placeID.String(), value, now, now); err != nil {
		t.Fatalf("seed rating: %v", err)
	}
}

func insertAccountReport(t *testing.T, database *sql.DB, reporterID uuid.UUID, placeID, commentID *uuid.UUID, now time.Time) {
	t.Helper()
	var placeArg, commentArg any
	if placeID != nil {
		placeArg = placeID.String()
	}
	if commentID != nil {
		commentArg = commentID.String()
	}
	if _, err := database.Exec(`INSERT INTO reports (id, user_id, place_id, comment_id, reason, created_at) VALUES (?, ?, ?, ?, 'spam', ?)`,
		uuid.NewString(), reporterID.String(), placeArg, commentArg, now); err != nil {
		t.Fatalf("seed report: %v", err)
	}
}

func insertAccountToken(t *testing.T, database *sql.DB, userID uuid.UUID, now time.Time) {
	t.Helper()
	if _, err := database.Exec(`INSERT INTO refresh_tokens (id, user_id, token_hash, family_id, expires_at, created_at, is_revoked) VALUES (?, ?, 'hash', ?, ?, ?, 0)`,
		uuid.NewString(), userID.String(), uuid.NewString(), now.Add(time.Hour), now); err != nil {
		t.Fatalf("seed refresh token: %v", err)
	}
}

func TestDeleteMeRemovesPlacesAndAnonymizesComments(t *testing.T) {
	f := newAccountFixture(t)
	storage := &recordingStorage{}
	repo, err := NewSQLiteRepository(f.db)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewAccountHandler(repo, storage)

	tokenManager := NewTokenManager("test-secret", time.Minute)
	token, err := tokenManager.GenerateAccessToken(f.userA)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/users/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	RequireAuth(tokenManager, http.HandlerFunc(handler.DeleteMe)).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}

	assertCount(t, f.db, `SELECT COUNT(*) FROM places WHERE id = ?`, f.placeOwn.String(), 0)
	assertCount(t, f.db, `SELECT COUNT(*) FROM places WHERE id = ?`, f.placeOth.String(), 1)

	var name, email, passwordHash string
	if err := f.db.QueryRow(`SELECT name, email, password_hash FROM users WHERE id = ?`, f.userA.String()).Scan(&name, &email, &passwordHash); err != nil {
		t.Fatal(err)
	}
	if name != "Usuario eliminado" || email != deletedAccountEmail(f.userA) || passwordHash != "!" {
		t.Fatalf("user was not anonymized: name=%q email=%q hash=%q", name, email, passwordHash)
	}
	assertCount(t, f.db, `SELECT COUNT(*) FROM refresh_tokens WHERE user_id = ? AND is_revoked = 1`, f.userA.String(), 1)
	assertCount(t, f.db, `SELECT COUNT(*) FROM refresh_tokens WHERE user_id = ? AND is_revoked = 0`, f.userB.String(), 1)

	assertCount(t, f.db, `SELECT COUNT(*) FROM photos WHERE uploaded_by = ?`, f.userA.String(), 0)
	assertCount(t, f.db, `SELECT COUNT(*) FROM photos WHERE url = ?`, f.fileB, 1)

	var deleted int
	var body sql.NullString
	if err := f.db.QueryRow(`SELECT deleted, body FROM comments WHERE place_id = ? AND parent_id IS NULL`, f.placeOth.String()).Scan(&deleted, &body); err != nil {
		t.Fatal(err)
	}
	if deleted != 1 || body.Valid {
		t.Fatalf("comment was not anonymized: deleted=%d body=%v", deleted, body)
	}
	assertCount(t, f.db, `SELECT COUNT(*) FROM comments WHERE place_id = ? AND parent_id IS NOT NULL`, f.placeOth.String(), 1)
	assertCount(t, f.db, `SELECT COUNT(*) FROM comments WHERE place_id = ?`, f.placeOwn.String(), 0)

	var commentCount, replyCount, ratingCount int
	var ratingAvg sql.NullFloat64
	if err := f.db.QueryRow(`SELECT comment_count, rating_count, rating_avg FROM places WHERE id = ?`, f.placeOth.String()).Scan(&commentCount, &ratingCount, &ratingAvg); err != nil {
		t.Fatal(err)
	}
	if commentCount != 1 || ratingCount != 1 || !ratingAvg.Valid || ratingAvg.Float64 != 4 {
		t.Fatalf("place aggregates = comment_count:%d rating_count:%d rating_avg:%v", commentCount, ratingCount, ratingAvg)
	}
	if err := f.db.QueryRow(`SELECT reply_count FROM comments WHERE place_id = ? AND parent_id IS NULL`, f.placeOth.String()).Scan(&replyCount); err != nil {
		t.Fatal(err)
	}
	if replyCount != 1 {
		t.Fatalf("reply_count = %d, want 1", replyCount)
	}

	assertCount(t, f.db, `SELECT COUNT(*) FROM reports WHERE user_id = ?`, f.userA.String(), 0)
	// Queda solo la denuncia de B sobre el comentario anonimizado de A.
	assertCount(t, f.db, `SELECT COUNT(*) FROM reports`, "", 1)

	wantFiles := map[string]bool{f.fileA: true, f.fileATh: true, f.fileA2: true, f.fileA2Th: true}
	if len(storage.deleted) != len(wantFiles) {
		t.Fatalf("deleted files = %v", storage.deleted)
	}
	for _, file := range storage.deleted {
		if !wantFiles[file] {
			t.Fatalf("unexpected file deleted: %s", file)
		}
		delete(wantFiles, file)
	}
	if len(wantFiles) != 0 {
		t.Fatalf("missing files: %v", wantFiles)
	}

	// Una segunda llamada es idempotente y no borra ficheros ya eliminados.
	urls, err := handler.repo.DeleteAccountData(context.Background(), f.userA)
	if err != nil {
		t.Fatalf("second delete: %v", err)
	}
	if len(urls) != 0 {
		t.Fatalf("second delete returned files: %v", urls)
	}
}

func assertCount(t *testing.T, database *sql.DB, query, arg string, want int) {
	t.Helper()
	var got int
	var err error
	if arg == "" {
		err = database.QueryRow(query).Scan(&got)
	} else {
		err = database.QueryRow(query, arg).Scan(&got)
	}
	if err != nil {
		t.Fatalf("count query: %v", err)
	}
	if got != want {
		t.Fatalf("count = %d, want %d (%s)", got, want, query)
	}
}
