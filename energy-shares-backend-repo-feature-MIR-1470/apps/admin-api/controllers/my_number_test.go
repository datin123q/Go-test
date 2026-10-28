package controllers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// MIR-1276③: マイナンバー承認で本登録完了(COMPLETED_AUDIT)にしてよいかの判定。
// isEkyc単独ではなく、審査ステータスが最終承認待機(WAIT_FINAL_APPROVAL_AUDIT)の
// ときだけtrueになることを、全審査ステータス×isEkycで確認する。
func TestCanCompleteAuditOnMyNumberApprove(t *testing.T) {
	allStatuses := []domain.AUDIT_STATUS{
		domain.PROVISIONAL_AUDIT,
		domain.WAIT_AUDIT,
		domain.WAIT_IDENTITY_REGISTER_AUDIT,
		domain.HOLD_AUDIT,
		domain.WAIT_ANTISOCIAL_AUDIT,
		domain.CHECK_ANTISOCIAL_AUDIT,
		domain.WAIT_FINAL_APPROVAL_AUDIT,
		domain.COMPLETED_AUDIT,
		domain.INVALIDITY_AUDIT,
		domain.CHECK_ADDRESS_AUDIT,
	}

	for _, status := range allStatuses {
		status := status
		t.Run(string(status)+"_isEkyc=false", func(t *testing.T) {
			require.False(t, canCompleteAuditOnMyNumberApprove(status, false), "isEkycがfalseならどの審査ステータスでもfalse")
		})
	}

	for _, status := range allStatuses {
		status := status
		want := status == domain.WAIT_FINAL_APPROVAL_AUDIT
		t.Run(string(status)+"_isEkyc=true", func(t *testing.T) {
			require.Equal(t, want, canCompleteAuditOnMyNumberApprove(status, true),
				"isEkyc=trueでも、審査ステータスが最終承認待機のときだけtrueになること")
		})
	}
}
