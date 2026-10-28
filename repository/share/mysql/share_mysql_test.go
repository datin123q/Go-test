package mysql

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/illuminarean/energy-shares-backend-repo/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type lockDBFactoryMock struct{}

type sqliteForUpdateDB struct {
	xorm.Interface
}

func (d sqliteForUpdateDB) SQL(sqlOrArgs interface{}, args ...interface{}) *xorm.Session {
	if query, ok := sqlOrArgs.(string); ok {
		sqlOrArgs = strings.TrimSuffix(query, " FOR UPDATE")
	}
	return d.Interface.SQL(sqlOrArgs, args...)
}

func (lockDBFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	db, err := (dbFactoryMock{}).EnergySharesDB(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = db.Insert(mockDeletedShare); err != nil {
		return nil, err
	}
	if _, err = db.ID(mockDeletedShare.Id).Delete(&domain.Share{}); err != nil {
		return nil, err
	}
	return sqliteForUpdateDB{Interface: db}, nil
}

func (lockDBFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	return (dbFactoryMock{}).MyNumberDB(ctx)
}

var mockDeletedShare = domain.Share{
	Id:           "SHARE_ID4",
	OwnerId:      1003,
	ProjectId:    2000,
	InvestStatus: domain.REQUEST_INVEST,
}

func Test_shareMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.Share
		wantErr bool
	}{
		{
			name: "특정 지분 ID로 조회합니다. => 지분 ID가 1000인 공유가 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  "SHARE_ID1",
			},
			want:    mockShare1,
			wantErr: false,
		},
		{
			name: "조회할 수 없는 지분 ID로 조회합니다. => 에러가 발생하기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  "SHARE_ID4",
			},
			want:    domain.Share{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &shareMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := r.GetById(tt.args.ctx, tt.args.id)

			assert.Equalf(t, tt.wantErr, err != nil, "shareMysqlRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(got, tt.want, ignoreTime), true, "shareMysqlRepository.GetById() = %v, want %v", got, tt.want)
		})
	}
}

type failingDBFactoryMock struct{}

func (failingDBFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	return nil, errors.New("db unavailable")
}

func (failingDBFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	return nil, errors.New("db unavailable")
}

// 呼び出し元のhttp層はerrors.Is(err, ErrItemNotFound)で404へ振り分ける。
// 行が無いときにこれを返さないと、振り分けに当たらず500になる(MIR-1300)。
func Test_shareMysqlRepository_GetById_NotFound(t *testing.T) {
	tests := []struct {
		name         string
		dbFactory    factory.DBFactory
		id           string
		wantNotFound bool
	}{
		{
			name:         "存在しないShareはNotFoundを返す",
			dbFactory:    dbFactoryMock{},
			id:           "SHARE_ID_NOT_FOUND",
			wantNotFound: true,
		},
		{
			name:         "論理削除済みのShareはNotFoundを返す",
			dbFactory:    lockDBFactoryMock{},
			id:           mockDeletedShare.Id,
			wantNotFound: true,
		},
		{
			name:         "DBを取得できないときはNotFoundにしない(500のまま)",
			dbFactory:    failingDBFactoryMock{},
			id:           mockShare1.Id,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := shareMysqlRepository{dbFactory: tt.dbFactory}

			got, err := repository.GetById(context.Background(), tt.id)

			require.Error(t, err)
			assert.Equal(t, domain.Share{}, got)
			assert.Equal(t, tt.wantNotFound, errors.Is(err, cdomain.ErrItemNotFound))
		})
	}
}

func Test_shareMysqlRepository_LockById(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		want      domain.Share
		wantErrIs error
	}{
		{
			name: "Shareを排他取得できる",
			id:   mockShare1.Id,
			want: mockShare1,
		},
		{
			name:      "存在しないShareはNotFoundを返す",
			id:        "SHARE_ID_NOT_FOUND",
			wantErrIs: cdomain.ErrItemNotFound,
		},
		{
			name:      "論理削除済みのShareはNotFoundを返す",
			id:        mockDeletedShare.Id,
			wantErrIs: cdomain.ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := shareMysqlRepository{dbFactory: lockDBFactoryMock{}}

			got, err := repository.LockById(context.Background(), tt.id)

			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Id, got.Id)
			assert.Equal(t, tt.want.OwnerId, got.OwnerId)
			assert.Equal(t, tt.want.ProjectId, got.ProjectId)
		})
	}
}

