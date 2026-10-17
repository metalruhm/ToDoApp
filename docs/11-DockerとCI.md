# Section 11: Docker化とCI

> 📖 **詰まったら**：各Stepの答え（コード・コマンド）はこのファイル末尾の「付録: 各Stepの答え」にあります。まず自分で書いてから見ましょう。

## 1. 目的と完成の定義

**目的**: アプリをコンテナにして「どのPCでも同じように起動できる」状態にし、GitHub Actionsでテストとビルドを自動化する。

**完成の定義**:
- [ ] `docker compose up --build` で、フロント（`:3000`）とAPI（`:8080`）が起動し、ブラウザで使える
- [ ] コンテナを作り直してもデータが残る（ボリューム）
- [ ] GitHub Actionsで、push / PR のたびに `go test` と `npm run lint` / `build` が自動実行される
- [ ] PRでCIが赤くなる→直して緑になる流れを1回体験した
- [ ] `v0.4.0` のタグを打った

## 2. 用語解説

| 用語 | 意味 |
|------|------|
| コンテナ | アプリと実行環境をひとまとめにして隔離して動かす仕組み |
| イメージ | コンテナの設計図。`Dockerfile` から作る |
| マルチステージビルド | ビルド用と実行用のイメージを分け、最終イメージを小さくする書き方 |
| `docker compose` | 複数のコンテナを1つの設定ファイルでまとめて起動する |
| ボリューム | コンテナを消してもデータを残す保存場所 |
| `.dockerignore` | イメージに入れないファイルの指定（`node_modules` 等） |
| CI | 変更のたびに自動でテスト・ビルドする仕組み |
| GitHub Actions | GitHubのCI。`.github/workflows/*.yml` に書く |
| ワークフロー / ジョブ / ステップ | CIの定義 / 並行して動く単位 / 順に実行する命令 |
| ブランチ保護 | CIが通らないとマージできないようにする設定 |

## 3. 事前準備
- Docker Desktop（Mac用）をインストールして起動する。`docker --version` と `docker compose version` で確認
- 前提: Section 10までの内容が `main` にマージされている
- 注意: `NEXT_PUBLIC_*` 変数は**ビルド時**にブラウザ用コードへ埋め込まれる。実行時に環境変数を変えても反映されない

## 4. 手順（コミット単位）

### Step 1: ブランチ
```
git switch -c feat/11-docker
```

### Step 2: バックエンドのDockerfile（コミット1）
- ビルド用ステージ（Go）と実行用ステージ（軽量）に分ける
- SQLiteドライバーはPure Go（`modernc.org/sqlite`）なので `CGO_ENABLED=0` で静的にビルドできる
- 実行は一般ユーザー（root以外）で行う
- DBファイルの置き場は `/data`（ボリューム）
- `backend/.dockerignore` も作る

`feat: add backend Dockerfile`

### Step 3: フロントエンドのDockerfile（コミット2）
- `next.config.ts` に `output: "standalone"` を追加（実行に必要な最小限のファイルだけが出力される）
- 依存インストール → ビルド → 実行、の3ステージ
- `NEXT_PUBLIC_API_URL` はビルド引数（`ARG`）で受け取る
- `frontend/.dockerignore` も作る

`feat: add frontend Dockerfile`

### Step 4: docker compose（コミット3）
- `compose.yaml` をプロジェクトルートに置く
- `backend` と `frontend` の2サービス。DBのボリュームを `backend` に付ける
- 環境変数（`PORT` `DB_PATH` `ALLOWED_ORIGIN` `COOKIE_SECURE`）を設定する
- `frontend` は `backend` の起動後に始める（`depends_on`）

`feat: add docker compose`

### Step 5: 起動と確認
```
docker compose up --build
```
- `http://localhost:3000` で登録→ログイン→タスク操作ができる
- `Ctrl + C` で止め、`docker compose up` で再起動してもデータが残る
- `docker compose down -v` を実行すると**ボリュームも消えてデータが消える**ことを確認する

### Step 6: CIワークフロー（コミット4）
`.github/workflows/ci.yml` を作る。

要件:
- `push`（main）と `pull_request` で起動
- ジョブ `backend`: gofmtチェック、`go vet`、`go test -race`、`go build`
- ジョブ `frontend`: `npm ci`、`npm run lint`、`npm run build`
- 依存キャッシュを有効にする（速くなる）

`ci: add GitHub Actions workflow`（`ci:` もConventional Commitsの種類）

