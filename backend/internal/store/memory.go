package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"todoapp/backend/internal/model"
)

type Memory struct {
	mu     sync.Mutex
	items  map[int64]model.Todo
	nextID int64
}

func NewMemory() *Memory {
	return &Memory{items: map[int64]model.Todo{}, nextID: 1}
}

func (m *Memory) List(ctx context.Context, status string) ([]model.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]model.Todo, 0, len(m.items)) // 空でも null にならない
	for _, t := range m.items {
		if status == "active" && t.Done {
			continue
		}
		if status == "done" && !t.Done {
			continue
		}
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		if !result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].CreatedAt.After(result[j].CreatedAt)
		}
		return result[i].ID > result[j].ID
	})
	return result, nil
}

func (m *Memory) Get(ctx context.Context, id int64) (model.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.items[id]
	if !ok {
		return model.Todo{}, ErrNotFound
	}
	return t, nil
}

func (m *Memory) Create(ctx context.Context, title string) (model.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC().Truncate(time.Second)
	t := model.Todo{ID: m.nextID, Title: title, CreatedAt: now, UpdatedAt: now}
	m.items[t.ID] = t
	m.nextID++
	return t, nil
}

func (m *Memory) Update(ctx context.Context, id int64, in UpdateInput) (model.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.items[id]
	if !ok {
		return model.Todo{}, ErrNotFound
	}
	if in.Title != nil {
		t.Title = *in.Title
	}
	if in.Done != nil {
		t.Done = *in.Done
	}
	t.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	m.items[id] = t
	return t, nil
}

func (m *Memory) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return ErrNotFound
	}
	delete(m.items, id)
	return nil
}
