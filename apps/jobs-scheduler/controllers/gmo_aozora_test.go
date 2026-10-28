package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"energyshares-job-scheduler/jobs"

	client "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/contexts"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/eventbus"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

const concurrentVaDepositTestTimeout = 5 * time.Second

// concurrentVaDepositInsertHook は両経路の事前確認が完了し、
// 同じVA入金をINSERTしようとする点で同期する。
// 先着側のINSERT完了後、後着側にMySQLの対象一意制約違反を返す。
type concurrentVaDepositInsertHook struct {
	mu          sync.Mutex
	insertCount int
	bothReady   chan struct{}
	winnerDone  chan struct{}
	winner      *contexts.ContextHook
	readyOnce   sync.Once
	doneOnce    sync.Once
}

func newConcurrentVaDepositInsertHook() *concurrentVaDepositInsertHook {
	return &concurrentVaDepositInsertHook{
		bothReady:  make(chan struct{}),
		winnerDone: make(chan struct{}),
	}
}

func (h *concurrentVaDepositInsertHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	if !strings.Contains(strings.ToUpper(c.SQL), "INSERT INTO `VA_DEPOSIT`") {
		return c.Ctx, nil
	}

	h.mu.Lock()
	h.insertCount++
	insertOrder := h.insertCount
	if insertOrder == 1 {
		h.winner = c
	}
	if insertOrder == 2 {
		h.readyOnce.Do(func() { close(h.bothReady) })
	}
	h.mu.Unlock()

	switch insertOrder {
	case 1:
		select {
		case <-h.bothReady:
			return c.Ctx, nil
		case <-time.After(concurrentVaDepositTestTimeout):
			return c.Ctx, stderrors.New("timed out waiting for both va deposit inserts")
		}
	case 2:
		select {
		case <-h.winnerDone:
			return c.Ctx, &mysql.MySQLError{
				Number:  1062,
				Message: "Duplicate entry for key 'uq_va_deposit_account_item_key'",
			}
		case <-time.After(concurrentVaDepositTestTimeout):
			return c.Ctx, stderrors.New("timed out waiting for winning va deposit insert")
		}
	default:
		return c.Ctx, nil
	}
}

func (h *concurrentVaDepositInsertHook) AfterProcess(c *contexts.ContextHook) error {
	h.mu.Lock()
	isWinner := c == h.winner
	h.mu.Unlock()
	if isWinner {
		h.doneOnce.Do(func() { close(h.winnerDone) })
	}

	return nil
}

func (h *concurrentVaDepositInsertHook) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.insertCount
}

func setupAozoraWebHookTestDB(t *testing.T) (context.Context, *xorm.Engine) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	require.NoError(t, engine.Sync(new(domain.UserVirtualAccount), new(domain.VaDeposit)))
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	return ctx, engine
}

