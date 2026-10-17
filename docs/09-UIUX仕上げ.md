# Section 9: UI/UX仕上げ

> 📖 **詰まったら**：各Stepの答え（コード・コマンド）はこのファイル末尾の「付録: 各Stepの答え」にあります。まず自分で書いてから見ましょう。

## 1. 目的と完成の定義

**目的**: 使いやすく見やすい画面に仕上げる。Tailwind CSS、フィルタ、入力フィードバック、アクセシビリティの基本を学ぶ。履歴の整理（rebase）を練習する。

**完成の定義**:
- [ ] Tailwind CSSでレイアウト・配色が整っている
- [ ] フィルタ（全て/未完了/完了）が動く（F-06）
- [ ] 入力エラー、0件表示（「タスクがありません」）、送信中表示がある
- [ ] キーボードだけで主要操作ができる（Tab / Enter / Space）
- [ ] 完了タスクが見た目で区別できる（F-03）
- [ ] 画面幅を狭めても崩れない
- [ ] ローカルブランチで `git rebase` による履歴整理を1回行った

## 2. 用語解説

| 用語 | 意味 |
|------|------|
| Tailwind CSS | クラス名を組み合わせてスタイルを書くCSSフレームワーク |
| ユーティリティクラス | `p-4` `text-lg` のような1機能1クラス |
| レスポンシブ | 画面幅に応じてレイアウトを変えること |
| a11y | アクセシビリティ（accessibility）の略 |
| `aria-*` | 支援技術に情報を伝える属性 |
| フォーカスリング | キーボード操作時に現在位置を示す枠 |
| rebase | ブランチの土台を付け替え、履歴を直線にする操作 |

## 3. 手順（コミット単位）

### Step 1: ブランチ
```
git switch -c feat/09-ui-polish
```

### Step 2: デザインを決める（コミットなし or docs）
簡単なメモでよい。
- 配色: 背景/文字/強調色/エラー色の4色
- 余白と文字サイズの方針
- 状態: 通常 / ホバー / フォーカス / 無効 / エラー
`docs/ui-notes.md` にまとめて、AIに初稿作成を依頼しレビューしてもよい。

### Step 3: レイアウト（コミット）
- 中央寄せの1カラム、最大幅 `max-w-xl` 程度
- リストの1行をカード風にする
- 完了タスクは `line-through` と薄い文字色
`style: improve layout with Tailwind`（`style:` は見た目の変更用に追加してよい種類）

### Step 4: フィルタ（コミット）
- APIの `?status=` を使う方式、またはフロントで絞り込む方式を選び、**選んだ理由を1行メモ**
- 選択中のボタンを強調表示、`aria-pressed` を付ける
`feat: add status filter`

### Step 5: 入力フィードバック（コミット）
- 空/100文字超: エラーメッセージを入力欄の下に表示（`role="alert"`）
- 文字数カウンタ（例: `12/100`）
- 送信中はボタン無効化と「追加中…」
- 0件のとき「タスクがありません」
`feat: improve form feedback and empty state`

### Step 6: アクセシビリティ（コミット）
- 入力欄に `<label>`（見えなくても `sr-only` で付ける）
- 完了切替は `<input type="checkbox">` を使う（`div` に `onClick` を付けない）
- ボタンにはテキストか `aria-label`
- フォーカスリングを消さない
- 文字色と背景色のコントラスト比（目安 4.5:1）をブラウザ開発者ツールで確認
`fix: improve accessibility`

### Step 7: レスポンシブ確認
ブラウザの開発者ツールでスマホ幅（375px）にして崩れを確認 → 修正。

### Step 8: rebaseで履歴整理（ローカル限定）
コミットが細かすぎる/誤字修正が多い場合に整理する。

**必ず先にバックアップブランチを作る。**
```
git branch backup/09
git rebase -i main
```
エディタでコミットの先頭を `pick` / `squash` / `reword` に変更する。
うまくいかなければ `git rebase --abort`、または `git reset --hard backup/09`。

**ルール: すでに `push` したコミットには使わない。**

### Step 9: マージ
```
git switch main && git merge feat/09-ui-polish
git tag -a v0.2.0 -m "UI polish"
```

## 4. 動作確認チェック
- [ ] Tabキーで入力欄 → 追加 → 各行の操作へ移動できる
- [ ] Enterで追加できる
- [ ] 100文字ちょうど/101文字で期待どおり
- [ ] APIを止めたときにエラーが分かる
- [ ] スマホ幅で横スクロールが出ない

