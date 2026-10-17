# Section 11 補足: Docker入門と詳細手順

> [11-DockerとCI.md](11-DockerとCI.md) を、Dockerが初めての人向けに細かく説明した資料。
> 順番に進めれば、Dockerの考え方 → 手で触る → ToDoアプリのコンテナ化 → CI、と段階的に理解できる。
> 未検証の注意: Dockerのバージョンや画面は変わる。動かないときはエラー文を読み、公式ドキュメントを優先する。

## 0. この資料の使い方

| パート | やること | 目安 |
|---|---|---|
| 1〜2 | Dockerの考え方を知る | 15分 |
| 3 | 小さなコマンドで触って慣れる（アプリは使わない） | 30分 |
| 4 | バックエンドをコンテナにする | 45分 |
| 5 | フロントエンドをコンテナにする | 45分 |
| 6 | まとめて起動する（Compose） | 30分 |
| 7 | CI（GitHub Actions） | 60分 |
| 8〜9 | 困ったとき・Git手順 | 必要なとき |

**いきなりアプリで試さず、パート3で小さく触る**のがコツ。エラーが出たとき、原因がDockerなのかアプリなのか切り分けやすい。

---

## 1. Dockerとは何か

### 1-1. 解決したい問題

「自分のPCでは動くのに、他のPCでは動かない」。原因は、Goのバージョン、Nodeのバージョン、設定、入っているライブラリの違いなど。

**Docker は、アプリと、その動作に必要なものをひとまとめにして、どこでも同じように動かす仕組み。**

### 1-2. 用語（まず4つだけ）

| 用語 | たとえ | 説明 |
|---|---|---|
| **イメージ** | 冷凍パックの料理 | アプリと必要なもの（OSの最小部分、ライブラリ、設定）を固めた「ひな形」。**読み取り専用で、動かない** |
| **コンテナ** | 解凍して食べている状態 | イメージから作った**実行中のもの**。同じイメージから何個でも作れる。消しても、イメージは残る |
| **Dockerfile** | レシピ | イメージの作り方を書いたテキストファイル |
| **ボリューム** | 外付けの保存場所 | コンテナを消しても残したいデータ（DBなど）の置き場所 |

```
Dockerfile --(docker build)--> イメージ --(docker run)--> コンテナ(実行中)
  レシピ                       冷凍パック                    食べている状態
```

### 1-3. 仮想マシンとの違い（参考）

| | 仮想マシン | コンテナ |
|---|---|---|
| 中身 | OSまるごと | アプリと必要最小限 |
| 起動 | 遅い（分） | 速い（秒） |
| 大きさ | GB | 数十〜数百MB |

Macで動かすときは、Docker Desktop が裏で小さなLinuxを動かし、その中でコンテナを動かしている。そのため、**コンテナの中はLinux**（`ls` `cat` などLinuxのコマンド）。

### 1-4. コンテナの重要な性質

1. **コンテナの中は独立している**: 中のファイルは、コンテナを消すと**消える**。残したいものはボリュームに置く。
2. **外からは繋がらない**: コンテナ内で `:8080` を待ち受けていても、Macから見えない。**ポートの公開（`-p`）が必要**。
3. **コンテナ内の `localhost` は、そのコンテナ自身**: Macの `localhost` ではない。

この3つを覚えておくと、つまずきの大半が理解できる。

---

## 2. 事前準備

### 2-1. Docker Desktop の起動を確認

1. `Applications` の **Docker** を起動
2. メニューバー（画面右上）のクジラのアイコンが「running」になるまで待つ（初回は1〜2分）
3. ターミナルで確認
   ```
   docker --version
   docker compose version
   ```

`Cannot connect to the Docker daemon` と出たら、**Docker Desktop が起動していない**。アプリを起動して待つ。

### 2-2. コマンドの基本形

```
docker <対象> <操作>
```
例: `docker image ls`（イメージ一覧）、`docker container ls`（コンテナ一覧）。
`docker images` `docker ps` のような短縮形もよく使われる。

---

## 3. 手で触って慣れる（アプリを使わない）

### 3-1. 最初のコンテナ

```
docker run --rm hello-world
```
何が起きたか:
1. `hello-world` イメージが手元にないので、**Docker Hub**（イメージの公開場所）からダウンロード
2. イメージからコンテナを作って実行
3. メッセージを表示して終了
4. `--rm` により、終了したコンテナを自動で削除

