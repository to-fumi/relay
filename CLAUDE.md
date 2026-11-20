# Relay

備品の受け渡しを記録し、そのログを分析・可視化するWebアプリ。

## 機能

### 受け渡しの記録（QR）
1. 渡す側：item（自由入力）と qty を入力し、QRを発行する
2. 受け取る側：QRを読み取る → `from=発行者、to=読み取った人` としてログに記録する

- QRの中身はサーバーが署名した短いトークン（`handoff/adapter/token`）。トークンはどこにも保存しない
  - 形式：`version(1) | jti(8) | from(16) | exp(uint32) | qty(uvarint) | item(UTF-8) | HMAC-SHA256の先頭12バイト` を Base32（パディングなし）にしたもの
  - Base32 は QR の英数字モードに収まるので、QR を荒くできる（`LANケーブル` で 92 文字、バージョン4、33×33マス）。QR の誤り訂正レベルは L
  - 有効期限は60秒程度
- 読み取りのときに署名と期限を検証する。`from == to` は拒否する
- 同じQRの二重記録は、`handoffs.jti` の一意制約で防ぐ

### 閲覧・分析（全ユーザー）
- 自分の履歴
- 全体とグループの受け渡しの流れ、分析（集計内容は別途決める）
- 関係図（人を点、受け渡しを線として表示する。Cytoscape.js）

### 管理（管理者）
- ユーザーとグループの管理（グループはアプリ内で管理する）
- ログの削除

ユーザーは初回ログイン時に自動で登録する。最初の管理者は環境変数 `ADMIN_EMAILS` で指定する。

## データモデル

```
users        id, entra_oid(unique), display_name, email, is_admin, created_at
groups       id, name
user_groups  user_id, group_id
handoffs     id, jti(unique), from_user, to_user, item(text), qty, created_at
```

グループごとの集計は、今の所属で行う。

## 認証

- Entra ID。React（MSAL React）でアクセストークンを取得し、`Authorization: Bearer` ヘッダーで送る
- バックエンドはJWTの署名、iss、aud、exp を検証する
- MSALのトークンは `sessionStorage` に保存する

## アーキテクチャ

モノレポ ＋ モジュラーモノリス ＋ Clean Architecture ＋ DDD。依存は常に片方向（DIP）。

```
relay/
├── api/openapi.yaml        # API仕様の正本（OpenAPI 3.1）
├── web/                    # React + Vite + TypeScript
└── backend-go/             # Go
```

### Backend（Go）

```
backend-go/
├── cmd/relay/main.go       # Composition Root：依存の組み立てはここだけで行う
└── internal/
    ├── handoff/            # 受け渡しの記録、QRトークン
    ├── identity/           # ユーザー、グループ、管理者
    ├── analytics/          # 集計、関係図用のデータ
    └── platform/           # DB接続、JWT検証ミドルウェア、設定、ログ
```

各モジュールの中は次のように分ける。

```
<module>/
├── domain/        # エンティティ、値オブジェクト、リポジトリのインターフェース。何にも依存しない
├── application/   # ユースケース、ポート（TokenSigner、Clock などのインターフェース）。domain だけに依存する
└── adapter/
    ├── http/      # oapi-codegen で生成したサーバーインターフェースの実装
    ├── postgres/  # リポジトリの実装（pgx）
    └── ...        # ポートの実装（例：handoff/adapter/token）
```

- 依存の向き：`adapter → application → domain`
- 他のモジュールの `domain` と `adapter` は import しない。モジュールの間で何かを使うときは、使う側の `application` にインターフェースを定義し、`main.go` で提供する側の実装を注入する
- 不変条件はドメインで守る（例：Handoff は `from != to`、`qty > 0`、`item` は空でない）

### Frontend（Feature-based）

```
web/src/
├── app/            # ルーティング、Provider（MSAL、TanStack Query）、レイアウト
├── features/
│   ├── handoff-issue/     # QR発行
│   ├── handoff-receive/   # QR読み取り
│   ├── history/           # 自分の履歴
│   ├── analytics/         # 流れ、分析、関係図
│   └── admin/             # ユーザーとグループの管理
└── shared/         # 生成したAPIクライアント、UI部品、共通の関数
```

- 依存の向き：`app → features → shared`。feature同士は import しない
- 各featureは `index.ts` で公開するものだけを外に出す