func setupConcurrentAozoraDepositTestDB(t *testing.T) (context.Context, *xorm.Engine) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "va-deposit-concurrency.db")
	engine, err := xorm.NewEngine("sqlite3", "file:"+dbPath+"?_busy_timeout=5000&_journal_mode=WAL")
	require.NoError(t, err)
	engine.SetMaxOpenConns(4)
	require.NoError(t, engine.Sync(
		new(domain.UserVirtualAccount),
		new(domain.VaDeposit),
		new(domain.User),
		new(domain.UserAccount),
		new(domain.UserAssetView),
		new(domain.UserTransaction),
		new(domain.EventMessage),
		new(domain.DepositRequest),
		new(domain.DepositRequestLackHistory),
		new(domain.Share),
		new(domain.Project),
		new(domain.Spv),
		new(domain.ProjectNews),
		new(domain.Attachment),
		new(domain.ProjectFile),
		new(domain.RepaymentDividendSchedule),
		new(domain.ProjectInvestmentExecute),
		new(domain.ProjectViewCount),
		new(domain.ProjectMilestone),
		new(domain.AdminUser),
		new(domain.Role),
		new(domain.OperatorHistory),
	))
	_, err = engine.Exec(
		"CREATE UNIQUE INDEX uq_va_deposit_account_item_key " +
			"ON va_deposit(va_account_number, item_key)",
	)
	require.NoError(t, err)

	superAdmin := domain.AdminUser{
		Email: "super-admin@example.test",
		Name:  "concurrency test admin",
	}
	_, err = engine.Insert(&superAdmin)
	require.NoError(t, err)
	_, err = engine.Insert(&domain.Role{
		UserId:   superAdmin.Id,
		RoleCode: string(domain.ROLE_SUPER_ADMIN),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	return ctx, engine
}

func seedConcurrentAozoraDepositUser(t *testing.T, engine *xorm.Engine, transaction VaTransaction) int64 {
	t.Helper()

	user := domain.User{
		Uid:      "concurrency-test-user",
		CommonId: "concurrency-test-common-id",
		Name:     "concurrency test user",
		Email:    "user@example.test",
		Password: "test-password",
		Salt:     "test-salt",
	}
	_, err := engine.Insert(&user)
	require.NoError(t, err)

	initialBalance := int64(0)
	_, err = engine.Insert(&domain.UserAccount{
		UserId:  user.Id,
		Balance: &initialBalance,
		Version: 1,
	})
	require.NoError(t, err)
	_, err = engine.Insert(&domain.UserAssetView{
		UserId:  user.Id,
		Deposit: 100000,
	})
	require.NoError(t, err)
	_, err = engine.Insert(&domain.UserVirtualAccount{
		UserId:           user.Id,
		VaBranchCode:     transaction.VaBranchCode,
		VaAccountNumber:  transaction.VaAccountNumber,
		VaHolderNameKana: "TEST USER",
		ProjectType:      domain.EQUITY,
		LinkedStatus:     domain.ON_VA_ACCOUNT,
	})
	require.NoError(t, err)

	project := domain.Project{
		CommonId:          "concurrency-test-project",
		Name:              "concurrency test project",
		CategoryType:      domain.SOLAR,
		Description:       "concurrency test project",
		ProjectType:       domain.EQUITY,
		ProjectStatus:     domain.OPEN_FUNDING,
		OwnerId:           user.Id,
		UpdateId:          user.Id,
		TargetShareQty:    2,
		TargetShareAmount: 200000,
	}
	_, err = engine.Insert(&project)
	require.NoError(t, err)

	now := jsonutils.JSONTime(time.Now())
	share := domain.Share{
		Id:           "concurrency-test-share",
		ProjectId:    project.Id,
		ProjectType:  domain.EQUITY,
		OwnerId:      user.Id,
		UpdateId:     user.Id,
		InvestStatus: domain.REQUEST_INVEST,
		Amount:       200000,
		Qty:          2,
		Price:        100000,
		InvestDate:   now,
	}
	_, err = engine.Insert(&share)
	require.NoError(t, err)
	_, err = engine.Insert(&domain.DepositRequest{
		UserId:               user.Id,
		ShareId:              share.Id,
		ProjectId:            project.Id,
		ProjectType:          domain.EQUITY,
		RequestAmount:        share.Amount,
		DepositAmount:        0,
		DepositCheckStatus:   domain.NO_DEPOSIT,
		ExpireDepositDate:    jsonutils.JSONTime(time.Now().Add(48 * time.Hour)),
		DepositBankName:      "test bank",
		DepositBranchName:    "test branch",
		DepositBranchCode:    transaction.VaBranchCode,
		DepositAccountNumber: transaction.VaAccountNumber,
	})
	require.NoError(t, err)

	return user.Id
}

func aozoraWebHookTransactionForTest() VaTransaction {
	return VaTransaction{
		VaId:            "test-va-id-001",
		TransactionDate: "2026-01-15",
		ValueDate:       "2026-01-15",
		VaBranchCode:    "001",
		VaAccountNumber: "1234567",
		DepositAmount:   "100000",
		ItemKey:         "test-item-key-001",
	}
}

func executeAozoraWebHookForTest(ctx context.Context, transaction VaTransaction) (*httptest.ResponseRecorder, error) {
	body, err := json.Marshal(AozoraWebHookMessageInput{VaTransaction: transaction})
	if err != nil {
		return nil, err
	}
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body)).WithContext(ctx)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	echoContext := echo.New().NewContext(req, recorder)

	err = (AozoraApiControllers{}).AozoraWebHook(echoContext)
	return recorder, err
}