`docker image ls` で、`hello-world` イメージが手元に残っていることを確認する。

### 3-2. コンテナの中でコマンドを実行する

```
docker run --rm alpine echo "hello from container"
```
`alpine` は、とても小さなLinux（約5MB）。`echo ...` をコンテナ内で実行して終了する。

中に入ってみる（`-it` = 対話モード）:
```
docker run --rm -it alpine sh
```
プロンプトが変わる。次を試す。
```
cat /etc/os-release     # Alpine Linux と表示される
ls /                    # Linuxのフォルダ構成
echo hello > /tmp/a.txt
cat /tmp/a.txt
exit                    # 抜ける
```
もう一度 `docker run --rm -it alpine sh` して `cat /tmp/a.txt` すると、**ファイルは無い**。コンテナを消すと中のデータも消えるため。

### 3-3. バックグラウンドで動かし、ポートを公開する

Webサーバー（nginx）を動かす。
```
docker run -d --name web -p 8081:80 nginx
```
| オプション | 意味 |
|---|---|
| `-d` | バックグラウンドで実行（detach） |
| `--name web` | コンテナに名前を付ける |
| `-p 8081:80` | **Macの8081番を、コンテナの80番につなぐ**（`-p 外:中`） |

確認:
```
curl -i http://localhost:8081      # nginxの歓迎ページ（HTML）が返る
docker ps                          # 実行中のコンテナ一覧
docker logs web                    # コンテナのログ
docker logs -f web                 # ログを流し続ける（Ctrl+Cで抜ける。コンテナは止まらない）
docker exec -it web sh             # 動いているコンテナの中に入る（exitで抜ける）
```
後片付け:
```
docker stop web        # 停止
docker rm web          # 削除
```
`-p 8081:80` を付けずに起動すると、`curl localhost:8081` は繋がらない（性質2）。

### 3-4. ボリューム（データを残す）

```
# 1. ボリュームを作る
docker volume create demo-data

# 2. コンテナ1: ボリュームを /data に付けて、ファイルを書く
docker run --rm -v demo-data:/data alpine sh -c "echo saved > /data/note.txt"

# 3. コンテナ2（別のコンテナ）: 同じボリュームを付けて読む
docker run --rm -v demo-data:/data alpine cat /data/note.txt
```
`saved` と表示される。**コンテナは消えているのに、ボリュームにデータが残っている。**
`-v ボリューム名:コンテナ内のパス` と覚える。

後片付け:
```
docker volume rm demo-data
```

### 3-5. よく使う確認・掃除コマンド

| 目的 | コマンド |
|---|---|
| 実行中のコンテナ | `docker ps` |
| 停止中も含む | `docker ps -a` |
| イメージ一覧 | `docker image ls` |
| ボリューム一覧 | `docker volume ls` |
| イメージを削除 | `docker image rm <名前>` |
| 使っていないものを一括削除 | `docker system prune`（**確認メッセージを読む**。`-a` や `--volumes` は消えるものが増えるので慎重に） |

---

## 4. バックエンド（Go）をコンテナにする

### 4-1. 先に理解する: マルチステージビルド

Goのアプリをビルドするには、Goのコンパイラが必要（約数百MB）。しかし、**動かすだけなら、できあがった実行ファイル1つで足りる**。

```
ステージ1「build」: Goの入った大きなイメージで、実行ファイルを作る
ステージ2「実行」 : 小さなイメージに、実行ファイルだけをコピーする
```
最終的なイメージには、ステージ2だけが残る。小さく、安全になる。

### 4-2. Dockerfile を1行ずつ読む

`backend/Dockerfile`
```dockerfile
FROM golang:1.27-alpine AS build
```
**ベースにするイメージ**を指定。`golang:1.27-alpine` はGo 1.27入りの小さなLinux。`AS build` はこのステージの名前。
→ `backend/go.mod` の `go` 行以上のバージョンにする。

```dockerfile
WORKDIR /src
```
以降のコマンドを実行する**作業ディレクトリ**（なければ作られる）。

```dockerfile
COPY go.mod go.sum ./
RUN go mod download
```
`COPY 元 先`: Macのファイルをイメージにコピー。`RUN`: イメージ作成中にコマンドを実行。
**依存関係の定義だけ先にコピーして取得する**のがコツ。ソースを変更しても、`go.mod` が変わらなければこの2行の結果は**キャッシュ**されて再利用され、ビルドが速くなる。

