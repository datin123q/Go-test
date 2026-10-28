package mysql

import (
	"context"
	"strings"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// sqliteNoWaitDB は"FOR UPDATE NOWAIT"を持たないsqliteでTryLockByIdのテストを
// 通すため、末尾の"FOR UPDATE NOWAIT"を取り除く(share_mysql_test.goの
// sqliteForUpdateDBと同じ考え方)。
type sqliteNoWaitDB struct {
	xorm.Interface
}

func (d sqliteNoWaitDB) SQL(sqlOrArgs interface{}, args ...interface{}) *xorm.Session {
	if query, ok := sqlOrArgs.(string); ok {
		sqlOrArgs = strings.TrimSuffix(query, " FOR UPDATE NOWAIT")
	}
	return d.Interface.SQL(sqlOrArgs, args...)
}

type tryLockDBFactoryMock struct{}

func (tryLockDBFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	xormEngine.Sync(new(domain.Project), new(domain.Share))
	xormEngine.ShowSQL(false)

	// SUCCESS_INVESTのshareを紐付け、raised_amountの相関サブクエリが実際に
	// 値を返すデータにする。project.*ではなくPROJECT_DETAIL_COLUMN_LISTを
	// 使っていないと、この値は常に0のまま返ってしまう(MIR-1204で検出)。
	project := domain.Project{
		Id:             3000,
		Name:           "TryLockByIdのテスト用プロジェクト",
		TargetShareQty: 100,
	}
	share := domain.Share{
		Id:           "SHARE_ID_TRYLOCK1",
		ProjectId:    3000,
		Amount:       50000,
		Qty:          10,
		InvestStatus: domain.SUCCESS_INVEST,
	}
	if _, err := xormEngine.Insert(project, share); err != nil {
		return nil, err
	}

	return sqliteNoWaitDB{Interface: xormEngine}, nil
}

func (tryLockDBFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}

func Test_projectMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.Project
		wantErr bool
	}{
		{
			name: "プロジェクト単件照会 : IDが1であるプロジェクト照会 => 1件が照会されることが期待されます。",
			args: args{
				ctx: context.Background(),
				id:  1000,
			},
			want: domain.Project{
				Id:                     1000,
				CommonId:               "0cb05069-19e1-4ea8-8cfd-a50388eec4c5",
				Name:                   "水資源プロジェクト",
				CategoryType:           domain.SOLAR,
				SelectionType:          domain.PROJECT_SELECTION_LOTTERY,
				OwnerId:                2000,
				TargetShareQty:         1000,
				IsWaitInvest:           false,
				TransactionCount:       1,
				RemainingQty:           1000,
				InvestWaitRemainingQty: 1000,
			},
			wantErr: false,
		},
		{
			name: "プロジェクト単件照会 : IDが2であるプロジェクト照会 => 照会されるプロジェクトがありません",
			args: args{
				ctx: context.Background(),
				id:  1001,
			},
			want:    domain.Project{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := projectMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := p.GetById(tt.args.ctx, tt.args.id)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "projectMysqlRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(got, tt.want, ignoreTime), "projectMysqlRepository.GetById() = %v, want %v", got, tt.want)
		})
	}
}

// MIR-1204の回帰テスト: TryLockByIdがGetByIdと同じPROJECT_DETAIL_COLUMN_LISTを
// 使っており、RaisedAmount(shareへの相関サブクエリで埋まる列)が正しく
// 埋まることを確認する。"SELECT project.*"に戻すと、この値は常に0になる。
func Test_projectMysqlRepository_TryLockById(t *testing.T) {
	tests := []struct {
		name       string
		id         int64
		wantRaised int64
		wantErr    bool
	}{
		{
			name:       "SUCCESS_INVESTのshareがあるプロジェクトはRaisedAmountが正しく埋まる",
			id:         3000,
			wantRaised: 50000,
			wantErr:    false,
		},
		{
			name:    "存在しないIDはErrItemNotFound",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := projectMysqlRepository{
				dbFactory: tryLockDBFactoryMock{},
			}
			got, err := p.TryLockById(context.Background(), tt.id)

			assert.Equalf(t, tt.wantErr, err != nil, "projectMysqlRepository.TryLockById() error = %v, wantErr %v", err, tt.wantErr)
			if !tt.wantErr {
				assert.Equal(t, tt.wantRaised, got.RaisedAmount, "RaisedAmount should be populated from the raised_amount subquery, not left at zero")
			}
		})
	}
}

func Test_projectMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx     context.Context
		project domain.Project
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "プロジェクト修正 : IDが1000であるプロジェクト修正 => 修正が成功する必要があります。",
			args: args{
				ctx: context.Background(),
				project: domain.Project{
					Id:            1000,
					ProjectStatus: domain.END_FUNDING,
					IsWaitInvest:  false,
					OfferingType:  domain.OFFERING_PUBLIC,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := projectMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := p.Update(tt.args.ctx, tt.args.project)
			assert.Equalf(t, tt.wantErr, err != nil, "projectMysqlRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

// singleEngineDBFactoryMockは、dbFactoryMockと違い呼び出しのたびに新しい
// :memory: DBを作らず、同じengineを返し続ける。Update→GetByIdのように
// 1テスト内で複数回dbFactoryを呼ぶ検証で、書き込みが後続の呼び出しから
// 見えることを確かめるために使う(MIR-1314)。
type singleEngineDBFactoryMock struct {
	engine *xorm.Engine
}

func (d *singleEngineDBFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	return d.engine, nil
}

func (d *singleEngineDBFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}

// Test_projectMysqlRepository_Update_DoesNotWriteProjectStatusは、
// project_statusがColsから外れており、渡した値に関わらず書き込まれないこと
// を確認する(MIR-1314)。呼び出し元(メンテナンス・share_svc)はどちらも
// GetByIdで読んだ値をそのまま渡しているだけで状態を変える意図は無いが、
// 書いたままにすると並行して動くジョブや状態変更APIによる変更をこの保存で
// 古い値に戻してしまう。mockProject1はProjectStatusを設定していない
// (ゼロ値の空文字列)ため、END_FUNDINGを渡しても変わらないことで検証する。
func Test_projectMysqlRepository_Update_DoesNotWriteProjectStatus(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync(new(domain.Project), new(domain.Share)))
	_, err = engine.Insert(mockProject1, mockShare1)
	require.NoError(t, err)

	p := projectMysqlRepository{
		dbFactory: &singleEngineDBFactoryMock{engine: engine},
	}

	project := mockProject1
	project.ProjectStatus = domain.END_FUNDING
	project.OfferingType = domain.OFFERING_PUBLIC

	err = p.Update(context.Background(), project)
	require.NoError(t, err)

	got, err := p.GetById(context.Background(), mockProject1.Id)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, mockProject1.ProjectStatus, got.ProjectStatus, "project_statusは書き換わらないこと")
	assert.Equal(t, domain.OFFERING_PUBLIC, got.OfferingType, "他の対象列(offering_type)は通常どおり書き込まれること")
}
