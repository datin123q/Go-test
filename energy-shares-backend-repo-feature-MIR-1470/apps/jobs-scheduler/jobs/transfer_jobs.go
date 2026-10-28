package jobs

import (
	"context"
	"energyshares-job-scheduler/pkg/gmo_aozora"
	"energyshares-job-scheduler/pkg/timeutils"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/antihax/optional"
	client "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/gmoaozorautils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

const (
	GMO_TRANSFER_STATUS_REMAND  = "3"  // 差戻
	GMO_TRANSFER_STATUS_SUCCESS = "20" // 手続済
	GMO_TRANSFER_STATUS_FAIL    = "30" // 不能・組戻あり
	GMO_TRANSFER_STATUS_CANCEL  = "40" // 手続不成立
)

// transferTransactionIds は Transfer のスライスから transaction_id だけを拾う。
// 一括処理のログで対象範囲を追跡できるようにするための最小限の識別子(MIR-1116)。
func transferTransactionIds(transfers []domain.Transfer) []string {
	return lo.Map(transfers, func(t domain.Transfer, _ int) string {
		return t.TransactionId
	})
}

// httpStatusCode は *http.Response からステータスコードだけを取り出す。
//
// *http.Response をそのまま logrus.Fields に渡すと、標準の JSON エンコーダは
// Request.GetBody(関数値)を含む構造体を丸ごと直列化しようとして失敗する
// (json: unsupported type: func() (io.ReadCloser, error))。JSONFormatter
// 使用時はこの失敗でログ行全体が標準出力に出ず、"Failed to obtain reader"
// が標準エラーに残るだけになり、同じ Fields にある error_code 等の
// 診断情報まで失われる。ステータスコードだけを拾えば直列化に失敗しない
// (MIR-1116)。
func httpStatusCode(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

// sumTransferFeesは、1件ごとの実際の手数料(Transfer.TransferFee)を合計して返す。
// 手数料見積もりAPIが失敗しても本振込は止めない(MIR-1123)ため、依頼の手数料
// 合計(TransferRequest.TransferTotalFee)は見積もり失敗時に0円のまま残り得る。
// 結果確認ジョブで銀行から実際の手数料が確定したあと、この関数で合計を
// 入れ直す。
func sumTransferFees(transfers []domain.Transfer) int64 {
	var sum int64
	for _, t := range transfers {
		sum += t.TransferFee
	}
	return sum
}

// isTransferErrorDetailIndexValid は、GMOのTransferErrorDetail.ItemId(1始まり)をAtoiした値が
// Transfers への1始まりインデックスとして安全かを確認する。GMOが想定外の値を返した場合に
// Transfers[idx-1] が範囲外アクセスとなり panic するのを避けるためのガード(MIR-1127)。
func isTransferErrorDetailIndexValid(idx, transferCount int) bool {
	return idx >= 1 && idx <= transferCount
}

func (a *AozoraJobs) BulkTransferJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_BULK_TRANSFER,
		JobName: jobNameBulkTransfer,
	}, a.bulkTransferJob)
}

