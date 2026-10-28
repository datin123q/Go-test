package jobs

import (
	"net/http"
	"testing"

	client "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// transferTransactionIds は、GMO振込ログに氏名・口座番号が平文で残っていた問題
// (MIR-1116)の対応で、一括処理のログに Transfer を丸ごと載せる代わりに使う。
// 拾うのが transaction_id だけで、氏名・口座番号を含まないことを固定する。
func TestTransferTransactionIds(t *testing.T) {
	tests := []struct {
		name      string
		transfers []domain.Transfer
		want      []string
	}{
		{
			name:      "0件なら空スライスを返す(nilではない)",
			transfers: nil,
			want:      []string{},
		},
		{
			name: "1件ならその1件のtransaction_idを返す",
			transfers: []domain.Transfer{
				{TransactionId: "id-1", Depositor: "ヤマダ タロウ", AccountNumber: "1234567"},
			},
			want: []string{"id-1"},
		},
		{
			name: "複数件なら全件分のtransaction_idを返す",
			transfers: []domain.Transfer{
				{TransactionId: "id-1", Depositor: "ヤマダ タロウ", AccountNumber: "1234567"},
				{TransactionId: "id-2", Depositor: "スズキ ハナコ", AccountNumber: "7654321"},
			},
			want: []string{"id-1", "id-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, transferTransactionIds(tt.transfers))
		})
	}
}

// httpStatusCode は、*http.Response をそのまま logrus.Fields に渡すと
// Request.GetBody(関数値)を含むため標準JSONエンコーダで直列化に失敗し、
// ログ行そのものが出力されなくなる問題への対応(MIR-1116)。
// ステータスコードだけを取り出せば直列化に失敗しない。
func TestHttpStatusCode(t *testing.T) {
	tests := []struct {
		name string
		resp *http.Response
		want int
	}{
		{
			name: "nilのときは0を返す",
			resp: nil,
			want: 0,
		},
		{
			name: "2xxはそのままステータスコードを返す",
			resp: &http.Response{StatusCode: 200},
			want: 200,
		},
		{
			name: "4xx/5xxもそのままステータスコードを返す",
			resp: &http.Response{StatusCode: 429},
			want: 429,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				assert.Equal(t, tt.want, httpStatusCode(tt.resp))
			})
		})
	}
}

// isTransferErrorDetailIndexValid は、GMOのTransferErrorDetail.ItemIdをAtoiした値を
// そのままTransfers[idx-1]へ使うと、想定外の値(0以下・依頼件数超過)でpanicしていた
// 問題への対応(MIR-1127)。1始まりの範囲に収まっているかだけを判定する。
func TestIsTransferErrorDetailIndexValid(t *testing.T) {
	tests := []struct {
		name          string
		idx           int
		transferCount int
		want          bool
	}{
		{name: "0以下は範囲外", idx: 0, transferCount: 3, want: false},
		{name: "負の値も範囲外", idx: -1, transferCount: 3, want: false},
		{name: "下限の1は範囲内", idx: 1, transferCount: 3, want: true},
		{name: "上限のtransferCountは範囲内", idx: 3, transferCount: 3, want: true},
		{name: "transferCountを超えると範囲外", idx: 4, transferCount: 3, want: false},
		{name: "Transfersが空なら1でも範囲外", idx: 1, transferCount: 0, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isTransferErrorDetailIndexValid(tt.idx, tt.transferCount))
		})
	}
}

// validateBulkTransferStatusResponse は、bulkTransferInfosTask 内で
// item.BulkTransferDetails[0].BulktransferResponses[0].BulkTransferInfos を無条件に
// 参照していたことで、GMOが明細を返さない、または明細件数が依頼件数と一致しないと
// req.Transfers[idx] が範囲外になり panic していた問題への対応(MIR-1127)。
func TestValidateBulkTransferStatusResponse(t *testing.T) {
	statusResponse := func(infoCount int) *client.BulkTransferStatusResponse {
		return &client.BulkTransferStatusResponse{
			BulkTransferDetails: []client.BulkTransferDetail{{
				BulktransferResponses: []client.BulkTransferResponse{{
					BulkTransferInfos: make([]client.BulkTransferInfo, infoCount),
				}},
			}},
		}
	}
	transferRequest := func(transferCount int) *domain.TransferRequest {
		return &domain.TransferRequest{Transfers: make([]domain.Transfer, transferCount)}
	}

	tests := []struct {
		name        string
		item        *client.BulkTransferStatusResponse
		req         *domain.TransferRequest
		wantErr     string
		wantSuccess bool
	}{
		{
			name:    "レスポンスがnil",
			req:     transferRequest(1),
			wantErr: "bulk transfer status response is nil",
		},
		{
			name:    "振込依頼がnil",
			item:    statusResponse(1),
			wantErr: "transfer request is nil",
		},
		{
			name:    "総合振込照会明細が空",
			item:    &client.BulkTransferStatusResponse{},
			req:     transferRequest(1),
			wantErr: "bulk transfer details are empty",
		},
		{
			name:    "DB側の送金明細が空",
			item:    statusResponse(0),
			req:     transferRequest(0),
			wantErr: "transfer request details are empty",
		},
		{
			name: "振込一括照会レスポンスが空",
			item: &client.BulkTransferStatusResponse{
				BulkTransferDetails: []client.BulkTransferDetail{{}},
			},
			req:     transferRequest(1),
			wantErr: "bulk transfer responses are empty",
		},
		{
			name:    "GMO明細数とDB明細数が一致しない",
			item:    statusResponse(1),
			req:     transferRequest(2),
			wantErr: "bulk transfer info count mismatch: response=1 request=2",
		},
		{
			name:        "件数が一致すれば成功する",
			item:        statusResponse(2),
			req:         transferRequest(2),
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var detail *client.BulkTransferDetail
			var response *client.BulkTransferResponse
			var err error
			assert.NotPanics(t, func() {
				detail, response, err = validateBulkTransferStatusResponse(tt.item, tt.req)
			})

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, detail)
				assert.NotNil(t, response)
				return
			}
			assert.Nil(t, detail)
			assert.Nil(t, response)
			assert.ErrorContains(t, err, tt.wantErr)
			assert.ErrorIs(t, err, errBulkTransferStatusResponseInvalid)
		})
	}
}

// bulkTransferInfosTask のエラーが errBulkTransferStatusResponseInvalid かどうかで、呼び出し側
// (bulkTransferStatusCheckJob)がこの1件をスキップするかジョブを失敗させるかを分けている(MIR-1127)。
// validateBulkTransferStatusResponse 由来のエラーは errors.Wrap を経由しても errors.Is で
// 判別できることを確認する。DB書き込み失敗等の内部エラーとの取り違えは、前者を誤ってジョブ失敗に
// 昇格させるだけで安全側に倒れるため、ここでは正しい種類を確実に判別できることだけを確認する。
func TestBulkTransferInfosTask_ErrorIsDistinguishableFromInternalError(t *testing.T) {
	err := bulkTransferInfosTask(nil, nil, &domain.TransferRequest{Transfers: []domain.Transfer{{}}})
	require := assert.New(t)
	require.Error(err)
	require.ErrorIs(err, errBulkTransferStatusResponseInvalid)

	internalErr := errors.New("db write failed")
	require.False(errors.Is(internalErr, errBulkTransferStatusResponseInvalid))
}
