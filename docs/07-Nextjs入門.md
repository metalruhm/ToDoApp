# Section 7: Next.js入門

> 📖 **詰まったら**：各Stepの答え（コード・コマンド）はこのファイル末尾の「付録: 各Stepの答え」にあります。まず自分で書いてから見ましょう。

## 1. 目的と完成の定義

**目的**: Next.js（App Router）・React・TypeScriptの基本を学び、バックエンドなしでToDo画面を動かす。

**完成の定義**:
- [ ] `frontend/` で `npm run dev` が起動し、`http://localhost:3000` が表示される
- [ ] ダミーデータでタスク一覧が表示される
- [ ] 追加・完了切替・編集・削除が（メモリ上で）動く
- [ ] コンポーネントが分割されている（最低3つ）
- [ ] `npm run lint` と `npm run build` が通る

## 2. 用語解説

| 用語 | 意味 |
|------|------|
| React | UIをコンポーネント（部品）で作るライブラリ |
| Next.js | Reactのフレームワーク。ルーティング・ビルド等を提供 |
| App Router | `app/` ディレクトリでページを作る方式 |
| コンポーネント | 画面の部品。関数として書く |
| props | 親から子へ渡す値 |
| state / `useState` | 画面が覚えている値。変わると再描画される |
| `"use client"` | ブラウザで動くコンポーネントの宣言 |
| TypeScript | 型付きJavaScript |
| JSX/TSX | JavaScript内にHTML風に書く記法 |

## 3. 手順（コミット単位）

### Step 1: ブランチ
```
git switch -c feat/07-nextjs-basics
```

### Step 2: プロジェクト作成（コミット1：chore）
プロジェクトルートで:
```
npx create-next-app@latest frontend
```
質問の推奨回答: TypeScript = Yes / ESLint = Yes / Tailwind = Yes / `src/` = Yes / App Router = Yes / import alias は既定。

生成物が大量なので**このコミットは生成物だけ**にする。
```
git status        # node_modules が出ていないか確認
git add frontend
git commit -m "chore: scaffold Next.js app"
```
※ `frontend/.git` が作られていたら削除（入れ子のリポジトリになるため）。`ls -a frontend` で確認。

### Step 3: 動かしてみる（コミットなし）
```
cd frontend && npm run dev
```
`src/app/page.tsx` を編集して、画面が変わるのを確認する。

### Step 4: Todo型とダミーデータ（コミット2）
`src/types/todo.ts` に [data-model.md](data-model.md) のTypeScript型を書く。
`src/app/page.tsx` に、ダミーの配列を表示する。
`feat: show dummy todo list`

### Step 5: コンポーネント分割（コミット3）
`src/components/` に分ける。
- `TodoForm`（入力欄＋追加ボタン）
- `TodoList`
- `TodoItem`（1件）

ヒント: 一覧は `map` で描画し、`key` に `id` を指定。

### Step 6: stateで動かす（コミット4〜5）
- `page.tsx`（または親コンポーネント）に `"use client"` と `useState<Todo[]>` を置く
- 追加: 空/空白のみは拒否してエラー表示（F-07）、100文字上限
- 完了切替・削除
- 編集: 編集モードの切替（`editingId` のstate）
- state更新は**元の配列を書き換えず**、新しい配列を作る（`map` `filter` `[...]`）

仮のID発行は `Date.now()` などでよい（後でAPIのIDに置き換わる）。

### Step 7: 品質チェック（コミット6）
```
npm run lint
npm run build
```
エラーは直してコミット。

### Step 8: マージ
```
git switch main && git merge feat/07-nextjs-basics
```

## 4. Git練習ポイント
- **生成物のコミットと自作コードのコミットを分ける**
- `git stash`: 作業中に別の作業を挟むとき
  ```
  git stash push -m "wip: edit mode"
  git switch main
  git switch -
  git stash pop
  ```