func Test_shareMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx   context.Context
		share domain.Share
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Share 모델에서 Status를 변경합니다 : INVEST_SUCCESS => 정상적으로 변경되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				share: domain.Share{
					Id:           "SHARE_ID1",
					InvestStatus: domain.SUCCESS_INVEST,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &shareMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			err := r.Update(tt.args.ctx, tt.args.share)
			assert.Equalf(t, tt.wantErr, err != nil, "shareMysqlRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_shareMysqlRepository_GetInvestorCount(t *testing.T) {
	type args struct {
		ctx       context.Context
		projectId int64
		options   []cdomain.ShareQueryOption
	}

	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "특정 프로젝트의 투자자 수를 전부 조회합니다. => 3명이 조회되기를 기대합니다.",
			args: args{
				ctx:       context.Background(),
				projectId: 2000,
				options:   []cdomain.ShareQueryOption{},
			},
			want:    3,
			wantErr: false,
		},
		{
			name: "특정 프로젝트의 투자성공(SUCCESS_INVEST)을 필터링합니다. => 1명이 조회되기를 기대합니다.",
			args: args{
				ctx:       context.Background(),
				projectId: 2000,
				options: []cdomain.ShareQueryOption{
					cdomain.ShareQueryOptionFunc().WithInvestStatuses(domain.SUCCESS_INVEST),
				},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "특정 프로젝트의 투자성공(SUCCESS_INVEST), 투자요청(REQUEST_INVEST)을 필터링합니다. => 2명이 조회되기를 기대합니다.",
			args: args{
				ctx:       context.Background(),
				projectId: 2000,
				options: []cdomain.ShareQueryOption{
					cdomain.ShareQueryOptionFunc().WithInvestStatuses(domain.SUCCESS_INVEST, domain.REQUEST_INVEST),
				},
			},
			want:    2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &shareMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := r.GetInvestorCount(tt.args.ctx, tt.args.projectId, tt.args.options...)
			assert.Equalf(t, tt.wantErr, err != nil, "GetInvestorCount() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetInvestorCount() = %v, want %v", got, tt.want)
		})
	}
}

func Test_shareMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx       context.Context
		projectId int64
		options   []cdomain.ShareQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.Share
		wantErr bool
	}{
		{
			name: "특정 프로젝트의 모든 지분을 조회합니다. => 3개의 지분이 조회되기를 기대합니다.",
			args: args{
				ctx:       context.Background(),
				projectId: 2000,
				options: []cdomain.ShareQueryOption{
					cdomain.ShareQueryOptionFunc().WithProjectId(2000),
					cdomain.ShareQueryOptionFunc().WithOffset(0),
					cdomain.ShareQueryOptionFunc().WithLimit(100),
				},
			},
			want:    []domain.Share{mockShare1, mockShare2, mockShare3},
			wantErr: false,
		},
		{
			name: "특정 프로젝트, 상태의 지분을 조회합니다 : 2000, [SUCCESS_INVEST, REQUEST_INVEST] => 2개의 지분이 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.ShareQueryOption{
					cdomain.ShareQueryOptionFunc().WithProjectId(2000),
					cdomain.ShareQueryOptionFunc().WithInvestStatuses(domain.SUCCESS_INVEST, domain.REQUEST_INVEST),
					cdomain.ShareQueryOptionFunc().WithOffset(0),
					cdomain.ShareQueryOptionFunc().WithLimit(100),
				},
			},
			want:    []domain.Share{mockShare1, mockShare2},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &shareMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := r.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equalf(t, tt.wantErr, err != nil, "shareMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(got, tt.want, ignoreTime), "shareMysqlRepository.GetAll() = %v, want %v", got, tt.want)
		})
	}
}

// mir1239DbFactoryMockはMIR-1239の検証専用。③の判定(OPERATION_INVESTと
// REFUND_INVESTのamount合計)で、論理削除された行が合計から除外されることを
// 実DB(sqlite)相当のフィルタで確認するため、独立したデータを用意する。
// ":memory:"のsqliteは接続(NewEngine呼び出し)のたびに別インスタンスになる
// ため、engineをキャッシュして毎回同じインスタンスを返す(自動レビューで
// 指摘。以前はGetSum内部が使うインスタンスと、テストが裏取りに使う
// インスタンスが別物になっていた)。
type mir1239DbFactoryMock struct {
	engine *xorm.Engine
}

