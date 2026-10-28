package http

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swaggest/openapi-go/openapi31"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

const investmentWaitersCancelPath = "/api/v1/users/:id/waitlist/:waitlist_id/cancel"

type stubInvestmentWaitersUsecase struct {
	cancel func(ctx context.Context, request udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error)
}

func (s stubInvestmentWaitersUsecase) Cancel(ctx context.Context, request udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
	return s.cancel(ctx, request)
}

func investmentWaitersCancelContext(userID, investmentWaiterID string, claims map[string]interface{}) func(e *echo.Echo) echo.Context {
	return pkgtest.RequestContext(pkgtest.Request{
		Path:   investmentWaitersCancelPath,
		Method: http.MethodPost,
		Params: map[string]string{
			"id":          userID,
			"waitlist_id": investmentWaiterID,
		},
		Header: pkgtest.Header{
			Headers:   echo.HeaderContentType,
			MIMETypes: echo.MIMEApplicationJSON,
		},
		Claims: claims,
	})
}

func validInvestmentWaitersClaims() map[string]interface{} {
	return map[string]interface{}{
		"user":  "1000",
		"email": "user@example.com",
		"name":  "テストユーザー",
		"role":  []string{"ROLE_USER"},
	}
}

func TestInvestmentWaitersHttpDelivery_Cancel(t *testing.T) {
	tests := []struct {
		name       string
		echoCtx    func(e *echo.Echo) echo.Context
		usecase    stubInvestmentWaitersUsecase
		wantStatus int
		wantBody   string
	}{
		{
			name:    "本人のキャンセル待ち申請を200空ボディで取り消す",
			echoCtx: investmentWaitersCancelContext("1000", "2000", validInvestmentWaitersClaims()),
			usecase: stubInvestmentWaitersUsecase{cancel: func(_ context.Context, request udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
				assert.Equal(t, int64(1000), request.UserID)
				assert.Equal(t, int64(2000), request.InvestmentWaiterID)
				assert.Equal(t, "192.0.2.1", request.ClientIp)
				assert.Equal(t, jwthelper.User{
					Id:    1000,
					Email: "user@example.com",
					Name:  "テストユーザー",
					Roles: map[string]int{"ROLE_USER": 1},
				}, request.JWTUser)
				return udomain.InvestmentWaitersCancelResponse{}, nil
			}},
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
		{
			name:    "ユーザーIDが不正な場合は400を返す",
			echoCtx: investmentWaitersCancelContext("invalid", "2000", validInvestmentWaitersClaims()),
			usecase: stubInvestmentWaitersUsecase{cancel: func(context.Context, udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
				t.Fatal("Usecaseが呼び出されてはいけません")
				return udomain.InvestmentWaitersCancelResponse{}, nil
			}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "キャンセル待ち申請IDが不正な場合は400を返す",
			echoCtx: investmentWaitersCancelContext("1000", "invalid", validInvestmentWaitersClaims()),
			usecase: stubInvestmentWaitersUsecase{cancel: func(context.Context, udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
				t.Fatal("Usecaseが呼び出されてはいけません")
				return udomain.InvestmentWaitersCancelResponse{}, nil
			}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "キャンセル待ち申請IDが0の場合は400を返す",
			echoCtx: investmentWaitersCancelContext("1000", "0", validInvestmentWaitersClaims()),
			usecase: stubInvestmentWaitersUsecase{cancel: func(context.Context, udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
				t.Fatal("Usecaseが呼び出されてはいけません")
				return udomain.InvestmentWaitersCancelResponse{}, nil
			}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "JWTユーザーを取得できない場合は401を返す",
			echoCtx: investmentWaitersCancelContext("1000", "2000", map[string]interface{}{}),
			usecase: stubInvestmentWaitersUsecase{cancel: func(context.Context, udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
				t.Fatal("Usecaseが呼び出されてはいけません")
				return udomain.InvestmentWaitersCancelResponse{}, nil
			}},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "対象が存在しない場合は404を返す",
			echoCtx: investmentWaitersCancelContext("1000", "2000", validInvestmentWaitersClaims()),
			usecase: stubInvestmentWaitersUsecase{cancel: func(context.Context, udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
				return udomain.InvestmentWaitersCancelResponse{}, cdomain.ErrItemNotFound
			}},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "権限不正の場合は403を返す",
			echoCtx: investmentWaitersCancelContext("1000", "2000", validInvestmentWaitersClaims()),
			usecase: stubInvestmentWaitersUsecase{cancel: func(context.Context, udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
				return udomain.InvestmentWaitersCancelResponse{}, cdomain.ErrForbidden
			}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:    "状態不正の場合は400を返す",
			echoCtx: investmentWaitersCancelContext("1000", "2000", validInvestmentWaitersClaims()),
			usecase: stubInvestmentWaitersUsecase{cancel: func(context.Context, udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
				return udomain.InvestmentWaitersCancelResponse{}, cdomain.ErrBadRequest
			}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "予期しないエラーの場合は500を返す",
			echoCtx: investmentWaitersCancelContext("1000", "2000", validInvestmentWaitersClaims()),
			usecase: stubInvestmentWaitersUsecase{cancel: func(context.Context, udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
				return udomain.InvestmentWaitersCancelResponse{}, assert.AnError
			}},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			ctx := tt.echoCtx(e)
			delivery := investmentWaitersHttpDelivery{
				investmentWaitersUsecase: tt.usecase,
				extractor:                jwthelper.NewUserExtractor(),
			}

			err := delivery.Cancel(ctx)
			if tt.wantStatus == http.StatusOK {
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, ctx.Response().Status)
				recorder, ok := ctx.Response().Writer.(*httptest.ResponseRecorder)
				require.True(t, ok)
				assert.Equal(t, tt.wantBody, recorder.Body.String())
				return
			}

			require.Error(t, err)
			var factoryErr *factory.Error
			require.True(t, stderrors.As(err, &factoryErr), "error = %v", err)
			assert.Equal(t, tt.wantStatus, factoryErr.Status())
		})
	}
}

func TestInvestmentWaitersHttpDelivery_OpenAPI(t *testing.T) {
	e := echo.New()
	reflector := openapi31.NewReflector()
	reflector.Spec.SetHTTPBearerTokenSecurity("Authorization", "JWT", "JWTアクセストークン")
	api := echoutils.NewEchoSwagger(e, reflector)

	New(api, stubInvestmentWaitersUsecase{cancel: func(context.Context, udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
		return udomain.InvestmentWaitersCancelResponse{}, nil
	}}, jwthelper.NewUserExtractor())
	api.RegisterOpenAPI()

	data, err := json.Marshal(reflector.Spec)
	require.NoError(t, err)
	var document map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &document))

	paths, ok := document["paths"].(map[string]interface{})
	require.True(t, ok)
	path, ok := paths["/api/v1/users/{id}/waitlist/{waitlist_id}/cancel"].(map[string]interface{})
	require.True(t, ok)
	post, ok := path["post"].(map[string]interface{})
	require.True(t, ok)

	parameters, ok := post["parameters"].([]interface{})
	require.True(t, ok)
	parameterNames := make(map[string]bool, len(parameters))
	for _, parameter := range parameters {
		parameterMap, ok := parameter.(map[string]interface{})
		require.True(t, ok)
		name, _ := parameterMap["name"].(string)
		parameterNames[name] = true
		assert.Equal(t, "path", parameterMap["in"])
		assert.Equal(t, true, parameterMap["required"])
	}
	assert.True(t, parameterNames["id"])
	assert.True(t, parameterNames["waitlist_id"])

	responses, ok := post["responses"].(map[string]interface{})
	require.True(t, ok)
	for _, status := range []string{"200", "400", "401", "403", "404", "500"} {
		assert.Contains(t, responses, status)
	}
	successResponse, ok := responses["200"].(map[string]interface{})
	require.True(t, ok)
	assert.NotContains(t, successResponse, "content")

	security, ok := post["security"].([]interface{})
	require.True(t, ok)
	require.NotEmpty(t, security)
}