- `git status` で `node_modules` `.next` が出ていないこと（`.gitignore` の効果確認）

## 5. つまずきやすい点
- `useState` の関数がサーバーコンポーネントでエラー → `"use client"` を書く
- stateを直接書き換える（`todos.push(...)`）と画面が更新されない
- `key` 警告 → リスト要素にユニークな `key` を付ける
- Nodeのバージョンが古い → `node -v` を確認
- ポート3000使用中 → `npm run dev -- -p 3001`

## 6. 自己チェック
1. stateとpropsの違いは？
2. なぜstateを直接書き換えてはいけないのか？
3. `"use client"` はいつ必要か？
4. `key` は何のためにあるか？

## 7. 完了したら
`learning-log.md` に追記し、**Section 8: フロントとバックの接続** へ。

---

# 付録: 各Stepの答え

> まず自分で書いてみて、詰まったときに開く。コピペせず、**1行ずつ意味を説明できる**ようにする。
> 前提: `create-next-app` で TypeScript / Tailwind / `src/` / App Router を選択済み。パスのエイリアス `@/` は `src/` を指す。

<details>
<summary>Step 2 の答え: プロジェクト作成の回答とコミット</summary>

```
npx create-next-app@latest frontend
```
推奨の回答: TypeScript = Yes / ESLint = Yes / Tailwind CSS = Yes / `src/` directory = Yes / App Router = Yes / import alias = 既定(`@/*`)

```
ls -a frontend          # frontend/.git があれば rm -rf frontend/.git（入れ子リポジトリを防ぐ）
git status              # node_modules / .next が出ていないこと
git add frontend
git commit -m "chore: scaffold Next.js app"
```
</details>

<details>
<summary>Step 4 の答え: Todo型とダミーデータ表示</summary>

`frontend/src/types/todo.ts`
```ts
export type Todo = {
  id: number;
  title: string;
  done: boolean;
  createdAt: string; // ISO 8601
  updatedAt: string;
};
```

`frontend/src/app/page.tsx`
```tsx
import type { Todo } from "@/types/todo";

const dummy: Todo[] = [
  { id: 1, title: "牛乳を買う", done: false, createdAt: "2026-10-12T00:00:00Z", updatedAt: "2026-10-12T00:00:00Z" },
  { id: 2, title: "レポート提出", done: true, createdAt: "2026-10-12T00:00:00Z", updatedAt: "2026-10-12T00:00:00Z" },
];

export default function Home() {
  return (
    <main>
      <h1>ToDo</h1>
      <ul>
        {dummy.map((todo) => (
          <li key={todo.id}>{todo.title}</li>
        ))}
      </ul>
    </main>
  );
}
```
`key` は、リストの各要素を見分けるための印（ユニークな値）。
</details>

<details>
<summary>Step 5 の答え: コンポーネント分割</summary>

`src/components/TodoForm.tsx`
```tsx
"use client";

import { useState } from "react";

type Props = {
  onAdd: (title: string) => void;
};

export function TodoForm({ onAdd }: Props) {
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault(); // ページ再読み込みを止める
    const trimmed = title.trim();
    if (trimmed.length < 1 || trimmed.length > 100) {
      setError("タイトルは1〜100文字で入力してください");
      return;
    }
    setError("");
    onAdd(trimmed);
    setTitle("");
  }

  return (
    <form onSubmit={handleSubmit}>
      <label htmlFor="title">新しいタスク</label>
      <input
        id="title"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="新しいタスクを入力"
      />
      <button type="submit">追加</button>
      {error && <p role="alert">{error}</p>}
    </form>
  );
}
```

