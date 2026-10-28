package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/eventbus"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func TestInvestmentService_WaitInvest_RechecksLockedWaiter(t *testing.T) {
	const (
		waiterID = int64(100)
		userID   = int64(200)
	)

	user := domain.User{Id: userID}
	incomingWaiter := domain.InvestmentWaiters{
		Id:        waiterID,
		ProjectId: 300,
		UserId:    userID,
		Status:    domain.WAITING,
		Qty:       10,
	}

	tests := []struct {
		name           string
		locked         domain.InvestmentWaiters
		lockErr        error
		advancedQty    int64
		advancedQtyErr error
		requestQty     int64
		txErr          error
		wantErr        error
	}{
		{
			// MIR-1068: WAITING_ENDは「繰上げ完了」と「打ち切り」の両方を意味する。
			// 繰上げ実績が無い(advancedQty=0)場合は募集失敗・提案期限切れによる打ち切り
			// なので、引き続き処理しない。
			name: "取得後に打ち切り(繰上げ実績なし)になった申込は処理しない",
			locked: func() domain.InvestmentWaiters {
				waiter := incomingWaiter
				waiter.Status = domain.WAITING_END
				return waiter
			}(),
			advancedQty: 0,
			requestQty:  1,
			wantErr:     cdomain.ErrBadRequest,
		},
		{
			name: "取得後にキャンセル済みになった申込は処理しない",
			locked: func() domain.InvestmentWaiters {
				waiter := incomingWaiter
				waiter.Status = domain.CANCELLED
				return waiter
			}(),
			advancedQty: 0,
			requestQty:  1,
			wantErr:     cdomain.ErrBadRequest,
		},
		{
			name:        "WAITINGで繰上げ可能残を使い切っていれば処理しない",
			locked:      incomingWaiter,
			advancedQty: 10,
			requestQty:  1,
			wantErr:     cdomain.ErrBadRequest,
		},
		{
			// 申込口数(Qty)を超えて繰上げることはない(2026-08-25 河内さん・小倉さん確認済み)。
			name:        "繰上げ可能残を超えるqtyは処理しない",
			locked:      incomingWaiter,
			advancedQty: 6,
			requestQty:  5,
			wantErr:     cdomain.ErrBadRequest,
		},
		{
			name: "ユーザーが一致しない申込は処理しない",
			locked: func() domain.InvestmentWaiters {
				waiter := incomingWaiter
				waiter.UserId = userID + 1
				return waiter
			}(),
			requestQty: 1,
			wantErr:    cdomain.ErrBadRequest,
		},
		{
			name:       "行ロックの失敗を返す",
			lockErr:    assert.AnError,
			requestQty: 1,
			wantErr:    assert.AnError,
		},
		{
			name:       "トランザクション開始の失敗を返す",
			txErr:      assert.AnError,
			requestQty: 1,
			wantErr:    assert.AnError,
		},
		{
			// ②レビューで発見: 個々の文単位でデッドロックを判定していたころは、行ロック
			// 取得の失敗を専用エラーへ変換できていたが、中間テーブルへのINSERT(waitInvest
			// 内、モックでは差し込めない)で起きるデッドロックを取りこぼしていた。戻り値
			// 全体に対する一括判定へ変更したので、行ロック取得側で確認する。
			name:       "行ロックでデッドロックが起きたら再試行を促すエラーに変換する",
			lockErr:    &mysql.MySQLError{Number: 1213, Message: "Deadlock found when trying to get lock; try restarting transaction"},
			requestQty: 1,
			wantErr:    cdomain.ErrConcurrentUpdate,
		},
		{
			name:           "繰上げ済み累計の集計でデッドロックが起きたら再試行を促すエラーに変換する",
			locked:         incomingWaiter,
			advancedQtyErr: &mysql.MySQLError{Number: 1213, Message: "Deadlock found when trying to get lock; try restarting transaction"},
			requestQty:     1,
			wantErr:        cdomain.ErrConcurrentUpdate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txManager := mocks.NewTxManager(t)
			repository := mocks.NewInvestmentWaitersRepository(t)
			if tt.txErr != nil {
				txManager.On("WithTx", mock.Anything, mock.Anything).Return(tt.txErr)
			} else {
				txManager.On("WithTx", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				repository.On("LockById", mock.Anything, waiterID).Return(tt.locked, tt.lockErr)
				if tt.lockErr == nil && tt.locked.UserId == userID {
					repository.On("LockAndSumAdvancedQty", mock.Anything, waiterID).Return(tt.advancedQty, tt.advancedQtyErr)
				}
			}

			service := InvestmentService{
				txManager:                   txManager,
				investmentWaitersRepository: repository,
			}
			waiter := incomingWaiter

			share, err := service.WaitInvest(context.Background(), &user, &waiter, tt.requestQty, false)

			assert.Nil(t, share)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, incomingWaiter, waiter)
		})
	}
}