```dockerfile
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server
```
ソース全体をコピーしてビルド。
- `CGO_ENABLED=0`: C言語に依存しない、単独で動く実行ファイルにする（`modernc.org/sqlite` はPure Goなので可能）
- `-ldflags="-s -w"`: デバッグ情報を削って小さくする
- `-o /out/server`: 出力先

```dockerfile
FROM alpine:3.20
```
ステージ2の開始。**ここから新しい小さなイメージ**。ステージ1のGoコンパイラは含まれない。

```dockerfile
RUN adduser -D -u 10001 app && mkdir /data && chown app /data
USER app
```
一般ユーザー `app` を作り、データ置き場 `/data` を `app` が書けるようにして、以降は `app` で実行。**rootで動かさない**のはセキュリティの基本。

```dockerfile
WORKDIR /app
COPY --from=build /out/server /app/server
```
`--from=build`: ステージ1「build」から、実行ファイルだけをコピー。

```dockerfile
ENV PORT=8080 DB_PATH=/data/todo.db
EXPOSE 8080
CMD ["/app/server"]
```
- `ENV`: コンテナ内の環境変数の既定値
- `EXPOSE 8080`: 「このコンテナは8080を使う」という**メモ**。実際の公開は `docker run -p`
- `CMD`: コンテナ起動時に実行するコマンド

### 4-3. .dockerignore

`backend/.dockerignore` は「イメージに入れないファイル」の指定（`COPY . .` で入るのを防ぐ）。
```
*.db
.env
.env.*
!.env.example
cover.out
*.test
```
`.env` や `todo.db`（開発データ）を入れないことが重要。秘密情報の漏えいを防ぐ。

### 4-4. ビルドして動かす

```
cd backend
docker build -t todo-backend .
```
- `-t todo-backend`: イメージに名前（タグ）を付ける
- 最後の `.`: ビルドの基準フォルダ（Dockerfileのある場所）

初回は数分かかる。ログに `[build 3/6] RUN go mod download` のように、ステップごとの進行が出る。もう一度実行すると `CACHED` と表示され、速くなる。

```
docker image ls todo-backend      # SIZE が小さい（数十MB程度）ことを確認
```

起動:
```
docker run --rm -p 8080:8080 -v todo-data:/data -e ALLOWED_ORIGIN=http://localhost:3000 todo-backend
```
| オプション | 意味 |
|---|---|
| `-p 8080:8080` | Macの8080をコンテナの8080につなぐ |
| `-v todo-data:/data` | ボリューム `todo-data` を `/data` に付ける（DBが残る） |
| `-e ...` | 環境変数を渡す |

**Mac側で `go run` のAPIが既に8080を使っている場合**は、先に止めるか `-p 18080:8080` にする（`:18080` でアクセス）。

別のターミナルで確認:
```
curl -i http://localhost:8080/health
curl -s -X POST http://localhost:8080/auth/register -H "Content-Type: application/json" \
  -d '{"email":"a@example.com","password":"password123"}'
```
`Ctrl + C` で止める。もう一度同じ `docker run` で起動して、`/auth/login` が成功すれば、**ボリュームにDBが残っている**。

### 4-5. コミット
```
git add backend/Dockerfile backend/.dockerignore
git commit -m "feat: add backend Dockerfile"
```

---

## 5. フロントエンド（Next.js）をコンテナにする

### 5-1. 先に理解する: ビルド時と実行時

Next.jsの `NEXT_PUBLIC_API_URL` は、**ビルドする瞬間に、ブラウザ用のコードへ文字として埋め込まれる**。後から環境変数を変えても反映されない。

→ イメージを作るとき（`docker build`）に値を渡す必要がある。Dockerfileでは `ARG`（ビルド時だけの変数）を使う。

**値は「ブラウザから見えるURL」にする。** コンテナ名（`backend`）はブラウザからは解決できない。

### 5-2. standalone出力

`next.config.ts` に `output: "standalone"` を足すと、実行に必要な最小限のファイルが `.next/standalone` に出力され、イメージが小さくなる。

```ts
const nextConfig: NextConfig = {
  output: "standalone",
  // ...既存の設定を残す
};
```
追加後、`npm run build` が通ることを先にMac上で確認する。

### 5-3. Dockerfile を読む

```dockerfile
FROM node:24-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
```
ステージ「deps」: 依存だけ入れる。`npm ci` は `package-lock.json` どおりに厳密にインストール（キャッシュが効く）。

