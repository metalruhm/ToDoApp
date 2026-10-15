package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"todoapp/backend/internal/store"
)

// テストごとに新しい store を作り、テスト同士が影響しないようにする
func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	NewTodoHandler(store.NewMemory()).Register(mux)
	return mux
}

func do(mux *http.ServeMux, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"正常", `{"title":"牛乳"}`, http.StatusCreated},
		{"空タイトル", `{"title":""}`, http.StatusBadRequest},
		{"空白のみ", `{"title":"   "}`, http.StatusBadRequest},
		{"壊れたJSON", `{`, http.StatusBadRequest},
		{"未知のフィールド", `{"title":"a","x":1}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(newMux(), "POST", "/todos", tt.body)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, tt.want, rec.Body)
			}
		})
	}
}

func TestListEmptyIsArray(t *testing.T) {
	rec := do(newMux(), "GET", "/todos", "")
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Errorf("want [], got %s", rec.Body)
	}
}

func TestNotFoundAndBadID(t *testing.T) {
	mux := newMux()
	tests := []struct {
		name, method, path, body string
		want                     int
	}{
		{"GET 存在しない", "GET", "/todos/999", "", 404},
		{"PATCH 存在しない", "PATCH", "/todos/999", `{"done":true}`, 404},
		{"DELETE 存在しない", "DELETE", "/todos/999", "", 404},
		{"GET IDが文字列", "GET", "/todos/abc", "", 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(mux, tt.method, tt.path, tt.body)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestPatchEmptyBody(t *testing.T) {
	mux := newMux()
	do(mux, "POST", "/todos", `{"title":"a"}`)
	rec := do(mux, "PATCH", "/todos/1", `{}`)
	if rec.Code != 400 {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestFlow(t *testing.T) {
	mux := newMux()
	do(mux, "POST", "/todos", `{"title":"牛乳"}`)

	rec := do(mux, "PATCH", "/todos/1", `{"done":true}`)
	if rec.Code != 200 {
		t.Fatalf("patch status = %d", rec.Code)
	}
	var got struct {
		Done  bool   `json:"done"`
		Title string `json:"title"`
	}
	json.Unmarshal(rec.Body.Bytes(), &got)
	if !got.Done || got.Title != "牛乳" {
		t.Errorf("unexpected: %+v", got)
	}

	if rec := do(mux, "DELETE", "/todos/1", ""); rec.Code != 204 {
		t.Errorf("delete status = %d", rec.Code)
	}
}
