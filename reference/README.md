# 完成版コード（Section 1〜10 相当）

学習用の**参照実装**。自分の `backend/` `frontend/` とは別に置いてあり、自分のコードには一切触れていない。
詰まったとき・答え合わせのときに読むこと。いきなり写すと学習にならないので、まず自分で書く。

## 内容

| 領域 | 内容 | 対応する章 |
|---|---|---|
| backend | Go（`net/http`）の REST API。model / store / handler の層構造 | 3, 4 |
| backend | SQLite 保存、マイグレーション（実行済み管理つき） | 6 |
| backend | テスト（model・store・handler・auth） | 5 |
| backend | CORS（Cookie 用の `Allow-Credentials` つき） | 8 |
| backend | ユーザー登録・ログイン・ログアウト、bcrypt、サーバー側セッション、ユーザー別ToDo | 10 |
| frontend | Next.js（App Router）+ TypeScript + Tailwind CSS | 7 |
| frontend | API 連携、読み込み・エラー表示 | 8 |
| frontend | フィルタ、入力フィードバック、アクセシビリティ | 9 |
| frontend | ログイン / 新規登録画面、未ログイン時のリダイレクト | 10 |

Docker・CI・デプロイ（Section 11, 12）は含めていない。

## 起動

ターミナルを2つ使う。

```
# 1. API（http://localhost:8080）
cd reference/backend
go run ./cmd/server

# 2. 画面（http://localhost:3000）
cd reference/frontend
npm install
npm run dev
```

ブラウザで `http://localhost:3000` を開き、新規登録 → タスク操作。

環境変数（任意）:

| 変数 | 場所 | 既定値 | 説明 |
|---|---|---|---|
| `PORT` | backend | `8080` | 待ち受けポート |
| `DB_PATH` | backend | `./todo.db` | SQLite ファイル |
| `ALLOWED_ORIGIN` | backend | `http://localhost:3000` | CORS で許可する画面のURL |
| `COOKIE_SECURE` | backend | `false` | HTTPS 配信のときだけ `true` |
| `NEXT_PUBLIC_API_URL` | frontend | `http://localhost:8080` | ブラウザから見える API の URL |

## テストとチェック

```
cd reference/backend  && go vet ./... && go test -race ./...
cd reference/frontend && npm run lint && npm run build
```

## 構成

```
backend/
├── cmd/server/main.go          起動・配線・安全な終了
├── internal/
│   ├── model/                  Todo 型、タイトル検証
│   ├── store/                  TodoStore インターフェース、SQLite 実装、ユーザー/セッション
│   ├── auth/                   パスワードハッシュ、トークン
│   └── handler/                HTTP ハンドラー、認証ミドルウェア、CORS
└── migrations/                 001 todos、002 users / sessions と todos.user_id

frontend/src/
├── app/                        page.tsx（一覧）、login/page.tsx、layout.tsx
├── components/                 TodoForm / TodoList / TodoItem / StatusFilter
├── lib/api.ts                  fetch をここに集約
└── types/todo.ts
```

## 設計メモ（docs との対応）

- API は [api-spec.md](../docs/api-spec.md) どおり。更新は `PATCH`、存在しないIDの `DELETE` は `404`。
- 未知のフィールドを含む JSON は `400 INVALID_JSON`（test-plan の H-23 は「拒否する」側で実装）。
- 他ユーザーのタスクは `403` ではなく `404`（存在を知らせない）。ログイン失敗は常に同じ文言の `401`。
- セッションは「ランダムなトークンを HttpOnly Cookie に、DB には SHA-256 ハッシュだけ保存」。有効期間は7日。
- Section 4 のメモリ実装は、Section 10 で `userID` が必要になったため含めていない（全テストは SQLite で実行）。

## 既知の制限（学習用として割り切った点）

- ログイン試行の回数制限・パスワード再設定・メール確認はない。**公開環境に置かない**こと。
- マイグレーションは簡易実装（ファイル名順、実行済みを `schema_migrations` に記録）。ロールバックはない。
- 002 は `todos` を作り直す（開発用データは消える）。
- 画面の自動テストはない（lint とビルドのみ）。
