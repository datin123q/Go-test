package jobs

import (
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
)

// sumTransferFeesは、手数料見積もりAPIが失敗しても本振込は止めない(MIR-1123)
// ため0円のまま残り得る依頼の手数料合計を、結果確認ジョブで実際の手数料から
// 入れ直すために使う。
func TestSumTransferFees(t *testing.T) {
	tests := []struct {
		name      string
		transfers []domain.Transfer
		want      int64
	}{
		{
			name:      "0件なら0円",
			transfers: nil,
			want:      0,
		},
		{
			name:      "1件ならその手数料",
			transfers: []domain.Transfer{{TransferFee: 143}},
			want:      143,
		},
		{
			name: "複数件は合計する",
			transfers: []domain.Transfer{
				{TransferFee: 143},
				{TransferFee: 0},
				{TransferFee: 99},
			},
			want: 242,
		},
		{
			name: "全件0円なら合計も0円",
			transfers: []domain.Transfer{
				{TransferFee: 0},
				{TransferFee: 0},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sumTransferFees(tt.transfers))
		})
	}
}
