# 解説2: Go とは何か（このアプリのサーバー側）

> 対象: Go が初めての人。このアプリのコード（[reference/backend](../reference/backend/)）を例に説明する。
> 関連: [explain-01-Nextjs.md](explain-01-Nextjs.md) / [explain-03-フロントとバックの接続.md](explain-03-フロントとバックの接続.md) / [03](03-Go入門-HTTPサーバー.md)・[04](04-GoでToDo-API.md)

## 1. Go の役割

```
ブラウザ ──HTTP──> Go のサーバー ──SQL──> SQLite（todo.db）
                   │
                   ├ リクエストを受け取る
                   ├ ログイン中の人を確認する
                   ├ 入力が正しいか検証する
                   ├ データを保存・取得する
                   └ JSON で結果を返す
```
**Go は「データと決まりごとを守る側」**。画面（Next.js）を信用せず、**大事なチェック（入力検証、ログイン確認、他人のデータを見せない）は必ず Go 側で行う**。画面側のチェックは、親切のためのもので、守りのためではない。

## 2. Go の特徴

| 特徴 | 意味 |
|---|---|
| **コンパイル言語** | 実行前に機械語へ変換する。できあがるのは1つの実行ファイル（Docker でも小さく済む） |
| **静的型付け** | 変数に型がある。間違いをビルド時に教えてくれる |
| **シンプルな文法** | 機能が少なく、読みやすさ重視。書き方がほぼ1通りに揃う（`go fmt` で自動整形） |
| **標準ライブラリが強い** | HTTP サーバー、JSON、テストが標準で付いている。このアプリは外部フレームワークを使っていない |
| **並行処理が得意** | ゴルーチンで、たくさんのリクエストを同時に処理できる |

## 3. プロジェクトの構造

```
backend/
├── go.mod                     モジュール名と依存ライブラリの一覧
├── go.sum                     依存の改ざん検知用ハッシュ（コミットする）
├── cmd/server/main.go         起動の入口（package main）
├── internal/                  このプロジェクト専用のコード（外部から import 不可）
│   ├── model/                 データの型と検証
│   ├── store/                 保存と取得（DB）
│   ├── auth/                  パスワード、トークン
│   └── handler/               HTTP の入口と出口
└── migrations/                DB の構造を変える SQL
```

| 用語 | 意味 |
|---|---|
| **モジュール** | プロジェクト全体の単位。`go.mod` の `module todoapp/backend` |
| **パッケージ** | フォルダ単位のコードのまとまり。ファイル先頭の `package handler` |
| `package main` | 実行できるプログラム。`func main()` が起動点 |
| `internal/` | この名前のフォルダの中身は、同じモジュールからしか import できない |

import は `import "todoapp/backend/internal/model"` のように書き、使うときは `model.Todo` と、**パッケージ名で修飾**する。

## 4. 文法の基本（このアプリに出てくる範囲）

### 変数と関数
```go
name := "牛乳"                 // := で宣言と代入（型は自動で決まる）
var count int = 3              // 型を明示する書き方もある

func add(a, b int) int {       // 引数の型、戻り値の型
	return a + b
}
```

### 複数の戻り値とエラー処理（最重要）
Go には例外（try/catch）がない。**関数が結果とエラーを両方返し、呼ぶ側が確認する。**

```go
todo, err := h.store.Create(ctx, userID, title)
if err != nil {
	// エラーのときの処理
	return
}
// ここに来たら成功
```
`if err != nil` は Go で最もよく見る形。**エラーを無視しない**（`_` で捨てるのは理由があるときだけ）。

### 構造体（struct）と JSON
```go
type Todo struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}
```
- 複数の値をまとめた型
- バッククォートの `json:"id"` は**構造体タグ**。JSON にしたときのキー名を指定する
- **フィールド名が大文字で始まらないと、JSON に出ない**（公開/非公開のルール）

### スライスとマップ
```go
todos := []model.Todo{}        // スライス（可変長の配列）
todos = append(todos, t)       // 追加
for _, t := range todos { }    // 繰り返し

m := map[int64]model.Todo{}    // マップ（キーと値）
```
空のスライスは `[]` のまま JSON になるが、**`nil` スライスは `null` になる**。API が `[]` を返したいので `[]model.Todo{}` で初期化している。