func TestAozoraWebHook_ReturnsOKForMatchingRegisteredTransaction(t *testing.T) {
	ctx, engine := setupAozoraWebHookTestDB(t)
	existing, err := aozoraWebHookTransactionForTest().ToModel()
	require.NoError(t, err)
	_, err = engine.Insert(existing)
	require.NoError(t, err)

	recorder, err := executeAozoraWebHookForTest(ctx, aozoraWebHookTransactionForTest())

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	count, err := engine.Unscoped().Count(new(domain.VaDeposit))
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestAozoraWebHook_ReturnsErrorForMismatchedRegisteredTransaction(t *testing.T) {
	ctx, engine := setupAozoraWebHookTestDB(t)
	existing, err := aozoraWebHookTransactionForTest().ToModel()
	require.NoError(t, err)
	_, err = engine.Insert(existing)
	require.NoError(t, err)

	incoming := aozoraWebHookTransactionForTest()
	incoming.DepositAmount = "200000"
	_, err = executeAozoraWebHookForTest(ctx, incoming)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrVaDepositTransactionConflict)
	var factoryErr *factory.Error
	require.True(t, stderrors.As(err, &factoryErr))
	assert.Equal(t, http.StatusInternalServerError, factoryErr.Status())
	count, countErr := engine.Unscoped().Count(new(domain.VaDeposit))
	require.NoError(t, countErr)
	assert.Equal(t, int64(1), count)
}

func TestAozoraWebHook_ReturnsBadRequestForEmptyItemKey(t *testing.T) {
	tests := []struct {
		name    string
		itemKey string
	}{
		{name: "未指定", itemKey: ""},
		{name: "空白のみ", itemKey: " \t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, engine := setupAozoraWebHookTestDB(t)
			transaction := aozoraWebHookTransactionForTest()
			transaction.ItemKey = tt.itemKey

			_, err := executeAozoraWebHookForTest(ctx, transaction)

			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrVaDepositItemKeyRequired)
			var factoryErr *factory.Error
			require.True(t, stderrors.As(err, &factoryErr))
			assert.Equal(t, http.StatusBadRequest, factoryErr.Status())
			count, countErr := engine.Unscoped().Count(new(domain.VaDeposit))
			require.NoError(t, countErr)
			assert.Equal(t, int64(0), count)
		})
	}
}

func TestVaDepositInputToModel_RejectsEmptyItemKey(t *testing.T) {
	_, err := (VaDepositInput{ItemKey: "\n"}).ToModel()

	assert.ErrorIs(t, err, domain.ErrVaDepositItemKeyRequired)
}

