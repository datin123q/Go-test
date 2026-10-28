package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"
)

func Test_contentImageMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx          context.Context
		contentImage domain.ContentImage
	}

	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "ContentImage 생성 => 생성되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				contentImage: domain.ContentImage{
					FileName:      "test.jpg",
					OriginalName:  "test.jpg",
					OriginalUrl:   "http://test.com/test.jpg",
					FileSize:      12737123,
					FileExtension: "jpg",
					FilePath:      "/test/test.jpg",
					FileUrl:       "http://test.com/test.jpg",
					ResourceId:    "e3ae2c14-b617-4870-b057-a8ae915cdee6",
				},
			},
			want:    1003,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := contentImageMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.Create(tt.args.ctx, tt.args.contentImage)

			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "Create() = %v, want %v", got, tt.want)
		})
	}

}

func Test_contentImageMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.ContentImageQueryOption
	}

	tests := []struct {
		name    string
		args    args
		want    []domain.ContentImage
		wantErr bool
	}{
		{
			name: "ContentImage 전체조회 => 전체 조회되기를 기대합니다.",
			args: args{
				ctx:     context.Background(),
				options: []cdomain.ContentImageQueryOption{},
			},
			want: []domain.ContentImage{
				mockContentImage1, mockContentImage2, mockContentImage3,
			},
		},
		{
			name: "ContentImage 리소스 ID로 필터링 조회 : resourceId : e3ae2c14-b617-4870-b057-a8ae915cdee6 => 2건이 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.ContentImageQueryOption{
					cdomain.ContentImageQueryOptionFunc().WithResourceId("e3ae2c14-b617-4870-b057-a8ae915cdee6"),
				},
			},
			want: []domain.ContentImage{
				mockContentImage2, mockContentImage3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := contentImageMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "GetAll() = %v, want %v", got, tt.want)
		})
	}
}

func Test_contentImageMysqlRepository_DeleteAllByResourceId(t *testing.T) {
	type args struct {
		ctx        context.Context
		resourceId string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ContentImage 삭제 => resourceId 기반으로 삭제되기를 기대합니다.",
			args: args{
				ctx:        context.Background(),
				resourceId: mockContentImage1.ResourceId,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := contentImageMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			err := repo.DeleteByResourceId(tt.args.ctx, tt.args.resourceId)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_contentImageMysqlRepository_Delete(t *testing.T) {
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
			name: "ContentImage 삭제 => ID 기반으로 삭제되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  mockContentImage1.Id,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := contentImageMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			err := repo.Delete(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
