package controllers

import (
	"context"
	"energyshares-job-scheduler/jobs"
	"energyshares-job-scheduler/pkg/gmo_aozora"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/antihax/optional"
	client "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/feequote"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/gmoaozorautils"
)

const (
	gmoAozoraBankCode            = "0310"
	gmoAozoraFallbackTransferFee = 0
	otherBankFallbackTransferFee = 99
)

type BulkTransferApiControllers struct{}

// 複数件手数料見積り入力
type BulkTransferFeeInputs struct {
	Transfers []TransferInput `json:"transfers"`
}

// convertFeeResponseToDTO maps GMO fee response (itemId) to internal DTO (withdrawalId).
func convertFeeResponseToDTO(res client.TransferFeeResponse, itemToWithdrawal map[string]string) (feequote.WithdrawalFeeBulkResponse, error) {
	var conv feequote.WithdrawalFeeBulkResponse

	totalFee, err := strconv.ParseInt(res.TotalFee, 10, 64)
	if err != nil {
		return conv, errors.Wrap(err, "fail parse totalFee")
	}
	conv.TotalFee = totalFee
	conv.Details = make([]feequote.WithdrawalFeeDetail, 0, len(res.TransferFeeDetails))

	for i := range res.TransferFeeDetails {
		wid, ok := itemToWithdrawal[res.TransferFeeDetails[i].ItemId]
		if !ok {
			return conv, errors.Errorf("missing withdrawalId for itemId=%s", res.TransferFeeDetails[i].ItemId)
		}
		widInt, err := strconv.ParseInt(wid, 10, 64)
		if err != nil {
			return conv, errors.Wrap(err, "fail parse withdrawalId")
		}
		feeInt, err := strconv.ParseInt(res.TransferFeeDetails[i].TransferFee, 10, 64)
		if err != nil {
			return conv, errors.Wrap(err, "fail parse transferFee")
		}
		conv.Details = append(conv.Details, feequote.WithdrawalFeeDetail{
			WithdrawalID: widInt,
			TransferFee:  feeInt,
		})
	}
	return conv, nil
}

func (c BulkTransferApiControllers) Init(g echoswagger.ApiGroup) {
	g.POST("/transfer-job/execute", c.TransferJobExecute).
		AddResponse(http.StatusOK, "일괄 송금 요청 잡 실행", domain.Transfer{}, nil).
		SetSummary("(TEST) 일괄 송금 요청").
		SetDescription("일괄 송금 요청")

	g.POST("/transfer-result-job/execute", c.TransferResultJobExecute).
		AddResponse(http.StatusOK, "일괄 결과 조회 잡 실행", domain.Transfer{}, nil).
		SetSummary("(TEST) 일괄 송금 결과 조회").
		SetDescription("일괄 송금 결과 조회")

	g.POST("/request", c.GetTransferRequests).
		AddParamBody(TransferInput{}, "body", "일괄 송금 추가", true).
		AddResponse(http.StatusOK, "일괄 송금 요청 완료", domain.Transfer{}, nil).
		SetSummary("(TEST) 일괄 송금 요청").
		SetDescription("일괄 송금 요청")

	g.POST("/fee", c.GetTransferRequestsFee).
		AddParamBody(TransferInput{}, "body", "単件の送金手数料を事前見積り", true).
		AddResponse(http.StatusOK, "単件の送金手数料を事前見積り", nil, nil).
		SetSummary("(TEST) 単件送金手数料見積り").
		SetDescription("単件の送金手数料を事前見積りします。")

	// 複数件手数料見積り
	g.POST("/fee-bulk", c.GetTransferRequestsFeeBulk).
		AddParamBody(BulkTransferFeeInputs{}, "body", "複数件の送金手数料をまとめて事前見積り", true).
		AddResponse(http.StatusOK, "複数件の送金手数料をまとめて事前見積り", nil, nil).
		SetSummary("(TEST) 複数件送金手数料見積り").
		SetDescription("複数件の送金手数料をまとめて事前見積りします。")

	g.POST("/result", c.TransferResultCheck).
		AddParamBody(BulkTransferCancelInput{}, "body", "일괄 송금 추가", true).
		AddResponse(http.StatusOK, "일괄 송금 결과 조회", domain.Transfer{}, nil).
		SetSummary("(TEST) 일괄 송금 결과 조회").
		SetDescription("일괄 송금 결과 조회")

	g.POST("/cancel", c.CancelBulkTransfer).
		AddParamBody(BulkTransferCancelInput{}, "body", "일괄 송금 취소", true).
		AddResponse(http.StatusOK, "일괄 송금 취소", nil, nil).
		SetSummary("(TEST) 일괄 송금 취소").
		SetDescription("일괄 송금 취소")

	g.POST("/status", c.GetStatusBulkTransfer).
		AddParamBody(BulkTransferStatusInput{}, "body", "일괄 송금 상태 조회", true).
		AddResponse(http.StatusOK, "일괄 송금 상태 조회", nil, nil).
		SetSummary("(TEST) 일괄 송금 상태 조회").
		SetDescription("일괄 송금 상태 조회")
}

