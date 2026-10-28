package mysql

import (
	"context"
	"testing"

	"github.com/samber/lo"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
)

func Test_investmentRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.InvestmentQueryOption
	}
	tests := []struct {
		name      string
		args      args
		wantCount int64
		wantErr   bool
	}{
		{
			name: "全て照会",
			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentQueryOption{
					cdomain.InvestmentQueryOptionFunc().WithLimit(100),
					cdomain.InvestmentQueryOptionFunc().WithOffset(0),
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "Offset/Limitが正しく動作することを確認",
			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentQueryOption{
					cdomain.InvestmentQueryOptionFunc().WithLimit(1),
					cdomain.InvestmentQueryOptionFunc().WithOffset(5),
				},
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "投資状態がSUCCESS_INVESTである投資件を照会 => 1件照会されることを期待",
			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentQueryOption{
					cdomain.InvestmentQueryOptionFunc().WithLimit(100),
					cdomain.InvestmentQueryOptionFunc().WithOffset(0),
					cdomain.InvestmentQueryOptionFunc().WithInvestStatus(domain.SUCCESS_INVEST),
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "投資条件を照会します。" +
				"1. プロジェクトタイプをフィルタリングします : PROJECT_TYPE.EQUITY" +
				"2. 投資家名でフィルタリングします : ホン・ギルドン" +
				"3. 開発会社でフィルタリングします : エネルギー" + // エネルギー開発会社の後方一致LIKE検索
				"=> 1件の投資件が照会されることを期待します。",

			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentQueryOption{
					cdomain.InvestmentQueryOptionFunc().WithLimit(100),
					cdomain.InvestmentQueryOptionFunc().WithOffset(0),
					cdomain.InvestmentQueryOptionFunc().WithProjectType(lo.ToPtr(string(domain.EQUITY))),
					cdomain.InvestmentQueryOptionFunc().WithInvestorName(lo.ToPtr("ホン・ギルドン")),
					cdomain.InvestmentQueryOptionFunc().WithSpvName(lo.ToPtr("エネルギー")),
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "投資を\"テスト商品\"商品名で照会 => 照会される商品名が存在しないため0件が出ることを期待します。",

			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentQueryOption{
					cdomain.InvestmentQueryOptionFunc().WithLimit(100),
					cdomain.InvestmentQueryOptionFunc().WithOffset(0),
					cdomain.InvestmentQueryOptionFunc().WithProjectName(lo.ToPtr("テスト商品")),
				},
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "投資総額が5000000以上10000000以下の値を照会 => 総額が5000000である1件が照会されることを期待します。",

			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentQueryOption{
					cdomain.InvestmentQueryOptionFunc().WithLimit(100),
					cdomain.InvestmentQueryOptionFunc().WithOffset(0),
					cdomain.InvestmentQueryOptionFunc().WithStartAmount(lo.ToPtr(5000000)),
					cdomain.InvestmentQueryOptionFunc().WithEndAmount(lo.ToPtr(10000000)),
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := investmentRepository{
				dbFactory: &dbFactoryMock{},
			}
			_, got, err := c.GetAll(tt.args.ctx, tt.args.options...)

			assert.Equalf(t, tt.wantErr, err != nil, "investmentRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, int(tt.wantCount), len(got), "investmentRepository.GetAll() = %v, want %v", len(got), tt.wantCount)
		})
	}
}

func Test_investmentRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.Investment
		wantErr bool
	}{
		{
			name: "IDがshareId1である投資件を照会 => 照会されることを期待",
			args: args{
				ctx: context.Background(),
				id:  "shareId1",
			},
			want: domain.Investment{
				Share: mockShare1,
				Project: domain.Project{
					Id:       projectId1,
					SpvId:    spvId1,
					Name:     "太陽光発電プロジェクト",
					Capacity: "100MW",
				},
				InvestmentId: "shareId1",
				UserType:     "一般",
				FirstName:    "ギルドン",
				FamilyName:   "ホン",
				SpvName:      "エネルギー開発会社",
				UserName:     "ホン・ギルドン",
				InvestorType: domain.PRIVATE_PRO_INVESTOR,
				Email:        "honggildong@example.com",
				MobilePhone:  "010-1234-5678",
				Address:      "ソウル特別市江南区",
				City:         "ソウル",
				ZipCode:      "12345",
				Province:     "江南区",
			},
			wantErr: false,
		},
		{
			name: "IDが100である投資件を照会 => 照会されないことを期待",
			args: args{
				ctx: context.Background(),
				id:  "shareId2",
			},

			want:    domain.Investment{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := investmentRepository{
				dbFactory: &dbFactoryMock{},
			}
			got, err := c.GetById(tt.args.ctx, tt.args.id)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "investmentRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(got, tt.want, ignoreTime), "investmentRepository.GetById() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investmentRepository_GroupByProject(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.InvestmentQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.Investment
		wantErr bool
	}{
		{
			name: "特定プロジェクトグループ投資リスト返却",
			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentQueryOption{
					cdomain.InvestmentQueryOptionFunc().WithProjectId(projectId1),
				},
			},
			want: []domain.Investment{
				{
					Share:        mockShare1,
					UserType:     "一般",
					UserName:     "ホン・ギルドン",
					InvestorType: domain.PRIVATE_PRO_INVESTOR,
					Email:        "honggildong@example.com",
					MobilePhone:  "010-1234-5678",
					ZipCode:      "12345",
					Address:      "ソウル特別市江南区",
					City:         "ソウル",
					Province:     "江南区",
				},
			},
			wantErr: false,
		},
		{
			name: "プロジェクトでグループ投資リスト返却 : InvestStatus - FAIL_INVEST : 0件が照会されます。",
			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentQueryOption{
					cdomain.InvestmentQueryOptionFunc().WithProjectId(projectId1),
					cdomain.InvestmentQueryOptionFunc().WithInvestStatus(domain.FAIL_INVEST),
				},
			},
			want:    []domain.Investment{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := investmentRepository{
				dbFactory: &dbFactoryMock{},
			}
			got, err := c.GroupByOwner(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreFields := cmpopts.IgnoreFields(domain.Investment{}, "Project")

			assert.Equalf(t, tt.wantErr, err != nil, "investmentRepository.GroupByOwner() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(got, tt.want, ignoreTime, ignoreFields), "investmentRepository.GroupByOwner() = %v, want %v", got, tt.want)
		})
	}

}
