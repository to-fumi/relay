# Google でログインする（AUTH_MODE=oidc）

普段の開発は開発用ログイン（`AUTH_MODE=dev`）で行う。Google でのログインを確かめるときだけ、ここの手順で設定する。料金はかからない（支払い情報の登録も要らない）。

## 1. Google Cloud で OAuth クライアントを作る

1. [Google Cloud コンソール](https://console.cloud.google.com/) でプロジェクトを作る（例：`relay`）
2. 「API とサービス」→「OAuth 同意画面」（Google Auth Platform）
   - ユーザーの種類：外部（個人のアカウント）。会社の Google Workspace の中だけで使うなら内部
   - アプリ名とサポートのメールを入れる
   - スコープは `openid`・`email`・`profile` だけ（審査は要らない）
   - 公開ステータスは「テスト中」のままでよい。ログインする人を「テストユーザー」に足す（100 人まで）
3. 「認証情報」→「認証情報を作成」→「OAuth クライアント ID」
   - 種類：ウェブ アプリケーション
   - 承認済みのリダイレクト URI（使う URL をすべて足す。`localhost` 以外は HTTPS が要る）
     - Vite で開発するとき：`http://localhost:5173/api/auth/callback`
     - API サーバーから配信するとき：`http://localhost:8080/api/auth/callback`
     - スマホで確かめるとき（#13 のトンネル）：`https://<トンネルの URL>/api/auth/callback`
4. 表示されたクライアント ID とクライアント シークレットを控える（コミットしない）

## 2. バックエンドの設定（`backend-go/.env`）

```
AUTH_MODE=oidc
OIDC_ISSUER=https://accounts.google.com
OIDC_CLIENT_ID=<クライアント ID>
OIDC_CLIENT_SECRET=<クライアント シークレット>
OIDC_REDIRECT_URL=http://localhost:5173/api/auth/callback
ALLOWED_EMAILS=you@gmail.com
ADMIN_EMAILS=you@gmail.com
```

- `ALLOWED_EMAILS` にない人はログインできない（`/?login=forbidden` に戻る）。`@example.com` と書くと、そのドメインの人を許す
- `OIDC_REDIRECT_URL` が `https://` なら、Cookie に `Secure` を付ける

## 3. フロントエンドの設定（`web/.env`）

```
VITE_AUTH_MODE=oidc
```

## 4. 確かめる

```bash
pnpm task dev:api
pnpm task dev:web
```

1. Mac のブラウザで `http://localhost:5173` を開くと「Google でログイン」が出る
2. 押して Google でログインすると、受け渡しの画面に戻る
3. 右上に Google アカウントの名前が出る
4. `ALLOWED_EMAILS` にないアカウントでログインすると、「このアカウントでは使えません」が出る

Dev Container の中から Google に届かないときは、コンテナを作り直す（`.devcontainer/allowed-domains.txt` に Google のドメインを足してある）。
