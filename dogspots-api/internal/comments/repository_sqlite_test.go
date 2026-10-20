package comments

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/angellsan94/dogspots-api/db/migrations"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func TestCreateReplyAndDeleteComment(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if err := migrations.Up(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	placeID := uuid.New()
	userA := uuid.New()
	userB := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`, userA.String(), "alice@example.test", "hash", "Alice", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`, userB.String(), "bob@example.test", "hash", "Bob", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO places (id, author_id, title, description, lat, lng, category, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, placeID.String(), userA.String(), "Parque", "desc", 40.0, -3.0, "park", now, now); err != nil {
		t.Fatal(err)
	}

	repo := NewSQLiteRepository(db)
	root, err := repo.Create(context.Background(), placeID, userA, "comentario raíz", nil)
	if err != nil {
		t.Fatal(err)
	}
	if root.ParentID != nil {
		t.Fatalf("expected root comment to have nil parent, got %v", *root.ParentID)
	}

	reply, err := repo.Create(context.Background(), placeID, userB, "respuesta directa", &root.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reply.ParentID == nil || *reply.ParentID != root.ID {
		t.Fatalf("expected reply parent to equal root id, got %#v", reply.ParentID)
	}
	if reply.ReplyTo == nil || reply.ReplyTo.ID != userA {
		t.Fatalf("expected reply_to to target root author, got %#v", reply.ReplyTo)
	}

	nested, err := repo.Create(context.Background(), placeID, userA, "respuesta a respuesta", &reply.ID)
	if err != nil {
		t.Fatal(err)
	}
	if nested.ParentID == nil || *nested.ParentID != root.ID {
		t.Fatalf("expected nested parent to equal root id, got %#v", nested.ParentID)
	}
	if nested.ReplyTo == nil || nested.ReplyTo.ID != userB {
		t.Fatalf("expected nested reply_to to target direct reply author, got %#v", nested.ReplyTo)
	}
	if _, err := db.Exec(`UPDATE comments SET created_at = ? WHERE id = ?`, time.Now().UTC().Add(-10*time.Minute).Format(time.RFC3339Nano), reply.ID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE comments SET created_at = ? WHERE id = ?`, time.Now().UTC().Add(10*time.Minute).Format(time.RFC3339Nano), nested.ID.String()); err != nil {
		t.Fatal(err)
	}

	secondRoot, err := repo.Create(context.Background(), placeID, userB, "segundo comentario raíz", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE comments SET created_at = ? WHERE id = ?`, time.Now().UTC().Add(-time.Hour).Format(time.RFC3339Nano), root.ID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE comments SET created_at = ? WHERE id = ?`, time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano), secondRoot.ID.String()); err != nil {
		t.Fatal(err)
	}
	firstPage, err := repo.ListByPlace(context.Background(), placeID, "recent", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(firstPage.Items) != 1 || firstPage.NextCursor == nil {
		t.Fatalf("expected one item and a next cursor, got %#v", firstPage)
	}
	secondPage, err := repo.ListByPlace(context.Background(), placeID, "recent", 1, *firstPage.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].ID != root.ID || secondPage.Items[0].ID == secondRoot.ID {
		t.Fatalf("unexpected second page: %#v", secondPage)
	}
	if secondPage.Items[0].ReplyCount != 2 {
		t.Fatalf("expected root reply_count=2, got %d", secondPage.Items[0].ReplyCount)
	}
	repliesPage, err := repo.ListReplies(context.Background(), root.ID, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(repliesPage.Items) != 1 || repliesPage.Items[0].ID != reply.ID || repliesPage.NextCursor == nil {
		t.Fatalf("unexpected first replies page: %#v", repliesPage)
	}
	repliesPage, err = repo.ListReplies(context.Background(), root.ID, 1, *repliesPage.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	if len(repliesPage.Items) != 1 || repliesPage.Items[0].ID != nested.ID {
		t.Fatalf("unexpected second replies page: %#v", repliesPage)
	}
	oldestPage, err := repo.ListByPlace(context.Background(), placeID, "oldest", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(oldestPage.Items) != 1 || oldestPage.Items[0].ID != root.ID {
		t.Fatalf("unexpected oldest comments page: %#v", oldestPage)
	}
	if _, err := repo.Update(context.Background(), reply.ID, userA, "edición ajena"); err != ErrCommentForbidden {
		t.Fatalf("expected forbidden edit, got %v", err)
	}
	if err := repo.Delete(context.Background(), reply.ID, userA); err != ErrCommentForbidden {
		t.Fatalf("expected forbidden delete, got %v", err)
	}
	var commentCount int
	if err := db.QueryRow(`SELECT comment_count FROM places WHERE id = ?`, placeID.String()).Scan(&commentCount); err != nil {
		t.Fatal(err)
	}
	if commentCount != 4 {
		t.Fatalf("expected comment_count to include roots and replies, got %d", commentCount)
	}

	updated, err := repo.Update(context.Background(), reply.ID, userB, "respuesta actualizada")
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Edited {
		t.Fatal("expected updated comment to be marked as edited")
	}
	if updated.Body == nil || *updated.Body != "respuesta actualizada" {
		t.Fatalf("unexpected updated body: %#v", updated.Body)
	}

	if err := repo.Delete(context.Background(), reply.ID, userB); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByID(context.Background(), reply.ID); err != sql.ErrNoRows {
		t.Fatalf("expected reply without children to be physically deleted, got %v", err)
	}
	if err := db.QueryRow(`SELECT comment_count FROM places WHERE id = ?`, placeID.String()).Scan(&commentCount); err != nil {
		t.Fatal(err)
	}
	if commentCount != 3 {
		t.Fatalf("expected deleting reply to decrement comment_count, got %d", commentCount)
	}
	if err := repo.Delete(context.Background(), root.ID, userA); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetByID(context.Background(), root.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Body != nil || stored.Author != nil || !stored.Deleted {
		t.Fatalf("expected root with replies to be logically deleted, got %#v", stored)
	}
	if err := db.QueryRow(`SELECT comment_count FROM places WHERE id = ?`, placeID.String()).Scan(&commentCount); err != nil {
		t.Fatal(err)
	}
	if commentCount != 2 {
		t.Fatalf("expected logical delete to decrement active comment_count, got %d", commentCount)
	}
}
