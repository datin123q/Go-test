package http

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

func Test_qnaHttpDelivery_GetChangelogs(t *testing.T) {
	type fields struct {
		cfg              *config.Config
		changelogUsecase func() *mocks.ChangelogUsecase
		extractor        jwthelper.UserExtractor
	}
	type args struct {
		echoCtx func(e *echo.Echo) echo.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "변경로그 목록 조회",
			fields: fields{
				cfg: &config.Config{CommonConfig: config.CommonConfig{Env: "dev"}},
				changelogUsecase: func() *mocks.ChangelogUsecase {
					usecase := new(mocks.ChangelogUsecase)
					usecase.On("GetAll", mock.Anything, adomain.ChangelogGetAllRequest{
						JWTUser: jwthelper.User{
							Id:    1,
							Email: "admin@energyshares.jp",
							Name:  "admin",
							Roles: map[string]int{
								string(domain.ROLE_OPERATOR): 1,
								string(domain.ROLE_ADMIN):    1,
							},
						},
						ClientIp: "192.0.2.1",
						ChangelogGetAllQueryParam: adomain.ChangelogGetAllQueryParam{
							KsqlSearchInput: cdomain.KsqlSearchInput{
								StartDate:      lo.ToPtr("2024-10-01 00:00:00"),
								EndDate:        lo.ToPtr("2024-11-02 16:47:00"),
								MaxResultCount: 10,
							},
							Id: lo.ToPtr("133"),
						},
					}).Return(adomain.ChangelogGetAllResponse{
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
					}, nil)
					return usecase
				},
			},
			args: args{
				echoCtx: pkgtest.RequestContext(pkgtest.Request{
					Path:   "/api/v1/qna/changelogs",
					Method: http.MethodGet,
					Claims: map[string]interface{}{
						"user":  "1",
						"email": "admin@energyshares.jp",
						"name":  "admin",
						"role":  []string{"ROLE_OPERATOR", "ROLE_ADMIN"},
					},
					QueryParams: map[string][]string{
						"start_date":       {"2024-10-01 00:00:00"},
						"end_date":         {"2024-11-02 16:47:00"},
						"max_result_count": {"10"},
						"id":               {"133"},
					},
				}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := qnaHttpDelivery{
				cfg:              tt.fields.cfg,
				changelogUsecase: tt.fields.changelogUsecase(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetChangelogs(tt.args.echoCtx(e))
			assert.NoErrorf(t, err, "GetChangelogs() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