### ポインタ（`*`）
```go
var req struct {
	Title *string `json:"title"`   // string ではなく *string
	Done  *bool   `json:"done"`
}
if req.Done == nil { /* done は送られなかった */ }
```
`*bool` は「bool の置き場所を指す値」で、**「値がない（nil）」を表せる**。PATCH で「`done: false` が送られた」と「`done` が送られなかった」を区別するために使っている。`bool` だけだと、どちらも `false` になって区別できない。

### インターフェース
```go
type TodoStore interface {
	List(ctx context.Context, userID int64, status string) ([]model.Todo, error)
	Get(...)
	// ...
}
```
「**このメソッドを持つ型なら何でも**」という約束。handler は `TodoStore` だけを知っていて、中身が SQLite かメモリかは知らない。

→ **テスト時やDB変更時に、実装を差し替えられる**。Go のインターフェースは、実装側が「実装します」と宣言しなくても、メソッドが揃っていれば自動的に満たす。

### defer
```go
rows, err := s.db.QueryContext(ctx, q)
if err != nil { return nil, err }
defer rows.Close()      // この関数を抜けるときに必ず実行される
```
閉じ忘れを防ぐ。

## 5. HTTP サーバーの仕組み（`net/http`）

### 部品は3つだけ
```
ルーター(ServeMux)  ──>  ハンドラー  ──>  レスポンス
「どのURLか」           「何をするか」      「何を返すか」
```

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /todos", h.list)       // メソッドとパスを関数に結びつける
mux.HandleFunc("PATCH /todos/{id}", h.update)
http.ListenAndServe(":8080", mux)          // ポート8080で待ち受け
```
- `{id}` は `r.PathValue("id")` で取り出せる

### ハンドラーの形
```go
func (h *TodoHandler) list(w http.ResponseWriter, r *http.Request) {
	// r: 届いたリクエスト（URL, ヘッダー, 本文）
	// w: レスポンスを書き込む先
}
```

### レスポンスを書く順番（重要）
1. ヘッダーを設定（`w.Header().Set(...)`）
2. ステータスコード（`w.WriteHeader(201)`）
3. 本文を書く

**本文を書き始めると、ヘッダーとステータスは変更できなくなる。** 順番を間違えると、意図したステータスにならない。

### ミドルウェア
「ハンドラーを受け取って、ハンドラーを返す関数」。**複数の入口に共通する処理を前後に挟む**ために使う。

```go
// CORS: 全リクエストに許可ヘッダーを付ける
handler.CORS(allowedOrigin, mux)

// RequireAuth: ログインしていなければ 401 を返し、していれば本処理を呼ぶ
mux.HandleFunc("GET /todos", authH.RequireAuth(h.list))
```

```
リクエスト → [CORS] → [ルーター] → [RequireAuth] → [list ハンドラー] → レスポンス
```

## 6. このアプリの層構造

```
handler  （HTTP の言葉: JSON、ステータスコード、Cookie）
   ↓ 呼ぶ
store    （保存の言葉: SQL、トランザクション）
   ↓ 使う
