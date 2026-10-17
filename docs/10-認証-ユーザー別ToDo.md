# Section 10: 認証・ユーザー別ToDo（発展）

> 📖 **詰まったら**：各Stepの答え（コード・コマンド）はこのファイル末尾の「付録: 各Stepの答え」にあります。まず自分で書いてから見ましょう。

## 1. 目的と完成の定義

**目的**: ユーザー登録・ログインを実装し、タスクをユーザーごとに分ける。パスワードの扱い、トークン認証、認可の基本を学ぶ。大きめの機能をPull Requestで進める。

**完成の定義**:
- [ ] 登録・ログイン・ログアウトができる
- [ ] パスワードはハッシュ化して保存され、平文では保存・ログ出力されない
- [ ] 未ログインでは `/todos` が 401 を返す
- [ ] **他人のタスクを取得/更新/削除できない**（認可。他人のIDは404にする）
- [ ] 認証まわりのテストがある
- [ ] PRを作成してセルフレビューし、マージした

## 2. 先にドキュメントを作る（AI駆動）

認証は設計ミスが危険なため、**実装前に必ずドキュメント化してレビュー**する。AIに依頼:
> requirements.md, architecture.md, api-spec.md, data-model.md を元に、認証機能の追加設計を作って。方式の候補（Cookieセッション / JWT）の比較と推奨、DB変更、API追加、セキュリティ上の注意点を含めて。コードは書かないで。

出力を既存4ドキュメントへの追記とし、`docs/auth-design.md` も作る。

## 3. 用語解説

| 用語 | 意味 |
|------|------|
| 認証 (Authentication) | 「誰か」を確かめる |
| 認可 (Authorization) | 「何をしてよいか」を判断する |
| ハッシュ化 | 元に戻せない変換。パスワード保存に使う |
| ソルト | ハッシュに混ぜるランダム値。bcryptは内部で自動付与 |
| bcrypt / argon2 | パスワード用のハッシュ方式 |
| JWT | 署名付きトークン。サーバーが状態を持たない |
| セッション + Cookie | サーバーがログイン状態を保持し、Cookieで識別 |
| HttpOnly Cookie | JavaScriptから読めないCookie（XSS対策） |
| CSRF | ログイン中のブラウザを悪用した不正リクエスト |
| ミドルウェア | リクエスト前後に共通処理を挟む仕組み |

## 4. 方式の選択（推奨）

| 方式 | 長所 | 短所 |
|------|------|------|
| **HttpOnly Cookie にトークン/セッションID**（推奨） | JSから盗めにくい | CORSで `credentials` 設定、CSRF対策が必要 |
| localStorage に JWT | 実装が簡単 | XSSで盗まれる |

→ 学習として**HttpOnly Cookie + サーバー側セッション or JWT**を推奨。選んだ理由は `auth-design.md` に書く。

## 5. 手順（コミット単位）

### Step 1: ドキュメント作成とレビュー
```
git switch -c docs/10-auth-design
```
`docs: add auth design (AI draft)` → レビュー → `docs: refine auth design after review` → mainへマージ。

### Step 2: DB変更（feat/10-auth-backend）
```
git switch -c feat/10-auth-backend
```
マイグレーション `002_create_users.sql`:
- `users`（id, email UNIQUE, password_hash, created_at）
- `todos.user_id`（外部キー、NOT NULL）

既存データの扱い（開発DBを作り直す/移行する）を決める。
`feat: add users table and todos.user_id`

### Step 3: 登録・ログイン（コミット数回）
- `POST /auth/register`: email形式・パスワード長（8文字以上等）を検証、bcryptでハッシュ化
- `POST /auth/login`: 照合し、Cookieをセット。**「メールが無い」と「パスワード違い」で同じメッセージ**を返す（アカウント存在の漏えい防止）
- `POST /auth/logout`
- `GET /auth/me`: ログイン中ユーザーの確認用
- ハッシュは `golang.org/x/crypto/bcrypt`

### Step 4: 認証ミドルウェア（コミット）
- Cookieを検証し、ユーザーIDを `context` に入れる
- 失敗は 401
- `/todos` 系をこのミドルウェアで保護

