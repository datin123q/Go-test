package gmoaozorautils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_HankakuLength(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{
			name: "半角英数字はそのままの文字数",
			s:    "TANAKA TARO",
			want: 11,
		},
		{
			name: "半角カナはそのままの文字数",
			s:    "ﾀﾅｶ",
			want: 3,
		},
		{
			name: "濁点が分離された半角カナは1文字ずつ数える(変換前の全角より長くなる)",
			s:    "ｶﾞｷﾞｸﾞｹﾞｺﾞ",
			want: 10,
		},
		{
			name: "30文字ちょうどは境界値",
			s:    strings.Repeat("A", 30),
			want: 30,
		},
		{
			name: "31文字も境界値",
			s:    strings.Repeat("A", 31),
			want: 31,
		},
		{
			name: "見た目20文字の濁点付き全角カタカナは変換後に半角40文字相当になる",
			s:    ZenkakuToHankaku("ガギグゲゴザジズゼゾダヂヅデドバビブベボ"),
			want: 40,
		},
		{
			name: "空文字は0",
			s:    "",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HankakuLength(tt.s))
		})
	}
}
