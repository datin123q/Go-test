package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"energyshares-job-scheduler/pkg/gmo_aozora"
	"energyshares-job-scheduler/pkg/timeutils"

	"github.com/antihax/optional"
	client "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type AccountApiControllers struct {
}

func (c AccountApiControllers) Init(g echoswagger.ApiGroup) {
	g.POST("/info", c.GetAccountInfo).
		AddResponse(http.StatusOK, "계좌 조회", nil, nil).
		SetSummary("(TEST) 계좌 정보 조회").
		SetDescription("계좌 정보 조회")

	g.POST("/deposit-transactions", c.AccountsDepositTransactionsUsingGET).AddParamBody(TransferInput{}, "body", "송금 추가", true).
		AddResponse(http.StatusOK, "송금 모델 추가 완료", domain.Transfer{}, nil).
		SetSummary("(TEST) 계좌 입금 내역 조회").
		SetDescription("계좌 입금 내역 조회")

	g.POST("/transactions", c.TransactionsUsingGET).AddParamBody(DateInput{}, "body", "송금 추가", true).
		AddResponse(http.StatusOK, "거래 내역 조회", nil, nil).
		SetSummary("(TEST) 계좌 거래 내역 조회").
		SetDescription("계좌 거래 내역 조회")
}

func (AccountApiControllers) GetAccountInfo(c echo.Context) error {
	err := gmo_aozora.GetAccountInfo(c.Request().Context(), gmo_aozora.BankAccounts)
	if err != nil {
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"count": len(gmo_aozora.BankAccounts),
			"error": err,
		}).Error(" Fail Account Info ")
		return factory.ErrorSystem.New(err)
	}
	return renderSucc(c, http.StatusOK, nil)
}

func (AccountApiControllers) AccountsDepositTransactionsUsingGET(c echo.Context) error {
	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}

	ctx := c.Request().Context()

	account, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_DEPOSIT)
	if err != nil {
		return factory.ErrorParameter.New(errors.Wrap(err, "fail account deposit transaction"))
	}
	xAccessToken := gmo_aozora.GetAccessToken()

	opts := client.AccountsDepositTransactionsUsingGETOpts{DateFrom: optional.NewString(timeutils.LocalZoneTime().Format(time.DateOnly))}

	res, _, err := gmo_aozora.CallWith429Retry(ctx, "AccountApi.AccountsDepositTransactionsUsingGET", func(ctx context.Context) (client.DepositTransactionsResponse, *http.Response, error) {
		return corporateClient.AccountApi.AccountsDepositTransactionsUsingGET(ctx, account.AccountId, xAccessToken, &opts)
	})
	if err != nil {
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"error": err,
		}).Error(" Fail Account Deposit Transaction ")
		return factory.ErrorParameter.New(err)
	}
	return renderSucc(c, http.StatusOK, res)
}

func (AccountApiControllers) TransactionsUsingGET(c echo.Context) error {
	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}

	ctx := c.Request().Context()

	account, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_DEPOSIT)
	if err != nil {
		return factory.ErrorParameter.New(errors.Wrap(err, "fail account deposit transaction"))
	}
	xAccessToken := gmo_aozora.GetAccessToken()

	opts := client.TransactionsUsingGETOpts{DateFrom: optional.NewString(time.Now().Format(time.DateOnly))}

	res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "AccountApi.TransactionsUsingGET", func(ctx context.Context) (client.TransactionsResponse, *http.Response, error) {
		return corporateClient.AccountApi.TransactionsUsingGET(ctx, account.AccountId, xAccessToken, &opts)
	})
	if err != nil {
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"error":          err,
			"httpStatusCode": gmo_aozora.HTTPStatusCode(httpResp),
			"errorDetail":    gmo_aozora.GetSwaggerErrorDetailString(err),
		}).Error(" Fail Transaction ")
		return factory.ErrorParameter.New(err)
	}
	return renderSucc(c, http.StatusOK, res)
}
