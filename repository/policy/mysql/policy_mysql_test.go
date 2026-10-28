package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func Test_policyRepository_Create(t *testing.T) {
	type args struct {
		ctx    context.Context
		policy domain.Policy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "policy 생성",
			args: args{
				ctx: context.Background(),
				policy: domain.Policy{
					CommonId:       "00000000-0000-0000-0000-000000000000",
					OwnerId:        1,
					ModifierId:     1,
					PolicyViewType: domain.POLICY_VIEW_FOOTER,
					PolicyType:     domain.POLICY_CAUTION,
					ContentHtml:    "content",
					ModifierEmail:  "admin@energyshares.jp",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := policyMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, err := repo.Create(tt.args.ctx, tt.args.policy)

			assert.Equalf(t, tt.wantErr, err != nil, "PolicyRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_policyRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.PolicyQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.Policy
		wantErr bool
	}{
		{
			name: "policy 목록 조회 - 조건 없음",
			args: args{
				ctx: context.Background(),
				options: []cdomain.PolicyQueryOption{
					cdomain.PolicyQueryOptionFunc().WithSortBy("id"),
					cdomain.PolicyQueryOptionFunc().WithOrder("asc"),
					cdomain.PolicyQueryOptionFunc().WithOffset(0),
					cdomain.PolicyQueryOptionFunc().WithLimit(10),
				},
			},
			want: []domain.Policy{
				mockPolicy1,
				mockPolicy2,
			},
			wantErr: false,
		},
		{
			name: "policy 목록 조회 - 단일 조건",
			args: args{
				ctx: context.Background(),
				options: []cdomain.PolicyQueryOption{
					cdomain.PolicyQueryOptionFunc().WithPolicyViewType(lo.ToPtr(string(domain.POLICY_VIEW_FOOTER))),
					cdomain.PolicyQueryOptionFunc().WithSortBy("id"),
					cdomain.PolicyQueryOptionFunc().WithOrder("desc"),
					cdomain.PolicyQueryOptionFunc().WithOffset(0),
					cdomain.PolicyQueryOptionFunc().WithLimit(10),
				},
			},
			want: []domain.Policy{
				mockPolicy2,
			},
			wantErr: false,
		},
		{
			name: "policy 목록 조회 - 다중 조건 - policyType, policyViewType",
			args: args{
				ctx: context.Background(),
				options: []cdomain.PolicyQueryOption{
					cdomain.PolicyQueryOptionFunc().WithPolicyViewType(lo.ToPtr(string(domain.POLICY_VIEW_FOOTER))),
					cdomain.PolicyQueryOptionFunc().WithPolicyType(lo.ToPtr(string(domain.POLICY_PROTECTION))),
					cdomain.PolicyQueryOptionFunc().WithSortBy("id"),
					cdomain.PolicyQueryOptionFunc().WithOrder("desc"),
					cdomain.PolicyQueryOptionFunc().WithOffset(0),
					cdomain.PolicyQueryOptionFunc().WithLimit(10),
				},
			},
			want:    []domain.Policy{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := policyMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "PolicyRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "GetAll() diff = %v", cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}

func Test_policyRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.Policy
		wantErr bool
	}{
		{
			name: "policy 조회",
			args: args{
				ctx: context.Background(),
				id:  mockPolicy1.Id,
			},
			want:    mockPolicy1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := policyMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, err := repo.GetById(tt.args.ctx, tt.args.id)

			assert.Equalf(t, tt.wantErr, err != nil, "PolicyRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_policyRepository_Update(t *testing.T) {
	type args struct {
		ctx    context.Context
		policy domain.Policy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "policy 수정",
			args: args{
				ctx: context.Background(),
				policy: domain.Policy{
					Id:             mockPolicy1.Id,
					CommonId:       mockPolicy1.CommonId,
					OwnerId:        mockPolicy1.OwnerId,
					ModifierId:     mockPolicy1.ModifierId,
					PolicyViewType: mockPolicy1.PolicyViewType,
					PolicyType:     mockPolicy1.PolicyType,
					ContentHtml:    "changed content",
					ModifierEmail:  "test@example.com",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := policyMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.policy)

			assert.Equalf(t, tt.wantErr, err != nil, "PolicyRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_policyRepository_Delete(t *testing.T) {
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
			name: "policy 삭제",
			args: args{
				ctx: context.Background(),
				id:  mockPolicy1.Id,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := policyMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Delete(tt.args.ctx, tt.args.id)

			assert.Equalf(t, tt.wantErr, err != nil, "PolicyRepository.Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
