package mysql

import (
	"context"
	"github.com/samber/lo"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_userAccountMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx         context.Context
		userAccount domain.UserAccount
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) 유저 생성",
			args: args{
				ctx:         context.Background(),
				userAccount: mockUserAccountToInsert,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userAccountMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			err := repo.Create(tt.args.ctx, tt.args.userAccount)

			assert.Equalf(t, tt.wantErr, err != nil, "userAccountMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_user_account_mysql_repository_Update(t *testing.T) {
	type args struct {
		ctx         context.Context
		userAccount domain.UserAccount
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "특정 유저의 계좌를 업데이트합니다 : 1000 => 업데이트가 성공해야 합니다.",
			args: args{
				ctx: context.Background(),
				userAccount: domain.UserAccount{
					Id:      2000,
					UserId:  1000,
					Balance: lo.ToPtr(int64(100000)),
					Version: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "특정 유저의 계좌를 업데이트합니다 : 0 => 업데이트가 성공해야 합니다.",
			args: args{
				ctx: context.Background(),
				userAccount: domain.UserAccount{
					Id:      2000,
					UserId:  1000,
					Balance: lo.ToPtr(int64(0)),
					Version: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "잔액을 업데이트하지 않고 버전만 업데이트 합니다. => 업데이트가 성공해야 합니다.",
			args: args{
				ctx: context.Background(),
				userAccount: domain.UserAccount{
					Id:      2000,
					UserId:  1000,
					Version: 1,
				},
			},
			wantErr: false,
		},
		{
			// MIR-1247: シードデータ(mockUserAccount1)のversionは1。ここでは
			// version不一致(=2)を渡し、0行更新がエラーとして呼び出し元に
			// 伝わることを固定する。以前はUpdate()の影響行数を見ておらず、
			// この場合もwantErr=falseだった。
			name: "버전이 일치하지 않는 상태로 업데이트합니다. => 0행 갱신이므로 에러가 되어야 합니다.",
			args: args{
				ctx: context.Background(),
				userAccount: domain.UserAccount{
					Id:      2000,
					UserId:  1000,
					Balance: lo.ToPtr(int64(100000)),
					Version: 2,
				},
			},
			wantErr:   true,
			wantErrIs: cdomain.ErrItemNotAffected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := userAccountMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			err := u.Update(tt.args.ctx, tt.args.userAccount)
			assert.Equalf(t, tt.wantErr, err != nil, "userAccountMysqlRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs, "呼び出し元がerrors.Isで判別できること")
			}
		})
	}
}

func Test_user_account_mysql_repository_GetByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.UserAccount
		wantErr bool
	}{
		{
			name: "특정 유저의 계좌를 조회합니다 : 1000 => 1건 조회되기를 기대합니다.",
			args: args{
				ctx:    context.Background(),
				userId: 1000,
			},
			want: domain.UserAccount{
				Id:      2000,
				UserId:  1000,
				Balance: lo.ToPtr(int64(100000)),
				Version: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := userAccountMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := u.GetByUserId(tt.args.ctx, tt.args.userId)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.NoErrorf(t, err, "userAccountMysqlRepository.GetByUserId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "userAccountMysqlRepository.GetByUserId() = %v, want %v", got, tt.want)
		})
	}
}
