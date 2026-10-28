package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-share-core-api/service"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type UserVirtaulAccountApiController struct {
	userService *service.UserAdminService
}

func (c UserVirtaulAccountApiController) Init(g echoswagger.ApiGroup) {
	c.userService = service.NewUserAdminService()
	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_BASE)).
		AddParamQueryNested(VirtualAccountSearchInput{}).
		SetDescription("仮想口座リストを照会します。").
		SetSummary("仮想口座リスト照会").
		SetSecurity("Authorization")

	g.GET("/check-reserve", c.CheckReserve,
		echoutils.AllowRoles(domain.ROLE_ADMIN)).
		AddResponse(http.StatusOK, "ユーザー仮想口座アラート照会", nil, nil).
		SetDescription("仮想口座の不足有無を照会します。").
		SetSummary("仮想口座不足有無照会").
		SetSecurity("Authorization")

	g.POST("/upload-loan-account-csv", c.UploadLoanVirtualAccount,
		echoutils.AllowRoles(domain.ROLE_REPORT)).
		AddResponse(http.StatusOK, "仮想口座のCSVファイルアップロードします。.", nil, nil).
		AddParamFile("virtual_account_upload_file", "仮想口座アップロードファイル", true).
		SetDescription("仮想口座のCSVファイルアップロード。").
		SetSummary("仮想口座のCSVファイルアップロードします。").
		SetSecurity("Authorization")

	g.POST("/upload-equity-account-csv", c.UploadEquityVirtualAccount,
		echoutils.AllowRoles(domain.ROLE_REPORT)).
		AddResponse(http.StatusOK, "仮想口座のCSVファイルアップロードします。.", nil, nil).
		AddParamFile("virtual_account_upload_file", "仮想口座アップロードファイル", true).
		SetDescription("仮想口座のCSVファイルアップロード。").
		SetSummary("仮想口座のCSVファイルアップロードします。").
		SetSecurity("Authorization")

	g.GET("/download", c.Download,
		echoutils.AllowRoles(domain.ROLE_BASE)).
		AddParamQueryNested(VirtualAccountSearchInput{}).
		AddResponse(http.StatusOK, "仮想口座リストをダウンロードします。.", nil, nil).
		SetDescription("仮想口座リストをダウンロードします。").
		SetSummary("仮想口座リストをダウンロードします。").
		SetSecurity("Authorization")

}

func (UserVirtaulAccountApiController) GetAll(c echo.Context) error {
	var v VirtualAccountSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}
	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	totalCount, items, err := domain.UserVirtualAccount{}.GetAll(c.Request().Context(), v.VaBranchCode, v.VaAccountNumber, v.LinkStatus, v.UserId, v.ProjectType, v.StartDate, v.EndDate, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_virtualaccount_manage", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (UserVirtaulAccountApiController) CheckReserve(c echo.Context) error {
	isNotSufficient, err := domain.UserVirtualAccount{}.CheckIfNotSufficient(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, isNotSufficient)
}

func (u UserVirtaulAccountApiController) UploadLoanVirtualAccount(c echo.Context) error {
	var items []domain.UserVirtualAccount

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	formFile, err := c.FormFile("virtual_account_upload_file")
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}
	file, err := formFile.Open()
	if err != nil {
		return factory.ErrorNotSupportedFiles.New(err)
	}

	lines, err := csvutils.ReadMultpartCsv(&file)

	for i, line := range lines {
		if i == 0 {
			if "支店" != line[0] {
				return factory.ErrorNotSupportedFiles.New(err)
			}
			if "口座番号" != line[2] {
				return factory.ErrorNotSupportedFiles.New(err)
			}
			break
		}
	}

	if err != nil {
		return factory.ErrorNotSupportedFiles.New(err)
	}
	_, err = u.userService.UploadVaCsv(c.Request().Context(), domain.LOAN, lines)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_virtualaccount_manage", "", c.RealIP(), domain.UPLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, items)
}
func (u UserVirtaulAccountApiController) UploadEquityVirtualAccount(c echo.Context) error {
	var items []domain.UserVirtualAccount

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	formFile, err := c.FormFile("virtual_account_upload_file")
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}
	file, err := formFile.Open()
	if err != nil {
		return factory.ErrorNotSupportedFiles.New(err)
	}

	lines, err := csvutils.ReadMultpartCsv(&file)

	for i, line := range lines {
		if i == 0 {
			if "支店" != line[0] {
				return factory.ErrorNotSupportedFiles.New(err)
			}
			if "口座番号" != line[2] {
				return factory.ErrorNotSupportedFiles.New(err)
			}
			break
		}
	}

	if err != nil {
		return factory.ErrorNotSupportedFiles.New(err)
	}
	_, err = u.userService.UploadVaCsv(c.Request().Context(), domain.EQUITY, lines)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_virtualaccount_manage", "", c.RealIP(), domain.UPLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, items)
}

func (UserVirtaulAccountApiController) Download(c echo.Context) error {
	var v VirtualAccountSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}
	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby": v.Sortby,
		"order":  v.Order,
	}).Info("SearchStart")

	err = domain.UserVirtualAccount{}.DownloadCsv(c.Request().Context(), c.Response().Writer, v.VaBranchCode, v.VaAccountNumber, v.ProjectType, v.LinkStatus, v.UserId, v.StartDate, v.EndDate, v.Sortby, v.Order)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_virtualaccount_manage", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}