### Step 7: PRとCI（練習）
```
git push -u origin feat/11-docker
```
GitHubでPRを作成すると、下部の「Checks」にCIの状況が出る。

**わざと失敗させる練習**
1. テストを1つ壊して（例: 期待値を変える）コミット・push
2. PRのChecksが赤くなる→「Details」でログを読む
3. 直して push → 緑になる

### Step 8: ブランチ保護（任意）
GitHub の Settings → Branches → Branch protection rule で、`main` に「Require status checks to pass before merging」を設定する。

### Step 9: マージとタグ
```
git switch main && git pull
git tag -a v0.4.0 -m "Add Docker and CI"
git push origin main --tags
```

## 5. Git練習ポイント
- PRでのCI確認と、失敗→修正のコミット履歴
- `git tag -a` と `git push --tags`、`git tag -l`、`git show v0.4.0`
- `.gitignore` と `.dockerignore` の役割の違い（Gitに入れない / イメージに入れない）

## 6. つまずきやすい点
- `docker compose up` でフロントからAPIに繋がらない → `NEXT_PUBLIC_API_URL` がビルド時に正しく入っているか確認。ブラウザから見えるURL（`localhost:8080`）を指定する。コンテナ名（`backend`）はブラウザからは解決できない
- ログインできない → `ALLOWED_ORIGIN` が `http://localhost:3000` と完全一致しているか。`COOKIE_SECURE` はHTTPなので `false`
- DBが毎回消える → ボリュームを付けていない、または `down -v` を実行した
- Goのバージョン不一致 → `go.mod` の `go` 行以上のイメージタグを使う
- `permission denied` で `/data` に書けない → 実行ユーザーに `/data` の書き込み権限が必要
- CIで `npm ci` が失敗 → `package-lock.json` がコミットされていない
- CIはグリーンなのにローカルで失敗（またはその逆）→ Node / Go のバージョン差。CIでバージョンを固定する

## 7. 自己チェック
1. イメージとコンテナの違いは？
2. マルチステージビルドにすると何が良いか？
3. `NEXT_PUBLIC_` 変数が実行時に変更できないのはなぜか？
4. CIが自動で実行されると、どんな問題を防げるか？
5. `docker compose down` と `down -v` の違いは？

## 8. 完了したら
`learning-log.md` に追記し、**Section 12: デプロイとふりかえり** へ。

---

# 付録: 各Stepの答え

> まず自分で書いてみて、詰まったときに開く。コピペせず、**1行ずつ意味を説明できる**ようにする。
> Dockerやアクションのバージョンは変わる。動かない場合はエラーを読み、公式ドキュメントで最新を確認する。

<details>
<summary>Step 2 の答え: backend/Dockerfile と .dockerignore</summary>

`backend/Dockerfile`
```dockerfile
# ---- ビルド ----
# go.mod の go 行以上のバージョンを指定する
FROM golang:1.27-alpine AS build
WORKDIR /src

# 依存だけ先に取得（ソースが変わってもキャッシュが効く）
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- 実行 ----
FROM alpine:3.20
RUN adduser -D -u 10001 app && mkdir /data && chown app /data
USER app
WORKDIR /app
COPY --from=build /out/server /app/server

ENV PORT=8080 \
    DB_PATH=/data/todo.db
EXPOSE 8080
CMD ["/app/server"]
```

`backend/.dockerignore`
```
*.db
.env
.env.*
!.env.example
cover.out
*.test
```
- `CGO_ENABLED=0` … Cが不要な静的バイナリにする（Pure GoのSQLiteドライバーだから可能）
- `COPY go.mod go.sum` を先にする … 依存の取得結果をキャッシュするため
- `USER app` … rootで動かさない
</details>

<details>
<summary>Step 3 の答え: frontend/Dockerfile と設定</summary>

`frontend/next.config.ts`
```ts
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
};

export default nextConfig;
```

`frontend/Dockerfile`
```dockerfile
FROM node:24-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:24-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
# NEXT_PUBLIC_* はビルド時にコードへ埋め込まれる
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
RUN npm run build

FROM node:24-alpine AS run
WORKDIR /app
ENV NODE_ENV=production \
    PORT=3000 \
    HOSTNAME=0.0.0.0
RUN adduser -D -u 10001 app
COPY --from=build --chown=app /app/.next/standalone ./
COPY --from=build --chown=app /app/.next/static ./.next/static
COPY --from=build --chown=app /app/public ./public
USER app
EXPOSE 3000
CMD ["node", "server.js"]
```

