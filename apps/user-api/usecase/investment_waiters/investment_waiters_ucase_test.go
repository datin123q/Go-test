package investment_waiters

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

func newInvestmentWaitersCancelTxManager(t *testing.T) *mocks.TxManager {
	t.Helper()

	txManager := mocks.NewTxManager(t)
	txManager.On("WithTx", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	return txManager
}

func TestInvestmentWaitersUsecase_Cancel(t *testing.T) {
	const (
		userId              = int64(100)
		investmentWaitersId = int64(200)
	)

	actor := jwthelper.User{
		Id:    userId,
		Name:  "テストユーザー",
		Email: "user@example.com",
	}
	request := udomain.InvestmentWaitersCancelRequest{
		JWTUser:            actor,
		UserID:             userId,
		InvestmentWaiterID: investmentWaitersId,
	}
	waitingInvestmentWaiters := domain.InvestmentWaiters{
		Id:        investmentWaitersId,
		ProjectId: 300,
		UserId:    userId,
		Qty:       2,
		Amount:    20000,
		Status:    domain.WAITING,
	}

	tests := []struct {
		name        string
		request     udomain.InvestmentWaitersCancelRequest
		txManager   func(t *testing.T) *mocks.TxManager
		waitersRepo func(t *testing.T) *mocks.InvestmentWaitersRepository
		historyRepo func(t *testing.T) *mocks.InvestmentWaitHistoryRepository
		wantErr     error
	}{
		{
			name:      "キャンセル待ち申請を取り消して会員操作履歴を登録する",
			request:   request,
			txManager: newInvestmentWaitersCancelTxManager,
			waitersRepo: func(t *testing.T) *mocks.InvestmentWaitersRepository {
				repo := mocks.NewInvestmentWaitersRepository(t)
				repo.On("LockById", mock.Anything, investmentWaitersId).
					Return(waitingInvestmentWaiters, nil)
				repo.On("LockAndSumAdvancedQty", mock.Anything, investmentWaitersId).
					Return(int64(0), nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(got domain.InvestmentWaiters) bool {
					return got.Id == waitingInvestmentWaiters.Id &&
						got.UserId == waitingInvestmentWaiters.UserId &&
						got.Status == domain.CANCELLED
				})).Return(int64(1), nil)
				return repo
			},
			historyRepo: func(t *testing.T) *mocks.InvestmentWaitHistoryRepository {
				repo := mocks.NewInvestmentWaitHistoryRepository(t)
				repo.On("Create", mock.Anything, domain.InvestmentWaitHistory{
					InvestmentWaitId: investmentWaitersId,
					SubstanceType:    string(domain.INVESTMENT_WAIT_HISTORY_STATUS),
					Name:             actor.Name,
					Email:            actor.Email,
					PrevStatus:       domain.WAITING,
					CurrStatus:       domain.CANCELLED,
					ProcessMethod:    domain.USER_HISTORY_PROCESS_MEMBER,
				}).Return(nil)
				return repo
			},
		},
		{
			name: "JWTユーザーとパスのユーザーが不一致の場合は権限エラーを返す",
			request: func() udomain.InvestmentWaitersCancelRequest {
				invalid := request
				invalid.UserID = 999
				return invalid
			}(),
			txManager: func(t *testing.T) *mocks.TxManager {
				return mocks.NewTxManager(t)
			},
			waitersRepo: func(t *testing.T) *mocks.InvestmentWaitersRepository {
				return mocks.NewInvestmentWaitersRepository(t)
			},
			historyRepo: func(t *testing.T) *mocks.InvestmentWaitHistoryRepository {
				return mocks.NewInvestmentWaitHistoryRepository(t)
			},
			wantErr: cdomain.ErrForbidden,
		},
		{
			name:      "キャンセル待ち申請が存在しない場合はNotFoundエラーを返す",
			request:   request,
			txManager: newInvestmentWaitersCancelTxManager,
			waitersRepo: func(t *testing.T) *mocks.InvestmentWaitersRepository {
				repo := mocks.NewInvestmentWaitersRepository(t)
				repo.On("LockById", mock.Anything, investmentWaitersId).
					Return(domain.InvestmentWaiters{}, cdomain.ErrItemNotFound)
				return repo
			},
			historyRepo: func(t *testing.T) *mocks.InvestmentWaitHistoryRepository {
				return mocks.NewInvestmentWaitHistoryRepository(t)
			},
			wantErr: cdomain.ErrItemNotFound,
		},
		{
			name:      "他ユーザーのキャンセル待ち申請は取り消せない",
			request:   request,
			txManager: newInvestmentWaitersCancelTxManager,
			waitersRepo: func(t *testing.T) *mocks.InvestmentWaitersRepository {
				repo := mocks.NewInvestmentWaitersRepository(t)
				otherUserWaiter := waitingInvestmentWaiters
				otherUserWaiter.UserId = 999
				repo.On("LockById", mock.Anything, investmentWaitersId).
					Return(otherUserWaiter, nil)
				return repo
			},
			historyRepo: func(t *testing.T) *mocks.InvestmentWaitHistoryRepository {
				return mocks.NewInvestmentWaitHistoryRepository(t)
			},
			wantErr: cdomain.ErrForbidden,
		},
		{
			name:      "WAITING_END状態のキャンセル待ち申請は取り消せない",
			request:   request,
			txManager: newInvestmentWaitersCancelTxManager,
			waitersRepo: func(t *testing.T) *mocks.InvestmentWaitersRepository {
				repo := mocks.NewInvestmentWaitersRepository(t)
				waitingEnd := waitingInvestmentWaiters
				waitingEnd.Status = domain.WAITING_END
				repo.On("LockById", mock.Anything, investmentWaitersId).
					Return(waitingEnd, nil)
				return repo
			},
			historyRepo: func(t *testing.T) *mocks.InvestmentWaitHistoryRepository {
				return mocks.NewInvestmentWaitHistoryRepository(t)
			},
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name:      "WAITINGでもShareが作成済みのキャンセル待ち申請は取り消せない",
			request:   request,
			txManager: newInvestmentWaitersCancelTxManager,
			waitersRepo: func(t *testing.T) *mocks.InvestmentWaitersRepository {
				repo := mocks.NewInvestmentWaitersRepository(t)
				repo.On("LockById", mock.Anything, investmentWaitersId).
					Return(waitingInvestmentWaiters, nil)
				repo.On("LockAndSumAdvancedQty", mock.Anything, investmentWaitersId).
					Return(int64(2), nil)
				return repo
			},
			historyRepo: func(t *testing.T) *mocks.InvestmentWaitHistoryRepository {
				return mocks.NewInvestmentWaitHistoryRepository(t)
			},
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name:      "キャンセル済みの申請を再度取り消せない",
			request:   request,
			txManager: newInvestmentWaitersCancelTxManager,
			waitersRepo: func(t *testing.T) *mocks.InvestmentWaitersRepository {
				repo := mocks.NewInvestmentWaitersRepository(t)
				cancelled := waitingInvestmentWaiters
				cancelled.Status = domain.CANCELLED
				repo.On("LockById", mock.Anything, investmentWaitersId).
					Return(cancelled, nil)
				return repo
			},
			historyRepo: func(t *testing.T) *mocks.InvestmentWaitHistoryRepository {
				return mocks.NewInvestmentWaitHistoryRepository(t)
			},
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name:      "キャンセル待ち申請の更新に失敗した場合はエラーを返す",
			request:   request,
			txManager: newInvestmentWaitersCancelTxManager,
			waitersRepo: func(t *testing.T) *mocks.InvestmentWaitersRepository {
				repo := mocks.NewInvestmentWaitersRepository(t)
				repo.On("LockById", mock.Anything, investmentWaitersId).
					Return(waitingInvestmentWaiters, nil)
				repo.On("LockAndSumAdvancedQty", mock.Anything, investmentWaitersId).
					Return(int64(0), nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("domain.InvestmentWaiters")).
					Return(int64(0), assert.AnError)
				return repo
			},
			historyRepo: func(t *testing.T) *mocks.InvestmentWaitHistoryRepository {
				return mocks.NewInvestmentWaitHistoryRepository(t)
			},
			wantErr: assert.AnError,
		},
		{
			name:      "更新件数が0件の場合はエラーを返す",
			request:   request,
			txManager: newInvestmentWaitersCancelTxManager,
			waitersRepo: func(t *testing.T) *mocks.InvestmentWaitersRepository {
				repo := mocks.NewInvestmentWaitersRepository(t)
				repo.On("LockById", mock.Anything, investmentWaitersId).
					Return(waitingInvestmentWaiters, nil)
				repo.On("LockAndSumAdvancedQty", mock.Anything, investmentWaitersId).
					Return(int64(0), nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("domain.InvestmentWaiters")).
					Return(int64(0), nil)
				return repo
			},
			historyRepo: func(t *testing.T) *mocks.InvestmentWaitHistoryRepository {
				return mocks.NewInvestmentWaitHistoryRepository(t)
			},
			wantErr: cdomain.ErrItemNotAffected,
		},
		{
			name:      "履歴登録に失敗した場合はトランザクションへエラーを返す",
			request:   request,
			txManager: newInvestmentWaitersCancelTxManager,
			waitersRepo: func(t *testing.T) *mocks.InvestmentWaitersRepository {
				repo := mocks.NewInvestmentWaitersRepository(t)
				repo.On("LockById", mock.Anything, investmentWaitersId).
					Return(waitingInvestmentWaiters, nil)
				repo.On("LockAndSumAdvancedQty", mock.Anything, investmentWaitersId).
					Return(int64(0), nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("domain.InvestmentWaiters")).
					Return(int64(1), nil)
				return repo
			},
			historyRepo: func(t *testing.T) *mocks.InvestmentWaitHistoryRepository {
				repo := mocks.NewInvestmentWaitHistoryRepository(t)
				repo.On("Create", mock.Anything, mock.AnythingOfType("domain.InvestmentWaitHistory")).
					Return(assert.AnError)
				return repo
			},
			wantErr: assert.AnError,
		},
		{
			name:    "トランザクション開始に失敗した場合はエラーを返す",
			request: request,
			txManager: func(t *testing.T) *mocks.TxManager {
				txManager := mocks.NewTxManager(t)
				txManager.On("WithTx", mock.Anything, mock.Anything).Return(assert.AnError)
				return txManager
			},
			waitersRepo: func(t *testing.T) *mocks.InvestmentWaitersRepository {
				return mocks.NewInvestmentWaitersRepository(t)
			},
			historyRepo: func(t *testing.T) *mocks.InvestmentWaitHistoryRepository {
				return mocks.NewInvestmentWaitHistoryRepository(t)
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentWaitersUsecase{
				txManager:                       tt.txManager(t),
				investmentWaitersRepository:     tt.waitersRepo(t),
				investmentWaitHistoryRepository: tt.historyRepo(t),
			}

			got, err := usecase.Cancel(context.Background(), tt.request)

			assert.Equal(t, udomain.InvestmentWaitersCancelResponse{}, got)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}

			assert.Error(t, err)
			assert.True(t, stderrors.Is(err, tt.wantErr), "error = %v, want %v", err, tt.wantErr)
		})
	}
}
