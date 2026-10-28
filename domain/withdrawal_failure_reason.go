package domain

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// MIR-1170: 出金依頼の一覧と送金管理の「確認」モーダルに出す「失敗の理由」。
// 出金依頼ごとに最新の振込(transfer.id最大)と、その振込バッチ(transfer_request)だけを
// 見て、上から最初に当てはまったものを理由にする。

type WithdrawalFailureKind string

const (
	// WithdrawalFailureKindNone は理由を出さない(対象外の出金依頼)。
	WithdrawalFailureKindNone WithdrawalFailureKind = ""
	// 締めの送金承認で口座情報が不備のため、振込データを作らずにFAILにした。
	WithdrawalFailureKindNoTransferData WithdrawalFailureKind = "NO_TRANSFER_DATA"
	// その出金依頼の振込自身にエラーが記録されている。
	WithdrawalFailureKindTransferError WithdrawalFailureKind = "TRANSFER_ERROR"
	// 同じ振込バッチの別の出金依頼のエラーで、バッチ全体が不成立になった。
	WithdrawalFailureKindBatchCausedByOther WithdrawalFailureKind = "BATCH_CAUSED_BY_OTHER"
	// 振込バッチ全体のエラー(明細の内訳なし)。
	WithdrawalFailureKindBatchError WithdrawalFailureKind = "BATCH_ERROR"
	// 振込の期限切れ(結果コード8)。
	WithdrawalFailureKindBatchExpired WithdrawalFailureKind = "BATCH_EXPIRED"
	// 銀行の応答を読めず失敗扱いにした。銀行側では受け付けられている可能性がある。
	WithdrawalFailureKindResponseUnreadable WithdrawalFailureKind = "RESPONSE_UNREADABLE"
	// 出金依頼の状態と振込の記録が食い違っている。
	WithdrawalFailureKindStatusMismatch WithdrawalFailureKind = "STATUS_MISMATCH"
	// 上のどれにも当てはまらない。
	WithdrawalFailureKindNoRecord WithdrawalFailureKind = "NO_RECORD"
)

const (
	withdrawalFailureMessageNoTransferData     = "口座情報の不備で送金データを作成できず"
	withdrawalFailureMessageBatchExpired       = "振込の期限切れ(結果コード8)"
	withdrawalFailureMessageResponseUnreadable = "銀行の応答を読めず失敗扱い。銀行側で受け付けられている可能性あり。再送金の前に GMO で確認"
	withdrawalFailureMessageStatusMismatch     = "状態と振込記録が一致しない(詳細を確認)"
	withdrawalFailureMessageNoRecord           = "理由の記録なし(詳細を確認)"
	// 振込全体のエラーは、GMOが500/503/504などで返したタイムアウト系のものも同じ形で入る。その場合
	// 銀行側で処理済みのことがあるため、エラーの内容にかかわらず再送金の前の確認を添える。
	withdrawalFailureSuffixBatchError = "。再送金の前に GMO で確認"
)

// WithdrawalFailure は出金依頼1件の「失敗の理由」。Kindが空なら理由を出さない。
type WithdrawalFailure struct {
	Kind              WithdrawalFailureKind
	Message           string
	TransferRequestId int64
}

// WithdrawalFailureSource は、出金依頼ごとの最新の振込(t)とその振込バッチ(r)のうち、
// 理由の判定に要る列だけを持つ。transferとtransfer_requestはerror_code・error_msg・
// id・created等の列名が同じため、取得時はすべて別名を付けて読む(別名が無いとxormが
// 片方へまとめ、220011のバッチで全明細のerror_codeが同じ値に上書きされて、原因の
// 1件が埋もれる)。口座番号や受取人名など、判定に要らない列は読まない。
type WithdrawalFailureSource struct {
	WithdrawalId      int64                   `xorm:"withdrawal_id <-"`
	TransferId        int64                   `xorm:"transfer_id <-"`
	TransferRequestId int64                   `xorm:"transfer_request_id <-"`
	TransferStatus    TRANSFER_STATUS_TYPE    `xorm:"transfer_status_value <-"`
	TransferErrorCode string                  `xorm:"transfer_error_code <-"`
	TransferErrorMsg  string                  `xorm:"transfer_error_msg <-"`
	TransferCreated   jsonutils.JSONTime      `xorm:"transfer_created <-"`
	RequestRowId      int64                   `xorm:"request_row_id <-"` // 0ならtransfer_requestの行が無い
	RequestStatus     TRANSFER_REQUEST_STATUS `xorm:"request_status_value <-"`
	RequestErrorCode  string                  `xorm:"request_error_code <-"`
	RequestErrorMsg   string                  `xorm:"request_error_msg <-"`
	RequestResultCode TRANSFER_RESULT_CODE    `xorm:"request_result_code <-"`
}