func (a *AozoraJobs) bulkTransferJob(ctx context.Context) error {
	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return errors.Wrap(err, "fail to get corporate client")
	}

	_, requests, err := domain.TransferRequest{}.GetTransferRequestByRequestStatus(ctx, domain.TRANSFER_REQUEST_WAIT)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":  err,
			"time": time.Now().String(),
		}).Error("Get TransferRequest By RequestStatus")
		return err
	}

	// NOTE: このファイルの "request" は TransferRequest を丸ごと載せており、
	// 受取人氏名・口座番号・振込金額が Info で残る。MIR-1089 ではアクセスログと
	// イベントログの範囲に絞ったため、GMO 振込まわりは手を付けていない。
	// 別チケットで識別子だけに絞ること。
	// 対象は `grep -n '"request":' apps/jobs-scheduler/jobs/transfer_jobs.go`。
	logrus.WithFields(logrus.Fields{
		"count": len(requests),
		"time":  time.Now().String(),
	}).Info("Transfer Request Start ")

	jobExecuteTime, err := domain.JobExecuteTime{}.GetJobLastExecuteTime(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to get job last execute time at bulk transfer job")
	}

	account, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_WITHDRAWAL)
	if err != nil {
		return errors.Wrap(err, "fail to get account by account use type at bulk transfer job")
	}

	withdrawalIds := make([]int64, 0)

	var btrs []client.BulkTransfer
	for i := range requests {
		btrs = make([]client.BulkTransfer, 0, 0)
		for j, t := range requests[i].Transfers {
			amount := strconv.FormatInt(requests[i].Transfers[j].TransferAmount, 10)
			if err != nil {
				return err
			}
			itemId := strconv.Itoa(requests[i].Transfers[j].ItemId)
			// 処理は止めずログのみ残す(MIR-1178)。切り捨ては別人への誤送金では
			// なく名義相違による組戻しを招くだけのため。受取人名そのものは
			// ログへ出さない。
			beneficiaryName := gmoaozorautils.ZenkakuToHankaku(requests[i].Transfers[j].Depositor)
			if depositorLength := gmoaozorautils.HankakuLength(beneficiaryName); depositorLength > gmoaozorautils.MaxDepositorLength {
				logrus.WithFields(logrus.Fields{
					"transferId":        t.Id,
					"transferRequestId": requests[i].Id,
					"itemId":            itemId,
					"transactionId":     t.TransactionId,
					"depositorLength":   depositorLength,
					"time":              timeutils.LocalZoneTime(),
				}).Warn(gmoaozorautils.DepositorTruncationWarning)
			}
			bt := client.BulkTransfer{
				ItemId:                itemId,
				BeneficiaryBankCode:   requests[i].Transfers[j].BankCode,
				BeneficiaryBranchCode: requests[i].Transfers[j].BranchCode,
				AccountTypeCode:       requests[i].Transfers[j].AccountTypeCode,
				AccountNumber:         requests[i].Transfers[j].AccountNumber,
				BeneficiaryName:       beneficiaryName,
				TransferAmount:        amount,
			}

			logrus.WithFields(logrus.Fields{
				"itemId":          itemId,
				"transaction_id":  t.TransactionId,
				"BankCode":        t.BankCode,
				"BranchCode":      t.BranchCode,
				"BranchName":      t.BranchName,
				"AccountTypeCode": t.AccountTypeCode,
				"TransferAmount":  amount,
				"time":            time.Now(),
			}).Info("Transfer Request Process ")

			btrs = append(btrs, bt)
		}

		totalAmountStr := strconv.FormatInt(requests[i].RequestTotalAmount, 10)
		totalCntStr := strconv.Itoa(requests[i].RequestTotalCnt)

		// MIR-1172: 見積もり側(RequestFee/RequestFeeBulk)とこの本振込の両方が
		// domain.TransferDesignatedDateで同じ計算規則(JSTの翌日)を使う。
		// 呼ばれる時刻が違うため日付そのものが必ず一致するとは限らないが、
		// 銀行側のバリデーションを通す「翌日・日付のみ」という形式は揃う。
		transferDesignatedDate := domain.TransferDesignatedDate(time.Now())

		body := client.BulkTransferRequest{
			AccountId:              account.GetAccountId(),
			TransferDesignatedDate: transferDesignatedDate, // 송금 지정일
			TotalCount:             totalCntStr,
			TotalAmount:            totalAmountStr,
			BulkTransfers:          btrs,
		}
		logrus.WithFields(logrus.Fields{
			"AccountId":              account.GetAccountId(),
			"transfer_request_id":    requests[i].Id,
			"TransferTotalAmount":    totalAmountStr,
			"TransferTotalCount":     totalCntStr,
			"TransferDesignatedDate": transferDesignatedDate,
			"time":                   timeutils.LocalZoneTime(),
		}).Info("Transfer Request Process ")

		res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "BulkTransferApi.BulkTransferFeeUsingPOST", func(ctx context.Context) (client.TransferFeeResponse, *http.Response, error) {
			return corporateClient.BulkTransferApi.BulkTransferFeeUsingPOST(ctx, body, gmo_aozora.GetAccessToken())
		})
		if err != nil {
			errDetail := gmo_aozora.GetSwaggerErrorDetail(err)

			logrus.WithFields(logrus.Fields{
				"event":          "fail transfer fee Post",
				"id":             requests[i].Id,
				"httpStatusCode": httpStatusCode(httpResp),
				"errDetail":      errDetail,
				"time":           timeutils.LocalZoneTime(),
			}).Error("Fail Transfer Request Fee Process ")

			errorResponse, err := ModelToErrorResponse(errDetail)
			if err != nil {
				// errDetail が JSON でない場合(GMOが502等でHTMLを返す場合など)。エラーコードは取得できないが、下の errorResponse != nil 分岐は元々nilを許容しているため、ここで打ち切らず後続リクエストの処理を続ける(MIR-1127)。
				logrus.WithFields(logrus.Fields{
					"id":   requests[i].Id,
					"err":  err,
					"time": timeutils.LocalZoneTime(),
				}).Warn("fail errorResponse convert")
				errorResponse = nil
			}
			// MIR-1123: 手数料の見積もりに失敗しても、依頼のRequestStatus/ErrorCodeは
			// 書き換えない。このあと本振込(BulkTransferRequestUsingPOST)が成功すると
			// TRANSFER_REQUEST_SUCCESSで上書きされ、ErrorCodeだけが取り残されて
			// 「成功なのにエラーコードがある」依頼ができてしまっていた(desk、JIRA
			// コメント13888)。振込額は見積もりに左右されないため、ここでは失敗を
			// ログに残すだけにして本振込へ進む(手数料の合計は、本振込後の結果確認
			// ジョブでsumTransferFeesにより実際の値へ入れ直す)。
			if errorResponse != nil {
				logrus.WithFields(logrus.Fields{
					"id":         requests[i].Id,
					"error_code": errorResponse.ErrorCode,
					"error_msg":  errorResponse.ErrorMessage,
					"time":       timeutils.LocalZoneTime(),
				}).Error("Fail Transfer Request Fee Process: error detail")
			}
		}

		// 手数料見積もりAPI自体が失敗した場合、res はゼロ値のままここへ落ちてくる(228行に continue/return が無く、
		// 手数料が見積もれなくても実際の送金依頼へ進む既存の挙動)。この分岐は元々 res.TransferFeeDetails が
		// 空であることを前提に何もしなかった(range が0回)ため、以下の件数チェック・パース処理は
		// fee 取得が成功した場合(feeRequestSucceeded)に限定し、既存の挙動を変えないようにする(MIR-1127)。
		feeRequestSucceeded := err == nil
		if feeRequestSucceeded {
			fee, parseErr := strconv.ParseInt(res.TotalFee, 10, 0)
			if parseErr == nil {
				requests[i].TransferTotalFee = fee
			} else {
				// MIR-1123: 以前は握りつぶして手数料0円のままDBへ保存していた。
				// 値そのものは後続の結果確認ジョブ(sumTransferFees)で実際の
				// 手数料へ入れ直されるため、ここではログだけ残す。
				logrus.WithFields(logrus.Fields{
					"transfer_request_id": requests[i].Id,
					"total_fee":           res.TotalFee,
					"err":                 parseErr,
					"time":                timeutils.LocalZoneTime(),
				}).Error("fail parse TotalFee")
			}

			// GMOが返す手数料明細の件数は、こちらが依頼した送金明細(requests[i].Transfers)の件数と
			// 一致している前提でインデックスアクセスしている。件数不一致のまま進むと配列外アクセスで
			// panic する。この手数料明細は実際の送金額(body.BulkTransfers)には影響せず、
			// Transfers[].TransferFee という記帳用の値にしか使わないため、件数が合わない場合は
			// ジョブ全体やこの1件を止めるのではなく、パースだけ諦めて実際の送金依頼へ進む(MIR-1127)。
			if len(res.TransferFeeDetails) != len(requests[i].Transfers) {
				logrus.WithFields(logrus.Fields{
					"transfer_request_id": requests[i].Id,
					"response":            len(res.TransferFeeDetails),
					"request":             len(requests[i].Transfers),
					"time":                timeutils.LocalZoneTime(),
				}).Error("transfer fee detail count mismatch")
			} else {
				for j := range res.TransferFeeDetails {
					requests[i].Transfers[j].TransferFee, _ = strconv.ParseInt(res.TransferFeeDetails[j].TransferFee, 10, 0)
				}
			}
		}

		res2, httpResp2, err2 := gmo_aozora.CallWith429Retry(ctx, "BulkTransferApi.BulkTransferRequestUsingPOST", func(ctx context.Context) (client.BulkTransferRequestResponse, *http.Response, error) {
			return corporateClient.BulkTransferApi.BulkTransferRequestUsingPOST(ctx, body, gmo_aozora.GetAccessToken())
		})
		if err2 != nil {
			requests[i].RequestStatus = domain.TRANSFER_REQUEST_FAIL

			for j := range requests[i].Transfers {
				requests[i].Transfers[j].TransferStatus = domain.TRANSFER_STATUS_FAIL
			}

			logrus.WithFields(
				logrus.Fields{
					"event": "fail transfer request Post",
					"id":    requests[i].Id,
					"err":   err2,
					"time":  timeutils.LocalZoneTime(),
				}).Error("Fail Transfer Request ")

			if httpResp2 != nil {
				logrus.WithFields(
					logrus.Fields{
						"id":             requests[i].Id,
						"httpStatusCode": httpStatusCode(httpResp2),
						"time":           timeutils.LocalZoneTime(),
					}).Error("Fail Transfer Request for httpResp2")
			}

			errDetail := gmo_aozora.GetSwaggerErrorDetail(err2)
			logrus.WithFields(logrus.Fields{
				"id":        requests[i].Id,
				"res":       res2,
				"errDetail": errDetail,
				"time":      timeutils.LocalZoneTime(),
			}).Error("Fail Transfer Request with Error Detail")

			transferErrorResponse, err := ModelToTransferErrorResponse(errDetail)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"errDetail": errDetail,
					"err":       err,
					"time":      time.Now(),
				}).Error("Fail Error Detail Convert ")
			}

			if transferErrorResponse != nil {
				requests[i].ErrorCode = transferErrorResponse.ErrorCode
				requests[i].ErrorMsg = transferErrorResponse.ErrorMessage

				logrus.WithFields(logrus.Fields{
					"id":                    requests[i].Id,
					"transferCount":         len(requests[i].Transfers),
					"transferErrorResponse": transferErrorResponse,
					"time":                  time.Now(),
				}).Info("Fail Transfer Error Response")

				for _, e := range transferErrorResponse.TransferErrorDetails {
					idx, err := strconv.Atoi(e.ItemId)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"id":            requests[i].Id,
							"transferCount": len(requests[i].Transfers),
							"itemId":        e.ItemId,
							"err":           err,
							"time":          time.Now(),
						}).Info("Fail to parse transferErrorResponse.TransferErrorDetails.ItemId to int")
						break
					}
					// idx は GMO の ItemId(1始まり)由来。想定外の値が返る、または ErrorDetails が空で返るケースは
					// Transfers[idx-1] や ErrorDetails[0] へのインデックスアクセスが範囲外になり panic するため、
					// 事前に範囲を確認してから代入する(MIR-1127)。
					if !isTransferErrorDetailIndexValid(idx, len(requests[i].Transfers)) {
						logrus.WithFields(logrus.Fields{
							"id":            requests[i].Id,
							"transferCount": len(requests[i].Transfers),
							"itemId":        e.ItemId,
							"time":          time.Now(),
						}).Error("transferErrorResponse.TransferErrorDetails.ItemId is out of range")
						continue
					}
					if len(e.ErrorDetails) == 0 {
						logrus.WithFields(logrus.Fields{
							"id":     requests[i].Id,
							"itemId": e.ItemId,
							"time":   time.Now(),
						}).Error("transferErrorResponse.TransferErrorDetails.ErrorDetails is empty")
						continue
					}
					requests[i].Transfers[idx-1].ErrorCode = e.ErrorDetails[0].ErrorDetailsCode
					requests[i].Transfers[idx-1].ErrorMsg = e.ErrorDetails[0].ErrorDetailsMessage
				}
			}

			// 출금 의뢰 실패 - 관리자가 출금 의뢰 마감 목록에서 송금 승인 한 경우
			// update withdrawal - {status: FAIL}
			for j := range requests[i].Transfers {
				if requests[i].Transfers[j].TransactionDetailType == domain.WITHDRAWAL_INVESTOR {
					withdrawalId, err := updateFailedWithdrawal(ctx, requests[i].Transfers[j])
					if err == nil {
						withdrawalIds = append(withdrawalIds, withdrawalId)
					}
				}
			}

		} else {
			requests[i].ResultCode = domain.TRANSFER_RESULT_CODE(res2.ResultCode)
			requests[i].ApplyNo = res2.ApplyNo
			requests[i].ApplyEndDatetime = res2.ApplyEndDatetime
			requests[i].RequestStatus = domain.TRANSFER_REQUEST_SUCCESS
			requests[i].CompletedTransfer = domain.COMPLETE_TRANSFER_WAIT // 송금 완료 대기

			logrus.WithFields(logrus.Fields{
				"id":               requests[i].Id,
				"applyNo":          res2.ApplyNo,
				"applyEndDatetime": res2.ApplyEndDatetime,
				"resultCode":       res2.ResultCode,
				"time":             time.Now(),
			}).Info("success transfer request")
		}

		err = requests[i].Update(ctx)
		if err != nil {
			return errors.Wrap(err, "fail requests update at transfer request")
		}
		_, err = domain.Transfer{}.UpdateMany(ctx, requests[i].Transfers)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":  err,
				"time": time.Now(),
			}).Error("Fail Transfer Update Many")
			return err
		}

	}

	// 출금 의뢰 송금 완료 처리 - 관리가자 출금 의뢰 마감 목록에서 송금 승인 한 경우
	// update withdrawal_close - {status: COMPLETE}
	updateWithdrawalClose(ctx, withdrawalIds)

	if err := jobExecuteTime.UpdateTransferRequestTime(ctx, jsonutils.JSONTime(time.Now())); err != nil {
		return errors.Wrap(err, "fail update transfer request time")
	}

	logrus.WithFields(logrus.Fields{
		"count": len(requests),
		"time":  time.Now().String(),
	}).Info("transfer request end ")
	return nil
}

