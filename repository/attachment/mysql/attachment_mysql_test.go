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

func Test_attachmentMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx        context.Context
		attachment domain.Attachment
	}

	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "Attachment 생성 => 생성되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				attachment: domain.Attachment{
					FileName:      "test.jpg",
					OriginalName:  "test.jpg",
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
			repo := attachmentMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.Create(tt.args.ctx, tt.args.attachment)

			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "Create() = %v, want %v", got, tt.want)
		})
	}
}

func Test_attachmentMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.AttachmentQueryOption
	}

	tests := []struct {
		name    string
		args    args
		want    []domain.Attachment
		wantErr bool
	}{
		{
			name: "Attachment 전체조회 => 전체 조회되기를 기대합니다.",
			args: args{
				ctx:     context.Background(),
				options: []cdomain.AttachmentQueryOption{},
			},
			want: []domain.Attachment{
				mockAttachment1, mockAttachment2, mockAttachment3,
			},
		},
		{
			name: "Attachment 리소스 ID로 필터링 조회 : resourceId : e3ae2c14-b617-4870-b057-a8ae915cdee6 => 2건이 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.AttachmentQueryOption{
					cdomain.AttachmentQueryOptionFunc().WithResourceId("e3ae2c14-b617-4870-b057-a8ae915cdee6"),
				},
			},
			want: []domain.Attachment{
				mockAttachment2, mockAttachment3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := attachmentMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "GetAll() = %v, want %v", got, tt.want)
		})
	}
}

func Test_attachmentMysqlRepository_DeleteAllByResourceId(t *testing.T) {
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
			name: "Attachment 삭제 => resourceId 기반으로 삭제되기를 기대합니다.",
			args: args{
				ctx:        context.Background(),
				resourceId: mockAttachment1.ResourceId,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := attachmentMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			err := repo.DeleteByResourceId(tt.args.ctx, tt.args.resourceId)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_attachmentMysqlRepository_Delete(t *testing.T) {
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
			name: "Attachment 삭제 => ID 기반으로 삭제되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  mockAttachment1.Id,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := attachmentMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			err := repo.Delete(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
