package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShare_CanCancelInvest(t *testing.T) {
	tests := []struct {
		name         string
		investStatus INVEST_STATUS
		want         bool
	}{
		{name: "投資申請中はキャンセル可能", investStatus: REQUEST_INVEST, want: true},
		{name: "投資完了はキャンセル可能", investStatus: SUCCESS_INVEST, want: true},
		{name: "投資不成立はキャンセル不可", investStatus: FAIL_INVEST, want: false},
		{name: "キャンセル済みはキャンセル不可", investStatus: CANCEL_INVEST, want: false},
		{name: "運用中はキャンセル不可", investStatus: OPERATION_INVEST, want: false},
		{name: "運用終了はキャンセル不可", investStatus: REFUND_INVEST, want: false},
		{name: "未設定はキャンセル不可", investStatus: INVEST_STATUS(""), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			share := Share{InvestStatus: tt.investStatus}

			assert.Equal(t, tt.want, share.CanCancelInvest())
		})
	}
}
