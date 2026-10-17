# Section 4: Go — ToDo APIをメモリ上で実装

> 📖 **詰まったら**：各Stepの答え（コード・コマンド）はこのファイル末尾の「付録: 各Stepの答え」にあります。まず自分で書いてから見ましょう。

## 1. 目的と完成の定義

**目的**: [api-spec.md](api-spec.md) どおりのCRUD APIを、まずメモリ上のデータで実装する。層分割（handler / model / store）を体験する。

**完成の定義**:
- [ ] api-spec.md の6エンドポイントがすべて動く（`curl` で確認）
- [ ] 検証エラー・404・不正JSONで仕様どおりのエラーJSONが返る
- [ ] 並行アクセスでデータが壊れない（`sync.Mutex` を使用）
- [ ] `handler` `model` `store` にパッケージが分かれている
- [ ] 機能ごとにコミットしている（目安6コミット以上）

## 2. 用語解説

| 用語 | 意味 |
|------|------|
| CRUD | Create / Read / Update / Delete |
| インターフェース | 「このメソッドを持つ型」という約束。実装を差し替え可能にする |
| ミューテックス | 同時アクセスから共有データを守る鍵 |
| ゴルーチン | Goの軽量スレッド。HTTPリクエストは並行で処理される |
| ステータスコード | 200/201/204/400/404/500 など結果を示す番号 |
| `context.Context` | キャンセルやタイムアウトを伝える仕組み |

## 3. 目標ディレクトリ

```
backend/internal/
├── model/todo.go          # Todo構造体、タイトル検証
├── store/store.go         # TodoStoreインターフェース、エラー(ErrNotFound)
├── store/memory.go        # メモリ実装
└── handler/todo.go        # HTTPハンドラー、JSONヘルパー
```

## 4. 手順（コミット単位）

### Step 1: ブランチ
```
git switch -c feat/04-todo-api
```

### Step 2: model（コミット1）
- `Todo` 構造体（[data-model.md](data-model.md) 参照）
- タイトル検証関数: 前後の空白除去、`utf8.RuneCountInString` で1〜100文字
- 検証失敗時のエラー型を決める

`feat: add Todo model and title validation`

### Step 3: store インターフェースとメモリ実装（コミット2）
- `TodoStore` インターフェース（architecture.md の方針）
- `ErrNotFound` を定義
- メモリ実装: `map[int64]Todo` + `nextID` + `sync.Mutex`
- `List` は新しい順に並べる（`sort.Slice`）

`feat: add in-memory todo store`

### Step 4: 一覧・追加（コミット3）
- `GET /todos`、`POST /todos`
- JSONレスポンス/エラーレスポンスの**共通ヘルパー**を作る（`writeJSON`, `writeError`）
- 0件は `[]`（`null`にならないよう注意）
- `POST` は 201 と `Location` ヘッダー

`feat: add list and create todo endpoints`

### Step 5: 取得（コミット4）
- `GET /todos/{id}`: `r.PathValue("id")` → `strconv.ParseInt`
- 数値でない → 400、見つからない → 404（`errors.Is`）

### Step 6: 更新（コミット5）
- `PATCH /todos/{id}`: ポインタ型フィールドで「送られたか」を判別（`*string`, `*bool`）
- ボディが空 `{}` → 400
- 更新で `updatedAt` を更新

### Step 7: 削除（コミット6）
- `DELETE /todos/{id}` → 204

### Step 8: 横断的な処理（コミット7）
- 本文サイズ制限（`http.MaxBytesReader`）
- `status` クエリ（`all|active|done`）
- 未知のフィールドを拒否するか検討（`DisallowUnknownFields`）
- `main.go` の配線（store → handler → mux）

### Step 9: 動作確認
api-spec.md の「curl での動作確認例」をすべて実行し、レスポンスを仕様と見比べる。

### Step 10: マージ
```
git switch main && git merge feat/04-todo-api
```

## 5. ヒント集
- ハンドラーはstoreの**具体型でなくインターフェース**を受け取る
- `json.NewDecoder(r.Body).Decode(&req)` のエラーは `INVALID_JSON`
- 500のときは詳細をレスポンスに出さず、`log.Printf` に出す
- 完成形は末尾の付録にある。まず自分で書き、詰まったらAIに質問してから付録を見る

## 6. Git練習ポイント
- 1機能1コミットで進め、`git log --oneline --graph` で履歴を眺める
- 途中で気づいた別の修正は `git stash` で退避するか、別コミットに分ける
- `git add -p` で変更の一部だけをステージする

## 7. つまずきやすい点
- `map` の並行書き込みで `fatal error: concurrent map writes` → Mutexを忘れている
- 空スライスが `null` になる → `[]Todo{}` で初期化
- `PATCH` で `done: false` を送っても無視される → `bool` ではなく `*bool` を使う
- ステータスコードは `WriteHeader` を**ボディ書き込み前**に呼ぶ

## 8. 自己チェック
1. なぜstoreをインターフェースにしたのか？
2. `*bool` を使う理由は？
3. 400と404はどう使い分けるか？
4. Mutexがないとなぜ壊れるのか？

## 9. 完了したら
`learning-log.md` に追記し、**Section 5: テスト** へ。

---

# 付録: 各Stepの答え

> まず自分で書いてみて、詰まったときに開く。コピペせず、**1行ずつ意味を説明できる**ようにする。
> モジュール名は `todoapp/backend`。importは `todoapp/backend/internal/...` になる。

<details>
<summary>Step 2 の答え: model（internal/model/todo.go）</summary>

```go
package model

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

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
	if n < 1 || n > 100 {
		return "", ErrInvalidTitle
	}
	return t, nil
}
```
</details>

