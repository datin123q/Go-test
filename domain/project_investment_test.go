package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineInvestmentExecutionState(t *testing.T) {
	tests := []struct {
		name     string
		executes []ProjectInvestmentExecute
		want     InvestmentExecutionState
	}{
		{
			name:     "1件も無ければ記録なし",
			executes: nil,
			want:     InvestmentExecutionNone,
		},
		{
			name:     "空スライスでも記録なし",
			executes: []ProjectInvestmentExecute{},
			want:     InvestmentExecutionNone,
		},
		{
			name: "1件でCOMPLETEDなら完了",
			executes: []ProjectInvestmentExecute{
				{Id: 1, ExecuteStatus: COMPLETED},
			},
			want: InvestmentExecutionCompleted,
		},
		{
			name: "全行COMPLETEDなら完了",
			executes: []ProjectInvestmentExecute{
				{Id: 1, ExecuteStatus: COMPLETED},
				{Id: 2, ExecuteStatus: COMPLETED},
			},
			want: InvestmentExecutionCompleted,
		},
		{
			name: "1件でもSCHEDULEDがあれば途中",
			executes: []ProjectInvestmentExecute{
				{Id: 1, ExecuteStatus: COMPLETED},
				{Id: 2, ExecuteStatus: SCHEDULED},
			},
			want: InvestmentExecutionInProgress,
		},
		{
			name: "1件でもCONFIRMがあれば途中",
			executes: []ProjectInvestmentExecute{
				{Id: 1, ExecuteStatus: CONFIRM},
			},
			want: InvestmentExecutionInProgress,
		},
		{
			name: "空文字は安全側に倒して途中扱いする",
			executes: []ProjectInvestmentExecute{
				{Id: 1, ExecuteStatus: ""},
			},
			want: InvestmentExecutionInProgress,
		},
		{
			name: "未知の値も安全側に倒して途中扱いする",
			executes: []ProjectInvestmentExecute{
				{Id: 1, ExecuteStatus: "UNKNOWN"},
			},
			want: InvestmentExecutionInProgress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, DetermineInvestmentExecutionState(tt.executes))
		})
	}
}

// CanManuallyChangeInvestStatus は遷移元×遷移先×出資実行の状態×自己募集
// かどうかの全組み合わせをカバーする(MIR-1295の設計要求)。許可するのは
// 申込中→取消、成立→取消(出資実行が途中でない場合)の2つだけ。
func TestCanManuallyChangeInvestStatus(t *testing.T) {
	allStatuses := []INVEST_STATUS{
		REQUEST_INVEST, FAIL_INVEST, CANCEL_INVEST,
		SUCCESS_INVEST, OPERATION_INVEST, REFUND_INVEST,
	}
	allExecutionStates := []InvestmentExecutionState{
		InvestmentExecutionNone, InvestmentExecutionInProgress, InvestmentExecutionCompleted,
	}

	isAllowedInTable := func(from, to INVEST_STATUS, execution InvestmentExecutionState) bool {
		switch {
		case from == REQUEST_INVEST && to == CANCEL_INVEST:
			return true
		case from == SUCCESS_INVEST && to == CANCEL_INVEST:
			return execution != InvestmentExecutionInProgress
		default:
			return false
		}
	}

	t.Run("全組み合わせ(自己募集ではない場合)", func(t *testing.T) {
		for _, from := range allStatuses {
			for _, to := range allStatuses {
				for _, execution := range allExecutionStates {
					wantAllowed := isAllowedInTable(from, to, execution)
					err := CanManuallyChangeInvestStatus(from, to, execution, false)
					if wantAllowed {
						assert.NoErrorf(t, err, "from=%s to=%s execution=%s", from, to, execution)
					} else {
						assert.Errorf(t, err, "from=%s to=%s execution=%s", from, to, execution)
					}
				}
			}
		}
	})

	t.Run("自己募集ファンドはどの組み合わせでも拒否する", func(t *testing.T) {
		for _, from := range allStatuses {
			for _, to := range allStatuses {
				for _, execution := range allExecutionStates {
					err := CanManuallyChangeInvestStatus(from, to, execution, true)
					assert.Errorf(t, err, "from=%s to=%s execution=%s (self fund)", from, to, execution)
				}
			}
		}
	})

	t.Run("拒否したときのメッセージに表示名を含める", func(t *testing.T) {
		err := CanManuallyChangeInvestStatus(OPERATION_INVEST, FAIL_INVEST, InvestmentExecutionNone, false)
		assert.EqualError(t, err, "この出資は運用中のため、投資不成立には変更できません")
	})
}
