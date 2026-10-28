package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type myNumberHttpDelivery struct {
	myNumberUsecase udomain.MyNumberUsecase
	extractor       jwthelper.UserExtractor
}

func New(
	api echoswagger.ApiRoot,
	myNumberUsecase udomain.MyNumberUsecase,
	extractor jwthelper.UserExtractor,
) *myNumberHttpDelivery {
	handler := &myNumberHttpDelivery{
		myNumberUsecase: myNumberUsecase,
		extractor:       extractor,
	}

	g := api.Group("v1/my-number", "api/v1/my-number",
		echoutils.AllowRoles(domain.ROLE_USER))

	g.POST("", handler.Create).
		AddParamBody(udomain.MyNumberCreateRequest{}, "body", "My Number情報", true).
		AddResponse(http.StatusOK, "マイナンバー登録成功", udomain.MyNumberCreateResponse{}, nil).
		SetSummary("個人会員My Number登録").
		SetDescription("個人会員My Number登録（/private-users/signup/my-numberと機能は同じ）").
		SetSecurity("Authorization")

	g.GET("", handler.Get).
		AddResponse(http.StatusOK, "本人のMy Number照会成功", udomain.MyNumberGetResponse{}, nil).
		SetSummary("本人のMy Number照会").
		SetDescription("本人のMy Number照会").
		SetSecurity("Authorization")

	return handler
}

func (m myNumberHttpDelivery) Create(c echo.Context) error {
	user, err := m.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// Request DTO作成及びバインド
	var request = udomain.MyNumberCreateRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecaseを呼び出してエラーをハンドル
	response, err := m.myNumberUsecase.Create(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorMyNumberAlreadyExist.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (m myNumberHttpDelivery) Get(c echo.Context) error {
	user, err := m.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// Request DTO作成及びバインド
	var request = udomain.MyNumberGetRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecaseを呼び出してエラーをハンドル
	response, err := m.myNumberUsecase.Get(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorMyNumberAlreadyExist.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
