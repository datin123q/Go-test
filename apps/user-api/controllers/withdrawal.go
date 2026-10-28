package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/samber/lo"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

type WithdrawalApiController struct {
	withdrawalService service.IWithdrawalService
}

func NewWithdrawalController() *WithdrawalApiController {
	return &WithdrawalApiController{
		withdrawalService: service.NewWithdrawalService(),
	}
}

func (api WithdrawalApiController) Init(g echoswagger.ApiGroup) {
	g.POST("", api.Create).
		AddParamForm(WithdrawalInput{}, "body", "", false).
		AddResponse(http.StatusCreated, "出金申請登録", WithdrawalDto{}, nil).
		SetDescription("出金を申請").
		SetSummary("出金申請").
		SetSecurity("Authorization")

	g.GET("", api.GetAll).
		AddResponse(http.StatusOK, "出金申請一覧取得", []WithdrawalDto{}, nil).
		SetDescription("出金申請の一覧を取得").
		SetSummary("出金申請一覧取得").
		SetSecurity("Authorization")

	g.GET("/:id", api.GetOne).
		AddParamPath("", "id", "出金申請ID").
		AddResponse(http.StatusOK, "出金申請詳細情報", WithdrawalDetailDto{}, nil).
		SetDescription("出金申請の詳細を取得").
		SetSummary("出金申請詳細取得").
		SetSecurity("Authorization")

	g.PUT("/:id", api.UpdateStatus).
		AddParamPath("", "id", "").
		AddParamQuery("", "status", "CANCEL", true).
		AddResponse(http.StatusOK, "出金申請ステータス変更", WithdrawalDto{}, nil).
		SetDescription("出金申請のステータスを変更（未完了 → キャンセル）。").
		SetSummary("出金申請ステータス変更").
		SetSecurity("Authorization")
}

// Create は投資家ユーザーの出金申請を受け付けるハンドラ。
//
// 以下の順序で処理を行う:
//  1. リクエストボディをバインド・バリデーション
//  2. JWTトークンから投資家IDを取得
//  3. ユーザーの存在確認および本人確認審査完了チェック（COMPLETED_AUDIT かつ 非ADMIN）
//  4. UserAccount の残高チェック（要求額 > 0 かつ 残高 >= 要求額）
//  5. 外部APIで送金手数料を取得（Transfer.RequestFee）
//  6. 残高から出金額を引く（UserAccount.WithdrawBalance）
//  7. Withdrawal レコードを作成（ステータス: INCOMPLETE）
//  8. 出金申請受付メールを送信
//
// エラー時は以下を返す:
//   - 400: バインド・バリデーションエラー、無効なステータス
//   - 403: 審査未完了ユーザー、管理者ユーザー
//   - 404: ユーザーが存在しない
//   - 残高不足: ErrorInsufficientBalance
//   - 手数料取得失敗: ErrorWithdrawalRequest
//   - DB エラー: ErrorDB
//   - メール送信失敗: ErrorAWSSendMail
func (WithdrawalApiController) Create(c echo.Context) error {
	var v WithdrawalInput

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	investorId, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	user, err := domain.User{}.GetById(c.Request().Context(), investorId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if user == nil {
		return factory.ErrorNotFound.New(errors.New("user not found"))
	}

	if user.AuditStatus != domain.COMPLETED_AUDIT || user.UserType == domain.ADMIN {
		return factory.ErrorNotHaveRole.New(errors.New("user is invalid"))
	}

	account, err := domain.UserAccount{}.GetByUserId(c.Request().Context(), investorId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if lo.FromPtr(account.Balance) < v.Amount || v.Amount <= 0 {
		return factory.ErrorInsufficientBalance.New(err)
	}

	fee, err := domain.Transfer{}.RequestFee(c.Request().Context(), *user, v.Amount)
	if err != nil {
		return factory.ErrorWithdrawalRequest.New(err)
	}

	withdrawal, err := v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	withdrawal.Fee = fee
	withdrawal.RequestDate = jsonutils.JSONTime(time.Now())
	withdrawal.Status = domain.WITHDRAWAL_STATUS_INCOMPLETE
	withdrawal.IsManual = false
	withdrawal.InvestorName = user.Name
	withdrawal.InvestorId = strconv.FormatInt(investorId, 10)

	_, err = account.WithdrawBalance(c.Request().Context(), withdrawal.Amount, domain.WITHDRAWAL_INVESTOR)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if _, err := withdrawal.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	var mail domain.Mail

	err = mail.WithdrawalRequestMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), user.Email,
		withdrawal.InvestorName, withdrawal.RequestDate.Time().Format(time.DateOnly), strconv.FormatInt(withdrawal.Amount, 10))
	if err != nil {
		return factory.ErrorAWSSendMail.New(err)
	}

	return renderSucc(c, http.StatusCreated, WithdrawalDto{}.New(withdrawal))
}