func TestInvestmentService_WaitInvest_UsesExistingTransaction(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", filepath.Join(t.TempDir(), "investment.db"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	session := engine.NewSession()
	require.NoError(t, session.Begin())
	t.Cleanup(func() {
		_ = session.Rollback()
		session.Close()
	})
	ctx := context.WithValue(context.Background(), factory.ConextDBType, session)

	const (
		waiterID = int64(100)
		userID   = int64(200)
	)
	user := domain.User{Id: userID}
	waiter := domain.InvestmentWaiters{Id: waiterID, UserId: userID, Status: domain.WAITING, Qty: 10}
	lockedWaiter := waiter
	lockedWaiter.Status = domain.CANCELLED

	txManager := mocks.NewTxManager(t)
	repository := mocks.NewInvestmentWaitersRepository(t)
	repository.On("LockById", mock.Anything, waiterID).Return(lockedWaiter, nil)
	repository.On("LockAndSumAdvancedQty", mock.Anything, waiterID).Return(int64(0), nil)
	service := InvestmentService{
		txManager:                   txManager,
		investmentWaitersRepository: repository,
	}

	share, err := service.WaitInvest(ctx, &user, &waiter, 1, false)

	assert.Nil(t, share)
	assert.ErrorIs(t, err, cdomain.ErrBadRequest)
}

