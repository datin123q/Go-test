package mysql

import (
	"context"
	"strings"
	"testing"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
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
	if _, err = db.Insert(mockDeletedInvestmentWaiters); err != nil {
		return nil, err
	}
	if _, err = db.ID(mockDeletedInvestmentWaiters.Id).Delete(&domain.InvestmentWaiters{}); err != nil {
		return nil, err
	}
	return sqliteForUpdateDB{Interface: db}, nil
}

func (lockDBFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	return (dbFactoryMock{}).MyNumberDB(ctx)
}

var mockDeletedInvestmentWaiters = domain.InvestmentWaiters{
	Id:        1003,
	UserId:    3000,
	ProjectId: 2000,
	Status:    domain.WAITING,
}

func Test_investmentWaitersMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx               context.Context
		investmentWaiters domain.InvestmentWaiters
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "InvestmentWaiters 모델에서 Status를 변경합니다 : DEPOSIT_REQUEST => 정상적으로 변경되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				investmentWaiters: domain.InvestmentWaiters{
					Id:     1000,
					Status: domain.DEPOSIT_REQUEST,
				},
			},
			wantErr: false,
		},
		{
			name: "조회할 수 없는 Id로 변경합니다 : 에러가 발생하기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				investmentWaiters: domain.InvestmentWaiters{
					Id:     2000,
					Status: domain.DEPOSIT_REQUEST,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := investmentWaitersMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			_, err := d.Update(tt.args.ctx, tt.args.investmentWaiters)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentWaitersMysqlRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_investmentWaitersMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "InvestmentWaitersをIDで取得するとProjectStatusとInvestStatusも取得される",
			args: args{
				ctx: context.Background(),
				id:  1001,
			},
			wantErr: false,
		},
		{
			name: "조회할 수 없는 Id로 조회합니다 : 에러가 발생하기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  2000,
			},
			wantErr:   true,
			wantErrIs: cdomain.ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := investmentWaitersMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := d.GetById(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentWaitersMysqlRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
			}
			if !tt.wantErr {
				assert.Equal(t, mockInvestmentWaiters2.ProjectStatus, got.ProjectStatus)
				assert.Equal(t, mockInvestmentWaiters2.InvestStatus, got.InvestStatus)
			}
		})
	}
}

func Test_investmentWaitersMysqlRepository_LockById(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		want      domain.InvestmentWaiters
		wantErrIs error
	}{
		{
			name: "キャンセル待ちを排他取得できる",
			id:   mockInvestmentWaiters1.Id,
			want: mockInvestmentWaiters1,
		},
		{
			// mockInvestmentWaiters2 は中間テーブル(investment_waiters_winning)経由で
			// 当選済みだが、LockByIdはもう当選済みかどうかの解決を行わない(MIR-1118)。
			// investment_waiters.share_id の実カラムをそのまま返すだけであり、この待機者は
			// 実カラムへの直接書き込みが無いため空文字になる。当選済みかどうかの解決は
			// LockAndSumAdvancedQtyの責務(TestInvestmentWaitersMysqlRepository_LockAndSumAdvancedQty参照)。
			name: "中間テーブル経由で当選済みでもLockByIdは実カラムの値(空文字)をそのまま返す",
			id:   mockInvestmentWaiters2.Id,
			want: func() domain.InvestmentWaiters {
				w := mockInvestmentWaiters2
				w.ShareId = ""
				return w
			}(),
		},
		{
			name:      "存在しないキャンセル待ちはNotFoundを返す",
			id:        2000,
			wantErrIs: cdomain.ErrItemNotFound,
		},
		{
			name:      "論理削除済みのキャンセル待ちはNotFoundを返す",
			id:        mockDeletedInvestmentWaiters.Id,
			wantErrIs: cdomain.ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := investmentWaitersMysqlRepository{dbFactory: lockDBFactoryMock{}}

			got, err := repository.LockById(context.Background(), tt.id)

			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Id, got.Id)
			assert.Equal(t, tt.want.UserId, got.UserId)
			assert.Equal(t, tt.want.ProjectId, got.ProjectId)
			assert.Equal(t, tt.want.ShareId, got.ShareId)
		})
	}
}

