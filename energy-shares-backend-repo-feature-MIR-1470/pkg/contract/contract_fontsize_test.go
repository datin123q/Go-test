package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// widthPerCharは、境界値テストで使う簡易widthAtSize。フォントファイルに
// 依存せず、文字数×サイズという単純な比例関係で幅を返す。
func widthPerChar(charCount int) func(size float64) float64 {
	return func(size float64) float64 {
		return float64(charCount) * size
	}
}

func TestFitFontSize(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		widthPt     float64
		heightPt    float64
		widthAtSize func(size float64) float64
		want        float64
	}{
		{
			name:        "空文字列は幅を計算せず基準サイズを返す",
			text:        "",
			widthPt:     100,
			heightPt:    13.08,
			widthAtSize: widthPerChar(2),
			want:        13.08 * AutoFontSizeFraction,
		},
		{
			name:        "高さ0なら基準サイズも0",
			text:        "山田 太郎",
			widthPt:     100,
			heightPt:    0,
			widthAtSize: widthPerChar(1),
			want:        0,
		},
		{
			name:        "幅に余裕があれば基準サイズをそのまま使う",
			text:        "山田",
			widthPt:     100,
			heightPt:    13.08,
			widthAtSize: widthPerChar(1),
			want:        13.08 * AutoFontSizeFraction,
		},
		{
			// MIR-1201: 判定は「左マージン(2.0pt) + 推定幅」と矩形の幅を
			// 比較する(矩形の幅を先に0.95倍するのではない)。ここでは
			// ちょうど「2.0 + w == widthPt」になるようwidthAtSizeを
			// 組み立て、境界(超過しない側)を検証する。
			name:     "左マージン込みの推定幅が矩形の幅にちょうど一致する境界",
			text:     "山田太郎山田太郎",
			widthPt:  100,
			heightPt: 13.08,
			widthAtSize: func(size float64) (w float64) {
				base := 13.08 * AutoFontSizeFraction
				return size / base * (100 - textLeftMarginPt)
			},
			want: 13.08 * AutoFontSizeFraction,
		},
		{
			name:        "文字数が多くても幅の上限内なら縮小しない",
			text:        "26文字の長い氏名がここに入ることを想定したテキストです",
			widthPt:     50,
			heightPt:    13.08,
			widthAtSize: widthPerChar(3),
			// baseSize = 13.08*0.65 = 8.502
			// w = 3 * 8.502 = 25.506
			// 2.0 + 25.506 = 27.506 <= 50 → 超過しないので縮小なし
			want: 13.08 * AutoFontSizeFraction,
		},
		{
			name:        "極端に長い文字列は大きく縮小する",
			text:        "非常に長い文字列非常に長い文字列非常に長い文字列非常に長い文字列",
			widthPt:     50,
			heightPt:    13.08,
			widthAtSize: widthPerChar(10),
			// baseSize = 8.502, w = 85.02
			// 2.0 + 85.02 = 87.02 > 50 → 縮小する
			// want = 0.95 * (50-2.0) * 8.502 / 85.02
			want: fontSizeSafetyMargin * (50 - textLeftMarginPt) * 8.502 / 85.02,
		},
		{
			name:        "widthAtSizeが0以下を返す異常値は基準サイズを返す",
			text:        "山田",
			widthPt:     100,
			heightPt:    13.08,
			widthAtSize: func(size float64) float64 { return 0 },
			want:        13.08 * AutoFontSizeFraction,
		},
		{
			// ZOLOME側の自動レビューで指摘(MIR-1201): widthPtが左マージン
			// (textLeftMarginPt=2.0pt)より狭い、壊れた矩形の場合、
			// 縮小後のフォントサイズが負になり得る。検証済みの本番
			// テンプレートには該当するフィールドは無いが、0にクランプ
			// されることを確認する。
			name:        "矩形の幅が左マージンより狭い異常値は0を返す(負値にならない)",
			text:        "山田太郎",
			widthPt:     1.0,
			heightPt:    13.08,
			widthAtSize: widthPerChar(1),
			want:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.widthAtSize)
			got := FitFontSize(tt.text, tt.widthPt, tt.heightPt, tt.widthAtSize)
			assert.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

func TestTextOrigin(t *testing.T) {
	// notoSansCJKtcCapHeightRatioは、NotoSansCJKtc-Regular.ttfの実際の
	// CapHeight(733、1000 unitsPerEmスケール)を1000で割った値。
	const notoSansCJKtcCapHeightRatio = 0.733

	tests := []struct {
		name           string
		fontSize       float64
		widthPt        float64
		heightPt       float64
		textWidthPt    float64
		capHeightRatio float64
		quadding       int
		wantDx         float64
		wantDy         float64
	}{
		{
			// invest_name(矩形高さ13.08pt)、縮小なしの基準サイズ8.502ptの
			// ケース。UniPDFの実測値dy=3.424ptと一致する。
			name:           "UniPDF実測値その1(高さ13.08pt・フォントサイズ8.502pt、左揃え)を再現する",
			fontSize:       8.502,
			widthPt:        100,
			heightPt:       13.08,
			textWidthPt:    50,
			capHeightRatio: notoSansCJKtcCapHeightRatio,
			quadding:       QuaddingLeft,
			wantDx:         2.0,
			wantDy:         3.424,
		},
		{
			// invest_name(矩形高さ15.12pt)、DA=7.2ptだが無視され基準サイズ
			// 9.828ptになるケース。UniPDFの実測値dy=3.959ptと一致する。
			name:           "UniPDF実測値その2(高さ15.12pt・フォントサイズ9.828pt、左揃え)を再現する",
			fontSize:       9.828,
			widthPt:        100,
			heightPt:       15.12,
			textWidthPt:    50,
			capHeightRatio: notoSansCJKtcCapHeightRatio,
			quadding:       QuaddingLeft,
			wantDx:         2.0,
			wantDy:         3.959,
		},
		{
			name:           "フォントサイズ0ならdyは高さの半分(左揃え)",
			fontSize:       0,
			widthPt:        100,
			heightPt:       13.08,
			textWidthPt:    0,
			capHeightRatio: notoSansCJKtcCapHeightRatio,
			quadding:       QuaddingLeft,
			wantDx:         2.0,
			wantDy:         6.54,
		},
		{
			// MIR-1201: 敵対的検証で発見。本番テンプレートに/Q=1(中央揃え)の
			// フィールド(spv_name)が実在し、揃えを無視すると位置がずれていた。
			name:           "中央揃えは(矩形の幅-文字列の幅)/2",
			fontSize:       8.502,
			widthPt:        100,
			heightPt:       13.08,
			textWidthPt:    40,
			capHeightRatio: notoSansCJKtcCapHeightRatio,
			quadding:       QuaddingCenter,
			wantDx:         30, // (100-40)/2
			wantDy:         3.424,
		},
		{
			// MIR-1201: 敵対的検証で発見。本番テンプレートに/Q=2(右揃え)の
			// フィールド(date1)が本番44件中3件実在した。
			name:           "右揃えは矩形の幅-文字列の幅(左マージンの控除は無い)",
			fontSize:       8.502,
			widthPt:        100,
			heightPt:       13.08,
			textWidthPt:    40,
			capHeightRatio: notoSansCJKtcCapHeightRatio,
			quadding:       QuaddingRight,
			wantDx:         60, // 100-40
			wantDy:         3.424,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dx, dy := TextOrigin(tt.fontSize, tt.widthPt, tt.heightPt, tt.textWidthPt, tt.capHeightRatio, tt.quadding)
			assert.InDelta(t, tt.wantDx, dx, 0.0001)
			assert.InDelta(t, tt.wantDy, dy, 0.001)
		})
	}
}
