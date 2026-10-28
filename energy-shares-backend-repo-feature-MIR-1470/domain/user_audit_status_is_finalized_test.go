package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuditStatus_IsFinalized(t *testing.T) {
	tests := []struct {
		name   string
		status AUDIT_STATUS
		want   bool
	}{
		{name: "本登録承認完了はtrue", status: COMPLETED_AUDIT, want: true},
		{name: "無効はtrue", status: INVALIDITY_AUDIT, want: true},
		{name: "仮登録はfalse", status: PROVISIONAL_AUDIT, want: false},
		{name: "本登録承認待機はfalse", status: WAIT_AUDIT, want: false},
		{name: "本人確認待ちはfalse", status: WAIT_IDENTITY_REGISTER_AUDIT, want: false},
		{name: "本登録保留はfalse", status: HOLD_AUDIT, want: false},
		{name: "反社待機中はfalse", status: WAIT_ANTISOCIAL_AUDIT, want: false},
		{name: "反社確認中はfalse", status: CHECK_ANTISOCIAL_AUDIT, want: false},
		{name: "最終承認待機はfalse", status: WAIT_FINAL_APPROVAL_AUDIT, want: false},
		{name: "住所確認中はfalse", status: CHECK_ADDRESS_AUDIT, want: false},
		{name: "未知の値はfalse", status: AUDIT_STATUS("UNKNOWN"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsFinalized())
		})
	}
}
