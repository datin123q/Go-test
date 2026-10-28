package gmoaozorautils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidDepositorCharset(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "半角カタカナのみ", in: "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ", want: true},
		{name: "全角カタカナのみ", in: "カブシキガイシャテスト", want: true},
		{name: "半角/全角英数字と許可された記号", in: "TEST(1) ｶﾌﾞｼｷｶﾞｲｼｬ.,/-", want: true},
		{name: "空文字", in: "", want: true},
		{name: "漢字を含む", in: "株式会社テスト", want: false},
		{name: "ひらがなを含む", in: "かぶしきがいしゃ", want: false},
		{name: "許可されていない記号を含む", in: "テスト＠会社", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidDepositorCharset(tt.in))
		})
	}
}
