package controllers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MIR-1236: operator_history.title(varchar(255))は文字数(コードポイント数)での
// 上限であり、バイト数ではない。日本語や絵文字(utf8mb4で4バイトの文字)が混ざった
// 件名でも、DBと同じ基準(文字数)で255文字に収まることを確認する。
func TestTruncateMailAuditTitle(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		wantTitle string
		wantLen   int
	}{
		{
			name:      "上限-1文字はそのまま",
			title:     strings.Repeat("あ", mailAuditTitleMaxLen-1),
			wantTitle: strings.Repeat("あ", mailAuditTitleMaxLen-1),
			wantLen:   mailAuditTitleMaxLen - 1,
		},
		{
			name:      "上限ちょうどはそのまま",
			title:     strings.Repeat("あ", mailAuditTitleMaxLen),
			wantTitle: strings.Repeat("あ", mailAuditTitleMaxLen),
			wantLen:   mailAuditTitleMaxLen,
		},
		{
			name:      "上限+1文字は上限文字数に切り詰める",
			title:     strings.Repeat("あ", mailAuditTitleMaxLen+1),
			wantTitle: strings.Repeat("あ", mailAuditTitleMaxLen),
			wantLen:   mailAuditTitleMaxLen,
		},
		{
			name:      "日本語だけの長い件名も文字数で切り詰める",
			title:     strings.Repeat("お知らせ", 100), // 400文字
			wantTitle: string([]rune(strings.Repeat("お知らせ", 100))[:mailAuditTitleMaxLen]),
			wantLen:   mailAuditTitleMaxLen,
		},
		{
			name:      "絵文字(4バイト文字)を含んでもバイト数ではなく文字数で切り詰める",
			title:     strings.Repeat("🎉", mailAuditTitleMaxLen+1),
			wantTitle: strings.Repeat("🎉", mailAuditTitleMaxLen),
			wantLen:   mailAuditTitleMaxLen,
		},
		{
			name:      "短い件名はそのまま",
			title:     "通常のお知らせ件名",
			wantTitle: "通常のお知らせ件名",
			wantLen:   9,
		},
		{
			name:      "空文字はそのまま",
			title:     "",
			wantTitle: "",
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateMailAuditTitle(tt.title)
			assert.Equal(t, tt.wantTitle, got)
			assert.Equal(t, tt.wantLen, len([]rune(got)))
		})
	}
}
