package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type UserTransactionApiController struct{}

func (c UserTransactionApiController) Init(g echoswagger.ApiGroup) {

	g.GET("", c.GetAll).
		AddParamQueryNested(SearchInput{}).
		SetDescription("ユーザーの入出金履歴を照会します。").
		SetSummary("ユーザーの入出金履歴照会").
		SetSecurity("Authorization")

	g.GET("/download", c.Download).
		AddParamQueryNested(SearchInput{}).
		SetDescription("ユーザーの入出金履歴CSVファイルをダウンロードします。").
		SetSummary("ユーザーの入出金履歴CSVファイルダウンロード").
		SetSecurity("Authorization")
}

func (UserTransactionApiController) GetAll(c echo.Context) error {
	var v UserTransactionSearchInput
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
		"userId":           v.UserId,
		"userName":         v.UserName,
		"transactionType":  v.TransactionType,
		"startAmount":      v.StartAmount,
		"endAmount":        v.EndAmount,
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("UserTransactionSearchInput")

	totalCount, items, err := domain.UserTransaction{}.GetAll(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.UserId, v.TransactionType, v.UserName, v.StartAmount, v.EndAmount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.deposit_detail", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

// 入出金履歴CSVダウンロード
func (UserTransactionApiController) Download(c echo.Context) error {

	var v UserTransactionSearchInput
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
		"userId":           v.UserId,
		"userName":         v.UserName,
		"transactionType":  v.TransactionType,
		"startAmount":      v.StartAmount,
		"endAmount":        v.EndAmount,
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("UserTransactionSearchInput")

	err = domain.UserTransaction{}.GetAllDownloadCsv(c.Request().Context(), c.Response().Writer, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.UserId, v.TransactionType, v.UserName, v.StartAmount, v.EndAmount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.deposit_detail", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}
