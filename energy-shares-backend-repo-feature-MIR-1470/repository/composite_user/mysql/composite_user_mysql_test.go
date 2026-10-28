package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func Test_CompositeUserMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.CompositeUserQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []cdomain.CompositeUser
		wantErr bool
	}{
		{
			name: "ユーザー照会: boolフィルタ確認 - ニュースレター同意者の照会",
			args: args{
				ctx: context.Background(),
				options: []cdomain.CompositeUserQueryOption{
					cdomain.CompositeUserQueryOptionFunc().WithAllowsNewsletter(lo.ToPtr(true)),
				},
			},
			want: []cdomain.CompositeUser{compositeUser1, compositeUser2, compositeUser3},
		},
		{
			name: "ユーザー照会: boolフィルタ確認 - my number APPROVEステータスの人の照会",
			args: args{
				ctx: context.Background(),
				options: []cdomain.CompositeUserQueryOption{
					cdomain.CompositeUserQueryOptionFunc().WithHasMyNumber(lo.ToPtr(true)),
				},
			},
			want: []cdomain.CompositeUser{compositeUser1},
		},
		// TODO: マイナンバー登録復活後に元に戻す
		// {
		// 	name: "ユーザー照会: 文字列INフィルタ確認 - auditStatusがCOMPLETED_AUDIT、CHECK_MY_NUMBER_AUDITステータスの人の照会",
		// 	args: args{
		// 		ctx: context.Background(),
		// 		options: []cdomain.CompositeUserQueryOption{
		// 			cdomain.CompositeUserQueryOptionFunc().WithAuditStatuses(domain.COMPLETED_AUDIT, domain.CHECK_MY_NUMBER_AUDIT),
		// 		},
		// 	},
		// 	want: []cdomain.CompositeUser{compositeUser1, compositeUser2, compositeUser3},
		// },
		{
			name: "ユーザー照会: 日付範囲フィルタ確認 - 最後のログイン日が2024-11-1以前の人の照会",
			args: args{
				ctx: context.Background(),
				options: []cdomain.CompositeUserQueryOption{
					cdomain.CompositeUserQueryOptionFunc().WithMaxLastLoginDate(lo.ToPtr("2024-11-01")),
				},
			},
			want: []cdomain.CompositeUser{compositeUser1, compositeUser2},
		},
		{
			name: "ユーザー照会: 数値範囲フィルタ確認 - 総資産が10000以上の人の照会",
			args: args{
				ctx: context.Background(),
				options: []cdomain.CompositeUserQueryOption{
					cdomain.CompositeUserQueryOptionFunc().WithMinTotalAsset(lo.ToPtr(int64(10000))),
				},
			},
			want: []cdomain.CompositeUser{compositeUser3},
		},
		{
			name: "ユーザー照会: 複合フィルタ確認",
			args: args{
				ctx: context.Background(),
				options: []cdomain.CompositeUserQueryOption{
					cdomain.CompositeUserQueryOptionFunc().WithMaxTotalAsset(lo.ToPtr(int64(10000))),
					cdomain.CompositeUserQueryOptionFunc().WithAllowsNewsletter(lo.ToPtr(true)),
					cdomain.CompositeUserQueryOptionFunc().WithMinBirthday(lo.ToPtr("1988-02-09")),
					cdomain.CompositeUserQueryOptionFunc().WithMaxBirthday(lo.ToPtr("2000-02-09")),
				},
			},
			want: []cdomain.CompositeUser{compositeUser1, compositeUser2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := compositeUserRepository{
				dbFactory: dbFactoryMock{},
			}
			_, got, err := r.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "compositeUserRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "GetAll() diff = %v", cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}
