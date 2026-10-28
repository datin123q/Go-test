package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/rs/xid"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/service"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type WithdrawalApiController struct {
	withdrawalService service.IWithdrawalService
}

func NewWithdrawalController() *WithdrawalApiController {
	return &WithdrawalApiController{
		withdrawalService: service.NewWithdrawalService(),
	}
}

// withdrawalTransferValidationErrorはMIR-1256の対応。送金データ(domain.Transfer)を
// 組み立てる前に、投資家IDが数値であること・銀行コード/支店コードが登録されている
// ことを確認する。問題が無ければnilを返す。以前はinvestorIdが数値でない場合に
// logrus.Fatal(os.Exit(1)、admin-api全体が落ちる)、bank_code/branch_codeが
// nil(DBでNULLを許容、Go側は*string)の場合にnilポインタ参照のパニック(500)が
// それぞれ起きており、1人分のデータ不整合が一括送金全体・admin-api全体を巻き込んで
// いた。呼び出し側はこのエラーを見て、この1件を除いて残りを続けるか、
// リクエストごと拒否するかを選ぶ。
func withdrawalTransferValidationError(investorId string, bankCode, branchCode *string, bankAccount string, depositType domain.DEPOSIT_TYPE) error {
	if _, err := strconv.ParseInt(investorId, 10, 64); err != nil {
		return errors.Wrap(err, "investor id is not numeric")
	}
	// 銀行口座情報4項目の判定はdomain.WithdrawalBankInfoErrorへ集約している
	// (domain/transfer.goのRequestFeeとも共有するため)。
	return domain.WithdrawalBankInfoError(bankCode, branchCode, bankAccount, depositType)
}

// buildWithdrawalTransfersForCloseはMIR-1256の対応。締めに含まれる出金依頼を
// 順に見て、送金データが組み立てられるものだけをtransfers/mappingsに積み、
// 組み立てられないものはidを集めて返す。
//
// 単に対象から外してcontinueするだけでは、この締めがSENDING→COMPLETEへ進んだ後、
// 外した出金依頼はWAITのまま取り残される。WAITから個別再送金(FAILのみ受付)にも
// キャンセル(FAILまたはINCOMPLETEのみ受付、checkStatusValidation)にも進めず、
// 出金申請時に引き落とし済みの会員残高を戻す経路も無くなる(DB直接修正以外に
// 復旧手段が無くなる)。これを避けるため、外す出金依頼はその場でFAILへ更新し、
// 個別再送金画面からデータを直して再送金するか、キャンセルして返金する、という
// 既存の経路に乗せる。
//
// このFAIL更新は呼び出し元(RequestTransferByClosedId)のリクエスト単位の暗黙
// トランザクション内で行われる(session.Begin/Commit/Rollbackはapps/admin-api/
// server/server.goのミドルウェア)。呼び出し元が最終的にエラーを返せば、ここで
// 行ったFAIL更新も含め全体がロールバックされる。全件がスキップされ
// transfers が空になるケースが呼び出し元のこの経路にあたり、その場合は
// 締め・出金依頼ともWAITのまま変わらない(個別再送金・キャンセルの経路は開かない。
// 復旧はデータを直して一括送金を再実行するのみ)。
func buildWithdrawalTransfersForClose(ctx context.Context, items []domain.WithdrawalDetail) (transfers []domain.Transfer, mappings []domain.WithdrawalTransferMapping, skippedWithdrawalIds []int64, err error) {
	transfers = make([]domain.Transfer, 0)

	for _, item := range items {
		if validationErr := withdrawalTransferValidationError(item.Withdrawal.InvestorId, item.User.BankCode, item.User.BranchCode, item.User.BankAccount, item.User.DepositType); validationErr != nil {
			logrus.WithField("withdrawal_id", item.Withdrawal.Id).WithError(validationErr).
				Error("Withdrawal data is invalid for transfer; marking as FAIL and skipping this withdrawal")

			failed := domain.Withdrawal{Id: item.Withdrawal.Id, Status: domain.WITHDRAWAL_STATUS_FAIL}
			if updateErr := failed.Update(ctx); updateErr != nil {
				return nil, nil, nil, updateErr
			}

			skippedWithdrawalIds = append(skippedWithdrawalIds, item.Withdrawal.Id)
			continue
		}
		// 直前のwithdrawalTransferValidationErrorで数値であることを確認済み
		userId, _ := strconv.ParseInt(item.Withdrawal.InvestorId, 10, 64)

		t := domain.Transfer{
			TransactionId:         xid.New().String(),
			UserId:                userId,
			TransferDate:          jsonutils.JSONTime(time.Now()),
			TransferAmount:        item.Amount,
			AccountNumber:         item.User.BankAccount,
			BankCode:              *item.User.BankCode,
			BankName:              item.User.BankName,
			BranchCode:            *item.User.BranchCode,
			BranchName:            item.User.BranchName,
			Depositor:             item.User.Depositor,
			AccountTypeCode:       domain.DepositAccountTypeCheck(item.User.DepositType),
			TransactionDetailType: domain.WITHDRAWAL_INVESTOR,
			TransferStatus:        domain.TRANSFER_STATUS_WAIT,
		}

		transfers = append(transfers, t)

		mappings = append(mappings, domain.WithdrawalTransferMapping{
			WithdrawalId:  item.Withdrawal.Id,
			TransactionId: t.TransactionId,
		})
	}

	return transfers, mappings, skippedWithdrawalIds, nil
}

