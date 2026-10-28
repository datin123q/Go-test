package repayment_dividend

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	eventMessageMysql "github.com/illuminarean/energy-shares-backend-repo/repository/event_message/mysql"
	eventMessageSvc "github.com/illuminarean/energy-shares-backend-repo/service/event_message"
)

// publishLegalRecordEvent の Outbox パターン(MIR-1202)を、実際の event_message
// リポジトリ(sqlite)とサービス実装を通して検証する。トランザクション境界は
// admin-api のミドルウェア(contextDBWithName)を模して手動で再現する。

func setupOutboxTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(domain.EventMessage)))
	return engine
}

func newOutboxTestEventMessagesService(publishFunc func(v []byte) error) cdomain.EventMessageService {
	eventBusFactory := new(mocks.EventBusFactory)
	eventBusFactory.On("StatutoryLedgerEventBus", mock.Anything).Return(&eventMessageSvc.MockEventBus{
		PublishFunc: publishFunc,
	}, nil)
	return eventMessageSvc.New(
		factory.NewLoggerFactory(),
		eventBusFactory,
		eventMessageMysql.New(factory.NewDBFactory()),
	)
}

func newOutboxTestShareRepository(shares []domain.Share) *mocks.ShareRepository {
	repo := new(mocks.ShareRepository)
	repo.On("GetAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(shares, nil)
	return repo
}

func newOutboxTestProjectInvestmentExecuteRepository() *mocks.ProjectInvestmentExecuteRepository {
	repo := new(mocks.ProjectInvestmentExecuteRepository)
	repo.On("IncrementPublishFailedCount", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return repo
}

func newOutboxTestUsecase(eventMessagesService cdomain.EventMessageService, shareRepo *mocks.ShareRepository, withdrawalErrOnCall int) repaymentDividendUsecase {
	return newOutboxTestUsecaseWithExecuteRepo(eventMessagesService, shareRepo, withdrawalErrOnCall, newOutboxTestProjectInvestmentExecuteRepository())
}

func newOutboxTestUsecaseWithExecuteRepo(eventMessagesService cdomain.EventMessageService, shareRepo *mocks.ShareRepository, withdrawalErrOnCall int, projectInvestmentExecuteRepo *mocks.ProjectInvestmentExecuteRepository) repaymentDividendUsecase {
	callCount := 0
	return repaymentDividendUsecase{
		shareRepository: shareRepo,
		userAccountRepository: func() *mocks.UserAccountRepository {
			repo := new(mocks.UserAccountRepository)
			repo.On("GetByUserId", mock.Anything, mock.Anything).Return(domain.UserAccount{}, nil)
			return repo
		}(),
		projectInvestmentExecuteRepository: projectInvestmentExecuteRepo,
		eventMessagesService:               eventMessagesService,
		calculateCustomerLedgerBalance: func(ctx context.Context, userId int64) (int64, error) {
			return 100, nil
		},
		withdrawalUserTransaction: func(ctx context.Context, userId, balance, amount, projectId int64, projectType domain.PROJECT_TYPE, accountNumber string, transactionDetailType domain.USER_TRANSACTION_DETAIL_TYPE, transactionDate jsonutils.JSONTime) error {
			callCount++
			if withdrawalErrOnCall > 0 && callCount == withdrawalErrOnCall {
				return errors.New("simulated failure")
			}
			return nil
		},
		updateShareToOperation: func(ctx context.Context, share *domain.Share, operatorId int64) error {
			share.InvestStatus = domain.OPERATION_INVEST
			return nil
		},
		timeNow: time.Now,
	}
}

func TestPublishLegalRecordEvent_Outbox(t *testing.T) {
	project := mockProject1
	execute := domain.ProjectInvestmentExecute{Id: 1, ProjectId: project.Id}

	t.Run("成功時: コミット前はis_publish=falseで記録され、コミット後の送信でtrueになる", func(t *testing.T) {
		engine := setupOutboxTestEngine(t)
		session := engine.NewSession()
		defer session.Close()
		require.NoError(t, session.Begin())

		sentCount := 0
		eventMessagesService := newOutboxTestEventMessagesService(func(v []byte) error {
			sentCount++
			return nil
		})

		ctx := context.WithValue(context.Background(), factory.ConextDBType, session)
		ctx, ownsAfterCommit := factory.WithAfterCommit(ctx)
		require.True(t, ownsAfterCommit)

		shares := []domain.Share{{Id: "share-1", OwnerId: 3000, InvestStatus: domain.SUCCESS_INVEST}}
		r := newOutboxTestUsecase(eventMessagesService, newOutboxTestShareRepository(shares), 0)

		_, err := r.publishLegalRecordEvent(ctx, &project, execute, 1)
		require.NoError(t, err)

		// コミット前: まだ送信されておらず、is_publish=falseで1件(EVENT_TRANSACTION、
		// EVENT_FUND_TRANSFERはMIR-1202/#543で発行を停止済み)記録されていること。
		assert.Equal(t, 0, sentCount, "コミット前に送信してはいけない")
		unpublishedBefore, err := session.Where("is_publish = ?", false).Count(new(domain.EventMessage))
		require.NoError(t, err)
		assert.EqualValues(t, 1, unpublishedBefore)

		require.NoError(t, session.Commit())

		afterCommitCtx := context.WithValue(context.Background(), factory.ConextDBType, engine)
		require.NoError(t, factory.RunAfterCommit(ctx, afterCommitCtx))

		assert.Equal(t, 1, sentCount, "コミット後にまとめて送信されること")
		publishedAfter, err := engine.Where("is_publish = ?", true).Count(new(domain.EventMessage))
		require.NoError(t, err)
		assert.EqualValues(t, 1, publishedAfter)
	})

	t.Run("途中失敗時: ロールバックされevent_messageは1件も残らない", func(t *testing.T) {
		engine := setupOutboxTestEngine(t)
		session := engine.NewSession()
		defer session.Close()
		require.NoError(t, session.Begin())

		eventMessagesService := newOutboxTestEventMessagesService(func(v []byte) error {
			return nil
		})

		ctx := context.WithValue(context.Background(), factory.ConextDBType, session)
		ctx, _ = factory.WithAfterCommit(ctx)

		// 2件目のshareでwithdrawalUserTransactionを失敗させる。
		shares := []domain.Share{
			{Id: "share-1", OwnerId: 3000, InvestStatus: domain.SUCCESS_INVEST},
			{Id: "share-2", OwnerId: 3000, InvestStatus: domain.SUCCESS_INVEST},
		}
		r := newOutboxTestUsecase(eventMessagesService, newOutboxTestShareRepository(shares), 2)

		_, err := r.publishLegalRecordEvent(ctx, &project, execute, 1)
		require.Error(t, err, "2件目で失敗するはず")

		require.NoError(t, session.Rollback())

		count, err := engine.Count(new(domain.EventMessage))
		require.NoError(t, err)
		assert.EqualValues(t, 0, count, "ロールバックされたら1件目分のevent_messageも残らないこと")
	})

	t.Run("送信失敗があった実行レコードへ、失敗件数が加算される(MIR-1202の④)", func(t *testing.T) {
		engine := setupOutboxTestEngine(t)
		session := engine.NewSession()
		defer session.Close()
		require.NoError(t, session.Begin())

		// share-1(EVENT_TRANSACTION)は送信失敗、share-2は成功させる。
		// EVENT_FUND_TRANSFERはMIR-1202/#543で発行を停止済みのため、1shareにつき
		// イベントは1件。合計1件の送信失敗を期待する。
		callCount := 0
		eventMessagesService := newOutboxTestEventMessagesService(func(v []byte) error {
			callCount++
			if callCount <= 1 {
				return errors.New("kafka down")
			}
			return nil
		})

		ctx := context.WithValue(context.Background(), factory.ConextDBType, session)
		ctx, _ = factory.WithAfterCommit(ctx)

		shares := []domain.Share{
			{Id: "share-1", OwnerId: 3000, InvestStatus: domain.SUCCESS_INVEST},
			{Id: "share-2", OwnerId: 3000, InvestStatus: domain.SUCCESS_INVEST},
		}
		executeRepo := newOutboxTestProjectInvestmentExecuteRepository()
		r := newOutboxTestUsecaseWithExecuteRepo(eventMessagesService, newOutboxTestShareRepository(shares), 0, executeRepo)

		_, err := r.publishLegalRecordEvent(ctx, &project, execute, 1)
		require.NoError(t, err)

		require.NoError(t, session.Commit())

		afterCommitCtx := context.WithValue(context.Background(), factory.ConextDBType, engine)
		// SendRecordedEvent自体は送信失敗をエラーとして返さない(is_publish=falseで
		// 記録するだけ)ため、RunAfterCommit全体はエラーなく終わる。
		require.NoError(t, factory.RunAfterCommit(ctx, afterCommitCtx))

		executeRepo.AssertCalled(t, "IncrementPublishFailedCount", mock.Anything, execute.Id, int64(1))
	})

	t.Run("送信自体は成功してもDB更新が失敗した場合も、送信失敗としてカウントされる", func(t *testing.T) {
		// SendRecordedEventがpublished=trueかつerr!=nil(Kafka送信は成功したが
		// is_publishへのUpdateだけ失敗した)ケースをモックで再現する。この場合も
		// DB上はis_publish=falseのまま残り再送対象になるため、送信失敗として
		// カウントすべき。
		eventMessagesService := new(mocks.EventMessageService)
		eventMessagesService.On("RecordEvent", mock.Anything, mock.Anything, mock.Anything).
			Return(domain.EventMessage{Uuid: "test-uuid"}, nil)
		eventMessagesService.On("SendRecordedEvent", mock.Anything, mock.Anything).
			Return(true, errors.New("update failed"))

		ctx, _ := factory.WithAfterCommit(context.Background())

		shares := []domain.Share{{Id: "share-1", OwnerId: 3000, InvestStatus: domain.SUCCESS_INVEST}}
		executeRepo := newOutboxTestProjectInvestmentExecuteRepository()
		r := newOutboxTestUsecaseWithExecuteRepo(eventMessagesService, newOutboxTestShareRepository(shares), 0, executeRepo)

		_, err := r.publishLegalRecordEvent(ctx, &project, execute, 1)
		require.NoError(t, err)

		// RunAfterCommitはSendRecordedEventのUpdate失敗をJoinして返すが、
		// ここではIncrementPublishFailedCountが正しい件数で呼ばれたことだけを見る。
		// EVENT_FUND_TRANSFERはMIR-1202/#543で発行を停止済みのため、1shareにつき
		// イベントは1件(EVENT_TRANSACTIONのみ)。
		_ = factory.RunAfterCommit(ctx, context.Background())

		executeRepo.AssertCalled(t, "IncrementPublishFailedCount", mock.Anything, execute.Id, int64(1))
	})
}