model    （アプリの言葉: Todo、タイトルは1〜100文字）
```

| 層 | 知っていること | 知らないこと |
|---|---|---|
| handler | HTTP | SQL |
| store | SQL / DB | HTTP |
| model | アプリのルール | HTTP も DB も |

**層を分ける理由**: 1つの変更が1か所で済む。たとえば「DB を SQLite から PostgreSQL に変える」なら store だけ、「エラーの JSON の形を変える」なら handler だけ。テストも層ごとに書ける。

## 7. リクエストが処理される流れ（タスク追加）

`POST /todos` のとき、[handler/todo.go](../reference/backend/internal/handler/todo.go) の `create`:

```
1. RequireAuth       Cookie から「誰か」を特定 → context にユーザーIDを入れる
2. decodeJSON        本文を {title} に変換（壊れていたら 400 INVALID_JSON）
3. NormalizeTitle    前後の空白を除き、1〜100文字か検証（NG なら 400 VALIDATION_ERROR）
4. store.Create      INSERT（user_id も保存）
5. writeJSON(201)    作成したタスクを JSON で返す。Location ヘッダー付き
```
どの段階で失敗しても、**エラーの形は同じ**（`{"error":{"code":"...","message":"..."}}`）。画面側はこの形を前提にメッセージを表示できる。

## 8. 並行処理と context

### ゴルーチン
HTTP サーバーは、**リクエストごとにゴルーチン（軽量スレッド）を作って並行処理する**。つまり、**複数のリクエストが同時に同じデータにアクセスする**ことがある。

- メモリ上の `map` を複数のゴルーチンが同時に書くとクラッシュする → `sync.Mutex`（鍵）で守る（Section 4 のメモリ実装）
- SQLite は書き込みが1つずつなので、`SetMaxOpenConns(1)` で接続を1本にして、競合を避けている
- 競合の検出には `go test -race`

### context
`context.Context` は、**リクエストのキャンセルや、リクエスト単位の値**を運ぶもの。

- `r.Context()` を DB 操作に渡すと、**ブラウザが接続を切ったら DB 操作も中断できる**
- ログイン中のユーザーID を context に入れ、`userID(r)` で取り出している

## 9. 認証まわり（Go の側の仕事）

| 項目 | このアプリの実装 |
|---|---|
| パスワード | **bcrypt** でハッシュ化して保存。元に戻せない |
| ログイン後 | ランダムなトークンを作り、**Cookie には元のトークン、DB にはそのハッシュ**を保存 |
| Cookie | `HttpOnly`（JS から読めない）、`SameSite=Lax`、本番は `Secure` |
| 認可 | 全ての SQL に `WHERE user_id = ?`。他人のIDは **404**（存在を知らせない） |
| ログイン失敗 | 「メールがない」も「パスワード違い」も**同じメッセージ**（アカウントの存在を漏らさない） |

## 10. テスト

Go は標準でテストを書ける。

```go
func TestNormalizeTitle(t *testing.T) {
	tests := []struct{ name, input string; wantErr bool }{
		{"通常", "牛乳", false},
		{"空", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeTitle(tt.input)
			if (err != nil) != tt.wantErr { t.Errorf(...) }
		})
	}
}
```
- ファイル名は `_test.go`、関数名は `TestXxx`
- **テーブル駆動テスト**: 入力と期待値を表にして、ループで回す。ケースの追加が簡単
- HTTP は `httptest` で、サーバーを起動せずに試せる
- 実行: `go test ./...`、競合検出つきは `go test -race ./...`

## 11. よく使うコマンド

| コマンド | 意味 |
|---|---|
| `go run ./cmd/server` | ビルドして実行 |
| `go build ./...` | コンパイルできるか確認 |
| `go test ./...` | テスト実行 |
| `go fmt ./...` | 書式を自動整形 |
| `go vet ./...` | 怪しい書き方の検出 |
| `go mod tidy` | 使っていない依存を整理、足りないものを追加 |
| `go get <パッケージ>` | 依存を追加 |

## 12. よくあるつまずき

| 症状 | 原因と対処 |
|---|---|
| `declared and not used` | 使っていない変数・import はエラー（Go の仕様）。消すか `_` を使う |
| JSON にフィールドが出ない | フィールド名が小文字始まり（非公開）。大文字にする |
| 空の一覧が `null` になる | `nil` スライス。`[]T{}` で初期化 |
| `PATCH` で `false` が無視される | `bool` ではなく `*bool` を使う |
| `concurrent map writes` | 複数のゴルーチンが `map` を書いている。Mutex で守る |
| `database is locked` | `rows.Close()` を忘れた、または接続が複数 |
| ステータスコードが意図と違う | `WriteHeader` の前に本文を書いた |
| `cannot find main module` | `go.mod` のあるフォルダで実行していない |

## 13. 自己チェック

1. なぜ入力検証を、画面側だけでなく Go 側でも行うのか？
2. `if err != nil` は何のためにあるか？
3. `*bool` を使う理由は？
4. インターフェースを使うと、何が嬉しいか？
5. ミドルウェアとは何か。このアプリのどこで使っているか？
6. HTTP サーバーで、複数のリクエストが同時に来るとどうなるか？
7. 他人のタスクIDにアクセスされたとき、403 でなく 404 を返す理由は？

## 14. 次に読む
- 実際に作る手順: [03](03-Go入門-HTTPサーバー.md)・[04](04-GoでToDo-API.md)・[05](05-Goのテスト.md)・[06](06-データベース.md)
- 公式: A Tour of Go（日本語あり）、Effective Go