## 5. つまずきやすい点
- Tailwindのクラスが効かない → `globals.css` の読込み、`content` 設定、バージョンによる設定差
- クラスが長くなる → コンポーネントに切り出す（CSSにまとめず部品化）
- 動的に作ったクラス名（`bg-${color}`）はTailwindが検出できない
- rebaseで混乱 → すぐ `--abort`。バックアップがあれば戻れる

## 6. 自己チェック
1. なぜ `div` に `onClick` を付けるのはよくないか？
2. エラーメッセージに `role="alert"` を付ける効果は？
3. `merge` と `rebase` の履歴の違いは？
4. rebaseを `push` 済みのブランチにしてはいけない理由は？

## 7. 完了したら
`learning-log.md` に追記。**ここまでで「動くアプリ」の目標達成**。続けて **Section 10: 認証（発展）** へ。

---

# 付録: 各Stepの答え

> まず自分で書いてみて、詰まったときに開く。コピペせず、**1行ずつ意味を説明できる**ようにする。
> Tailwind のクラス名はバージョンで一部異なる場合がある。効かないときは公式ドキュメントで確認する。

<details>
<summary>Step 2 の答え: デザインメモ（docs/ui-notes.md の例）</summary>

```markdown
# UIメモ
- 配色: 背景 slate-50 / 文字 slate-900 / 強調 indigo-600 / エラー red-600
- 余白: 部品間は 4（1rem）、カード内は p-3
- 文字: タイトル text-2xl font-bold / 本文 text-base
- 状態: hover で背景を少し濃く、focus は ring-2 ring-indigo-500、disabled は opacity-50
- 幅: 中央寄せ max-w-xl、スマホは左右に px-4
```
</details>

<details>
<summary>Step 3 の答え: レイアウト（Tailwind）</summary>

`page.tsx` の `return` 部分
```tsx
<main className="mx-auto max-w-xl px-4 py-8">
  <h1 className="mb-6 text-2xl font-bold text-slate-900">ToDo</h1>
  <TodoForm onAdd={addTodo} />
  {error && (
    <p role="alert" className="mt-3 text-sm text-red-600">{error}</p>
  )}
  ...
</main>
```

`TodoItem.tsx` の1行
```tsx
<li className="flex items-center gap-3 rounded-lg border border-slate-200 bg-white p-3">
  <input
    type="checkbox"
    className="h-5 w-5 accent-indigo-600"
    checked={todo.done}
    onChange={() => onToggle(todo.id)}
    aria-label={`${todo.title} を完了にする`}
  />
  <span className={`flex-1 break-words ${todo.done ? "text-slate-400 line-through" : "text-slate-900"}`}>
    {todo.title}
  </span>
  <button className="rounded px-2 py-1 text-sm text-slate-600 hover:bg-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500">
    編集
  </button>
  <button className="rounded px-2 py-1 text-sm text-red-600 hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-red-500">
    削除
  </button>
</li>
```
一覧: `<ul className="mt-4 space-y-2">`

注意: クラス名は**完成した文字列**で書く。`bg-${color}-500` のような動的な組み立ては検出されず、効かない。
</details>

<details>
<summary>Step 4 の答え: フィルタ</summary>

方式は「APIの `?status=` を使う」を選んだ場合。理由メモ例: 「データが増えてもサーバー側で絞れるため」。

`src/components/StatusFilter.tsx`
```tsx
import type { TodoStatus } from "@/lib/api";

const OPTIONS: { value: TodoStatus; label: string }[] = [
  { value: "all", label: "全て" },
  { value: "active", label: "未完了" },
  { value: "done", label: "完了" },
];

type Props = {
  value: TodoStatus;
  onChange: (value: TodoStatus) => void;
};

export function StatusFilter({ value, onChange }: Props) {
  return (
    <div role="group" aria-label="表示の絞り込み" className="mt-4 flex gap-2">
      {OPTIONS.map((o) => (
        <button
          key={o.value}
          aria-pressed={value === o.value}
          onClick={() => onChange(o.value)}
          className={`rounded-full px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 ${
            value === o.value ? "bg-indigo-600 text-white" : "bg-slate-100 text-slate-700 hover:bg-slate-200"
          }`}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}
```

