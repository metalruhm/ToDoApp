package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

var ErrEmailTaken = errors.New("email already registered")

func (s *SQLite) CreateUser(ctx context.Context, email, passwordHash string) (User, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users (email, password_hash, created_at) VALUES (?, ?, ?)`,
		email, passwordHash, now())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}
	id, err := res.LastInsertId()
	return User{ID: id, Email: email}, err
}

// UserByEmail はユーザーとパスワードハッシュを返す。
func (s *SQLite) UserByEmail(ctx context.Context, email string) (User, string, error) {
	var u User
	var hash string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash FROM users WHERE email = ?`, email).
		Scan(&u.ID, &u.Email, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, "", ErrNotFound
	}
	return u, hash, err
}

// CreateSession はセッションを保存する。ついでに期限切れのセッションを掃除する。
func (s *SQLite) CreateSession(ctx context.Context, userID int64, tokenHash string, expires time.Time) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= ?`, now()); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		tokenHash, userID, expires.UTC().Format(time.RFC3339))
	return err
}

// UserBySession は有効なセッションのユーザーを返す。期限切れ・不明は ErrNotFound。
func (s *SQLite) UserBySession(ctx context.Context, tokenHash string) (User, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		`SELECT u.id, u.email FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = ? AND s.expires_at > ?`, tokenHash, now()).
		Scan(&u.ID, &u.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

func (s *SQLite) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, tokenHash)
	return err
}