<details>
<summary>Step 3 の答え: store（internal/store/store.go と memory.go）</summary>

`store.go`
```go
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

type TodoStore interface {
	List(ctx context.Context, status string) ([]model.Todo, error) // status: all|active|done
	Get(ctx context.Context, id int64) (model.Todo, error)
	Create(ctx context.Context, title string) (model.Todo, error)
	Update(ctx context.Context, id int64, in UpdateInput) (model.Todo, error)
	Delete(ctx context.Context, id int64) error
}
```

`memory.go`
```go
package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"todoapp/backend/internal/model"
)

type Memory struct {
	mu     sync.Mutex
	items  map[int64]model.Todo
	nextID int64
}

func NewMemory() *Memory {
	return &Memory{items: map[int64]model.Todo{}, nextID: 1}
}

func (m *Memory) List(ctx context.Context, status string) ([]model.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]model.Todo, 0, len(m.items)) // 空でも null にならない
	for _, t := range m.items {
		if status == "active" && t.Done {
			continue
		}
		if status == "done" && !t.Done {
			continue
		}
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		if !result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].CreatedAt.After(result[j].CreatedAt)
		}
		return result[i].ID > result[j].ID
	})
	return result, nil
}

func (m *Memory) Get(ctx context.Context, id int64) (model.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.items[id]
	if !ok {
		return model.Todo{}, ErrNotFound
	}
	return t, nil
}

func (m *Memory) Create(ctx context.Context, title string) (model.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC().Truncate(time.Second)
	t := model.Todo{ID: m.nextID, Title: title, CreatedAt: now, UpdatedAt: now}
	m.items[t.ID] = t
	m.nextID++
	return t, nil
}

func (m *Memory) Update(ctx context.Context, id int64, in UpdateInput) (model.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.items[id]
	if !ok {
		return model.Todo{}, ErrNotFound
	}
	if in.Title != nil {
		t.Title = *in.Title
	}
	if in.Done != nil {
		t.Done = *in.Done
	}
	t.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	m.items[id] = t
	return t, nil
}

func (m *Memory) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return ErrNotFound
	}
	delete(m.items, id)
	return nil
}
```
- `Lock` / `defer Unlock` … 同時アクセスから `map` を守る
- `make([]model.Todo, 0, ...)` … 0件でも JSON が `[]` になる
- map の走査順は不定なので、`sort.Slice` で並べる
</details>

<details>
<summary>Step 4〜7 の答え: handler（internal/handler/todo.go）</summary>

```go
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"todoapp/backend/internal/model"
	"todoapp/backend/internal/store"
)

type TodoHandler struct {
	store store.TodoStore // 具体型ではなくインターフェース
}

func NewTodoHandler(s store.TodoStore) *TodoHandler {
	return &TodoHandler{store: s}
}

func (h *TodoHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /todos", h.list)
	mux.HandleFunc("POST /todos", h.create)
	mux.HandleFunc("GET /todos/{id}", h.get)
	mux.HandleFunc("PATCH /todos/{id}", h.update)
	mux.HandleFunc("DELETE /todos/{id}", h.delete)
}

// ---------- 共通ヘルパー ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // ボディを書く前に呼ぶ
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

// decodeJSON は本文サイズを1MBに制限し、未知のフィールドを拒否する。
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func parseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

// writeStoreError は store のエラーを HTTP に変換する。詳細はログにだけ出す。
func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "タスクが見つかりません")
		return
	}
	log.Printf("internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーが発生しました")
}

// ---------- ハンドラー ----------

func (h *TodoHandler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *TodoHandler) list(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	switch status {
	case "":
		status = "all"
	case "all", "active", "done":
	default:
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "status は all / active / done のいずれかです")
		return
	}
	todos, err := h.store.List(r.Context(), status)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todos)
}

func (h *TodoHandler) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの形式が正しくありません")
		return
	}
	title, err := model.NormalizeTitle(req.Title)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	todo, err := h.store.Create(r.Context(), title)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/todos/%d", todo.ID))
	writeJSON(w, http.StatusCreated, todo)
}

func (h *TodoHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "IDが正しくありません")
		return
	}
	todo, err := h.store.Get(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "IDが正しくありません")
		return
	}
	// ポインタ型: 「送られなかった」と「false / 空文字が送られた」を区別できる
	var req struct {
		Title *string `json:"title"`
		Done  *bool   `json:"done"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの形式が正しくありません")
		return
	}
	if req.Title == nil && req.Done == nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "title または done を指定してください")
		return
	}
	in := store.UpdateInput{Done: req.Done}
	if req.Title != nil {
		title, err := model.NormalizeTitle(*req.Title)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		in.Title = &title
	}
	todo, err := h.store.Update(r.Context(), id, in)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "IDが正しくありません")
		return
	}
	if err := h.store.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```
**コミットの分け方の例**: Step 4（共通ヘルパー + list + create）→ Step 5（get）→ Step 6（update）→ Step 7（delete）→ Step 8（status・MaxBytesReader・main配線）。ファイルは同じでも、`git add -p` で機能ごとに分けられる。
</details>

<details>
<summary>Step 8 の答え: main.go の配線</summary>

`backend/cmd/server/main.go`
```go
package main

import (
	"log"
	"net/http"
	"os"

	"todoapp/backend/internal/handler"
	"todoapp/backend/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	st := store.NewMemory() // Section 6 で SQLite に差し替える
	h := handler.NewTodoHandler(st)

	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("server started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
```
`curl` での確認は [api-spec.md](api-spec.md) の「curl での動作確認例」を使う。
</details>
