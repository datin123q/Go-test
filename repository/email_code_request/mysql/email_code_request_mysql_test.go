package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"
)

func Test_emailCodeRequestRepository_Create(t *testing.T) {
	type args struct {
		ctx              context.Context
		emailCodeRequest domain.EmailCodeRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) EmailCodeRequest 추가 -> 정상적으로 insert 되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				emailCodeRequest: domain.EmailCodeRequest{
					UserId: 100,
					Email:  "gildong@email.com",
					Code:   "123456",
					Expire: jsonutils.JSONTime(time.Now().Add(time.Minute * 3)),
				},
			},
			wantErr: false,
		},
		{
			name: "(실패) EmailCodeRequest 추가 -> id 중복으로 실패하기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				emailCodeRequest: domain.EmailCodeRequest{
					Id:     1,
					UserId: 100,
					Email:  "gildong@email.com",
					Code:   "123456",
					Expire: jsonutils.JSONTime(time.Now().Add(time.Minute * 3)),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := emailCodeRequestRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.emailCodeRequest)
			assert.Equalf(t, tt.wantErr, err != nil, "emailCodeRequest.Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_emailCodeRequestRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.EmailCodeRequestQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.EmailCodeRequest
		wantErr bool
	}{
		{
			name: "(성공) EmailCodeRequest 특정 유저로 조회 -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.EmailCodeRequestQueryOption{
					cdomain.EmailCodeRequestQueryOptionFunc().WithUserId(100),
				},
			},

			want: []domain.EmailCodeRequest{
				mockEmailCodeRequest1,
			},
			wantErr: false,
		},
		{
			name: "(성공) 현재 시간보다 이틀 이후로 생성날짜를 필터링 -> 조회되는 값이 없습니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.EmailCodeRequestQueryOption{
					cdomain.EmailCodeRequestQueryOptionFunc().WithGTECreated(time.Now().Add(time.Hour * 48).Format(time.DateTime)),
				},
			},

			want:    []domain.EmailCodeRequest{},
			wantErr: false,
		},
		{
			name: "(성공) 최초 요청여부로 필터링 -> 조회되는 값이 없습니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.EmailCodeRequestQueryOption{
					cdomain.EmailCodeRequestQueryOptionFunc().WithFirstReq(true),
				},
			},

			want:    []domain.EmailCodeRequest{},
			wantErr: false,
		},
		{
			name: "(성공) 특정 코드 값(654321)으로 필터링 -> 조회되는 값이 없습니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.EmailCodeRequestQueryOption{
					cdomain.EmailCodeRequestQueryOptionFunc().WithCode("654321"),
				},
			},

			want:    []domain.EmailCodeRequest{},
			wantErr: false,
		},
		{
			name: "(성공) 특정 이메일로 필터링 (dool2@email.com) -> 조회되는 값이 없습니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.EmailCodeRequestQueryOption{
					cdomain.EmailCodeRequestQueryOptionFunc().WithEmail("dool2@email.com"),
				},
			},

			want:    []domain.EmailCodeRequest{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := emailCodeRequestRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.GetAll(tt.args.ctx, tt.args.options...)
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "emailCodeRequest.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "emailCodeRequest.GetAll() = %v, want %v", got, tt.want)
		})
	}
}

func Test_emailCodeRequestRepository_Update(t *testing.T) {
	type args struct {
		ctx              context.Context
		emailCodeRequest domain.EmailCodeRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) EmailCodeRequest 코드 값 수정 -> 정상적으로 update 되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				emailCodeRequest: domain.EmailCodeRequest{
					Id:   1,
					Code: "654321",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := emailCodeRequestRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.emailCodeRequest)
			assert.Equalf(t, tt.wantErr, err != nil, "emailCodeRequest.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
