package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"energyshares-job-scheduler/jobs"
	"energyshares-job-scheduler/pkg/gmo_aozora"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"

	"github.com/antihax/optional"
	client "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type VaAccountApiControllers struct {
}

func (c VaAccountApiControllers) Init(g echoswagger.ApiGroup) {
	g.POST("/va-deposit/execute", c.ExecuteAccountVaDeposit).
		AddParamBody(VaDepositInput{}, "body", "仮想口座入金内訳追加", true).
		SetSummary("(TEST) 仮想口座預金入金実行 (QAテスト用)").
		SetDescription("仮想口座預金入金実行 (QAテスト用)")

	g.POST("/job/execute", c.GetDepositCheck).
		AddResponse(http.StatusOK, "入金内訳照会実行", nil, nil).
		SetSummary("入金内訳照会実行").
		SetDescription("入金内訳照会実行。顧客仮想口座入金確認に使用")

	g.POST("/va-issue-job/execute", c.VaIssue).
		AddResponse(http.StatusOK, "仮想口座発行", nil, nil).
		SetSummary("仮想口座追加発行実行").
		SetDescription("仮想口座追加発行実行。仮想口座を大量に発行する際に使用")

	g.POST("/va-deposits", c.CreateVaDeposit).
		AddParamBody(VaDepositInput{}, "body", "仮想口座入金内訳", true).
		AddResponse(http.StatusOK, "仮想口座入金内訳追加完了", domain.VaDeposit{}, nil).
		SetSummary("(TEST) 仮想口座入金内訳追加").
		SetDescription("仮想口座入金内訳追加")

	g.POST("/change-status", c.VaStatusChange).
		AddParamBody(VaSearchInput{}, "body", "仮想口座ステータス変更申請リクエスト ", true).
		AddResponse(http.StatusOK, "仮想口座ステータス変更完了", nil, nil).
		SetSummary("仮想口座ステータス変更申請").
		SetDescription("仮想口座ステータス変更申請")
}

func (VaAccountApiControllers) GetDepositCheck(c echo.Context) error {
	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{}).Info("get deposit check start")

	a := &jobs.AozoraJobs{}
	err := a.GetEquityVaTransactionJob(c.Request().Context())
	if err != nil {
		factory.Logger(c.Request().Context()).Error("fail get va transactions")
		return factory.ErrorDB.New(err)
	}
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{}).Info("get deposit check end")
	return renderSucc(c, http.StatusOK, nil)
}

func (VaAccountApiControllers) CreateVaDeposit(c echo.Context) error {
	var v VaDepositInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	d, err := v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	_, err = d.Create(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return renderSucc(c, http.StatusOK, d)
}

func (VaAccountApiControllers) VaStatusChange(c echo.Context) error {
	var v VaSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}

	if err != nil {
		return factory.ErrorSystem.New(err)
	}
	vaIdList := []client.VaId{
		{VaId: v.VaId},
	}

	body := client.VaStatusChangeRequest{VaStatusChangeCode: v.VaStatusCode, VaIdList: vaIdList}
	opts := client.StatusChangeUsingPOSTOpts{Body: optional.NewInterface(body)}

	res, httpResp, err := gmo_aozora.CallWith429Retry(c.Request().Context(), "VirtualAccountApi.StatusChangeUsingPOST", func(ctx context.Context) (client.VaStatusChangeResponse, *http.Response, error) {
		return corporateClient.VirtualAccountApi.StatusChangeUsingPOST(ctx, gmo_aozora.GetAccessToken(), &opts)
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"vaId":           v.VaId,
			"vaStatusCode":   v.VaStatusCode,
			"httpStatusCode": gmo_aozora.HTTPStatusCode(httpResp),
			"errDetail":      gmo_aozora.GetSwaggerErrorDetailString(err),
			"error":          err,
			"time":           time.Now().String(),
		}).Error("Fail Va List  ")
		return err
	}

	return renderSucc(c, http.StatusOK, res)
}

func (VaAccountApiControllers) VaIssue(c echo.Context) error {
	a := &jobs.AozoraJobs{}
	err := a.IssueVirtualAccountJob(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return renderSucc(c, http.StatusOK, nil)
}

func (VaAccountApiControllers) ExecuteAccountVaDeposit(c echo.Context) error {
	var v VaDepositInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	// Prod 에서는 불가능하도록 하는 코드 절대 지우지 말 것
	if serviceutils.GetEnv() == "prod" {
		return factory.ErrorParameter.New(errors.New("not allowed"))
	}
	logrus.WithFields(logrus.Fields{
		"vaId":              v.VaId,
		"user_id":           v.UserId,
		"transfer_date":     v.TransactionDate,
		"deposit_amount":    v.DepositAmount,
		"deposit_date":      v.DepositDate,
		"deposit_status":    v.DepositStatus,
		"item_key":          v.ItemKey,
		"remarks":           v.Remarks,
		"va_account_number": v.VaAccountNumber,
	}).Warning("execute account va deposit start")
	d, err := v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	_, err = d.Create(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	err = d.DepositEquityProcess(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(errors.Wrap(err, "fail deposit check process"))
	}
	return renderSucc(c, http.StatusOK, d)
}
