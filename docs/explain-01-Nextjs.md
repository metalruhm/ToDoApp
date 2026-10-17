# 解説1: Next.js とは何か（このアプリの画面側）

> 対象: Next.js / React が初めての人。このアプリのコード（[reference/frontend](../reference/frontend/)）を例に説明する。
> 関連: [explain-02-Go.md](explain-02-Go.md) / [explain-03-フロントとバックの接続.md](explain-03-フロントとバックの接続.md) / [07-Nextjs入門.md](07-Nextjs入門.md)

## 1. 先に全体像

```
ブラウザ（画面）                      サーバー
┌────────────────────┐   HTTP    ┌──────────────┐
│ Next.js (React) で  │ ───────>  │ Go の API     │
│ 作った画面          │ <───────  │               │
└────────────────────┘   JSON    └──────────────┘
   見た目と操作を担当              データと決まりごとを担当
```

**Next.js は「画面を作るための道具」**。ボタンを押す、一覧を表示する、入力欄にエラーを出す、といった、ユーザーが見て触る部分を作る。データの保存や認証の判断は Go 側が行う。

## 2. React と Next.js の関係

| 名前 | 役割 |
|---|---|
| **React** | 画面を「部品（コンポーネント）」の組み合わせで作るライブラリ |
| **Next.js** | React を使ったアプリを作りやすくする**フレームワーク**。ページの振り分け（ルーティング）、ビルド、最適化、開発サーバーなどを用意してくれる |
| **TypeScript** | JavaScript に「型」を足した言語。間違いを書いた時点で教えてくれる |
| **Tailwind CSS** | `px-4 rounded-lg` のような小さなクラス名を組み合わせてデザインする CSS の道具 |

たとえ: React は「レゴのブロック」、Next.js は「ブロックを置く土台と説明書と、完成品を運ぶ仕組み」。

## 3. コンポーネント: 画面は部品でできている

コンポーネントは、**画面の一部を返す関数**。

```tsx
type Props = { title: string };

export function Greeting({ title }: Props) {
  return <p>こんにちは、{title}</p>;
}
```
- `return` の中の HTML のような書き方が **JSX（TSX）**。`{ }` の中に変数や式を書ける
- 使うときは HTML のタグのように書く: `<Greeting title="太郎" />`
- 親から子へ渡す値を **props** と呼ぶ

このアプリの部品の関係:

```
Home（page.tsx）
├── TodoForm        入力欄と「追加」ボタン
├── StatusFilter    「全て / 未完了 / 完了」
└── TodoList        一覧（0件なら「タスクがありません」）
    └── TodoItem    1件分（チェック・編集・削除）
```
部品に分けると、「1件の見た目を変えたい」→ `TodoItem.tsx` だけ見ればよい、となる。

## 4. state: 画面が覚えている値

ボタンを押したら表示が変わる、を実現するのが **state**（`useState`）。

```tsx
const [title, setTitle] = useState("");
//     ↑現在の値  ↑更新する関数        ↑最初の値
```
- `setTitle("牛乳")` を呼ぶと、**値が変わり、画面が自動で再描画される**
- 普通の変数 `let title = ...` を書き換えても、画面は更新されない

**大事なルール: state を直接書き換えない。新しい値を渡す。**

```tsx
// ダメ: 元の配列を書き換える（React が変更に気づけない）
todos.push(newTodo);

// 良い: 新しい配列を作って渡す
setTodos((prev) => [newTodo, ...prev]);
setTodos((prev) => prev.map((t) => (t.id === id ? { ...t, done: true } : t)));
setTodos((prev) => prev.filter((t) => t.id !== id));
```
`{ ...t, done: true }` は「`t` をコピーして、`done` だけ変える」という意味。

## 5. useEffect: 画面を出した後にやること

API からデータを取ってくる処理は、画面を描画した**後**に行う。それが `useEffect`。

```tsx
useEffect(() => {
  listTodos(status).then(...);   // 描画後に実行される
}, [status]);                    // status が変わったら、もう一度実行する
```
- 最後の `[ ... ]` を**依存配列**と呼ぶ。中の値が変わったときだけ再実行される
- `[]`（空）なら、最初の1回だけ
- このアプリでは、`[status, reloadKey]` にして「絞り込みを変えたとき」「変更が成功したとき」に一覧を取り直している（[page.tsx](../reference/frontend/src/app/page.tsx)）

## 6. App Router: フォルダがそのままURLになる

Next.js は `src/app/` の中のフォルダ構成で、URLが決まる。

```
src/app/
├── layout.tsx          全ページ共通の枠（<html>, <body>）
├── page.tsx            →  /          （一覧画面）
└── login/
    └── page.tsx        →  /login     （ログイン画面）
```
- `page.tsx` が**そのURLで表示される画面**
- `layout.tsx` が**共通の枠**（フォントや `lang="ja"` など）。`children` に各ページが入る
- フォルダを増やせばページが増える（`app/about/page.tsx` → `/about`）

## 7. "use client" とは: サーバーとブラウザ

Next.js のコンポーネントは、**どこで実行されるか**が2種類ある。

