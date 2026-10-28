package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func assertFactoryErrorCode(t *testing.T, err error, wantCode int) {
	t.Helper()

	require.Error(t, err)
	factoryErr, ok := err.(*factory.Error)
	require.True(t, ok, "error type = %T", err)
	assert.Equal(t, wantCode, factoryErr.Code)
}

func stubInvestmentWaitValidationDependencies(t *testing.T) {
	t.Helper()

	originalGetWaitingAmount := getInvestmentWaitersTotalAmount
	originalGetShareAmount := getShareAmountByUserID
	originalCheckUserOverAge := checkUserOverAge

	getInvestmentWaitersTotalAmount = func(context.Context, int64, int64) (int64, error) {
		return 0, nil
	}
	getShareAmountByUserID = func(context.Context, int64, int64) (int64, error) {
		return 0, nil
	}
	checkUserOverAge = func(context.Context, *domain.User) error {
		return nil
	}

	t.Cleanup(func() {
		getInvestmentWaitersTotalAmount = originalGetWaitingAmount
		getShareAmountByUserID = originalGetShareAmount
		checkUserOverAge = originalCheckUserOverAge
	})
}

func TestValidateInvestmentWaitProject(t *testing.T) {
	tests := []struct {
		name    string
		project *domain.Project
		wantErr bool
		code    int
	}{
		{
			name:    "(失敗) プロジェクトが nil の場合",
			project: nil,
			wantErr: true,
			code:    factory.ErrorParameter.Code,
		},
		{
			name: "(失敗) 募集中だがキャンセル待ち受付中ではない場合",
			project: &domain.Project{
				ProjectStatus: domain.OPEN_FUNDING,
				IsWaitInvest:  false,
			},
			wantErr: true,
			code:    factory.ErrorNotOpenFunding.Code,
		},
		{
			name: "(失敗) キャンセル待ち受付中でも投資受付可能ステータスではない場合",
			project: &domain.Project{
				ProjectStatus: domain.STANDBY_FUNDING,
				IsWaitInvest:  true,
			},
			wantErr: true,
			code:    factory.ErrorNotOpenFunding.Code,
		},
		{
			name: "(成功) 募集中かつキャンセル待ち受付中の場合",
			project: &domain.Project{
				ProjectStatus: domain.OPEN_FUNDING,
				IsWaitInvest:  true,
			},
			wantErr: false,
		},
		{
			name: "(成功) 募集終了かつキャンセル待ち受付中の場合",
			project: &domain.Project{
				ProjectStatus: domain.END_FUNDING,
				IsWaitInvest:  true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInvestmentWaitProject(tt.project)
			if tt.wantErr {
				assertFactoryErrorCode(t, err, tt.code)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateInvestmentWaitApplicationBeforeDBAccess(t *testing.T) {
	stubInvestmentWaitValidationDependencies(t)

	tests := []struct {
		name    string
		user    *domain.User
		project *domain.Project
		wantErr bool
		code    int
	}{
		{
			name:    "(失敗) ユーザーが nil の場合",
			user:    nil,
			project: &domain.Project{},
			wantErr: true,
			code:    factory.ErrorParameter.Code,
		},
		{
			name: "(失敗) 投資制限ユーザーの場合",
			user: &domain.User{
				Id:          1,
				BlockInvest: true,
			},
			project: &domain.Project{},
			wantErr: true,
			code:    factory.ErrorBlockInvest.Code,
		},
		{
			name: "(失敗) キャンセル待ち受付中ではない場合",
			user: &domain.User{
				Id: 1,
			},
			project: &domain.Project{
				Id:            5,
				ProjectStatus: domain.OPEN_FUNDING,
				IsWaitInvest:  false,
			},
			wantErr: true,
			code:    factory.ErrorNotOpenFunding.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInvestmentWaitApplication(context.Background(), tt.user, tt.project)
			if tt.wantErr {
				assertFactoryErrorCode(t, err, tt.code)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateInvestmentWaitApplication(t *testing.T) {
	baseUser := &domain.User{
		Id: 1,
	}
	baseProject := &domain.Project{
		Id:            5,
		ProjectStatus: domain.OPEN_FUNDING,
		IsWaitInvest:  true,
		SharePrice:    100,
	}

	tests := []struct {
		name    string
		setup   func()
		project *domain.Project
		wantErr bool
		code    int
	}{
		{
			name:    "(失敗) 年齢条件を満たさない場合",
			project: baseProject,
			setup: func() {
				checkUserOverAge = func(context.Context, *domain.User) error {
					return errors.New("over age")
				}
			},
			wantErr: true,
			code:    factory.ErrorOverAgeInvestor.Code,
		},
		{
			name:    "(成功) 事前チェックを通過する場合",
			project: baseProject,
			setup:   func() {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubInvestmentWaitValidationDependencies(t)
			tt.setup()

			err := validateInvestmentWaitApplication(context.Background(), baseUser, tt.project)
			if tt.wantErr {
				assertFactoryErrorCode(t, err, tt.code)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateInvestmentWaitAmountLimit(t *testing.T) {
	amountLimitProject := &domain.Project{
		Id:            5,
		ProjectStatus: domain.OPEN_FUNDING,
		IsWaitInvest:  true,
		IsAmountLimit: true,
		SharePrice:    100,
	}

	tests := []struct {
		name    string
		setup   func()
		project *domain.Project
		qty     int64
		wantErr bool
		code    int
	}{
		{
			name:    "(失敗) 既存投資金額により投資金額制限を超過する場合",
			project: amountLimitProject,
			qty:     1,
			setup: func() {
				getShareAmountByUserID = func(context.Context, int64, int64) (int64, error) {
					return LIMITED_AMOUNT - 50, nil
				}
			},
			wantErr: true,
			code:    factory.ErrorOverLimitedAmount.Code,
		},
		{
			name:    "(失敗) 既存のキャンセル待ち金額により投資金額制限を超過する場合",
			project: amountLimitProject,
			qty:     1,
			setup: func() {
				getInvestmentWaitersTotalAmount = func(context.Context, int64, int64) (int64, error) {
					return LIMITED_AMOUNT - 50, nil
				}
			},
			wantErr: true,
			code:    factory.ErrorOverLimitedAmount.Code,
		},
		{
			name:    "(失敗) 既存投資+既存キャンセル待ち+今回申込で投資金額制限を超過する場合",
			project: amountLimitProject,
			qty:     2,
			setup: func() {
				getShareAmountByUserID = func(context.Context, int64, int64) (int64, error) {
					return 2_000_000, nil
				}
				getInvestmentWaitersTotalAmount = func(context.Context, int64, int64) (int64, error) {
					return 2_999_900, nil
				}
			},
			wantErr: true,
			code:    factory.ErrorOverLimitedAmount.Code,
		},
		{
			name:    "(成功) 既存のキャンセル待ちがあっても合算が上限未満なら再申込できる",
			project: amountLimitProject,
			qty:     1,
			setup: func() {
				getShareAmountByUserID = func(context.Context, int64, int64) (int64, error) {
					return 1_000_000, nil
				}
				getInvestmentWaitersTotalAmount = func(context.Context, int64, int64) (int64, error) {
					return 2_000_000, nil
				}
			},
			wantErr: false,
		},
		{
			name: "(成功) 金額制限がない場合は常に通過する",
			project: &domain.Project{
				Id:            5,
				IsAmountLimit: false,
				SharePrice:    100,
			},
			qty: 1,
			setup: func() {
				getShareAmountByUserID = func(context.Context, int64, int64) (int64, error) {
					return LIMITED_AMOUNT, nil
				}
			},
			wantErr: false,
		},
		{
			name: "(成功) 合計金額が 4,999,999 円の場合は上限未満として許可する",
			project: &domain.Project{
				Id:            5,
				IsAmountLimit: true,
				SharePrice:    4_999_999,
			},
			qty:     1,
			setup:   func() {},
			wantErr: false,
		},
		{
			name: "(失敗) 合計金額が 5,000,000 円ちょうどの場合は上限到達として拒否する",
			project: &domain.Project{
				Id:            5,
				IsAmountLimit: true,
				SharePrice:    5_000_000,
			},
			qty:     1,
			setup:   func() {},
			wantErr: true,
			code:    factory.ErrorOverLimitedAmount.Code,
		},
		{
			name: "(失敗) 合計金額が 5,000,001 円の場合は上限超過として拒否する",
			project: &domain.Project{
				Id:            5,
				IsAmountLimit: true,
				SharePrice:    5_000_001,
			},
			qty:     1,
			setup:   func() {},
			wantErr: true,
			code:    factory.ErrorOverLimitedAmount.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubInvestmentWaitValidationDependencies(t)
			tt.setup()

			err := validateInvestmentWaitAmountLimit(context.Background(), 1, tt.project, tt.qty)
			if tt.wantErr {
				assertFactoryErrorCode(t, err, tt.code)
				return
			}
			require.NoError(t, err)
		})
	}
}

func setupInvestmentWaitApplicationDB(t *testing.T) (context.Context, *xorm.Engine) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	_, err = engine.Exec(`
CREATE TABLE user (
    id INTEGER PRIMARY KEY
);
CREATE TABLE share (
    id INTEGER PRIMARY KEY,
    project_id INTEGER,
    owner_id INTEGER,
    amount INTEGER,
    invest_status TEXT,
    deleted DATETIME
);
CREATE TABLE investment_waiters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER,
    user_id INTEGER,
    waiting_turn INTEGER,
    qty INTEGER,
    amount INTEGER,
    status TEXT,
    waiting_date DATETIME,
    share_id TEXT,
    created DATETIME,
    updated DATETIME,
    deleted DATETIME
);
INSERT INTO user (id) VALUES (1);
`)
	require.NoError(t, err)

	return context.WithValue(context.Background(), factory.ConextDBType, engine), engine
}

// setupInvestmentWaitApplicationTxDB は本番の POST ミドルウェアと同様に
// 既存 Session 上でトランザクションを開始した context を返す。
func setupInvestmentWaitApplicationTxDB(t *testing.T) (context.Context, *xorm.Session, *xorm.Engine) {
	t.Helper()

	ctx, engine := setupInvestmentWaitApplicationDB(t)
	session := engine.NewSession()
	t.Cleanup(func() {
		_ = session.Close()
	})
	require.NoError(t, session.Begin())

	txCtx := context.WithValue(ctx, factory.ConextDBType, session)
	return txCtx, session, engine
}

func stubInvestmentWaitNotifications(t *testing.T) {
	t.Helper()

	originalPublish := publishInvestmentWaitRequestEvent
	originalSendMail := sendInvestStandbyMail

	publishInvestmentWaitRequestEvent = func(context.Context, *domain.User, *domain.InvestmentWaiters, *domain.Project) error {
		return nil
	}
	sendInvestStandbyMail = func(context.Context, string, string, string, string, string, string, string) error {
		return nil
	}

	t.Cleanup(func() {
		publishInvestmentWaitRequestEvent = originalPublish
		sendInvestStandbyMail = originalSendMail
	})
}

func countInvestmentWaiters(t *testing.T, ctx context.Context, projectID, userID int64) int64 {
	t.Helper()

	count, err := factory.DB(ctx).
		Where("project_id = ? AND user_id = ? AND status = ?", projectID, userID, domain.WAITING).
		Count(&domain.InvestmentWaiters{})
	require.NoError(t, err)
	return count
}

func TestCreateInvestmentWaitApplication_AllowsMultipleApplications(t *testing.T) {
	ctx, _ := setupInvestmentWaitApplicationDB(t)
	user := &domain.User{Id: 1}
	project := &domain.Project{
		Id:            10,
		IsAmountLimit: true,
		SharePrice:    1_000_000,
	}

	// 1回目: 200万円
	first, err := createInvestmentWaitApplication(ctx, user, project, 2)
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.NotZero(t, first.Id)
	assert.Equal(t, int64(2_000_000), first.Amount)

	// 2回目: 同じユーザー・同じファンドへ再申込（合計 400万円 < 500万円）
	second, err := createInvestmentWaitApplication(ctx, user, project, 2)
	require.NoError(t, err)
	require.NotNil(t, second)
	assert.NotZero(t, second.Id)
	assert.Equal(t, int64(2_000_000), second.Amount)

	// 両方成功し、異なる ID のレコードが 2 件作成される
	assert.NotEqual(t, first.Id, second.Id)
	assert.Equal(t, int64(2), countInvestmentWaiters(t, ctx, project.Id, user.Id))

	// 2件目を含めた合計金額が上限判定に使われる
	waitingAmount, err := domain.InvestmentWaiters{}.GetTotalAmountForWaitingStatus(ctx, project.Id, user.Id)
	require.NoError(t, err)
	assert.Equal(t, int64(4_000_000), waitingAmount)

	// 3回目: 200万円追加で合計 600万円となり上限超過 → 新規レコードは作成されない
	_, err = createInvestmentWaitApplication(ctx, user, project, 2)
	assertFactoryErrorCode(t, err, factory.ErrorOverLimitedAmount.Code)
	assert.Equal(t, int64(2), countInvestmentWaiters(t, ctx, project.Id, user.Id))

	waitingAmount, err = domain.InvestmentWaiters{}.GetTotalAmountForWaitingStatus(ctx, project.Id, user.Id)
	require.NoError(t, err)
	assert.Equal(t, int64(4_000_000), waitingAmount)
}

func TestCreateInvestmentWaitApplication_UsesExistingTransaction(t *testing.T) {
	txCtx, session, engine := setupInvestmentWaitApplicationTxDB(t)
	user := &domain.User{Id: 1}
	project := &domain.Project{
		Id:            10,
		IsAmountLimit: true,
		SharePrice:    1_000_000,
	}

	created, err := createInvestmentWaitApplication(txCtx, user, project, 1)
	require.NoError(t, err)
	require.NotNil(t, created)

	// 外側トランザクション未コミットでも同一 Session からは見える
	assert.Equal(t, int64(1), countInvestmentWaiters(t, txCtx, project.Id, user.Id))

	// 外側トランザクションをロールバックすると investment_waiters も消える
	// （ネストした独立トランザクションではないことの検証）
	require.NoError(t, session.Rollback())

	committedCtx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	assert.Equal(t, int64(0), countInvestmentWaiters(t, committedCtx, project.Id, user.Id))
}

func TestApplyInvestmentWaitApplication_NotificationFailuresStillSucceed(t *testing.T) {
	stubInvestmentWaitNotifications(t)

	tests := []struct {
		name            string
		setupNotify     func()
		wantWaiterCount int64
	}{
		{
			name: "イベント通知失敗でも申込は成功し1件作成される",
			setupNotify: func() {
				publishInvestmentWaitRequestEvent = func(context.Context, *domain.User, *domain.InvestmentWaiters, *domain.Project) error {
					return errors.New("event publish failed")
				}
			},
			wantWaiterCount: 1,
		},
		{
			name: "メール送信失敗でも申込は成功し1件作成される",
			setupNotify: func() {
				sendInvestStandbyMail = func(context.Context, string, string, string, string, string, string, string) error {
					return errors.New("mail send failed")
				}
			},
			wantWaiterCount: 1,
		},
		{
			name: "イベント・メール両方失敗でも申込は成功し1件作成される",
			setupNotify: func() {
				publishInvestmentWaitRequestEvent = func(context.Context, *domain.User, *domain.InvestmentWaiters, *domain.Project) error {
					return errors.New("event publish failed")
				}
				sendInvestStandbyMail = func(context.Context, string, string, string, string, string, string, string) error {
					return errors.New("mail send failed")
				}
			},
			wantWaiterCount: 1,
		},
		{
			name:            "通知成功時も申込は成功し1件作成される",
			setupNotify:     func() {},
			wantWaiterCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubInvestmentWaitNotifications(t)
			tt.setupNotify()

			ctx, _ := setupInvestmentWaitApplicationDB(t)
			user := &domain.User{Id: 1, Email: "user@example.com", Name: "test"}
			project := &domain.Project{
				Id:            10,
				Name:          "test fund",
				IsAmountLimit: true,
				SharePrice:    1_000_000,
			}

			created, err := applyInvestmentWaitApplication(ctx, user, project, 1)
			require.NoError(t, err)
			require.NotNil(t, created)
			assert.Equal(t, tt.wantWaiterCount, countInvestmentWaiters(t, ctx, project.Id, user.Id))
		})
	}
}

func TestApplyInvestmentWaitApplication_AllowsMultipleApplications(t *testing.T) {
	stubInvestmentWaitNotifications(t)

	ctx, _ := setupInvestmentWaitApplicationDB(t)
	user := &domain.User{Id: 1, Email: "user@example.com", Name: "test"}
	project := &domain.Project{
		Id:            10,
		Name:          "test fund",
		IsAmountLimit: true,
		SharePrice:    1_000_000,
	}

	first, err := applyInvestmentWaitApplication(ctx, user, project, 2)
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := applyInvestmentWaitApplication(ctx, user, project, 2)
	require.NoError(t, err)
	require.NotNil(t, second)

	assert.NotEqual(t, first.Id, second.Id)
	assert.Equal(t, int64(2), countInvestmentWaiters(t, ctx, project.Id, user.Id))
}

// TestCreateInvestmentWaitApplication_ConcurrentRequestsDoNotExceedLimit は
// 同一ユーザーへの並行申込で金額上限を超えないことを確認する。
//
// 注意: このテストは User.LockById / FOR UPDATE の有効性を検証するものではない。
// SQLite は FOR UPDATE 非対応のため、MaxOpenConns(1) と各 goroutine の Session.Begin
// によりクリティカルセクション全体を直列化している（接続取得待ちによる代替）。
// LockById を削除しても成功しうる。本番 MySQL での行ロック検証は別途必要。
func TestCreateInvestmentWaitApplication_ConcurrentRequestsDoNotExceedLimit(t *testing.T) {
	_, engine := setupInvestmentWaitApplicationDB(t)
	engine.SetMaxOpenConns(1)
	_, err := engine.Exec(`
INSERT INTO investment_waiters (project_id, user_id, waiting_turn, qty, amount, status)
VALUES (10, 1, 1, 1, 4000000, 'WAITING');
`)
	require.NoError(t, err)

	user := &domain.User{Id: 1}
	project := &domain.Project{
		Id:            10,
		IsAmountLimit: true,
		SharePrice:    600000,
	}

	const goroutines = 2
	var (
		wg           sync.WaitGroup
		successCount atomic.Int64
		limitCount   atomic.Int64
	)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			// 本番の POST ミドルウェアと同様、リクエストごとに独立したトランザクションを開始する
			session := engine.NewSession()
			defer session.Close()
			if err := session.Begin(); err != nil {
				t.Errorf("begin transaction: %v", err)
				return
			}
			txCtx := context.WithValue(context.Background(), factory.ConextDBType, session)

			_, err := createInvestmentWaitApplication(txCtx, user, project, 1)
			if err == nil {
				if err := session.Commit(); err != nil {
					t.Errorf("commit transaction: %v", err)
					return
				}
				successCount.Add(1)
				return
			}
			_ = session.Rollback()

			factoryErr, ok := err.(*factory.Error)
			if ok && factoryErr.Code == factory.ErrorOverLimitedAmount.Code {
				limitCount.Add(1)
				return
			}
			t.Errorf("unexpected error: %v", err)
		}()
	}
	wg.Wait()

	assert.Equal(t, int64(1), successCount.Load(), "片方のみ成功すること")
	assert.Equal(t, int64(1), limitCount.Load(), "もう片方は金額上限エラーになること")

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	assert.Equal(t, int64(2), countInvestmentWaiters(t, ctx, project.Id, user.Id), "上限超過分のレコードは作成されないこと")

	waitingAmount, err := domain.InvestmentWaiters{}.GetTotalAmountForWaitingStatus(ctx, project.Id, user.Id)
	require.NoError(t, err)
	assert.Equal(t, int64(4_600_000), waitingAmount)
	assert.Less(t, waitingAmount, LIMITED_AMOUNT)
}

func stubInvestmentWaitHTTPDependencies(t *testing.T, user *domain.User, project *domain.Project) {
	t.Helper()
	stubInvestmentWaitValidationDependencies(t)
	stubInvestmentWaitNotifications(t)

	originalGetUser := getUserByIDForInvestmentWait
	originalGetProject := getProjectByIDForInvestmentWait

	getUserByIDForInvestmentWait = func(context.Context, int64) (*domain.User, error) {
		return user, nil
	}
	getProjectByIDForInvestmentWait = func(context.Context, int64) (*domain.Project, error) {
		return project, nil
	}

	t.Cleanup(func() {
		getUserByIDForInvestmentWait = originalGetUser
		getProjectByIDForInvestmentWait = originalGetProject
	})
}

func newInvestmentWaitCreateContext(t *testing.T, ctx context.Context, projectID int64, qty int64) echo.Context {
	t.Helper()

	e := echo.New()
	e.Validator = validator.NewCustomValidator()

	body, err := json.Marshal(InvestmentWaitersInput{Qty: qty})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/projects/"+strconv.FormatInt(projectID, 10)+"/investment-wait-apply", bytes.NewReader(body))
	req = req.WithContext(ctx)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/projects/:id/investment-wait-apply")
	c.SetParamNames("id")
	c.SetParamValues(strconv.FormatInt(projectID, 10))

	token, _ := CreateUserTokenForTest(1, "user@example.com")
	c.Set("user", token)
	return c
}

func beginInvestmentWaitHTTPTransaction(t *testing.T, engine *xorm.Engine) (context.Context, *xorm.Session) {
	t.Helper()

	session := engine.NewSession()
	require.NoError(t, session.Begin())

	ctx := context.WithValue(context.Background(), factory.ConextDBType, session)
	ctx, _ = factory.WithAfterCommit(ctx)
	return ctx, session
}

func commitInvestmentWaitHTTPTransaction(t *testing.T, ctx context.Context, session *xorm.Session, engine *xorm.Engine) {
	t.Helper()

	require.NoError(t, session.Commit())
	afterCommitCtx := context.WithValue(ctx, factory.ConextDBType, engine)
	require.NoError(t, factory.RunAfterCommit(ctx, afterCommitCtx))
}

func runInvestmentWaitCreateRequest(t *testing.T, engine *xorm.Engine, controller ProjectApiController, projectID, qty int64) echo.Context {
	t.Helper()

	txCtx, session := beginInvestmentWaitHTTPTransaction(t, engine)
	defer session.Close()

	c := newInvestmentWaitCreateContext(t, txCtx, projectID, qty)
	if err := controller.InvestmentWaitCreate(c); err != nil {
		_ = session.Rollback()
		require.NoError(t, err)
	}
	commitInvestmentWaitHTTPTransaction(t, txCtx, session, engine)
	return c
}

func TestInvestmentWaitCreate_HTTP_AllowsMultipleApplications(t *testing.T) {
	ctx, engine := setupInvestmentWaitApplicationDB(t)

	user := &domain.User{
		Id:       1,
		Email:    "user@example.com",
		Name:     "test",
		UserType: domain.CORPORATE,
	}
	project := &domain.Project{
		Id:            10,
		Name:          "test fund",
		ProjectStatus: domain.OPEN_FUNDING,
		IsWaitInvest:  true,
		IsAmountLimit: true,
		SharePrice:    1_000_000,
	}
	stubInvestmentWaitHTTPDependencies(t, user, project)

	controller := ProjectApiController{}

	c1 := runInvestmentWaitCreateRequest(t, engine, controller, project.Id, 2)
	assert.Equal(t, http.StatusCreated, c1.Response().Status)

	c2 := runInvestmentWaitCreateRequest(t, engine, controller, project.Id, 2)
	assert.Equal(t, http.StatusCreated, c2.Response().Status)

	assert.Equal(t, int64(2), countInvestmentWaiters(t, ctx, project.Id, user.Id))
}

func TestInvestmentWaitCreate_HTTP_NotificationFailuresReturnCreated(t *testing.T) {
	tests := []struct {
		name        string
		setupNotify func()
	}{
		{
			name: "イベント通知失敗でも201",
			setupNotify: func() {
				publishInvestmentWaitRequestEvent = func(context.Context, *domain.User, *domain.InvestmentWaiters, *domain.Project) error {
					return errors.New("event publish failed")
				}
			},
		},
		{
			name: "メール送信失敗でも201",
			setupNotify: func() {
				sendInvestStandbyMail = func(context.Context, string, string, string, string, string, string, string) error {
					return errors.New("mail send failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, engine := setupInvestmentWaitApplicationDB(t)

			user := &domain.User{
				Id:       1,
				Email:    "user@example.com",
				Name:     "test",
				UserType: domain.CORPORATE,
			}
			project := &domain.Project{
				Id:            10,
				Name:          "test fund",
				ProjectStatus: domain.OPEN_FUNDING,
				IsWaitInvest:  true,
				IsAmountLimit: true,
				SharePrice:    1_000_000,
			}
			stubInvestmentWaitHTTPDependencies(t, user, project)
			tt.setupNotify()

			controller := ProjectApiController{}
			c := runInvestmentWaitCreateRequest(t, engine, controller, project.Id, 1)
			assert.Equal(t, http.StatusCreated, c.Response().Status)
			assert.Equal(t, int64(1), countInvestmentWaiters(t, ctx, project.Id, user.Id))
		})
	}
}

func TestInvestmentWaitCreate_HTTP_NotificationsRunAfterCommit(t *testing.T) {
	ctx, engine := setupInvestmentWaitApplicationDB(t)
	user := &domain.User{
		Id:       1,
		Email:    "user@example.com",
		Name:     "test",
		UserType: domain.CORPORATE,
	}
	project := &domain.Project{
		Id:            10,
		Name:          "test fund",
		ProjectStatus: domain.OPEN_FUNDING,
		IsWaitInvest:  true,
		IsAmountLimit: true,
		SharePrice:    1_000_000,
	}
	stubInvestmentWaitHTTPDependencies(t, user, project)

	var publishCount, mailCount int
	publishInvestmentWaitRequestEvent = func(context.Context, *domain.User, *domain.InvestmentWaiters, *domain.Project) error {
		publishCount++
		return nil
	}
	sendInvestStandbyMail = func(context.Context, string, string, string, string, string, string, string) error {
		mailCount++
		return nil
	}

	txCtx, session := beginInvestmentWaitHTTPTransaction(t, engine)
	defer session.Close()

	controller := ProjectApiController{}
	c := newInvestmentWaitCreateContext(t, txCtx, project.Id, 1)
	require.NoError(t, controller.InvestmentWaitCreate(c))
	assert.Equal(t, http.StatusCreated, c.Response().Status)
	assert.Equal(t, int64(1), countInvestmentWaiters(t, txCtx, project.Id, user.Id))
	assert.Equal(t, 0, publishCount)
	assert.Equal(t, 0, mailCount)

	commitInvestmentWaitHTTPTransaction(t, txCtx, session, engine)

	assert.Equal(t, int64(1), countInvestmentWaiters(t, ctx, project.Id, user.Id))
	assert.Equal(t, 1, publishCount)
	assert.Equal(t, 1, mailCount)
}