func (a *AozoraJobs) BulkTransferResultCheckJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_BULK_TRANSFER_RESULT_CHECK,
		JobName: jobNameBulkTransferResultCheck,
	}, a.bulkTransferResultCheckJob)
}

func (a *AozoraJobs) bulkTransferResultCheckJob(ctx context.Context) error {
	logrus.WithFields(logrus.Fields{
		"time": time.Now(),
	}).Info("Start Transfer Result Check")

	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return errors.Wrap(err, "fail to get corporate client")
	}
	jobExecuteTime, err := domain.JobExecuteTime{}.GetJobLastExecuteTime(ctx)
	if err != nil {
		return err
	}
	account, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_WITHDRAWAL)
	if err != nil {
		return errors.Wrap(err, "fail to get account by account use type")
	}

	_, trs, err := domain.TransferRequest{}.GetTransferRequestByResultCode(ctx, domain.TRANSFER_RESULT_WAIT)
	if err != nil {
		return errors.Wrap(err, "fail Get TransferRequest By Result Code = TRANSFER_RESULT_WAIT")
	}

	withdrawalIds := make([]int64, 0)

	for i := range trs {
		res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "BulkTransferApi.BulkTransferRequestResultUsingGET", func(ctx context.Context) (client.TransferRequestResultResponse, *http.Response, error) {
			return corporateClient.BulkTransferApi.BulkTransferRequestResultUsingGET(ctx, account.GetAccountId(), trs[i].ApplyNo, gmo_aozora.GetAccessToken())
		})
		if err != nil {

			logrus.WithFields(logrus.Fields{
				"category":            "transfer result check",
				"transfer_request_id": trs[i].Id,
				"httpStatusCode":      httpStatusCode(httpResp),
				"err":                 err,
			}).Error("Fail BulkTransfer Request Result Get ")
			continue
		}
		if trs[i].ApplyNo == res.ApplyNo {
			trs[i].ApplyEndDatetime = res.ApplyEndDatetime
			trs[i].ResultCode = domain.TRANSFER_RESULT_CODE(res.ResultCode)

			if trs[i].ResultCode == domain.TRANSFER_RESULT_SUCCESS {
				applyEndTime, err := jsonutils.ParseJSONTimeJST("2006-01-02T15:04:05-07:00", trs[i].ApplyEndDatetime)
				if err != nil {
					return errors.Wrap(err, "fail Parse to ApplyEnd Time")
				}

				trs[i].CompletedTransfer = domain.COMPLETE_TRANSFER_WAIT // 송금 완료 대기

				for j := range trs[i].Transfers {
					trs[i].Transfers[j].TransferDate = applyEndTime
					trs[i].Transfers[j].TransferStatus = domain.TRANSFER_STATUS_WAIT
				}

				_, err = domain.Transfer{}.UpdateMany(ctx, trs[i].Transfers)
				if err != nil {
					return errors.Wrap(err, "fail Transfer Time Update")
				}

				logrus.WithFields(logrus.Fields{
					"transfer_request_id": trs[i].Id,
					"transaction_ids":     transferTransactionIds(trs[i].Transfers),
					"time":                time.Now(),
				}).Info("success transfer request result check")

			} else if trs[i].ResultCode == domain.TRANSFER_RESULT_FAIL {
				for j := range trs[i].Transfers {

					trs[i].Transfers[j].TransferStatus = domain.TRANSFER_STATUS_FAIL // 송금 실패

					if trs[i].Transfers[j].TransactionDetailType == domain.WITHDRAWAL_INVESTOR {
						// 출금 의뢰 실패 - 관리자가 출금 의뢰 마감 목록에서 송금 승인 한 경우
						// update withdrawal - {status: FAIL}
						withdrawalId, err := updateFailedWithdrawal(ctx, trs[i].Transfers[j])
						if err == nil {
							withdrawalIds = append(withdrawalIds, withdrawalId)
						}
					}
				}
				_, err = domain.Transfer{}.UpdateMany(ctx, trs[i].Transfers)
				if err != nil {
					return errors.Wrap(err, "fail Transfer Time Update")
				}
				logrus.WithFields(logrus.Fields{
					"transfer_request_id": trs[i].Id,
					"transaction_ids":     transferTransactionIds(trs[i].Transfers),
					"time":                time.Now(),
				}).Info("fail transfer request result check")
			}

			if err := trs[i].Update(ctx); err != nil {
				return errors.Wrap(err, "fail transfer request update")
			}
		}
	}

	// 출금 의뢰 송금 완료 처리 - 관리가자 출금 의뢰 마감 목록에서 송금 승인 한 경우
	// update withdrawal_close - {status: COMPLETE}
	updateWithdrawalClose(ctx, withdrawalIds)

	jobExecuteTime.UpdateTransferRequestTime(ctx, jsonutils.JSONTime(time.Now()))
	logrus.WithFields(logrus.Fields{
		"time": time.Now(),
	}).Info("Success Transfer Result Check Jobs")

	return nil
}

