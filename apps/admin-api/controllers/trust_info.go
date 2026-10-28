package controllers

import (
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type TrustInfoApiController struct{}

func (c TrustInfoApiController) Init(g echoswagger.ApiGroup) {

	g.GET("/download", c.Download).
		AddParamQueryNested(TrustInfoSearchInput{}).
		SetDescription("信託情報リストのCSVファイルをダウンロードします。").
		SetSummary("信託情報リストのCSVファイルダウンロード").
		SetSecurity("Authorization")

}

func (TrustInfoApiController) Download(c echo.Context) error {

	var v TrustInfoSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	err = domain.TrustInfo{}.DownloadCsv(c.Request().Context(), c.Response().Writer, v.Sortby, v.Order, v.UserId, v.Name, v.Email, v.StartDate, v.EndDate)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 出力処理ログ
	_, err = domain.TrustInfoHistory{}.Create(c.Request().Context(), userId, string(domain.PRINT_STATUS_COMPLETE), v.StartDate, v.EndDate)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.trust_info.list", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}
