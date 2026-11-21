# 作業の進め方（AI）

人は Issue の作成、プロンプトでの指示、PR のレビューとマージをする。AI は Issue やプロンプトをもとに実装し、確認まで済ませて PR を作る。

## 始める前

- 作業ブランチは `develop` から切る。名前は `<種類>/<内容>`（例：`feat/return-reminder`、`fix/qr-expire`）。種類は feat・fix・refactor・test・docs・chore
- 受け入れ条件が書かれていない Issue やプロンプトは、読み取った受け入れ条件を PR の概要に書いて進める
- 次のときは実装を始めず、先に Issue のコメント（プロンプトならその返答）で確認する
  - 読み方が 2 通り以上あり、どちらにするかで作るものが変わる
  - DB のマイグレーションを追加・変更する
  - `api/openapi.yaml` の既存の API を、互換性がなくなる形で変える
  - 依存パッケージを追加する、メジャーバージョンを上げる
  - 認証（`platform/auth`、`shared/auth`）やトークンの形式（`handoff/adapter/token`）を変える

## 完了の条件

- `pnpm task check` が通る（CI と同じ）
- 振る舞いを変えたらテストを足す（ドメインとユースケースは Go のテスト）
- UI を変えたら、画面を開いて確かめる（iPhone の幅 390×844、ライトとダーク）
  - スクショなど確認用のファイルは `artifacts/` に置く。`.gitignore` してあるので、コミットしない
- 機能・データモデル・UI の方針を変えたら、`CLAUDE.md` の該当する記述も直す

## コミットと PR

- 1 つの PR は 1 つの目的にする。大きくなるときは分ける
- コミットメッセージと PR のタイトルは英語の Conventional Commits（例：`feat: add return reminders`）
- lefthook を飛ばさない（`--no-verify` を使わない）

## Issue・PR・コメントの書き方

読むのはスマホのことが多いので、短く書く。日本語で、「〜する」「〜した」で終える。

- PR の本文は `.github/pull_request_template.md` に沿って書く。「概要」は 300 字以内。細かい説明は `<details>` の中に入れる
- PR と Issue へのコメントは 300 字以内。結論を先に書く
- Issue は `.github/ISSUE_TEMPLATE/` の形に沿って書く。各項目は短く箇条書きにする
- UI を変えたときは、確かめた画面と結果を PR の「確認」に書く

## CI が落ちたら

- ログを読んで直し、push し直す。同じ原因で 3 回直せなかったら手を止め、状況を PR にコメントして人に任せる
- テストを消したり、`t.Skip` や lint を無効にするコメントで通したりしない

## マージ

- マージは人がする。AI は頼まれたときだけ `gh pr merge` を使う（`.claude/settings.json` で確認を求めるようにしてある）

## ガードレールを育てる

- 同じ指摘を 2 回受けたら、このファイル・`CLAUDE.md`・CI のどれかで防げるようにする PR を出す