func (BulkTransferApiControllers) TransferJobExecute(c echo.Context) error {
	a := &jobs.AozoraJobs{}
	err := a.BulkTransferJob(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return renderSucc(c, http.StatusOK, nil)
}

func (BulkTransferApiControllers) TransferResultJobExecute(c echo.Context) error {
	a := &jobs.AozoraJobs{}
	err := a.BulkTransferResultCheckJob(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return renderSucc(c, http.StatusOK, nil)
}

func (BulkTransferApiControllers) TransferResultCheck(c echo.Context) error {
	var v BulkTransferCancelInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}

	ctx := c.Request().Context()

	account, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_WITHDRAWAL)
	if err != nil {
		return factory.ErrorSystem.New(errors.Wrap(err, "Fail Get Account"))
	}
	xAccessToken := gmo_aozora.GetAccessToken()

	logrus.WithFields(logrus.Fields{
		"accountId":            account.AccountId,
		"applyNo":              v.ApplyNo,
		"cancelTargetKeyClass": v.CancelTargetKeyClass,
		"time":                 time.Now().String(),
	}).Info("Get Transfer Result Code ")

	res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "BulkTransferApi.BulkTransferRequestResultUsingGET", func(ctx context.Context) (client.TransferRequestResultResponse, *http.Response, error) {
		return corporateClient.BulkTransferApi.BulkTransferRequestResultUsingGET(ctx, account.AccountId, v.ApplyNo, xAccessToken)
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"accountId":      account.AccountId,
			"applyNo":        v.ApplyNo,
			"httpStatusCode": gmo_aozora.HTTPStatusCode(httpResp),
			"errDetail":      gmo_aozora.GetSwaggerErrorDetailString(err),
			"error":          err,
			"time":           time.Now().String(),
		}).Error("Fail Transfer Result ")

		return err
	}
	return renderSucc(c, http.StatusOK, res)
}

func (BulkTransferApiControllers) CancelBulkTransfer(c echo.Context) error {
	var v BulkTransferCancelInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}

	ctx := c.Request().Context()

	account, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_WITHDRAWAL)
	if err != nil {
		return factory.ErrorSystem.New(errors.Wrap(err, "Fail Get Account"))
	}

	body := client.TransferCancelRequest{AccountId: account.AccountId, CancelTargetKeyClass: v.CancelTargetKeyClass, ApplyNo: v.ApplyNo}

	res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "BulkTransferApi.BulkTransferCancelUsingPOST", func(ctx context.Context) (client.TransferCancelResponse, *http.Response, error) {
		return corporateClient.BulkTransferApi.BulkTransferCancelUsingPOST(ctx, body, gmo_aozora.GetAccessToken())
	})
	if err != nil {
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"error":          err,
			"httpStatusCode": gmo_aozora.HTTPStatusCode(httpResp),
		}).Error("Fail Transfer Cancel Request ")
		return err
	}

	return renderSucc(c, http.StatusOK, res.ResultCode)
}

func (BulkTransferApiControllers) GetStatusBulkTransfer(c echo.Context) error {
	var v BulkTransferStatusInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}

	ctx := c.Request().Context()

	account, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_WITHDRAWAL)
	if err != nil {
		return factory.ErrorSystem.New(errors.Wrap(err, "Fail Get Account"))
	}
	GenerateTransferOption := func() *client.BulkTransferStatusUsingGETOpts {
		opt := &client.BulkTransferStatusUsingGETOpts{}
		if v.ApplyNo != "" {
			opt.ApplyNo = optional.NewString(v.ApplyNo)
		}
		if v.DateFrom != "" {
			opt.DateFrom = optional.NewString(v.DateFrom)
		}
		if v.DateTo != "" {
			opt.DateTo = optional.NewString(v.DateTo)
		}
		if v.DetailInfoNecessity != "" {
			opt.DetailInfoNecessity = optional.NewBool(v.DetailInfoNecessity == "Y")
		}

		return opt
	}

	res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "BulkTransferApi.BulkTransferStatusUsingGET", func(ctx context.Context) (client.BulkTransferStatusResponse, *http.Response, error) {
		return corporateClient.BulkTransferApi.BulkTransferStatusUsingGET(ctx, account.AccountId, v.QueryKeyClass, gmo_aozora.GetAccessToken(), GenerateTransferOption())
	})
	if err != nil {
		errDetail := gmo_aozora.GetSwaggerErrorDetailString(err)

		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"accountId":      account.AccountId,
			"queryKeyClass":  v.QueryKeyClass,
			"applyNo":        v.ApplyNo,
			"httpStatusCode": gmo_aozora.HTTPStatusCode(httpResp),
			"errDetail":      errDetail,
			"error":          err,
			"time":           time.Now().String(),
		}).Error("Fail Transfer Status Using GET ")
		return errors.Wrap(err, "Fail Transfer Status Using GET :"+errDetail)
	}

	return renderSucc(c, http.StatusOK, res)
}

