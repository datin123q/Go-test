package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_NewsRepository_Create(t *testing.T) {
	type args struct {
		ctx  context.Context
		news domain.News
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) News 생성",
			args: args{
				ctx:  context.Background(),
				news: mockNewsToCreate,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newsRepository{
				dbFactory: dbFactoryMock{},
			}
			_, err := repo.Create(tt.args.ctx, tt.args.news)

			assert.Equalf(t, tt.wantErr, err != nil, "newsRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_NewsRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.NewsQueryOption
	}
	tests := []struct {
		name      string
		args      args
		want      domain.News
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "News 모델 조회 -> 뉴스 둘을 모두 가져오기를 기대합니다.",
			args: args{
				ctx: context.Background(),
			},
			want:      mockNews1,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name: "News 모델 조회 -> Title1 인 뉴스 하나만 가져오기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.NewsQueryOption{
					cdomain.NewsQueryOptionFields{}.WithTitle("Title1"),
				},
			},
			want:      mockNews1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "News 모델 조회 -> Posted 인 뉴스 하나만 가져오기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.NewsQueryOption{
					cdomain.NewsQueryOptionFields{}.WithPosted(lo.ToPtr(true)),
				},
			},
			want:      mockNews1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "News 모델 조회 -> Posted 인 뉴스 하나만 가져오기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.NewsQueryOption{
					cdomain.NewsQueryOptionFields{}.WithPosted(lo.ToPtr(true)),
					cdomain.NewsQueryOptionFields{}.WithTitle(""),
				},
			},
			want:      mockNews1,
			wantTotal: 1,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newsRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			total, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equal(t, tt.wantTotal, total)
			assert.Equalf(t, tt.wantErr, err != nil,
				"newsRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, "", cmp.Diff(tt.want, got[0], ignoreTime),
				"myNumberRepository.GetAll() got = %v, want %v", got[0], tt.want)
		})
	}
}

func Test_NewsRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.News
		wantErr bool
	}{
		{
			name: "News 모델 조회 (id: 2000) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  mockNews1.Id,
			},
			want:    mockNews1,
			wantErr: false,
		},
		{
			name: "News 모델 조회 (User:2001) -> 존재하지 않는 User를 조회하므로 에러가 발생합니다.",
			args: args{
				ctx: context.Background(),
				id:  2001,
			},
			want:    domain.News{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newsRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			got, err := repo.GetById(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.wantErr, err != nil, "newsRepository.GetById() error = %v, wantErr %v",
				err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime),
				"myNumberRepository.GetByUserId() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_NewsRepository_Delete(t *testing.T) {
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
			name: "News 모델 삭제 (id: 2000) -> 정상적으로 식제되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  mockNews1.Id,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newsRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Delete(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.wantErr, err != nil, "newsRepository.Delete() error = %v, wantErr %v",
				err, tt.wantErr)
		})
	}
}
