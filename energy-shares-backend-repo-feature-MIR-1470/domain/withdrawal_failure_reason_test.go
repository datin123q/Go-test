package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func jt(t time.Time) jsonutils.JSONTime { return jsonutils.JSONTime(t) }

// MIR-1170: 判定表(上から最初に当てはまったものを出す)の#0〜#7と「出さない」を1つずつ確かめる。
func TestWithdrawalFailureReason(t *testing.T) {
	closeAt := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	after := jt(closeAt.Add(time.Minute))
	before := jt(closeAt.Add(-time.Minute))

	failW := Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_FAIL, CloseDate: jt(closeAt)}
	// 出金依頼に紐づく最新の振込。個別に上書きして使う。
	src := func(mod func(s *WithdrawalFailureSource)) *WithdrawalFailureSource {
		s := &WithdrawalFailureSource{
			WithdrawalId:      10,
			TransferId:        100,
			TransferRequestId: 7,
			TransferStatus:    TRANSFER_STATUS_FAIL,
			TransferCreated:   after,
			RequestRowId:      7,
			RequestStatus:     TRANSFER_REQUEST_FAIL,
		}
		if mod != nil {
			mod(s)
		}
		return s
	}

	tests := []struct {
		name        string
		w           Withdrawal
		src         *WithdrawalFailureSource
		causes      []int64
		includeWait bool
		wantKind    WithdrawalFailureKind
		wantMessage string
		wantRequest int64
	}{
		// --- 出さない ---
		{name: "COMPLETEは出さない", w: Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_COMPLETE}, src: src(nil), wantKind: WithdrawalFailureKindNone},
		{name: "INCOMPLETEは出さない", w: Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_INCOMPLETE}, src: src(nil), wantKind: WithdrawalFailureKindNone},
		{name: "WAITは(ミラリタでは)出さない", w: Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_WAIT}, src: src(nil), wantKind: WithdrawalFailureKindNone},
		{
			name: "ZOLOME: WAITで最新の振込がFAILなら出す(見積もり失敗)", includeWait: true,
			w:        Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_WAIT, CloseDate: jt(closeAt)},
			src:      src(func(s *WithdrawalFailureSource) { s.RequestErrorCode, s.RequestErrorMsg = "E1", "見積もり失敗" }),
			wantKind: WithdrawalFailureKindBatchError, wantMessage: "振込全体のエラー E1 見積もり失敗。再送金の前に GMO で確認", wantRequest: 7,
		},
		{
			name: "ZOLOME: WAITでも振込がFAILでなければ出さない", includeWait: true,
			w:        Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_WAIT},
			src:      src(func(s *WithdrawalFailureSource) { s.TransferStatus = TRANSFER_STATUS_WAIT }),
			wantKind: WithdrawalFailureKindNone,
		},
		{name: "ZOLOME: WAITで振込が無ければ出さない", includeWait: true, w: Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_WAIT}, src: nil, wantKind: WithdrawalFailureKindNone},
		{name: "CANCELで振込が無い(送金前の取消)は出さない", w: Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_CANCEL}, src: nil, wantKind: WithdrawalFailureKindNone},

		// --- #0 ---
		{name: "#0 FAILで振込が無い", w: failW, src: nil, wantKind: WithdrawalFailureKindNoTransferData, wantMessage: "口座情報の不備で送金データを作成できず"},
		{
			name: "#0 振込の作成が締めより前(前の締めの古い理由を出さない)", w: failW,
			src: src(func(s *WithdrawalFailureSource) {
				s.TransferCreated = before
				s.TransferErrorCode, s.TransferErrorMsg = "220001", "古い理由"
			}),
			wantKind: WithdrawalFailureKindNoTransferData, wantMessage: "口座情報の不備で送金データを作成できず", wantRequest: 0,
		},
		{
			name: "#0 振込の作成が締めと同時刻なら前ではない(#1へ進む)", w: failW,
			src: src(func(s *WithdrawalFailureSource) {
				s.TransferCreated = jt(closeAt)
				s.TransferErrorCode, s.TransferErrorMsg = "220001", "今の理由"
			}),
			wantKind: WithdrawalFailureKindTransferError, wantMessage: "220001 今の理由", wantRequest: 7,
		},
		{
			name: "#0 締め日が未設定なら前とみなさない", w: Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_FAIL},
			src:      src(func(s *WithdrawalFailureSource) { s.TransferErrorCode = "220001" }),
			wantKind: WithdrawalFailureKindTransferError, wantMessage: "220001", wantRequest: 7,
		},
		{
			name: "#0 作成日時が未設定なら前とみなさない", w: failW,
			src: src(func(s *WithdrawalFailureSource) {
				s.TransferCreated = jsonutils.JSONTime{}
				s.TransferErrorCode = "220001"
			}),
			wantKind: WithdrawalFailureKindTransferError, wantMessage: "220001", wantRequest: 7,
		},
		{
			name: "#0 CANCELでも振込が締めより前なら出る", w: Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_CANCEL, CloseDate: jt(closeAt)},
			src:      src(func(s *WithdrawalFailureSource) { s.TransferCreated = before }),
			wantKind: WithdrawalFailureKindNoTransferData, wantMessage: "口座情報の不備で送金データを作成できず", wantRequest: 0,
		},

		// --- #1 ---
		{
			name: "#1 error_codeとerror_msgの両方", w: failW,
			src: src(func(s *WithdrawalFailureSource) {
				s.TransferErrorCode, s.TransferErrorMsg = "220001", "受取人名として使用できない文字が含まれています。"
			}),
			wantKind: WithdrawalFailureKindTransferError, wantMessage: "220001 受取人名として使用できない文字が含まれています。", wantRequest: 7,
		},
		{
			name: "#1 error_codeだけ", w: failW,
			src:      src(func(s *WithdrawalFailureSource) { s.TransferErrorCode = "220001" }),
			wantKind: WithdrawalFailureKindTransferError, wantMessage: "220001", wantRequest: 7,
		},
		{
			name: "#1 error_msgだけ(払戻し等の固定文)", w: Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_CANCEL, CloseDate: jt(closeAt)},
			src: src(func(s *WithdrawalFailureSource) {
				s.TransferStatus = TRANSFER_STATUS_CANCEL
				s.TransferErrorMsg = "送金キャンセル"
			}),
			wantKind: WithdrawalFailureKindTransferError, wantMessage: "送金キャンセル", wantRequest: 7,
		},
		{
			name: "#1は#2より先: 原因の明細自身は、バッチのエラーがあっても自分の理由を出す", w: failW, causes: []int64{10, 11},
			src: src(func(s *WithdrawalFailureSource) {
				s.TransferErrorCode, s.TransferErrorMsg = "220001", "原因"
				s.RequestErrorCode = "220011"
			}),
			wantKind: WithdrawalFailureKindTransferError, wantMessage: "220001 原因", wantRequest: 7,
		},

		// --- #2 / #3 ---
		{
			name: "#2 原因が1件", w: failW, causes: []int64{42},
			src:      src(func(s *WithdrawalFailureSource) { s.RequestErrorCode = "220011" }),
			wantKind: WithdrawalFailureKindBatchCausedByOther, wantMessage: "同じ振込の出金ID 42 のエラーで不成立", wantRequest: 7,
		},
		{
			name: "#2 原因が複数なら最小のIDと「ほかN件」を出す(重複と自分は数えない)", w: failW, causes: []int64{50, 42, 10, 50, 43},
			src:      src(func(s *WithdrawalFailureSource) { s.RequestErrorCode = "220011" }),
			wantKind: WithdrawalFailureKindBatchCausedByOther, wantMessage: "同じ振込の出金ID 42(ほか 2 件)のエラーで不成立", wantRequest: 7,
		},
		{
			name: "#3 原因が自分だけ・0件なら振込全体のエラー", w: failW, causes: []int64{10},
			src: src(func(s *WithdrawalFailureSource) {
				s.RequestErrorCode, s.RequestErrorMsg = "WG_ERR_110", "Validation Error"
			}),
			wantKind: WithdrawalFailureKindBatchError, wantMessage: "振込全体のエラー WG_ERR_110 Validation Error。再送金の前に GMO で確認", wantRequest: 7,
		},
		{
			name: "#3 error_msgが空でも末尾に空白を残さない", w: failW,
			src:      src(func(s *WithdrawalFailureSource) { s.RequestErrorCode = "WG_ERR_110" }),
			wantKind: WithdrawalFailureKindBatchError, wantMessage: "振込全体のエラー WG_ERR_110。再送金の前に GMO で確認", wantRequest: 7,
		},

		{
			name: "#2/#3にならない: MIR-1123より前の古い行(成功したバッチにerror_codeだけ残っている)", w: failW,
			src: src(func(s *WithdrawalFailureSource) {
				s.RequestStatus = TRANSFER_REQUEST_SUCCESS
				s.RequestErrorCode, s.RequestErrorMsg = "E_FEE", "見積もり失敗"
				s.TransferStatus = TRANSFER_STATUS_SUCCESS
			}),
			wantKind: WithdrawalFailureKindStatusMismatch, wantMessage: "状態と振込記録が一致しない(詳細を確認)", wantRequest: 7,
		},
		{
			name: "#4は古いerror_codeに隠されない: 成功扱いのバッチに残ったerror_codeより結果コード8を優先", w: failW,
			src: src(func(s *WithdrawalFailureSource) {
				s.RequestStatus = TRANSFER_REQUEST_SUCCESS
				s.RequestErrorCode = "E_FEE"
				s.RequestResultCode = TRANSFER_RESULT_FAIL
			}),
			wantKind: WithdrawalFailureKindBatchExpired, wantMessage: "振込の期限切れ(結果コード8)", wantRequest: 7,
		},

		// --- #4 / #5 ---
		{
			name: "#4 結果コード8は期限切れ", w: failW,
			src:      src(func(s *WithdrawalFailureSource) { s.RequestResultCode = TRANSFER_RESULT_FAIL }),
			wantKind: WithdrawalFailureKindBatchExpired, wantMessage: "振込の期限切れ(結果コード8)", wantRequest: 7,
		},
		{
			name: "#5 応答が読めない失敗は、再送金の前にGMOで確認するよう案内する", w: failW,
			src:      src(nil),
			wantKind: WithdrawalFailureKindResponseUnreadable, wantMessage: "銀行の応答を読めず失敗扱い。銀行側で受け付けられている可能性あり。再送金の前に GMO で確認", wantRequest: 7,
		},
		{
			name: "#5にならない: 振込バッチの行が無い(LEFT JOINで取れない)", w: failW,
			src:      src(func(s *WithdrawalFailureSource) { s.RequestRowId = 0 }),
			wantKind: WithdrawalFailureKindNoRecord, wantMessage: "理由の記録なし(詳細を確認)", wantRequest: 7,
		},
		{
			name: "#5にならない: 振込バッチがFAILでなく結果コードがある", w: failW,
			src: src(func(s *WithdrawalFailureSource) {
				s.RequestStatus = TRANSFER_REQUEST_WAIT
				s.RequestResultCode = TRANSFER_RESULT_WAIT
			}),
			wantKind: WithdrawalFailureKindNoRecord, wantMessage: "理由の記録なし(詳細を確認)", wantRequest: 7,
		},

		// --- #6 / #7 ---
		{
			name: "#6 出金依頼=FAIL・振込=WAIT(ジョブが途中で落ちた)", w: failW,
			src: src(func(s *WithdrawalFailureSource) {
				s.TransferStatus = TRANSFER_STATUS_WAIT
				s.RequestStatus = TRANSFER_REQUEST_WAIT
			}),
			wantKind: WithdrawalFailureKindStatusMismatch, wantMessage: "状態と振込記録が一致しない(詳細を確認)", wantRequest: 7,
		},
		{
			name: "#6 出金依頼=FAIL・振込=SUCCESS", w: failW,
			src: src(func(s *WithdrawalFailureSource) {
				s.TransferStatus = TRANSFER_STATUS_SUCCESS
				s.RequestStatus = TRANSFER_REQUEST_SUCCESS
			}),
			wantKind: WithdrawalFailureKindStatusMismatch, wantMessage: "状態と振込記録が一致しない(詳細を確認)", wantRequest: 7,
		},
		{
			name: "#7 どれにも当てはまらない", w: Withdrawal{Id: 10, Status: WITHDRAWAL_STATUS_CANCEL, CloseDate: jt(closeAt)},
			src:      src(func(s *WithdrawalFailureSource) { s.TransferStatus = TRANSFER_STATUS_CANCEL; s.RequestRowId = 0 }),
			wantKind: WithdrawalFailureKindNoRecord, wantMessage: "理由の記録なし(詳細を確認)", wantRequest: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithdrawalFailureReason(tt.w, tt.src, tt.causes, tt.includeWait)

			assert.Equal(t, tt.wantKind, got.Kind)
			assert.Equal(t, tt.wantMessage, got.Message)
			assert.Equal(t, tt.wantRequest, got.TransferRequestId)
		})
	}
}

