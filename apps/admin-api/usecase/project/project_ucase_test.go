package project

import (
	"context"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_projectUsecase_GetStatistics(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.ProjectGetStatisticsRequest
	}
	tests := []struct {
		name              string
		args              args
		projectRepository func() cdomain.ProjectRepository
		shareRepository   func() cdomain.ShareRepository
		want              adomain.ProjectGetStatisticsResponse
		wantErr           bool
	}{
		{
			name: "特定プロジェクトの統計データ照会",
			args: args{
				ctx: context.Background(),
				request: adomain.ProjectGetStatisticsRequest{
					ProjectId: 146,
				},
			},
			projectRepository: func() cdomain.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("GetById", context.Background(), int64(146)).Return(mockProject1, nil)
				return repo
			},

			shareRepository: func() cdomain.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
						opt := cdomain.ShareQueryOptionFunc()
						field(&opt)
						return *opt.ProjectId() == "146"
					}),
				).
					Return([]domain.Share{mockShare1, mockShare2, mockShare3, mockShare4}, nil)
				return repo
			},
			want: adomain.ProjectGetStatisticsResponse{
				TotalShareCount:        4,
				TargetShareAmount:      100_000,
				TargetShareQty:         10,
				SuccessShareAmount:     100_000,
				RequestShareCount:      0,
				SuccessShareQty:        10,
				RemainingShareQty:      0,
				ReinvestCount:          3,
				SuccessShareCount:      4,
				CancellationShareCount: 0,
				SuccessShareAvgQty:     2.5,
			},
			wantErr: false,
		},
		{
			name: "投資件数がない場合に正しく動作するか確認",
			args: args{
				ctx: context.Background(),
				request: adomain.ProjectGetStatisticsRequest{
					ProjectId: 146,
				},
			},
			projectRepository: func() cdomain.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("GetById", context.Background(), int64(146)).Return(mockProject1, nil)
				return repo
			},

			shareRepository: func() cdomain.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ShareQueryOptionFields)) bool {
						opt := cdomain.ShareQueryOptionFunc()
						field(&opt)
						return *opt.ProjectId() == "146"
					}),
				).
					Return([]domain.Share{}, nil)
				return repo
			},
			want: adomain.ProjectGetStatisticsResponse{
				RemainingShareQty: 10,
				TargetShareAmount: 100_000,
				TargetShareQty:    10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := projectUsecase{
				shareRepository:   tt.shareRepository(),
				projectRepository: tt.projectRepository(),
			}
			got, err := p.GetStatistics(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "projectUsecase.GetStatistics() error = %v", err)
			assert.Equalf(t, tt.want, got, "projectUsecase.GetStatistics() = %v, want %v", got, tt.want)
		})
	}
}
