# 解説3: フロントエンドとバックエンドの接続

> 対象: 「画面（Next.js）とサーバー（Go）は、どうやって話しているのか」を知りたい人。
> 関連: [explain-01-Nextjs.md](explain-01-Nextjs.md) / [explain-02-Go.md](explain-02-Go.md) / [08-フロントとバックの接続.md](08-フロントとバックの接続.md) / [api-spec.md](api-spec.md)

## 1. 全体像

```
┌─────────────────────────┐            ┌──────────────────────┐
│ ブラウザ                 │            │ Go のサーバー          │
│  ┌───────────────────┐  │  HTTP      │  ┌────────────────┐  │
│  │ Next.js の画面     │──┼──────────>│  │ handler        │  │
│  │ (localhost:3000)   │<─┼──────────  │  │ (localhost:8080)│  │
│  └───────────────────┘  │  JSON      │  └──────┬─────────┘  │
└─────────────────────────┘            │         ↓ SQL        │
                                       │      SQLite          │
                                       └──────────────────────┘
```
- 画面とサーバーは**別のプログラム**（別のポート）。話すための共通言語が **HTTP** で、データの形が **JSON**
- 約束事を書いたものが [API仕様書](api-spec.md)。**画面とサーバーは、この仕様だけを信じて作る**ので、片方を作り替えてももう一方は影響を受けにくい

## 2. HTTP の基本

### 2-1. リクエストとレスポンス

**画面が頼み、サーバーが答える。** 1回のやりとりは必ず「リクエスト → レスポンス」のセット。

リクエスト（画面 → サーバー）:
```
POST /todos HTTP/1.1                       ← メソッドとパス
Host: localhost:8080
Content-Type: application/json             ← 本文は JSON ですよ
Cookie: session=abc...                      ← ログインの証明

{"title":"牛乳を買う"}                       ← 本文
```
レスポンス（サーバー → 画面）:
```
HTTP/1.1 201 Created                        ← ステータスコード
Content-Type: application/json
Location: /todos/5

{"id":5,"title":"牛乳を買う","done":false,...}
```

### 2-2. メソッド: 何をしたいか

| メソッド | 意味 | このアプリ |
|---|---|---|
| GET | 取得 | 一覧・1件の取得 |
| POST | 新規作成 | タスク追加、ログイン、登録 |
| PATCH | 一部だけ更新 | 完了切替・タイトル編集 |
| DELETE | 削除 | タスク削除 |

### 2-3. ステータスコード: 結果はどうだったか

| コード | 意味 | このアプリでの例 |
|---|---|---|
| 200 | 成功 | 一覧取得、更新 |
| 201 | 作成した | タスク追加、ユーザー登録 |
| 204 | 成功、返す本文なし | 削除、ログアウト |
| 400 | 依頼が不正 | 空のタイトル、壊れた JSON |
| 401 | ログインが必要 | Cookie がない・期限切れ |
| 404 | 見つからない | 存在しない、または他人のタスク |
| 409 | 衝突 | 登録済みのメールアドレス |
| 500 | サーバー内部エラー | 想定外の不具合 |

**4xx = 依頼した側（画面）の問題、5xx = サーバーの問題**、と覚える。

### 2-4. JSON
データを文字列で表す形式。画面もサーバーも、この形でやりとりする。

```json
{ "id": 1, "title": "牛乳", "done": false }
```
- Go: 構造体 ⇔ JSON を `encoding/json` が変換（フィールドの `json:"..."` タグで名前が決まる）
- TypeScript: `res.json()` で JavaScript のオブジェクトに変換。型は `Todo` を自分で書いて合わせる

## 3. 1つの操作を端から端まで追う（「追加」ボタン）

