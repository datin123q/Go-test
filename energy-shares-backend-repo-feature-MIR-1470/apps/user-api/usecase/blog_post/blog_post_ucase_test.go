package blog_post

import (
	"testing"

	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

func Test_blogPostUsecase_GetAll(t *testing.T) {
	type fields struct {
		blogPostRepo cdomain.BlogPostRepository
	}
	type args struct {
		c       context.Context
		request udomain.BlogPostGetAllRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		want1   udomain.BlogPostGetAllResponse
		wantErr bool
	}{
		{
			name: "ブログ_ポスト_リスト_照会_成功",
			fields: fields{
				blogPostRepo: func() *mocks.BlogPostRepository {
					repo := &mocks.BlogPostRepository{}
					repo.On("GetAll", context.Background(), 0, 6).
						Return(int64(2), []domain.BlogPost{
							mockBlogPost1,
							mockBlogPost2,
						}, nil)
					return repo
				}(),
			},
			args: args{
				c: context.Background(),
				request: udomain.BlogPostGetAllRequest{
					Offset: 0,
					Limit:  6,
				},
			},
			want: 2,
			want1: udomain.BlogPostGetAllResponse{
				TotalCount: 2,
				Items: []udomain.BlogPostResponse{
					mockBlogPostResponse1,
					mockBlogPostResponse2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := blogPostUsecase{
				blogPostRepo: tt.fields.blogPostRepo,
			}
			got, got1, err := u.GetAll(tt.args.c, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetAll() got = %v, want %v", got, tt.want)
			}
			cmpOpts := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			if !cmp.Equal(got1, tt.want1, cmpOpts) {
				t.Errorf("GetAll() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
