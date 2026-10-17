package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *SQLite {
	t.Helper()
	s, err := NewSQLite(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s
}

func newUser(t *testing.T, s *SQLite, email string) int64 {
	t.Helper()
	u, err := s.CreateUser(context.Background(), email, "hash")
	if err != nil {
		t.Fatal(err)
	}
	return u.ID
}

func TestMigrateIsIdempotent(t *testing.T) {
	s := newTestStore(t)
	if err := s.Migrate(context.Background()); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
}

func TestCreateAndGet(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	uid := newUser(t, s, "a@example.com")

	created, err := s.Create(ctx, uid, "牛乳")
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.Done || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Errorf("unexpected todo: %+v", created)
	}
	got, err := s.Get(ctx, uid, created.ID)
	if err != nil || got.Title != "牛乳" {
		t.Fatalf("Get: %+v, %v", got, err)
	}
}

func TestListOrderAndEmpty(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	uid := newUser(t, s, "a@example.com")

	empty, err := s.List(ctx, uid, "all")
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("want empty non-nil slice, got %#v, %v", empty, err)
	}

	s.Create(ctx, uid, "first")
	s.Create(ctx, uid, "second")
	list, _ := s.List(ctx, uid, "all")
	if len(list) != 2 || list[0].Title != "second" {
		t.Errorf("want newest first, got %+v", list)
	}
}

func TestListFilter(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	uid := newUser(t, s, "a@example.com")

	a, _ := s.Create(ctx, uid, "active")
	d, _ := s.Create(ctx, uid, "done")
	done := true
	s.Update(ctx, uid, d.ID, UpdateInput{Done: &done})

	tests := []struct {
		status string
		want   int64
	}{{"active", a.ID}, {"done", d.ID}}
	for _, tt := range tests {
		list, err := s.List(ctx, uid, tt.status)
		if err != nil || len(list) != 1 || list[0].ID != tt.want {
			t.Errorf("%s: %+v, %v", tt.status, list, err)
		}
	}
	if all, _ := s.List(ctx, uid, "all"); len(all) != 2 {
		t.Errorf("all: want 2, got %d", len(all))
	}
}

func TestUpdatePartial(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	uid := newUser(t, s, "a@example.com")
	c, _ := s.Create(ctx, uid, "a")

	done := true
	got, err := s.Update(ctx, uid, c.ID, UpdateInput{Done: &done})
	if err != nil || !got.Done || got.Title != "a" {
		t.Fatalf("done only: %+v, %v", got, err)
	}

	title := "b"
	got, err = s.Update(ctx, uid, c.ID, UpdateInput{Title: &title})
	if err != nil || got.Title != "b" || !got.Done {
		t.Fatalf("title only: %+v, %v", got, err)
	}

	notDone := false // false を指定しても「未指定」扱いにならない
	got, err = s.Update(ctx, uid, c.ID, UpdateInput{Done: &notDone})
	if err != nil || got.Done {
		t.Fatalf("done=false: %+v, %v", got, err)
	}
}

func TestNotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	uid := newUser(t, s, "a@example.com")
	done := true

	if _, err := s.Get(ctx, uid, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get: %v", err)
	}
	if _, err := s.Update(ctx, uid, 999, UpdateInput{Done: &done}); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update: %v", err)
	}
	if err := s.Delete(ctx, uid, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete: %v", err)
	}
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	uid := newUser(t, s, "a@example.com")
	c, _ := s.Create(ctx, uid, "a")

	if err := s.Delete(ctx, uid, c.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, uid, c.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// 他ユーザーのタスクは、取得・更新・削除のどれでも「存在しない」扱いになる
func TestIsolationBetweenUsers(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	alice := newUser(t, s, "alice@example.com")
	bob := newUser(t, s, "bob@example.com")
	c, _ := s.Create(ctx, alice, "secret")
	done := true

	if list, _ := s.List(ctx, bob, "all"); len(list) != 0 {
		t.Errorf("bob sees alice's todos: %+v", list)
	}
	if _, err := s.Get(ctx, bob, c.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get: %v", err)
	}
	if _, err := s.Update(ctx, bob, c.ID, UpdateInput{Done: &done}); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update: %v", err)
	}
	if err := s.Delete(ctx, bob, c.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete: %v", err)
	}
	if got, err := s.Get(ctx, alice, c.ID); err != nil || got.Done {
		t.Errorf("alice's todo was modified: %+v, %v", got, err)
	}
}

func TestConcurrentCreate(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	uid := newUser(t, s, "a@example.com")

	errs := make(chan error, 50)
	for i := 0; i < 50; i++ {
		go func() {
			_, err := s.Create(ctx, uid, "x")
			errs <- err
		}()
	}
	for i := 0; i < 50; i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	if list, _ := s.List(ctx, uid, "all"); len(list) != 50 {
		t.Errorf("want 50, got %d", len(list))
	}
}

func TestAccounts(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	u, err := s.CreateUser(ctx, "a@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateUser(ctx, "a@example.com", "hash"); !errors.Is(err, ErrEmailTaken) {
		t.Errorf("want ErrEmailTaken, got %v", err)
	}
	got, hash, err := s.UserByEmail(ctx, "a@example.com")
	if err != nil || got.ID != u.ID || hash != "hash" {
		t.Errorf("UserByEmail: %+v %q %v", got, hash, err)
	}
	if _, _, err := s.UserByEmail(ctx, "none@example.com"); !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestSessions(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	u, _ := s.CreateUser(ctx, "a@example.com", "hash")

	if err := s.CreateSession(ctx, u.ID, "valid", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if got, err := s.UserBySession(ctx, "valid"); err != nil || got.ID != u.ID {
		t.Errorf("valid session: %+v, %v", got, err)
	}

	if err := s.CreateSession(ctx, u.ID, "expired", time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UserBySession(ctx, "expired"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expired session: %v", err)
	}
	if _, err := s.UserBySession(ctx, "unknown"); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown session: %v", err)
	}

	if err := s.DeleteSession(ctx, "valid"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UserBySession(ctx, "valid"); !errors.Is(err, ErrNotFound) {
		t.Errorf("deleted session: %v", err)
	}
}
