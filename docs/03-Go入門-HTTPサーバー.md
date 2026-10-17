# Section 3: Go入門 — 最小のHTTPサーバー

> 📖 **詰まったら**：各Stepの答え（コード・コマンド）はこのファイル末尾の「付録: 各Stepの答え」にあります。まず自分で書いてから見ましょう。

## 1. 目的と完成の定義

**目的**: Goの基本文法と標準ライブラリ `net/http` を使い、JSONを返すHTTPサーバーを自分で書く。

**完成の定義**:
- [ ] `backend/` に Goモジュールがあり、`go run ./cmd/server` で起動する
- [ ] `curl http://localhost:8080/health` が `{"status":"ok"}` を返す
- [ ] `go fmt` `go vet` がエラーなしで通る
- [ ] `PORT` 環境変数でポートを変更できる
- [ ] 機能ごとに小さくコミットした（3コミット以上）

## 2. 用語解説

| 用語 | 意味 |
|------|------|
| モジュール | Goのプロジェクト単位。`go.mod` で管理 |
| パッケージ | ディレクトリ単位のコード集まり。`package main` は実行可能 |
| 構造体 (struct) | 複数の値をまとめた型 |
| 構造体タグ | `` `json:"id"` `` のようにJSON変換の名前を指定する注釈 |
| ハンドラー | HTTPリクエストを処理する関数 |
| `go.sum` | 依存の改ざん検知用ハッシュ。コミットする |

## 3. 事前準備：work/ の整理

現在 `work/` にある `go.mod` `go.sum` `main.go` を設計どおりの場所へ移す。

```
mkdir -p backend/cmd/server
git mv work/go.mod work/go.sum backend/    # 追跡前なら mv でよい
mv work/main.go backend/cmd/server/main.go
```
`go.mod` の `module` 名を確認する（例: `module todo-app/backend`）。

## 4. 手順（コミット単位）

### Step 1: ブランチ
```
git switch -c feat/03-http-server
```

### Step 2: Hello World（コミット1）
`cmd/server/main.go` に `package main` と `func main()` を書き、`fmt.Println` で文字を出す。
```
cd backend
go run ./cmd/server
```
```
git add backend && git commit -m "chore: set up Go module and entry point"
```

### Step 3: HTTPサーバーを起動（コミット2）
ヒント:
- `http.NewServeMux()` でルーターを作る
- `mux.HandleFunc("GET /health", handler)`（Go 1.22以上のメソッド付きパターン）
- ハンドラーの形: `func(w http.ResponseWriter, r *http.Request)`
- `http.ListenAndServe(":8080", mux)` の戻り値（error）を無視しない

まず文字列 `ok` を返し、`curl -i localhost:8080/health` で確認。
```
git commit -m "feat: add health endpoint"
```

### Step 4: JSONを返す（コミット3）
ヒント:
- `w.Header().Set("Content-Type", "application/json")`
- `json.NewEncoder(w).Encode(...)`
- まず `map[string]string{"status":"ok"}`、次に自分で定義した構造体で書き直す

```
git commit -m "feat: return JSON from health endpoint"
```

### Step 5: ポートを環境変数化（コミット4）
ヒント: `os.Getenv("PORT")`、空なら `8080` を使う。
```
PORT=9000 go run ./cmd/server
git commit -m "feat: make port configurable via PORT"
```

### Step 6: 品質チェック
```
go fmt ./...
go vet ./...
go build ./...
```
エディタの保存時に自動で `gofmt` が走る設定にしておく（VS Codeの Go拡張）。

### Step 7: 練習（Git）
- コミットメッセージの誤字 → `git commit --amend`
- 編集を破棄したい → `git restore <file>`
- ミスで増えたコミット数を確認 → `git log --oneline`

### Step 8: マージ
```
git switch main && git merge feat/03-http-server && git branch -d feat/03-http-server
```

## 5. 学ぶGo文法（このセクションで触れる範囲）
`package` / `import` / `func` / 変数 (`:=`) / `if err != nil` / 構造体 / 構造体タグ / `defer` / ポインタの基本

> 目安: [A Tour of Go](https://go.dev/tour/) の Basics と Methods まで。**全部覚えようとせず**、必要になったら戻る。

## 6. つまずきやすい点
- `go: cannot find main module` → `backend/`（`go.mod` のある場所）で実行する
- ポート使用中エラー → 前のプロセスを `Ctrl+C` で止める、または `PORT` を変える
- 使っていない変数/importはコンパイルエラーになる（Goの仕様）
- `err` を捨てない。Goでは `if err != nil` が基本パターン
- JSONに出ない → 構造体フィールドが小文字始まり（非公開）になっている

## 7. 自己チェック
1. `package main` と通常のパッケージの違いは？
2. なぜ `ListenAndServe` の戻り値を確認するのか？
3. 構造体タグ `json:"createdAt"` は何をしているか？
4. `go.sum` はコミットすべきか？

## 8. 完了したら
`learning-log.md` に追記し、**Section 4: ToDo API** へ。

---

# 付録: 各Stepの答え

> まず自分で書いてみて、詰まったときに開く。コピペせず、**1行ずつ意味を説明できる**ようにする。
> 前提: `backend/go.mod` の1行目は `module todoapp/backend` のみ（`module` 行は1つだけ）。

<details>
<summary>事前準備の答え（go.mod の整理と配置）</summary>

```
# go.mod の「module todoapp」の行を消し、module todoapp/backend だけにする
git mv backend/server backend/cmd/server      # すでに cmd/ がある場合
cd backend
go mod tidy                                    # 使っていない依存(gin等)が整理される
```
`" server"`（先頭スペース）のフォルダがあれば、`mv "backend/ server" backend/cmd/server` で直す。
</details>

<details>
<summary>Step 2 の答え（Hello World）</summary>

`backend/cmd/server/main.go`
```go
package main

import "fmt"

func main() {
	fmt.Println("hello")
}
```
実行: `cd backend && go run ./cmd/server`
</details>

<details>
<summary>Step 3 の答え（HTTPサーバー）</summary>

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	log.Println("server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```
- `http.NewServeMux()` … URLとハンドラーを結びつけるルーター
- `"GET /health"` … メソッドとパスの組（Go 1.22以上）
- `fmt.Fprint(w, ...)` … `w`（レスポンス）に文字を書く
- `log.Fatal(...)` … `ListenAndServe` はエラー時だけ戻ってくる。エラーを表示して終了する

確認: `curl -i localhost:8080/health`
</details>

<details>
<summary>Step 4 の答え（JSONを返す）</summary>

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type healthResponse struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	log.Println("server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```
- ヘッダーは**ボディを書く前**に設定する
- 構造体のフィールドは大文字始まり（公開）でないとJSONに出ない
- 確認: `{"status":"ok"}` が返る
</details>

<details>
<summary>Step 5 の答え（PORT環境変数）と最終形</summary>

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type healthResponse struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	log.Printf("server started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
```
確認: `PORT=9000 go run ./cmd/server` → `curl localhost:9000/health`
</details>

<details>
<summary>Step 6 の答え（品質チェック）</summary>

```
cd backend
go fmt ./...     # 整形（変更があればファイルが書き換わる）
go vet ./...     # 怪しい書き方の検出
go build ./...   # コンパイルできるか
```
</details>
