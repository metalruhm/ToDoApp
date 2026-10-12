# データモデル設計書

> ステータス: **自分レビュー済**
> 関連: [requirements.md](requirements.md) / [architecture.md](architecture.md) / [api-spec.md](api-spec.md)

## 1. エンティティ

現時点では **Todo** のみ。認証導入時（Section 10）に User を追加する。

## 2. todos テーブル

| カラム | 型（SQLite） | NULL | 初期値 | 説明 |
|--------|--------------|------|--------|------|
| id | INTEGER PRIMARY KEY AUTOINCREMENT | NO | 自動 | タスクID |
| title | TEXT | NO | - | タイトル（1〜100文字） |
| done | INTEGER（0/1） | NO | 0 | 完了フラグ。SQLiteに真偽型はないため0/1で持つ |
| created_at | TEXT（ISO 8601 UTC） | NO | 現在時刻 | 作成日時 |
| updated_at | TEXT（ISO 8601 UTC） | NO | 現在時刻 | 更新日時 |

```sql
-- 方針（マイグレーションは Section 6 で自分で書く）
CREATE TABLE todos (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  title      TEXT    NOT NULL CHECK (length(title) BETWEEN 1 AND 100),
  done       INTEGER NOT NULL DEFAULT 0 CHECK (done IN (0, 1)),
  created_at TEXT    NOT NULL,
  updated_at TEXT    NOT NULL
);
```

- 一覧は `created_at` の降順（新しい順）。同時刻対策で `id` 降順を第2キーにする
- 削除は**物理削除**（論理削除はしない）。理由: 要件にないため

## 3. 制約・ルール

| ルール | 実施場所 |
|--------|----------|
| タイトルは前後の空白を除去し、1〜100文字 | Go（model）で検証。DBのCHECKは最後の砦 |
| 文字数は「文字」単位（日本語もバイトでなく文字数） | Goでは `utf8.RuneCountInString` を使う |
| `updated_at` は更新のたびに更新 | store層 |

## 4. Goの型（方針）

```go
type Todo struct {
    ID        int64     `json:"id"`
    Title     string    `json:"title"`
    Done      bool      `json:"done"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

JSONではキャメルケース（`createdAt`）、DBではスネークケース（`created_at`）に揃える。

## 5. TypeScriptの型（方針）

```ts
type Todo = {
  id: number;
  title: string;
  done: boolean;
  createdAt: string; // ISO 8601
  updatedAt: string;
};
```

## 6. 将来の変更

| 変更 | 影響 |
|------|------|
| User追加（Section 10） | `users` テーブル追加、`todos.user_id`（外部キー）追加 |
| 説明文・期限（未採用） | `todos` にカラム追加。マイグレーションで対応 |
