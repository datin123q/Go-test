package mysql

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_adminUserMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.AdminUserQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.AdminUser
		wantErr bool
	}{
		{
			name: "전체 AdminUser 조회",
			args: args{
				ctx: context.Background(),
				options: []cdomain.AdminUserQueryOption{
					cdomain.AdminUserQueryOptionFunc().WithLimit(100),
					cdomain.AdminUserQueryOptionFunc().WithOffset(0),
				},
			},
			want: []domain.AdminUser{
				mockAdminUser1,
				mockAdminUser2,
			},
			wantErr: false,
		},
		{
			name: "Limit 1, Offset 1로 AdminUser 조회 : User2 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.AdminUserQueryOption{
					cdomain.AdminUserQueryOptionFunc().WithLimit(1),
					cdomain.AdminUserQueryOptionFunc().WithOffset(1),
				},
			},
			want: []domain.AdminUser{
				mockAdminUser2,
			},
			wantErr: false,
		},
		{
			name: "특정 전화번호 앞자리로 AdminUser 조회 : 0101234 => User1 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.AdminUserQueryOption{
					cdomain.AdminUserQueryOptionFunc().WithLimit(100),
					cdomain.AdminUserQueryOptionFunc().WithOffset(0),
					cdomain.AdminUserQueryOptionFunc().WithMobilePhone("0101234"),
				},
			},
			want: []domain.AdminUser{
				mockAdminUser1,
			},
			wantErr: false,
		},
		{
			name: "특정이름과 휴대폰번호로 AdminUser를 조회 : 홍길동, 01012345678 => User1 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.AdminUserQueryOption{
					cdomain.AdminUserQueryOptionFunc().WithLimit(100),
					cdomain.AdminUserQueryOptionFunc().WithOffset(0),
					cdomain.AdminUserQueryOptionFunc().WithMobilePhone("01012345678"),
					cdomain.AdminUserQueryOptionFunc().WithName("홍길동"),
				},
			},
			want: []domain.AdminUser{
				mockAdminUser1,
			},
			wantErr: false,
		},
		{
			name: "특정이름으로 조회했지만 찾지를 못합니다. : 도우너 => 빈리스트가 조회됩니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.AdminUserQueryOption{
					cdomain.AdminUserQueryOptionFunc().WithLimit(100),
					cdomain.AdminUserQueryOptionFunc().WithOffset(0),
					cdomain.AdminUserQueryOptionFunc().WithName("도우너"),
				},
			},
			want:    []domain.AdminUser{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := adminUserMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreField := cmpopts.IgnoreFields(domain.AdminUser{}, "RolesString")

			assert.Equalf(t, tt.wantErr, err != nil, "adminUserMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime, ignoreField), true, "adminUserMysqlRepository.GetAll() = %v, want %v", got, tt.want)
		})
	}
}

func Test_adminUserMysqlRepository_GetByEmail(t *testing.T) {
	type fields struct {
		dbFactory factory.DBFactory
	}
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    domain.AdminUser
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "AdminUser 이메일로 조회",
			fields: fields{
				dbFactory: dbFactoryMock{},
			},
			args: args{
				ctx:   context.Background(),
				email: "dool2@email.com",
			},
			want:    mockAdminUser2,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adminUserMysqlRepository{
				dbFactory: tt.fields.dbFactory,
			}
			got, err := a.GetByEmail(tt.args.ctx, tt.args.email)
			if !tt.wantErr(t, err, fmt.Sprintf("GetByEmail(%v, %v)", tt.args.ctx, tt.args.email)) {
				return
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "GetByEmail(%v, %v)", tt.args.ctx, tt.args.email)
		})
	}
}

func Test_adminUserMysqlRepository_GetById(t *testing.T) {
	type fields struct {
		dbFactory factory.DBFactory
	}
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    domain.AdminUser
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "AdminUser ID로 조회",
			fields: fields{
				dbFactory: dbFactoryMock{},
			},
			args: args{
				ctx:    context.Background(),
				userId: 1000,
			},
			want:    mockAdminUser1,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adminUserMysqlRepository{
				dbFactory: tt.fields.dbFactory,
			}
			got, err := a.GetById(tt.args.ctx, tt.args.userId)
			if !tt.wantErr(t, err, fmt.Sprintf("GetById(%v, %v)", tt.args.ctx, tt.args.userId)) {
				return
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreField := cmpopts.IgnoreFields(domain.AdminUser{}, "RolesString")
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime, ignoreField), true, "GetById(%v, %v)", tt.args.ctx, tt.args.userId)
		})
	}
}

func Test_adminUserMysqlRepository_Update(t *testing.T) {
	type fields struct {
		dbFactory factory.DBFactory
	}
	type args struct {
		ctx  context.Context
		user domain.AdminUser
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "AdminUser 업데이트",
			fields: fields{
				dbFactory: dbFactoryMock{},
			},
			args: args{
				ctx: context.Background(),
				user: func() domain.AdminUser {
					u := mockAdminUser1
					u.Name = "홍길순"
					u.MobilePhone = "01098765432"
					return u
				}(),
			},
			wantErr: assert.NoError,
		},
		{
			name: "AdminUser 아이디가 잘못된 경우 업데이트 실패하는 케이스",
			fields: fields{
				dbFactory: dbFactoryMock{},
			},
			args: args{
				ctx: context.Background(),
				user: func() domain.AdminUser {
					u := mockAdminUser1
					u.Id = 9999
					u.Name = "홍길순"
					return u
				}(),
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adminUserMysqlRepository{
				dbFactory: tt.fields.dbFactory,
			}
			tt.wantErr(t, a.Update(tt.args.ctx, tt.args.user), fmt.Sprintf("Update(%v, %v)", tt.args.ctx, tt.args.user))
		})
	}
}

func Test_adminUserMysqlRepository_Delete(t *testing.T) {
	type fields struct {
		dbFactory factory.DBFactory
	}
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "AdminUser 삭제",
			fields: fields{
				dbFactory: dbFactoryMock{},
			},
			args: args{
				ctx:    context.Background(),
				userId: 1000,
			},
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.NoError(t, err, msgAndArgs...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adminUserMysqlRepository{
				dbFactory: tt.fields.dbFactory,
			}
			tt.wantErr(t, a.Delete(tt.args.ctx, tt.args.userId), fmt.Sprintf("Delete(%v, %v)", tt.args.ctx, tt.args.userId))
		})
	}
}

func Test_adminUserMysqlRepository_Create(t *testing.T) {
	type fields struct {
		dbFactory factory.DBFactory
	}
	type args struct {
		ctx  context.Context
		user domain.AdminUser
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    domain.AdminUser
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "AdminUser 생성",
			fields: fields{
				dbFactory: dbFactoryMock{},
			},
			args: args{
				ctx:  context.Background(),
				user: mockCreateAdminUser3,
			},
			want:    mockCreateAdminUser3,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adminUserMysqlRepository{
				dbFactory: tt.fields.dbFactory,
			}
			got, err := a.Create(tt.args.ctx, tt.args.user)
			if !tt.wantErr(t, err, fmt.Sprintf("Create(%v, %v)", tt.args.ctx, tt.args.user)) {
				return
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreField := cmpopts.IgnoreFields(domain.AdminUser{}, "RolesString")
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime, ignoreField), true, "Create(%v, %v)", tt.args.ctx, tt.args.user)
		})
	}
}
