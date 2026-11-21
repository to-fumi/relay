# Relay

備品の受け渡しを記録し、そのログを分析・可視化するWebアプリ。

## 機能

### 受け渡しの記録（QR）
1. 渡す側：item（自由入力）と qty、任意で返却期限（本日・翌日・2週間・カレンダーで任意の日付）を入力し、QRを発行する
2. 受け取る側：QRを読み取る → `from=発行者、to=読み取った人` としてログに記録する

返却期限がない受け渡しはログに残るだけ。返却期限がある受け渡しは、返却されるまで返却タブの「返す」（自分が受け取ったもの）と「待ち」（自分が送ったもの）のリストに入る。

受け渡しは 2 人が同時に行うので、受け渡しタブ（⇄）に「送る」（QR を出す）と「受け取る」（QR を読む）を並べる。返却は返却タブにまとめる。

- 返却期限は日付のみで、その日の終わりまで。発行のときに今日（日本時間）以降であることを確かめる
- 返却もQRで行う：借りた側（to）が返却タブの「返す」で返却QRを発行する（`POST /return-tokens`）→ 貸した側（from）が返却タブの「待ち」の「読取」で読み取る（`POST /returns`）→ `returned_at` を記録する
- 間違えて「受け取る」で返却QRを読んだときも記録できるようにする：API が type `urn:relay:problem:return-token` の 400 を返すので、同じトークンで `POST /returns` を呼び直す

- QRの中身はサーバーが署名した短いトークン（`handoff/adapter/token`）。トークンはどこにも保存しない
  - 受け渡し：`tag(1)=2 | jti(8) | from(16) | exp(uint32) | due(uint16、1970-01-01からの日数、0は期限なし) | qty(uvarint) | item(UTF-8) | HMAC-SHA256の先頭12バイト` を Base32（パディングなし）にしたもの
  - 返却：`tag(1)=3 | handoff_id(16) | exp(uint32) | HMAC-SHA256の先頭12バイト`（53 文字）
  - Base32 は QR の英数字モードに収まるので、QR を荒くできる（`LANケーブル` で 95 文字、バージョン4、33×33マス）。QR の誤り訂正レベルは L
  - 有効期限は60秒程度
- 読み取りのときに署名と期限を検証する。`from == to` は拒否する
- 同じQRの二重記録は、`handoffs.jti` の一意制約で防ぐ
- 再発行した QR も、古い QR は期限まで有効なまま（トークンを保存しないので取り消せない）。同じものが 2 件記録されないよう、画面で次のようにする
  - 送る側は、発行の応答の `jti` と一致する記録（`GET /me/handoffs` の `jti`）が出てくるまで確認し続け、出てきたら QR を消して「完了」にする
  - 再発行は、今の QR の期限が切れてからだけにする（有効な QR を同時に 2 つ作らない）。返却QRは二重に記録されない（返却済みかで判定する）が、操作をそろえるため同じにする

### 閲覧・分析（全ユーザー）
- 自分の履歴（操作の記録。送った・受け取った・返した・返ってきたを時刻順に並べる。返却期限は出さない）
- 全体とグループの受け渡しの流れ、分析、関係図（Cytoscape.js）は一旦なし（画面とタブを外している。作るときは `features/analytics` に置く）

### 管理（管理者）
- ユーザーとグループの管理、ログの削除は一旦なし（画面とタブを外している。作るときは `features/admin` に置き、管理者だけに下部タブを出す）

ユーザーは初回ログイン時に自動で登録する。最初の管理者は環境変数 `ADMIN_EMAILS` で指定する。

## データモデル

```
users        id, entra_oid(unique), display_name, email, is_admin, created_at
groups       id, name
user_groups  user_id, group_id
handoffs     id, jti(unique), from_user, to_user, item(text), qty, created_at, due_date(date, null), returned_at(null)
```

`returned_at` は `due_date` があるときだけ入る。

グループごとの集計は、今の所属で行う。

## 認証

