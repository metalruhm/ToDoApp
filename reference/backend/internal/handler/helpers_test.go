package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"todoapp/backend/internal/store"
)

// newApp は、テストごとに新しい SQLite（一時ファイル）で組み立てたアプリを返す。
func newApp(t *testing.T) *http.ServeMux {
	t.Helper()
	st, err := store.NewSQLite(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	a := NewAuthHandler(st, false)
	a.Register(mux)
	NewTodoHandler(st).Register(mux, a.RequireAuth)
	return mux
}

func call(mux *http.ServeMux, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func sessionCookie(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == cookieName {
			return c
		}
	}
	return nil
}

// signup はユーザーを登録し、ログイン状態の Cookie を返す。
func signup(t *testing.T, mux *http.ServeMux, email string) *http.Cookie {
	t.Helper()
	rec := call(mux, "POST", "/auth/register", `{"email":"`+email+`","password":"password123"}`, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", rec.Code, rec.Body)
	}
	c := sessionCookie(rec)
	if c == nil {
		t.Fatal("no session cookie")
	}
	return c
}

type todoJSON struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func decodeTodo(t *testing.T, rec *httptest.ResponseRecorder) todoJSON {
	t.Helper()
	var todo todoJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &todo); err != nil {
		t.Fatalf("decode %q: %v", rec.Body, err)
	}
	return todo
}

func decodeTodos(t *testing.T, rec *httptest.ResponseRecorder) []todoJSON {
	t.Helper()
	var todos []todoJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &todos); err != nil {
		t.Fatalf("decode %q: %v", rec.Body, err)
	}
	return todos
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body %q: %v", rec.Body, err)
	}
	return body.Error.Code
}
