package mysql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_UserNoticeMappingRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.UserNoticeMappingQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   []domain.UserNoticeMapping
		wantErr bool
	}{
		{
			name: "메시지 ID로 매핑 리스트 조회 => 정상",
			args: args{
				ctx: context.Background(),
				options: []cdomain.UserNoticeMappingQueryOption{
					cdomain.UserNoticeMappingQueryOptionFunc().WithMsgId(1),
				},
			},
			want:    1,
			want1:   []domain.UserNoticeMapping{mockUserNoticeMapping1},
			wantErr: false,
		},
		{
			name: "메시지 ID, 사용자 ID로 매핑 리스트 조회 => 정상",
			args: args{
				ctx: context.Background(),
				options: []cdomain.UserNoticeMappingQueryOption{
					cdomain.UserNoticeMappingQueryOptionFunc().WithMsgId(1),
					cdomain.UserNoticeMappingQueryOptionFunc().WithUserId(111),
				},
			},
			want:    1,
			want1:   []domain.UserNoticeMapping{mockUserNoticeMapping1},
			wantErr: false,
		},
		{
			name: "사용자 ID로 매핑 리스트 조회 => 정상",
			args: args{
				ctx: context.Background(),
				options: []cdomain.UserNoticeMappingQueryOption{
					cdomain.UserNoticeMappingQueryOptionFunc().WithUserId(111),
				},
			},
			want:    4,
			want1:   []domain.UserNoticeMapping{mockUserNoticeMapping1, mockUserNoticeMapping2, mockUserNoticeMapping3, mockUserNoticeMapping4},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userNoticeMappingRepository{
				dbFactory: dbFactoryMock{},
			}

			count, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equal(t, tt.want, count)
			assert.Equalf(t, tt.want1, got, "GetAll() got = %v, want %v", got, tt.want1)
		})
	}
}

func Test_UserNoticeMappingRepository_Create(t *testing.T) {
	type args struct {
		ctx     context.Context
		mapping domain.UserNoticeMapping
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "create mapping",
			args: args{
				ctx: context.Background(),
				mapping: domain.UserNoticeMapping{
					MsgId:  12,
					UserId: 107,
				},
			},
			want:    12,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userNoticeMappingRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.Create(tt.args.ctx, tt.args.mapping)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "Create() got = %v, want %v", got, tt.want)
		})
	}
}
