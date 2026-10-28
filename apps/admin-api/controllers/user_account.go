package controllers

import (
	"net/http"
	"strconv"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/service"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type UserAccountApiController struct {
	withdrawalService service.IWithdrawalService
}

func NewUserAccountApiController() *UserAccountApiController {
	return &UserAccountApiController{
		withdrawalService: service.NewWithdrawalService(),
	}
}

func (c UserAccountApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll).
		AddParamQueryNested(UserAccountSearchInput{}).
		SetDescription("ユーザーの預金口座リストを照会します。").
		SetSummary("ユーザーの預金口座リスト照会").
		SetSecurity("Authorization")

	g.GET("/download", c.Download).
		AddParamQueryNested(UserAccountSearchInput{}).
		SetDescription("ユーザーの預金口座リストをダウンロードします。").
		SetSummary("ユーザーの預金口座リストダウンロード").
		SetSecurity("Authorization")

	g.GET("/:id", c.GetOne).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "ユーザーの口座照会", UserAccountDetailResponse{}, nil).
		SetDescription("ユーザーの預金口座情報を照会します。").
		SetSummary("ユーザーの預金口座詳細情報照会").
		SetSecurity("Authorization")

	g.GET("/:id/transactions", c.GetAccountTransactions).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "預金取引履歴照会", domain.UserAccount{}, nil).
		SetDescription("預金取引履歴を照会します。").
		SetSummary("預金取引履歴照会").
		SetSecurity("Authorization")
}

func (UserAccountApiController) GetAll(c echo.Context) error {

	tokenId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	var v UserAccountSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	totalCount, items, err := domain.UserAccount{}.GetAll(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.UserId, v.UserName, v.StartBalance, v.EndBalance)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), tokenId, email,
		"menu.manage_accounting.deposit_list", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (u UserAccountApiController) GetOne(c echo.Context) error {

	tokenId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	ctx := c.Request().Context()

	item, err := domain.UserAccount{}.GetById(ctx, id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 投資金額情報（運用中・投資完了・投資申請中）
	amountInfo, err := domain.UserShareAmountInfo{}.GetUserProjectShareAmount(ctx, item.UserId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	requestInvestAmount, err := domain.UserShareAmountInfo{}.SumDepositAmountOfRequestInvest(ctx, item.UserId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	amountInfo.InvestRequestAmount = requestInvestAmount
	// NOTE: 「投資完了金額」は「投資申請中金額」の中に含まれている
	amountInfo.InvestTotalAmount = amountInfo.InvestOperationAmount + amountInfo.InvestRequestAmount
	amountInfo.InvestTotalCount = amountInfo.InvestOperationCount + amountInfo.InvestSuccessCount + amountInfo.InvestRequestCount

	// 出金申請中金額（未完了 + 締切 + 送金失敗）
	withdrawalTotal, err := u.withdrawalService.GetTotalWithdrawalAmountByUserId(ctx, item.UserId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 投資入金中で未入金の金額
	noDepositRequests, err := domain.DepositRequest{}.GetNoDepositRequestByUserId(ctx, item.UserId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	var totalUnpaidAmount int64
	for _, ndr := range noDepositRequests {
		totalUnpaidAmount += ndr.RequestAmount - ndr.DepositAmount
	}

	// 顧客資産総額 = 運用中資産 + 投資申請中金額 + 預け金 + 出金申請中金額
	userTotalAsset := amountInfo.InvestTotalAmount + lo.FromPtr(item.Balance) + withdrawalTotal

	resp := UserAccountDetailResponse{
		UserAccount:         *item,
		Withdrawal:          withdrawalTotal,
		UserShareAmountInfo: *amountInfo,
		UserTotalAsset:      userTotalAsset,
		UserTotalUnpaid:     totalUnpaidAmount,
	}

	err = domain.OperatorHistory{}.Record(ctx, tokenId, email,
		"menu.manage_accounting.deposit_detail", "ID :"+strconv.FormatInt(item.UserId, 10), c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, resp)
}

func (UserAccountApiController) GetAccountTransactions(c echo.Context) error {
	var v SearchInput
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	account, err := domain.UserAccount{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 예치금 거래내역은 사업형에 한해서만
	totalCount, items, err := domain.UserTransaction{}.GetByUserIdAndType(c.Request().Context(), account.UserId, domain.EQUITY, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (UserAccountApiController) Download(c echo.Context) error {

	var v UserAccountSearchInput
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

	err = domain.UserAccount{}.DownloadCsv(c.Request().Context(), c.Response().Writer, v.Sortby, v.Order, v.UserId, v.UserName, v.StartBalance, v.EndBalance)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.deposit_download", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}
