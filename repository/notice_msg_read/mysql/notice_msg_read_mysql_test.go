package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_noticeMsgReadMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.NoticeMsgReadQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   []domain.NoticeMsgRead
		wantErr bool
	}{
		{
			name: "알림 정보 조회 => 정상",
			args: args{
				ctx:     context.Background(),
				options: []cdomain.NoticeMsgReadQueryOption{},
			},
			want:    2,
			want1:   []domain.NoticeMsgRead{mockNoticeMsgRead1, mockNoticeMsgRead2},
			wantErr: false,
		},
		{
			name: "알림 정보 조회 (유저 ID 조건) => 정상",
			args: args{
				ctx: context.Background(),
				options: []cdomain.NoticeMsgReadQueryOption{
					cdomain.NoticeMsgReadQueryOptionFunc().WithUserId(111),
				},
			},
			want:    1,
			want1:   []domain.NoticeMsgRead{mockNoticeMsgRead1},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := noticeMsgReadMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, got1, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			assert.NoErrorf(t, err, "GetAll() error = %v", err)
			assert.Equalf(t, tt.want, got, "GetAll() got = %v, want %v", got, tt.want)
			assert.Equalf(t, cmp.Equal(tt.want1, got1, ignoreTime), true, "GetAll() got1 = %v, want %v", got1, tt.want1)
		})
	}
}

func Test_noticeMsgReadMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx           context.Context
		noticeMsgRead domain.NoticeMsgRead
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "알림 정보 생성 => 정상",
			args: args{
				ctx: context.Background(),
				noticeMsgRead: domain.NoticeMsgRead{
					MsgId:  2,
					UserId: 111,
					IsRead: true,
					ReadAt: jsonutils.JSONTime{},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := noticeMsgReadMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.Create(tt.args.ctx, tt.args.noticeMsgRead)

			assert.NoErrorf(t, err, "Create() error = %v", err)
			assert.Greaterf(t, got, int64(0), "Create() got = %v, want > 0", got)
		})
	}
}

func Test_noticeMsgReadMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx           context.Context
		noticeMsgRead domain.NoticeMsgRead
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "알림 정보 업데이트 => 정상",
			args: args{
				ctx:           context.Background(),
				noticeMsgRead: mockNoticeMsgRead1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := noticeMsgReadMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.noticeMsgRead)
			assert.NoErrorf(t, err, "Update() error = %v", err)
		})
	}
}
