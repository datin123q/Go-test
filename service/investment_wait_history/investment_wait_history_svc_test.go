package investment_wait_history

import (
	"context"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
)

func Test_investmentWaitHistoryService_CreateStatusHistory(t *testing.T) {
	type args struct {
		ctx           context.Context
		comparer      domain.InvestmentWaitersComparer
		processMethod domain.USER_HISTORY_PROCESS_TYPE
		updateId      int64
	}
	tests := []struct {
		name                            string
		args                            args
		adminUserRepository             func() *mocks.AdminUserRepository
		investmentWaitHistoryRepository func() *mocks.InvestmentWaitHistoryRepository
		wantErr                         bool
	}{
		{
			name: "정상적으로 히스토리를 생성합니다.",
			args: args{
				ctx: context.Background(),
				comparer: domain.InvestmentWaitersComparer{
					Before: domain.InvestmentWaiters{
						Id:     1000,
						Status: domain.WAITING,
					},
					After: domain.InvestmentWaiters{
						Id:     1000,
						Status: domain.WAITING_END,
					},
				},
				processMethod: domain.USER_HISTORY_PROCESS_OPERATOR,
				updateId:      2000,
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetById", context.Background(), int64(2000)).Return(domain.AdminUser{
					Id:    2000,
					Name:  "고길동",
					Email: "gildong@email.com",
				}, nil)

				return repo
			},

			investmentWaitHistoryRepository: func() *mocks.InvestmentWaitHistoryRepository {
				investmentWaitHistoryRepository := &mocks.InvestmentWaitHistoryRepository{}
				investmentWaitHistoryRepository.On("Create", context.Background(), domain.InvestmentWaitHistory{
					InvestmentWaitId: 1000,
					SubstanceType:    string(domain.INVESTMENT_WAIT_HISTORY_STATUS),
					Name:             "고길동",
					Email:            "gildong@email.com",
					PrevStatus:       domain.WAITING,
					CurrStatus:       domain.WAITING_END,
					ProcessMethod:    domain.USER_HISTORY_PROCESS_OPERATOR,
				}).Return(nil)

				return investmentWaitHistoryRepository
			},

			wantErr: false,
		},
		{
			name: "변경전/후 상태값이 같은 경우에는 히스토리를 생성하지 않습니다.",
			args: args{
				ctx: context.Background(),
				comparer: domain.InvestmentWaitersComparer{
					Before: domain.InvestmentWaiters{
						Id:     1000,
						Status: domain.WAITING,
					},
					After: domain.InvestmentWaiters{
						Id:     1000,
						Status: domain.WAITING,
					},
				},
				processMethod: domain.USER_HISTORY_PROCESS_OPERATOR,
				updateId:      2000,
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetById", context.Background(), int64(2000)).Return(domain.AdminUser{
					Id:    2000,
					Name:  "고길동",
					Email: "gildong@email.com",
				}, nil)

				return repo
			},

			investmentWaitHistoryRepository: func() *mocks.InvestmentWaitHistoryRepository {
				return &mocks.InvestmentWaitHistoryRepository{}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := investmentWaitHistoryService{
				adminUserRepository:             tt.adminUserRepository(),
				investmentWaitHistoryRepository: tt.investmentWaitHistoryRepository(),
			}
			err := svc.CreateStatusHistory(tt.args.ctx, tt.args.comparer, tt.args.processMethod, tt.args.updateId)
			assert.Equalf(t, tt.wantErr, err != nil, "CreateStatusHistory() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
