package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// persistentDBFactoryMock はdbFactoryMockと違い、呼び出しのたびに新しい
// :memory:sqliteを作らず、同じエンジンを使い回す。Create→Update→GetByProjectId
// のように複数回EnergySharesDB(ctx)を呼ぶテストで、書き込んだ内容が
// 次の呼び出しでも見える必要があるために使う。
type persistentDBFactoryMock struct {
	engine *xorm.Engine
}

func newPersistentDBFactoryMock() *persistentDBFactoryMock {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	engine.Sync(new(domain.Project), new(domain.ProjectInvestmentExecute))
	engine.ShowSQL(false)
	return &persistentDBFactoryMock{engine: engine}
}

func (d *persistentDBFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	return d.engine, nil
}

func (d *persistentDBFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}

func Test_projectInvestmentExecuteMysqlRepository_GetByProjectId(t *testing.T) {
	type args struct {
		ctx       context.Context
		projectId int64
	}

	tests := []struct {
		name    string
		args    args
		want    domain.ProjectInvestmentExecute
		wantErr bool
	}{
		{
			name: "프로젝트 ID로 조회 1000 => 1건이 조회되기를 기대합니다.",
			args: args{
				ctx:       context.Background(),
				projectId: 1000,
			},
			want: domain.ProjectInvestmentExecute{
				Id:            3000,
				ProjectId:     1000,
				ExecuteStatus: domain.COMPLETED,
				Amount:        10000,
			},
			wantErr: false,
		},
		{
			name: "프로젝트 ID로 조회 1001 => 조회되지 않기를 기대합니다.",
			args: args{
				ctx:       context.Background(),
				projectId: 1001,
			},
			want:    domain.ProjectInvestmentExecute{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := projectInvestmentExecuteMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetByProjectId(tt.args.ctx, tt.args.projectId)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{}, time.Time{})
			assert.Equalf(t, tt.wantErr, err != nil, "GetByProjectId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "GetByProjectId() = %v, want %v", got, tt.want)
		})
	}
}

// MIR-1204の回帰テスト: Createがdb.Insert(&x)でポインタを渡し、自動採番された
// Idを呼び出し元へ正しく書き戻すことを確認する。値渡し(db.Insert(x))だと
// Idが常に0のまま返り、直後のUpdate(ExecuteInvestmentが分割完了時に呼ぶ)が
// "WHERE id = 0"で対象0件のまま静かに失敗し、DB上はSCHEDULEDのまま残る
// (実際にCloudeレビューで検出された)。
func Test_projectInvestmentExecuteMysqlRepository_CreateThenUpdate(t *testing.T) {
	repo := projectInvestmentExecuteMysqlRepository{
		dbFactory: newPersistentDBFactoryMock(),
	}
	ctx := context.Background()

	created, err := repo.Create(ctx, domain.ProjectInvestmentExecute{
		ProjectId:     2000,
		ExecuteStatus: domain.SCHEDULED,
		Amount:        5000,
	})
	assert.NoError(t, err)
	assert.NotZerof(t, created.Id, "Create()が返すIdは自動採番された値であるべきで、0のままではいけない")

	created.ExecuteStatus = domain.COMPLETED
	err = repo.Update(ctx, created)
	assert.NoError(t, err)

	got, err := repo.GetByProjectId(ctx, 2000)
	assert.NoError(t, err)
	assert.Equalf(t, domain.COMPLETED, got.ExecuteStatus,
		"UpdateはCreateが返した本物のIdに対して行われ、DB上の値もCOMPLETEDへ変わっているべき")
}

// Test_projectInvestmentExecuteMysqlRepository_IncrementPublishFailedCount は、
// SQLレベルでの加算(UPDATE ... SET publish_failed_count = publish_failed_count + ?)
// が累積・0件更新・delta<=0のいずれでも正しく振る舞うことを確認する(MIR-1202)。
func Test_projectInvestmentExecuteMysqlRepository_IncrementPublishFailedCount(t *testing.T) {
	repo := projectInvestmentExecuteMysqlRepository{
		dbFactory: newPersistentDBFactoryMock(),
	}
	ctx := context.Background()

	created, err := repo.Create(ctx, domain.ProjectInvestmentExecute{
		ProjectId:     3000,
		ExecuteStatus: domain.SCHEDULED,
		Amount:        5000,
	})
	assert.NoError(t, err)

	t.Run("加算するたびに累積される", func(t *testing.T) {
		assert.NoError(t, repo.IncrementPublishFailedCount(ctx, created.Id, 2))
		got, err := repo.GetByProjectId(ctx, 3000)
		assert.NoError(t, err)
		assert.EqualValues(t, 2, got.PublishFailedCount)

		assert.NoError(t, repo.IncrementPublishFailedCount(ctx, created.Id, 3))
		got, err = repo.GetByProjectId(ctx, 3000)
		assert.NoError(t, err)
		assert.EqualValues(t, 5, got.PublishFailedCount, "累積であり上書きではないこと")
	})

	t.Run("delta<=0のときは何もしない", func(t *testing.T) {
		before, err := repo.GetByProjectId(ctx, 3000)
		assert.NoError(t, err)

		assert.NoError(t, repo.IncrementPublishFailedCount(ctx, created.Id, 0))
		assert.NoError(t, repo.IncrementPublishFailedCount(ctx, created.Id, -1))

		after, err := repo.GetByProjectId(ctx, 3000)
		assert.NoError(t, err)
		assert.Equal(t, before.PublishFailedCount, after.PublishFailedCount)
	})

	t.Run("存在しないidを指定してもエラーにならない(0件更新)", func(t *testing.T) {
		assert.NoError(t, repo.IncrementPublishFailedCount(ctx, 999999, 1))
	})
}
