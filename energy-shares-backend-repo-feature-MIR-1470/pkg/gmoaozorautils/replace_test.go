package gmoaozorautils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZenkakuToHankaku(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "全角数字", in: "０１２３", want: "0123"},
		{name: "全角カタカナ", in: "カブシキガイシャ", want: "ｶﾌﾞｼｷｶﾞｲｼｬ"},
		// MIR-1171: depositorValidationは全角英字を許可するが、ここが変換しないと
		// GMOあおぞらネット銀行へ全角のまま送られてしまう。
		{name: "全角英大文字", in: "ＡＢＣ", want: "ABC"},
		{name: "全角英小文字", in: "ａｂｃ", want: "abc"},
		{name: "全角英数字と全角カタカナの混在", in: "ＡＢＣｶﾌﾞｼｷ会社", want: "ABCｶﾌﾞｼｷ会社"},
		{name: "半角文字は変化しない", in: "ABCabc123", want: "ABCabc123"},
		{name: "漢字は変化しない", in: "株式会社", want: "株式会社"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ZenkakuToHankaku(tt.in))
		})
	}
}