### Step 5: 認可（コミット）
- `store` の全メソッドに `userID` を渡す
- SQLに `WHERE user_id = ?` を付ける
- 他人のタスクIDは 404（存在を知らせない）

### Step 6: テスト（コミット）
- 登録/ログインの正常・異常
- 未ログインで401
- **ユーザーAのタスクをユーザーBが操作できない**こと（最重要）

### Step 7: フロント（feat/10-auth-frontend）
- ログイン/登録画面、ログアウトボタン
- `fetch` に `credentials: "include"`、Go側CORSに `Access-Control-Allow-Credentials: true`（このときオリジンは `*` 不可）
- 未ログインならログイン画面へ誘導、401を受けたら同様に処理

### Step 8: Pull Request（GitHub）
```
git push -u origin feat/10-auth-backend
```
GitHubでPRを作成 → **自分でレビュー**（Files changedタブ）。
セルフレビューの観点:
- [ ] パスワード/トークンがログやエラーに出ていない
- [ ] 認可漏れのエンドポイントがない
- [ ] `.env` や秘密鍵がコミットされていない
- [ ] テストが通る

AIにレビューさせたい場合: `/code-review` や `/security-review` を使える。指摘は鵜呑みにせず自分で判断する。
マージ後、ローカルを最新化:
```
git switch main && git pull && git branch -d feat/10-auth-backend
```

## 6. セキュリティチェックリスト
- [ ] パスワードを平文で保存・出力していない
- [ ] 署名鍵/セッション秘密は環境変数（`.env.example` には値を書かない）
- [ ] Cookieに `HttpOnly` `SameSite=Lax` 以上、本番は `Secure`
- [ ] ログイン失敗の回数制限を検討（発展）
- [ ] 全APIで認可を確認した

## 7. Git練習ポイント
- 大きな機能を複数ブランチ・複数PRに分割
- PR説明の書き方（目的・変更点・確認方法・スクリーンショット）
- `git pull` の挙動（fetch + merge）と、`git fetch` で状況確認する習慣

## 8. つまずきやすい点
- Cookieがブラウザに保存されない → `credentials`、CORSの Allow-Credentials、Origin 指定を確認
- `localhost` の別ポートでCookieが共有される/されない挙動（ポートは区別されない。オリジンとは別）
- パスワード検証で `bcrypt.CompareHashAndPassword` の戻り値の扱い
- 既存の `todos` に `user_id NOT NULL` を足すと失敗 → 移行手順を決める

## 9. 自己チェック
1. 認証と認可の違いは？
2. パスワードをハッシュ化する理由と、暗号化との違いは？
3. なぜ他人のIDは403でなく404にするのか？
4. HttpOnly Cookie がXSS対策になる理由は？

## 10. 完了したら
`learning-log.md` に追記し、`git tag -a v0.3.0`。続けて **Section 11: Docker化とCI** へ（Section 11, 12 の詳細資料は必要になったら作成する）。

---

# 付録: 各Stepの答え

> まず自分で書いてみて、詰まったときに開く。コピペせず、**1行ずつ意味を説明できる**ようにする。
> 方式は「**サーバー側セッション + HttpOnly Cookie**」を採用した例。JWTより仕組みが単純で、ログアウト時にサーバー側で無効化できる。
> 設計ドキュメント（auth-design.md）には、この選択理由を書いておく。

<details>
<summary>auth-design.md に書く内容の答え（Step 1）</summary>

| 項目 | 決定 | 理由 |
|---|---|---|
| 方式 | ランダムなセッショントークン + HttpOnly Cookie | JSから読めない。サーバーで無効化できる |
| トークン保存 | DBには**ハッシュ（SHA-256）**だけ保存 | DBが漏れてもCookieを偽造できない |
| パスワード | bcrypt、8〜72バイト | 72バイトを超える部分は無視される仕様のため上限を設ける |
| 有効期限 | 7日 | 学習用。短くしてもよい |
| Cookie属性 | `HttpOnly` `SameSite=Lax` `Path=/`、本番のみ `Secure` | XSS・CSRF対策 |
| 認可 | 全ての `todos` クエリに `user_id` 条件。他人のIDは404 | 存在を知らせない |
| エンドポイント | `POST /auth/register` `POST /auth/login` `POST /auth/logout` `GET /auth/me` | - |
| エラー | ログイン失敗は常に同じ文言（401） | アカウントの存在を漏らさない |
</details>

