package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_refreshTokenMysqlRepository_Create(t *testing.T) {
	adminUser10 := domain.AdminUser{
		Id:    10,
		Name:  "My Admin10",
		Email: "admin10@gmail.com",
	}
	token, _ := domain.CreateAdminRefreshToken(context.Background(), adminUser10)

	type args struct {
		ctx          context.Context
		refreshToken domain.RefreshToken
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "토큰 생성 (admin) -> 정상",
			args: args{
				ctx: context.Background(),
				refreshToken: domain.RefreshToken{
					UserId:   adminUser10.Id,
					UserType: domain.ADMIN,
					Token:    token,
					IsActive: true,
					Email:    adminUser10.Email,
					ExpireIn: time.Now().Add(time.Hour * 24).Unix(),
					IpAddr:   "192.168.0.10",
					Agent:    "Mozilla/5.0",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := refreshTokenMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			resultToken, err := repo.Create(tt.args.ctx, tt.args.refreshToken)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equal(t, tt.args.refreshToken.UserId, resultToken.UserId)
		})
	}
}

func Test_refreshTokenMysqlRepository_GetByToken(t *testing.T) {
	type args struct {
		ctx   context.Context
		token string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "토큰 조회 (by token) -> 정상",
			args: args{
				ctx:   context.Background(),
				token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJlbmVyZ3lzaGFyZXMiLCJhdWRpdF9zdGF0dXMiOiJJTlZBTElESVRZX0FVRElUIiwiZW1haWwiOiJhZG1pbkBlbmVyZ3lzaGFyZXMuanAiLCJleHAiOjE3MDM2NDAxNTQsImlzcyI6ImVuZXJneXNoYXJlcyIsIm5hbWUiOiJBRE1JTiIsIm5iZiI6MTcwMzYzOTg1NCwicm9sZSI6WyJST0xFX0lOVkVTVE9SIiwiUk9MRV9NRVNTRU5HRVIiLCJST0xFX1BBWU1FTlQiLCJST0xFX1BST0RVQ1QiLCJST0xFX09QRVJBVE9SIiwiUk9MRV9CQVNFIiwiUk9MRV9SRVBPUlQiLCJST0xFX0FETUlOIiwiUk9MRV9TVVBFUl9BRE1JTiJdLCJ1c2VyIjoiMSIsInVzZXJfdHlwZSI6IkFETUlOIn0.w8b3v8-bBtr5YmDBCGVY5FOV9TA_dAtdkTBDx_7knMU",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := refreshTokenMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			resultToken, err := repo.GetByToken(tt.args.ctx, tt.args.token)
			assert.Equalf(t, tt.wantErr, err != nil, "GetByToken() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equal(t, 1, int(resultToken.UserId))
		})
	}
}

func Test_refreshTokenMysqlRepository_Update(t *testing.T) {
	adminUser2 := domain.AdminUser{
		Id:    2,
		Name:  "My Admin2 modified",
		Email: "admin2@example.com",
	}

	type args struct {
		ctx          context.Context
		refreshToken domain.RefreshToken
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "토큰 업데이트 -> 정상",
			args: args{
				ctx: context.Background(),
				refreshToken: domain.RefreshToken{
					Id:       2,
					UserId:   adminUser2.Id,
					Token:    "new-token",
					IsActive: true,
					Email:    adminUser2.Email,
					ExpireIn: time.Now().Add(time.Hour * 24).Unix(),
					IpAddr:   "192.168.0.10",
					Agent:    "Mozilla/5.0",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := refreshTokenMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.refreshToken)
			assert.Equalf(t, tt.wantErr, err != nil, "Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