// LockByIdは`investment_waiters.*`で実カラムをそのまま返すだけなので、xormが
// share_idの生値を正しくShareIdへマッピングできることを直接確認する(MIR-1118)。
// この値はLockAndSumAdvancedQtyの当選判定にはもう使われない(PR3で削除)が、LockByIdの
// 戻り値としては引き続き埋まる。呼び出し元がこの生値を当選判定に誤用しないよう、
// WaitInvest・Cancel側は必ずLockAndSumAdvancedQtyの戻り値を使うこと。
// ShareIdは xorm:"<-" のため通常のInsertでは書き込めず、生SQLで直接挿入する。
func Test_investmentWaitersMysqlRepository_LockById_ReadsLegacyShareIdColumn(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	require.NoError(t, engine.Sync(new(domain.InvestmentWaiters)))

	const waiterId = int64(9200)
	_, err = engine.Insert(&domain.InvestmentWaiters{Id: waiterId, UserId: 8200, ProjectId: 7200, Status: domain.WAITING})
	require.NoError(t, err)
	_, err = engine.Exec("UPDATE investment_waiters SET share_id = ? WHERE id = ?", "SHARE_ID_LEGACY_PHYSICAL", waiterId)
	require.NoError(t, err)

	// LockByIdは生SQLに"FOR UPDATE"を含むが、sqliteはこの構文を持たないため、
	// lockDBFactoryMockと同じくsqliteForUpdateDBで剥がしてから実行する。
	repository := investmentWaitersMysqlRepository{
		dbFactory: sqliteForUpdateEngineDbFactoryMock{engine: engine},
	}

	got, err := repository.LockById(context.Background(), waiterId)
	require.NoError(t, err)
	assert.Equal(t, "SHARE_ID_LEGACY_PHYSICAL", got.ShareId)
}

// LockAndSumAdvancedQty は Cancel API・WaitInvestの二重処理防止ガード・繰上げ可能残の
// 計算が使う唯一の情報源なので、LockById(実カラムをそのまま返すだけ)と混同していないかを
// 直接確認する(MIR-1118のCurrentShareIdをMIR-1068で複数当選対応に拡張)。旧列
// investment_waiters.share_idへのフォールバックはPR3(MIR-1163)で削除したため、中間
// テーブルに行が無い場合は常に0を返す。
func TestInvestmentWaitersMysqlRepository_LockAndSumAdvancedQty(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, engine *xorm.Engine, waiterId int64)
		wantSum int64
	}{
		{
			name: "中間テーブルに未削除の当選行がある場合はそのshare.qtyを返す",
			setup: func(t *testing.T, engine *xorm.Engine, waiterId int64) {
				_, err := engine.Insert(&domain.Share{Id: "SHARE_1", Qty: 3})
				require.NoError(t, err)
				_, err = engine.Insert(&domain.InvestmentWaitersWinning{
					InvestmentWaitersId: waiterId,
					ShareId:             "SHARE_1",
				})
				require.NoError(t, err)
			},
			wantSum: 3,
		},
		{
			name:    "中間テーブルに行が無い場合は0を返す",
			setup:   func(t *testing.T, engine *xorm.Engine, waiterId int64) {},
			wantSum: 0,
		},
		{
			name: "中間テーブルに論理削除済みの当選行しか無い場合は0を返す",
			setup: func(t *testing.T, engine *xorm.Engine, waiterId int64) {
				_, err := engine.Insert(&domain.Share{Id: "SHARE_CANCELLED", Qty: 5})
				require.NoError(t, err)
				winning := domain.InvestmentWaitersWinning{InvestmentWaitersId: waiterId, ShareId: "SHARE_CANCELLED"}
				_, err = engine.Insert(&winning)
				require.NoError(t, err)
				_, err = engine.ID(winning.Id).Delete(&domain.InvestmentWaitersWinning{})
				require.NoError(t, err)
			},
			wantSum: 0,
		},
		{
			// MIR-1068: 複数回の繰上げで当選行が複数になっても、未削除分をすべて合計する。
			name: "中間テーブルに複数の未削除当選行がある場合は合計を返す",
			setup: func(t *testing.T, engine *xorm.Engine, waiterId int64) {
				_, err := engine.Insert(&domain.Share{Id: "SHARE_A", Qty: 3})
				require.NoError(t, err)
				_, err = engine.Insert(&domain.Share{Id: "SHARE_B", Qty: 4})
				require.NoError(t, err)
				_, err = engine.Insert(&domain.InvestmentWaitersWinning{InvestmentWaitersId: waiterId, ShareId: "SHARE_A"})
				require.NoError(t, err)
				_, err = engine.Insert(&domain.InvestmentWaitersWinning{InvestmentWaitersId: waiterId, ShareId: "SHARE_B"})
				require.NoError(t, err)
			},
			wantSum: 7,
		},
		{
			// 一部の当選が取消(論理削除)されていても、未削除分だけを合計する。
			name: "一部の当選行が論理削除されている場合は未削除分だけ合計する",
			setup: func(t *testing.T, engine *xorm.Engine, waiterId int64) {
				_, err := engine.Insert(&domain.Share{Id: "SHARE_KEEP", Qty: 3})
				require.NoError(t, err)
				_, err = engine.Insert(&domain.Share{Id: "SHARE_DROP", Qty: 4})
				require.NoError(t, err)
				_, err = engine.Insert(&domain.InvestmentWaitersWinning{InvestmentWaitersId: waiterId, ShareId: "SHARE_KEEP"})
				require.NoError(t, err)
				dropped := domain.InvestmentWaitersWinning{InvestmentWaitersId: waiterId, ShareId: "SHARE_DROP"}
				_, err = engine.Insert(&dropped)
				require.NoError(t, err)
				_, err = engine.ID(dropped.Id).Delete(&domain.InvestmentWaitersWinning{})
				require.NoError(t, err)
			},
			wantSum: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := xorm.NewEngine("sqlite3", ":memory:")
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, engine.Close()) })

			require.NoError(t, engine.Sync(
				new(domain.InvestmentWaiters),
				new(domain.InvestmentWaitersWinning),
				new(domain.Share),
			))

			const waiterId = int64(9100)
			_, err = engine.Insert(&domain.InvestmentWaiters{Id: waiterId, UserId: 8100, ProjectId: 7100, Status: domain.WAITING})
			require.NoError(t, err)

			tt.setup(t, engine, waiterId)

			repository := investmentWaitersMysqlRepository{
				dbFactory: sqliteEngineDbFactoryMock{engine: engine},
			}

			got, err := repository.LockAndSumAdvancedQty(context.Background(), waiterId)
			require.NoError(t, err)
			assert.Equal(t, tt.wantSum, got)
		})
	}
}