<details>
<summary>Step 2 の答え: DB変更とマイグレーション管理</summary>

**問題**: `Migrate` は今まで毎回すべてのSQLを実行していた。`DROP TABLE` を含むマイグレーションを毎回流すとデータが消える。実行済みを記録する仕組みを足す。

`store/sqlite.go` の `Migrate` を置き換える
```go
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
```
importに `"fmt"` `"strings"` を追加。`NewSQLite` に外部キーを有効にする1行も足す:
```go
if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil { ... }
```

`migrations/002_users_and_sessions.sql`
```sql
CREATE TABLE users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  email         TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at    TEXT NOT NULL
);

CREATE TABLE sessions (
  token_hash TEXT PRIMARY KEY,
  user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TEXT NOT NULL
);

-- 開発用データは破棄して作り直す（本番なら移行手順が必要）
DROP TABLE todos;
CREATE TABLE todos (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title      TEXT    NOT NULL CHECK (length(title) BETWEEN 1 AND 100),
  done       INTEGER NOT NULL DEFAULT 0 CHECK (done IN (0, 1)),
  created_at TEXT    NOT NULL,
  updated_at TEXT    NOT NULL
);
CREATE INDEX idx_todos_user_id ON todos(user_id);
```
</details>

<details>
<summary>Step 3 の答え: パスワード・トークン・ユーザー保存</summary>

依存追加: `cd backend && go get golang.org/x/crypto/bcrypt`

`internal/auth/auth.go`
```go
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// NewToken は Cookie に入れるトークンと、DBに保存するハッシュを返す。
func NewToken() (token, tokenHash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, HashToken(token), nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
```

`internal/store/account.go`
```go
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

func (s *SQLite) CreateSession(ctx context.Context, userID int64, tokenHash string, expires time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		tokenHash, userID, expires.UTC().Format(time.RFC3339))
	return err
}

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
```
（`expires_at` は UTC の RFC3339 文字列なので、文字列比較で時刻の前後が正しく判定できる）
</details>

<details>
<summary>Step 3〜4 の答え: 登録・ログイン・ログアウト・ミドルウェア（internal/handler/auth.go）</summary>

```go
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

const cookieName = "session"
const sessionTTL = 7 * 24 * time.Hour

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
	if _, err := mail.ParseAddress(c.Email); err != nil {
		return "メールアドレスの形式が正しくありません"
	}
	if n := len(c.Password); n < 8 || n > 72 { // bcryptは72バイトまで
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
	c.Email = strings.ToLower(strings.TrimSpace(c.Email))

	// 「メールがない」と「パスワード違い」を同じ応答にする
	fail := func() {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "メールアドレスまたはパスワードが正しくありません")
	}
	u, hash, err := a.accounts.UserByEmail(r.Context(), c.Email)
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
	http.SetCookie(w, a.cookie("", -1)) // Cookieを削除
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
		HttpOnly: true,           // JavaScriptから読めない
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
```
`main.go` での配線
```go
auth := handler.NewAuthHandler(st, os.Getenv("COOKIE_SECURE") == "true")
auth.Register(mux)
todos := handler.NewTodoHandler(st)
todos.Register(mux, auth.RequireAuth) // 全 /todos を保護
```
`TodoHandler.Register` は `func (h *TodoHandler) Register(mux *http.ServeMux, protect func(http.HandlerFunc) http.HandlerFunc)` に変更し、各ルートを `protect(h.list)` のように包む（`/health` は包まない）。
</details>

<details>
<summary>Step 5 の答え: 認可（user_id で絞る）</summary>

`TodoStore` の全メソッドの第2引数に `userID int64` を追加する。

