# テストチェックリスト

バックエンド(Go)のテストを書くときに見る観点。`energy-shares-backend-repo` と
`mirapro-backend` の両方で成り立つ前提だけを書いている。

## 前提

- ORM は **xorm**。gorm ではない。
- テスト用DBは sqlite(`github.com/mattn/go-sqlite3`)のインメモリ。
- アサーションは testify。モックは mockery。
- `go.mod` の `replace` でモジュールを差し替えている。
  import は `github.com/illuminarean/energy-share-core-api/...` と書くが、
  実体は `./apps/admin-api` 配下。`energyshares-user-api` は `./apps/user-api`。

## 1. 基本の形 — table-driven

`tests := []struct{...}` を回して `t.Run` でサブテストにする。

```go
func TestProject_CheckExistFundNumber(t *testing.T) {
	ctx := setupProjectDuplicateCheck(t)

	tests := []struct {
		name       string
		id         int64
		fundNumber *int64
		want       bool
	}{
		{
			name:       "新規登録で重複していれば拾う",
			id:         0,
			fundNumber: lo.ToPtr(int64(202600001)),
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (Project{}).CheckExistFundNumber(ctx, tt.id, tt.fundNumber)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
```

- [ ] ケース名は日本語で、何を固定したいのかがわかる文にする
- [ ] ポインタ引数は `lo.ToPtr`(`github.com/samber/lo`)で作る
- [ ] **`require` は前提の検証、`assert` は結果の検証。**
      `require` は即中断するので、nil のまま先へ進んで無関係な panic になるのを防ぐ

## 2. モック(mockery)

- [ ] **`.EXPECT()` 形式は使えない。** 設定が `with-expecter: false` なので
      `usecase.On("MethodName", mock.Anything, args).Return(...)` と書く
- [ ] モックはテストケースごとに `func() *mocks.Xxx` フィールドで組む。
      ケース間でインスタンスを共有しない
- [ ] `ctx` は `mock.Anything`。それ以外の引数はできるだけ具体値で固定する
      (`mock.Anything` で埋めると引数の取り違えを検出できない)
- [ ] **モックの置き場所が3系統ある。** import 元を間違えると
      「同名だが別物」を渡すことになる

  | 生成先 | 対象 | miralita | ZOLOME |
  |---|---|---|---|
  | `clean-arch-domain/mocks` | ルートの `.mockery.yaml` | 118件 | 147件 |
  | `apps/admin-api/domain/mocks` | 管理API のドメイン | 24件 | 32件 |
  | `apps/user-api/domain/mocks` | ユーザーAPI のドメイン | 15件 | 28件 |

  `.mockery.yaml` は両プロダクトとも3つ(ルート / admin-api / user-api)。
  生成し直す前にどの設定が効くか確認する。

- [ ] `AssertExpectations` は既存では少数派(miralita 225ファイル中6、
      ZOLOME 337ファイル中9)。呼ぶ場合は「呼ばれないこと」まで検証したい
      ときに限る

## 3. 実DBを使うテスト(sqlite)

```go
func setupXxx(t *testing.T) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = engine.Exec(`CREATE TABLE ...`)
	require.NoError(t, err)

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}
```

- [ ] セットアップは `t.Helper()` を付けた関数に切る
- [ ] `CREATE TABLE` は生SQL。本番スキーマ全部ではなく、テスト対象が触る列だけ
- [ ] **engine を渡す定数は `factory.ConextDBType`。**
      `Context` ではなく `Conext`(コードベース側の綴り。型名は `ContextDBType`
      なので紛らわしい)
- [ ] マイナンバー用の別DBを触る場合は `factory.ConextMyNumberDBType`

## 4. 論理削除(soft delete)

- [ ] 論理削除は `Deleted jsonutils.JSONTime` に `xorm:"deleted"` タグ。
      **カラム名は `deleted`。`deleted_at` ではない**
- [ ] 通常のクエリは xorm が自動で除外する。含めたいときは `Unscoped()`
- [ ] **DB のユニーク制約は `deleted` を見ない。** 論理削除済みの行もユニーク
      キーを占有し続ける。重複チェックで削除済みを除外すると、
      **画面では通るのに保存で落ちる**という食い違いが起きる
      (`domain/project_duplicate_check_test.go` に実例)
- [ ] テストデータに **論理削除済みの行を必ず1件混ぜる。**
      `CREATE TABLE` に `deleted DATETIME NULL` を含め、
      `INSERT` の1件に日時を入れる
- [ ] 件数・集計・一覧のソートで、削除済みが混ざらない/落ちないことを確認する

## 5. 境界値

- [ ] 0件 / 1件 / 複数件
- [ ] `nil` ポインタと空文字・ゼロ値の区別。`*int64` の `nil` と `0` は別物
- [ ] 上限と上限+1(口数、金額、ページサイズ)
- [ ] 日付は期間の境界。開始日当日 / 終了日当日 / 終了日+1日
- [ ] 金額は口数 × 単価の桁あふれと、手数料の端数処理
- [ ] ステータスが `NULL` のレコード(既存データに存在しうる)

## 6. 異常系

- [ ] 対象レコードが無い
- [ ] 想定外のステータス値が入っている
- [ ] 権限が不足している(`ROLE_*` の組み合わせ)
- [ ] 外部サービス(S3 / SNS / メール)がエラーを返す
- [ ] トランザクションがロールバックされる経路
- [ ] 同じ処理が二重に走った場合(重複登録、二重出金)

## 7. 実行と後始末

```bash
go test ./domain/ -run TestXxx -v   # 単体
make test                           # 全体
```

- [ ] **`make test` は `bin/coverage.out` を生成する。** `/bin` は `.gitignore`
      済みなので追跡はされないが、実行後に `git status` を確認する
- [ ] 失敗件数は変更前と比較し、自分の変更による増加か既存の問題かを切り分ける