func (c WithdrawalApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll).
		AddParamQueryNested(WithdrawalSearchInput{}).
		AddResponse(http.StatusOK, "출금 요청 목록 조회", WithdrawalDetailDto{}, nil).
		SetDescription("출금 요청 목록을 조회 합니다.").
		SetSummary("출금 요청 목록 조회").
		SetSecurity("Authorization")

	g.GET("/:id", c.GetOne).
		AddParamPath("", "id", "출금의뢰 ID").
		AddResponse(http.StatusOK, "특정 출금 요청의 상세 정보", WithdrawalDetailDto{}, nil).
		SetDescription("출금 요청을 조회 합니다.").
		SetSummary("출금 요청 상세 조회").
		SetSecurity("Authorization")

	g.PUT("/:id", c.UpdateStatus).
		AddParamPath("", "id", "").
		AddParamQuery("", "status", "COMPLETE | INCOMPLETE | CANCEL", true).
		AddResponse(http.StatusOK, "출금 의뢰 상태 변경", WithdrawalDetailDto{}, nil).
		SetDescription("출금 의뢰의 상태 변경 (대기,실패 -> 완료 || 실패 -> 미완료,취소)").
		SetSummary("출금 의뢰 상태 변경").
		SetSecurity("Authorization")

	g.POST("/:id/transfer", c.RequestTransferByWithdrawalId).
		AddParamPath("", "id", "출금의뢰 ID").
		AddResponse(http.StatusOK, "출금의뢰 송금 요청", domain.WithdrawalTransferMapping{}, nil).
		SetDescription("마감된 특정 출금의뢰를 송금요청 합니다.").
		SetSummary("마감된 특정 출금의뢰 송금 요청.").
		SetSecurity("Authorization")

	g.GET("/closed/:id/download", c.DownloadClosed).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "마감 목록 다운로드", nil, nil).
		SetDescription("마감된 목록을 다운로드 합니다.").
		SetSummary("마감된 목록 다운로드").
		SetSecurity("Authorization")

	g.GET("/closed", c.GetClosed).
		AddParamQueryNested(SearchInput{}).
		AddResponse(http.StatusOK, "마감된 출금 요청 목록 조회", WithdrawalCloseDetailDto{}, nil).
		SetDescription("마감된 출금 요청 목록을 조회 합니다.").
		SetSummary("마감된 출금 요청 목록 조회").
		SetSecurity("Authorization")

	g.PUT("/closed", c.UpdateClosed).
		AddResponse(http.StatusOK, "출금 요청 목록 일괄 마감 처리", nil, nil).
		SetDescription("'미완료' 상태의 출금 요청 목록을 일괄 '송금 대기' 상태로 변경 합니다.").
		SetSummary("출금 요청 목록 일괄 마감 처리").
		SetSecurity("Authorization")

	g.GET("/closed/:id", c.GetResult).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "출금 요청 결과 조회", WithdrawalDetailDto{}, nil).
		SetDescription("출금 요청 결과를 조회 합니다.").
		SetSummary("출금 요청 결과 조회").
		SetSecurity("Authorization")

	g.POST("/closed/:id/transfer", c.RequestTransferByClosedId).
		AddParamPath("", "id", "출금의뢰 마감 ID").
		AddResponse(http.StatusOK, "마감된 출금 요청 송금 승인", nil, nil).
		SetDescription("마감된 출금의뢰를 송금 승인 합니다.").
		SetSummary("마감된 출금 요청 송금 승인").
		SetSecurity("Authorization")
}

