package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func Test_faqRepository_Create(t *testing.T) {
	type args struct {
		ctx context.Context
		faq domain.Faq
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "(성공) Faq 생성",
			args: args{
				ctx: context.Background(),
				faq: domain.Faq{
					CommonId:     "baedd3ff-6f36-4423-9fb8-0bd814d75969",
					Pin:          false,
					Title:        "FAQ 테스트",
					OwnerId:      100,
					UpdateId:     100,
					ContentHtml:  "<body>FAQ 테스트</body>",
					Posted:       true,
					CategoryType: domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION,
				},
			},
			want:    1003,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := faqRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.Create(tt.args.ctx, tt.args.faq)

			assert.Equalf(t, tt.wantErr, err != nil, "faqRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "faqRepository.Create() = %v, want %v", got, tt.want)
		})
	}

}

func Test_faqRepository_Update(t *testing.T) {
	type args struct {
		ctx context.Context
		faq domain.Faq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) Faq 수정",
			args: args{
				ctx: context.Background(),
				faq: domain.Faq{
					Id:           1000,
					CommonId:     "2000",
					Pin:          false,
					Title:        "FAQ 테스트 수정",
					OwnerId:      100,
					UpdateId:     100,
					ContentHtml:  "<body>FAQ 테스트 수정</body>",
					Posted:       true,
					CategoryType: domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := faqRepository{
				dbFactory: dbFactoryMock{},
			}
			err := repo.Update(tt.args.ctx, tt.args.faq)

			assert.Equalf(t, tt.wantErr, err != nil, "faqRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}

}

func Test_faqRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.FaqQueryOption
	}
	tests := []struct {
		name      string
		args      args
		want      []domain.Faq
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "(성공) Faq 전체 조회",
			args: args{
				ctx:     context.Background(),
				options: []cdomain.FaqQueryOption{},
			},
			want:      []domain.Faq{mockFaq1, mockFaq2, mockFaq3},
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name: "(성공) Faq '중요' Title 조회",
			args: args{
				ctx: context.Background(),
				options: []cdomain.FaqQueryOption{
					cdomain.FaqQueryOptionFunc().WithTitle(lo.ToPtr("중요")),
				},
			},
			want:      []domain.Faq{mockFaq1},
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "(성공) Faq CategoryType 조회 (FAQ_CATEGORY_FUNDS)",
			args: args{
				ctx: context.Background(),
				options: []cdomain.FaqQueryOption{
					cdomain.FaqQueryOptionFunc().WithCategoryType(lo.ToPtr(string(domain.FAQ_CATEGORY_FUNDS))),
				},
			},
			want:      []domain.Faq{mockFaq3},
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "(성공) Faq Posted 조회 (true)",
			args: args{
				ctx: context.Background(),
				options: []cdomain.FaqQueryOption{
					cdomain.FaqQueryOptionFunc().WithPosted(lo.ToPtr(true)),
				},
			},
			want:      []domain.Faq{mockFaq1},
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "(성공) Faq Pin 조회 (true)",
			args: args{
				ctx: context.Background(),
				options: []cdomain.FaqQueryOption{
					cdomain.FaqQueryOptionFunc().WithPin(lo.ToPtr(true)),
				},
			},
			want:      []domain.Faq{mockFaq3},
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "(성공) 특정 날짜안에 있는 Faq 조회 (2024-01-01 ~ 2024-02-01)",
			args: args{
				ctx: context.Background(),
				options: []cdomain.FaqQueryOption{
					cdomain.FaqQueryOptionFunc().WithStartDate(lo.ToPtr("2024-01-01")),
					cdomain.FaqQueryOptionFunc().WithEndDate(lo.ToPtr("2024-02-01")),
				},
			},
			want:      nil,
			wantTotal: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := faqRepository{
				dbFactory: dbFactoryMock{},
			}
			count, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			ignoreType := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.NoErrorf(t, err, "faqRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.wantTotal, count, "faqRepository.GetAll() total = %v, want %v", count, tt.wantTotal)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreType), "faqRepository.GetAll() = %v, want %v", got, tt.want)
		})
	}
}

func Test_faqRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.Faq
		wantErr bool
	}{
		{
			name: "(성공) Faq ID 1000 조회",
			args: args{
				ctx: context.Background(),
				id:  1000,
			},
			want:    mockFaq1,
			wantErr: false,
		},
		{
			name: "(실패) 존재하지 않는 Faq ID 1003 조회",
			args: args{
				ctx: context.Background(),
				id:  1003,
			},
			want:    domain.Faq{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := faqRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetById(tt.args.ctx, tt.args.id)

			ignoreType := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "faqRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreType), "faqRepository.GetById() = %v, want %v", got, tt.want)
		})
	}
}

func Test_faqRepository_Delete(t *testing.T) {
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
			name: "(성공) Faq ID 1000 삭제",
			args: args{
				ctx: context.Background(),
				id:  1000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := faqRepository{
				dbFactory: dbFactoryMock{},
			}
			err := repo.Delete(tt.args.ctx, tt.args.id)

			assert.Equalf(t, tt.wantErr, err != nil, "faqRepository.Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