func newMir1239DbFactoryMock() (*mir1239DbFactoryMock, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	if err := xormEngine.Sync(new(domain.Share)); err != nil {
		return nil, err
	}
	xormEngine.ShowSQL(false)

	if _, err := xormEngine.Insert(&domain.Share{
		Id:           "MIR1239_OPERATION",
		ProjectId:    3000,
		OwnerId:      4000,
		Amount:       1000,
		InvestStatus: domain.OPERATION_INVEST,
	}); err != nil {
		return nil, err
	}
	if _, err := xormEngine.Insert(&domain.Share{
		Id:           "MIR1239_REFUND",
		ProjectId:    3000,
		OwnerId:      4001,
		Amount:       500,
		InvestStatus: domain.REFUND_INVEST,
	}); err != nil {
		return nil, err
	}
	// 論理削除された出資。xormのInsertはDeletedフィールドを明示的にセット
	// しても無視してNULLで挿入するため、挿入後にDeleteを呼んで論理削除する
	// (MIR-1233で確認済みの挙動)。amountを大きくしておき、③の合計に紛れ込む
	// と検知できるようにする。
	if _, err := xormEngine.Insert(&domain.Share{
		Id:           "MIR1239_DELETED_OPERATION",
		ProjectId:    3000,
		OwnerId:      4002,
		Amount:       99999,
		InvestStatus: domain.OPERATION_INVEST,
	}); err != nil {
		return nil, err
	}
	if _, err := xormEngine.ID("MIR1239_DELETED_OPERATION").Delete(&domain.Share{}); err != nil {
		return nil, err
	}
	// 別プロジェクトの出資。project_idの絞り込みが効くことを確認するために
	// 混ぜておく。
	if _, err := xormEngine.Insert(&domain.Share{
		Id:           "MIR1239_OTHER_PROJECT",
		ProjectId:    3001,
		OwnerId:      4003,
		Amount:       88888,
		InvestStatus: domain.OPERATION_INVEST,
	}); err != nil {
		return nil, err
	}

	return &mir1239DbFactoryMock{engine: xormEngine}, nil
}

func (d *mir1239DbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	return d.engine, nil
}

func (d *mir1239DbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}

// MIR-1239: ③の判定で使うGetSum(OPERATION_INVESTとREFUND_INVESTの2ステータス、
// amount)が、論理削除された行を除外し、他プロジェクトの行も含めないことを
// 実DB(sqlite)相当のフィルタで確認する。
func TestShareMysqlRepository_GetSum_ExcludesDeletedForInvestmentExecutionCheck(t *testing.T) {
	mock, err := newMir1239DbFactoryMock()
	require.NoError(t, err)
	r := &shareMysqlRepository{
		dbFactory: mock,
	}

	got, err := r.GetSum(context.Background(),
		cdomain.ShareQueryOptionFunc().WithProjectId(3000),
		cdomain.ShareQueryOptionFunc().WithInvestStatuses(domain.OPERATION_INVEST, domain.REFUND_INVEST),
		cdomain.ShareQueryOptionFunc().WithSumColumn("amount"),
	)

	assert.NoError(t, err)
	// 1000(OPERATION_INVEST) + 500(REFUND_INVEST) = 1500。論理削除された
	// 99999と、別プロジェクトの88888は含まれない。
	assert.Equal(t, float64(1500), got)

	// MIR1239_DELETED_OPERATIONが物理削除ではなく論理削除であることを
	// Unscoped()で裏取りする(通常のGetでは見えないが、行自体は残っている)。
	db, err := mock.EnergySharesDB(context.Background())
	require.NoError(t, err)
	var deletedShare domain.Share
	has, err := db.ID("MIR1239_DELETED_OPERATION").Get(&deletedShare)
	assert.NoError(t, err)
	assert.False(t, has, "通常のGetでは論理削除済みの行は見えないはず")

	var unscopedShare domain.Share
	has, err = db.Unscoped().ID("MIR1239_DELETED_OPERATION").Get(&unscopedShare)
	assert.NoError(t, err)
	assert.True(t, has, "Unscoped()なら論理削除済みの行が見えるはず")
	assert.False(t, unscopedShare.Deleted.Time().IsZero(), "deletedカラムに削除日時が入っているはず")
}
