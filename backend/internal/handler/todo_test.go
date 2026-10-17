package handler

import (
	"net/http"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	rec := call(newApp(t), "GET", "/health", "", nil) // ログイン不要
	if rec.Code != 200 || strings.TrimSpace(rec.Body.String()) != `{"status":"ok"}` {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
}

func TestCreate(t *testing.T) {
	long := strings.Repeat("あ", 101)
	ok100 := strings.Repeat("あ", 100)
	tests := []struct {
		name     string
		body     string
		wantCode int
		wantErr  string
	}{
		{"正常", `{"title":"牛乳を買う"}`, 201, ""},
		{"前後の空白は除去", `{"title":"  牛乳  "}`, 201, ""},
		{"100文字", `{"title":"` + ok100 + `"}`, 201, ""},
		{"空タイトル", `{"title":""}`, 400, "VALIDATION_ERROR"},
		{"空白のみ", `{"title":"   "}`, 400, "VALIDATION_ERROR"},
		{"101文字", `{"title":"` + long + `"}`, 400, "VALIDATION_ERROR"},
		{"titleなし", `{}`, 400, "VALIDATION_ERROR"},
		{"型が違う", `{"title":123}`, 400, "INVALID_JSON"},
		{"壊れたJSON", `{`, 400, "INVALID_JSON"},
		{"空のボディ", ``, 400, "INVALID_JSON"},
		{"未知のフィールド", `{"title":"a","x":1}`, 400, "INVALID_JSON"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := newApp(t)
			c := signup(t, mux, "a@example.com")
			rec := call(mux, "POST", "/todos", tt.body, c)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body)
			}
			if tt.wantErr != "" && errorCode(t, rec) != tt.wantErr {
				t.Errorf("code = %q, want %q", errorCode(t, rec), tt.wantErr)
			}
		})
	}
}

func TestCreateResponse(t *testing.T) {
	mux := newApp(t)
	c := signup(t, mux, "a@example.com")
	rec := call(mux, "POST", "/todos", `{"title":"  牛乳  "}`, c)

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q", got)
	}
	todo := decodeTodo(t, rec)
	if todo.ID == 0 || todo.Title != "牛乳" || todo.Done || todo.CreatedAt == "" || todo.UpdatedAt == "" {
		t.Errorf("unexpected todo: %+v", todo)
	}
	if got := rec.Header().Get("Location"); got != "/todos/1" {
		t.Errorf("Location = %q", got)
	}
}

func TestListEmptyIsArray(t *testing.T) {
	mux := newApp(t)
	c := signup(t, mux, "a@example.com")
	rec := call(mux, "GET", "/todos", "", c)
	if rec.Code != 200 || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
}