func TestAozoraWebHookAndPolling_ProcessSameTransactionConcurrently(t *testing.T) {
	ctx, engine := setupConcurrentAozoraDepositTestDB(t)
	webhookTransaction := aozoraWebHookTransactionForTest()
	userID := seedConcurrentAozoraDepositUser(t, engine, webhookTransaction)
	_, err := eventbus.NewMockEventBus()
	require.NoError(t, err)

	var mailRequestCount atomic.Int32
	mailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/mails/send-request" {
			mailRequestCount.Add(1)
		}
		w.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		_, _ = w.Write([]byte("[1]"))
	}))
	defer mailServer.Close()
	previousEmailServiceURL := serviceutils.GetEmailServiceUrl()
	serviceutils.SetEmailServiceUrl(mailServer.URL)
	t.Cleanup(func() {
		serviceutils.SetEmailServiceUrl(previousEmailServiceURL)
	})

	hook := newConcurrentVaDepositInsertHook()
	engine.AddHook(hook)

	pollingTransaction := client.VaTransaction{
		VaId:              webhookTransaction.VaId,
		TransactionDate:   webhookTransaction.TransactionDate,
		ValueDate:         webhookTransaction.ValueDate,
		VaBranchCode:      webhookTransaction.VaBranchCode,
		VaBranchNameKana:  webhookTransaction.VaBranchNameKana,
		VaAccountNumber:   webhookTransaction.VaAccountNumber,
		VaAccountNameKana: webhookTransaction.VaAccountNameKana,
		DepositAmount:     webhookTransaction.DepositAmount,
		RemitterNameKana:  webhookTransaction.RemitterNameKana,
		PaymentBankName:   webhookTransaction.PaymentBankName,
		PaymentBranchName: webhookTransaction.PaymentBranchName,
		PartnerName:       webhookTransaction.PartnerName,
		Remarks:           webhookTransaction.Remarks,
		ItemKey:           webhookTransaction.ItemKey,
	}
	laterTransaction := pollingTransaction
	laterTransaction.VaId = "test-va-id-002"
	laterTransaction.VaAccountNumber = "1234568"
	laterTransaction.ItemKey = "test-item-key-002"

	type webhookResult struct {
		recorder *httptest.ResponseRecorder
		err      error
	}
	start := make(chan struct{})
	webhookResults := make(chan webhookResult, 1)
	jobResults := make(chan error, 1)

	go func() {
		<-start
		recorder, err := executeAozoraWebHookForTest(ctx, webhookTransaction)
		webhookResults <- webhookResult{recorder: recorder, err: err}
	}()
	go func() {
		<-start
		jobResults <- jobs.ProcessEquityVaTransactions(ctx, []client.VaTransaction{
			pollingTransaction,
			laterTransaction,
		})
	}()
	close(start)

	var webhookOutcome webhookResult
	select {
	case webhookOutcome = <-webhookResults:
	case <-time.After(concurrentVaDepositTestTimeout):
		t.Fatal("timed out waiting for webhook processing")
	}
	select {
	case err := <-jobResults:
		require.NoError(t, err)
	case <-time.After(concurrentVaDepositTestTimeout):
		t.Fatal("timed out waiting for polling processing")
	}

	require.NoError(t, webhookOutcome.err)
	require.NotNil(t, webhookOutcome.recorder)
	assert.Equal(t, http.StatusOK, webhookOutcome.recorder.Code)
	assert.Equal(t, 3, hook.Count(), "shared transaction attempts twice and later transaction once")

	sharedCount, err := engine.Unscoped().
		Where("va_account_number = ? AND item_key = ?", webhookTransaction.VaAccountNumber, webhookTransaction.ItemKey).
		Count(new(domain.VaDeposit))
	require.NoError(t, err)
	assert.Equal(t, int64(1), sharedCount)
	laterCount, err := engine.Unscoped().
		Where("va_account_number = ? AND item_key = ?", laterTransaction.VaAccountNumber, laterTransaction.ItemKey).
		Count(new(domain.VaDeposit))
	require.NoError(t, err)
	assert.Equal(t, int64(1), laterCount)
	totalCount, err := engine.Unscoped().Count(new(domain.VaDeposit))
	require.NoError(t, err)
	assert.Equal(t, int64(2), totalCount)

	userTransactionCount, err := engine.Where("user_id = ?", userID).
		Count(new(domain.UserTransaction))
	require.NoError(t, err)
	assert.Equal(t, int64(1), userTransactionCount)
	eventMessageCount, err := engine.Where("event_category = ?", domain.EVENT_TRANSACTION).
		Count(new(domain.EventMessage))
	require.NoError(t, err)
	assert.Equal(t, int64(1), eventMessageCount)
	assert.Equal(t, int32(1), mailRequestCount.Load())
}