// GetWinningsByWaiterId は管理画面の繰上げ履歴表示(MIR-1068)が使う唯一の情報源。
// GetAll/GetByIdと違い、この待機者に紐づく複数当選の各行を意図的に複数行で返す
// クエリなので、複数当選が水増しではなく正しく複数行返ること、論理削除済みの当選は
// 含まれないこと、deposit_requestが無い当選でも(LEFT JOINで)行が欠けないことを確認する。
func TestInvestmentWaitersMysqlRepository_GetWinningsByWaiterId(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	require.NoError(t, engine.Sync(
		new(domain.InvestmentWaiters),
		new(domain.InvestmentWaitersWinning),
		new(domain.Share),
		new(domain.DepositRequest),
	))

	const waiterId = int64(9300)
	_, err = engine.Insert(&domain.InvestmentWaiters{Id: waiterId, UserId: 8300, ProjectId: 7300, Status: domain.WAITING})
	require.NoError(t, err)

	// 1回目の繰上げ(4口、入金確認済み)。
	_, err = engine.Insert(&domain.Share{Id: "SHARE_A", Qty: 4, Price: 1000, Amount: 4000, InvestStatus: domain.SUCCESS_INVEST})
	require.NoError(t, err)
	_, err = engine.Insert(&domain.InvestmentWaitersWinning{InvestmentWaitersId: waiterId, ShareId: "SHARE_A"})
	require.NoError(t, err)
	_, err = engine.Insert(&domain.DepositRequest{ShareId: "SHARE_A", DepositCheckStatus: domain.OK_DEPOSIT})
	require.NoError(t, err)

	// 2回目の繰上げ(6口、入金依頼中)。deposit_requestは未作成でも行が欠けないこと。
	_, err = engine.Insert(&domain.Share{Id: "SHARE_B", Qty: 6, Price: 1000, Amount: 6000, InvestStatus: domain.REQUEST_INVEST})
	require.NoError(t, err)
	_, err = engine.Insert(&domain.InvestmentWaitersWinning{InvestmentWaitersId: waiterId, ShareId: "SHARE_B"})
	require.NoError(t, err)

	// investment_waiters_winning行自体が論理削除されている場合は含まれないこと。
	// (取消時に実際にこの行が論理削除される経路はコード上確認できていない。ここでは
	// NotDeletedConditionが効いていることをクエリ単体で検証する)。
	_, err = engine.Insert(&domain.Share{Id: "SHARE_WINNING_DELETED", Qty: 99, Price: 1000, Amount: 99000, InvestStatus: domain.CANCEL_INVEST})
	require.NoError(t, err)
	deletedWinning := domain.InvestmentWaitersWinning{InvestmentWaitersId: waiterId, ShareId: "SHARE_WINNING_DELETED"}
	_, err = engine.Insert(&deletedWinning)
	require.NoError(t, err)
	_, err = engine.ID(deletedWinning.Id).Delete(&domain.InvestmentWaitersWinning{})
	require.NoError(t, err)

	// shareが論理削除されている場合も含まれないこと(INNER JOINの未削除条件)。
	_, err = engine.Insert(&domain.Share{Id: "SHARE_DELETED", Qty: 98, Price: 1000, Amount: 98000, InvestStatus: domain.SUCCESS_INVEST})
	require.NoError(t, err)
	_, err = engine.ID("SHARE_DELETED").Delete(&domain.Share{})
	require.NoError(t, err)
	_, err = engine.Insert(&domain.InvestmentWaitersWinning{InvestmentWaitersId: waiterId, ShareId: "SHARE_DELETED"})
	require.NoError(t, err)

	// deposit_requestが論理削除されている場合はdeposit_check_statusが空であること
	// (LEFT JOINの未削除条件、②の敵対的検証で追加)。
	_, err = engine.Insert(&domain.Share{Id: "SHARE_C", Qty: 2, Price: 1000, Amount: 2000, InvestStatus: domain.SUCCESS_INVEST})
	require.NoError(t, err)
	_, err = engine.Insert(&domain.InvestmentWaitersWinning{InvestmentWaitersId: waiterId, ShareId: "SHARE_C"})
	require.NoError(t, err)
	deletedDeposit := domain.DepositRequest{ShareId: "SHARE_C", DepositCheckStatus: domain.OK_DEPOSIT}
	_, err = engine.Insert(&deletedDeposit)
	require.NoError(t, err)
	_, err = engine.ID(deletedDeposit.Id).Delete(&domain.DepositRequest{})
	require.NoError(t, err)

	repository := investmentWaitersMysqlRepository{dbFactory: sqliteEngineDbFactoryMock{engine: engine}}

	items, err := repository.GetWinningsByWaiterId(context.Background(), waiterId)
	require.NoError(t, err)
	require.Len(t, items, 3, "論理削除済みの当選・shareは含まれず、3件になること")

	// 並び順(investment_waiters_winning.id降順、繰上げを実行した順の新しい順)を確認する。
	// 最後に挿入したSHARE_Cが先頭に来るはず。
	assert.Equal(t, "SHARE_C", items[0].ShareId, "並び順が実装(id降順)どおりであること")

	byShareId := make(map[string]domain.InvestmentWaitersWinningItem, len(items))
	for _, item := range items {
		byShareId[item.ShareId] = item
	}

	itemA := byShareId["SHARE_A"]
	assert.Equal(t, int64(4), itemA.Qty)
	assert.Equal(t, int64(1000), itemA.Price)
	assert.Equal(t, int64(4000), itemA.Amount)
	assert.Equal(t, domain.SUCCESS_INVEST, itemA.InvestStatus)
	assert.Equal(t, domain.OK_DEPOSIT, itemA.DepositCheckStatus)

	itemB := byShareId["SHARE_B"]
	assert.Equal(t, int64(6), itemB.Qty)
	assert.Equal(t, domain.REQUEST_INVEST, itemB.InvestStatus)
	assert.Equal(t, domain.DEPOSIT_CHECK_STATUS(""), itemB.DepositCheckStatus, "deposit_requestが無い当選はdeposit_check_statusが空であること")

	itemC := byShareId["SHARE_C"]
	assert.Equal(t, domain.DEPOSIT_CHECK_STATUS(""), itemC.DepositCheckStatus, "deposit_requestが論理削除済みの当選はdeposit_check_statusが空であること")
}

