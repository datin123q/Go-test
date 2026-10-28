package domain

import (
	"testing"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"
)

func TestUser_ApproveSignUp(t *testing.T) {
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		signupStatus SIGNUP_STATUS
		auditStatus  AUDIT_STATUS
	}{
		{name: "最終承認待機・本人確認未完了からでもSIGNUP_COMPLETEになる", signupStatus: SIGNUP_KYC, auditStatus: WAIT_FINAL_APPROVAL_AUDIT},
		{name: "手動登録で会員登録途中のままでもSIGNUP_COMPLETEになる", signupStatus: SIGNUP_NONE, auditStatus: WAIT_AUDIT},
		{name: "既にSIGNUP_COMPLETEならそのまま", signupStatus: SIGNUP_COMPLETE, auditStatus: WAIT_FINAL_APPROVAL_AUDIT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{SignupStatus: tt.signupStatus, AuditStatus: tt.auditStatus}

			u.ApproveSignUp(now)

			assert.Equal(t, COMPLETED_AUDIT, u.AuditStatus)
			assert.Equal(t, SIGNUP_COMPLETE, u.SignupStatus)
			assert.Equal(t, jsonutils.JSONTime(now), u.SignUpApprovalDate)
		})
	}
}