// 単件の手数料見積り
func (BulkTransferApiControllers) GetTransferRequestsFee(c echo.Context) error {
	var v TransferInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}

	ctx := c.Request().Context()
	account, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_WITHDRAWAL)
	if err != nil {
		return factory.ErrorSystem.New(errors.Wrap(err, "Fail Get Account"))
	}
	xAccessToken := gmo_aozora.GetAccessToken()

	amount := strconv.FormatInt(v.RequestAmount, 10)
	transfer := client.BulkTransfer{
		ItemId:                "1",
		BeneficiaryBankCode:   v.BankCode,
		BeneficiaryBankName:   v.BankName,
		BeneficiaryBranchCode: v.BranchCode,
		BeneficiaryBranchName: v.BranchName,
		AccountTypeCode:       v.AccountTypeCode,
		AccountNumber:         v.AccountNumber,
		BeneficiaryName:       gmoaozorautils.ZenkakuToHankaku(v.Depositor),
		TransferAmount:        amount,
	}

	body := client.BulkTransferRequest{
		AccountId:              account.AccountId,
		TransferDesignatedDate: v.RequestDate,
		TotalCount:             "1",
		TotalAmount:            amount,
		BulkTransfers:          []client.BulkTransfer{transfer},
	}

	res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "BulkTransferApi.BulkTransferFeeUsingPOST", func(ctx context.Context) (client.TransferFeeResponse, *http.Response, error) {
		return corporateClient.BulkTransferApi.BulkTransferFeeUsingPOST(ctx, body, xAccessToken)
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"accountId":      account.AccountId,
			"httpStatusCode": gmo_aozora.HTTPStatusCode(httpResp),
			"errDetail":      gmo_aozora.GetSwaggerErrorDetailString(err),
			"error":          err,
			"time":           time.Now().String(),
		}).Error("Fail Transfer Request Process ")

		// MIR-1172: 単件の応答(result)は裸の数値のまま変えない。is_fallbackを構造体として
		// 足すと、rollout中に新旧バージョンの型が混在して出金申請が全件エラーになる
		// (この単件見積もりを呼ぶdomain.Transfer.RequestFeeはisFallbackを使っておらず、
		// 実利が無いまま互換性だけ壊すことになる)。失敗したこと自体は直前のErrorログに残る。
		// GMOあおぞらなら0円、それ以外は99円でフォールバック
		if v.BankCode == gmoAozoraBankCode {
			return renderSucc(c, http.StatusOK, gmoAozoraFallbackTransferFee)
		}
		return renderSucc(c, http.StatusOK, otherBankFallbackTransferFee)
	}

	return renderSucc(c, http.StatusOK, res.TotalFee)
}