func (WithdrawalApiController) GetAll(c echo.Context) error {
	var v WithdrawalSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"InvestorId":        v.InvestorId,
		"InvestorName":      v.Orderer,
		"AmountStart":       v.AmountStart,
		"AmountEnd":         v.AmountEnd,
		"RequestStartDate":  v.RequestStartDate,
		"RequestEndDate":    v.RequestEndDate,
		"CompleteStartDate": v.CompleteStartDate,
		"CompleteEndDate":   v.CompleteEndDate,
		"Status":            v.Status,
	}).Info("SearchStart")

	totalCount, items, err := domain.Withdrawal{}.GetAll(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.InvestorId, v.Orderer, v.AmountStart, v.AmountEnd, v.RequestStartDate, v.RequestEndDate, v.CompleteStartDate, v.CompleteEndDate, v.Status)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// ステータスがCOMPLETEのものだけまとめて最新transferを取得してスナップショット上書き
	completeIds := make([]int64, 0)
	for i := range items {
		if items[i].Status == domain.WITHDRAWAL_STATUS_COMPLETE {
			completeIds = append(completeIds, items[i].Withdrawal.Id)
		}
	}
	if len(completeIds) > 0 {
		latestMap, err := domain.Transfer{}.GetLatestByWithdrawalIds(c.Request().Context(), completeIds)
		if err != nil {
			logrus.WithError(err).Warn("failed to fetch latest transfers for snapshot")
		} else {
			for i := range items {
				if items[i].Status != domain.WITHDRAWAL_STATUS_COMPLETE {
					continue
				}
				if latest, ok := latestMap[items[i].Withdrawal.Id]; ok {
					items[i].Fee = latest.TransferFee
					items[i].BankName = latest.BankName
					items[i].BankAccount = latest.AccountNumber
					items[i].BankCode = &latest.BankCode
					items[i].BranchName = latest.BranchName
					items[i].BranchCode = &latest.BranchCode
				}
			}
		}
	}

	// 未完了件数を照会 -> '締切'ボタン有効化判定
	incompleteCnt, err := domain.Withdrawal{}.CountIncomplete(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	type ArrayResult struct {
		Items           interface{} `json:"items"`
		TotalCount      int64       `json:"total_count"`
		TotalIncomplete int64       `json:"totalIncomplete"`
	}

	dtos := WithdrawalDetailDto{}.NewList(items)
	applyWithdrawalFailureReasons(c.Request().Context(), items, dtos)

	return apirender.RenderSucc(c, http.StatusOK, ArrayResult{
		Items:           dtos,
		TotalCount:      totalCount,
		TotalIncomplete: incompleteCnt,
	})
}