```dockerfile
FROM node:24-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
RUN npm run build
```
ステージ「build」: ソースをコピーしてビルド。`ARG` で受け取った値を `ENV` にして、`npm run build` に渡す。

```dockerfile
FROM node:24-alpine AS run
WORKDIR /app
ENV NODE_ENV=production PORT=3000 HOSTNAME=0.0.0.0
RUN adduser -D -u 10001 app
COPY --from=build --chown=app /app/.next/standalone ./
COPY --from=build --chown=app /app/.next/static ./.next/static
COPY --from=build --chown=app /app/public ./public
USER app
EXPOSE 3000
CMD ["node", "server.js"]
```
ステージ「run」: ビルド結果だけをコピーして実行。`HOSTNAME=0.0.0.0` は「コンテナの外からの接続も受け付ける」設定（これがないと外から繋がらないことがある）。

注意: `public/` フォルダがないと `COPY ... public` が失敗する。このリポジトリには `public/` があるので問題ない。

`frontend/.dockerignore`
```
node_modules
.next
.env
.env.*
!.env.example
```
`node_modules` をイメージに入れない（Macのものはコンテナ内のLinuxでは使えない場合があり、巨大でもある）。

### 5-4. ビルドして動かす

```
cd frontend
docker build -t todo-frontend --build-arg NEXT_PUBLIC_API_URL=http://localhost:8080 .
docker run --rm -p 3000:3000 todo-frontend
```
ブラウザで `http://localhost:3000`。ただしAPIが別に動いていないとログインできない（次のパートでまとめる）。

### 5-5. コミット
```
git add frontend/Dockerfile frontend/.dockerignore frontend/next.config.ts
git commit -m "feat: add frontend Dockerfile"
```

---

## 6. Docker Compose でまとめて起動する

### 6-1. なぜ必要か

ここまでで、コンテナを2つ、それぞれ長い `docker run ...` で起動する必要がある。**Composeは、複数コンテナの設定を1つのファイルに書いて、1コマンドで起動する**仕組み。

### 6-2. compose.yaml を読む

プロジェクトのルート（`ToDoアプリ/`）に作る。

```yaml
services:                       # 起動するコンテナの定義
  backend:                      # サービス名（任意の名前）
    build: ./backend            # このフォルダのDockerfileでイメージを作る
    ports:
      - "8080:8080"             # Mac:コンテナ
    environment:                # 環境変数
      PORT: "8080"
      DB_PATH: /data/todo.db
      ALLOWED_ORIGIN: http://localhost:3000
      COOKIE_SECURE: "false"    # HTTPのローカルでは false
    volumes:
      - db-data:/data           # ボリューム db-data を /data に付ける
    healthcheck:                # 「正常に動いているか」の確認方法
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 10s
      timeout: 3s
      retries: 5

  frontend:
    build:
      context: ./frontend
      args:
        NEXT_PUBLIC_API_URL: http://localhost:8080   # ビルド引数（ブラウザから見えるURL）
    ports:
      - "3000:3000"
    depends_on:
      backend:
        condition: service_healthy   # backend が healthy になってから起動

volumes:                        # ボリュームの宣言
  db-data:
```

読むときのポイント:
- **YAMLはインデント（スペース2個）で構造を表す**。タブは使えない。ずれるとエラー
- `healthcheck` の `localhost` は、**そのコンテナ自身**
- `wget` はAlpineに標準で入っているので追加インストール不要
- `ports` の `"8080:8080"` は、必ず**引用符で囲む**（YAMLの解釈ミス防止）

### 6-3. 起動して確認

```
docker compose up --build
```
- `--build`: 変更を反映してイメージを作り直す
- ログが流れ続ける（`Ctrl + C` で停止）

別のターミナルで:
```
docker compose ps                    # 状態。backend が healthy になっているか
curl -i http://localhost:8080/health
```
ブラウザで `http://localhost:3000` を開く → 新規登録 → タスク追加。

バックグラウンドで起動する場合:
```
docker compose up -d --build
docker compose logs -f backend       # ログを見る
```

### 6-4. 停止と、データの残り方

| コマンド | 結果 |
|---|---|
| `docker compose stop` | 停止（コンテナは残る） |
| `docker compose down` | コンテナとネットワークを削除。**ボリュームは残る** |
| `docker compose down -v` | ボリュームも削除。**DBのデータが消える** |