```
【画面 / ブラウザ】
 ① TodoForm: 入力を検証 → onAdd(title)
 ② page.tsx: run(() => createTodo(title))
 ③ lib/api.ts: fetch("http://localhost:8080/todos", { method:"POST", credentials:"include", body })
      │
      │  ※ 先に OPTIONS（プリフライト）が飛ぶ。→ 4章
      ▼
【Go / サーバー】
 ④ CORS ミドルウェア: 許可ヘッダーを付ける
 ⑤ RequireAuth: Cookie → セッションを探す → ユーザーを特定（なければ 401）
 ⑥ create ハンドラー: JSON を解析 → タイトルを検証（NG なら 400）
 ⑦ store.Create: INSERT INTO todos (user_id, title, ...)
 ⑧ writeJSON: 201 とタスクの JSON を返す
      │
      ▼
【画面 / ブラウザ】
 ⑨ api.ts: res.ok なら JSON を返す（失敗なら ApiError を投げる）
 ⑩ page.tsx: 成功 → reloadKey を +1 → useEffect が listTodos を再実行
 ⑪ 取得した一覧で state を更新 → 画面が再描画される
```

実際のコード:
- 画面側: [api.ts](../reference/frontend/src/lib/api.ts) / [page.tsx](../reference/frontend/src/app/page.tsx)
- サーバー側: [handler/todo.go](../reference/backend/internal/handler/todo.go) / [handler/auth.go](../reference/backend/internal/handler/auth.go)

## 4. CORS: なぜ必要で、何をしているのか

### 4-1. 問題
`http://localhost:3000`（画面）から `http://localhost:8080`（API）への通信は、**ブラウザにとって「別のオリジン」**。

> **オリジン = スキーム（http） + ホスト（localhost） + ポート（3000）**
> ポートが違うだけで別オリジン。

ブラウザは安全のため、**別オリジンへの通信の結果を、初期状態ではページの JavaScript に渡さない**。悪意のあるサイトが、あなたがログインしている別のサービスを勝手に操作するのを防ぐための仕組み。

### 4-2. 解決: サーバーが「許可」を伝える
API が、レスポンスのヘッダーで「`http://localhost:3000` からの利用は許可する」と伝えると、ブラウザは通信を通す。

```
Access-Control-Allow-Origin: http://localhost:3000     ← このオリジンは許可
Access-Control-Allow-Credentials: true                  ← Cookie を使う通信も許可
Access-Control-Allow-Methods: GET, POST, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type
```
実装: [handler/cors.go](../reference/backend/internal/handler/cors.go)。許可するオリジンは環境変数 `ALLOWED_ORIGIN` で設定する。

### 4-3. プリフライト（事前確認）
`POST` + `Content-Type: application/json` や、`PATCH` `DELETE` のようなリクエストは、本番の送信**前に**ブラウザが自動で確認を送る。

```
ブラウザ → サーバー : OPTIONS /todos  （「この通信、していいですか？」）
サーバー → ブラウザ : 204 + Access-Control-Allow-* ヘッダー  （「いいですよ」）
ブラウザ → サーバー : POST /todos  （本番）
```
**OPTIONS に 204 を返さない（405 など）と、本番のリクエストは送られず、`fetch` が失敗する。** 「追加だけエラーになる」はこれが典型。

### 4-4. 大事な理解
- CORS を**制限しているのはブラウザ**。`curl` ではCORSは起きない（「curlでは動くのに画面では動かない」の原因）
- 許可を**設定するのはサーバー**（Go）。画面側のコードでは解決できない
- Cookie を使うときは、`Allow-Origin` に `*` は使えない。具体的なオリジンを書く必要がある

## 5. ログイン状態の保ち方（Cookie とセッション）

HTTP は「前のリクエストを覚えていない」ので、**毎回「自分は誰か」を伝える必要がある**。その運び屋が Cookie。

```
① ログイン:  画面 → POST /auth/login {email, password}
②            サーバー: パスワードを確認 → ランダムなトークンを作る
             Set-Cookie: session=<トークン>; HttpOnly; SameSite=Lax; Path=/
③ 以降:      ブラウザが、API へのリクエストに自動で Cookie を付ける
④            サーバー: Cookie のトークン → セッション → ユーザーを特定
```

