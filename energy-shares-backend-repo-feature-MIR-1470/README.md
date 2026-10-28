# ESJP (EnergyShares Japan) バックエンドプロジェクト

<!-- toc -->

- [概要](#%EA%B0%9C%EC%9A%94)
- [共通環境](#%EA%B3%B5%ED%86%B5%ED%99%98%EA%B2%BD)
    * [言語](#%EC%96%B8%EC%96%B4)
    * [プロジェクトインフラ](#%ED%94%84%EB%A1%9C%EC%A0%9D%ED%8A%B8-%EC%9D%B8%ED%94%84%EB%9D%BC)
- [設定](#%EC%84%A4%EC%A0%95)
    * [ワークスペース 設定](#%EC%9B%8C%ED%81%AC%EC%8A%A4%ED%8E%98%EC%9D%B4%EC%8A%A4-%EC%84%A4%EC%A0%95)
    * [テスト](#%ED%85%8C%EC%8A%A4%ED%8A%B8)
    * [Mocking](#mocking)
    * [Swagger](#swagger)
- [プロジェクトレイアウト (進行中)](#%ED%94%84%EB%A1%9C%EC%A0%9D%ED%8A%B8-%EB%A0%88%EC%9D%B4%EC%95%84%EC%9B%83-%EC%A7%84%ED%96%89%EC%A4%91)
    * [Go Clean Architecture テンプレート](#go-clean-architecture-%ED%85%9C%ED%94%8C%EB%A6%BF)
    * [インターフェースの位置](#%EC%9D%B8%ED%84%B0%ED%8E%98%EC%9D%B4%EC%8A%A4-%EC%9C%84%EC%B9%98)
    * [Repository Layer](#repository-layer)
        + [作成](#%EC%83%9D%EC%84%B1)
        + [リスト](#%EB%A6%AC%EC%8A%A4%ED%8A%B8)
        + [詳細・単件取得](#%EC%83%81%EC%84%B8-%EB%8B%A8%EA%B1%B4%EC%A1%B0%ED%9A%8C)
        + [更新](#%EC%97%85%EB%8D%B0%EC%9D%B4%ED%8A%B8)
        + [削除](#%EC%82%AD%EC%A0%9C)
    * [Service Layer](#service-layer)
    * [Usecase Layer](#usecase-layer)
    * [Delivery Layer](#delivery-layer)
- [Swagger 管理](#swagger-%EA%B4%80%EB%A6%AC)
    * [実行方法および概要](#%EC%8B%A4%ED%96%89%EB%B0%A9%EB%B2%95-%EB%B0%8F-%EA%B0%9C%EC%9A%94)
    * [実装パッケージの主な機能](#%EA%B5%AC%ED%98%84-%ED%8C%A8%ED%82%A4%EC%A7%80-%EC%A3%BC%EC%9A%94%EA%B8%B0%EB%8A%A5)
        + [主な構造体](#%EC%A3%BC%EC%9A%94-%EA%B5%AC%EC%A1%B0%EC%B2%B4)
        + [使い方](#%EC%82%AC%EC%9A%A9-%EB%B0%A9%EB%B2%95)
    * [Request/Response 関連タグ](#requestresponse-%EA%B4%80%EB%A0%A8-%ED%83%9C%EA%B7%B8)
    * [注意事項](#%EC%A3%BC%EC%9D%98-%EC%82%AC%ED%95%AD)
- [DB Migration手順](#db-migration-%EC%A0%88%EC%B0%A8)
    * [ローカルMySQLコンテナの実行](#%EB%A1%9C%EC%BB%AC-mysql-%EC%BB%A8%ED%85%8C%EC%9D%B4%EB%84%88-%EC%8B%A4%ED%96%89)
    * [gooseの使い方](#goose-%EC%82%AC%EC%9A%A9%EB%B2%95)
    * [PR 作成](#pr-%EC%83%9D%EC%84%B1)
- [環境変数管理](#%ED%99%98%EA%B2%BD-%EB%B3%80%EC%88%98%EC%9D%98-%EA%B4%80%EB%A6%AC)

<!-- tocstop -->

## 概要

本リポジトリは、ESJP サービスを実装するための API サーバーを実装するリポジトリです。モノレポで構成され、各アプリごとに独立したサービスを提供します。

- admin-api: 管理者APIサーバー
- user-api: UserAPIサーバー
- job-scheduler: schedulerサーバー

## 共通環境

### 言語

現在のプロジェクトは Go 1.25.1 を使用しています。

### プロジェクトインフラ

- AWS RDS MYSQL 8.0.32 (dev, qa)
- AWS RDS MYSQL 5.7 (stage, prod)
- AWS MSK (dev, stage, prod)
- AWS ECS (dev, qa, stage, prod)
- AWS S3
- AWS CloudFront

## 設定

### ワークスペース設定

workspaceを使用して複数のプロジェクトを管理しています。本プロジェクトを実行するにはワークスペースの初期化が必要です。

```bash
# ワークスペース 作成 - 既存のgo.workを削除してから実行すること
$ make work

# ワークスペース同期
$ make work-sync
```

### テスト
すべてのアプリをテストします。

```bash
$ make test
```

⚠️ テストの実行は、VPN 接続済みで MFA 認証セッションが有効なターミナルで行ってください。設定が済んでいない場合は、以下のドキュメントを参照して設定してください。

- [VPN Client(MacOS) のインストールおよび使用方法](https://illuminarean.atlassian.net/wiki/spaces/development/pages/179241004?atlOrigin=eyJpIjoiZjVhNzYzZjM1MTU3NDI3Mzk1YzY2NGM3MDQ3ZGE4ZGYiLCJwIjoiY29uZmx1ZW5jZS1jaGF0cy1pbnQifQ)
- [AWS アカウントの MFA 設定方法](https://illuminarean.atlassian.net/wiki/spaces/development/pages/833781844?atlOrigin=eyJpIjoiODc0YTlmNzA3ZDYxNDk5ZThmZTY4MmMzNDdiOGJkNmEiLCJwIjoiY29uZmx1ZW5jZS1jaGF0cy1pbnQifQ)

それでも失敗する場合は、権限が不足している可能性があります。DevOps にお問い合わせください。

### Mocking

テストのため、プロジェクトではGo Mockeryを使用しています。

```bash
# インストール
brew install mockery

# アップグレード
brew upgrade mockery
```

以下のコマンドを実行すると、各プロジェクト内に Mocking ファイルが生成されます。

```bash
$ make mock
```

```≈
.
├── clean-arch-domain/
│   └── mocks
└── apps/
    ├── admin-api/
    │   └── domain/
    │       └── mocks
    ├── user-api/
    │   └── domain/
    │       └── mocks
    └── job-scheduler/
        └── domain/
            └── mocks
```

### Swagger

各プロジェクトは Swagger を使用してAPIドキュメントを提供します。以下のコマンドで各プロジェクトの Swagger ドキュメントを確認できます。

```bash
$ make gen-swagger

Generating ./swagger/swagger-admin-api.json...
Energy Shares Admin API Swagger JSON generating.....

...

Generating ./swagger/swagger-user-api.json...
Energy Shares User API Swagger JSON generating.....

...

Generating ./swagger/swagger-job-scheduler-api.json...
Energy Shares Jobs Scheduler API Swagger JSON generating.....

```

```
.
└── swagger/
    ├── swagger-admin-api.json
    ├── swagger-job-scheduler-api.json
    └── swagger-user-api.json
```

## プロジェクトレイアウト (進行中)

プロジェクトレイアウトは Clean Architecture に基づいて構成されています。各レイヤーは独立しており、golang-fx
ライブラリを用いて依存性注入を行っています。

🚧現在、プロジェクトは Clean Architecture 構造への移行リファクタリング中です。以下の原則に基づいて進行しています。🚧

1. 新規に追加する API は Clean Architecture 構造で実装します。
2. 既存の API はドメイン単位でリファクタリングします。
3. 改善された API には v1 プレフィックスを付けて実装します。
4. 内部テスト後、フロントエンドに共有し、v1 への移行に問題がなければ既存 API を削除します。
5. 移行の過程では、必要に応じてロジックに触れない範囲でリファクタリングを行い、リクエストやレスポンスの値は変更しません。

### Go Clean Architecture テンプレート

Clean Architecture 構造の理解が難しい場合は、[Go Clean Architecture 스켈레톤](https://github.com/illuminarean/go-clean-arch)を参照してください。

### インターフェースの位置

インターフェースは、共通的に使用する部分はルートフォルダに、ドメインごとに使用するインターフェースは各ドメインフォルダに配置します。

- 共通で使用するインターフェース
    - repository layer
    - service layer
- アプリごとに使用するインターフェース
    - usecase layer

```
.
├── clean-arch-domain/
│   └── {domain-name}.go // 共通インターフェースの位置
├── repository/
│   └── mysql/
│       └── {domain-name}_mysql.go
├── service/
│   └── {domain-name}_svc.go
└── apps/
    └── {app_name}/
        ├── domain/
        │   └── {domain-name}.go // アプリ別インターフェースの位置
        ├── usecase/
        │   └── {domain-name}_ucase.go
        └── delivery/
            └── http/
                └── {domain-name}_http.go
```

### Repository Layer

Repository Layer は、データの保存や照会を担当します。MySQL などのデータベースを使用する場合、各 Repository を実装します。

Repository Layer におけるドメインは、論理的または物理的に DB モデルと同一にマッピングされます。（将来的に改善可能）

- 作業時の命名は原則として CRUD に基づき、以下の関数を基準に可能な限り実装します。状況に応じて追加のインターフェースが必要になる場合があります。
    - Create (作成)
    - GetAll(リスト)
    - GetById (詳細・単件取得)
    - Update (更新)
    - Delete (削除)

#### 作成

一般的に、モデルを受け取り、そのまま xorm.Session オブジェクトを利用して保存します。

```go
func (c contractMysqlRepository) Create(ctx context.Context, contract domain.Contract) (int64, error) {
db, err := c.dbFactory.EnergySharesDB(ctx)
if err != nil {
return 0, errors.Wrap(err, "err get db")
}

if _, err = db.InsertOne(&contract); err != nil {
return 0, errors.Wrap(err, "err insert contract")
}

return contract.Id, nil
}

```

#### リスト

リストでは必須で受け取る必要のないデータについては Optional パターンを用いてフィルタリングやページングを実装します。内部実装は、独自に xorm をラップした orm パッケージを用いて行います。

🚧リファクタリング時には、リストを返す関数をすべて `GetAll` に統合できるよう改善します。🚧

```go
func (c contractMysqlRepository) GetAll(ctx context.Context, options ...cdomain.ContractQueryOption) (domain.Contracts, error) {
q, err := c.dbFactory.EnergySharesDB(ctx)
if err != nil {
return nil, errors.Wrap(err, "err get db")
}

opts := cdomain.ContractQueryOptionFunc()
for _, opt := range options {
opt(&opts)
}

contracts := make([]domain.Contract, 0)

err = orm.Table(q, domain.Contract{}).
Select(domain.Contract{}).
SortOrder(opts.SortBy(), opts.Order()).
Equal(domain.Contract{}.TableName()+".share_id", opts.ShareId()).
Limit(opts.Limit(), opts.Offset()).
Find(&contracts)

if err != nil {
return nil, errors.Wrap(err, "err get contracts")
}

return contracts, nil
}
```

#### 詳細・単件取得

単件照会は ID を用いて行い、ID がない場合はエラーを返します。ID 以外の条件で照会して単件が保証される場合は、単件照会を追加で実装できます。
`(ex. GetBy{Field})`

```go
func (s spvMysqlRepository) GetById(ctx context.Context, id int64) (domain.Spv, error) {
db, err := s.dbFactory.EnergySharesDB(ctx)
if err != nil {
return domain.Spv{}, errors.Wrap(err, "failed to get db")
}

var spv domain.Spv

has, err := orm.Table(db, domain.Spv{}).Id(id).Get(&spv)

if err != nil {
return domain.Spv{}, errors.Wrap(err, "failed to get spv by id")
}

if !has {
return domain.Spv{}, errors.Wrap(cdomain.ErrItemNotFound, "spv not found")
}

return spv, nil
}
```

#### 更新

更新は、モデルを受け取り、そのモデルの ID を用いて行います。実装者の判断でフィールド全体をカバーしても、必要な特定フィールドのみ変更しても構いません。ただし、このインターフェース実装を利用する側の観点から、簡単でもコメントを付けておくと望ましいです。

```go

// 全フィールドの部分更新（Partial Update）
func (s spvMysqlRepository) Update(ctx context.Context, spv domain.Spv) error {
db, err := s.dbFactory.EnergySharesDB(ctx)
if err != nil {
return errors.Wrap(err, "failed to get db")
}

if _, err = orm.Table(db, domain.Spv{}).Id(spv.Id).Update(&spv); err != nil {
return errors.Wrap(err, "failed to update spv")
}

return nil
}

// 一部フィールドの部分更新（Partial Update）
// name, email フィールドのみ更新
func (s spvMysqlRepository) Update(ctx context.Context, spv domain.Spv) error {
db, err := s.dbFactory.EnergySharesDB(ctx)
if err != nil {
return errors.Wrap(err, "failed to get db")
}

if _, err = orm.Table(db, domain.Spv{}).Id(spv.Id).Cols("name", "email").Update(&spv); err != nil {
return errors.Wrap(err, "failed to update spv")
}

return nil
}
```

#### 削除

削除は ID を用いて行います。すべての削除は Soft Delete で実施します。

```go
// Soft Delete
func (s spvMysqlRepository) Delete(ctx context.Context, id int64) error {
db, err := s.dbFactory.EnergySharesDB(ctx)
if err != nil {
return errors.Wrap(err, "failed to get db")
}

if _, err = orm.Table(db, domain.Spv{}).Id(id).Update(&domain.Spv{DeletedAt: time.Now()}); err != nil {
return errors.Wrap(err, "failed to delete spv")
}

return nil
}
```

### Service Layer

Service Layer は Usecase Layer で必要となるロジックを実装するためのレイヤーです。以下の特徴があります。

- Usecase Layer で必要なロジックを実装します。
- Repository Layer または他の Service Layer を参照して実装できます。
- Usecase Layer で頻繁に使用されるロジックは Service Layer として実装され得ます（会員認証、権限チェック、プロジェクト終了処理など）。
- Service Layer は必須ではなく、必要に応じて実装します。

```
Repository -----> Service -----> Usecase3
    |                |
    |                |
    |                v
    |             Service
    |                |
    |                |
    v                v
Usecase1          Usecase2
```

### Usecase Layer

Usecase Layer はビジネスロジックを担当するレイヤーです。以下の特徴があります。

- Usecase インターフェースのリクエストとレスポンスは、Delivery Layer で定義した DTO を使用します。
- Usecase の DTO は以下の命名規則を持ちます。
    - リクエスト : {domain-name}{function-name}Request
    - 응レスポンス답 : {domain-name}{function-name}Response

 ```go
type InvestmentUsecase interface {
GetAll(ctx context.Context, request InvestmentGetAllRequest) (InvestmentGetAllResponse, error)
}
``` 

- Usecase Layer は Repository Layer および Service Layer を参照できます。
- Usecase Layer は Delivery Layer と 1:1 で対応付けられる必要があります。

### Delivery Layer

Delivery Layer は外部からのリクエストを受け取り、Usecase Layer へ渡す役割を担います。以下の特徴があります。

- Delivery Layer は HTTP、gRPC、Kafka などのプロトコルを使用できます。
- Delivery Layer は Usecase Layer を参照し、ユーザーから受け取ったデータを Usecase Layer に渡します。
- Delivery Layer は Usecase Layer と 1:1 で対応付けられる必要があります。
- Delivery Layer では受け取ったリクエストのバリデーションを行います。
- Delivery Layer では Usecase から受け取ったエラーを返却するか、ユーザーに合わせたエラーに変換して返却します。
- Delivery Layer では認証済みユーザーの情報を Usecase Layer へ渡す必要があります。

```go
func (d investmentHttpDelivery) UpdateStatus(c echo.Context) error {
// ユーザー情報の抽出
user, err := d.extractor.Extract(c)
if err != nil {
return apirender.RenderFail(c, err)
}

ctx := c.Request().Context()

// Request DTO 生成
var request = adomain.InvestmentUpdateStatusRequest{
JWTUser: user,
}

// Request DTO Bind
if err = c.Bind(&request); err != nil {
return apirender.RenderFail(c, factory.ErrorParameter.New(err))
}

// Request DTO バリデーション
if err = c.Validate(&request); err != nil {
return apirender.RenderFail(c, factory.ErrorParameter.New(err))
}

response, err := d.investmentUsecase.UpdateStatus(ctx, request)
// Usecaseからのエラー処理
switch {
case err == nil:
return apirender.RenderSucc(c, http.StatusOK, response)
case errors.Is(err, cdomain.ErrItemNotFound):
return apirender.RenderFail(c, factory.ErrorNotFound.New(err))
default:
return apirender.RenderFail(c, factory.ErrorDB.New(err))
}
}
```

## Swagger 管理

現在 Swagger はアプリごとに分けて生成し、2 つの方式で管理しています。

- https://github.com/pangpanglabs/echoswagger ライブラリで管理する方式 (過去に運用していた方式で今後は使用しない）
- https://github.com/swaggest/openapi-go ライブラリを用い、直接管理する方式（今後運用する方式）

### 実行方法および概要

従来と異なり、現在のバージョンは dev/qa 環境で起動すると swagger ファイルが自動生成され、静的に提供されます。実装は openapi-go を使用していますが、実際の使用感では既存の echoswagger の利便性も考慮して開発しています。

**バージョン情報**

- swagger-ui : 5.17.0
- openapi : 3.1.0

### 実装パッケージの主な機能

- pkg/openapi/swagger.go
    - swagger-uiのダウンロード
    - swagger-uiに関する設定
- pkg/echoutils/swagger.go
    - Echo のルーターと OpenAPI ドキュメントを同時生成
    - HTTP メソッド（GET, POST, PUT, PATCH, DELETE）をサポート
    - ルートのグルーピングをサポート
    - タグ、サマリー、説明の追加
    - リクエスト/レスポンス構造の定義
    - エラーレスポンスの自動生成
    - セキュリティ設定の追加

#### 主な構造体

- EchoSwagger: Echo インスタンスと OpenAPI リフレクターをラップするメイン構造体
- echoGroupSwagger: グループ化されたルートを処理する構造体
- echoMethodSwagger: 個別の HTTP メソッドおよびパスに対する OpenAPI の操作を処理する構造体

#### 使い方

- NewEchoSwagger を使用して EchoSwagger インスタンスを生成します。
- Group メソッドでルートグループを作成します。
- HTTP メソッド（GET, POST など）でルートを定義します。
- 各ルートに対して、タグ、サマリー、説明、リクエスト/レスポンス構造、エラーなどを定義します。
- RegisterOpenAPI を呼び出して OpenAPI ドキュメント（openapi.yaml）を登録します。
- AddErrors は factory パッケージ内で宣言された `ErrorTemplate` 構造体を使用してエラーを再定義します。

```go
e := echo.New()
reflector := openapi31.Reflector{}
swagger := NewEchoSwagger(e, &reflector)

api := swagger.Group("/api")

api.GET("/users/:id", getUserHandler).
SetTags("users").
SetSummary("Get user by ID").
AddRequest(UserIDParam{}).
AddResponse(User{}).
AddErrors(factory.ErrUserNotFound)

swagger.RegisterOpenAPI()
```

### Request/Response 関連タグ

- description: リクエスト/レスポンス構造体の説明
- required: 必須フィールド
    - true/false
- example: 例示値
- format: データ形式
    - ex) int64, date-time
- enums: 列挙型の値
    - ex) enum:"A,B,C"
- json: JSONリクエストおよびレスポンス値
    - json payloadまたはレスポンス値で使用
- query: クエリパラメータ
- path: パスパラメータ
    - eecho は param タグで受け取るが、openapi-go は path で定義（この不一致をどう統一するか検討中）

**SearchInput 構造体の例**

```go
type SearchInput struct {
Sortby         string `query:"sortby" description:"並び替え基準"`
Order          string `query:"order" description:"並び順" enum:"asc,desc"`
SkipCount      int    `query:"skip_count" description:"スキップ件数" format:"gte=0"`
MaxResultCount int    `query:"max_result_count" description:"最大表示件数" format:"gte=0,lte=100"`
}
```

### 注意事項

openapi-go は不適切な構成の場合にエラーが発生します。現在のコードでもエラーが発生した場合は panic を投げる実装になっているため注意が必要です。

⚠️特にパスパラメータは、echo との不一致により param に加えて path を必ず明示しないと正しく動作しません。

## DB Migration手順

1. DEV の DB をダンプで取得し、ローカルの MySQL コンテナで起動します。
2. goose を利用して .sql ファイルを生成し、マイグレーションを実行します。
3. ローカル MySQL コンテナで変更内容を確認後、PR を作成します。

### ローカルMySQLコンテナの実行

[Confluenceドキュメント](https://bit.ly/43c0sQQ)を参照し、.cnf ファイルおよびローカル環境設定ファイルである config.local.yml が作成済みであることを前提とします。

以下の手順で DEV 環境の MySQL をダンプし、Docker で実行できます。

1. make mysql_dump
2. make mysql_run

### gooseの使い方

1. `goose create <migration_name> sql` でマイグレーションファイルを生成します。
2. 目的の修正内容を SQL で記述します。このとき

- `-- +goose StatementBegin` と `-- +goose StatementEnd` は削除してください。エラーになります。

2. `cd ./apps/admin-api && go run main.go migrate --env=local` でマイグレーションを実行します。

### PR 作成

ローカル MySQL コンテナで変更内容を確認後、PR を作成します。

## 環境変数管理

環境変数は AWS System Manager Parameter Store を使用して管理し、各アプリごとに 1 つのパラメータに YAML ファイルの内容をすべて格納して管理します。

環境変数の確認・修正・ローカルでのテストには、`paramstore.sh` スクリプトを使用します。

- AWS MFAの MFA 権限が必要です。詳しい使用方法は[Confluenceドキュメント](https://illuminarean.atlassian.net/wiki/x/BwCQP)を参照してください。

```bash
./paramstore.sh get dev
    # - パラメータストアの /esjp/backend/dev パラメータの値を読み取り
    # - ./config/config.dev.yml に保存します。（config ディレクトリ内の 5 つの yaml は .gitignore 済み）

./paramstore.sh get all
    # - パラメータストアの /esjp/backend/ をプレフィックスとする 5 つのパラメータの値を読み取り
    # - ./config/config.{env}.yml に保存します。local, dev, qa, stg, prod

./paramstore.sh set dev qa
    # - ./config/config.dev.yml と ./config/config.qa.yml を読み取り
    # - パラメータストアの /esjp/backend/dev と /esjp/backend/qa に保存します。
./paramstore.sh set qa stg local
    # - 上と同様ですが、qa, stg, local に対して実行します。
```

## ローカルのポート(MIR-1329)

ZOLOME・IKKAQ・mira-point・帳簿と同時に起動できるよう、ポートを割り振っている。
**このリポジトリのポートは今までと変わらない。**

| 用途 | ポート |
| --- | --- |
| ユーザー画面(energy-share-next) | 3000 |
| 管理画面(energy-shares-next-admin) | 3001 |
| user-api | 8080 |
| admin-api | 8081 |
| jobs-scheduler | 8082 |
| jobs-scheduler の cron 状態(/ping・/entries) | 8083 |
| DB 本体(energysharesdb) | 3306 |
| DB マイナンバー(mynumberdb) | 3307 |
| localstack | 4566 |

全体のポート表は `miralita-inc/github-workflows` の `docs/local-ports.md` を参照。

### worktree(wt2-\*)を使うとき

`docker-compose.yaml` のポートは環境変数から読む(`${ESJ_DB_PORT:-3306}` など)。
worktree のディレクトリに `.env` を置き、本体のポート番号に一律 +20000 した値
(`ESJ_DB_PORT=23306`・`ESJ_MYNUMBER_DB_PORT=23307`・`ESJ_LOCALSTACK_PORT=24566`)
を指定する。合わせて `config/config.local.yml` の DB 接続先・localstack の URL
(`cdn`・`base_endpoint`・`invest_worker_queue_url`)も同じ番号に変える。書き換え
たら API を再起動し、localstack は作り直す(CORS の再設定のため)。

`jobs-scheduler` の cron 状態確認サーバー(`:8083`)も `config/config.local.yml`
の `jobSchedulerConfig.cronHttpPort` で変更できる(worktree では `28083` を想定)。

固定コンテナ名(`container_name`)は使っていないため、`Makefile` から DB に
接続するコマンド(`make db-setup`・`make mysql-client` など)は
`docker compose exec` 経由で動く。worktree から実行しても、そのディレクトリの
コンテナにだけ接続する。