// setupFailureReasonTestDB は、既存のsetupTransferTestDB(transfer・withdrawal_transfer_mapping)に
// transfer_requestを足す。本番スキーマ(db/sql/version/energyshares.base.schema.sql)と揃える。
func setupFailureReasonTestDB(t *testing.T) context.Context {
	t.Helper()
	ctx := setupTransferTestDB(t)
	session := ctx.Value(factory.ConextDBType).(*xorm.Session)
	_, err := session.Exec(`
		CREATE TABLE transfer_request (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			transfer_date DATETIME NULL,
			request_total_amount INTEGER NOT NULL,
			request_total_cnt INTEGER NOT NULL,
			transfer_fee INTEGER NULL,
			error_code TEXT NULL,
			error_msg TEXT NULL,
			request_status TEXT NULL,
			result_code TEXT NULL,
			apply_no TEXT NULL,
			apply_end_datetime TEXT NULL,
			created DATETIME NULL,
			updated DATETIME NULL,
			completed_transfer TEXT NULL
		);
	`)
	require.NoError(t, err)
	return ctx
}

func insertTransferRequest(t *testing.T, ctx context.Context, r TransferRequest) int64 {
	t.Helper()
	session := ctx.Value(factory.ConextDBType).(*xorm.Session)
	_, err := session.Insert(&r)
	require.NoError(t, err)
	return r.Id
}