func (s *WithdrawalFailureSource) hasRequest() bool {
	return s.RequestRowId != 0
}

// WithdrawalFailureReason は、出金依頼(w)・最新の振込とその振込バッチ(src)・同じ振込
// バッチでerror_codeを持つ別の出金依頼のID(causeWithdrawalIds)から、画面に出す理由を
// 決める。DBを読まず引数だけで決まる。
//
// includeWaitWithFailedTransfer は、WAITのまま振込がFAILになっている出金依頼も対象に
// するか(ZOLOMEは手数料の見積もり失敗でWAITのまま止まるためtrue)。
func WithdrawalFailureReason(w Withdrawal, src *WithdrawalFailureSource, causeWithdrawalIds []int64, includeWaitWithFailedTransfer bool) WithdrawalFailure {
	switch w.Status {
	case WITHDRAWAL_STATUS_FAIL, WITHDRAWAL_STATUS_CANCEL:
	case WITHDRAWAL_STATUS_WAIT:
		if !includeWaitWithFailedTransfer || src == nil || src.TransferStatus != TRANSFER_STATUS_FAIL {
			return WithdrawalFailure{}
		}
	default:
		return WithdrawalFailure{}
	}
	// 送金前の取消。振込が無いのは理由ではなく、そもそも送っていないため。口座不備でFAILになってから
	// 取り消された出金依頼も、振込を一度も作っていなければここに入り、理由は出ない。
	if w.Status == WITHDRAWAL_STATUS_CANCEL && src == nil {
		return WithdrawalFailure{}
	}

	// #0: 振込が無い、または今回の締めより前の振込(前の締めの古い理由を出さないため)。
	if src == nil || transferPredatesClose(src.TransferCreated, w.CloseDate) {
		// 前の締めの振込バッチのIDを返すと、画面からリンクしたときに無関係なバッチへ飛ぶ。
		return failureOf(WithdrawalFailureKindNoTransferData, withdrawalFailureMessageNoTransferData, nil)
	}

	// #1: その出金依頼の振込自身のエラー。#2より先に見るので、220011のバッチでは原因の1件が
	// ここで拾われ、残りが#2になる。
	if src.TransferErrorCode != "" || src.TransferErrorMsg != "" {
		return failureOf(WithdrawalFailureKindTransferError, joinCodeAndMessage(src.TransferErrorCode, src.TransferErrorMsg), src)
	}

	if src.hasRequest() {
		// MIR-1123より前は、手数料の見積もりに失敗すると振込バッチをFAIL+error_codeにしてから本振込へ
		// 進み、本振込が成功してもerror_codeだけが残った。status=FAILのときだけ見ないと、成功した
		// バッチの古いコードを振込全体のエラーとして出し、送金済みなのに再送金を促してしまう。
		if src.RequestStatus == TRANSFER_REQUEST_FAIL && src.RequestErrorCode != "" {
			// #2: 件数でなく原因の出金IDを出す。締めのモーダルは1ページ10件で、原因が別ページに
			// あることがあるため。
			if causes := sortedOthers(causeWithdrawalIds, w.Id); len(causes) > 0 {
				return failureOf(WithdrawalFailureKindBatchCausedByOther, causedByOtherMessage(causes), src)
			}
			// #3
			return failureOf(WithdrawalFailureKindBatchError, "振込全体のエラー "+joinCodeAndMessage(src.RequestErrorCode, src.RequestErrorMsg)+withdrawalFailureSuffixBatchError, src)
		}
		// #4
		if src.RequestResultCode == TRANSFER_RESULT_FAIL {
			return failureOf(WithdrawalFailureKindBatchExpired, withdrawalFailureMessageBatchExpired, src)
		}
		// #5: 「記録なし」と出すと、そのまま再送金して二重に振り込むおそれがあるため、文言を分ける
		// (GMOの仕様: 更新系処理がタイムアウトした場合、処理自体は実行されている可能性がある)。
		if src.RequestStatus == TRANSFER_REQUEST_FAIL && src.RequestResultCode == "" {
			return failureOf(WithdrawalFailureKindResponseUnreadable, withdrawalFailureMessageResponseUnreadable, src)
		}
	}

	// #6: ジョブが途中で落ちて、出金依頼=FAIL・振込=WAITのまま残った場合など。
	if src.TransferStatus == TRANSFER_STATUS_WAIT || src.TransferStatus == TRANSFER_STATUS_SUCCESS {
		return failureOf(WithdrawalFailureKindStatusMismatch, withdrawalFailureMessageStatusMismatch, src)
	}

	// #7
	return failureOf(WithdrawalFailureKindNoRecord, withdrawalFailureMessageNoRecord, src)
}

