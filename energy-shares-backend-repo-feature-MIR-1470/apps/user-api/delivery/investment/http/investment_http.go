package http

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type investmentHttpDelivery struct {
	investmentUsecase udomain.InvestmentUsecase
	extractor         jwthelper.UserExtractor
}

func New(
	api echoswagger.ApiRoot,
	investmentUsecase udomain.InvestmentUsecase,
	extractor jwthelper.UserExtractor,
) *investmentHttpDelivery {
	handler := &investmentHttpDelivery{
		investmentUsecase: investmentUsecase,
		extractor:         extractor,
	}

	g := api.Group("v1/investments", "api/v1/investments",
		echoutils.AllowRoles(domain.ROLE_USER))

	g.GET("/:id/result", handler.GetResult).
		AddParamPath("", "id", "投資ID（取引番号）").
		SetDescription("投資結果ページに必要な情報をリクエストします。").
		SetSummary("投資結果情報返却").
		SetSecurity("Authorization")

	return handler
}

func (i investmentHttpDelivery) GetResult(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザー情報抽出
	user, err := i.extractor.Extract(c)
	if err != nil {
		return err
	}

	// Request DTO作成
	var request = udomain.InvestmentGetResultRequest{
		JWTUser: user,
	}

	// Request DTOバインド
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := i.investmentUsecase.GetResult(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
