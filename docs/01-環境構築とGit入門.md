# Section 1: 環境構築とGit入門

## 1. 目的と完成の定義

**目的**: 開発に必要なツールを揃え、Gitの基本サイクル（変更 → ステージ → コミット → push）を一通り体験する。

**完成の定義**（すべて満たせば完了）:
- [ ] `git --version` `go version` `node -v` `npm -v` が表示される
- [ ] `git config` に名前とメールが設定されている
- [ ] `.gitignore` がある
- [ ] `main` に最初のコミットが2〜3個ある（`git log --oneline` で確認）
- [ ] GitHubにリポジトリがあり、`git push` 済み
- [ ] `docs/learning-log.md` に学んだことを書いた

## 2. 用語解説

| 用語 | 意味 |
|------|------|
| リポジトリ | Gitが履歴を管理しているフォルダ（`.git/` が本体） |
| 作業ツリー | 今ファイルを編集している場所 |
| ステージ（インデックス） | 次のコミットに含める変更の「準備置き場」 |
| コミット | ステージした変更を履歴として記録したもの |
| ブランチ | 履歴の枝。`main` が本流 |
| リモート / origin | GitHub上のリポジトリ / その標準的な呼び名 |
| push | ローカルのコミットをリモートへ送る |

```
作業ツリー --git add--> ステージ --git commit--> ローカル履歴 --git push--> GitHub
```

## 3. 手順（コミット単位）

> 現在の状態: `git init` 済み、`main` ブランチ、コミットなし。
> `go.mod` `go.sum` `main.go` が既にあるが、**今回はコミットしない**（Section 3で扱う）。

### Step 1: ツールの確認とインストール

```
git --version
go version      # 1.22以上推奨
node -v         # 20以上推奨（LTS）
npm -v
```
未インストールなら: Go（go.dev/dl）、Node.js（nodejs.org の LTS）、macOSでHomebrewを使うなら `brew install go node`。

つまずきやすい点: インストール後はターミナルを開き直さないとPATHが反映されない。

### Step 2: Gitの初期設定

```
git config --global user.name "あなたの名前"
git config --global user.email "メールアドレス"
git config --global init.defaultBranch main
git config --global --list      # 確認
```
- GitHubでメールを公開したくない場合は、GitHubの Settings > Emails にある `xxxx@users.noreply.github.com` を使う。

### Step 3: `.gitignore` を作る（コミット1つ目の準備）

「Gitに入れたくないファイル」を列挙する。このプロジェクトで入れるべきもの:

| 分類 | 例 | 理由 |
|------|----|------|
| OS/エディタ | `.DS_Store`, `.idea/`, `.vscode/` | 個人環境の産物 |
| 秘密情報 | `.env`, `.env.local` | パスワード・APIキーを公開しない |
| Node/Next.js | `node_modules/`, `.next/`, `out/` | 自動生成・巨大 |
| Go | ビルドしたバイナリ, `*.test`, `*.out` | 自動生成 |
| DB | `*.db`, `*.sqlite` | ローカルデータ |

**課題**: 上の表をもとに、自分で `.gitignore` を書く。分からない箇所はAIに質問してよいが、コピペせず意味を理解してから書くこと。
`.env.example`（値を空にした見本）はコミットしてよい点に注意。

### Step 4: 状態を確認して最初のコミット

```
git status                 # 未追跡ファイルが赤で出る
git add .gitignore docs/   # 必要なものだけ明示的に追加（git add . は使わない）
git status                 # 緑（ステージ済み）になったか確認
git diff --staged          # コミットされる内容を確認
git commit -m "chore: add .gitignore"
```
`docs/` のコミットは分ける練習として、`.gitignore` と別にしてもよい:
```
git add docs/
git commit -m "docs: add learning roadmap and section 1 guide"
```
確認: `git log --oneline` / `git status`（`go.mod` などが未追跡のまま残っていればOK）

### Step 5: GitHubにpush

1. GitHubで新規リポジトリを作成（**READMEや.gitignoreは追加しない**。ローカルと衝突するため）
2. 表示されるURLを使う:
```
git remote add origin https://github.com/<ユーザー名>/<リポジトリ名>.git
git remote -v
git push -u origin main
```
- 認証エラー時: パスワードではなくPersonal Access Token、またはGitHub CLI（`gh auth login`）を使う。
- `-u` は「以降 `git push` だけで済むようにする」設定。

### Step 6: 練習用ブランチで一周（サイクルの体験）

```
git switch -c docs/01-practice
# docs/learning-log.md を作成し、今日学んだこと・つまずきを3行以上書く
git add docs/learning-log.md
git commit -m "docs: add learning log"
git switch main
git merge docs/01-practice
git branch -d docs/01-practice
git push
```
確認: `git log --oneline --graph`

### Step 7: AI駆動サイクルの練習

1. AIに「Section 2の `docs/requirements.md` を作って」と依頼（ここは次のセクションの予習）
2. 生成されたファイルを読み、**気になる点を3つ以上**メモ
3. 修正を依頼するか、自分で直す
4. 初稿と修正後を別コミットにしてみる（`git diff` で差分を確認できる）

※ これはSection 2の先取りなので、Section 2で行ってもよい。

## 4. コマンド早見表

| 目的 | コマンド |
|------|----------|
| 状態確認 | `git status` |
| 未ステージの差分 | `git diff` |
| ステージ済みの差分 | `git diff --staged` |
| 履歴 | `git log --oneline --graph` |
| ステージ取り消し | `git restore --staged <file>` |
| 変更を破棄（※戻せない） | `git restore <file>` |
| 直前のコミットメッセージ修正（push前のみ） | `git commit --amend` |

## 5. つまずきやすい点

- `git add .` で意図しないファイル（`.env` など）を入れてしまう → 常に `git status` で確認、`add` はパス指定
- コミットメッセージを `-m` なしで実行するとエディタ（vim）が開く → `:wq` で保存終了、中止は `:q!`
- `main` と `master` の違い → 古いGitの標準名。今回は `main` に統一
- push前にコミットした内容が違うと気づいたら → `git commit --amend`。push後は履歴を書き換えない

## 6. 自己チェック（口頭で説明できるか）

1. 作業ツリー・ステージ・ローカル履歴・リモートの違いは？
2. なぜ `git add .` より `git add <path>` が安全？
3. `.gitignore` に書いたファイルが既にコミット済みの場合、どうなる？（調べてみよう）
4. `git commit` と `git push` の違いは？

## 7. 完了したら

- `docs/learning-log.md` に振り返りを追記
- 次は **Section 2: 要件定義・設計ドキュメント** へ（AIに作成を依頼）