// BulkTransferStatusCheckJob 종합 송금 상태 조회
func (a *AozoraJobs) BulkTransferStatusCheckJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_BULK_TRANSFER_STATUS_CHECK,
		JobName: jobNameBulkTransferStatusCheck,
	}, a.bulkTransferStatusCheckJob)
}

func (a *AozoraJobs) bulkTransferStatusCheckJob(ctx context.Context) error {
	logrus.WithFields(logrus.Fields{
		"time": time.Now(),
	}).Info("Start Transfer Status Check")

	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return errors.Wrap(err, "fail to get corporate client at transfer status check")
	}
	jobExecuteTime, err := domain.JobExecuteTime{}.GetJobLastExecuteTime(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to get job last execute time at transfer status check")
	}
	account, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_WITHDRAWAL)
	if err != nil {
		return errors.Wrap(err, "fail to get account by account use type at transfer status check")
	}

	_, request, err := domain.TransferRequest{}.GetTransferRequestByCompletedTransfer(ctx, domain.COMPLETE_TRANSFER_WAIT)
	if err != nil {
		return errors.Wrap(err, "fail to get transfer request by completed transfer")
	}

	for _, req := range request {
		logrus.WithFields(logrus.Fields{
			"transfer_request_id": req.Id,
			"time":                time.Now().String(),
		}).Info("start transfer request")

		item, httpResp, err := getTransferStatusWithApplyNo(ctx, corporateClient, account, true, req.ApplyNo)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"transfer_request_id": req.Id,
				"httpStatusCode":      httpStatusCode(httpResp),
				"errDetail":           gmo_aozora.GetSwaggerErrorDetailString(err),
				"time":                time.Now().String(),
			}).Error("fail get transfer status with apply no")

			return errors.Wrap(err, "fail get transfer status with apply no")
		}

		if !hasBulkTransferDetails(item) {
			logrus.WithFields(logrus.Fields{
				"transfer_request_id": req.Id,
				"detailInfoResult":    item.DetailInfoResult,
				"bulkTransferDetails": len(item.BulkTransferDetails),
				"time":                time.Now().String(),
			}).Info("in complete transfer status continue")
			continue
		}

		// GMOのレスポンス形状が想定と異なる(件数不一致等、errBulkTransferStatusResponseInvalid)場合は、
		// この1件を諦めてジョブ全体を止めると、この依頼が COMPLETE_TRANSFER_WAIT のまま残り続けて
		// 次回以降も同じ箇所で毎回止まり、後続の依頼が永久に処理されなくなるため、直前の
		// hasBulkTransferDetails の不成立時と同じくこの1件だけスキップして次へ進む。一方、
		// transferDetailResponseParse/checkTransferProcedures/Transfers[idx].Update(ctx) 由来の
		// DB書き込み失敗等はGMOのレスポンスと無関係の内部障害であり、continueで握りつぶすと
		// ジョブが正常終了扱いになり検知できなくなるため、こちらは従来どおりジョブを失敗させる(MIR-1127)。
		if err := bulkTransferInfosTask(ctx, item, &req); err != nil {
			logrus.WithFields(logrus.Fields{
				"transfer_request_id": req.Id,
				"transferCount":       len(req.Transfers),
				"err":                 err,
				"time":                time.Now().String(),
			}).Error("fail to bulk transfer infos task")
			if errors.Is(err, errBulkTransferStatusResponseInvalid) {
				continue
			}
			return errors.Wrap(err, "fail to bulk transfer infos task")
		}

		// MIR-1123: 手数料見積もりAPIの失敗時、依頼の手数料合計(TransferTotalFee)は
		// 0円のまま残っていた。ここで明細ごとの実際の手数料(直前のbulkTransferInfosTask/
		// transferDetailResponseParseが設定したreq.Transfers[].TransferFee)を合計して
		// 入れ直し、見積もりの成否に関わらず最終的に正しい合計へ揃える。受取口座が
		// 全員GMOあおぞら宛て(同行宛ては0円)の場合、合計が本当に0円になり、
		// 通常のUpdate()だとXORMがゼロ値を書き込まず古い合計が残るため、
		// UpdateTransferTotalFee(MustCols)で確実に書き込む。
		req.TransferTotalFee = sumTransferFees(req.Transfers)
		req.CompletedTransfer = domain.COMPLETE_TRANSFER_SUCCESS
		if err := req.UpdateTransferTotalFee(ctx); err != nil {
			return errors.Wrap(err, "fail to update transfer total fee")
		}
		logrus.WithFields(logrus.Fields{
			"transfer_request_id": req.Id,
			"time":                time.Now().String(),
		}).Info("Success Transfer Request Status Update")

	}
	if err := jobExecuteTime.UpdateTransferStatusTime(ctx, jsonutils.JSONTime(time.Now())); err != nil {
		return errors.Wrap(err, "fail to update transfer status time")
	}
	logrus.WithFields(logrus.Fields{
		"time": time.Now(),
	}).Info("Success Transfer Status Check Jobs")
	return nil
}