// addFailedTransfer は、出金依頼withdrawalIdの振込を1行(と、そのマッピング)作る。
func addFailedTransfer(t *testing.T, ctx context.Context, withdrawalId, requestId int64, txId string, status TRANSFER_STATUS_TYPE, errCode, errMsg string) {
	t.Helper()
	insertTransfer(t, ctx, Transfer{
		ItemId: 1, RequestId: requestId, TransactionId: txId,
		TransferStatus: status, ErrorCode: errCode, ErrorMsg: errMsg,
		// 判定に使わない列。読み出されないことの確認のため、値を入れておく。
		AccountNumber: "9999999", Depositor: "ヤマダ タロウ", BankName: "テスト銀行",
	})
	insertMapping(t, ctx, WithdrawalTransferMapping{WithdrawalId: withdrawalId, TransactionId: txId})
}

// MIR-1170: 2026-09-03の220011(79件中、原因は1件)の再現。transferとtransfer_requestは
// error_code・error_msg・idの列名が同じため、別名を付けずに読むと、全明細のerror_codeが
// バッチの220011に上書きされ、原因の1件(220001)が埋もれる。
func TestWithdrawalFailureReasons_CauseAndBystandersAreDistinguished(t *testing.T) {
	ctx := setupFailureReasonTestDB(t)

	requestId := insertTransferRequest(t, ctx, TransferRequest{
		RequestTotalAmount: 1, RequestTotalCnt: 3, RequestStatus: TRANSFER_REQUEST_FAIL,
		ErrorCode: "220011", ErrorMsg: "振込データに誤りがあります",
	})
	addFailedTransfer(t, ctx, 101, requestId, "tx-101", TRANSFER_STATUS_FAIL, "220001", "受取人名として使用できない文字が含まれています。")
	addFailedTransfer(t, ctx, 102, requestId, "tx-102", TRANSFER_STATUS_FAIL, "", "")
	addFailedTransfer(t, ctx, 103, requestId, "tx-103", TRANSFER_STATUS_FAIL, "", "")

	withdrawals := []Withdrawal{
		{Id: 101, Status: WITHDRAWAL_STATUS_FAIL},
		{Id: 102, Status: WITHDRAWAL_STATUS_FAIL},
		{Id: 103, Status: WITHDRAWAL_STATUS_FAIL},
	}

	// 取得した行そのもので、列名の衝突が起きていないことを確かめる。
	sources, err := Transfer{}.GetLatestFailureSourcesByWithdrawalIds(ctx, []int64{101, 102, 103})
	require.NoError(t, err)
	assert.Equal(t, "220001", sources[101].TransferErrorCode, "原因の明細自身のerror_code")
	assert.Equal(t, "", sources[102].TransferErrorCode, "巻き込まれた明細のerror_codeが、バッチのものに上書きされていない")
	assert.Equal(t, "220011", sources[102].RequestErrorCode)
	assert.Equal(t, requestId, sources[102].TransferRequestId)

	reasons, err := WithdrawalFailureReasons(ctx, withdrawals, false)
	require.NoError(t, err)

	assert.Equal(t, WithdrawalFailureKindTransferError, reasons[101].Kind)
	assert.Equal(t, "220001 受取人名として使用できない文字が含まれています。", reasons[101].Message)
	for _, id := range []int64{102, 103} {
		assert.Equal(t, WithdrawalFailureKindBatchCausedByOther, reasons[id].Kind)
		assert.Equal(t, "同じ振込の出金ID 101 のエラーで不成立", reasons[id].Message, "別の出金依頼の理由を取り違えていない")
		assert.Equal(t, requestId, reasons[id].TransferRequestId)
	}
}