// applyWithdrawalFailureReasons は、一覧と送金管理の確認モーダルのDTOに「失敗の理由」を足す
// (MIR-1170)。dtosはitemsと同じ並び。取得に失敗しても一覧は落とさず、理由を空のまま返す。
// ログには取得した行を載せない(口座情報や受取人名を含みうるため、エラーだけ)。
func applyWithdrawalFailureReasons(ctx context.Context, items []domain.WithdrawalDetail, dtos []*WithdrawalDetailDto) {
	withdrawals := make([]domain.Withdrawal, len(items))
	for i := range items {
		withdrawals[i] = items[i].Withdrawal
	}
	reasons, err := domain.WithdrawalFailureReasons(ctx, withdrawals, false)
	if err != nil {
		logrus.WithError(err).Warn("failed to fetch withdrawal failure reasons")
		return
	}
	for i, dto := range dtos {
		if reason, ok := reasons[items[i].Withdrawal.Id]; ok {
			dto.FailureKind = string(reason.Kind)
			dto.FailureMessage = reason.Message
		}
	}
}

func (WithdrawalApiController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	withdrawal, err := domain.WithdrawalDetail{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if withdrawal == nil {
		return factory.ErrorNotFound.New(err)
	}

	res := WithdrawalDetailDto{}.New(withdrawal)

	// 송금 요청 기록 조회
	_, transferList, err := domain.Transfer{}.GetByWithdrawalId(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err, "Failed to get transfer list by withdrawal ID.")
	}

	// ステータスがCOMPLETEかつtransferがある場合は最新transferでスナップショット上書き
	if res.Status == domain.WITHDRAWAL_STATUS_COMPLETE && len(transferList) > 0 {
		latest := transferList[0]
		res.Fee = latest.TransferFee
		res.BankName = latest.BankName
		res.BankAccount = latest.AccountNumber
		res.BankCode = &latest.BankCode
		res.BranchName = latest.BranchName
		res.BranchCode = &latest.BranchCode
		switch latest.AccountTypeCode {
		case "1":
			res.DepositType = domain.ORDINARY_ACCOUNT
		case "2":
			res.DepositType = domain.CHECKING_ACCOUNT
		}
	}

	transferDto := &TransferDto{}
	res.Transfer = transferDto.NewList(transferList)

	return apirender.RenderSucc(c, http.StatusOK, res)
}

func (a WithdrawalApiController) UpdateStatus(c echo.Context) error {
	uid, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	// get withdrawal detail
	withdrawalDetail, err := a.withdrawalService.GetWithdrawalDetailById(c.Request().Context(), id)
	if err != nil {
		return err
	}

	// get status from query param & set data
	status := domain.WITHDRAWAL_STATUS_TYPE(c.QueryParam("status"))
	if _, err := a.withdrawalService.ConvertWithdrawalStatus(c.Request().Context(), &withdrawalDetail.Withdrawal, status); err != nil {
		return err
	}

	userAccount, err := domain.UserAccount{}.GetByUserId(c.Request().Context(), withdrawalDetail.User.Id)
	if err != nil {
		return factory.ErrorDB.New(errors.Wrap(err, "fail get user account"))
	}

	if status == domain.WITHDRAWAL_STATUS_CANCEL { // cancel withdrawal
		// 예치금 환불
		_, err = userAccount.DepositBalance(c.Request().Context(), withdrawalDetail.Amount)
		if err != nil {
			return factory.ErrorDB.New(errors.Wrap(err, "fail deposit balance"))
		}

		// 출금의뢰 취소 메일 발송
		var mail domain.Mail
		err = mail.WithdrawalCancelMail(c.Request().Context(), withdrawalDetail.User.Email, withdrawalDetail.InvestorName)
		if err != nil {
			return factory.ErrorAWSSendMail.New(err)
		}

		// 입금 필요 금액 처리
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
	} else if status == domain.WITHDRAWAL_STATUS_COMPLETE { // complete withdrawal
		now := time.Now()
		if err := (domain.UserTransaction{}).WithdrawalUserTransaction(c.Request().Context(), withdrawalDetail.User.Id, withdrawalDetail.Amount, 0, domain.EQUITY, withdrawalDetail.User.BankAccount, domain.WITHDRAWAL_INVESTOR, jsonutils.JSONTime(now)); err != nil {
			logrus.WithError(err).Error("failed to create user transaction for withdrawal completion")
		}
		err = domain.SendCustomerLedgerWithdrawalEvent(c.Request().Context(), withdrawalDetail.User.Id, withdrawalDetail.InvestorName, withdrawalDetail.Amount, now)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": errors.Wrap(err, "fail to send customer ledger event"),
			})
		}
	}

	// create log
	err = domain.OperatorHistory{}.Record(c.Request().Context(), uid, email,
		"menu.manage_accounting.manage_withdrawal_close", string(status), c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, WithdrawalDetailDto{}.New(withdrawalDetail))
}