## 技術スタック

- **Frontend**：React 19、Vite 8、TypeScript 6.0（strict）、TanStack Query、React Router、MSAL React、qrcode.react（表示）、@yudiel/react-qr-scanner（読み取り）、Cytoscape.js。APIクライアントと型は @hey-api/openapi-ts で生成する
- **UI**：Tailwind CSS v4（`@tailwindcss/vite`）、lucide-react（アイコン）。スマホを前提にした1カラム（`max-w-lg`）と下部タブ。ダークモードは `prefers-color-scheme` に従う
  - className の結合は `shared/lib` の `cn`（tailwind-merge で衝突を解決する）を使う
  - 共通の UI 部品は `shared/ui`（Button、Card、Field、Alert、EmptyState、PageHeader）
  - 日時は `shared/lib` の `formatDateTime` などで ja-JP 形式にする
  - 「渡す」以外の画面は `React.lazy` で遅延読み込みする
- **Backend**：Go 1.27（`go.mod` の `toolchain` で patch バージョンまで固定する。標準ライブラリの net/http、oapi-codegen の strict-server、pgx、goose、golang-jwt）、PostgreSQL 18
- **Node**：24（`.node-version`）、pnpm（`packageManager` で固定する）
- **配置**：SPAはAPIサーバーから一緒に配信し、1つのコンテナにまとめる
- **テスト**：APIテストはバックエンドの実装に依存しない形で書く（トークンは1回しか使えない、期限切れは失敗する、自分には渡せない、など）

## コマンド

```bash
pnpm install                 # 依存パッケージと Git フック（lefthook）を入れる
pnpm task db:up              # PostgreSQL を起動する（ポートは RELAY_DB_PORT で変更できる。初期値は 5432）
pnpm task dev:api            # API を起動する（backend-go/.env を読む。.env.example をコピーして作る）
pnpm task dev:web            # Vite を起動する（/api は localhost:8080 にプロキシする）
pnpm task gen                # openapi.yaml を変更したら実行する（Go と TS のコードを生成する）
pnpm task fmt                # 整形する
pnpm task check              # CI と同じチェックをすべて実行する
```

- `task` は npm の `@go-task/cli` で入る（`pnpm task ...` で呼ぶ）。Go の開発ツールは `backend-go/tools/go.mod` で管理し、`go tool -modfile=tools/go.mod <tool>` で呼ぶ
- 開発中は `AUTH_MODE=dev`（バックエンド）と `VITE_AUTH_MODE=dev`（フロントエンド）で仮ログインを使う。トークンは `dev:<email>`
- API を追加するときの順番：`api/openapi.yaml` を変更する → `pnpm task gen` → `cmd/relay/server.go` の `server` がまだ満たしていないメソッドがコンパイルエラーになるので、担当するモジュールの `adapter/http` に実装する
- DB のマイグレーションは `backend-go/internal/platform/postgres/migrations/` に goose の形式で追加する（起動時に自動で適用する）

## 開発ツールとCI

すべてのコマンドは Taskfile（`Taskfile.yml`）に定義する。lefthook と CI はどちらも `task ...` を呼ぶだけにする。

| 対象 | フォーマット・Lint | 型・テスト | 依存の向き |
|---|---|---|---|
| web/ | Biome（`pnpm exec biome`） | `tsc --noEmit` | dependency-cruiser（app → features → shared、feature同士の import 禁止） |
| backend-go/ | golangci-lint v2（gofumpt、goimports を含む） | `go test -race ./...`、govulncheck | depguard（domain は何も import しない、application は adapter を import しない、モジュール同士の直接の import 禁止） |
| api/ | Redocly CLI | — | — |

- **パッケージ管理**：pnpm（`--frozen-lockfile`）、Go modules
- **生成コードの確認**：CIでコードを生成し直し、`git diff --exit-code` で差分がないことを確認する（Go側とTS側の両方）
- **Git フック**：lefthook（pre-commit でフォーマットと Lint を実行する）
- **CI**：GitHub Actions。paths フィルターで、変更されたディレクトリのジョブだけを実行する
- **依存の更新**：Renovate
- **改行コード**：LF に統一する（`.gitattributes` に `* text=auto eol=lf`、`biome.json` に `formatter.lineEnding: "lf"`）