`src/components/TodoItem.tsx`
```tsx
"use client";

import { useState } from "react";
import type { Todo } from "@/types/todo";

type Props = {
  todo: Todo;
  onToggle: (id: number) => void;
  onDelete: (id: number) => void;
  onEdit: (id: number, title: string) => void;
};

export function TodoItem({ todo, onToggle, onDelete, onEdit }: Props) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(todo.title);

  function save() {
    const trimmed = draft.trim();
    if (trimmed.length < 1 || trimmed.length > 100) return;
    onEdit(todo.id, trimmed);
    setEditing(false);
  }

  return (
    <li>
      <input
        type="checkbox"
        checked={todo.done}
        onChange={() => onToggle(todo.id)}
        aria-label={`${todo.title} を完了にする`}
      />
      {editing ? (
        <>
          <input value={draft} onChange={(e) => setDraft(e.target.value)} />
          <button onClick={save}>保存</button>
          <button onClick={() => { setDraft(todo.title); setEditing(false); }}>キャンセル</button>
        </>
      ) : (
        <>
          <span style={{ textDecoration: todo.done ? "line-through" : "none" }}>{todo.title}</span>
          <button onClick={() => setEditing(true)}>編集</button>
          <button onClick={() => onDelete(todo.id)}>削除</button>
        </>
      )}
    </li>
  );
}
```

`src/components/TodoList.tsx`
```tsx
import type { Todo } from "@/types/todo";
import { TodoItem } from "./TodoItem";

type Props = {
  todos: Todo[];
  onToggle: (id: number) => void;
  onDelete: (id: number) => void;
  onEdit: (id: number, title: string) => void;
};

export function TodoList({ todos, ...handlers }: Props) {
  if (todos.length === 0) return <p>タスクがありません</p>;
  return (
    <ul>
      {todos.map((todo) => (
        <TodoItem key={todo.id} todo={todo} {...handlers} />
      ))}
    </ul>
  );
}
```
</details>

<details>
<summary>Step 6 の答え: stateで動かす（src/app/page.tsx）</summary>

```tsx
"use client";

import { useState } from "react";
import type { Todo } from "@/types/todo";
import { TodoForm } from "@/components/TodoForm";
import { TodoList } from "@/components/TodoList";

export default function Home() {
  const [todos, setTodos] = useState<Todo[]>([]);

  function addTodo(title: string) {
    const now = new Date().toISOString();
    const todo: Todo = { id: Date.now(), title, done: false, createdAt: now, updatedAt: now };
    setTodos((prev) => [todo, ...prev]); // 元の配列を書き換えず、新しい配列を作る
  }

  function toggleTodo(id: number) {
    setTodos((prev) => prev.map((t) => (t.id === id ? { ...t, done: !t.done } : t)));
  }

  function deleteTodo(id: number) {
    setTodos((prev) => prev.filter((t) => t.id !== id));
  }

  function editTodo(id: number, title: string) {
    setTodos((prev) => prev.map((t) => (t.id === id ? { ...t, title } : t)));
  }

  return (
    <main>
      <h1>ToDo</h1>
      <TodoForm onAdd={addTodo} />
      <TodoList todos={todos} onToggle={toggleTodo} onDelete={deleteTodo} onEdit={editTodo} />
    </main>
  );
}
```
- `setTodos((prev) => ...)` … 直前の値を元に新しい値を作る書き方
- `{ ...t, done: !t.done }` … 元のオブジェクトをコピーして `done` だけ変える
- `todos.push(...)` のような直接変更は、Reactが変更に気づけず画面が更新されない
</details>

<details>
<summary>Step 7 の答え: lint と build</summary>

```
cd frontend
npm run lint
npm run build
```
よくある指摘と直し方:
- 未使用のimport/変数 → 削除する
- `<img>` の警告 → 今回は画像を使わないので出ないはず
- 型エラー → 型を明示する、`any` を使わない
</details>

<details>
<summary>Git練習の答え: stash</summary>

```
git stash push -m "wip: edit mode"   # 作業中の変更を退避
git switch main                      # 別のブランチで作業
git switch -                         # 元のブランチに戻る
git stash list                       # 退避の一覧
git stash pop                        # 退避を戻す（一覧から消える）
```
</details>
