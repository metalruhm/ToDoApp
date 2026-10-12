# API仕様書

> ステータス: **自分レビュー済**
> 関連: [requirements.md](requirements.md) / [data-model.md](data-model.md)

## 1. 共通事項

| 項目 | 内容 |
|------|------|
| ベースURL | `http://localhost:8080` |
| 形式 | リクエスト/レスポンスともにJSON（`Content-Type: application/json`） |
| 文字コード | UTF-8 |
| 日時 | ISO 8601（UTC）例: `2026-10-12T06:30:00Z` |
| 認証 | なし（Section 10 で追加） |

### Todoオブジェクト

```json
{
  "id": 1,
  "title": "牛乳を買う",
  "done": false,
  "createdAt": "2026-10-12T06:30:00Z",
  "updatedAt": "2026-10-12T06:30:00Z"
}
```

### エラーレスポンス

すべてのエラーは次の形式で返す。

```json
{ "error": { "code": "VALIDATION_ERROR", "message": "タイトルは1〜100文字で入力してください" } }
```

| HTTP | code | 意味 |
|------|------|------|
| 400 | `VALIDATION_ERROR` | 入力値が不正 |
| 400 | `INVALID_JSON` | JSONとして解析できない / IDが数値でない |
| 404 | `NOT_FOUND` | 指定IDのタスクがない |
| 500 | `INTERNAL_ERROR` | サーバー内部エラー（詳細は返さず、ログに残す） |

## 2. エンドポイント一覧

| メソッド | パス | 概要 | 要件 |
|----------|------|------|------|
| GET | `/health` | 稼働確認 | - |
| GET | `/todos` | 一覧取得 | F-02, F-06 |
| POST | `/todos` | 追加 | F-01, F-07 |
| GET | `/todos/{id}` | 1件取得 | - |
| PATCH | `/todos/{id}` | 部分更新（タイトル編集・完了切替） | F-03, F-04 |
| DELETE | `/todos/{id}` | 削除 | F-05 |

> 補足: ロードマップでは更新を `PUT` と書いたが、「完了だけ切り替える」「タイトルだけ変える」を1つの口で扱えるため `PATCH` を採用。PUTとPATCHの違いは学習ポイント。

## 3. 各エンドポイント詳細

### GET /health

- 200: `{"status":"ok"}`

### GET /todos

クエリ（任意）: `status=all|active|done`（省略時 `all`）

- 200: Todoの配列（`created_at` の新しい順）。0件のときは `[]`（`null` にしない）
- 400: `status` が上記以外 → `VALIDATION_ERROR`

```json
[
  { "id": 2, "title": "レポート提出", "done": true,  "createdAt": "...", "updatedAt": "..." },
  { "id": 1, "title": "牛乳を買う",   "done": false, "createdAt": "...", "updatedAt": "..." }
]
```

### POST /todos

リクエスト:
```json
{ "title": "牛乳を買う" }
```
| フィールド | 型 | 必須 | 検証 |
|-----------|----|------|------|
| title | string | ○ | 前後の空白除去後に1〜100文字 |

- 201: 作成したTodo（`done` は `false`）。`Location: /todos/{id}` ヘッダーを付ける
- 400: 検証エラー / JSON不正

### GET /todos/{id}

- 200: Todo
- 400: IDが数値でない
- 404: 存在しない

### PATCH /todos/{id}

リクエスト（**送ったフィールドだけ**更新。1つ以上必須）:
```json
{ "title": "牛乳とパンを買う", "done": true }
```
| フィールド | 型 | 必須 | 検証 |
|-----------|----|------|------|
| title | string | 任意 | `POST` と同じ |
| done | boolean | 任意 | true / false |

- 200: 更新後のTodo（`updatedAt` が更新される）
- 400: 空のボディ `{}`、検証エラー、JSON不正
- 404: 存在しない

### DELETE /todos/{id}

- 204: 本文なし
- 404: 存在しない（※ 冪等性の観点で204にする案もある → レビューで判断）

## 4. CORS

フロント（`http://localhost:3000`）からの呼び出しを許可する。

- `Access-Control-Allow-Origin`: 環境変数 `ALLOWED_ORIGIN` の値
- `Access-Control-Allow-Methods`: `GET, POST, PATCH, DELETE, OPTIONS`
- `Access-Control-Allow-Headers`: `Content-Type`
- `OPTIONS`（プリフライト）には 204 を返す

## 5. curl での動作確認例

```bash
curl -i http://localhost:8080/health
curl -i -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d '{"title":"牛乳を買う"}'
curl -i http://localhost:8080/todos
curl -i -X PATCH http://localhost:8080/todos/1 -H "Content-Type: application/json" -d '{"done":true}'
curl -i -X DELETE http://localhost:8080/todos/1
curl -i -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d '{"title":"  "}'   # 400になること
```

## 6. テスト観点（Section 5 で使用）

- 正常系: 追加 → 一覧に出る / 完了切替 / 編集 / 削除
- 異常系: 空タイトル、空白のみ、101文字、存在しないID、IDが文字列、壊れたJSON、`PATCH` に `{}`
- 境界値: タイトル1文字、100文字（日本語100文字を含む）
