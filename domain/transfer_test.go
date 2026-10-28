package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/feequote"
)

// TransferDesignatedDateは、GMOあおぞら銀行への振込指定日(YYYY-MM-DD)を
// 「JSTの翌日」として返す(MIR-1172)。境界値(JSTの23:59・0:00、月末・年末
// をまたぐ日)で、UTCの日付とJSTの日付がずれないことを確認する。
func TestTransferDesignatedDate(t *testing.T) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)

	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{
			name: "JSTの23:59 → 翌日の日付",
			now:  time.Date(2026, 10, 5, 23, 59, 0, 0, jst),
			want: "2026-10-06",
		},
		{
			name: "JSTの0:00 → 翌日の日付",
			now:  time.Date(2026, 10, 5, 0, 0, 0, 0, jst),
			want: "2026-10-06",
		},
		{
			name: "月末をまたぐ(JSTの10/31 23:00 → 11/1)",
			now:  time.Date(2026, 10, 31, 23, 0, 0, 0, jst),
			want: "2026-11-01",
		},
		{
			name: "年末をまたぐ(JSTの12/31 23:00 → 翌年1/1)",
			now:  time.Date(2026, 12, 31, 23, 0, 0, 0, jst),
			want: "2027-01-01",
		},
		{
			name: "UTCで渡されてもJSTへ変換してから翌日にする(UTCの10/5 15:30 = JSTの10/6 0:30 → 10/7)",
			now:  time.Date(2026, 10, 5, 15, 30, 0, 0, time.UTC),
			want: "2026-10-07",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, TransferDesignatedDate(tt.now))
		})
	}
}

// parseFeeResultは、resultがstring/float64のときだけ手数料として受け入れ、
// それ以外(null・オブジェクト・配列・bool等)はエラーにする(MIR-1172、
// ①敵対的検証の指摘)。resultがnullの応答(GMOトークン取得失敗時の共通
// エラーハンドラ等)を手数料0円として静かに受け入れてしまう退行がないか
// を固定する。
func TestParseFeeResult(t *testing.T) {
	tests := []struct {
		name    string
		result  any
		want    int64
		wantErr bool
	}{
		{name: "文字列の数値", result: "143", want: 143},
		{name: "float64の数値", result: float64(143), want: 143},
		{name: "0円(文字列)", result: "0", want: 0},
		{name: "0円(float64)", result: float64(0), want: 0},
		{name: "nil(result無し・エラー応答)はエラー", result: nil, wantErr: true},
		{name: "オブジェクト(map)はエラー", result: map[string]any{"fee": 143.0}, wantErr: true},
		{name: "真偽値はエラー", result: true, wantErr: true},
		{name: "数値に変換できない文字列はエラー", result: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFeeResult(tt.result)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRequestFeeBulk_UnwrapResult(t *testing.T) {
	body := []byte(`{
        "result": {
            "totalFee": 429,
            "details": [
                {"withdrawalId": 1, "transferFee": 143},
                {"withdrawalId": 2, "transferFee": 143},
                {"withdrawalId": 3, "transferFee": 0},
                {"withdrawalId": 4, "transferFee": 143}
            ]
        }
    }`)

	var wrapper struct {
		Result feequote.WithdrawalFeeBulkResponse `json:"result"`
	}

	err := json.Unmarshal(body, &wrapper)
	require.NoError(t, err)
	require.Equal(t, int64(429), wrapper.Result.TotalFee)
	require.Len(t, wrapper.Result.Details, 4)

	require.Equal(t, int64(1), wrapper.Result.Details[0].WithdrawalID)
	require.Equal(t, int64(143), wrapper.Result.Details[0].TransferFee)
	require.Equal(t, int64(3), wrapper.Result.Details[2].WithdrawalID)
	require.Equal(t, int64(0), wrapper.Result.Details[2].TransferFee)
	require.Equal(t, int64(4), wrapper.Result.Details[3].WithdrawalID)
	require.Equal(t, int64(143), wrapper.Result.Details[3].TransferFee)
}

func TestRequestFeeBulk_TotalFeeZeroAccepted(t *testing.T) {
	body := []byte(`{
        "result": {
            "totalFee": 0,
            "details": [
                {"withdrawalId": 10, "transferFee": 0},
                {"withdrawalId": 11, "transferFee": 0}
            ]
        }
    }`)

	var wrapper struct {
		Result feequote.WithdrawalFeeBulkResponse `json:"result"`
	}

	err := json.Unmarshal(body, &wrapper)
	require.NoError(t, err)

	require.Equal(t, int64(0), wrapper.Result.TotalFee)
	require.Len(t, wrapper.Result.Details, 2)
}
