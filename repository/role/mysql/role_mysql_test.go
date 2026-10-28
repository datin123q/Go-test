package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_roleMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx  context.Context
		role domain.Role
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "(성공) Role 생성",
			args: args{
				ctx: context.Background(),
				role: domain.Role{
					UserId:   1000,
					RoleCode: string(domain.ROLE_USER),
				},
			},
			want:    mockRole4.Id + 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := roleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.Create(tt.args.ctx, tt.args.role)

			assert.Equalf(t, tt.wantErr, err != nil, "roleMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "Create() = %v, want %v", got, tt.want)
		})
	}
}

func Test_roleMysqlRepository_Delete(t *testing.T) {
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
			name: "(성공) Role 삭제",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := roleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			if err := r.Delete(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_roleMysqlRepository_DeleteByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) User ID로 Role 삭제",
			args: args{
				ctx:    context.Background(),
				userId: 2000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := roleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			if err := r.DeleteByUserId(tt.args.ctx, tt.args.userId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteByUserId() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_roleMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.RoleQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   []domain.Role
		wantErr bool
	}{
		{
			name: "(성공) Role 전체 조회",
			args: args{
				ctx: context.Background(),
				options: []cdomain.RoleQueryOption{
					cdomain.RoleQueryOptionFunc().WithUserId(1000),
				},
			},
			want:    3,
			want1:   []domain.Role{mockRole1, mockRole2, mockRole3},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := roleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			_, got, err := r.GetAll(tt.args.ctx, tt.args.options...)

			assert.Equalf(t, tt.wantErr, err != nil, "roleMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, int64(len(got)), "GetAll() length = %v, want %v", len(got), tt.want)
			assert.Equalf(t, true, cmp.Equal(tt.want1, got, ignoreTime), "GetAll() diff = %v", cmp.Diff(tt.want1, got, ignoreTime))
		})
	}
}

func Test_roleMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.Role
		wantErr bool
	}{
		{
			name: "(성공) Role ID로 Role 조회",
			args: args{
				ctx: context.Background(),
				id:  mockRole1.Id,
			},
			want:    mockRole1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := roleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := r.GetById(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.wantErr, err != nil, "roleMysqlRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "GetById() diff = %v", cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}

func Test_roleMysqlRepository_GetByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.Role
		wantErr bool
	}{
		{
			name: "(성공) User ID로 Role 조회",
			args: args{
				ctx:    context.Background(),
				userId: 1000,
			},
			want:    []domain.Role{mockRole1, mockRole2, mockRole3},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := roleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := r.GetByUserId(tt.args.ctx, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "roleMysqlRepository.GetByUserId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "GetByUserId() diff = %v", cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}