// hasBulkTransferDetails 송금 상태 조회 결과에 송금 상세 정보가 있는지 확인
func hasBulkTransferDetails(item *client.BulkTransferStatusResponse) bool {
	if item == nil {
		return false
	}
	if !item.DetailInfoResult {
		return false
	}
	if len(item.BulkTransferDetails) == 0 {
		return false
	}

	if isIncompleteTransfers(item.BulkTransferDetails[0].TransferStatus) {
		return false
	}
	return true
}

// errBulkTransferStatusResponseInvalid は、validateBulkTransferStatusResponse がGMOのレスポンス形状の
// 問題(明細なし・件数不一致等)を検出したことを示す番兵エラー。呼び出し元はこの種類のエラーだけを
// 安全にスキップ対象と判断できる。DB書き込み失敗など他の内部エラーと区別せず一律スキップすると、
// ジョブが正常終了扱いになり本来検知すべき障害が見えなくなるため(MIR-1127、レビュー指摘対応)。
var errBulkTransferStatusResponseInvalid = errors.New("bulk transfer status response invalid")

// validateBulkTransferStatusResponse は item.BulkTransferDetails[0].BulktransferResponses[0] への無条件アクセスを避けるためのヘルパー(MIR-1127)。GMOが明細を返さない、または明細件数がこちらの依頼件数と一致しない場合、以降のインデックスアクセス(req.Transfers[idx] 等)が範囲外になり panic するため、その手前で必ずエラーとして弾く。
func validateBulkTransferStatusResponse(item *client.BulkTransferStatusResponse, req *domain.TransferRequest) (*client.BulkTransferDetail, *client.BulkTransferResponse, error) {
	if item == nil {
		return nil, nil, errors.Wrap(errBulkTransferStatusResponseInvalid, "bulk transfer status response is nil")
	}
	if req == nil {
		return nil, nil, errors.Wrap(errBulkTransferStatusResponseInvalid, "transfer request is nil")
	}
	if len(item.BulkTransferDetails) == 0 {
		return nil, nil, errors.Wrap(errBulkTransferStatusResponseInvalid, "bulk transfer details are empty")
	}
	if len(req.Transfers) == 0 {
		return nil, nil, errors.Wrap(errBulkTransferStatusResponseInvalid, "transfer request details are empty")
	}

	detail := &item.BulkTransferDetails[0]
	if len(detail.BulktransferResponses) == 0 {
		return nil, nil, errors.Wrap(errBulkTransferStatusResponseInvalid, "bulk transfer responses are empty")
	}

	response := &detail.BulktransferResponses[0]
	if len(response.BulkTransferInfos) != len(req.Transfers) {
		return nil, nil, errors.Wrapf(
			errBulkTransferStatusResponseInvalid,
			"bulk transfer info count mismatch: response=%d request=%d",
			len(response.BulkTransferInfos),
			len(req.Transfers),
		)
	}

	return detail, response, nil
}

