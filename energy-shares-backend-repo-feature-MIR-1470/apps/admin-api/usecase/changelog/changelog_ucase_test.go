package changelog

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

func Test_changelogUsecase_GetAll(t *testing.T) {
	type fields struct {
		cfg                 *config.Config
		changelogRepository func() *mocks.ChangelogRepository
	}
	type args struct {
		ctx     context.Context
		request adomain.ChangelogGetAllRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    adomain.ChangelogGetAllResponse
		wantErr bool
	}{
		{
			name: "変更ログリスト照会への応答",
			fields: fields{
				cfg: &config.Config{},
				changelogRepository: func() *mocks.ChangelogRepository {
					repo := new(mocks.ChangelogRepository)
					repo.On("GetAll", context.Background(),
						mock.AnythingOfType("func(*clean_arch_domain.ChangelogQueryOptionFields)"),
						mock.AnythingOfType("func(*clean_arch_domain.ChangelogQueryOptionFields)"),
						mock.AnythingOfType("func(*clean_arch_domain.ChangelogQueryOptionFields)"),
						mock.AnythingOfType("func(*clean_arch_domain.ChangelogQueryOptionFields)"),
					).Return([]domain.ChangelogRow{
						{
							Id:      133,
							Op:      "u",
							Updated: "2024-10-08T08:03:10.000",
							Actor: domain.ChangelogActor{
								Id:    1,
								Name:  "admin",
								Email: "admin@energyshares.jp",
							},
							Columns: map[string]interface{}{
								"CONTENT": map[string]interface{}{
									"AFTER":  "test11",
									"BEFORE": "test10",
								},
							},
						},
					}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: adomain.ChangelogGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    1,
						Email: "admin@energyshares.jp",
						Name:  "admin",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR): 1,
							string(domain.ROLE_ADMIN):    1,
						},
					},
					ClientIp: "127.0.0.1",
					ChangelogGetAllQueryParam: adomain.ChangelogGetAllQueryParam{
						KsqlSearchInput: cdomain.KsqlSearchInput{
							StartDate:      lo.ToPtr("2024-10-01 00:00:00"),
							EndDate:        lo.ToPtr("2024-11-02 16:47:00"),
							MaxResultCount: 1,
						},
						Id: lo.ToPtr("133"),
					},
				},
			},
			want: adomain.ChangelogGetAllResponse{
				Items: []domain.ChangelogRow{
					{
						Id:      133,
						Op:      "u",
						Updated: "2024-10-08T08:03:10.000",
						Actor: domain.ChangelogActor{
							Id:    1,
							Name:  "admin",
							Email: "admin@energyshares.jp",
						},
						Columns: map[string]interface{}{
							"CONTENT": map[string]interface{}{
								"AFTER":  "test11",
								"BEFORE": "test10",
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := changelogUsecase{
				cfg:                 tt.fields.cfg,
				changelogRepository: tt.fields.changelogRepository(),
			}

			got, err := u.GetAll(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, got, tt.want, "GetAll() got = %v, want %v", got, tt.want)
		})
	}
}