| 設定 | 意味 |
|---|---|
| `HttpOnly` | JavaScript から Cookie を読めない（盗まれにくい） |
| `SameSite=Lax` | 別サイトから送られるリクエストには付けない（なりすまし対策） |
| `Secure` | HTTPS のときだけ送る（本番） |
| 画面側 `credentials: "include"` | **別オリジンの API にも Cookie を付けて送る**指定。これを忘れるとログインしていない扱いになる |
| サーバー側 `Allow-Credentials: true` | 上を許可する |

**`credentials: "include"`（画面）と `Allow-Credentials`（サーバー）は、両方そろって初めて動く。**

画面は Cookie の中身を見られない（`HttpOnly`）ので、「ログイン中か」は `GET /auth/me` を呼んで、**401 か 200 か**で判断する。

## 6. エラーの扱い: 両側で約束をそろえる

### サーバーの約束（[API仕様](api-spec.md)）
どんなエラーも同じ形で返す。
```json
{ "error": { "code": "VALIDATION_ERROR", "message": "タイトルは1〜100文字で入力してください" } }
```

### 画面の処理（[api.ts](../reference/frontend/src/lib/api.ts)）
`fetch` の失敗には**2種類**ある。

| 種類 | 例 | 見分け方 | 画面の対応 |
|---|---|---|---|
| ネットワーク失敗 | API が止まっている、CORS で拒否 | `fetch` が例外を投げる | 「サーバーに接続できません」 |
| HTTP エラー | 400, 401, 404, 500 | `res.ok === false` | JSON の `error.message` を表示 |

- **401 を受けたら `/login` へ移動**する（未ログインまたは期限切れ）
- `res.ok` を確認せずに `res.json()` すると、エラー応答を成功と勘違いする

### 検証は両方でやる
| 場所 | 目的 |
|---|---|
| 画面 | すぐ教える（親切）。サーバーに無駄な通信をしない |
| サーバー | **本当の守り**。画面を通さず直接 API を呼ばれても守れる |

## 7. 設定: どのURLを呼ぶか

| 場所 | 変数 | 例 | 意味 |
|---|---|---|---|
| 画面 | `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | **ブラウザから見える** API の URL |
| サーバー | `ALLOWED_ORIGIN` | `http://localhost:3000` | 許可する**画面の** URL |
| サーバー | `COOKIE_SECURE` | `false`（本番 `true`） | HTTPS のときだけ Cookie を送る |

**この2つ（画面側の API URL とサーバー側の許可オリジン）は、お互いを指す。** どちらか一方だけ変えると繋がらなくなる。

## 8. 環境ごとの違い

### 8-1. ローカル（開発）
```
画面: npm run dev      → http://localhost:3000
API : go run ./cmd/server → http://localhost:8080
```
画面から API へは別オリジンの通信 → **CORS が必要**。

### 8-2. Docker Compose
```
backend  コンテナ → ホストの 8080
frontend コンテナ → ホストの 3000
```
ポイント: **API を呼ぶのはコンテナ内のサーバーではなく、ユーザーのブラウザ**。だから、`NEXT_PUBLIC_API_URL` は `http://backend:8080`（コンテナ名）ではなく、**ブラウザから見える `http://localhost:8080`** にする。

### 8-3. 本番（別ドメインに置くとき）
`xxx.vercel.app`（画面）と `yyy.fly.dev`（API）は**別サイト**になり、`SameSite=Lax` の Cookie が送られない → ログインが保てない。

対策の一つが、**画面のサーバーが `/api/...` を API に転送する**（Next.js の `rewrites`）。ブラウザからは同一オリジンに見えるので、CORS も Cookie の問題も起きにくい。詳しくは [12-デプロイとふりかえり.md](12-デプロイとふりかえり.md)。

## 9. 動作を確かめる方法

### 9-1. ブラウザの開発者ツール（`Cmd + Option + I`）
- **Network タブ**: 送ったリクエスト、ステータスコード、ヘッダー、レスポンスの中身が見える
  - `OPTIONS` が出ていて 204 か？
  - `POST /todos` が 201 か？ 赤（失敗）なら、行を選んで Headers と Response を見る