```go
type TodoStore interface {
	List(ctx context.Context, userID int64, status string) ([]model.Todo, error)
	Get(ctx context.Context, userID, id int64) (model.Todo, error)
	Create(ctx context.Context, userID int64, title string) (model.Todo, error)
	Update(ctx context.Context, userID, id int64, in UpdateInput) (model.Todo, error)
	Delete(ctx context.Context, userID, id int64) error
}
```
メモリ実装はここで役目を終えたので削除し、テストはSQLiteのみで行う（残す場合は `userID` ごとにmapを分ける）。

SQLite実装の変更点（**全クエリに `user_id` 条件**）
```go
// List
q := "SELECT " + columns + " FROM todos WHERE user_id = ?"
switch status {
case "active":
	q += " AND done = 0"
case "done":
	q += " AND done = 1"
}
q += " ORDER BY created_at DESC, id DESC"
rows, err := s.db.QueryContext(ctx, q, userID)

// Get
s.db.QueryRowContext(ctx, "SELECT "+columns+" FROM todos WHERE id = ? AND user_id = ?", id, userID)

// Create
"INSERT INTO todos (user_id, title, done, created_at, updated_at) VALUES (?, ?, 0, ?, ?)"
// 引数: userID, title, ts, ts

// Update: WHERE の部分と args
" WHERE id = ? AND user_id = ?"   // args = append(args, id, userID)

// Delete
"DELETE FROM todos WHERE id = ? AND user_id = ?", id, userID
```
他人のIDは「行が見つからない」→ `ErrNotFound` → 404 になる。**403にしない**ので存在を知らせない。

ハンドラーでは `userID(r)` を渡す
```go
todos, err := h.store.List(r.Context(), userID(r), status)
```
`model.Todo` に `UserID int64 `json:"-"`` を足してもよい（JSONには出さない）。
</details>

<details>
<summary>Step 6 の答え: テスト（最重要は「他人のタスクを触れない」）</summary>

`internal/handler/auth_test.go`
```go
package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"todoapp/backend/internal/store"
)

func newApp(t *testing.T) *http.ServeMux {
	t.Helper()
	st, err := store.NewSQLite(filepath.Join(t.TempDir(), "t.db"))
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

// 登録してセッションCookieを返す
func signup(t *testing.T, mux *http.ServeMux, email string) *http.Cookie {
	t.Helper()
	rec := call(mux, "POST", "/auth/register", `{"email":"`+email+`","password":"password123"}`, nil)
	if rec.Code != 201 {
		t.Fatalf("register: %d %s", rec.Code, rec.Body)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == cookieName {
			return c
		}
	}
	t.Fatal("no session cookie")
	return nil
}

func TestUnauthenticated(t *testing.T) {
	rec := call(newApp(t), "GET", "/todos", "", nil)
	if rec.Code != 401 {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestRegisterValidation(t *testing.T) {
	mux := newApp(t)
	tests := []struct {
		name, body string
		want       int
	}{
		{"短いパスワード", `{"email":"a@example.com","password":"short"}`, 400},
		{"不正なメール", `{"email":"nope","password":"password123"}`, 400},
		{"正常", `{"email":"a@example.com","password":"password123"}`, 201},
		{"重複", `{"email":"a@example.com","password":"password123"}`, 409},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := call(mux, "POST", "/auth/register", tt.body, nil); rec.Code != tt.want {
				t.Errorf("got %d, want %d", rec.Code, tt.want)
			}
		})
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
		if rec := call(mux, tt.method, "/todos/1", tt.body, bob); rec.Code != 404 {
			t.Errorf("%s: want 404, got %d", tt.method, rec.Code)
		}
	}
	if rec := call(mux, "GET", "/todos/1", "", alice); rec.Code != 200 {
		t.Errorf("alice should still see it: %d", rec.Code)
	}
}
```
`go test ./...` で、既存のテスト（Section 5のhandlerテスト）が `Register` の引数変更で壊れる。`newMux` も `newApp` を使うか、ログイン済みCookieを付けるように直す。
</details>

<details>
<summary>Step 7 の答え: フロントの認証対応</summary>

**Go側CORS**（`cors.go` に1行追加。`Allow-Origin` は `*` ではなく具体的なオリジンにする）
```go
w.Header().Set("Access-Control-Allow-Credentials", "true")
```

