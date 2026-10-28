package faq

import (
	"context"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_faqUsecase_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.FaqGetAllRequest
	}
	tests := []struct {
		name                 string
		args                 args
		faqRepository        func() *mocks.FaqRepository
		attachmentRepository func() *mocks.AttachmentRepository
		want                 udomain.FaqGetAllResponse
		wantErr              bool
	}{
		{
			name: "正常にFAQリストを照会する。",
			args: args{
				ctx: context.Background(),
				request: udomain.FaqGetAllRequest{
					CategoryType: lo.ToPtr(string(domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION)),
					SearchInput: cdomain.SearchInput{
						Sortby:         "title",
						Order:          "asc",
						SkipCount:      0,
						MaxResultCount: 10,
					},
				},
			},
			faqRepository: func() *mocks.FaqRepository {
				repo := new(mocks.FaqRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything, mock.Anything, mock.Anything,
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "pin,title"
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc,asc"
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}),
				).Return(int64(1), []domain.Faq{mockFaq1}, nil)
				return repo
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := new(mocks.AttachmentRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.AttachmentQueryOptionFields)) bool {
						opt := cdomain.AttachmentQueryOptionFunc()
						field(&opt)
						return *opt.ResourceId() == "bddad349-d0cc-4a1d-8843-431d54fcd2c7"
					}),
				).Return([]domain.Attachment{mockAttachment1}, nil)
				return repo
			},
			want: udomain.FaqGetAllResponse{
				TotalCount: 1,
				Items: []domain.Faq{
					mockFaq1.WithAttachments([]domain.Attachment{mockAttachment1}),
				},
			},
			wantErr: false,
		},
		{
			// MIR-1113: sortby/order未指定のとき "pin," / "desc," のように
			// 末尾に空要素が残る。pkg/orm.SortOrder は空要素を列指定なしの
			// プレースホルダとして読み飛ばすため、既定の pin desc がそのまま
			// 効くことを確認する。
			name: "sortby/orderが未指定でも既定のpin descで並ぶ。",
			args: args{
				ctx: context.Background(),
				request: udomain.FaqGetAllRequest{
					CategoryType: lo.ToPtr(string(domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION)),
					SearchInput: cdomain.SearchInput{
						Sortby:         "",
						Order:          "",
						SkipCount:      0,
						MaxResultCount: 10,
					},
				},
			},
			faqRepository: func() *mocks.FaqRepository {
				repo := new(mocks.FaqRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything, mock.Anything, mock.Anything,
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "pin,"
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc,"
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}),
				).Return(int64(1), []domain.Faq{mockFaq1}, nil)
				return repo
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := new(mocks.AttachmentRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.AttachmentQueryOptionFields)) bool {
						opt := cdomain.AttachmentQueryOptionFunc()
						field(&opt)
						return *opt.ResourceId() == "bddad349-d0cc-4a1d-8843-431d54fcd2c7"
					}),
				).Return([]domain.Attachment{mockAttachment1}, nil)
				return repo
			},
			want: udomain.FaqGetAllResponse{
				TotalCount: 1,
				Items: []domain.Faq{
					mockFaq1.WithAttachments([]domain.Attachment{mockAttachment1}),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := faqUsecase{
				faqRepository:        tt.faqRepository(),
				attachmentRepository: tt.attachmentRepository(),
			}

			got, err := f.GetAll(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetAll() = %v, want %v", got, tt.want)
		})
	}
}
