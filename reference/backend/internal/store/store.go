// Package store はデータの保存・取得を担当する。HTTPは知らない。
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

// TodoStore は、すべてのメソッドで userID を受け取り、他ユーザーのデータに触れない。
type TodoStore interface {
	List(ctx context.Context, userID int64, status string) ([]model.Todo, error) // status: all|active|done
	Get(ctx context.Context, userID, id int64) (model.Todo, error)
	Create(ctx context.Context, userID int64, title string) (model.Todo, error)
	Update(ctx context.Context, userID, id int64, in UpdateInput) (model.Todo, error)
	Delete(ctx context.Context, userID, id int64) error
}
