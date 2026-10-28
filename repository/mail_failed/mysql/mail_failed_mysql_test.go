package mysql

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_mailFailedMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx        context.Context
		mailFailed domain.MailFailed
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    int64
	}{
		{
			name: "MailFailedテーブルにMailFailedモデルを作成 => 正常に作成されることが期待されます。",
			args: args{
				ctx:        context.Background(),
				mailFailed: mockMailFailed1,
			},
			wantErr: false,
			want:    1001,
		},
		{
			name: "既に存在するMailFailedモデル生成: 1002 => エラーが発生することが期待されます。",
			args: args{
				ctx:        context.Background(),
				mailFailed: mockMailFailed2,
			},
			wantErr: true,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mailFailedMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.Create(tt.args.ctx, tt.args.mailFailed)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "Create() = %v, want %v", got, tt.want)
		})
	}
}

func Test_mailFailedMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.MailFailedQueryOption
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		wantTotal int64
		want      []domain.MailFailed
	}{
		{
			name: "メール送信リクエストの最終失敗メールリストを期待します。 テンプレート名: NORMAL_MAIL",
			args: args{
				ctx: context.Background(),
				options: []cdomain.MailFailedQueryOption{
					cdomain.MailFailedQueryOptionFunc().WithTemplateName(lo.ToPtr("NORMAL_MAIL")),
				},
			},
			wantErr:   false,
			wantTotal: int64(2),
			want:      []domain.MailFailed{mockMailFailed2, mockMailFailed3},
		},
		// SQLiteではJSON_UNQUOTE関数を使用できないため、テストを実行しません。
		// {
		// 	name: "メール送信リクエストの最終失敗メールリストを期待します。 receivers: jane.doe@example.com",
		// 	args: args{
		// 		ctx: context.Background(),
		// 		options: []cdomain.MailFailedQueryOption{
		// 			cdomain.MailFailedQueryOptionFunc().WithReceiver(lo.ToPtr("jane.doe@example.com")),
		// 		},
		// 	},
		// 	wantErr:   false,
		// 	wantTotal: int64(1),
		// 	want:      []domain.MailFailed{mockMailFailed2},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mailFailedMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.wantTotal, int64(len(got)), "GetAll() total = %v, wantTotal %v", len(got), tt.wantTotal)

			// json.RawMessage、jsonutils.JSONTimeタイプは比較から除外します。
			ignoreTypes := cmpopts.IgnoreTypes(json.RawMessage{}, jsonutils.JSONTime{})
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTypes), "GetAll() itmes = %v, want %v", got, tt.want)
		})
	}
}
