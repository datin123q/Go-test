package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_passwordResetRequestMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx                  context.Context
		passwordResetRequest domain.PasswordResetRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    int64
	}{
		{
			name: "(성공) PasswordResetRequest 모델 생성",
			args: args{
				ctx: context.Background(),
				passwordResetRequest: domain.PasswordResetRequest{
					RequestId: "11110000-0000-0000-0000-000000000000",
					Email:     "ilmuser@email.com",
					Url:       "https://es-jp-dev.illuminarean.com/reset/password/00000000-0000-0000-0000-000000000000",
					SendCount: 0,
				},
			},
			wantErr: false,
			want:    3,
		},
		{
			name: "(실패) 기존에 이미 있는 PasswordResetRequest 모델 생성 => ID 중복으로 에러 발생",
			args: args{
				ctx: context.Background(),
				passwordResetRequest: domain.PasswordResetRequest{
					Id:        1,
					RequestId: "00000000-0000-0000-0000-000000000000",
					Url:       "https://es-jp-dev.illuminarean.com/reset/password/00000000-0000-0000-0000-000000000000",
				},
			},
			wantErr: true,
			want:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := passwordResetRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.Create(tt.args.ctx, tt.args.passwordResetRequest)
			if (err != nil) != tt.wantErr {
				t.Errorf("passwordResetRequestMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("passwordResetRequestMysqlRepository.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_passwordResetRequestMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx                  context.Context
		passwordResetRequest domain.PasswordResetRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) PasswordResetRequest 모델 업데이트",
			args: args{
				ctx: context.Background(),
				passwordResetRequest: domain.PasswordResetRequest{
					Id:        1,
					RequestId: "00000000-0000-0000-0000-000000000000",
					Email:     "changed@email.com",
					Url:       "changed_url",
					SendCount: 100,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := passwordResetRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.passwordResetRequest)
			if (err != nil) != tt.wantErr {
				t.Errorf("passwordResetRequestMysqlRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_passwordResetRequestMysqlRepository_GetLatestByEmail(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.PasswordResetRequest
		wantErr bool
	}{
		{
			name: "(성공) 제일 최근에 생성된 PasswordResetRequest 모델을 email로 조회",
			args: args{
				ctx:   context.Background(),
				email: "gildong@email.com",
			},
			want: domain.PasswordResetRequest{
				Id:        2,
				RequestId: "20000000-0000-0000-0000-000000000000",
				Email:     "gildong@email.com",
				Url:       "https://es-jp-dev.illuminarean.com/reset/password/20000000-0000-0000-0000-000000000000",
				SendCount: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := passwordResetRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.GetLatestByEmail(tt.args.ctx, tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("passwordResetRequestMysqlRepository.GetLatestByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "passwordResetRequestMysqlRepository.GetLatestByEmail() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_passwordResetRequestMysqlRepository_GetByRequestId(t *testing.T) {
	type args struct {
		ctx       context.Context
		requestId string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.PasswordResetRequest
		wantErr bool
	}{
		{
			name: "(성공) PasswordResetRequest 모델을 requestId로 조회",
			args: args{
				ctx:       context.Background(),
				requestId: "00000000-0000-0000-0000-000000000000",
			},
			want:    mockPasswordResetRequest1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := passwordResetRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.GetByRequestId(tt.args.ctx, tt.args.requestId)
			if (err != nil) != tt.wantErr {
				t.Errorf("passwordResetRequestMysqlRepository.GetByRequestId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "passwordResetRequestMysqlRepository.GetByRequestId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(got, tt.want, ignoreTime), "passwordResetRequestMysqlRepository.GetByRequestId() = %v, want %v", got, tt.want)
		})
	}
}