**練習**
1. タスクを追加
2. `docker compose down` → `docker compose up` → タスクが残っている ✓
3. `docker compose down -v` → `docker compose up` → 登録したユーザーもタスクも消えている ✓

### 6-5. コンテナ同士の通信（理解のために）

Compose内のコンテナは、**サービス名**で互いを呼べる（例: `http://backend:8080`）。
ただし、今回のフロントはブラウザからAPIを呼ぶので、**ブラウザから見えるURL（`localhost:8080`）**を指定する。ここを間違えるのが典型的なつまずき。

```
ブラウザ(Mac) --localhost:3000--> frontend コンテナ
ブラウザ(Mac) --localhost:8080--> backend コンテナ   ← ブラウザが直接APIを呼ぶ
```

### 6-6. コミット
```
git add compose.yaml
git commit -m "feat: add docker compose"
```

---

## 7. CI（GitHub Actions）を細かく

### 7-1. CIとは

「コードをpush/PRしたら、**GitHubの用意するサーバー上で自動でテストやビルドを実行**し、結果を表示する」仕組み。自分のPCでテストを忘れても、毎回チェックされる。

### 7-2. ワークフローファイルの構造

場所は決まっている: `.github/workflows/ci.yml`（フォルダ名もそのまま、先頭のドットも必要）。

```yaml
name: CI                         # 表示名

on:                              # いつ動くか
  push:
    branches: [main]             # mainにpushされたとき
  pull_request:                  # PRが作成/更新されたとき

jobs:                            # 実行する仕事の集まり（並行して動く）
  backend:                       # ジョブ名
    runs-on: ubuntu-latest       # 実行する環境（GitHubのLinuxサーバー）
    defaults:
      run:
        working-directory: backend    # runのコマンドを実行する場所
    steps:                       # 順番に実行する手順
      - uses: actions/checkout@v4     # 他人が作った部品を使う（リポジトリを取得）
      - uses: actions/setup-go@v5     # Goをインストール
        with:                          # 部品への入力
          go-version-file: backend/go.mod
          cache-dependency-path: backend/go.sum
      - name: gofmt
        run: test -z "$(gofmt -l .)"   # 自分でコマンドを書く
      - run: go vet ./...
      - run: go test -race ./...
      - run: go build ./...
```
- `uses:` = 既製の部品を使う / `run:` = 自分でコマンドを書く
- どれか1つのステップが失敗すると、そのジョブは失敗（赤）になる
- `test -z "$(gofmt -l .)"` は「整形が必要なファイルが1つもなければ成功」という意味

フロントのジョブも同様（`setup-node` と `npm ci` / `npm run lint` / `npm run build`）。全文は [11-DockerとCI.md](11-DockerとCI.md) の付録を参照。

### 7-3. 進め方（小さく始める）

いきなり全部書かず、**1つずつ足して動かす**とエラーの原因が分かりやすい。

1. **最小のワークフロー**を作って push する
   ```yaml
   name: CI
   on: [push]
   jobs:
     hello:
       runs-on: ubuntu-latest
       steps:
         - run: echo "hello CI"
   ```
2. GitHubのリポジトリの **Actions** タブで、実行され、緑になるのを確認
3. `backend` ジョブを足す → push → 確認
4. `frontend` ジョブを足す → push → 確認

### 7-4. 結果の見方

- **Actions** タブ → 実行をクリック → ジョブをクリック → ステップを展開してログを読む
- 赤いステップを開き、エラー文を読む（`go test` の失敗なら、`--- FAIL:` の行）
- PRでは、下部の **Checks** にも表示される

### 7-5. よくあるエラー

| 症状 | 原因と対処 |
|---|---|
| ワークフローが動かない | ファイルの場所が違う（`.github/workflows/` か）。YAMLのインデント |
| `npm ci` が失敗 | `package-lock.json` がコミットされていない |
| `gofmt` ステップが失敗 | ローカルで `go fmt ./...` してコミット |
| `go.sum` が見つからない | `cache-dependency-path` のパスが違う |
| ローカルでは通るのにCIで失敗 | Goやnodeのバージョン違い、コミットし忘れたファイル、環境変数 |

### 7-6. コミット
```
git add .github/workflows/ci.yml
git commit -m "ci: add GitHub Actions workflow"
```

---

## 8. 困ったとき

### 8-1. エラーと対処

