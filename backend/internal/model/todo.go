// Package model はアプリのデータ型と入力検証を持つ。HTTPやDBは知らない。
package model

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxTitleLength = 100

type Todo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

var ErrInvalidTitle = errors.New("タイトルは1〜100文字で入力してください")

// NormalizeTitle は前後の空白を除いたタイトルを返す。1〜100文字でなければエラー。
func NormalizeTitle(s string) (string, error) {
	t := strings.TrimSpace(s)
	n := utf8.RuneCountInString(t) // バイト数ではなく文字数
	if n < 1 || n > MaxTitleLength {
		return "", ErrInvalidTitle
	}
	return t, nil
}