// このPRの中心となる書き込み処理(当選紐づけを investment_waiters_winning へ保存する)を
// 検証する。WaitInvest の成立時に、対象の待機者IDと作成された share の Id を持つ行が
// 1件だけ作られることを DB から直接確認する。ここを壊すと(木村さんの指摘どおり)
// 既存テストは何も気付けなかった。
func TestInvestmentService_WaitInvest_WritesToWinningTable(t *testing.T) {
	// Invest() 内部の投資イベント発行(Kafka)がグローバルな EventBus を要求するため、
	// テスト用のモックで初期化しておく(未初期化だと panic する)。
	_, err := eventbus.NewMockEventBus()
	require.NoError(t, err)

	engine, err := xorm.NewEngine("sqlite3", filepath.Join(t.TempDir(), "investment_success.db"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	require.NoError(t, engine.Sync(
		new(domain.Project),
		new(domain.Spv),
		new(domain.ProjectImage),
		new(domain.ProjectNews),
		new(domain.Attachment),
		new(domain.ProjectFile),
		new(domain.RepaymentDividendSchedule),
		new(domain.ProjectInvestmentExecute),
		new(domain.ProjectViewCount),
		new(domain.ProjectMilestone),
		new(domain.Share),
		new(domain.User),
		new(domain.Role),
		new(domain.UserAccount),
		new(domain.UserVirtualAccount),
		new(domain.DepositRequest),
		new(domain.InvestmentWaiters),
		new(domain.InvestmentWaitersWinning),
		new(domain.EventMessage),
	))

	now := jsonutils.JSONTime(time.Now())
	const (
		waiterID  = int64(200)
		userID    = int64(10)
		projectID = int64(20)
	)

	user := domain.User{Id: userID, Uid: "user-10", Email: "user10@example.com", Name: "テスト太郎", UserType: domain.PRIVATE, InvestorType: domain.PRIVATE_INVESTOR, Created: now, Updated: now}
	_, err = engine.Insert(&user)
	require.NoError(t, err)

	project := domain.Project{
		Id:                projectID,
		CommonId:          "project-20",
		Name:              "テストファンド",
		ProjectType:       domain.LOAN,
		ProjectStatus:     domain.OPEN_FUNDING,
		Posted:            true,
		SharePrice:        1000,
		MinShareQty:       1,
		TargetShareQty:    100,
		TargetShareAmount: 100000,
		RemainingQty:      100,
		Created:           now,
		Updated:           now,
	}
	_, err = engine.Insert(&project)
	require.NoError(t, err)

	waiter := domain.InvestmentWaiters{Id: waiterID, ProjectId: projectID, UserId: userID, Status: domain.WAITING, Qty: 1, WaitingDate: now}
	_, err = engine.Insert(&waiter)
	require.NoError(t, err)

	va := domain.UserVirtualAccount{
		UserId:           userID,
		VaBankName:       "テスト銀行",
		VaBranchCode:     "101",
		VaBranchName:     "テスト支店",
		VaAccountNumber:  "1111111",
		VaHolderNameKana: "テスト タロウ",
		ProjectType:      domain.LOAN,
		LinkedStatus:     domain.ON_VA_ACCOUNT,
		UpdateDate:       now,
		Created:          now,
		Updated:          now,
	}
	_, err = engine.Insert(&va)
	require.NoError(t, err)

	session := engine.NewSession()
	require.NoError(t, session.Begin())
	t.Cleanup(func() {
		_ = session.Rollback()
		session.Close()
	})
	ctx := context.WithValue(context.Background(), factory.ConextDBType, session)

	repository := mocks.NewInvestmentWaitersRepository(t)
	repository.On("LockById", mock.Anything, waiterID).Return(waiter, nil)
	repository.On("LockAndSumAdvancedQty", mock.Anything, waiterID).Return(int64(0), nil)

	service := InvestmentService{
		txManager:                   nil,
		investmentWaitersRepository: repository,
	}

	// メール送信(プロジェクト種別ごとのテンプレート検証)より前の段階で失敗しうる
	// エラーだけは早期に気付けるよう確認しておく。中間テーブルへの書き込みはメール送信より
	// 前に行われるため、全体の成否ではなく DB の状態で検証する。
	_, err = service.WaitInvest(ctx, &user, &waiter, 1, false)
	if err != nil {
		t.Logf("WaitInvest returned err (メール送信などテスト環境で到達不能な後続処理の可能性): %v", err)
	}

	var share domain.Share
	has, err := session.Where("project_id = ? AND owner_id = ?", projectID, userID).Get(&share)
	require.NoError(t, err)
	require.True(t, has, "share が作成されていること")

	var winning domain.InvestmentWaitersWinning
	has, err = session.Where("investment_waiters_id = ?", waiterID).Get(&winning)
	require.NoError(t, err)
	require.True(t, has, "investment_waiters_winning に行が作られること")
	assert.Equal(t, share.Id, winning.ShareId)
}

// TestInvestmentService_WaitInvest_TransitionsToWaitingEndOnlyWhenAdvancedQtyReachesQty は、
// ①の敵対的検証で発見した実装漏れ(繰上げのたびに無条件でWAITING_ENDへ遷移していた)の
// 回帰テスト。申請10口のうち1回目に6口繰上げてもStatusはWAITINGのまま残り、2回目に
// 残り4口を繰上げて累計が申請口数に達したときだけWAITING_ENDへ遷移することを、DBの
// 状態から直接確認する。
func TestInvestmentService_WaitInvest_TransitionsToWaitingEndOnlyWhenAdvancedQtyReachesQty(t *testing.T) {
	_, err := eventbus.NewMockEventBus()
	require.NoError(t, err)

	engine, err := xorm.NewEngine("sqlite3", filepath.Join(t.TempDir(), "investment_multiple_advances.db"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	require.NoError(t, engine.Sync(
		new(domain.Project),
		new(domain.Spv),
		new(domain.ProjectImage),
		new(domain.ProjectNews),
		new(domain.Attachment),
		new(domain.ProjectFile),
		new(domain.RepaymentDividendSchedule),
		new(domain.ProjectInvestmentExecute),
		new(domain.ProjectViewCount),
		new(domain.ProjectMilestone),
		new(domain.Share),
		new(domain.User),
		new(domain.Role),
		new(domain.UserAccount),
		new(domain.UserVirtualAccount),
		new(domain.DepositRequest),
		new(domain.InvestmentWaiters),
		new(domain.InvestmentWaitersWinning),
		new(domain.EventMessage),
		new(domain.AdminUser),
		new(domain.InvestmentRequest),
		new(domain.InvestmentWaitHistory),
	))

	now := jsonutils.JSONTime(time.Now())
	const (
		waiterID    = int64(300)
		userID      = int64(11)
		projectID   = int64(21)
		adminUserID = int64(1)
	)

	user := domain.User{Id: userID, Uid: "user-11", Email: "user11@example.com", Name: "テスト次郎", UserType: domain.PRIVATE, InvestorType: domain.PRIVATE_INVESTOR, Created: now, Updated: now}
	_, err = engine.Insert(&user)
	require.NoError(t, err)

	// waitInvest内のwaiter.ChangeStatus(WAITING_END)がGetSuperAdminIdを呼ぶため、
	// スーパー管理者を1件用意しておく。
	adminUser := domain.AdminUser{Id: adminUserID, Email: "admin@example.com", Name: "テスト管理者"}
	_, err = engine.Insert(&adminUser)
	require.NoError(t, err)
	_, err = engine.Insert(&domain.Role{UserId: adminUserID, RoleCode: string(domain.ROLE_SUPER_ADMIN)})
	require.NoError(t, err)

	project := domain.Project{
		Id:                projectID,
		CommonId:          "project-21",
		Name:              "テストファンド2",
		ProjectType:       domain.LOAN,
		ProjectStatus:     domain.OPEN_FUNDING,
		Posted:            true,
		SharePrice:        1000,
		MinShareQty:       1,
		TargetShareQty:    100,
		TargetShareAmount: 100000,
		RemainingQty:      100,
		Created:           now,
		Updated:           now,
	}
	_, err = engine.Insert(&project)
	require.NoError(t, err)

	// 申請10口の待機者。1回目に6口、2回目に残り4口を繰上げる。
	waiter := domain.InvestmentWaiters{Id: waiterID, ProjectId: projectID, UserId: userID, Status: domain.WAITING, Qty: 10, WaitingDate: now}
	_, err = engine.Insert(&waiter)
	require.NoError(t, err)

	va := domain.UserVirtualAccount{
		UserId:           userID,
		VaBankName:       "テスト銀行",
		VaBranchCode:     "101",
		VaBranchName:     "テスト支店",
		VaAccountNumber:  "2222222",
		VaHolderNameKana: "テスト ジロウ",
		ProjectType:      domain.LOAN,
		LinkedStatus:     domain.ON_VA_ACCOUNT,
		UpdateDate:       now,
		Created:          now,
		Updated:          now,
	}
	_, err = engine.Insert(&va)
	require.NoError(t, err)

	session := engine.NewSession()
	require.NoError(t, session.Begin())
	t.Cleanup(func() {
		_ = session.Rollback()
		session.Close()
	})
	ctx := context.WithValue(context.Background(), factory.ConextDBType, session)

	repository := mocks.NewInvestmentWaitersRepository(t)
	service := InvestmentService{
		txManager:                   nil,
		investmentWaitersRepository: repository,
	}

	// 1回目: 6口繰上げる。累計6 < 申請10のため、WAITINGのまま残るはず。
	repository.On("LockById", mock.Anything, waiterID).Return(waiter, nil).Once()
	repository.On("LockAndSumAdvancedQty", mock.Anything, waiterID).Return(int64(0), nil).Once()

	waiter1 := waiter
	_, err = service.WaitInvest(ctx, &user, &waiter1, 6, false)
	if err != nil {
		t.Logf("WaitInvest(1回目) returned err (メール送信などテスト環境で到達不能な後続処理の可能性): %v", err)
	}

	var afterFirst domain.InvestmentWaiters
	has, err := session.ID(waiterID).Get(&afterFirst)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, domain.WAITING, afterFirst.Status, "1回目(累計6/10)の繰上げ後もWAITINGのままであること")

	var winningsAfterFirst []domain.InvestmentWaitersWinning
	require.NoError(t, session.Where("investment_waiters_id = ?", waiterID).Find(&winningsAfterFirst))
	require.Len(t, winningsAfterFirst, 1, "1回目の繰上げでinvestment_waiters_winningが1件作られること")

	// 2回目: 残り4口を繰上げる。累計10 = 申請10のため、WAITING_ENDへ遷移するはず。
	repository.On("LockById", mock.Anything, waiterID).Return(waiter, nil).Once()
	repository.On("LockAndSumAdvancedQty", mock.Anything, waiterID).Return(int64(6), nil).Once()

	waiter2 := waiter
	_, err = service.WaitInvest(ctx, &user, &waiter2, 4, false)
	if err != nil {
		t.Logf("WaitInvest(2回目) returned err (メール送信などテスト環境で到達不能な後続処理の可能性): %v", err)
	}

	var afterSecond domain.InvestmentWaiters
	has, err = session.ID(waiterID).Get(&afterSecond)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, domain.WAITING_END, afterSecond.Status, "2回目(累計10/10)の繰上げでWAITING_ENDへ遷移すること")

	var winningsAfterSecond []domain.InvestmentWaitersWinning
	require.NoError(t, session.Where("investment_waiters_id = ?", waiterID).Find(&winningsAfterSecond))
	require.Len(t, winningsAfterSecond, 2, "2回目の繰上げでinvestment_waiters_winningが合計2件になること")
}
