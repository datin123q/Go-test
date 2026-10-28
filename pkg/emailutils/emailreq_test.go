package emailutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClient_DebugIsDisabled はMIR-1345の回帰テスト。init()で作る共通の
// clientとretryClient(clientのClone)がSetDebug(true)になっていた。
// デバッグ出力(~~~ REQUEST ~~~)には宛先・氏名・振込先がそのまま含まれ、
// 本番のログ(CloudWatch、保持30日)に出ていた。
func TestClient_DebugIsDisabled(t *testing.T) {
	assert.False(t, client.Debug, "共通のclientはSetDebug(true)になっていないこと")
	assert.False(t, retryClient.Debug, "retryClient(clientのClone)もSetDebug(true)になっていないこと")
}

func Test_countReceivers(t *testing.T) {
	tests := []struct {
		name     string
		mailData []MailData
		want     int
	}{
		{"MailDataがnilなら0", nil, 0},
		{"MailDataが空なら0", []MailData{}, 0},
		{"Receiversが1件なら1", []MailData{{Receivers: []string{"a@example.com"}}}, 1},
		{"Receiversが複数件なら合計", []MailData{{Receivers: []string{"a@example.com", "b@example.com"}}}, 2},
		{"MailDataが複数件なら全件を合算", []MailData{
			{Receivers: []string{"a@example.com"}},
			{Receivers: []string{"b@example.com", "c@example.com"}},
		}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, countReceivers(tt.mailData))
		})
	}
}
