package banner

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

func Test_bannerUsecase_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.BannerGetAllRequest
	}

	tests := []struct {
		name                    string
		args                    args
		bannerRepository        func() cdomain.BannerRepository
		bannerSettingRepository func() cdomain.BannerSettingRepository
		want                    udomain.BannerGetAllResponse
		wantErr                 bool
	}{
		{
			name: "(成功) 掲載中バナー sort_order順で全体照会",
			args: args{
				ctx: context.Background(),
			},
			bannerRepository: func() cdomain.BannerRepository {
				repo := &mocks.BannerRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.BannerQueryOptionFields)) bool {
						opt := cdomain.BannerQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "sort_order"
					}), // SortBy
					mock.MatchedBy(func(field func(*cdomain.BannerQueryOptionFields)) bool {
						opt := cdomain.BannerQueryOptionFunc()
						field(&opt)
						return opt.Order() == "asc"
					}), // SortOrder
					mock.MatchedBy(func(field func(*cdomain.BannerQueryOptionFields)) bool {
						opt := cdomain.BannerQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.BannerQueryOptionFields)) bool {
						opt := cdomain.BannerQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}), // MaxResultCount
					mock.MatchedBy(func(field func(*cdomain.BannerQueryOptionFields)) bool {
						opt := cdomain.BannerQueryOptionFunc()
						field(&opt)
						return *opt.IsPosted() == true
					}), // IsPosted
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(int64(3), []domain.Banner{
					mockBanner1,
					mockBanner2,
					mockBanner3,
				}, nil)
				return repo
			},
			bannerSettingRepository: func() cdomain.BannerSettingRepository {
				repo := &mocks.BannerSettingRepository{}
				repo.On("GetByKey", context.Background(), domain.ROLLING_SECOND).Return(mockRollingSecond, nil)
				return repo
			},
			want: udomain.BannerGetAllResponse{
				TotalCount:    int64(3),
				RollingSecond: int64(5),
				Items: []udomain.BannerGetAllItem{
					{
						Id:        mockBanner1.Id,
						Title:     mockBanner1.Title,
						IsPosted:  mockBanner1.IsPosted,
						SortOrder: lo.FromPtr(mockBanner1.SortOrder),
						Url:       mockBanner1.Url,
					},
					{
						Id:        mockBanner2.Id,
						Title:     mockBanner2.Title,
						IsPosted:  mockBanner2.IsPosted,
						SortOrder: lo.FromPtr(mockBanner2.SortOrder),
						Url:       mockBanner2.Url,
					},
					{
						Id:        mockBanner3.Id,
						Title:     mockBanner3.Title,
						IsPosted:  mockBanner3.IsPosted,
						SortOrder: lo.FromPtr(mockBanner3.SortOrder),
						Url:       mockBanner3.Url,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bannerUsecase{
				bannerRepository:        tt.bannerRepository(),
				bannerSettingRepository: tt.bannerSettingRepository(),
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := b.GetAll(tt.args.ctx)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAllOnDate() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "GetAll() = %+v, want %+v", got, tt.want)
		})
	}
}