func TestListOrderAndFilter(t *testing.T) {
	mux := newApp(t)
	c := signup(t, mux, "a@example.com")
	call(mux, "POST", "/todos", `{"title":"first"}`, c)
	call(mux, "POST", "/todos", `{"title":"second"}`, c)
	call(mux, "PATCH", "/todos/1", `{"done":true}`, c) // first を完了に

	tests := []struct {
		query     string
		wantCode  int
		wantTitle []string
	}{
		{"", 200, []string{"second", "first"}}, // 新しい順
		{"?status=all", 200, []string{"second", "first"}},
		{"?status=active", 200, []string{"second"}},
		{"?status=done", 200, []string{"first"}},
		{"?status=abc", 400, nil},
		{"?status=ACTIVE", 400, nil},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			rec := call(mux, "GET", "/todos"+tt.query, "", c)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if tt.wantCode != 200 {
				if errorCode(t, rec) != "VALIDATION_ERROR" {
					t.Errorf("code = %q", errorCode(t, rec))
				}
				return
			}
			todos := decodeTodos(t, rec)
			if len(todos) != len(tt.wantTitle) {
				t.Fatalf("got %+v, want %v", todos, tt.wantTitle)
			}
			for i, w := range tt.wantTitle {
				if todos[i].Title != w {
					t.Errorf("[%d] = %q, want %q", i, todos[i].Title, w)
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	mux := newApp(t)
	c := signup(t, mux, "a@example.com")
	call(mux, "POST", "/todos", `{"title":"a"}`, c)

	tests := []struct {
		name string
		path string
		want int
	}{
		{"存在する", "/todos/1", 200},
		{"存在しない", "/todos/999", 404},
		{"IDが文字列", "/todos/abc", 400},
		{"IDが0", "/todos/0", 400},
		{"IDが負数", "/todos/-1", 400},
		{"IDが巨大", "/todos/99999999999999999999", 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := call(mux, "GET", tt.path, "", c); rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestPatch(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		body      string
		wantCode  int
		wantErr   string
		wantTitle string
		wantDone  bool
	}{
		{"完了にする", "/todos/1", `{"done":true}`, 200, "", "orig", true},
		{"タイトル編集", "/todos/1", `{"title":"新"}`, 200, "", "新", false},
		{"両方", "/todos/1", `{"title":"新","done":true}`, 200, "", "新", true},
		{"空のボディ{}", "/todos/1", `{}`, 400, "VALIDATION_ERROR", "", false},
		{"空タイトル", "/todos/1", `{"title":""}`, 400, "VALIDATION_ERROR", "", false},
		{"101文字", "/todos/1", `{"title":"` + strings.Repeat("あ", 101) + `"}`, 400, "VALIDATION_ERROR", "", false},
		{"doneの型違い", "/todos/1", `{"done":"yes"}`, 400, "INVALID_JSON", "", false},
		{"壊れたJSON", "/todos/1", `{`, 400, "INVALID_JSON", "", false},
		{"存在しないID", "/todos/999", `{"done":true}`, 404, "NOT_FOUND", "", false},
		{"IDが文字列", "/todos/abc", `{"done":true}`, 400, "INVALID_JSON", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := newApp(t)
			c := signup(t, mux, "a@example.com")
			call(mux, "POST", "/todos", `{"title":"orig"}`, c)

			rec := call(mux, "PATCH", tt.path, tt.body, c)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantCode, rec.Body)
			}
			if tt.wantErr != "" && errorCode(t, rec) != tt.wantErr {
				t.Errorf("code = %q, want %q", errorCode(t, rec), tt.wantErr)
			}
			if tt.wantCode == 200 {
				got := decodeTodo(t, rec)
				if got.Title != tt.wantTitle || got.Done != tt.wantDone {
					t.Errorf("got %+v", got)
				}
			}
		})
	}
}

func TestPatchToggleBack(t *testing.T) {
	mux := newApp(t)
	c := signup(t, mux, "a@example.com")
	call(mux, "POST", "/todos", `{"title":"a"}`, c)
	call(mux, "PATCH", "/todos/1", `{"done":true}`, c)
	rec := call(mux, "PATCH", "/todos/1", `{"done":false}`, c)
	if got := decodeTodo(t, rec); got.Done {
		t.Errorf("done should be false again: %+v", got)
	}
}

func TestDelete(t *testing.T) {
	mux := newApp(t)
	c := signup(t, mux, "a@example.com")
	call(mux, "POST", "/todos", `{"title":"a"}`, c)

	if rec := call(mux, "DELETE", "/todos/abc", "", c); rec.Code != 400 {
		t.Errorf("bad id: %d", rec.Code)
	}
	if rec := call(mux, "DELETE", "/todos/1", "", c); rec.Code != 204 || rec.Body.Len() != 0 {
		t.Errorf("delete: %d %q", rec.Code, rec.Body)
	}
	if rec := call(mux, "GET", "/todos/1", "", c); rec.Code != 404 {
		t.Errorf("get after delete: %d", rec.Code)
	}
	if rec := call(mux, "GET", "/todos", "", c); strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Errorf("list after delete: %s", rec.Body)
	}
	if rec := call(mux, "DELETE", "/todos/1", "", c); rec.Code != 404 {
		t.Errorf("second delete: %d", rec.Code)
	}
}

func TestRoutingErrors(t *testing.T) {
	mux := newApp(t)
	c := signup(t, mux, "a@example.com")
	if rec := call(mux, "GET", "/unknown", "", c); rec.Code != http.StatusNotFound {
		t.Errorf("unknown path: %d", rec.Code)
	}
	if rec := call(mux, "PUT", "/todos/1", `{}`, c); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("PUT: %d", rec.Code)
	}
}

func TestFullFlow(t *testing.T) {
	mux := newApp(t)
	c := signup(t, mux, "a@example.com")

	call(mux, "POST", "/todos", `{"title":"牛乳"}`, c)
	call(mux, "PATCH", "/todos/1", `{"done":true}`, c)

	if got := decodeTodos(t, call(mux, "GET", "/todos?status=done", "", c)); len(got) != 1 {
		t.Errorf("done list: %+v", got)
	}
	if got := decodeTodos(t, call(mux, "GET", "/todos?status=active", "", c)); len(got) != 0 {
		t.Errorf("active list: %+v", got)
	}

	call(mux, "PATCH", "/todos/1", `{"title":"牛乳とパン"}`, c)
	if got := decodeTodo(t, call(mux, "GET", "/todos/1", "", c)); got.Title != "牛乳とパン" || !got.Done {
		t.Errorf("after edit: %+v", got)
	}
}