func bulkTransferInfosTask(ctx context.Context, item *client.BulkTransferStatusResponse, req *domain.TransferRequest) error {
	detail, response, err := validateBulkTransferStatusResponse(item, req)
	if err != nil {
		return errors.Wrap(err, "invalid bulk transfer status response")
	}

	for idx, infos := range response.BulkTransferInfos {
		if parseErr := transferDetailResponseParse(infos, &req.Transfers[idx], response.TransferDesignatedDate); parseErr != nil {
			return errors.Wrap(parseErr, "fail to transfer detail response")
		}

		if checkErr := checkTransferProcedures(ctx, infos, &req.Transfers[idx], detail.TransferStatus); checkErr != nil {
			return errors.Wrap(checkErr, "fail to check transfer procedures")
		}
		if updateErr := req.Transfers[idx].Update(ctx); updateErr != nil {
			return errors.Wrap(updateErr, "fail to update transfer")
		}
	}
	return nil
}

// transferDetailResponseParse 송금 상태 상세 정보 확인
func transferDetailResponseParse(infos client.BulkTransferInfo, tr *domain.Transfer, transferDesignatedDate string) error {
	if len(infos.TransferDetailResponses) > 0 {
		for j := range infos.TransferDetailResponses {
			if infos.TransferDetailResponses[j].TransferFee != "" {
				fee, err := strconv.ParseInt(infos.TransferDetailResponses[j].TransferFee, 10, 64)
				if err != nil {
					return errors.Wrap(err, "fail Parse to TransferFee")
				}
				tr.TransferFee = fee
			}

			if transferDesignatedDate != "" {
				designateDate, err := jsonutils.ParseJSONTimeJST(time.DateOnly, transferDesignatedDate)
				if err != nil {
					return errors.Wrap(err, "fail Parse to TransferDesignatedDate")
				}
				tr.TransferDate = designateDate
			}
		}
	}
	return nil
}

// repaymentStatusMessage 송금 상태 상세 정보 반화
func repaymentStatusMessage(code string) string {
	switch code {
	case "1":
		return "組戻手続中" // 재조합 절차 중
	case "2":
		return "組戻済" // 재조합 완료
	case "3":
		return "組戻不成立" // 재조합 불성립
	default:
		return ""
	}
}