| 種類 | 実行される場所 | できること |
|---|---|---|
| サーバーコンポーネント（既定） | サーバー（またはビルド時） | データ取得、秘密情報の利用。`useState` などは使えない |
| クライアントコンポーネント | ブラウザ | `useState` `useEffect`、クリックなどの操作 |

ファイルの先頭に `"use client";` と書くと、そのファイルはクライアントコンポーネントになる。

このアプリは、**ボタン操作やログイン状態の判断など、ブラウザでの動きが中心**なので、`page.tsx` `login/page.tsx` `TodoForm.tsx` `TodoItem.tsx` に `"use client"` を付けている。`TodoList.tsx` や `StatusFilter.tsx` のように、state を持たない部品には不要（親がクライアント側なので、その子としてブラウザで動く）。

迷ったら: **`useState` / `useEffect` / `onClick` を使うファイルには `"use client"` が必要**、と覚える。エラー（"You're importing a component that needs useState..."）が出たらここを疑う。

## 8. ビルドと実行の流れ

| コマンド | 何をする | いつ使う |
|---|---|---|
| `npm run dev` | 開発サーバー起動。ファイルを保存すると画面が自動更新 | 開発中 |
| `npm run build` | 本番用に最適化して出力（型チェックも行う） | 公開前、CI |
| `npm run start` | `build` した結果を起動 | 本番 |
| `npm run lint` | コードの書き方をチェック（ESLint） | コミット前、CI |

`npm run build` の最後に出る一覧の `○ (Static)` は、「ビルド時に作り置きされたページ」という意味。

**ビルドは型エラーがあると失敗する。** 「開発中は動いたのに build で失敗」は、型の問題が原因のことが多い。

## 9. 環境変数と `NEXT_PUBLIC_`

画面からAPIを呼ぶときのURLなどは、環境変数で渡す。

```ts
const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
```
- 名前が **`NEXT_PUBLIC_`** で始まるものだけが、ブラウザに送られるコードへ**文字として埋め込まれる**
- つまり、**ブラウザを使う人は誰でも値を見られる。秘密情報（鍵やパスワード）を入れてはいけない**
- 埋め込みは**ビルド時**に行われる。値を変えたら、`npm run dev` の再起動や再ビルドが必要
- ローカルは `.env.local` に書く（Git には入れない）。見本として `.env.example` を置く

## 10. このアプリの画面の動き（一覧画面）

[page.tsx](../reference/frontend/src/app/page.tsx) の流れ:

```
① 画面が表示される
   └─ useEffect: me() で「ログイン中のユーザー」を取得
   └─ useEffect: listTodos(status) で一覧を取得
        ├─ 成功: state に保存 → 一覧を描画
        └─ 401（未ログイン）: /login へ移動
② 「追加」を押す
   └─ createTodo(title)  →  成功したら reloadKey を +1  →  ①の一覧取得が再実行される
③ チェック・編集・削除も同じ形（API を呼ぶ → 成功したら一覧を取り直す）
```
**API が成功してから画面を更新する**ことで、「画面では追加されたのに、サーバーには保存されていない」状態を避けている。

## 11. Next.js 16 についての注意

このリポジトリの `frontend/AGENTS.md` には、「このバージョンは以前と違う点がある。`node_modules/next/dist/docs/` の資料を読むこと」と書かれている。Next.js は更新が速く、ネット上の古い記事のコードが動かないことがある。

- 迷ったら、**同梱のドキュメント**（`node_modules/next/dist/docs/`）や公式サイトを優先する
- AI に質問するときは、「Next.js 16 の App Router」とバージョンを伝える
- 設定ファイル `next.config.ts` に、試験的な機能（`experimental` など）が入っていることがある。意味が分からない設定は、削除せずに調べてから触る

## 12. よくあるつまずき

| 症状 | 原因と対処 |
|---|---|
| `Cannot find module '@/...'` | `tsconfig.json` の `paths`（`@/*` → `./src/*`）と、ファイルの場所がずれている |
| `useState` でエラー | `"use client"` が足りない |
| `.next/...validator.ts` の型エラー | ファイルを移動した後の古いキャッシュ。`.next` フォルダを削除して再実行 |
| 環境変数が反映されない | `.env.local` を変えた後、開発サーバーを再起動する。`NEXT_PUBLIC_` が付いているか確認 |
| 画面が更新されない | state を直接書き換えている。新しい配列・オブジェクトを作って渡す |
| `key` の警告 | 一覧の各要素にユニークな `key`（`todo.id`）を付ける |
| 開発中に API が2回呼ばれる | 開発モードの仕様（React の Strict Mode）。本番では1回 |

## 13. 自己チェック

1. React と Next.js の違いは？
2. state を直接書き換えると、なぜ画面が更新されないのか？
3. `useEffect` の依存配列は何のためにあるか？
4. `"use client"` はどんなときに必要か？
5. `NEXT_PUBLIC_API_URL` に秘密情報を入れてはいけない理由は？
6. `npm run build` と `npm run dev` の違いは？

## 14. 次に読む
- 実際に作る手順: [07-Nextjs入門.md](07-Nextjs入門.md)
- 公式: Next.js 公式サイトの Learn（日本語あり）、React 公式の「Learn」
