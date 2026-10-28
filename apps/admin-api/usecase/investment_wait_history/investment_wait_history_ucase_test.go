package investment_wait_history

import (
	"context"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"testing"

	"github.com/samber/lo"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"

	"github.com/stretchr/testify/mock"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"

	"github.com/illuminarean/energy-share-core-api/controllers"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
)

func Test_investmentWaitHistoryUsecase_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.InvestmentWaitHistoryGetAllRequest
	}

	tests := []struct {
		name                            string
		args                            args
		investmentWaitHistoryRepository func() *mocks.InvestmentWaitHistoryRepository
		want                            adomain.InvestmentWaitHistoryGetAllResponse
		wantErr                         bool
	}{
		{
			name: "投資待機履歴を照会します: SkipCount(0), MaxResultCount(2) => 正常にリストが照会されます。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitHistoryGetAllRequest{
					SearchInput: controllers.SearchInput{
						SkipCount:      0,
						MaxResultCount: 2,
					},
				},
			},

			investmentWaitHistoryRepository: func() *mocks.InvestmentWaitHistoryRepository {
				repo := new(mocks.InvestmentWaitHistoryRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.MatchedBy(func(field func(*cdomain.InvestmentWaitHistoryQueryOptionFields)) bool {
						opt := cdomain.InvestmentWaitHistoryQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.InvestmentWaitHistoryQueryOptionFields)) bool {
						opt := cdomain.InvestmentWaitHistoryQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 2
					}), // MaxResultCount
				).Return(int64(2), []domain.InvestmentWaitHistory{
					mockInvestmentWaitHistory1,
					mockInvestmentWaitHistory2,
				}, nil)

				return repo
			},
			want: adomain.InvestmentWaitHistoryGetAllResponse{
				TotalCount: 2,
				Items: []domain.InvestmentWaitHistory{
					mockInvestmentWaitHistory1,
					mockInvestmentWaitHistory2,
				},
			},
			wantErr: false,
		},
		{
			name: "待機者IDで投資待機履歴を照会します: InvestmentWaitId(1000) => 正常にリストが照会されます。",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestmentWaitHistoryGetAllRequest{
					InvestmentWaitId: lo.ToPtr(int64(1000)),
				},
			},

			investmentWaitHistoryRepository: func() *mocks.InvestmentWaitHistoryRepository {
				repo := new(mocks.InvestmentWaitHistoryRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.InvestmentWaitHistoryQueryOptionFields)) bool {
						opt := cdomain.InvestmentWaitHistoryQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.InvestmentWaitId()) == "1000"
					}), // InvestmentWaitId,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(int64(3), []domain.InvestmentWaitHistory{
					mockInvestmentWaitHistory1,
					mockInvestmentWaitHistory2,
					mockInvestmentWaitHistory3,
				}, nil)

				return repo
			},
			want: adomain.InvestmentWaitHistoryGetAllResponse{
				TotalCount: 3,
				Items: []domain.InvestmentWaitHistory{
					mockInvestmentWaitHistory1,
					mockInvestmentWaitHistory2,
					mockInvestmentWaitHistory3,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentWaitHistoryUsecase{
				investmentWaitHistoryRepository: tt.investmentWaitHistoryRepository(),
			}

			got, err := usecase.GetAll(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "investmentWaitHistoryUsecase.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "investmentWaitHistoryUsecase.GetAll() = %v, want %v", got, tt.want)
		})
	}
}