func (WithdrawalApiController) DownloadClosed(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	err = domain.Withdrawal{}.DownloadCsvClosed(c.Request().Context(), c.Response().Writer, id)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	// create log
	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.manage_withdrawal_download", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}

func (WithdrawalApiController) GetClosed(c echo.Context) error {
	var si SearchInput
	if err := c.Bind(&si); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if si.MaxResultCount == 0 {
		si.MaxResultCount = DefaultMaxResultCount
	}

	detail := &domain.WithdrawalCloseDetail{}
	totalCount, items, err := detail.GetAll(c.Request().Context(), si.Sortby, si.Order, si.SkipCount, si.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      WithdrawalCloseDetailDto{}.NewList(items),
		TotalCount: totalCount,
	})
}

func (WithdrawalApiController) UpdateClosed(c echo.Context) error {
	w := &domain.Withdrawal{}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	// INCOMPLETE 対象の手数料を最新口座で再計算（複数件まとめて手数料見積もり）
	withdrawals, _, err := domain.Withdrawal{}.GetIncompleteWithdrawal(c.Request().Context(), 0, 99)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	transferInputs := make([]domain.TransferInput, 0, len(withdrawals))
	wdIndexes := make([]int, 0, len(withdrawals))

	for i, wd := range withdrawals {
		investorId, err := strconv.ParseInt(wd.InvestorId, 10, 64)
		if err != nil {
			logrus.Warnf("failed to parse investor id %s: %v", wd.InvestorId, err)
			continue
		}

		user, err := domain.User{}.GetById(c.Request().Context(), investorId)
		if err != nil {
			logrus.WithError(err).Warnf("failed to get user %d for withdrawal %d", investorId, wd.Id)
			continue
		}

		// MIR-1256: bank_code/branch_codeはDBでNULLを許容している(Go側は*string)。
		// nilチェック無しで参照するとパニックし、締め処理(このハンドラ)全体が
		// 失敗していた。既に上のinvestorId・ユーザー取得の失敗と同じくcontinueで
		// 1件を除いて処理を続ける(このループは手数料の再計算のみが目的で、
		// 対象をWAITへ進めるかどうかとは無関係なため)。
		if err := domain.WithdrawalBankInfoError(user.BankCode, user.BranchCode, user.BankAccount, user.DepositType); err != nil {
			logrus.WithField("withdrawal_id", wd.Id).WithError(err).Warn("Withdrawal user bank info is not fully registered; skipping fee recalculation for this withdrawal")
			continue
		}

		transferInputs = append(transferInputs, domain.TransferInput{
			WithdrawalId:    wd.Id,
			UserId:          user.Id,
			RequestDate:     domain.TransferDesignatedDate(time.Now()),
			RequestAmount:   wd.Amount,
			AccountNumber:   user.BankAccount,
			BankCode:        *user.BankCode,
			BankName:        user.BankName,
			BranchCode:      *user.BranchCode,
			BranchName:      user.BranchName,
			Depositor:       user.Depositor,
			AccountTypeCode: domain.DepositAccountTypeCheck(user.DepositType),
		})
		wdIndexes = append(wdIndexes, i)
	}

	fees, totalFee, isFallback, err := domain.Transfer{}.RequestFeeBulk(c.Request().Context(), transferInputs)
	if err != nil {
		return factory.ErrorSystem.New(err)
	}
	if len(fees) != len(wdIndexes) {
		return factory.ErrorSystem.New(errors.New("fee-bulk result count mismatch"))
	}

	// fee が変わるものだけ更新
	for i, fee := range fees {
		idx := wdIndexes[i]
		if fee != withdrawals[idx].Fee {
			if _, err := factory.DB(c.Request().Context()).
				ID(withdrawals[idx].Id).
				Cols("fee").
				Update(&domain.Withdrawal{Fee: fee}); err != nil {
				logrus.WithError(err).Warnf("failed to update fee for withdrawal %d", withdrawals[idx].Id)
				return factory.ErrorDB.New(err)
			}
			withdrawals[idx].Fee = fee
		}
	}

	// totalFee はジョブレスポンスの合計値を採用
	totalAmount, err := w.SumIncompleteAmount(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// bulk update INCOMPLETE to WAIT
	updated, err := w.BulkIncompleteToWait(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// create withdrawal closing
	var wc domain.WithdrawalClose
	wc.CloseDate = jsonutils.JSONTime(time.Now())
	wc.CloserId = userId
	wc.Total = updated
	wc.Fee = totalFee
	wc.Amount = totalAmount
	wc.Status = domain.WITHDRAWAL_REQUEST_STATUS_WAIT
	_, err = wc.Create(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// create withdrawal closing mapping
	var wcm domain.WithdrawalCloseMapping
	var list []domain.WithdrawalCloseMapping
	for _, withdrawal := range withdrawals {
		tmp := domain.WithdrawalCloseMapping{
			WithdrawalId: withdrawal.Id,
			CloseId:      wc.Id,
		}
		list = append(list, tmp)
	}
	wcm.CreateMulti(c.Request().Context(), list)

	// create log
	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.manage_withdrawal_close", string(domain.WITHDRAWAL_STATUS_WAIT), c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// MIR-1172: 手数料見積もりが固定値へフォールバックしたまま締め処理が行われた場合、
	// 操作した運用側がその場で気付けるようにis_fallbackを返す。河内さんの判断により、
	// 締め処理の実行結果(この応答)にだけ乗せる。WithdrawalCloseへは永続化しない
	// (締め処理の一覧をあとから見るときの表示は対象外)。
	return apirender.RenderSucc(c, http.StatusOK, struct {
		Updated    int64 `json:"updated"`
		IsFallback bool  `json:"is_fallback"`
	}{Updated: updated, IsFallback: isFallback})
}

func (WithdrawalApiController) GetResult(c echo.Context) error {
	var si SearchInput
	if err := c.Bind(&si); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if si.MaxResultCount == 0 {
		si.MaxResultCount = DefaultMaxResultCount
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	total, items, err := domain.WithdrawalDetail{}.GetByCloseId(c.Request().Context(), id, si.Sortby, si.Order, si.SkipCount, si.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	dtos := WithdrawalDetailDto{}.NewList(items)
	applyWithdrawalFailureReasons(c.Request().Context(), items, dtos)

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      dtos,
		TotalCount: total,
	})
}

func (WithdrawalApiController) RequestTransferByClosedId(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// 締め行をロックし、多重実行(二重クリック・再送信)で同じ締めが2回処理されるのを
	// 防ぐ(MIR-1175)。ロック中なら待たずに409、ロックできてもStatusがWAITでなければ
	// 既に処理済み・進行中なので409を返す。
	withdrawalClose, err := domain.WithdrawalClose{}.TryLockById(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrWithdrawalCloseLocked) {
			return factory.ErrorWithdrawalCloseLocked.New(err)
		}
		if errors.Is(err, domain.ErrWithdrawalCloseNotFound) {
			return factory.ErrorWithdrawalCloseNotFound.New(err)
		}
		return factory.ErrorDB.New(err)
	}
	if !withdrawalClose.CanRequestTransfer() {
		return factory.ErrorWithdrawalCloseNotWaiting.New(nil)
	}

	// 마감된 출금 의뢰 목록 조회
	// GetByCloseIdはlimit=1000決め打ちで、締めが1000件を超えると超過分が静かに
	// 送金対象から漏れていた(MIR-1181)。上限の無いGetAllByCloseIdに差し替える。
	items, err := domain.WithdrawalDetail{}.GetAllByCloseId(ctx, id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	total := int64(len(items))
	if total == 0 {
		// 対象0件(GetIncompleteWithdrawalが0件でも作られてしまう空のWithdrawalClose)
		// では送金すべきものが無い。ここでstatusをSENDINGへ進めてしまうと、この締めを
		// WAITへ戻す経路がリポジトリ内に無く、以後CanRequestTransferが常にfalseを
		// 返して永久に処理できなくなるため、statusには触れず返す。
		return apirender.RenderSucc(c, http.StatusOK, RequestTransferByClosedIdResult{Total: total})
	}

	// 출금 의뢰 목록을 순회하며 송금 승인 요청
	//
	// MIR-1256: 対象者の中に1人でもinvestorIdが数値でない、またはbank_code/
	// branch_codeが未登録の者がいると、以前はlogrus.Fatal(admin-api全体が落ちる)
	// か、nilポインタ参照のパニック(500、一括送金全体が失敗)のどちらかで、その1件が
	// 残り全員の送金を巻き込んで止めていた。該当する1件だけを送金対象から外して
	// 残りの対象者への送金は進める(buildWithdrawalTransfersForClose)。
	transferList, wtmList, skippedWithdrawalIds, err := buildWithdrawalTransfersForClose(ctx, items)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if len(transferList) == 0 {
		// 全件がデータ不整合でスキップされ、送金するものが無い。この分岐へ来た時点で
		// 全件のFAIL更新はリクエスト単位の暗黙トランザクション内でロールバック済み
		// (session.Begin/Rollbackはapps/admin-api/server/server.goのミドルウェア。
		// エラーを返せば全体がロールバックされる)なので、締め・出金依頼とも
		// WAITのまま変わらない。復旧はデータを直して一括送金を再実行するのみで、
		// 個別再送金・キャンセルの経路は開かない。
		return factory.ErrorIllegalRequest.New(
			fmt.Errorf("skipped withdrawal ids: %v", skippedWithdrawalIds),
			"All withdrawals in this close were skipped due to data issues.",
		)
	}

	err = domain.TransferRequest{}.NewTransferRequest(ctx, transferList)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"transfer_request": transferList,
			"err":              err,
		}).Error("Fail New Transfer Request")
		return factory.ErrorDB.New(err)
	}

	// update transfer date and status
	if err := withdrawalClose.UpdateTransferDateAndStatus(ctx, jsonutils.JSONTime(time.Now())); err != nil {
		return factory.ErrorDB.New(err)
	}

	// create withdrawal transfer mapping
	withdrawalTransferMapping := domain.WithdrawalTransferMapping{}
	if err := withdrawalTransferMapping.CreateMulti(ctx, wtmList); err != nil {
		return factory.ErrorDB.New(err)
	}

	// create log
	err = domain.OperatorHistory{}.Record(ctx, userId, email,
		"menu.manage_accounting.send_manage", "SENDING", c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, RequestTransferByClosedIdResult{
		Total:                total,
		SkippedWithdrawalIds: skippedWithdrawalIds,
	})
}