func failureOf(kind WithdrawalFailureKind, message string, src *WithdrawalFailureSource) WithdrawalFailure {
	f := WithdrawalFailure{Kind: kind, Message: message}
	if src != nil {
		f.TransferRequestId = src.TransferRequestId
	}
	return f
}

// transferPredatesClose は、振込の作成日時が今回の締めより前かどうか。どちらかが未設定の
// ときは前とみなさない(締め日が無い=まだ締めていない、作成日時が無い=比べようがない)。
func transferPredatesClose(created, closeDate jsonutils.JSONTime) bool {
	c, d := time.Time(created), time.Time(closeDate)
	if c.IsZero() || d.IsZero() {
		return false
	}
	return c.Before(d)
}

func joinCodeAndMessage(code, msg string) string {
	return strings.TrimSpace(code + " " + msg)
}

// sortedOthers は、自分自身を除いた原因の出金IDを、重複を除いて昇順で返す。
func sortedOthers(ids []int64, self int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	others := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id == self {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		others = append(others, id)
	}
	sort.Slice(others, func(i, j int) bool { return others[i] < others[j] })
	return others
}

func causedByOtherMessage(causes []int64) string {
	if len(causes) == 1 {
		return fmt.Sprintf("同じ振込の出金ID %d のエラーで不成立", causes[0])
	}
	return fmt.Sprintf("同じ振込の出金ID %d(ほか %d 件)のエラーで不成立", causes[0], len(causes)-1)
}

// GetLatestFailureSourcesByWithdrawalIds は、出金依頼ごとの最新の振込(transfer.id最大)と
// その振込バッチを返す。キーはwithdrawal_id。既存のGetLatestByWithdrawalIdsは、COMPLETEの
// 口座情報の上書きで使っているため変えず、別に持つ。
func (Transfer) GetLatestFailureSourcesByWithdrawalIds(ctx context.Context, withdrawalIds []int64) (map[int64]WithdrawalFailureSource, error) {
	if len(withdrawalIds) == 0 {
		return map[int64]WithdrawalFailureSource{}, nil
	}

	items := make([]WithdrawalFailureSource, 0, len(withdrawalIds))
	err := factory.DB(ctx).Table("transfer").
		Select("withdrawal_transfer_mapping.withdrawal_id AS withdrawal_id, "+
			"transfer.id AS transfer_id, transfer.request_id AS transfer_request_id, "+
			"IFNULL(transfer.transfer_status, '') AS transfer_status_value, "+
			"IFNULL(transfer.error_code, '') AS transfer_error_code, "+
			"IFNULL(transfer.error_msg, '') AS transfer_error_msg, "+
			"transfer.created AS transfer_created, "+
			"IFNULL(transfer_request.id, 0) AS request_row_id, "+
			"IFNULL(transfer_request.request_status, '') AS request_status_value, "+
			"IFNULL(transfer_request.error_code, '') AS request_error_code, "+
			"IFNULL(transfer_request.error_msg, '') AS request_error_msg, "+
			"IFNULL(transfer_request.result_code, '') AS request_result_code").
		Join("INNER", "withdrawal_transfer_mapping", "withdrawal_transfer_mapping.transaction_id = transfer.transaction_id").
		Join("LEFT", "transfer_request", "transfer_request.id = transfer.request_id").
		In("withdrawal_transfer_mapping.withdrawal_id", withdrawalIds).
		Desc("transfer.id").
		Find(&items)
	if err != nil {
		return nil, err
	}

	latest := make(map[int64]WithdrawalFailureSource, len(withdrawalIds))
	for _, item := range items {
		// transfer.idの降順なので、withdrawal_idごとに最初に現れたものが最新。1つの出金依頼に
		// 複数の振込(再送金)が紐づきうる。
		if _, exists := latest[item.WithdrawalId]; !exists {
			latest[item.WithdrawalId] = item
		}
	}
	return latest, nil
}

