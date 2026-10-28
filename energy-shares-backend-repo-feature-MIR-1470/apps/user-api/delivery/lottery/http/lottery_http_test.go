package http

import (
	"context"
	"net/http"
	"testing"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
	"github.com/stretchr/testify/require"
)

const lotteryApi = "/api/v1/lottery"

type stubLotteryUsecase struct {
	getApplications func(ctx context.Context, request udomain.LotteryGetApplicationsRequest) (udomain.LotteryGetApplicationsResponse, error)
}

func (s stubLotteryUsecase) Application(ctx context.Context, request udomain.LotteryApplicationRequest) (udomain.LotteryApplicationResponse, error) {
	return udomain.LotteryApplicationResponse{}, nil
}

func (s stubLotteryUsecase) CheckApplication(ctx context.Context, request udomain.LotteryCheckApplicationRequest) (udomain.LotteryCheckApplicationResponse, error) {
	return udomain.LotteryCheckApplicationResponse{}, nil
}

func (s stubLotteryUsecase) Cancel(ctx context.Context, request udomain.LotteryCancelRequest) (udomain.LotteryCancelResponse, error) {
	return udomain.LotteryCancelResponse{}, nil
}

func (s stubLotteryUsecase) GetApplications(ctx context.Context, request udomain.LotteryGetApplicationsRequest) (udomain.LotteryGetApplicationsResponse, error) {
	return s.getApplications(ctx, request)
}

func TestLotteryHttpDelivery_Cancel(t *testing.T) {
	tests := []struct {
		name           string
		echoCtx        func(e *echo.Echo) echo.Context
		lotteryUsecase func() *mocks.LotteryUsecase
		wantErr        bool
	}{
		{
			name: "(成功) POST /api/v1/lottery/applications/:lottery_application_id/cancel",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   lotteryApi + "/applications/:lottery_application_id/cancel",
				Method: http.MethodPost,
				Params: map[string]string{"lottery_application_id": "app-001"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "user@email.com",
					"name":  "ユーザー",
					"role":  []string{"ROLE_USER"},
				},
			}),
			lotteryUsecase: func() *mocks.LotteryUsecase {
				usecase := &mocks.LotteryUsecase{}
				usecase.On("Cancel", mock.Anything, udomain.LotteryCancelRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "user@email.com",
						Name:  "ユーザー",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					LotteryApplicationID: "app-001",
				}).Return(udomain.LotteryCancelResponse{}, nil)
				return usecase
			},
		},
		{
			name: "(失敗) POST /api/v1/lottery/applications/:lottery_application_id/cancel - 対象が存在しない",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   lotteryApi + "/applications/:lottery_application_id/cancel",
				Method: http.MethodPost,
				Params: map[string]string{"lottery_application_id": "app-999"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "user@email.com",
					"name":  "ユーザー",
					"role":  []string{"ROLE_USER"},
				},
			}),
			lotteryUsecase: func() *mocks.LotteryUsecase {
				usecase := &mocks.LotteryUsecase{}
				usecase.On("Cancel", mock.Anything, udomain.LotteryCancelRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "user@email.com",
						Name:  "ユーザー",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					LotteryApplicationID: "app-999",
				}).Return(udomain.LotteryCancelResponse{}, cdomain.ErrItemNotFound)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "(失敗) POST /api/v1/lottery/applications/:lottery_application_id/cancel - キャンセル不可",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   lotteryApi + "/applications/:lottery_application_id/cancel",
				Method: http.MethodPost,
				Params: map[string]string{"lottery_application_id": "app-002"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "user@email.com",
					"name":  "ユーザー",
					"role":  []string{"ROLE_USER"},
				},
			}),
			lotteryUsecase: func() *mocks.LotteryUsecase {
				usecase := &mocks.LotteryUsecase{}
				usecase.On("Cancel", mock.Anything, udomain.LotteryCancelRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "user@email.com",
						Name:  "ユーザー",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					LotteryApplicationID: "app-002",
				}).Return(udomain.LotteryCancelResponse{}, cdomain.ErrBadRequest)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			delivery := lotteryHttpDelivery{
				lotteryUsecase: tt.lotteryUsecase(),
				extractor:      jwthelper.NewUserExtractor(),
			}

			err := delivery.Cancel(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Cancel() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_lotteryHttpDelivery_GetApplications(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		lotteryUsecase    stubLotteryUsecase
		wantErr           bool
		wantUsecaseCalled bool
	}{
		{
			name: "GET /api/v1/users/:user_id/lottery/applications binds path and query",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/users/:user_id/lottery/applications",
				Method: http.MethodGet,
				Params: map[string]string{
					"user_id": "1000",
				},
				QueryParams: map[string][]string{
					"limit":  {"2"},
					"offset": {"2"},
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "hello@email.com",
					"name":  "hello",
					"role":  []string{"ROLE_USER"},
				},
			}),
			lotteryUsecase: stubLotteryUsecase{
				getApplications: func(ctx context.Context, request udomain.LotteryGetApplicationsRequest) (udomain.LotteryGetApplicationsResponse, error) {
					require.Equal(t, int64(1000), request.UserID)
					require.Equal(t, 2, request.Limit)
					require.Equal(t, 2, request.Offset)
					require.Equal(t, int64(1000), request.JWTUser.Id)
					return udomain.LotteryGetApplicationsResponse{}, nil
				},
			},
			wantErr:           false,
			wantUsecaseCalled: true,
		},
		{
			name: "GET /api/v1/users/:user_id/lottery/applications rejects other user id",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/users/:user_id/lottery/applications",
				Method: http.MethodGet,
				Params: map[string]string{
					"user_id": "9999",
				},
				QueryParams: map[string][]string{
					"limit":  {"2"},
					"offset": {"0"},
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "hello@email.com",
					"name":  "hello",
					"role":  []string{"ROLE_USER"},
				},
			}),
			lotteryUsecase: stubLotteryUsecase{
				getApplications: func(ctx context.Context, request udomain.LotteryGetApplicationsRequest) (udomain.LotteryGetApplicationsResponse, error) {
					t.Fatalf("usecase should not be called")
					return udomain.LotteryGetApplicationsResponse{}, nil
				},
			},
			wantErr:           true,
			wantUsecaseCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := lotteryHttpDelivery{
				lotteryUsecase: tt.lotteryUsecase,
				extractor:      jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetApplications(tt.echoCtx(e))
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