// 再送金: 出金依頼ごとに最新の振込(transfer.id最大)だけを見る。
func TestWithdrawalFailureReasons_UsesLatestTransferOnly(t *testing.T) {
	ctx := setupFailureReasonTestDB(t)

	failedBatch := func(code string) int64 {
		return insertTransferRequest(t, ctx, TransferRequest{
			RequestTotalAmount: 1, RequestTotalCnt: 1, RequestStatus: TRANSFER_REQUEST_FAIL, ErrorCode: code,
		})
	}
	// 出金依頼201: 古い振込がFAIL、新しい振込がSUCCESS(再送金で成功)。状態はCOMPLETEなので出さない。
	oldBatch, newBatch := failedBatch("WG_ERR_110"), failedBatch("")
	addFailedTransfer(t, ctx, 201, oldBatch, "tx-201-old", TRANSFER_STATUS_FAIL, "", "")
	addFailedTransfer(t, ctx, 201, newBatch, "tx-201-new", TRANSFER_STATUS_SUCCESS, "", "")
	// 出金依頼202: 新旧どちらもFAIL。新しい方の理由を出す。
	addFailedTransfer(t, ctx, 202, oldBatch, "tx-202-old", TRANSFER_STATUS_FAIL, "220001", "古い理由")
	addFailedTransfer(t, ctx, 202, newBatch, "tx-202-new", TRANSFER_STATUS_FAIL, "220002", "新しい理由")

	reasons, err := WithdrawalFailureReasons(ctx, []Withdrawal{
		{Id: 201, Status: WITHDRAWAL_STATUS_COMPLETE},
		{Id: 202, Status: WITHDRAWAL_STATUS_FAIL},
	}, false)
	require.NoError(t, err)

	_, has201 := reasons[201]
	assert.False(t, has201, "再送金で成功した出金依頼に、古い失敗の理由を出していない")
	assert.Equal(t, "220002 新しい理由", reasons[202].Message)
	assert.Equal(t, newBatch, reasons[202].TransferRequestId)
}

