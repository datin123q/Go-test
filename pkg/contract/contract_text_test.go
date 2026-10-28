package contract

import "testing"

func TestStripVariationSelectors(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "SVS(U+FE00台)を取り除く",
			text: "辻︀田太郎",
			want: "辻田太郎",
		},
		{
			name: "IVS(U+E0100台)を取り除く",
			text: "辻\U000E0100田太郎",
			want: "辻田太郎",
		},
		{
			name: "セレクタが無ければそのまま返す",
			text: "山田太郎",
			want: "山田太郎",
		},
		{
			name: "空文字列はそのまま",
			text: "",
			want: "",
		},
		{
			name: "セレクタの直前・直後の境界値(範囲の端)も取り除く",
			text: "辻️\U000E01EF田",
			want: "辻田",
		},
		{
			name: "範囲外の類似コードポイントは残す",
			text: "辻﷿\U000E01F0田",
			want: "辻﷿\U000E01F0田",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripVariationSelectors(tt.text)
			if got != tt.want {
				t.Errorf("stripVariationSelectors(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}
