package deposit_request

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_depositRequestService_OkDeposit(t *testing.T) {
	type args struct {
		ctx            context.Context
		depositRequest domain.DepositRequest
	}
	tests := []struct {
		name                     string
		args                     args
		depositRequestRepository func() *mocks.DepositRequestRepository
		wantErr                  bool
	}{
		{
			name: "입금 요청을 성공적으로 처리합니다 : 입금액 : 10000 => 정상적으로 업데이트를 기대합니다.",
			args: args{
				ctx: context.Background(),
				depositRequest: domain.DepositRequest{
					Id:            1000,
					RequestAmount: 10000,
					DepositDate:   jsonutils.JSONTime(time.Unix(0, 0)),
				},
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := &mocks.DepositRequestRepository{}
				repo.On("Update", context.Background(), mock.MatchedBy(func(depositRequest domain.DepositRequest) bool {
					ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(domain.DepositRequest{
						Id:                 1000,
						RequestAmount:      10000,
						DepositCheckStatus: domain.OK_DEPOSIT,
						DepositAmount:      10000,
					}, depositRequest, ignoreTypes)
				})).Return(nil)
				return repo
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &depositRequestService{
				depositRequestRepository: tt.depositRequestRepository(),
			}

			err := svc.OkDeposit(tt.args.ctx, tt.args.depositRequest)
			assert.Equalf(t, tt.wantErr, err != nil, "OkDeposit() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_depositRequestService_CancelDepositSkipsAutoDepositCheckDuringFundFailureProcessing(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	assert.NoError(t, err)
	defer engine.Close()

	assert.NoError(t, engine.Sync(new(domain.User), new(domain.UserAccount)))
	_, err = engine.Insert(&domain.User{
		Id:       2000,
		Uid:      "user-2000",
		CommonId: "common-2000",
		Email:    "user-2000@example.com",
		Password: "password",
		Salt:     "salt",
	})
	assert.NoError(t, err)
	_, err = engine.Insert(&domain.UserAccount{
		Id:      3000,
		UserId:  2000,
		Balance: lo.ToPtr(int64(0)),
	})
	assert.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	fundFailureCtx := domain.WithFundFailureProcessing(ctx)
	repo := &mocks.DepositRequestRepository{}
	repo.On("Update", mock.Anything, mock.MatchedBy(func(depositRequest domain.DepositRequest) bool {
		return depositRequest.Id == 1000 &&
			depositRequest.RefundAmount == 10000 &&
			depositRequest.DepositCheckStatus == domain.CANCEL_DEPOSIT
	})).Return(nil)

	svc := &depositRequestService{
		depositRequestRepository: repo,
	}

	err = svc.CancelDeposit(fundFailureCtx, domain.DepositRequest{
		Id:            1000,
		UserId:        2000,
		DepositAmount: 10000,
		RefundAmount:  5000,
	})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "GetNoDepositRequestByUserId", mock.Anything, int64(2000))

	var account domain.UserAccount
	has, err := engine.ID(3000).Get(&account)
	assert.NoError(t, err)
	assert.True(t, has)
	assert.Equal(t, int64(5000), lo.FromPtr(account.Balance))
}

func Test_depositRequestService_CancelEquityDeposit(t *testing.T) {
	// TODO: DB接続処理をDIできるようにしてテストを復活させる
	t.Skip()

	type args struct {
		ctx            context.Context
		depositRequest domain.DepositRequest
	}
	tests := []struct {
		name                      string
		args                      args
		depositRequestRepository  func() *mocks.DepositRequestRepository
		userAccountRepository     func() *mocks.UserAccountRepository
		userTransactionRepository func() *mocks.UserTransactionRepository
		wantErr                   bool
	}{
		{
			name: "입금요청 취소로 인한 실제총액이 0보다 클 경우 => 정상적으로 업데이트를 기대합니다.",
			args: args{
				ctx: context.Background(),
				depositRequest: domain.DepositRequest{
					Id:                   1000,
					UserId:               2000,
					DepositAmount:        10000,
					RefundAmount:         5000,
					DepositAccountNumber: "123456789",
				},
			},
			userAccountRepository: func() *mocks.UserAccountRepository {
				repo := &mocks.UserAccountRepository{}
				repo.On("GetByUserId", context.Background(), int64(2000)).Return(domain.UserAccount{
					Id:      3000,
					UserId:  2000,
					Balance: lo.ToPtr(int64(0)),
				}, nil)

				repo.On("Update", mock.Anything, mock.MatchedBy(func(userAccount domain.UserAccount) bool {
					return lo.FromPtr(userAccount.Balance) == 5000
				})).Return(nil)
				return repo
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := &mocks.DepositRequestRepository{}
				repo.On("GetNoDepositRequestByUserId", context.Background(), int64(2000)).
					Return([]domain.DepositRequest{}, nil)
				repo.On("Update", context.Background(), domain.DepositRequest{
					Id:                   1000,
					UserId:               2000,
					DepositAmount:        10000,
					RefundAmount:         10000,
					DepositAccountNumber: "123456789",
					DepositCheckStatus:   domain.CANCEL_DEPOSIT,
				}).Return(nil)
				return repo
			},
			userTransactionRepository: func() *mocks.UserTransactionRepository { return &mocks.UserTransactionRepository{} },
			wantErr:                   false,
		},
		{
			name: "입금요청 취소로 인한 실제총액이 0일 경우 => 입금 요청상태만 변경하고 종료합니다.",
			args: args{
				ctx: context.Background(),
				depositRequest: domain.DepositRequest{
					Id:                   1000,
					UserId:               2000,
					DepositAmount:        10000,
					RefundAmount:         10000,
					DepositAccountNumber: "123456789",
				},
			},
			userAccountRepository: func() *mocks.UserAccountRepository {
				return &mocks.UserAccountRepository{}
			},
			depositRequestRepository: func() *mocks.DepositRequestRepository {
				repo := &mocks.DepositRequestRepository{}
				repo.On("Update", context.Background(), domain.DepositRequest{
					Id:                   1000,
					UserId:               2000,
					DepositAmount:        10000,
					RefundAmount:         10000,
					DepositAccountNumber: "123456789",
					DepositCheckStatus:   domain.CANCEL_DEPOSIT,
				}).Return(nil)
				return repo
			},
			userTransactionRepository: func() *mocks.UserTransactionRepository {
				return &mocks.UserTransactionRepository{}
			},
			wantErr: false,
		},

		{
			name: "이미 입금 요청이 취소된 경우에는 에러를 반환합니다.",
			args: args{
				ctx: context.Background(),
				depositRequest: domain.DepositRequest{
					Id:                 1000,
					DepositCheckStatus: domain.CANCEL_DEPOSIT,
				},
			},
			userTransactionRepository: func() *mocks.UserTransactionRepository {
				return &mocks.UserTransactionRepository{}
			},

			userAccountRepository: func() *mocks.UserAccountRepository {
				return &mocks.UserAccountRepository{}
			},

			depositRequestRepository: func() *mocks.DepositRequestRepository {
				return &mocks.DepositRequestRepository{}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &depositRequestService{
				depositRequestRepository:  tt.depositRequestRepository(),
				userTransactionRepository: tt.userTransactionRepository(),
				userAccountRepository:     tt.userAccountRepository(),
			}

			err := svc.CancelDeposit(tt.args.ctx, tt.args.depositRequest)
			assert.Equalf(t, tt.wantErr, err != nil, "CancelDeposit() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