// 複数件の手数料見積り
func (BulkTransferApiControllers) GetTransferRequestsFeeBulk(c echo.Context) error {
	var v BulkTransferFeeInputs
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if len(v.Transfers) == 0 {
		return factory.ErrorParameter.New(errors.New("transfers is empty"))
	}

	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}

	ctx := c.Request().Context()
	account, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_WITHDRAWAL)
	if err != nil {
		return factory.ErrorSystem.New(errors.Wrap(err, "Fail Get Account"))
	}
	xAccessToken := gmo_aozora.GetAccessToken()

	totalAmount := int64(0)
	transfers := make([]client.BulkTransfer, 0, len(v.Transfers))
	itemToWithdrawal := make(map[string]string, len(v.Transfers)) // GMO itemId -> withdrawalId
	for i, t := range v.Transfers {
		amtStr := strconv.FormatInt(t.RequestAmount, 10)
		totalAmount += t.RequestAmount
		itemId := strconv.Itoa(i + 1) // GMO要件: 1〜の連番
		if t.WithdrawalId != 0 {
			itemToWithdrawal[itemId] = strconv.FormatInt(t.WithdrawalId, 10)
		} else if t.ItemId != "" {
			itemToWithdrawal[itemId] = t.ItemId
		}
		transfers = append(transfers, client.BulkTransfer{
			ItemId:                itemId,
			BeneficiaryBankCode:   t.BankCode,
			BeneficiaryBankName:   t.BankName,
			BeneficiaryBranchCode: t.BranchCode,
			BeneficiaryBranchName: t.BranchName,
			AccountTypeCode:       t.AccountTypeCode,
			AccountNumber:         t.AccountNumber,
			BeneficiaryName:       gmoaozorautils.ZenkakuToHankaku(t.Depositor),
			TransferAmount:        amtStr,
		})
	}

	body := client.BulkTransferRequest{
		AccountId:              account.AccountId,
		TransferDesignatedDate: v.Transfers[0].RequestDate,
		TotalCount:             strconv.Itoa(len(v.Transfers)),
		TotalAmount:            strconv.FormatInt(totalAmount, 10),
		BulkTransfers:          transfers,
	}

	res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "BulkTransferApi.BulkTransferFeeUsingPOST", func(ctx context.Context) (client.TransferFeeResponse, *http.Response, error) {
		return corporateClient.BulkTransferApi.BulkTransferFeeUsingPOST(ctx, body, xAccessToken)
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"accountId":      account.AccountId,
			"transferCount":  len(v.Transfers),
			"httpStatusCode": gmo_aozora.HTTPStatusCode(httpResp),
			"errDetail":      gmo_aozora.GetSwaggerErrorDetailString(err),
			"error":          err,
			"time":           time.Now().String(),
		}).Error("Fail Transfer Request Fee Process (Batch)")

		// 各口座ごとに判定（0310:0円、その他:99円）
		fallbackFees := make([]int64, len(v.Transfers))
		for i, t := range v.Transfers {
			if t.BankCode == gmoAozoraBankCode {
				fallbackFees[i] = gmoAozoraFallbackTransferFee
			} else {
				fallbackFees[i] = otherBankFallbackTransferFee
			}
		}
		details := make([]client.TransferFeeDetail, 0, len(v.Transfers))
		total := int64(0)
		for i := range v.Transfers {
			itemId := strconv.Itoa(i + 1)
			details = append(details, client.TransferFeeDetail{
				ItemId:      itemId,
				TransferFee: strconv.FormatInt(fallbackFees[i], 10),
			})
			total += fallbackFees[i]
		}
		fallback := client.TransferFeeResponse{
			TotalFee:           strconv.FormatInt(total, 10),
			TransferFeeDetails: details,
		}
		conv, err := convertFeeResponseToDTO(fallback, itemToWithdrawal)
		if err != nil {
			return factory.ErrorSystem.New(err)
		}
		// MIR-1172: 固定値へフォールバックしたことを応答から分かるようにする。
		conv.IsFallback = true
		return renderSucc(c, http.StatusOK, conv)
	}

	conv, err := convertFeeResponseToDTO(res, itemToWithdrawal)
	if err != nil {
		return factory.ErrorSystem.New(err)
	}

	return renderSucc(c, http.StatusOK, conv)
}

func (BulkTransferApiControllers) GetTransferRequests(c echo.Context) error {
	var v TransferInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}

	ctx := c.Request().Context()
	account, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_WITHDRAWAL)
	if err != nil {
		return factory.ErrorSystem.New(errors.Wrap(err, "Fail Get Account"))
	}
	xAccessToken := gmo_aozora.GetAccessToken()

	amount := strconv.FormatInt(v.RequestAmount, 10)
	transfer := client.BulkTransfer{
		ItemId:                "1",
		BeneficiaryBankCode:   v.BankCode,
		BeneficiaryBankName:   v.BankName,
		BeneficiaryBranchCode: v.BranchCode,
		BeneficiaryBranchName: v.BranchName,
		AccountTypeCode:       v.AccountTypeCode,
		AccountNumber:         v.AccountNumber,
		BeneficiaryName:       gmoaozorautils.ZenkakuToHankaku(v.Depositor),
		TransferAmount:        amount,
	}
	transfers := []client.BulkTransfer{transfer}
	body := client.BulkTransferRequest{
		AccountId:              account.AccountId,
		TransferDesignatedDate: v.RequestDate,
		TotalCount:             "1",
		TotalAmount:            amount,
		BulkTransfers:          transfers,
	}
	res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "BulkTransferApi.BulkTransferRequestUsingPOST", func(ctx context.Context) (client.BulkTransferRequestResponse, *http.Response, error) {
		return corporateClient.BulkTransferApi.BulkTransferRequestUsingPOST(ctx, body, xAccessToken)
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"accountId":      account.AccountId,
			"httpStatusCode": gmo_aozora.HTTPStatusCode(httpResp),
			"errDetail":      gmo_aozora.GetSwaggerErrorDetailString(err),
			"error":          err,
			"time":           time.Now().String(),
		}).Error("Fail Transfer Request Process ")
		return factory.ErrorParameter.New(err)
	}

	return renderSucc(c, http.StatusOK, res)
}
