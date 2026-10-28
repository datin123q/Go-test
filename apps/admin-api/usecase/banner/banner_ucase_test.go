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
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

func Test_bannerUsecase_Create(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.BannerCreateRequest
	}

	tests := []struct {
		name             string
		args             args
		bannerRepository func() cdomain.BannerRepository
		want             adomain.BannerCreateResponse
		wantErr          bool
	}{
		{
			name: "(成功) バナー作成",
			args: args{
				ctx: context.Background(),
				request: adomain.BannerCreateRequest{
					Title:       mockBannerCreate.Title,
					IsPosted:    lo.ToPtr(mockBannerCreate.IsPosted),
					StartDate:   "2024-01-01 00:00:00",
					EndDate:     "2024-01-01 23:59:59",
					Url:         mockBannerCreate.Url,
					SortOrder:   mockBannerCreate.SortOrder,
					Description: nil,
					NewTab:      lo.ToPtr(mockBannerCreate.NewTab),
				},
			},
			bannerRepository: func() cdomain.BannerRepository {
				repo := &mocks.BannerRepository{}
				mockBannerCreate.Id = 0
				repo.On("Create", context.Background(),
					mock.MatchedBy(func(b domain.Banner) bool {
						return b.Title == mockBannerCreate.Title &&
							b.IsPosted == mockBannerCreate.IsPosted &&
							b.StartDate.ToString() == mockBannerCreate.StartDate.ToString() &&
							b.EndDate.ToString() == mockBannerCreate.EndDate.ToString() &&
							b.Url == mockBannerCreate.Url &&
							lo.FromPtr(b.SortOrder) == lo.FromPtr(mockBannerCreate.SortOrder) &&
							b.Description == mockBannerCreate.Description &&
							b.NewTab == mockBannerCreate.NewTab
					}),
				).Return(int64(1), nil)
				return repo
			},
			want: adomain.BannerCreateResponse{
				Id:          mockBannerCreate.Id,
				Title:       mockBannerCreate.Title,
				IsPosted:    mockBannerCreate.IsPosted,
				StartDate:   mockBannerCreate.StartDate,
				EndDate:     mockBannerCreate.EndDate,
				Url:         mockBannerCreate.Url,
				SortOrder:   mockBannerCreate.SortOrder,
				Description: mockBannerCreate.Description,
				NewTab:      mockBannerCreate.NewTab,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bannerUsecase{
				bannerRepository: tt.bannerRepository(),
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := b.Create(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("bannerUsecase.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "Create() = %+v, want %+v", got, tt.want)
			assert.Equal(t, tt.want.StartDate.ToString(), got.StartDate.ToString())
			assert.Equal(t, tt.want.EndDate.ToString(), got.EndDate.ToString())
		})
	}
}

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
		want                    adomain.BannerGetAllResponse
		wantErr                 bool
	}{
		{
			name: "(成功) バナー全体照会",
			args: args{
				ctx: context.Background(),
				request: adomain.BannerGetAllRequest{
					JWTUser: jwthelper.User{},
					BannerGetAllQueryParam: adomain.BannerGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							SkipCount:      0,
							MaxResultCount: 10,
							Order:          "asc",
							Sortby:         "id",
						},
						IsPosted: lo.ToPtr(true),
					},
				},
			},
			bannerRepository: func() cdomain.BannerRepository {
				repo := &mocks.BannerRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.BannerQueryOptionFields)) bool {
						opt := cdomain.BannerQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
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
					mock.Anything,
					mock.Anything,
					mock.Anything,
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
			want: adomain.BannerGetAllResponse{
				TotalCount:        int64(3),
				TotalBannerPosted: int64(3),
				RollingSecond:     int64(5),
				Items: []adomain.BannerGetAllItem{
					{
						Id:        mockBanner1.Id,
						Title:     mockBanner1.Title,
						IsPosted:  mockBanner1.IsPosted,
						StartDate: mockBanner1.StartDate,
						EndDate:   mockBanner1.EndDate,
						SortOrder: mockBanner1.SortOrder,
					},
					{
						Id:        mockBanner2.Id,
						Title:     mockBanner2.Title,
						IsPosted:  mockBanner2.IsPosted,
						StartDate: mockBanner2.StartDate,
						EndDate:   mockBanner2.EndDate,
						SortOrder: mockBanner2.SortOrder,
					},
					{
						Id:        mockBanner3.Id,
						Title:     mockBanner3.Title,
						IsPosted:  mockBanner3.IsPosted,
						StartDate: mockBanner3.StartDate,
						EndDate:   mockBanner3.EndDate,
						SortOrder: mockBanner3.SortOrder,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(成功) バナー掲示日基準照会",
			args: args{
				ctx: context.Background(),
				request: adomain.BannerGetAllRequest{
					JWTUser: jwthelper.User{},
					BannerGetAllQueryParam: adomain.BannerGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							SkipCount:      0,
							MaxResultCount: 10,
							Order:          "asc",
							Sortby:         "id",
						},
						StartDate: lo.ToPtr("2024-01-02 00:00:00"),
						EndDate:   lo.ToPtr("2024-01-05 23:59:59"),
						IsPosted:  lo.ToPtr(true),
					},
				},
			},
			bannerRepository: func() cdomain.BannerRepository {
				repo := &mocks.BannerRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.BannerQueryOptionFields)) bool {
						opt := cdomain.BannerQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}), // SortBy
					mock.MatchedBy(func(field func(*cdomain.BannerQueryOptionFields)) bool {
						opt := cdomain.BannerQueryOptionFunc()
						field(&opt)
						return opt.Order() == "asc"
					}), // Order
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
					mock.Anything, // Title
					mock.MatchedBy(func(field func(*cdomain.BannerQueryOptionFields)) bool {
						opt := cdomain.BannerQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.IsPosted()) == true
					}), // IsPosted
					mock.MatchedBy(func(field func(*cdomain.BannerQueryOptionFields)) bool {
						opt := cdomain.BannerQueryOptionFunc()
						field(&opt)
						return opt.GteEndDate() == "2024-01-02 00:00:00"
					}), // StartDate
					mock.MatchedBy(func(field func(*cdomain.BannerQueryOptionFields)) bool {
						opt := cdomain.BannerQueryOptionFunc()
						field(&opt)
						return opt.LteStartDate() == "2024-01-05 23:59:59"
					}), // EndDate
					mock.Anything, // CreatedStartDate
					mock.Anything, // CreatedEndDate
				).Return(int64(2), []domain.Banner{
					mockBanner1,
					mockBanner2,
				}, nil)
				return repo
			},
			bannerSettingRepository: func() cdomain.BannerSettingRepository {
				repo := &mocks.BannerSettingRepository{}
				repo.On("GetByKey", context.Background(), domain.ROLLING_SECOND).Return(mockRollingSecond, nil)
				return repo
			},
			want: adomain.BannerGetAllResponse{
				TotalCount:        int64(2),
				TotalBannerPosted: int64(2),
				RollingSecond:     int64(5),
				Items: []adomain.BannerGetAllItem{
					{
						Id:        mockBanner1.Id,
						Title:     mockBanner1.Title,
						IsPosted:  mockBanner1.IsPosted,
						StartDate: mockBanner1.StartDate,
						EndDate:   mockBanner1.EndDate,
						SortOrder: mockBanner1.SortOrder,
					},
					{
						Id:        mockBanner2.Id,
						Title:     mockBanner2.Title,
						IsPosted:  mockBanner2.IsPosted,
						StartDate: mockBanner2.StartDate,
						EndDate:   mockBanner2.EndDate,
						SortOrder: mockBanner2.SortOrder,
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

			got, err := b.GetAll(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("bannerUsecase.GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "GetAll() = %+v, want %+v", got, tt.want)
		})
	}
}

func Test_bannerUsecase_GetOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.BannerGetOneRequest
	}

	tests := []struct {
		name             string
		args             args
		bannerRepository func() cdomain.BannerRepository
		want             adomain.BannerGetOneResponse
		wantErr          bool
	}{
		{
			name: "(成功) バナー単件照会",
			args: args{
				ctx: context.Background(),
				request: adomain.BannerGetOneRequest{
					Id: 1,
				},
			},
			bannerRepository: func() cdomain.BannerRepository {
				repo := &mocks.BannerRepository{}
				repo.On("GetById", context.Background(), int64(1)).Return(mockBanner1, nil)
				return repo
			},
			want: adomain.BannerGetOneResponse{
				Id:          mockBanner1.Id,
				Title:       mockBanner1.Title,
				IsPosted:    mockBanner1.IsPosted,
				StartDate:   mockBanner1.StartDate,
				EndDate:     mockBanner1.EndDate,
				Url:         mockBanner1.Url,
				SortOrder:   mockBanner1.SortOrder,
				Description: mockBanner1.Description,
				Created:     mockBanner1.Created,
				Updated:     mockBanner1.Updated,
				NewTab:      mockBanner1.NewTab,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bannerUsecase{
				bannerRepository: tt.bannerRepository(),
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := b.GetOne(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("bannerUsecase.GetOne() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "GetOne() = %+v, want %+v", got, tt.want)
		})
	}
}

func Test_bannerUsecase_Update(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.BannerUpdateRequest
	}

	tests := []struct {
		name             string
		args             args
		bannerRepository func() cdomain.BannerRepository
		want             adomain.BannerUpdateResponse
		wantErr          bool
	}{
		{
			name: "(成功) バナー修正",
			args: args{
				ctx: context.Background(),
				request: adomain.BannerUpdateRequest{
					Id:          int64(1),
					Title:       mockBannerUpdateAfter.Title,
					IsPosted:    lo.ToPtr(mockBannerUpdateAfter.IsPosted),
					StartDate:   "2025-12-01 00:00:00",
					EndDate:     "2025-12-31 23:59:59",
					Url:         mockBannerUpdateAfter.Url,
					SortOrder:   mockBannerUpdateAfter.SortOrder,
					Description: lo.ToPtr(mockBannerUpdateAfter.Description),
					NewTab:      lo.ToPtr(mockBannerUpdateAfter.NewTab),
				},
			},
			bannerRepository: func() cdomain.BannerRepository {
				repo := &mocks.BannerRepository{}
				repo.On("GetById", context.Background(), int64(1)).Return(mockBannerUpdateBefore, nil)
				repo.On("Update", context.Background(), mock.MatchedBy(func(b domain.Banner) bool {
					return b.Id == mockBannerUpdateAfter.Id &&
						b.Title == mockBannerUpdateAfter.Title &&
						b.IsPosted == mockBannerUpdateAfter.IsPosted &&
						b.StartDate.ToString() == mockBannerUpdateAfter.StartDate.ToString() &&
						b.EndDate.ToString() == mockBannerUpdateAfter.EndDate.ToString() &&
						b.Url == mockBannerUpdateAfter.Url &&
						lo.FromPtr(b.SortOrder) == lo.FromPtr(mockBannerUpdateAfter.SortOrder) &&
						b.Description == mockBannerUpdateAfter.Description &&
						b.NewTab == mockBannerUpdateAfter.NewTab
				})).Return(nil)
				return repo
			},
			want: adomain.BannerUpdateResponse{
				Id:          int64(1),
				Title:       mockBannerUpdateAfter.Title,
				IsPosted:    mockBannerUpdateAfter.IsPosted,
				StartDate:   mockBannerUpdateAfter.StartDate,
				EndDate:     mockBannerUpdateAfter.EndDate,
				Url:         mockBannerUpdateAfter.Url,
				SortOrder:   mockBannerUpdateAfter.SortOrder,
				Description: mockBannerUpdateAfter.Description,
				NewTab:      mockBannerUpdateAfter.NewTab,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bannerUsecase{
				bannerRepository: tt.bannerRepository(),
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := b.Update(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("bannerUsecase.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "Update() = %+v, want %+v", got, tt.want)
			assert.Equal(t, tt.want.StartDate.ToString(), got.StartDate.ToString())
			assert.Equal(t, tt.want.EndDate.ToString(), got.EndDate.ToString())
		})
	}
}

func Test_bannerUsecase_Delete(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.BannerDeleteRequest
	}

	tests := []struct {
		name             string
		args             args
		bannerRepository func() cdomain.BannerRepository
		wantErr          bool
	}{
		{
			name: "(成功) バナー削除",
			args: args{
				ctx: context.Background(),
				request: adomain.BannerDeleteRequest{
					Id: 1,
				},
			},
			bannerRepository: func() cdomain.BannerRepository {
				repo := &mocks.BannerRepository{}
				repo.On("Delete", context.Background(), int64(1)).Return(nil)
				return repo
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bannerUsecase{
				bannerRepository: tt.bannerRepository(),
			}

			err := b.Delete(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("bannerUsecase.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_bannerUsecase_UpdateRollingTime(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.BannerUpdateRollingTimeRequest
	}

	tests := []struct {
		name                    string
		args                    args
		bannerSettingRepository func() cdomain.BannerSettingRepository
		wantErr                 bool
	}{
		{
			name: "(成功) バナーローリング時間修正",
			args: args{
				ctx: context.Background(),
				request: adomain.BannerUpdateRollingTimeRequest{
					JWTUser:       jwthelper.User{},
					RollingSecond: 1,
				},
			},
			bannerSettingRepository: func() cdomain.BannerSettingRepository {
				repo := &mocks.BannerSettingRepository{}
				repo.On("GetByKey", context.Background(), domain.ROLLING_SECOND).Return(mockRollingSecond, nil)
				mockRollingSecond.SettingValue = "1"
				repo.On("Update", context.Background(), mockRollingSecond).Return(nil)
				return repo
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bannerUsecase{
				bannerSettingRepository: tt.bannerSettingRepository(),
			}

			err := b.UpdateRollingTime(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("bannerUsecase.UpdateRollingTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