// transferDetailStatusMessage 송금 상태 메시지 반환
func transferDetailStatusMessage(code string) string {
	switch code {
	case "1":
		return "手続済" // 절차 완료
	case "2":
		return "手続不成立" // 절차 불성립
	default:
		return ""
	}
}

// checkTransferProcedures 송금 상태 상세 정보 확인하고 실패 했는지 성공했는지 확인
func checkTransferProcedures(ctx context.Context, infos client.BulkTransferInfo, tr *domain.Transfer, transferStatus string) error {
	ids := []int64{}

	if len(infos.UnableDetailInfos) > 0 {

		logrus.WithFields(logrus.Fields{
			"transfer_request_id":       tr.RequestId,
			"transfer_id":               tr.Id,
			"transaction_id":            tr.TransactionId,
			"item_id":                   infos.ItemId,
			"unableDetailCount":         len(infos.UnableDetailInfos),
			"firstTransferDetailStatus": infos.UnableDetailInfos[0].TransferDetailStatus,
			"firstIsRepayment":          infos.UnableDetailInfos[0].IsRepayment,
			"time":                      time.Now().String(),
			"transfer status":           tr.TransferStatus,
		}).Info("Unable Detail Info")

		err := unableDetailProcess(ctx, infos, tr, &ids)
		if err != nil {
			return errors.Wrap(err, "fail to unable detail process")
		}
	} else if transferStatus == GMO_TRANSFER_STATUS_REMAND {
		// NOTE: 差戻の場合は「送金キャンセル」とする
		tr.TransferStatus = domain.TRANSFER_STATUS_CANCEL
		tr.ErrorMsg = "差し戻しのため、送金キャンセル"
		withdrawId, err := updateFailedWithdrawal(ctx, *tr)
		if err != nil {
			return errors.Wrap(err, "fail to update failed withdrawal")
		}
		ids = append(ids, withdrawId)
	} else {
		err := successTransferStatus(ctx, infos, tr, &ids)
		if err != nil {
			return errors.Wrap(err, "fail to success transfer status")
		}

	}
	updateWithdrawalClose(ctx, ids)
	return nil
}

// successTransferStatus 송금 성공 상태 처리
func successTransferStatus(ctx context.Context, infos client.BulkTransferInfo, tr *domain.Transfer, ids *[]int64) error {
	tr.TransferStatus = domain.TRANSFER_STATUS_SUCCESS

	if tr.TransactionDetailType == domain.WITHDRAWAL_INVESTOR {
		// 출금 의뢰 성공 - 관리자가 출금 의뢰 마감 목록에서 송금 승인 한 경우
		// update withdrawal - {status: COMPLETE, complete_date: NOW}
		withdrawId, err := updateSuccessWithdrawal(ctx, *tr)
		if err != nil {
			return errors.Wrap(err, "fail to update failed withdrawal")
		}
		*ids = append(*ids, withdrawId)
	}
	now := time.Now()
	ua, err := domain.UserAccount{}.GetByUserId(ctx, tr.UserId)
	if err != nil {
		return errors.Wrap(err, "fail to get user account by user id")
	}
	if err := (domain.UserTransaction{}).WithdrawalUserTransaction(ctx, tr.UserId, tr.TransferAmount, tr.ProjectId, domain.EQUITY, tr.AccountNumber, tr.TransactionDetailType, jsonutils.JSONTime(now)); err != nil {
		logrus.WithError(err).Error("failed to create user transaction for transfer success")
	}
	logrus.WithFields(logrus.Fields{
		"transfer_request_id": tr.RequestId,
		"transfer_id":         tr.Id,
		"transaction_id":      tr.TransactionId,
		"item_id":             infos.ItemId,
		"time":                now.String(),
		"transfer status":     tr.TransferStatus,
	}).Info("Success Transfer Status")

	err = domain.SendCustomerLedgerWithdrawalEvent(ctx, tr.UserId, ua.UserName, tr.TransferAmount, now)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"transfer_id": tr.Id,
			"err":         err,
		}).Error("fail to send customer ledger event")
	}
	return nil
}

// unableDetailProcess unableDetailInfos가 있을 경우에 대한 송금 환불 및 취소 처리 로직
func unableDetailProcess(ctx context.Context, infos client.BulkTransferInfo, tr *domain.Transfer, ids *[]int64) error {
	const (
		TRANSFER_PROCEDURES_COMPLETED   = "1"
		TRANSFER_PROCEDURES_INCOMPLETED = "2"
	)

	for j := range infos.UnableDetailInfos {
		if infos.UnableDetailInfos[j].TransferDetailStatus == TRANSFER_PROCEDURES_INCOMPLETED {
			tr.TransferStatus = domain.TRANSFER_STATUS_CANCEL
			tr.ErrorMsg = "送金キャンセル" // 송금 취소
			withdrawId, err := updateFailedWithdrawal(ctx, *tr)
			if err != nil {
				return errors.Wrap(err, "fail to update failed withdrawal")
			}
			*ids = append(*ids, withdrawId)
			break
		}

		if infos.UnableDetailInfos[j].TransferDetailStatus == TRANSFER_PROCEDURES_COMPLETED {
			if infos.UnableDetailInfos[j].IsRepayment {
				tr.TransferStatus = domain.TRANSFER_STATUS_REPAYMENT
				repaymentDate, err := jsonutils.ParseJSONTimeJST(time.DateOnly, infos.UnableDetailInfos[j].RepaymentDate)
				if err != nil {
					return errors.Wrap(err, "fail Parse to RepaymentDate")
				}
				tr.RepaymentDate = repaymentDate
				tr.ErrorMsg = fmt.Sprintf("払戻し -  資金返却日:%s", infos.UnableDetailInfos[j].RepaymentDate) // 환불 - 자금 반환 일자
				withdrawId, err := updateFailedWithdrawal(ctx, *tr)
				if err != nil {
					return errors.Wrap(err, "fail to update failed withdrawal")
				}
				*ids = append(*ids, withdrawId)
			}
		}
	}
	return nil
}

