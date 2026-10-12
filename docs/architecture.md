# アーキテクチャ設計書

> ステータス: **自分レビュー済**
> 関連: [requirements.md](requirements.md) / [data-model.md](data-model.md) / [api-spec.md](api-spec.md)

## 1. 全体構成

```
┌──────────────┐   HTTP(JSON)   ┌──────────────┐   SQL    ┌──────────┐
│  ブラウザ     │ ─────────────> │  Go API      │ ───────> │ SQLite   │
│ (Next.js UI) │ <───────────── │ :8080        │ <─────── │ (todo.db)│
│ :3000        │                └──────────────┘          └──────────┘
└──────────────┘
```

- フロント（Next.js, `:3000`）とバックエンド（Go, `:8080`）は**別プロセス**。ポートが違うため、Go側で **CORS** を許可する必要がある（Section 8）
- 通信はJSON over HTTP。仕様は [api-spec.md](api-spec.md)
- DBはSQLite（ファイル1つ）。Section 4まではメモリ上で保持する

## 2. ディレクトリ構成

```
ToDoアプリ/
├── docs/                 # ドキュメント
├── backend/              # Go
│   ├── cmd/server/main.go        # 起動処理（エントリーポイント）
│   ├── internal/
│   │   ├── handler/              # HTTP層: リクエスト解析・レスポンス生成
│   │   ├── model/                # Todo構造体、バリデーション
│   │   └── store/                # 保存層: インターフェース + 実装(memory / sqlite)
│   ├── migrations/               # SQLファイル
│   └── go.mod
├── frontend/             # Next.js
│   └── src/
│       ├── app/                  # ページ・レイアウト
│       ├── components/           # UI部品
│       └── lib/api.ts            # APIクライアント(fetchをここに集約)
└── README.md
```

※ 現在 `work/` にある `main.go` などは、Section 3で `backend/` に整理する想定。

## 3. バックエンドの層構造

```
handler  ──>  store(interface)  ──>  memory 実装 / sqlite 実装
   │
   └──> model（Todo, 入力検証）
```

| 層 | 責務 | やらないこと |
|----|------|--------------|
| handler | JSONの読み書き、ステータスコード決定 | SQLを書く |
| model | Todoの型、タイトル検証 | HTTPを知る |
| store | 保存・取得 | HTTPを知る |

**なぜ分けるか**: `store` をインターフェースにすると、メモリ実装 → SQLite実装へ**handlerを変えずに**差し替えられ、テストではメモリ実装を使える。

```go
// 方針のみ（実装は自分で書く）
type TodoStore interface {
    List(ctx) ([]Todo, error)
    Get(ctx, id) (Todo, error)
    Create(ctx, title) (Todo, error)
    Update(ctx, id, patch) (Todo, error)
    Delete(ctx, id) error
}
```

## 4. フロントエンドの構造

| 要素 | 方針 |
|------|------|
| ルーティング | App Router。画面は `/` の1枚 |
| 状態管理 | まず `useState` のみ。ライブラリは入れない |
| API呼び出し | `lib/api.ts` に集約。コンポーネントから直接 `fetch` しない |
| 描画 | まずクライアントコンポーネント（`"use client"`）で作る。慣れたらサーバーコンポーネントを検討 |
| 設定 | APIのURLは `NEXT_PUBLIC_API_URL`（`.env.local`） |

## 5. 主要な処理フロー（タスク追加）

```
ユーザー: タイトル入力 → [追加]
  → フロント: 空チェック（簡易）
  → POST /todos {title}
      → handler: JSON解析 → model: 検証（空・100文字超は 400）
      → store.Create → DB INSERT
      ← 201 {todo}
  ← フロント: 一覧の先頭に追加、入力欄をクリア
  ※ 400 の場合: エラーメッセージを表示
```

## 6. 設計判断（ADR風メモ）

| 判断 | 採用 | 理由 / 代替案 |
|------|------|---------------|
| GoのHTTP | 標準 `net/http`（Go 1.22+ のパターン付きルーティング） | 学習のため。ginやechoは使わない |
| DB | SQLite | 環境構築が不要。後でPostgreSQLへ移行可能なように `store` で隠蔽 |
| ID | 整数の自動採番 | 単純さを優先。認証導入時に再検討 |
| フロント状態管理 | `useState` | 規模が小さい |
| モノレポ | 1リポジトリに `backend/` `frontend/` | Git学習の対象を1つにするため |

## 7. 開発・実行環境

| 用途 | コマンド（想定） |
|------|------------------|
| API起動 | `cd backend && go run ./cmd/server` |
| フロント起動 | `cd frontend && npm run dev` |
| APIテスト | `cd backend && go test ./...` |

環境変数:

| 変数 | 場所 | 例 |
|------|------|----|
| `PORT` | backend | `8080` |
| `DB_PATH` | backend | `./todo.db` |
| `ALLOWED_ORIGIN` | backend | `http://localhost:3000` |
| `NEXT_PUBLIC_API_URL` | frontend | `http://localhost:8080` |

`.env` はコミットしない。`.env.example` を用意する。

## 8. 将来の拡張ポイント

- 認証（Section 10）: `handler` の前段にミドルウェアを追加、`todos` に `user_id` を追加
- PostgreSQL化: `store` の新実装を追加するのみ
- Docker化（Section 11）: backend / frontend それぞれにDockerfile