**`src/lib/api.ts`**（変更点のみ）
```ts
export class ApiError extends Error {
  constructor(message: string, public status: number) {
    super(message);
  }
}

// request() の中
res = await fetch(`${BASE_URL}${path}`, {
  ...init,
  credentials: "include", // Cookieを送受信する
  headers: { "Content-Type": "application/json", ...init?.headers },
});
// ...
if (!res.ok) {
  const body = await res.json().catch(() => null);
  throw new ApiError(body?.error?.message ?? `エラーが発生しました (${res.status})`, res.status);
}

export type User = { id: number; email: string };
export const register = (email: string, password: string) =>
  request<User>("/auth/register", { method: "POST", body: JSON.stringify({ email, password }) });
export const login = (email: string, password: string) =>
  request<User>("/auth/login", { method: "POST", body: JSON.stringify({ email, password }) });
export const logout = () => request<void>("/auth/logout", { method: "POST" });
export const me = () => request<User>("/auth/me");
```

**`src/app/login/page.tsx`**
```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { login, register } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await (mode === "login" ? login : register)(email, password);
      router.push("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "エラーが発生しました");
    }
  }

  return (
    <main className="mx-auto max-w-sm px-4 py-8">
      <h1 className="mb-6 text-2xl font-bold">{mode === "login" ? "ログイン" : "新規登録"}</h1>
      <form onSubmit={handleSubmit} className="space-y-3">
        <div>
          <label htmlFor="email">メールアドレス</label>
          <input id="email" type="email" autoComplete="email" value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full rounded-lg border px-3 py-2" />
        </div>
        <div>
          <label htmlFor="password">パスワード（8文字以上）</label>
          <input id="password" type="password"
            autoComplete={mode === "login" ? "current-password" : "new-password"}
            value={password} onChange={(e) => setPassword(e.target.value)}
            className="w-full rounded-lg border px-3 py-2" />
        </div>
        {error && <p role="alert" className="text-sm text-red-600">{error}</p>}
        <button type="submit" className="w-full rounded-lg bg-indigo-600 py-2 text-white">
          {mode === "login" ? "ログイン" : "登録"}
        </button>
      </form>
      <button onClick={() => setMode(mode === "login" ? "register" : "login")}
        className="mt-4 text-sm text-indigo-600 underline">
        {mode === "login" ? "アカウントを作る" : "ログインへ戻る"}
      </button>
    </main>
  );
}
```

**`src/app/page.tsx`**（変更点）
```tsx
const router = useRouter();

// 一覧取得の .catch の中
.catch((e: Error) => {
  if (e instanceof ApiError && e.status === 401) {
    router.push("/login"); // 未ログインならログイン画面へ
    return;
  }
  setError(e.message);
})

// ログアウトボタン
<button onClick={async () => { await logout(); router.push("/login"); }}>ログアウト</button>
```
importに `useRouter`（`next/navigation`）、`ApiError`、`logout` を追加。
</details>

<details>
<summary>Step 8 の答え: Pull Request の流れ</summary>

```
git push -u origin feat/10-auth-backend
```
GitHub に表示される「Compare & pull request」から作成する。

PRの説明の例
```markdown
## 目的
ユーザー登録・ログインと、ユーザーごとのToDoを実現する。

## 変更点
- users / sessions テーブルを追加、todos に user_id を追加
- /auth/register, /login, /logout, /me を追加
- /todos をログイン必須にし、全クエリを user_id で絞る

## 確認方法
- `go test ./...`
- 2ユーザーを作り、互いのタスクが見えないことを確認

## 注意
- 002 マイグレーションで既存の todos を破棄する（開発用データのみ）
```
セルフレビュー: GitHub の **Files changed** タブで差分を読み、気づいた点に行コメントを付ける → 修正をコミット → 問題なければ **Merge**（Squash でもよい）。

マージ後
```
git switch main
git pull
git branch -d feat/10-auth-backend
git tag -a v0.3.0 -m "Add authentication"
git push origin main --tags
```
</details>