- Entra ID。React（MSAL React）でアクセストークンを取得し、`Authorization: Bearer` ヘッダーで送る
- バックエンドはJWTの署名、iss、aud、exp を検証する
- MSALのトークンは `sessionStorage` に保存する（タブを閉じれば消える）
- ログアウトのボタンは置かない。アクセストークンは MSAL が自動で更新し、更新できないとき（`InteractionRequiredAuthError`）や API が 401 を返したときは、保存済みのトークンを消してログインし直してもらう（`shared/auth` の `expireSession`）
- 開発用ログイン（`AUTH_MODE=dev`）も 12 時間で期限切れにし、期限切れのものは読むときに消す

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
│   ├── returns/           # 返却リスト（返す・待ち）、返却QRの発行と読み取り
│   └── history/           # 自分の履歴
└── shared/         # 生成したAPIクライアント、UI部品、共通の関数
```

- 依存の向き：`app → features → shared`。feature同士は import しない
- 各featureは `index.ts` で公開するものだけを外に出す

## 技術スタック

- **Frontend**：React 19、Vite 8、TypeScript 6.0（strict）、TanStack Query、React Router、MSAL React、qrcode.react（表示）、@yudiel/react-qr-scanner（読み取り）、Cytoscape.js。APIクライアントと型は @hey-api/openapi-ts で生成する
- **UI**：Tailwind CSS v4（`@tailwindcss/vite`）、lucide-react（アイコン）。スマホを前提にした1カラム（`max-w-lg`）と下部タブ。ダークモードは `prefers-color-scheme` に従う
  - className の結合は `shared/lib` の `cn`（tailwind-merge で衝突を解決する）を使う
  - 文章は極力書かない。説明文や補足は置かず、アイコンで伝えられるものはアイコンにする（アイコンだけの操作には `aria-label` を付ける）
  - 視線が Z 字（左 → 右 → 左下 → 右）に流れるように並べる。各行は左に主な情報（品目・相手など）、右に値や操作（数量・期限・ボタンなど）を置く。見出しは左にタイトル、右に操作
  - 共通の UI 部品は `shared/ui`（Button、IconButton、Fab、Card、Field、Alert、EmptyState、PageHeader、IconRow、DueBadge、TokenQr。カメラの QrScanner は重いので `shared/ui/QrScanner` から直接 import する）
  - 日時は `shared/lib` の `formatDateTime` などで ja-JP 形式にする
  - 「送る」以外の画面は `React.lazy` で遅延読み込みする
  - 下部タブは「⇄ 受け渡し／↩ 返却／履歴」。タブの中は上の切り替えで分ける（受け渡しは「送る／受け取る」、返却は「返す／待ち」。`app/HandoffLayout`、`app/ReturnLayout`）
  - 画面の主な操作（QR を発行、返却QRを読み取る、続けて読み取る）は、右下に浮かせたアイコンだけの丸いボタン（`shared/ui` の `Fab`）にする。「受け取る」はカメラがすぐ動くので、読み取り中のボタンは置かない。送るの入力欄で Enter を押しても発行しない（品目 → 数量へ移るだけ。日本語の変換確定の Enter はそのまま通す）
  - 1 画面の機能は少ないので、余白は広めにとる（カードは `p-6`、見出しの下は `mb-8`）。入力は中身に合った大きさにする（例：数量は右寄せの小さなステッパー）
- **Backend**：Go 1.27（`go.mod` の `toolchain` で patch バージョンまで固定する。標準ライブラリの net/http、oapi-codegen の strict-server、pgx、goose、golang-jwt）、PostgreSQL 18
- **Node**：24（`.node-version`）、pnpm（`packageManager` で固定する）
- **配置**：SPAはAPIサーバーから一緒に配信し、1つのコンテナにまとめる
- **外部への通信**：社内のプロキシで止まらないように、実行時に外部の CDN などへ取りに行くものは置かない（Entra ID のログインだけは例外）。QR を解読する wasm（`zxing-wasm`）もアプリと一緒に配信する（`shared/ui/QrScanner` の `prepareZXingModule`）。wasm と、ライブラリが使う JS のバージョンは `pnpm-workspace.yaml` の `overrides` で 1 つにそろえる
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