| エラー・症状 | 原因 | 対処 |
|---|---|---|
| `Cannot connect to the Docker daemon` | Docker Desktopが起動していない | アプリを起動して待つ |
| `port is already allocated` / `address already in use` | そのポートを別のものが使用中 | `lsof -i :8080` で確認し、止める。または `ports` の左側を変える（`"18080:8080"`） |
| `no such file or directory`（COPY） | `COPY` の元がビルドの基準フォルダの外、または `.dockerignore` で除外 | パスと `.dockerignore` を確認 |
| `go.mod requires go >= 1.xx` | イメージのGoが古い | `FROM golang:<バージョン>` を `go.mod` に合わせる |
| ブラウザから APIに繋がらない | `NEXT_PUBLIC_API_URL` が違う（コンテナ名など） | ビルド引数を `http://localhost:8080` にして**再ビルド** |
| 環境変数を変えたのに反映されない | `NEXT_PUBLIC_*` はビルド時に固定される | `docker compose up --build` |
| ログインできない | `ALLOWED_ORIGIN` が画面のURLと違う、`COOKIE_SECURE=true` のままHTTPで使用 | 設定を見直す |
| `permission denied`（/data） | 実行ユーザーが `/data` に書けない | Dockerfileの `chown` を確認。古いボリュームが原因なら `down -v` |
| 変更したのに反映されない | 古いイメージが使われている | `--build` を付ける、または `docker compose build --no-cache` |
| ディスクが足りない | イメージやキャッシュが溜まった | `docker system df` で確認、`docker system prune` |
| ビルドが遅い / 巨大 | `.dockerignore` がない | `node_modules` 等を除外 |

### 8-2. 調べ方の順番

1. **エラーメッセージの一番最初の赤い行**を読む（後ろは連鎖して出ているだけのことが多い）
2. `docker compose ps` で、どのコンテナが落ちているか
3. `docker compose logs <サービス名>` でそのコンテナのログ
4. `docker compose exec <サービス名> sh` で中に入って確認（フロントもAlpineなので `sh`）
5. それでも分からなければ、エラー文を丸ごとAIに貼って質問

### 8-3. コマンド早見表

| 目的 | コマンド |
|---|---|
| イメージを作る | `docker build -t 名前 .` |
| コンテナを起動 | `docker run --rm -p 外:中 -v ボリューム:パス 名前` |
| 実行中の一覧 | `docker ps` |
| ログ | `docker logs -f コンテナ名` |
| 中に入る | `docker exec -it コンテナ名 sh` |
| 止める / 消す | `docker stop 名` / `docker rm 名` |
| Compose起動（作り直し） | `docker compose up --build` |
| Compose起動（裏で） | `docker compose up -d` |
| Compose状態 | `docker compose ps` |
| Composeログ | `docker compose logs -f <サービス>` |
| Compose停止（データ残す） | `docker compose down` |
| Compose停止（データも消す） | `docker compose down -v` |

---

## 9. Section 11 全体のGit手順（まとめ）

```
git switch -c feat/11-docker

# パート4
git add backend/Dockerfile backend/.dockerignore
git commit -m "feat: add backend Dockerfile"

# パート5
git add frontend/Dockerfile frontend/.dockerignore frontend/next.config.ts
git commit -m "feat: add frontend Dockerfile"

# パート6
git add compose.yaml
git commit -m "feat: add docker compose"

# パート7
git add .github/workflows/ci.yml
git commit -m "ci: add GitHub Actions workflow"

git push -u origin feat/11-docker
# GitHubでPR作成 → Checks（CI）が緑になるのを確認 → マージ
git switch main && git pull
git tag -a v0.4.0 -m "Add Docker and CI"
git push origin v0.4.0
```

**CIを失敗→成功にする練習**は [11-DockerとCI.md](11-DockerとCI.md) の Step 7 を参照。

## 10. 自己チェック（口頭で説明できるか）

1. イメージとコンテナは、何が違うか？
2. コンテナを消すと、中のデータはどうなる？ 残すにはどうする？
3. `-p 8080:3000` の「8080」と「3000」はそれぞれどちら側か？
4. コンテナ内の `localhost` は、どこを指すか？
5. なぜ `go.mod` を先に `COPY` してから、残りを `COPY` するのか？
6. マルチステージビルドで、最終イメージが小さくなるのはなぜか？
7. `NEXT_PUBLIC_API_URL` を変更したら、なぜ再ビルドが必要か？
8. `docker compose down` と `down -v` で、何が違うか？
9. CIで `uses:` と `run:` はどう違うか？
