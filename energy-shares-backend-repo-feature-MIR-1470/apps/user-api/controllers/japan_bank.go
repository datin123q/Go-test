package controllers

import (
	"net/http"
	"regexp"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

type JapanBankApiController struct {
	bankCodeJPService service.IBankCodeJPService
}

func NewJapanBankApiController() *JapanBankApiController {
	return &JapanBankApiController{
		bankCodeJPService: service.NewBankCodeJPService(),
	}
}

const MAX_LIMIT = 2000

func (c JapanBankApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetBanks).
		AddParamQueryNested(JapanBankSearchRequest{}).
		AddResponse(http.StatusOK, "日本銀行リスト照会", JapanBankResponse{}, nil).
		SetDescription("日本銀行リストを照会します。").
		SetSummary("日本銀行リスト照会")

	g.GET("/:code/branches", c.GetBranches).
		AddParamPath("", "code", "銀行コード").
		AddParamQueryNested(JapanBankBranchSearchRequest{}).
		AddResponse(http.StatusOK, "日本銀行支店リスト照会", JapanBankBranchResponse{}, nil).
		SetDescription("日本銀行の支店リストを照会します。").
		SetSummary("日本銀行支店リスト照会")
}

func (api JapanBankApiController) GetBanks(c echo.Context) error {

	var v JapanBankSearchRequest
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.Limit == 0 {
		v.Limit = MAX_LIMIT
	}

	re := regexp.MustCompile(`\s+`) // 空白削除
	v.Keyword = re.ReplaceAllString(v.Keyword, "")

	if len(v.Keyword) == 0 {
		return apirender.RenderSucc(c, http.StatusOK, JapanBankResponse{
			Items:      nil,
			TotalCount: 0,
			HasNext:    false,
			NextCursor: nil,
		})
	}

	result, err := api.bankCodeJPService.GetBanksWithFreeword(v.Keyword, v.Limit, v.Cursor)
	if err != nil {
		return factory.ErrorServiceUnavailable.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, JapanBankResponse{}.New(*result))
}

func (api JapanBankApiController) GetBranches(c echo.Context) error {

	var v JapanBankBranchSearchRequest
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.Limit == 0 {
		v.Limit = MAX_LIMIT
	}

	bankCode := c.Param("code")

	re := regexp.MustCompile(`\s+`) // 空白削除
	v.Keyword = re.ReplaceAllString(v.Keyword, "")

	if len(v.Keyword) == 0 {
		return apirender.RenderSucc(c, http.StatusOK, JapanBankBranchResponse{
			Items:      nil,
			TotalCount: 0,
			HasNext:    false,
			NextCursor: nil,
		})
	}

	result, err := api.bankCodeJPService.GetBranchesWithFreeword(bankCode, v.Keyword, v.Limit, v.Cursor)
	if err != nil {
		return factory.ErrorServiceUnavailable.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, JapanBankBranchResponse{}.New(*result))
}
