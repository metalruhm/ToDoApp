package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"todoapp/backend/internal/auth"
	"todoapp/backend/internal/store"
)

const (
	cookieName = "session"
	sessionTTL = 7 * 24 * time.Hour
)

type AccountStore interface {
	CreateUser(ctx context.Context, email, passwordHash string) (store.User, error)
	UserByEmail(ctx context.Context, email string) (store.User, string, error)
	CreateSession(ctx context.Context, userID int64, tokenHash string, expires time.Time) error
	UserBySession(ctx context.Context, tokenHash string) (store.User, error)
	DeleteSession(ctx context.Context, tokenHash string) error
}

type AuthHandler struct {
	accounts     AccountStore
	secureCookie bool // 本番(HTTPS)では true
}

func NewAuthHandler(a AccountStore, secureCookie bool) *AuthHandler {
	return &AuthHandler{accounts: a, secureCookie: secureCookie}
}

func (a *AuthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", a.register)
	mux.HandleFunc("POST /auth/login", a.login)
	mux.HandleFunc("POST /auth/logout", a.logout)
	mux.HandleFunc("GET /auth/me", a.RequireAuth(a.me))
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *credentials) validate() string {
	c.Email = strings.ToLower(strings.TrimSpace(c.Email))
	if addr, err := mail.ParseAddress(c.Email); err != nil || addr.Address != c.Email {
		return "メールアドレスの形式が正しくありません"
	}
	if n := len(c.Password); n < auth.MinPasswordLength || n > auth.MaxPasswordLength {
		return "パスワードは8〜72文字で入力してください"
	}
	return ""
}

func (a *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := decodeJSON(w, r, &c); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの形式が正しくありません")
		return
	}
	if msg := c.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}
	hash, err := auth.HashPassword(c.Password)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	u, err := a.accounts.CreateUser(r.Context(), c.Email, hash)
	if errors.Is(err, store.ErrEmailTaken) {
		writeError(w, http.StatusConflict, "EMAIL_TAKEN", "このメールアドレスは登録済みです")
		return
	}
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if err := a.startSession(w, r, u.ID); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func (a *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := decodeJSON(w, r, &c); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの形式が正しくありません")
		return
	}
	email := strings.ToLower(strings.TrimSpace(c.Email))

	// 「メールがない」と「パスワード違い」を同じ応答にする（アカウントの存在を漏らさない）
	fail := func() {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "メールアドレスまたはパスワードが正しくありません")
	}
	u, hash, err := a.accounts.UserByEmail(r.Context(), email)
	if errors.Is(err, store.ErrNotFound) {
		fail()
		return
	}
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !auth.CheckPassword(hash, c.Password) {
		fail()
		return
	}
	if err := a.startSession(w, r, u.ID); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (a *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		if err := a.accounts.DeleteSession(r.Context(), auth.HashToken(c.Value)); err != nil {
			log.Printf("delete session: %v", err)
		}
	}
	http.SetCookie(w, a.cookie("", -1)) // Cookie を削除
	w.WriteHeader(http.StatusNoContent)
}

func (a *AuthHandler) me(w http.ResponseWriter, r *http.Request) {
	u, err := a.accounts.UserBySession(r.Context(), currentTokenHash(r))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "ログインが必要です")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// ---------- セッション / Cookie ----------

func (a *AuthHandler) startSession(w http.ResponseWriter, r *http.Request, userID int64) error {
	token, tokenHash, err := auth.NewToken()
	if err != nil {
		return err
	}
	if err := a.accounts.CreateSession(r.Context(), userID, tokenHash, time.Now().Add(sessionTTL)); err != nil {
		return err
	}
	http.SetCookie(w, a.cookie(token, int(sessionTTL.Seconds())))
	return nil
}

func (a *AuthHandler) cookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,           // JavaScript から読めない
		Secure:   a.secureCookie, // 本番(HTTPS)のみ
		SameSite: http.SameSiteLaxMode,
	}
}

func currentTokenHash(r *http.Request) string {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return auth.HashToken(c.Value)
}

// ---------- ミドルウェア ----------

type ctxKey struct{}

// RequireAuth はログイン済みのときだけ next を呼び、ユーザーIDを context に入れる。
func (a *AuthHandler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := a.accounts.UserBySession(r.Context(), currentTokenHash(r))
		if err != nil {
			if !errors.Is(err, store.ErrNotFound) {
				log.Printf("session lookup: %v", err)
			}
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "ログインが必要です")
			return
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, u.ID)
		next(w, r.WithContext(ctx))
	}
}

func userID(r *http.Request) int64 {
	id, _ := r.Context().Value(ctxKey{}).(int64)
	return id
}
