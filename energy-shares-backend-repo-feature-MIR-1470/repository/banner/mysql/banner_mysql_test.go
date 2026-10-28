package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_bannerRepository_Create(t *testing.T) {
	type args struct {
		ctx    context.Context
		banner domain.Banner
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "배너 생성",
			args: args{
				ctx: context.Background(),
				banner: domain.Banner{
					Title:       "test",
					Description: "test",
					StartDate:   jsonutils.JSONTime(time.Now()),
					EndDate:     jsonutils.JSONTime(time.Now().AddDate(0, 0, 1)),
					IsPosted:    false,
					SortOrder:   lo.ToPtr(int64(3)),
					NewTab:      false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := bannerRepository{
				dbFactory: dbFactoryMock{},
			}

			_, err := repo.Create(tt.args.ctx, tt.args.banner)
			if (err != nil) != tt.wantErr {
				t.Errorf("BannerRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_bannerRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.BannerQueryOption
	}

	tests := []struct {
		name    string
		args    args
		want    []domain.Banner
		wantErr bool
	}{
		{
			name: "배너 목록 조회 - 조건 없음",
			args: args{
				ctx: context.Background(),
				options: []cdomain.BannerQueryOption{
					cdomain.BannerQueryOptionFunc().WithSortBy("id"),
					cdomain.BannerQueryOptionFunc().WithOrder("asc"),
					cdomain.BannerQueryOptionFunc().WithLimit(10),
					cdomain.BannerQueryOptionFunc().WithOffset(0),
				},
			},
			want: []domain.Banner{
				mockBanner1, mockBanner2, mockBanner3, mockBanner4,
			},
			wantErr: false,
		},
		{
			name: "배너 목록 조회 - 페이징 처리 확인",
			args: args{
				ctx: context.Background(),
				options: []cdomain.BannerQueryOption{
					cdomain.BannerQueryOptionFunc().WithSortBy("id"),
					cdomain.BannerQueryOptionFunc().WithOrder("asc"),
					cdomain.BannerQueryOptionFunc().WithLimit(2),
					cdomain.BannerQueryOptionFunc().WithOffset(2),
				},
			},
			want: []domain.Banner{
				mockBanner3, mockBanner4,
			},
			wantErr: false,
		},
		{
			name: "배너 목록 조회 - startDate 있음, endDate 없음",
			args: args{
				ctx: context.Background(),
				options: []cdomain.BannerQueryOption{
					cdomain.BannerQueryOptionFunc().WithSortBy("id"),
					cdomain.BannerQueryOptionFunc().WithOrder("asc"),
					cdomain.BannerQueryOptionFunc().WithLimit(10),
					cdomain.BannerQueryOptionFunc().WithOffset(0),
					cdomain.BannerQueryOptionFunc().WithGTEStartDate(time.Now().AddDate(0, 0, -1).Format(time.DateTime)),
				},
			},
			want: []domain.Banner{
				mockBanner2, mockBanner3, mockBanner4,
			},
			wantErr: false,
		},
		{
			name: "배너 목록 조회 - 게시일 날짜 검색 (1건 기대)",
			args: args{
				ctx: context.Background(),
				options: []cdomain.BannerQueryOption{
					cdomain.BannerQueryOptionFunc().WithIsPosted(lo.ToPtr(true)),
					cdomain.BannerQueryOptionFunc().WithGTEStartDate(time.Now().AddDate(0, 0, -2).Format(time.DateTime)),
					cdomain.BannerQueryOptionFunc().WithLTEEndDate(time.Now().AddDate(0, 0, 2).Format(time.DateTime)),
					cdomain.BannerQueryOptionFunc().WithSortBy("id"),
					cdomain.BannerQueryOptionFunc().WithOrder("asc"),
					cdomain.BannerQueryOptionFunc().WithLimit(10),
					cdomain.BannerQueryOptionFunc().WithOffset(0),
				},
			},
			want:    []domain.Banner{mockBanner2},
			wantErr: false,
		},
		{
			name: "배너 목록 조회 - 생성일 날짜 검색 (2건 기대)",
			args: args{
				ctx: context.Background(),
				options: []cdomain.BannerQueryOption{
					cdomain.BannerQueryOptionFunc().WithIsPosted(lo.ToPtr(true)),
					cdomain.BannerQueryOptionFunc().WithCreatedStartDate(time.Now().AddDate(0, 0, -20).Format(time.DateTime)),
					cdomain.BannerQueryOptionFunc().WithCreatedEndDate(time.Now().AddDate(0, 0, -10).Format(time.DateTime)),
					cdomain.BannerQueryOptionFunc().WithSortBy("id"),
					cdomain.BannerQueryOptionFunc().WithOrder("asc"),
					cdomain.BannerQueryOptionFunc().WithLimit(10),
					cdomain.BannerQueryOptionFunc().WithOffset(0),
				},
			},
			want:    []domain.Banner{mockBanner1, mockBanner2},
			wantErr: false,
		},
		{
			name: "배너 목록 조회 - 제목 검색 (2건 기대)",
			args: args{
				ctx: context.Background(),
				options: []cdomain.BannerQueryOption{
					cdomain.BannerQueryOptionFunc().WithTitle("test"),
					cdomain.BannerQueryOptionFunc().WithSortBy("id"),
					cdomain.BannerQueryOptionFunc().WithOrder("asc"),
					cdomain.BannerQueryOptionFunc().WithLimit(10),
					cdomain.BannerQueryOptionFunc().WithOffset(0),
				},
			},
			want:    []domain.Banner{mockBanner1, mockBanner2},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := bannerRepository{
				dbFactory: dbFactoryMock{},
			}

			_, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("BannerRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i := range got {
				assert.Equalf(t, true, cmp.Equal(tt.want[i].Id, got[i].Id), "GetAll() diff = %v", cmp.Diff(tt.want[i].Id, got[i].Id))
			}
		})
	}
}

func Test_bannerRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}

	tests := []struct {
		name    string
		args    args
		want    domain.Banner
		wantErr bool
	}{
		{
			name: "배너 조회",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			want:    mockBanner1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := bannerRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := repo.GetById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("BannerRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "BannerRepository.GetById() diff = %v",
				cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}

func Test_bannerRepository_Update(t *testing.T) {
	type args struct {
		ctx    context.Context
		banner domain.Banner
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "배너 수정",
			args: args{
				ctx: context.Background(),
				banner: domain.Banner{
					Id:          3,
					Title:       "test",
					Description: "test",
					StartDate:   jsonutils.JSONTime(time.Now().AddDate(0, 0, 5)),
					EndDate:     jsonutils.JSONTime(time.Now().AddDate(0, 0, 10)),
					IsPosted:    true,
					SortOrder:   lo.ToPtr(int64(4)),
					NewTab:      false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := bannerRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.banner)
			if (err != nil) != tt.wantErr {
				t.Errorf("BannerRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_bannerRepository_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "배너 삭제",
			args: args{
				ctx: context.Background(),
				id:  4,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := bannerRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Delete(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("BannerRepository.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