// sqliteEngineDbFactoryMock は既存の engine をそのまま返すだけの DBFactory(MIR-1118)。
type sqliteEngineDbFactoryMock struct{ engine *xorm.Engine }

func (m sqliteEngineDbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	return m.engine, nil
}

func (m sqliteEngineDbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	return m.engine, nil
}

// sqliteForUpdateEngineDbFactoryMock は、LockByIdのように生SQLに"FOR UPDATE"を含む
// クエリをsqliteで実行するための DBFactory(MIR-1118)。lockDBFactoryMockと同じく
// sqliteForUpdateDBで" FOR UPDATE"を剥がす。
type sqliteForUpdateEngineDbFactoryMock struct{ engine *xorm.Engine }

func (m sqliteForUpdateEngineDbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	return sqliteForUpdateDB{Interface: m.engine}, nil
}

func (m sqliteForUpdateEngineDbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	return m.engine, nil
}

func Test_investmentWaitersMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.InvestmentWaitersQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.InvestmentWaiters
		wantErr bool
	}{
		{
			name: "InvestmentWaiters 모델 전체를 조회합니다 : 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentWaitersQueryOption{
					cdomain.InvestmentWaitersQueryOptionFunc().WithOffset(0),
					cdomain.InvestmentWaitersQueryOptionFunc().WithLimit(100),
				},
			},
			want: []domain.InvestmentWaiters{
				mockInvestmentWaiters1, mockInvestmentWaiters2, mockInvestmentWaiters3,
			},

			wantErr: false,
		},
		{
			name: "특정 프로젝트 ID로 조회합니다 : 2000 => 2건이 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentWaitersQueryOption{
					cdomain.InvestmentWaitersQueryOptionFunc().WithOffset(0),
					cdomain.InvestmentWaitersQueryOptionFunc().WithLimit(100),
					cdomain.InvestmentWaitersQueryOptionFunc().WithProjectId(2000),
				},
			},
			want: []domain.InvestmentWaiters{
				mockInvestmentWaiters1, mockInvestmentWaiters2,
			},

			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := investmentWaitersMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			_, got, err := d.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "investmentWaitersMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "investmentWaitersMysqlRepository.GetAll() = %v, want %v", got, tt.want)
		})
	}
}

