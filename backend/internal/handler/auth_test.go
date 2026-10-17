package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUnauthenticated(t *testing.T) {
	mux := newApp(t)
	for _, tt := range []struct{ method, path, body string }{
		{"GET", "/todos", ""},
		{"POST", "/todos", `{"title":"a"}`},
		{"GET", "/todos/1", ""},
		{"PATCH", "/todos/1", `{"done":true}`},
		{"DELETE", "/todos/1", ""},
		{"GET", "/auth/me", ""},
	} {
		rec := call(mux, tt.method, tt.path, tt.body, nil)
		if rec.Code != 401 || errorCode(t, rec) != "UNAUTHORIZED" {
			t.Errorf("%s %s: %d %s", tt.method, tt.path, rec.Code, rec.Body)
		}
	}
	// 不正なトークンも401
	rec := call(mux, "GET", "/todos", "", &http.Cookie{Name: cookieName, Value: "invalid"})
	if rec.Code != 401 {
		t.Errorf("invalid token: %d", rec.Code)
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"正常", `{"email":"a@example.com","password":"password123"}`, 201},
		{"メールは小文字化される", `{"email":"A@Example.com","password":"password123"}`, 201},
		{"短いパスワード", `{"email":"a@example.com","password":"short"}`, 400},
		{"長すぎるパスワード", `{"email":"a@example.com","password":"` + strings.Repeat("a", 73) + `"}`, 400},
		{"不正なメール", `{"email":"nope","password":"password123"}`, 400},
		{"表示名つきメールは拒否", `{"email":"Bob <b@example.com>","password":"password123"}`, 400},
		{"壊れたJSON", `{`, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := call(newApp(t), "POST", "/auth/register", tt.body, nil)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d (%s)", rec.Code, tt.want, rec.Body)
			}
		})
	}
}

func TestRegisterDuplicate(t *testing.T) {
	mux := newApp(t)
	signup(t, mux, "a@example.com")
	rec := call(mux, "POST", "/auth/register", `{"email":"A@example.com","password":"password123"}`, nil)
	if rec.Code != 409 || errorCode(t, rec) != "EMAIL_TAKEN" {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
}

func TestRegisterDoesNotLeakPassword(t *testing.T) {
	mux := newApp(t)
	rec := call(mux, "POST", "/auth/register", `{"email":"a@example.com","password":"password123"}`, nil)
	if strings.Contains(rec.Body.String(), "password") {
		t.Errorf("response must not contain password info: %s", rec.Body)
	}
}

func TestSessionCookieAttributes(t *testing.T) {
	mux := newApp(t)
	c := signup(t, mux, "a@example.com")
	if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.Path != "/" || c.MaxAge <= 0 {
		t.Errorf("unexpected cookie: %+v", c)
	}
}

func TestLogin(t *testing.T) {
	mux := newApp(t)
	signup(t, mux, "a@example.com")

	rec := call(mux, "POST", "/auth/login", `{"email":"a@example.com","password":"password123"}`, nil)
	if rec.Code != 200 {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}
	c := sessionCookie(rec)
	if c == nil {
		t.Fatal("no cookie")
	}
	if me := call(mux, "GET", "/auth/me", "", c); me.Code != 200 || !strings.Contains(me.Body.String(), "a@example.com") {
		t.Errorf("me: %d %s", me.Code, me.Body)
	}
}

func TestLoginFailureIsUniform(t *testing.T) {
	mux := newApp(t)
	signup(t, mux, "a@example.com")
	wrongPw := call(mux, "POST", "/auth/login", `{"email":"a@example.com","password":"wrongwrong"}`, nil)
	noUser := call(mux, "POST", "/auth/login", `{"email":"x@example.com","password":"password123"}`, nil)
	if wrongPw.Code != 401 || noUser.Code != 401 || wrongPw.Body.String() != noUser.Body.String() {
		t.Errorf("responses differ: %d %q / %d %q", wrongPw.Code, wrongPw.Body, noUser.Code, noUser.Body)
	}
	if sessionCookie(wrongPw) != nil || sessionCookie(noUser) != nil {
		t.Error("failed login must not set a session cookie")
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	mux := newApp(t)
	c := signup(t, mux, "a@example.com")

	rec := call(mux, "POST", "/auth/logout", "", c)
	if rec.Code != 204 {
		t.Fatalf("logout: %d", rec.Code)
	}
	if cleared := sessionCookie(rec); cleared == nil || cleared.MaxAge >= 0 {
		t.Errorf("cookie should be cleared: %+v", cleared)
	}
	// 古い Cookie を持っていても、サーバー側で無効化されている
	if rec := call(mux, "GET", "/todos", "", c); rec.Code != 401 {
		t.Errorf("after logout: %d", rec.Code)
	}
}

// 最重要: ユーザーBはユーザーAのタスクを見られず、変更・削除もできない
func TestCannotAccessOthersTodos(t *testing.T) {
	mux := newApp(t)
	alice := signup(t, mux, "alice@example.com")
	bob := signup(t, mux, "bob@example.com")

	if rec := call(mux, "POST", "/todos", `{"title":"aliceの秘密"}`, alice); rec.Code != 201 {
		t.Fatalf("create: %d", rec.Code)
	}

	if rec := call(mux, "GET", "/todos", "", bob); strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Errorf("bob sees alice's todos: %s", rec.Body)
	}
	for _, tt := range []struct{ method, body string }{
		{"GET", ""}, {"PATCH", `{"done":true}`}, {"DELETE", ""},
	} {
		// 403 ではなく 404（存在を知らせない）
		if rec := call(mux, tt.method, "/todos/1", tt.body, bob); rec.Code != 404 {
			t.Errorf("%s: want 404, got %d", tt.method, rec.Code)
		}
	}
	got := call(mux, "GET", "/todos/1", "", alice)
	if got.Code != 200 || decodeTodo(t, got).Done {
		t.Errorf("alice's todo was affected: %d %s", got.Code, got.Body)
	}
}

func TestCORS(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	h := CORS("http://localhost:3000", next)

	t.Run("ヘッダーが付く", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
		if rec.Code != 200 {
			t.Errorf("status = %d", rec.Code)
		}
		for k, want := range map[string]string{
			"Access-Control-Allow-Origin":      "http://localhost:3000",
			"Access-Control-Allow-Credentials": "true",
		} {
			if got := rec.Header().Get(k); got != want {
				t.Errorf("%s = %q, want %q", k, got, want)
			}
		}
	})
	t.Run("OPTIONSは204で次に渡さない", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("OPTIONS", "/todos", nil))
		if rec.Code != 204 {
			t.Errorf("status = %d", rec.Code)
		}
		if !strings.Contains(rec.Header().Get("Access-Control-Allow-Methods"), "PATCH") {
			t.Errorf("methods = %q", rec.Header().Get("Access-Control-Allow-Methods"))
		}
	})
}
