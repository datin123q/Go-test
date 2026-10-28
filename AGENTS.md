# AGENTS.md

このファイルは、`energy-shares-backend-repo` で作業するエージェント向けの実行ルールです。
以降の指示は、`CLAUDE.md` を含む互換ファイルより優先される一次情報として扱ってください。

## 全リポジトリ共通ルール

[AGENTS-COMMON.md](/AGENTS-COMMON.md) を参照する。

## energy-shares-backend-repo 専用ルール

### プロジェクト概要

ESJP (EnergyShares Japan) backend は、AWS ECS 上で動く Go モノレポ。

- `admin-api`: 管理者向け API
- `user-api`: エンドユーザー向け API
- `jobs-scheduler`: バックグラウンドジョブ実行
- `invest-worker`: 投資処理ワーカー

主要スタック:

- Go 1.25.1
- Echo v4
- XORM (MySQL)
- Kafka
- uber/fx (DI)
- OpenAPI 3.1.0

### 実装原則

- シンプルさ優先: 必要最小限の変更で解く。
- 最小影響: 関連しないレイヤーやドメインへ波及させない。
- 手抜き禁止: 問題の根本に対処する。

### 主要コマンド

```bash
# Workspace setup (run once, or after go.work changes)
make work
make work-sync

# Run all tests (requires VPN + MFA session)
make test

# Run tests for a single app
cd apps/admin-api && go test -v ./...
cd apps/user-api && go test -v ./...

# Run a specific test
go test -v -run Test_adminUsecase_GetAll ./apps/admin-api/usecase/admin_user/...

# Generate mocks (requires: brew install mockery)
make mock

# Generate Swagger JSON
make gen-swagger

# Run apps locally
make start-admin-api    # admin-api on local env
make start-user-api     # user-api on local env
make start-jobs-scheduler

# Local MySQL
make mysql_dump          # dump DEV DB
make mysql_run           # start MySQL container
make db-setup            # seed schemas + run migrations
make db-migrate          # run migrations only
make seed-sample-data    # load sample data
make mysql-client        # connect to local MySQL

# DB migration: create new migration file
goose create <migration_name> sql
# Run migration
cd apps/admin-api && go run main.go migrate --env local

# Code generation from YAML definitions
make codegen

# Environment config (requires AWS MFA)
./paramstore.sh get dev       # download dev config
./paramstore.sh get all       # download all envs
./paramstore.sh set dev qa    # upload configs
```

### アーキテクチャ

Clean Architecture (4層) を採用し、`go.uber.org/fx` で依存を注入する。

```text
Delivery (HTTP/Kafka) -> Usecase (business logic) -> Service (shared logic) -> Repository (data access)
```

### 主要ディレクトリ

| Path                                 | Purpose                                                     |
| ------------------------------------ | ----------------------------------------------------------- |
| `clean-arch-domain/`                 | Repository/Service の共有インターフェース、クエリオプション |
| `domain/`                            | XORM タグ付き DB モデル                                     |
| `repository/{domain}/mysql/`         | MySQL リポジトリ実装                                        |
| `service/`                           | 認証・メール・S3 など横断サービス                           |
| `apps/{app}/domain/`                 | アプリ固有の Usecase I/F と DTO                             |
| `apps/{app}/usecase/{domain}/`       | Usecase 実装                                                |
| `apps/{app}/delivery/{domain}/http/` | Echo ハンドラ                                               |
| `apps/{app}/server/bootstrap.go`     | fx の DI 構成                                               |
| `pkg/`                               | ORM ラッパー、factory、config、補助関数                     |

### レイヤー規約

- Repository:
  - CRUD 命名を維持 (`Create`, `GetAll`, `GetById`, `GetBy{Field}`, `Update`, `Delete`)。
  - Delete はすべて soft delete (`DeletedAt` を設定)。
  - 絞り込み・ページングは functional options を使う。
- Service:
  - 認証、権限、メール送信など再利用ロジックを配置。
  - ドメインによっては Service 層を省略可。
- Usecase:
  - Delivery と 1:1 で対応。
  - 入出力は `{Domain}{Action}Request` / `{Domain}{Action}Response`。
  - Request には `JWTUser` と `ClientIp` を含めて監査情報を保持。
  - トランザクションは `factory.TxManager` を使用。
- Delivery:
  - Echo で bind/validate/JWT 抽出/エラー変換を担当。

### エラーハンドリング

- ドメインエラーは `clean-arch-domain/common.go` に定義 (`ErrItemNotFound`, `ErrAlreadyExists` など)。
- Delivery では `factory.ErrorTemplate` (`ErrorNotFound`, `ErrorParameter`, `ErrorDB` など) にマッピングして HTTP 応答を返す。

### DB 運用

- DB は `EnergySharesDB` (main) と `MyNumberDB` (separate) の 2 系統。
- DB インスタンスは `context.Context` で運び、`factory.DBFactory` から取得する。
- `pkg/orm/orm.go` の XORM ラッパーを使い、チェーン可能なクエリビルダを利用する。
- トランザクションは `txManager.WithTx(ctx, func(txCtx) error { ... })` で管理。

### マイグレーション状況

- リポジトリは legacy 構成から Clean Architecture へ段階移行中。
- 新規 API は `v1` を利用し、既存 API は契約互換を維持しながらドメイン単位で移行する。
- DB マイグレーションの追加・適用は上記 `goose` / `migrate` コマンドを使い、差分の整合性を確認する。

### テストとモック

- テストは table-driven + `testify` (`assert`/`mock`) が基本。
- モックは `mockery` で生成し、出力先は `clean-arch-domain/mocks/` と `apps/{app}/domain/mocks/`。
- 設定はリポジトリ直下 `.mockery.yaml` を参照。
- モック境界は Repository/Service/Factory などインターフェース境界に置く。

### OpenAPI / Swagger

- ルーティング定義は `pkg/echoutils/swagger.go` を利用し、ルート定義と OpenAPI 生成を同時に扱う。
- 開発・QA 環境では Swagger UI を自動配信する。
- Path parameter では Echo 用 `param` タグと OpenAPI 用 `path` タグの両方を付与する。
