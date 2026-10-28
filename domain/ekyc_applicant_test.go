package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ekycTestStrPtr(s string) *string { return &s }

func TestBuildEkycAddress(t *testing.T) {
	tests := []struct {
		name     string
		province string
		city     string
		address  string
		address2 *string
		want     string
	}{
		{"全項目あり。建物名の前だけ空白が入る", "東京都", "千代田区", "丸の内1-1-1", ekycTestStrPtr("ビル101"), "東京都千代田区丸の内1-1-1 ビル101"},
		{"建物名なし(nil)", "東京都", "千代田区", "丸の内1-1-1", nil, "東京都千代田区丸の内1-1-1"},
		{"建物名が空白だけなら無いものとして扱う", "東京都", "千代田区", "丸の内1-1-1", ekycTestStrPtr("   "), "東京都千代田区丸の内1-1-1"},
		{"各項目の前後の空白は取り除く", " 東京都 ", "\t千代田区\n", " 丸の内1-1-1 ", ekycTestStrPtr(" ビル101 "), "東京都千代田区丸の内1-1-1 ビル101"},
		{"空の項目は飛ばす", "", "千代田区", "丸の内1-1-1", nil, "千代田区丸の内1-1-1"},
		{"すべて空なら空文字", "", "", "", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, BuildEkycAddress(tt.province, tt.city, tt.address, tt.address2))
		})
	}
}

func TestBuildEkycName(t *testing.T) {
	tests := []struct {
		name       string
		familyName string
		firstName  string
		want       string
	}{
		{"姓名はスペースなしで連結する(画面のbuildEkycAccountと同じ)", "山田", "太郎", "山田太郎"},
		{"前後の空白は取り除く", " 山田 ", " 太郎 ", "山田太郎"},
		{"名が空なら姓だけ", "山田", "", "山田"},
		{"姓が空なら名だけ", "", "太郎", "太郎"},
		{"空白だけの項目は無いものとして扱う", "山田", "  ", "山田"},
		{"両方空なら空文字", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, BuildEkycName(tt.familyName, tt.firstName))
		})
	}
}