type requestCauseRow struct {
	RequestId    int64 `xorm:"request_id <-"`
	WithdrawalId int64 `xorm:"withdrawal_id <-"`
}

// GetCauseWithdrawalIdsByRequestIds は、振込バッチごとに、error_codeを持つ明細の出金
// 依頼IDを返す(キーはtransfer.request_id)。transfer.request_idには索引が無く全件走査に
// なるが、いまの件数(約千行)では問題にならない。
func (Transfer) GetCauseWithdrawalIdsByRequestIds(ctx context.Context, requestIds []int64) (map[int64][]int64, error) {
	if len(requestIds) == 0 {
		return map[int64][]int64{}, nil
	}

	rows := make([]requestCauseRow, 0)
	err := factory.DB(ctx).Table("transfer").
		Select("DISTINCT transfer.request_id AS request_id, withdrawal_transfer_mapping.withdrawal_id AS withdrawal_id").
		Join("INNER", "withdrawal_transfer_mapping", "withdrawal_transfer_mapping.transaction_id = transfer.transaction_id").
		In("transfer.request_id", requestIds).
		Where("IFNULL(transfer.error_code, '') <> ''").
		Asc("transfer.request_id", "withdrawal_transfer_mapping.withdrawal_id").
		Find(&rows)
	if err != nil {
		return nil, err
	}

	causes := make(map[int64][]int64, len(requestIds))
	for _, row := range rows {
		causes[row.RequestId] = append(causes[row.RequestId], row.WithdrawalId)
	}
	return causes, nil
}

// WithdrawalFailureReasons は、ページに出す出金依頼のうち対象のものについて、失敗の理由を
// 出金依頼IDごとに返す。クエリは対象の数によらず最大2本(N+1にしない)。対象外の出金依頼は
// マップに含めない。取得に失敗したときはエラーを返すので、呼び出し側は一覧を落とさず、
// 理由を空にして返すこと。
func WithdrawalFailureReasons(ctx context.Context, withdrawals []Withdrawal, includeWaitWithFailedTransfer bool) (map[int64]WithdrawalFailure, error) {
	targetIds := make([]int64, 0, len(withdrawals))
	for _, w := range withdrawals {
		switch w.Status {
		case WITHDRAWAL_STATUS_FAIL, WITHDRAWAL_STATUS_CANCEL:
			targetIds = append(targetIds, w.Id)
		case WITHDRAWAL_STATUS_WAIT:
			if includeWaitWithFailedTransfer {
				targetIds = append(targetIds, w.Id)
			}
		}
	}
	reasons := make(map[int64]WithdrawalFailure, len(targetIds))
	if len(targetIds) == 0 {
		return reasons, nil
	}

	sources, err := Transfer{}.GetLatestFailureSourcesByWithdrawalIds(ctx, targetIds)
	if err != nil {
		return nil, err
	}

	requestIdSet := make(map[int64]struct{}, len(sources))
	for _, src := range sources {
		if src.hasRequest() {
			requestIdSet[src.TransferRequestId] = struct{}{}
		}
	}
	requestIds := make([]int64, 0, len(requestIdSet))
	for id := range requestIdSet {
		requestIds = append(requestIds, id)
	}
	causesByRequest, err := Transfer{}.GetCauseWithdrawalIdsByRequestIds(ctx, requestIds)
	if err != nil {
		return nil, err
	}

	for _, w := range withdrawals {
		var src *WithdrawalFailureSource
		var causes []int64
		if s, ok := sources[w.Id]; ok {
			src = &s
			causes = causesByRequest[s.TransferRequestId]
		}
		if failure := WithdrawalFailureReason(w, src, causes, includeWaitWithFailedTransfer); failure.Kind != WithdrawalFailureKindNone {
			reasons[w.Id] = failure
		}
	}
	return reasons, nil
}