`frontend/.dockerignore`
```
node_modules
.next
.env
.env.*
!.env.example
```
`public/` がないプロジェクトでは `COPY ... /app/public` の行が失敗する。`mkdir public` で空のフォルダを作るか、その行を消す。
</details>

<details>
<summary>Step 4 の答え: compose.yaml</summary>

```yaml
services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      PORT: "8080"
      DB_PATH: /data/todo.db
      ALLOWED_ORIGIN: http://localhost:3000
      COOKIE_SECURE: "false"      # HTTPのローカルでは false
    volumes:
      - db-data:/data
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 10s
      timeout: 3s
      retries: 5

  frontend:
    build:
      context: ./frontend
      args:
        # ブラウザから見えるURL（コンテナ名ではない）
        NEXT_PUBLIC_API_URL: http://localhost:8080
    ports:
      - "3000:3000"
    depends_on:
      backend:
        condition: service_healthy

volumes:
  db-data:
```
- ボリューム `db-data` がDBを保持する。`docker compose down` では残り、`down -v` で消える
- `healthcheck` の `wget` は alpine に標準で入っている
- `build.args` の値が、フロントの Dockerfile の `ARG NEXT_PUBLIC_API_URL` に渡る
</details>

<details>
<summary>Step 5 の答え: 動作確認のコマンド</summary>

```
docker compose up --build        # ビルドして起動（フォアグラウンド）
docker compose ps                # 状態（healthy になっているか）
docker compose logs -f backend   # APIのログ
curl -i http://localhost:8080/health

# 止める
docker compose down              # コンテナ削除、データは残る
docker compose up                # 再起動 → データが残っていることを確認

docker compose down -v           # ボリュームも削除（データも消える）
```
</details>

<details>
<summary>Step 6 の答え: .github/workflows/ci.yml</summary>

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  backend:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: backend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: backend/go.mod
          cache-dependency-path: backend/go.sum
      - name: gofmt
        run: test -z "$(gofmt -l .)"
      - run: go vet ./...
      - run: go test -race ./...
      - run: go build ./...

  frontend:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: frontend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 24
          cache: npm
          cache-dependency-path: frontend/package-lock.json
      - run: npm ci
      - run: npm run lint
      - run: npm run build
        env:
          NEXT_PUBLIC_API_URL: http://localhost:8080
```
- `on.push.branches: [main]` と `pull_request` … mainへのpushとPRで実行（PRのブランチをpushするたびに二重実行されない）
- `go-version-file` … `go.mod` と同じバージョンをCIでも使う
- `gofmt -l` は整形が必要なファイルを列挙する。出力があれば失敗させる
- アクションのバージョン（`@v4` など）は更新される。最新はGitHubのActionsページで確認する
</details>

<details>
<summary>Step 7 の答え: わざと失敗させる手順</summary>

```
# 例: テストの期待値を壊す
# backend/internal/model/todo_test.go の {"101文字(日本語)", ..., true} を false に変更
git add backend/internal/model/todo_test.go
git commit -m "test: break test on purpose"
git push
```
PR画面の「Checks」→「Details」→ 失敗したステップを開き、`--- FAIL: TestNormalizeTitle/101文字(日本語)` などのログを読む。

直す
```
git revert HEAD                   # 直前のコミットを打ち消すコミットを作る
# または修正して
git commit -m "test: restore expected value"
git push
```
→ Checks が緑になる。履歴に失敗と修正が残ることも確認する（`git log --oneline`）。
</details>

<details>
<summary>Step 8 の答え: ブランチ保護の設定手順</summary>

1. GitHub のリポジトリ → **Settings** → **Branches** → **Add branch protection rule**
2. Branch name pattern: `main`
3. **Require a pull request before merging** にチェック
4. **Require status checks to pass before merging** にチェックし、`backend` と `frontend` を検索して追加（一度CIが動いた後でないと候補に出ない）
5. 保存

以降、CIが赤い間はマージボタンが押せなくなる。個人リポジトリでは、プランや設定によって利用できる機能が違う場合がある。
</details>

<details>
<summary>Step 9 の答え: タグのコマンド</summary>

```
git switch main
git pull
git tag -a v0.4.0 -m "Add Docker and CI"
git tag -l                       # 一覧
git show v0.4.0                  # 内容
git push origin v0.4.0           # 1つだけ送る（または git push origin main --tags）
```
タグを間違えたとき
```
git tag -d v0.4.0                  # ローカルで削除
git push origin :refs/tags/v0.4.0  # リモートで削除（push済みの場合のみ）
```
</details>
