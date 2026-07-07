package likeable

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest-likeable/migrations"
	"github.com/nicolasbonnici/gorest/database"
	_ "github.com/nicolasbonnici/gorest/database/sqlite"
)

func newTestDB(t *testing.T) database.Database {
	t.Helper()

	dsn := "file:" + filepath.Join(t.TempDir(), "likeable_test.db")
	db, err := database.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	migs, err := migrations.GetMigrations().Migrations()
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}
	for _, m := range migs {
		if err := m.Executor.Up(ctx, db); err != nil {
			t.Fatalf("migrate %s: %v", m.Version, err)
		}
	}
	return db
}

func insertLike(t *testing.T, db database.Database, likerID *string, likeableType, likeableID string) {
	t.Helper()

	like := Like{
		Id:         uuid.New().String(),
		LikerId:    likerID,
		LikeableId: likeableID,
		Likeable:   likeableType,
		LikedAt:    time.Now(),
	}
	if err := NewLikeService(db).crud.Create(context.Background(), like); err != nil {
		t.Fatalf("insert like: %v", err)
	}
}

func ptr(s string) *string { return &s }

func TestCount(t *testing.T) {
	db := newTestDB(t)
	svc := NewLikeService(db)
	ctx := context.Background()

	insertLike(t, db, ptr("user-1"), "post", "post-1")
	insertLike(t, db, ptr("user-2"), "post", "post-1")
	insertLike(t, db, ptr("user-1"), "post", "post-2")

	got, err := svc.Count(ctx, "post", "post-1")
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if got != 2 {
		t.Errorf("post-1 count = %d, want 2", got)
	}

	got, err = svc.Count(ctx, "post", "missing")
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if got != 0 {
		t.Errorf("missing count = %d, want 0", got)
	}
}

func TestCountBatch(t *testing.T) {
	db := newTestDB(t)
	svc := NewLikeService(db)
	ctx := context.Background()

	insertLike(t, db, ptr("user-1"), "post", "post-1")
	insertLike(t, db, ptr("user-2"), "post", "post-1")
	insertLike(t, db, ptr("user-1"), "post", "post-2")
	insertLike(t, db, ptr("user-1"), "comment", "post-1")

	counts, err := svc.CountBatch(ctx, "post", []string{"post-1", "post-2", "post-3"})
	if err != nil {
		t.Fatalf("CountBatch: %v", err)
	}

	want := map[string]int64{"post-1": 2, "post-2": 1, "post-3": 0}
	for id, w := range want {
		if counts[id] != w {
			t.Errorf("count[%s] = %d, want %d", id, counts[id], w)
		}
	}
	if len(counts) != len(want) {
		t.Errorf("counts has %d entries, want %d", len(counts), len(want))
	}
}

func TestCountBatchEmpty(t *testing.T) {
	db := newTestDB(t)
	counts, err := NewLikeService(db).CountBatch(context.Background(), "post", nil)
	if err != nil {
		t.Fatalf("CountBatch: %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("expected empty map, got %v", counts)
	}
}

func TestLikedByBatch(t *testing.T) {
	db := newTestDB(t)
	svc := NewLikeService(db)
	ctx := context.Background()

	insertLike(t, db, ptr("user-1"), "post", "post-1")
	insertLike(t, db, ptr("user-2"), "post", "post-2")
	insertLike(t, db, ptr("user-1"), "post", "post-3")

	liked, err := svc.LikedByBatch(ctx, "user-1", "post", []string{"post-1", "post-2", "post-3"})
	if err != nil {
		t.Fatalf("LikedByBatch: %v", err)
	}

	want := map[string]bool{"post-1": true, "post-2": false, "post-3": true}
	for id, w := range want {
		if liked[id] != w {
			t.Errorf("liked[%s] = %v, want %v", id, liked[id], w)
		}
	}
}

func TestLikedByBatchAnonymous(t *testing.T) {
	db := newTestDB(t)
	liked, err := NewLikeService(db).LikedByBatch(context.Background(), "", "post", []string{"post-1"})
	if err != nil {
		t.Fatalf("LikedByBatch: %v", err)
	}
	if liked["post-1"] {
		t.Error("anonymous caller should never be reported as having liked")
	}
}
