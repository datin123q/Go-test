package contract

// isVariationSelectorは、rが異体字セレクタ(IVS/SVS)かどうかを返す。
// 直前の文字に「どの字形で表示するか」を指示するだけの符号で、それ自体は
// グリフを持たない。UniPDF(FlattenFields)も異体字セレクタを描かず、直前の
// 漢字だけを描いていたため、新方式でも同じく取り除いてから描画する
// (MIR-1201、河内さん指示2026-09-28)。
//
//   - U+FE00-U+FE0F: SVS(標準異体字セレクタ)
//   - U+E0100-U+E01EF: IVS(異体字セレクタ、SVSの拡張)
func isVariationSelector(r rune) bool {
	return (r >= 0xFE00 && r <= 0xFE0F) || (r >= 0xE0100 && r <= 0xE01EF)
}

// stripVariationSelectorsは、textから異体字セレクタだけを取り除いた文字列を
// 返す。直前の漢字自体は残るため、「令」に異体字セレクタが付いていた場合は
// 「令」がそのまま描画される(UniPDFと同じ挙動)。
func stripVariationSelectors(text string) string {
	hasSelector := false
	for _, r := range text {
		if isVariationSelector(r) {
			hasSelector = true
			break
		}
	}
	if !hasSelector {
		return text
	}

	runes := make([]rune, 0, len(text))
	for _, r := range text {
		if isVariationSelector(r) {
			continue
		}
		runes = append(runes, r)
	}
	return string(runes)
}
