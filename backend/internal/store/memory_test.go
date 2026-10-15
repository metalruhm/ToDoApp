package store

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestMemory_CreateAndGet(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()

	created, err := m.Create(ctx, "牛乳")
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.Get(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "牛乳" || got.Done {
		t.Errorf("unexpected todo: %+v", got)
	}
}

func TestMemory_ListNewestFirst(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()
	m.Create(ctx, "first")
	m.Create(ctx, "second")

	list, _ := m.List(ctx, "all")
	if len(list) != 2 || list[0].Title != "second" {
		t.Errorf("want newest first, got %+v", list)
	}
}

func TestMemory_NotFound(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()

	if _, err := m.Get(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get: want ErrNotFound, got %v", err)
	}
	if err := m.Delete(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete: want ErrNotFound, got %v", err)
	}
}

func TestMemory_UpdatePartial(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()
	created, _ := m.Create(ctx, "a")

	done := true
	got, err := m.Update(ctx, created.ID, UpdateInput{Done: &done})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Done || got.Title != "a" { // タイトルは変わらない
		t.Errorf("unexpected: %+v", got)
	}
}

func TestMemory_Concurrent(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Create(ctx, "x")
		}()
	}
	wg.Wait()

	list, _ := m.List(ctx, "all")
	if len(list) != 100 {
		t.Errorf("want 100, got %d", len(list))
	}
}