// RequestTransferByClosedIdResultはMIR-1256の対応。Totalは締めに含まれる出金
// 依頼の件数(スキップされた分も含む、従来どおりの意味)、
// SkippedWithdrawalIdsはデータ不整合のため送金対象から外した出金依頼のIDで、
// 無ければ空になる。
type RequestTransferByClosedIdResult struct {
	Total                int64   `json:"total"`
	SkippedWithdrawalIds []int64 `json:"skipped_withdrawal_ids,omitempty"`
}

func (a WithdrawalApiController) RequestTransferByWithdrawalId(c echo.Context) error {
	withdrawalId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	// 마감된 출금 의뢰 상세 정보 조회
	withdrawalDetail, err := domain.WithdrawalDetail{}.GetById(c.Request().Context(), withdrawalId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// Withdrawal.Status = FAIL 일 때만 개별 재송금 가능
	if withdrawalDetail.Status != domain.WITHDRAWAL_STATUS_FAIL {
		return factory.ErrorIllegalRequest.New(err, "Only failed withdrawal can be transferred.")
	}

	// MIR-1275: 個別再送金ボタンの二度押しで送金レコードが二重に作られる不具合の対応。
	// 直前のwithdrawalDetail.Statusの判定はメモリ上の値のため、同じ出金への個別再送金が
	// 並行して届くと両方ともこの判定を通過してしまう。ReserveWithdrawalForTransferで
	// ステータスをFAIL→WAITへ条件付き更新し、これに成功した側だけがTransfer/
	// WithdrawalTransferMappingの作成へ進めるようにする。この呼び出しが失敗した場合、
	// このリクエスト全体はミドルウェア側の暗黙のトランザクション内で実行されており
	// (pkg/apirender.RenderSuccがハンドラの最後でCommitする)、エラーを返せばリクエスト
	// 全体がロールバックされるため、ステータスだけWAITになって送金レコードが作られない
	// 状態は残らない。
	if err := a.withdrawalService.ReserveWithdrawalForTransfer(c.Request().Context(), withdrawalId); err != nil {
		return err
	}

	// 송금 요청
	transferList := make([]domain.Transfer, 0)
	// MIR-1256: 以前はinvestorIdが数値でない場合にlogrus.Fatal(os.Exit(1))を呼んで
	// おり、この1件のデータ不整合でadmin-api全体が落ち、その時管理画面を使っている
	// 全員の操作が巻き添えで切れていた。bank_code/branch_codeもDBでNULLを許容して
	// いる(Go側は*string)ため、nilのまま参照するとパニックし500になっていた。
	// 両方をwithdrawalTransferValidationErrorへまとめ、エラーを返すだけにする。
	if err := withdrawalTransferValidationError(withdrawalDetail.InvestorId, withdrawalDetail.User.BankCode, withdrawalDetail.User.BranchCode, withdrawalDetail.User.BankAccount, withdrawalDetail.User.DepositType); err != nil {
		logrus.WithField("withdrawal_id", withdrawalId).WithError(err).Error("Withdrawal data is invalid for transfer")
		return factory.ErrorIllegalRequest.New(err)
	}
	// 直前のwithdrawalTransferValidationErrorで数値であることを確認済み
	investorId, _ := strconv.ParseInt(withdrawalDetail.InvestorId, 10, 64)

	t := domain.Transfer{
		TransactionId:         xid.New().String(),
		UserId:                investorId,
		TransferDate:          jsonutils.JSONTime(time.Now()),
		TransferAmount:        withdrawalDetail.Amount,
		AccountNumber:         withdrawalDetail.User.BankAccount,
		BankCode:              *withdrawalDetail.User.BankCode,
		BankName:              withdrawalDetail.User.BankName,
		BranchCode:            *withdrawalDetail.User.BranchCode,
		BranchName:            withdrawalDetail.User.BranchName,
		Depositor:             withdrawalDetail.User.Depositor,
		AccountTypeCode:       domain.DepositAccountTypeCheck(withdrawalDetail.User.DepositType),
		TransactionDetailType: domain.WITHDRAWAL_INVESTOR,
		TransferStatus:        domain.TRANSFER_STATUS_WAIT,
	}
	transferList = append(transferList, t)

	err = domain.TransferRequest{}.NewTransferRequest(c.Request().Context(), transferList)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"transfer_request": transferList,
			"err":              err,
		}).Error("Fail New Transfer Request")
		return factory.ErrorDB.New(err)
	}

	// create withdrawal transfer mapping
	wtm := domain.WithdrawalTransferMapping{
		WithdrawalId:  withdrawalId,
		TransactionId: t.TransactionId,
	}
	_, err = wtm.Create(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// create log
	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.send_manage", "SENDING", c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, wtm)
}
