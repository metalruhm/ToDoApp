package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite" // ドライバー登録のためだけに import

	"todoapp/backend/internal/model"
	"todoapp/backend/migrations"
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite は書き込みが1つずつ。"database is locked" を避ける
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return &SQLite{db: db}, nil
}

func (s *SQLite) Close() error { return s.db.Close() }

// Migrate は migrations/*.sql をファイル名順に、未実行のものだけ実行する。
func (s *SQLite) Migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY)`); err != nil {
		return err
	}

	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return err
	}
	names := []string{}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var n int
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			continue // 実行済み
		}
		b, err := migrations.FS.ReadFile(name)
		if err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, string(b)); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if _, err := s.db.ExecContext(ctx,
			`INSERT INTO schema_migrations (name) VALUES (?)`, name); err != nil {
			return err
		}
	}
	return nil
}

const columns = "id, title, done, created_at, updated_at"

// *sql.Row と *sql.Rows の両方で使えるようにする
type scanner interface {
	Scan(dest ...any) error
}

func scanTodo(sc scanner) (model.Todo, error) {
	var t model.Todo
	var done int
	var created, updated string
	if err := sc.Scan(&t.ID, &t.Title, &done, &created, &updated); err != nil {
		return model.Todo{}, err
	}
	t.Done = done == 1
	var err error
	if t.CreatedAt, err = time.Parse(time.RFC3339, created); err != nil {
		return model.Todo{}, err
	}
	if t.UpdatedAt, err = time.Parse(time.RFC3339, updated); err != nil {
		return model.Todo{}, err
	}
	return t, nil
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (s *SQLite) List(ctx context.Context, userID int64, status string) ([]model.Todo, error) {
	q := "SELECT " + columns + " FROM todos WHERE user_id = ?" // 連結しているのは固定文字列のみ
	switch status {
	case "active":
		q += " AND done = 0"
	case "done":
		q += " AND done = 1"
	}
	q += " ORDER BY created_at DESC, id DESC"

	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []model.Todo{} // 0件でも [] にする
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

func (s *SQLite) Get(ctx context.Context, userID, id int64) (model.Todo, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT "+columns+" FROM todos WHERE id = ? AND user_id = ?", id, userID)
	t, err := scanTodo(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Todo{}, ErrNotFound
	}
	return t, err
}

func (s *SQLite) Create(ctx context.Context, userID int64, title string) (model.Todo, error) {
	ts := now()
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO todos (user_id, title, done, created_at, updated_at) VALUES (?, ?, 0, ?, ?)",
		userID, title, ts, ts)
	if err != nil {
		return model.Todo{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.Todo{}, err
	}
	return s.Get(ctx, userID, id)
}

func (s *SQLite) Update(ctx context.Context, userID, id int64, in UpdateInput) (model.Todo, error) {
	sets := []string{"updated_at = ?"}
	args := []any{now()}
	if in.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *in.Title)
	}
	if in.Done != nil {
		done := 0
		if *in.Done {
			done = 1
		}
		sets = append(sets, "done = ?")
		args = append(args, done)
	}
	args = append(args, id, userID)

	res, err := s.db.ExecContext(ctx,
		"UPDATE todos SET "+strings.Join(sets, ", ")+" WHERE id = ? AND user_id = ?", args...)
	if err != nil {
		return model.Todo{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.Todo{}, ErrNotFound
	}
	return s.Get(ctx, userID, id)
}

func (s *SQLite) Delete(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx,
		"DELETE FROM todos WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