- **Console タブ**: `blocked by CORS policy` などのエラーが出る
- **Application タブ → Cookies**: `session` Cookie があるか。`HttpOnly` が付いているか

### 9-2. curl で API 単体を確認
画面を介さず、API だけを試せる。**画面でエラーのとき、curl で成功するなら原因は CORS・Cookie・URL 設定**と切り分けられる。

```
curl -i http://localhost:8080/health

# ログインして Cookie を保存し、続けて使う
curl -c jar -X POST localhost:8080/auth/login -H "Content-Type: application/json" \
  -d '{"email":"a@example.com","password":"password123"}'
curl -b jar localhost:8080/todos

# CORS のプリフライトを再現
curl -i -X OPTIONS localhost:8080/todos \
  -H "Origin: http://localhost:3000" -H "Access-Control-Request-Method: POST"
```

## 10. 切り分けの順番

画面で「うまくいかない」とき:

```
① API は動いているか?          curl localhost:8080/health
② curl で API は成功するか?    成功 → API は正常。画面側か接続設定を疑う
③ Network タブで赤い行は?       どのリクエストが何番で失敗したか
④ OPTIONS が 204 か?           違う → CORS ミドルウェアが効いていない
⑤ Console に CORS エラー?       ALLOWED_ORIGIN と画面のURLが完全一致か
⑥ 401 ばかり?                  credentials: "include"、Cookie の属性、ALLOWED_ORIGIN
⑦ 400 の中身は?                 Response の error.message を読む
```

## 11. よくある症状と原因

| 症状 | 原因 |
|---|---|
| `blocked by CORS policy` | `ALLOWED_ORIGIN` が画面の URL と違う、または CORS ミドルウェアが `main.go` に組み込まれていない |
| 「追加」だけ失敗する（一覧は出る） | プリフライト（OPTIONS）が通っていない |
| ログインはできるが、すぐ未ログインに戻る | `credentials: "include"` がない、`Allow-Credentials` がない、`Secure` の Cookie を HTTP で使っている |
| 常に「サーバーに接続できません」 | API が停止、URL/ポートが違う |
| `localhost` と `127.0.0.1` で挙動が違う | 別オリジン扱い。`ALLOWED_ORIGIN` と画面のURLのホスト名を合わせる |
| Docker で画面は出るが API に繋がらない | `NEXT_PUBLIC_API_URL` がコンテナ名になっている。ブラウザから見える URL にして**再ビルド** |
| `.env.local` を変えたのに反映されない | 開発サーバーを再起動する（`NEXT_PUBLIC_` はビルド時に埋め込まれる） |
| 一覧が `null` でエラー | サーバーが空のとき `[]` ではなく `null` を返している |

## 12. 自己チェック

1. 画面とサーバーが別のプログラムなのに、やりとりできるのはなぜか？
2. 「オリジン」とは何か。`localhost:3000` と `localhost:8080` が別オリジンなのはなぜか？
3. CORS を制限しているのはブラウザとサーバーのどちらか？ 許可を設定するのはどちらか？
4. プリフライト（OPTIONS）とは何か？ なぜ「追加だけ」失敗することがあるのか？
5. `credentials: "include"` と `Access-Control-Allow-Credentials` は、なぜ両方必要か？
6. `fetch` の失敗2種類（ネットワーク失敗と HTTP エラー）を、どう見分けるか？
7. 検証を画面とサーバーの両方でやる理由は？
8. Docker で `NEXT_PUBLIC_API_URL` にコンテナ名を書いてはいけないのはなぜか？
9. 「curl では成功するのに、画面では失敗する」とき、まず何を疑うか？

## 13. 次に読む
- 実装手順: [08-フロントとバックの接続.md](08-フロントとバックの接続.md)
- 認証の詳細: [10-認証-ユーザー別ToDo.md](10-認証-ユーザー別ToDo.md)
- 公開時の構成: [12-デプロイとふりかえり.md](12-デプロイとふりかえり.md)