// 振込バッチの行が無くても(LEFT JOIN)、振込自身のエラーは出る。振込が無い出金依頼もある。
func TestWithdrawalFailureReasons_MissingRowsDoNotDropTheList(t *testing.T) {
	ctx := setupFailureReasonTestDB(t)

	addFailedTransfer(t, ctx, 301, 9999, "tx-301", TRANSFER_STATUS_FAIL, "220001", "理由")

	reasons, err := WithdrawalFailureReasons(ctx, []Withdrawal{
		{Id: 301, Status: WITHDRAWAL_STATUS_FAIL},
		{Id: 302, Status: WITHDRAWAL_STATUS_FAIL},   // 振込が無い
		{Id: 303, Status: WITHDRAWAL_STATUS_CANCEL}, // 振込が無い(送金前の取消)
		{Id: 304, Status: WITHDRAWAL_STATUS_COMPLETE},
	}, false)
	require.NoError(t, err)

	assert.Equal(t, "220001 理由", reasons[301].Message)
	assert.Equal(t, WithdrawalFailureKindNoTransferData, reasons[302].Kind)
	_, has303 := reasons[303]
	assert.False(t, has303)
	_, has304 := reasons[304]
	assert.False(t, has304)
}

// ZOLOME: 手数料の見積もりに失敗してWAITのまま止まった出金依頼は、includeWaitのときだけ出す。
func TestWithdrawalFailureReasons_IncludeWaitWithFailedTransfer(t *testing.T) {
	ctx := setupFailureReasonTestDB(t)

	requestId := insertTransferRequest(t, ctx, TransferRequest{
		RequestTotalAmount: 1, RequestTotalCnt: 1, RequestStatus: TRANSFER_REQUEST_FAIL, ErrorCode: "E_FEE", ErrorMsg: "見積もり失敗",
	})
	addFailedTransfer(t, ctx, 401, requestId, "tx-401", TRANSFER_STATUS_FAIL, "", "")
	withdrawals := []Withdrawal{{Id: 401, Status: WITHDRAWAL_STATUS_WAIT}}

	without, err := WithdrawalFailureReasons(ctx, withdrawals, false)
	require.NoError(t, err)
	assert.Empty(t, without)

	with, err := WithdrawalFailureReasons(ctx, withdrawals, true)
	require.NoError(t, err)
	assert.Equal(t, "振込全体のエラー E_FEE 見積もり失敗。再送金の前に GMO で確認", with[401].Message)
}