// GetAll はログイン中の投資家ユーザー自身の出金申請一覧を返すハンドラ。
//
// 以下の順序で処理を行う:
//  1. クエリパラメータ（ソート・ページング）をバインド
//  2. JWTトークンからユーザーIDを取得
//  3. ユーザーIDに紐づく出金申請一覧を取得
//
// レスポンスは items（出金申請リスト）と total_count（総件数）を返す。
func (api WithdrawalApiController) GetAll(c echo.Context) error {
	// parse query string
	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// get info from token
	uid, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	// get withdrawal list
	total, withdrawalList, err := api.withdrawalService.GetWithdrawalDetailListByUserId(c.Request().Context(), uid, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return err
	}

	type ArrayResult struct {
		Items      any   `json:"items"`
		TotalCount int64 `json:"total_count"`
	}

	return apirender.RenderSucc(c, http.StatusOK, ArrayResult{
		Items:      WithdrawalDto{}.NewListByDetail(withdrawalList),
		TotalCount: total,
	})
}

// GetOne は指定IDの出金申請詳細を返すハンドラ。
//
// 以下の順序で処理を行う:
//  1. JWTトークンからユーザーIDを取得
//  2. パスパラメータから出金申請IDを取得
//  3. 出金申請詳細を取得し、リクエスト元ユーザーが申請者本人か確認
//  4. ステータスが COMPLETE かつ送金履歴がある場合、最新 Transfer の情報（手数料・口座情報）でレスポンスを上書き
//  5. 現在の口座残高（ExpectedBalance）を付与
//  6. ステータスが CANCEL の場合、キャンセル理由を付与（Transfer 有無でADMIN/USERキャンセルを判別）
func (api WithdrawalApiController) GetOne(c echo.Context) error {
	// get info from token
	uid, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	// parse path variable
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	// get withdrawal detail
	withdrawalDetail, err := api.withdrawalService.GetWithdrawalDetailById(c.Request().Context(), id)
	if err != nil {
		return err
	}

	// check permission
	userId, err := strconv.ParseInt(withdrawalDetail.InvestorId, 10, 64)
	if err = jwthelper.ValidateUserToken(c, userId); err != nil {
		return err
	}

	withdrawalDetailDto := WithdrawalDetailDto{}.New(withdrawalDetail)

	// 送金履歴がある場合は最新transferでスナップショット上書き
	_, transferList, err := domain.Transfer{}.GetByWithdrawalId(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err, "Failed to get transfer list by withdrawal ID.")
	}
	// ステータスがCOMPLETEかつtransferがある場合は最新transferでスナップショット上書き
	if withdrawalDetailDto.Status == domain.WITHDRAWAL_STATUS_COMPLETE && len(transferList) > 0 {
		latest := transferList[0]
		withdrawalDetailDto.Fee = latest.TransferFee
		withdrawalDetailDto.BankName = latest.BankName
		withdrawalDetailDto.BankAccount = latest.AccountNumber
		withdrawalDetailDto.BankCode = &latest.BankCode
		withdrawalDetailDto.BranchName = latest.BranchName
		withdrawalDetailDto.BranchCode = &latest.BranchCode
		switch latest.AccountTypeCode {
		case "1":
			withdrawalDetailDto.DepositType = domain.ORDINARY_ACCOUNT
		case "2":
			withdrawalDetailDto.DepositType = domain.CHECKING_ACCOUNT
		}
	}

	// get expected Balance
	account, err := domain.UserAccount{}.GetByUserId(c.Request().Context(), uid)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	withdrawalDetailDto.ExpectedBalance = lo.FromPtr(account.Balance)

	// get reason for cancellation
	if withdrawalDetail.Status == domain.WITHDRAWAL_STATUS_CANCEL {
		total, _, err := domain.Transfer{}.GetByWithdrawalId(c.Request().Context(), id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		if total > 0 {
			withdrawalDetailDto.CancelledReason = api.withdrawalService.GetWithdrawalCancelledReason(*withdrawalDetail, "ADMIN")
		} else {
			withdrawalDetailDto.CancelledReason = api.withdrawalService.GetWithdrawalCancelledReason(*withdrawalDetail, "USER")
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, withdrawalDetailDto)
}

// UpdateStatus は出金申請のステータスをキャンセルに変更するハンドラ。
//
// ユーザーが行える操作はキャンセル（CANCEL）のみ。
// 以下の順序で処理を行う:
//  1. JWTトークンからユーザーIDを取得
//  2. パスパラメータから出金申請IDを取得
//  3. 出金申請を取得し、申請者本人かチェック（他人の申請はキャンセル不可）
//  4. クエリパラメータ status が CANCEL 以外の場合はエラー
//  5. ステータスを CANCEL に更新
//  6. キャンセル後処理:
//     - 出金額を口座残高に返金（DepositBalance）
//     - 未入金状態（NO_DEPOSIT）の入金必要リクエストを再チェック（DepositCheckProcess）
func (api WithdrawalApiController) UpdateStatus(c echo.Context) error {
	// get info from token
	uid, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	// parse path variable
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	// get withdrawal detail
	withdrawalDetail, err := api.withdrawalService.GetWithdrawalDetailById(c.Request().Context(), id)
	if err != nil {
		return err
	}

	// check permission
	investorId, err := strconv.ParseInt(withdrawalDetail.InvestorId, 10, 64)
	if investorId != uid {
		return factory.ErrorPermissionDenied.New(errors.New("Only the owner can cancel the withdrawal request"))
	}

	if err := jwthelper.ValidateUserToken(c, investorId); err != nil {
		return err
	}

	// get status from query param
	status := domain.WITHDRAWAL_STATUS_TYPE(c.QueryParam("status"))
	if status != domain.WITHDRAWAL_STATUS_CANCEL { // user only can CANCEL withdrawal
		return factory.ErrorParameter.New(errors.Wrap(err, "invalid status"))
	}

	// convert status
	if _, err := api.withdrawalService.ConvertWithdrawalStatus(c.Request().Context(), &withdrawalDetail.Withdrawal, status); err != nil {
		return err
	}

	// extra process by status
	if status == domain.WITHDRAWAL_STATUS_CANCEL { // cancel withdrawal
		userAccount, err := domain.UserAccount{}.GetByUserId(c.Request().Context(), withdrawalDetail.User.Id)
		if err != nil {
			return err
		}
		if _, err := userAccount.DepositBalance(c.Request().Context(), withdrawalDetail.Amount); err != nil {
			// MIR-1247: 戻り値を捨てていたため、返金のUPDATEが競合等で
			// 0行更新になっても出金キャンセル自体は成立してしまい、返金
			// されないまま処理が進んでいた。PUTハンドラのリクエストtx内
			// なので、ここでエラーを返せばConvertWithdrawalStatusによる
			// キャンセル状態への変更もロールバックされる。
			return factory.ErrorDB.New(errors.Wrap(err, "failed to deposit balance for withdrawal cancel"))
		}

		depositReqList, err := domain.DepositRequest{}.GetNoDepositRequestByUserId(c.Request().Context(), withdrawalDetail.User.Id)
		if err != nil {
			return factory.ErrorDB.New(errors.Wrap(err, "failed to get deposit request list (status: NO_DEPOSIT)"))
		}
		for _, depositReq := range depositReqList {
			account, err := domain.UserAccount{}.GetByUserId(c.Request().Context(), depositReq.UserId)
			if err != nil {
				return factory.ErrorDB.New(errors.Wrap(err, "failed to get user account"))
			}
			err = account.DepositCheckProcess(c.Request().Context(), &depositReq, domain.NOT_IMMEDIATE)
			if err != nil {
				return factory.ErrorSystem.New(errors.Wrap(err, "failed to deposit check process"))
			}
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, WithdrawalDto{}.NewByDetail(withdrawalDetail))
}
