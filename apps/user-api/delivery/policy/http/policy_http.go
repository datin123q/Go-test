package http

import (
	"net/http"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/labstack/echo/v4"
	openapigo "github.com/swaggest/openapi-go"
)

const (
	_DefaultMaxResultCount = 30
)

type policyHttpDelivery struct {
	policyUsecase udomain.PolicyUsecase
}

func New(
	api *echoutils.EchoSwagger,
	policyUsecase udomain.PolicyUsecase,
) *policyHttpDelivery {
	handler := &policyHttpDelivery{
		policyUsecase: policyUsecase,
	}

	g := api.Group("/api/v1/policies")

	g.GET("", handler.GetAll).
		SetTags("利用規約").
		SetSummary("利用規約のリストを照会します").
		AddRequest(udomain.PolicyGetAllRequest{}).
		AddResponse(udomain.PolicyGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "利用規約リスト応答値"
		}).
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorDB,
		)

	g.GET("/histories", handler.GetAllHistory).
		SetTags("利用規約").
		SetSummary("利用規約の履歴リストを照会します").
		AddRequest(udomain.PolicyGetAllHistoryRequest{}).
		AddResponse(udomain.PolicyGetAllHistoryResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "利用規約履歴リスト応答値"
		}).
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorDB,
		)

	return handler
}

func (h *policyHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = udomain.PolicyGetAllRequest{
		PolicyGetAllQueryParam: udomain.PolicyGetAllQueryParam{
			SearchInput: cdomain.SearchInput{
				MaxResultCount: _DefaultMaxResultCount,
			},
		},
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := h.policyUsecase.GetAll(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (h *policyHttpDelivery) GetAllHistory(c echo.Context) error {
	// (参考) ユーザー利用規約変更履歴照会 使用中のもの: 2個
	// - 利用規約/footer: POLICY_CAUTION/POLICY_VIEW_FOOTER
	// - 個人情報保護方針/footer: POLICY_PROTECTION/POLICY_VIEW_FOOTER

	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = udomain.PolicyGetAllHistoryRequest{
		PolicyGetAllHistoryQueryParam: udomain.PolicyGetAllHistoryQueryParam{
			SearchInput: cdomain.SearchInput{
				MaxResultCount: _DefaultMaxResultCount,
			},
		},
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := h.policyUsecase.GetAllHistory(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}
