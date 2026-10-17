# 完成版コード（Section 1〜12 相当）

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

| Docker | バックエンド / フロントの Dockerfile、`compose.yaml`（マルチステージ、一般ユーザー実行、ボリューム） | 11 |
| CI | GitHub Actions（gofmt・vet・test -race・lint・build） | 11 |
| 公開 | `/api` 中継（Next.js の rewrites）、Fly.io 用 `fly.toml`、Vercel の設定メモ | 12 |

> このフォルダは、**完成後のリポジトリのルート**と同じ並び（`backend/` `frontend/` `compose.yaml` `.github/`）にしてある。
> GitHub Actions は**リポジトリ直下の** `.github/workflows/` しか読まない。自分のリポジトリで使うときは、`reference/.github/` をリポジトリ直下へコピーする。

## 起動

### A. Docker（おすすめ。Docker Desktop が必要）

```
cd reference
docker compose up --build        # http://localhost:3000 を開く
docker compose down              # 停止（データは残る）
docker compose down -v           # 停止（データも消える）
```
ポートが使用中のとき: `BACKEND_PORT=18080 FRONTEND_PORT=13000 docker compose up --build`

### B. 個別に起動

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
| `NEXT_PUBLIC_API_URL` | frontend | `http://localhost:8080` | ブラウザから見える API の URL。本番は `/api`（ビルド時に固定される） |
| `API_ORIGIN` | frontend | `http://localhost:8080` | `/api/...` の転送先（rewrites。ビルド時に固定される） |

## テストとチェック

```
cd reference/backend  && go vet ./... && go test -race ./...
cd reference/frontend && npm run lint && npm run build
cd reference && docker compose config -q     # compose.yaml の文法チェック
```

## 公開（Section 12）

画面（Vercel など）と API（Fly.io など）は別サイト扱いになり、そのままでは `SameSite=Lax` の Cookie が送られない。
そこで、画面側の `/api/...` を API へ中継する（`frontend/next.config.ts` の `rewrites`）。

```
ブラウザ ──/api/*──> 画面のサーバー(Next.js) ──中継──> API(Go) ──> SQLite（ボリューム）
          同一オリジンなので Cookie も CORS も問題にならない
```

| 手順 | 内容 |
|---|---|
| 1. API | `cd backend && fly launch --no-deploy`（`fly.toml` の app 名を直す）→ `fly volumes create data --region nrt --size 1` → `fly scale count 1`（SQLite は1台）→ `fly deploy` |
| 2. 画面 | Vercel でリポジトリを連携し、**Root Directory を `frontend`** にする |
| 3. 環境変数（Vercel） | `NEXT_PUBLIC_API_URL=/api`、`API_ORIGIN=https://<アプリ名>.fly.dev` |
| 4. 確認 | 登録 → ログイン → 追加 → 再読み込み。Cookie に `HttpOnly` と `Secure` が付いていること |

環境変数を変えたら、Vercel は**再デプロイ**が必要（ビルド時に固定されるため）。
公開前に、ログイン試行の回数制限がない点（下の「既知の制限」）を理解しておく。

## 構成

```
compose.yaml                    api と画面をまとめて起動
.github/workflows/ci.yml        CI
backend/
├── Dockerfile / .dockerignore
├── fly.toml                    Fly.io 用
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
（ほか frontend/Dockerfile、frontend/next.config.ts の rewrites）
```

## 設計メモ（docs との対応）

- API は [api-spec.md](../docs/api-spec.md) どおり。更新は `PATCH`、存在しないIDの `DELETE` は `404`。
- 未知のフィールドを含む JSON は `400 INVALID_JSON`（test-plan の H-23 は「拒否する」側で実装）。
- 他ユーザーのタスクは `403` ではなく `404`（存在を知らせない）。ログイン失敗は常に同じ文言の `401`。
- セッションは「ランダムなトークンを HttpOnly Cookie に、DB には SHA-256 ハッシュだけ保存」。有効期間は7日。
- Docker のイメージは、API が約 30MB、画面が約 290MB。どちらも一般ユーザーで実行する。
- Section 4 のメモリ実装は、Section 10 で `userID` が必要になったため含めていない（全テストは SQLite で実行）。

## 既知の制限（学習用として割り切った点）

- ログイン試行の回数制限・パスワード再設定・メール確認はない。**公開環境に置かない**こと。
- マイグレーションは簡易実装（ファイル名順、実行済みを `schema_migrations` に記録）。ロールバックはない。
- 002 は `todos` を作り直す（開発用データは消える）。
- 画面の自動テストはない（lint とビルドのみ）。
- Fly.io / Vercel / GitHub Actions は、この環境では実際に動かしていない（設定は公式の仕様に沿って書いただけ）。Docker での起動、`/api` 中継、データの永続化は実際に確認済み。