// GetAll/GetByIdは investment_waiters_winning・share への実JOINをやめ、スカラーサブクエリ
// (InvestmentWaitersInvestStatusExpr・AdvancedQtyExpr)でinvest_status・advanced_qtyを
// 解決するようにした(MIR-1068)。1申込に複数当選があっても一覧の行が水増しされないこと、
// advanced_qtyが複数当選の合計になること、invest_statusが最新(中間テーブルのid降順で先頭)の
// 当選のものになることを確認する。
func Test_investmentWaitersMysqlRepository_GetAll_GetById_MultipleWinningsDoNotDuplicateRows(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	require.NoError(t, engine.Sync(
		new(domain.InvestmentWaiters),
		new(domain.InvestmentWaitersWinning),
		new(domain.User),
		new(domain.Project),
		new(domain.Share),
	))

	openFunding := domain.OPEN_FUNDING
	cancelInvest := domain.CANCEL_INVEST
	successInvest := domain.SUCCESS_INVEST

	// Qty(申請口数)とAdvancedQty(繰上げ済み累計)を別の値にしておく。両方10だと、
	// AdvancedQtyExprの取り違え(たとえばQtyをそのまま返す実装ミス)があっても
	// テストが偶然一致して見逃す(②の敵対的検証で指摘)。
	waiter := domain.InvestmentWaiters{Id: 9001, UserId: 8001, ProjectId: 7001, Qty: 12, Status: domain.WAITING}
	_, err = engine.Insert(&waiter)
	require.NoError(t, err)
	_, err = engine.Insert(&domain.User{Id: 8001})
	require.NoError(t, err)
	_, err = engine.Insert(&domain.Project{Id: 7001, ProjectStatus: openFunding})
	require.NoError(t, err)

	// 1回目の繰上げ(4口、入金期限切れでキャンセル)、2回目の繰上げ(6口、投資成立)。
	_, err = engine.Insert(&domain.Share{Id: "SHARE_A", Qty: 4, InvestStatus: cancelInvest})
	require.NoError(t, err)
	_, err = engine.Insert(&domain.Share{Id: "SHARE_B", Qty: 6, InvestStatus: successInvest})
	require.NoError(t, err)
	_, err = engine.Insert(&domain.InvestmentWaitersWinning{InvestmentWaitersId: waiter.Id, ShareId: "SHARE_A"})
	require.NoError(t, err)
	_, err = engine.Insert(&domain.InvestmentWaitersWinning{InvestmentWaitersId: waiter.Id, ShareId: "SHARE_B"})
	require.NoError(t, err)

	repository := investmentWaitersMysqlRepository{dbFactory: sqliteEngineDbFactoryMock{engine: engine}}

	totalCount, all, err := repository.GetAll(context.Background(), cdomain.InvestmentWaitersQueryOptionFunc().WithLimit(100))
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalCount, "複数当選でもtotalCountは水増しされないこと")
	require.Len(t, all, 1, "複数当選でも一覧の行は水増しされないこと")
	assert.Equal(t, int64(10), all[0].AdvancedQty, "繰上げ済み累計は複数当選の合計になること")
	require.NotNil(t, all[0].InvestStatus)
	assert.Equal(t, successInvest, *all[0].InvestStatus, "invest_statusは最新(id降順で先頭)の当選のものになること")

	got, err := repository.GetById(context.Background(), waiter.Id)
	require.NoError(t, err)
	assert.Equal(t, int64(10), got.AdvancedQty)
	require.NotNil(t, got.InvestStatus)
	assert.Equal(t, successInvest, *got.InvestStatus)
}
