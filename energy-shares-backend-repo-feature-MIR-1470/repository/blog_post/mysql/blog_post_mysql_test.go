package mysql

import (
	"testing"

	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_blogPostMysqlRepository_Create(t *testing.T) {

	type args struct {
		ctx      context.Context
		blogPost domain.BlogPost
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "블로그 포스트 등록 성공",
			args: args{
				ctx:      context.Background(),
				blogPost: mockCreateBlogPost,
			},
			want:    "f71c759d-5ff3-47d7-bfee-a034b2028040",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := blogPostMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := b.Create(tt.args.ctx, tt.args.blogPost)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Create() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_blogPostMysqlRepository_Exist(t *testing.T) {

	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "블로그 포스트 존재 확인 성공",
			args: args{
				ctx: context.Background(),
				id:  "c71c759d-5ff3-47d7-bfee-a034b2028040",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := blogPostMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := b.Exist(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exist() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_blogPostMysqlRepository_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "블로그 포스트 삭제 성공",
			args: args{
				ctx: context.Background(),
				id:  "c71c759d-5ff3-47d7-bfee-a034b2028040",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := blogPostMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			if err := b.Delete(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_blogPostMysqlRepository_GetAll(t *testing.T) {

	type args struct {
		ctx    context.Context
		offset int
		limit  int
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   []domain.BlogPost
		wantErr bool
	}{
		{
			name: "블로그 포스트 목록 조회 성공",
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  10,
			},
			want:    2,
			want1:   []domain.BlogPost{mockBlogPost1, mockBlogPost2},
			wantErr: false,
		},
		{
			name: "블로그 포스트 목록 조회 실패",
			args: args{
				ctx:    context.Background(),
				offset: 0,
				limit:  1,
			},
			want:    1,
			want1:   []domain.BlogPost{mockBlogPost1},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := blogPostMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, got1, err := b.GetAll(tt.args.ctx, tt.args.offset, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetAll() got = %v, want %v", got, tt.want)
			}
			opts := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			if cmp.Equal(got1, tt.want1, opts) == false {
				t.Errorf("GetAll() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_blogPostMysqlRepository_GetById(t *testing.T) {

	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		want1   domain.BlogPost
		wantErr bool
	}{
		{
			name: "블로그 포스트 조회 성공",
			args: args{
				ctx: context.Background(),
				id:  "c71c759d-5ff3-47d7-bfee-a034b2028040",
			},
			want:    true,
			want1:   mockBlogPost1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := blogPostMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, got1, err := b.GetById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetById() got = %v, want %v", got, tt.want)
			}
			cmpOpts := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			if !cmp.Equal(got1, tt.want1, cmpOpts) {
				t.Errorf("GetById() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_blogPostMysqlRepository_Update(t *testing.T) {

	type args struct {
		ctx context.Context
		p   domain.BlogPost
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "블로그 포스트 수정 성공",
			args: args{
				ctx: context.Background(),
				p:   func() domain.BlogPost { p := mockBlogPost1; p.Title = "title1-updated"; return p }(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := blogPostMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			if err := b.Update(tt.args.ctx, tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
