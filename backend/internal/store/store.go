package store

import (
	"context"
	"errors"

	"todoapp/backend/internal/model"
)

var ErrNotFound = errors.New("not found")

// UpdateInput は PATCH 用。nil のフィールドは「変更しない」。
type UpdateInput struct {
	Title *string
	Done  *bool
}

type TodoStore interface {
	List(ctx context.Context, status string) ([]model.Todo, error) // status: all|active|done
	Get(ctx context.Context, id int64) (model.Todo, error)
	Create(ctx context.Context, title string) (model.Todo, error)
	Update(ctx context.Context, id int64, in UpdateInput) (model.Todo, error)
	Delete(ctx context.Context, id int64) error
}