func TestWithdrawalFailureReasons_NoTargetsRunsNoQuery(t *testing.T) {
	// DBを用意しない: 対象が無ければクエリを発行しないこと(ctxにDBが無くてもpanicしない)。
	reasons, err := WithdrawalFailureReasons(context.Background(), []Withdrawal{
		{Id: 1, Status: WITHDRAWAL_STATUS_COMPLETE},
		{Id: 2, Status: WITHDRAWAL_STATUS_WAIT},
	}, false)

	require.NoError(t, err)
	assert.Empty(t, reasons)
}

// 振込の作成日時(transfer.created)を実際に読めていることを確かめる。この列を読み落とすと、再締めでも
// 口座不備で振込を作れなかった出金依頼に、前の締めの古い理由が出続ける。
func TestWithdrawalFailureReasons_TransferBeforeCloseIsNoTransferData(t *testing.T) {
	ctx := setupFailureReasonTestDB(t)

	requestId := insertTransferRequest(t, ctx, TransferRequest{RequestTotalAmount: 1, RequestTotalCnt: 1, RequestStatus: TRANSFER_REQUEST_FAIL})
	addFailedTransfer(t, ctx, 501, requestId, "tx-501", TRANSFER_STATUS_FAIL, "220001", "前の締めの理由")

	now := time.Now()
	reasons, err := WithdrawalFailureReasons(ctx, []Withdrawal{
		{Id: 501, Status: WITHDRAWAL_STATUS_FAIL, CloseDate: jt(now.Add(time.Hour))},
	}, false)
	require.NoError(t, err)
	assert.Equal(t, WithdrawalFailureKindNoTransferData, reasons[501].Kind, "振込より後の締めなら、前の振込の理由は出さない")
	assert.Equal(t, int64(0), reasons[501].TransferRequestId)

	reasons, err = WithdrawalFailureReasons(ctx, []Withdrawal{
		{Id: 501, Status: WITHDRAWAL_STATUS_FAIL, CloseDate: jt(now.Add(-time.Hour))},
	}, false)
	require.NoError(t, err)
	assert.Equal(t, "220001 前の締めの理由", reasons[501].Message, "振込より前の締めなら、その振込の理由を出す")
}
