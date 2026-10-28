package contract

// AutoFontSizeFractionはUniPDF v3.58.0のannotator.AppearanceStyle{}の
// デフォルト値(0.65)。DAのフォントサイズが0(自動)のとき、矩形の高さに
// この係数を掛けた値を基準フォントサイズにする(MIR-1201、置き換え前の
// UniPDFのFlattenFields(RegenerateTextFields: true)と同じ見た目にするため、
// 挙動をそのまま踏襲した)。
const AutoFontSizeFraction = 0.65

// fontSizeSafetyMarginは、文字列の推定幅が矩形の幅を超える場合に、矩形の
// 幅の何%まで許容するか。UniPDFのannotator実装と同じ0.95。
const fontSizeSafetyMargin = 0.95

// textLeftMarginPtはベースラインのx位置(矩形左下からの相対オフセット、pt)。
// UniPDFが実際に出力したPDFの外観ストリームを計測して得た値。
const textLeftMarginPt = 2.0

// FitFontSizeは、矩形(幅widthPt・高さheightPt)にtextを収めるフォントサイズを
// 計算する。UniPDF v3.58.0のannotatorが実際に使っている計算式を再現している。
//
//  1. 基準サイズ = 高さ × AutoFontSizeFraction
//  2. 「左マージン(textLeftMarginPt) + 基準サイズでの推定幅」が矩形の幅を
//     超える場合だけ、「(矩形の幅 - 左マージン) × fontSizeSafetyMargin」に
//     収まるようフォントサイズを比例で縮小する
//
// 2のマージン控除は、annotator.goのソース(定数 `_egd float64 = 2.0`、
// textLeftMarginPtと同じ値)を読んで特定した。当初は「矩形の幅 ×
// fontSizeSafetyMargin」を閾値にしていたが、この左マージン分を差し引いて
// いなかったため、実際には縮小しなくてよいケースまで誤って縮小していた
// (本番テンプレート全件との精密比較で、26文字の長い氏名を含むケースの
// 一部が一致しないことから発見した。MIR-1201)。
//
// DAでフィールドごとにフォントサイズが固定指定されていても無視する。UniPDFの
// FieldAppearance(OnlyIfMissing: true, RegenerateTextFields: true)+
// AppearanceFontStyle{Fallback: &AppearanceFont{Size: 0}, ForceReplace: true}
// という設定では、DAの固定サイズに関わらず常にこの自動計算が使われることを、
// 本番テンプレート全件との精密比較で確認した(MIR-1201)。当初は「DAが0なら
// 自動、非0ならそのまま使う」という実装にしていたが、DAが非0のテンプレートで
// UniPDFの実測値と一致せず、この誤りに気づいた。
//
// widthAtSizeは、指定サイズでのtextの推定幅を返す関数。呼び出し元がフォント
// 依存のロジック(gofpdfのGetStringWidth等)を注入することで、この関数自体は
// フォントファイルを読まずに境界値テストできる。
func FitFontSize(text string, widthPt, heightPt float64, widthAtSize func(size float64) float64) float64 {
	baseSize := heightPt * AutoFontSizeFraction
	if text == "" || baseSize <= 0 {
		return baseSize
	}

	w := widthAtSize(baseSize)
	if w <= 0 {
		return baseSize
	}

	if textLeftMarginPt+w <= widthPt {
		return baseSize
	}

	// widthPtが左マージン(textLeftMarginPt)より狭い、極端に壊れた矩形の
	// 場合、縮小後のフォントサイズが負になり得る。検証済みの本番テンプレート
	// (ミラリタ44件・ZOLOME4件)にはこれに該当するフィールドは無いが、
	// gofpdf側へ負のフォントサイズを渡すと未定義動作になり得るため
	// 0にクランプする(ZOLOME側の自動レビューで指摘、MIR-1201)。
	shrunk := fontSizeSafetyMargin * (widthPt - textLeftMarginPt) * baseSize / w
	if shrunk < 0 {
		return 0
	}
	return shrunk
}

// AcroFormの/Q(揃え)の値。PDF仕様(左揃え0、中央揃え1、右揃え2)そのまま。
const (
	QuaddingLeft   = 0
	QuaddingCenter = 1
	QuaddingRight  = 2
)

// TextOriginは、フォントサイズ・矩形の幅と高さ・実際に描画する文字列の幅・
// フォントのCapHeight比率(1000 unitsPerEmスケールのCapHeightを1000で割った
// 値)・揃え(quadding、AcroFormの/Q)から、矩形の左下を原点としたベースライン
// 位置(dx, dy)を計算する。UniPDF v3.58.0のannotator実装(シングルラインの
// テキストフィールドで、フィールドの高さがフォントサイズの2倍を超えない
// 通常のケース)をソースから写した式。
//
//	dy = (矩形の高さ - CapHeight比率 × フォントサイズ) / 2
//	dx = 左揃え: 左マージン(textLeftMarginPt)
//	     中央揃え: (矩形の幅 - 文字列の幅) / 2
//	     右揃え: 矩形の幅 - 文字列の幅(マージンの控除は無い)
//
// dyの式は「矩形の中でグリフの見た目の高さを上下中央に置く」計算になって
// いる。以前はUniPDFの実測値1点だけから「フォントサイズに比例する」近似式
// (baselineYRatio)を使っていたが、フォントサイズが基準値から変わる
// (文字列が矩形の幅に収まらず縮小される)ケースで実測と0.3〜4pt程度ずれる
// ことが分かり、ソースを読んで正確な式に置き換えた。
//
// dxの揃え分岐は、当初は左揃え固定だったが、敵対的検証で「右揃え・中央揃え
// のフィールドが本番テンプレートに実在し、無視すると位置がずれる」ことが
// 分かり追加した(MIR-1201)。
func TextOrigin(fontSize, widthPt, heightPt, textWidthPt, capHeightRatio float64, quadding int) (dx, dy float64) {
	dy = (heightPt - capHeightRatio*fontSize) / 2.0

	switch quadding {
	case QuaddingCenter:
		dx = (widthPt - textWidthPt) / 2.0
	case QuaddingRight:
		dx = widthPt - textWidthPt
	default:
		dx = textLeftMarginPt
	}
	return dx, dy
}
