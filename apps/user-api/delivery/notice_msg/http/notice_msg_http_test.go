package http

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
)

func Test_noticeMsgHttpDelivery_GetAll(t *testing.T) {
	type fields struct {
		noticeMsgUsecase func() *mocks.NoticeMsgUsecase
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
			name: "GET /api/v1/users/notice-msgs 정상적으로 호출",
			fields: fields{
				noticeMsgUsecase: func() *mocks.NoticeMsgUsecase {
					usecase := &mocks.NoticeMsgUsecase{}
					usecase.On("GetAll", mock.Anything, udomain.NoticeMsgGetAllRequest{
						JWTUser: jwthelper.User{
							Id:    1000,
							Email: "hello@email.com",
							Name:  "hello",
							Roles: map[string]int{"ROLE_USER": 1},
						},
						ClientIp: "192.0.2.1",
					}).Return(udomain.NoticeMsgGetAllResponse{
						TotalCount: 1,
						Items: []domain.NoticeMsg{
							{
								Id:              1,
								Title:           "test",
								ContentHtml:     "<p>test</p>",
								NoticeDate:      jsonutils.JSONTime{},
								NoticeUserCount: 0,
								Created:         jsonutils.JSONTime{},
								NoticeUserType:  domain.NOTICE_NEWSLETTER,
								ProjectId:       0,
								IsRead:          true,
							},
						},
					}, nil)
					return usecase
				},
				extractor: jwthelper.NewUserExtractor(),
			},
			args: args{
				echoCtx: pkgtest.RequestContext(pkgtest.Request{
					Path:   "/api/v1/notice-msgs",
					Method: http.MethodGet,
					Claims: map[string]interface{}{
						"user":  "1000",
						"email": "hello@email.com",
						"name":  "hello",
						"role":  []string{"ROLE_USER"},
					},
				}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			delivery := noticeMsgHttpDelivery{
				noticeMsgUsecase: tt.fields.noticeMsgUsecase(),
				extractor:        tt.fields.extractor,
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAll(tt.args.echoCtx(e))
			assert.NoErrorf(t, err, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_noticeMsgHttpDelivery_GetOne(t *testing.T) {
	type fields struct {
		noticeMsgUsecase func() *mocks.NoticeMsgUsecase
		extractor        jwthelper.UserExtractor
	}
	type args struct {
		echoCtx func(e *echo.Echo) echo.Context
	}
	tests := []struct {
		name string
		fields
		args    args
		wantErr bool
	}{
		{
			name: "사용자 알림 정보를 요청 => 정상 응답",
			fields: fields{
				noticeMsgUsecase: func() *mocks.NoticeMsgUsecase {
					usecase := new(mocks.NoticeMsgUsecase)
					usecase.On("GetOne", mock.Anything, udomain.NoticeMsgGetOneRequest{
						Id: 1,
						JWTUser: jwthelper.User{
							Id:    111,
							Email: "gildong@email.com",
							Name:  "홍길동",
							Roles: map[string]int{"ROLE_USER": 1},
						},
						ClientIp: "192.0.2.1",
					}).Return(udomain.NoticeMsgGetOneResponse{NoticeMsg: mockNoticeMsg1}, nil)
					return usecase
				},
				extractor: jwthelper.NewUserExtractor(),
			},
			args: args{
				echoCtx: pkgtest.RequestContext(pkgtest.Request{
					Path:   "/api/v1/notice-msgs/:id",
					Method: http.MethodGet,
					Params: map[string]string{"id": "1"},
					Claims: map[string]interface{}{
						"user":  "111",
						"email": "gildong@email.com",
						"name":  "홍길동",
						"role":  []string{"ROLE_USER"},
					},
				}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := noticeMsgHttpDelivery{
				noticeMsgUsecase: tt.fields.noticeMsgUsecase(),
				extractor:        tt.fields.extractor,
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := d.GetOne(tt.args.echoCtx(e))
			assert.NoErrorf(t, err, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