`page.tsx` 側
```tsx
const [status, setStatus] = useState<TodoStatus>("all");

useEffect(() => {
  setLoading(true);
  listTodos(status)
    .then(setTodos)
    .catch((e: Error) => setError(e.message))
    .finally(() => setLoading(false));
}, [status]); // status が変わるたびに再取得
```
絞り込み中に追加したタスクが条件外のとき（例: 「完了」表示で追加）は、`status === "done"` のとき追加後の先頭挿入をしない、または追加後に `listTodos(status)` で再取得する。
</details>

<details>
<summary>Step 5 の答え: 入力フィードバック（TodoForm.tsx）</summary>

```tsx
"use client";

import { useState } from "react";

type Props = {
  onAdd: (title: string) => void | Promise<void>;
};

export function TodoForm({ onAdd }: Props) {
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = title.trim();
    if (trimmed.length < 1 || trimmed.length > 100) {
      setError("タイトルは1〜100文字で入力してください");
      return;
    }
    setError("");
    setSubmitting(true);
    try {
      await onAdd(trimmed);
      setTitle("");
    } finally {
      setSubmitting(false); // 成功でも失敗でも必ず戻す
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <label htmlFor="title" className="sr-only">新しいタスク</label>
      <div className="flex gap-2">
        <input
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="新しいタスクを入力"
          aria-invalid={error ? true : undefined}
          aria-describedby={error ? "title-error" : undefined}
          className="flex-1 rounded-lg border border-slate-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <button
          type="submit"
          disabled={submitting}
          className="rounded-lg bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700 disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          {submitting ? "追加中…" : "追加"}
        </button>
      </div>
      <div className="mt-1 flex justify-between text-sm">
        {error ? (
          <p id="title-error" role="alert" className="text-red-600">{error}</p>
        ) : (
          <span />
        )}
        <span className={title.trim().length > 100 ? "text-red-600" : "text-slate-500"}>
          {title.trim().length}/100
        </span>
      </div>
    </form>
  );
}
```
0件表示は `TodoList` の `<p className="mt-6 text-center text-slate-500">タスクがありません</p>`。
</details>

<details>
<summary>Step 6 の答え: アクセシビリティのチェックと修正</summary>

| 項目 | 直し方 |
|---|---|
| 入力欄のラベル | `<label htmlFor="title" className="sr-only">`（Step 5で対応済み） |
| 完了切替 | `div` ではなく `<input type="checkbox">`（対応済み） |
| 編集モードの入力欄 | `aria-label="タイトルを編集"` を付ける |
| 編集・削除ボタン | 対象が分かるように `aria-label={`${todo.title} を編集`}` |
| フォーカスの枠 | `outline-none` を使うときは必ず `focus:ring-*` を併用 |
| 文字色 | 薄い文字（`text-slate-400`）は完了タスクだけ。コントラスト比はDevToolsの「Accessibility」で確認 |
| エラー | `role="alert"` で自動的に読み上げ |

確認のしかた: マウスを使わず `Tab` / `Shift+Tab` / `Enter` / `Space` だけで、追加・完了・編集・削除ができるか。
</details>

<details>
<summary>Step 7 の答え: レスポンシブ確認</summary>

1. ブラウザの開発者ツール → デバイスツールバー（`Cmd + Shift + M`）→ 幅 375px
2. 確認: 横スクロールが出ない / ボタンが押しやすい / 長いタイトルが折り返す（`break-words`）
3. 崩れる場合: 行を `flex-wrap` にする、または小さい画面で操作ボタンを下段に回す
   ```tsx
   <li className="flex flex-wrap items-center gap-3 ...">
   ```
</details>

<details>
<summary>Step 8 の答え: rebase でコミット整理</summary>

```
git branch backup/09          # 失敗時の保険
git log --oneline main..HEAD  # 整理したいコミットを確認
git rebase -i main
```
エディタに次のように表示される（古い順）:
```
pick a1b2c3 feat: add status filter
pick d4e5f6 fix typo
pick 789abc style: improve layout with Tailwind
```
整理例（2行目を直前に統合、3行目は文言変更）:
```
pick a1b2c3 feat: add status filter
squash d4e5f6 fix typo
reword 789abc style: improve layout with Tailwind
```
保存して閉じると、メッセージ編集画面が順に開く。

困ったとき:
```
git rebase --abort                 # 途中でやめる
git reset --hard backup/09         # 終わった後に元へ戻す（※ 未コミットの変更は消える）
```
整理後 `git log --oneline` で確認し、問題なければ `git branch -D backup/09`。
</details>