// isIncompleteTransfers 送金が未完了かどうかを確認する
func isIncompleteTransfers(transferStatus string) bool {
	if (transferStatus != GMO_TRANSFER_STATUS_REMAND && transferStatus != GMO_TRANSFER_STATUS_SUCCESS) && (transferStatus != GMO_TRANSFER_STATUS_FAIL) && (transferStatus != GMO_TRANSFER_STATUS_CANCEL) {
		return true
	}
	return false
}

// getTransferStatusWithApplyNo Apply No 를 기준으로 송금 상태를 확인한다.
func getTransferStatusWithApplyNo(ctx context.Context, corporateClient *client.APIClient, account *gmo_aozora.Account, DetailInfoNecessity bool, ApplyNo string) (*client.BulkTransferStatusResponse, *http.Response, error) {
	statusOpt := func() *client.BulkTransferStatusUsingGETOpts {
		opt := client.BulkTransferStatusUsingGETOpts{}
		if ApplyNo != "" {
			opt.ApplyNo = optional.NewString(ApplyNo)
		}
		opt.DetailInfoNecessity = optional.NewBool(DetailInfoNecessity)
		return &opt
	}
	logrus.WithFields(logrus.Fields{
		"category":            "transfer status check",
		"account_id":          account.AccountId,
		"ApplyNo":             ApplyNo,
		"DetailInfoNecessity": DetailInfoNecessity,
	}).Info("Start BulkTransfer Status Get ")

	res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "BulkTransferApi.BulkTransferStatusUsingGET", func(ctx context.Context) (client.BulkTransferStatusResponse, *http.Response, error) {
		return corporateClient.BulkTransferApi.BulkTransferStatusUsingGET(ctx, account.AccountId, "1", gmo_aozora.GetAccessToken(), statusOpt())
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"category":       "transfer status check",
			"httpStatusCode": httpStatusCode(httpResp),
			"errDetail":      gmo_aozora.GetSwaggerErrorDetailString(err),
			"err":            err,
		}).Error("Fail BulkTransfer Status Get ")
		return nil, httpResp, err
	}
	logrus.WithFields(logrus.Fields{
		"category":    "transfer status check",
		"detailCount": res.Count,
	}).Info("Success BulkTransfer Status Get ")
	return &res, httpResp, nil
}

func updateSuccessWithdrawal(ctx context.Context, transfer domain.Transfer) (int64, error) {
	withdrawal := domain.Withdrawal{}
	w, err := withdrawal.GetByTransactionId(ctx, transfer.TransactionId)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"category": "transfer result check",
			"detail":   "withdrawal update when transfer success",
			"err":      err,
		}).Error("Failed get withdrawal by transaction id")
		return -1, errors.Wrap(err, "fail get withdrawal by transaction id")
	}

	w.CompleteDate = transfer.TransferDate
	w.Status = domain.WITHDRAWAL_STATUS_COMPLETE
	w.Fee = transfer.TransferFee

	err = w.Update(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"category": "transfer result check",
			"detail":   "withdrawal update when transfer success",
			"err":      err,
		}).Error("Failed update withdrawal status")
		return w.Id, err
	}

	return w.Id, nil
}

func updateFailedWithdrawal(ctx context.Context, transfer domain.Transfer) (int64, error) {
	withdrawal := domain.Withdrawal{}
	w, err := withdrawal.GetByTransactionId(ctx, transfer.TransactionId)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"category": "transfer result check",
			"detail":   "withdrawal update when transfer fail",
			"err":      err,
		}).Error("Failed get withdrawal by transaction id")
		return -1, errors.Wrap(err, "fail get withdrawal by transaction id")
	}

	w.Status = domain.WITHDRAWAL_STATUS_FAIL
	w.Fee = transfer.TransferFee

	err = w.Update(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"category": "transfer result check",
			"detail":   "withdrawal update when transfer fail",
			"err":      err,
		}).Error("Failed update withdrawal status")
		return w.Id, err
	}

	return w.Id, nil
}

func updateWithdrawalClose(ctx context.Context, withdrawalIds []int64) {
	wc := domain.WithdrawalClose{}
	wcm := domain.WithdrawalCloseMapping{}
	if len(withdrawalIds) > 0 {
		ids, err := wcm.GetCloseIdsByWithdrawalIds(ctx, withdrawalIds)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"category": "transfer result check",
				"detail":   "withdrawal_close update when transfer completed",
				"err":      err,
			}).Error("Failed get withdrawal_close_mapping by withdrawal ids")
		}

		err = wc.UpdateStatusByIds(ctx, ids, domain.WITHDRAWAL_REQUEST_STATUS_COMPLETE)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"category": "transfer result check",
				"detail":   "withdrawal_close update when transfer completed",
				"err":      err,
			}).Error("Failed update withdrawal_close status")
		}

		for _, closeId := range ids {
			err = wc.UpdateFeeById(ctx, closeId)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"category": "transfer result check",
					"detail":   "withdrawal_close update when transfer completed",
					"err":      err,
				}).Error("Failed update withdrawal_close fee")
			}
		}
	}
}
