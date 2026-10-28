package contract

// privateUseAreaStart/Endは、gofpdfのSetGlyphAliasで「借りる」BMP私用領域の
// 範囲。NotoSansCJKtc-Regular.ttfはこの範囲(U+E000-U+F8FF、6400文字)に
// 元々グリフを持たないことを確認済み(MIR-1201)。
const privateUseAreaStart = 0xE000
const privateUseAreaEnd = 0xF8FF

// astralAliasTableは、1回の書面生成の中で、Unicode補助面の文字(U+10000
// 以上、拡張漢字等)をBMP私用領域の空き番号に対応づける表を保持する。
//
// gofpdfは内部でCID(PDFの文字コード)にUnicodeコードポイントをそのまま
// 使うため、Identity-Hエンコーディング(2バイト固定、0-0xFFFF)の範囲を
// 超える補助面の文字を正しく描画できない。呼び出し側でこの範囲に一時的に
// 詰め替え、gofpdf側には「私用領域の番号だが、実際のグリフは元の補助面の
// 文字のもの」というエイリアスを登録することで、CIDは2バイトに収まった
// まま正しいグリフを指せるようにする(MIR-1201、河内さんの指示による方式)。
//
// 同じ補助面の文字には同じ番号を割り当てる。1回の書面生成(1回の
// fillContractTemplate呼び出し)の中でだけ有効なテーブルで、生成のたびに
// 新しく作る。
//
// reservedには、書き込む値の中に本物の私用領域の文字(外字。ミラリタ9310・
// 9560の氏名に実在する)が既にある場合、その番号を渡しておく。渡さないと、
// 補助面の文字にたまたま同じ私用領域の番号を割り当ててしまい、本物の外字が
// 別の拡張漢字のグリフに置き換わって見える(レビューで発見した不具合。
// MIR-1201)。
type astralAliasTable struct {
	next     rune
	aliases  map[rune]rune // 補助面の文字 -> 割り当てた私用領域の番号
	reserved map[rune]bool // 書き込む値の中に既にある私用領域の文字
}

func newAstralAliasTable(reserved map[rune]bool) *astralAliasTable {
	return &astralAliasTable{
		next:     privateUseAreaStart,
		aliases:  map[rune]rune{},
		reserved: reserved,
	}
}

// collectPrivateUseRunesは、valuesに含まれる本物の私用領域の文字(BMP、
// U+E000-U+F8FF)を集める。newAstralAliasTableのreservedにそのまま渡す。
func collectPrivateUseRunes(values map[string]string) map[rune]bool {
	reserved := map[rune]bool{}
	for _, v := range values {
		for _, r := range v {
			if r >= privateUseAreaStart && r <= privateUseAreaEnd {
				reserved[r] = true
			}
		}
	}
	return reserved
}

// remapは、textの中の補助面の文字(U+10000以上)を、私用領域の番号に
// 置き換えた文字列を返す。私用領域の空きが尽きた場合(実運用の氏名・住所等
// では起こり得ない想定)は、それ以上の補助面文字は置き換えずそのまま残す。
// 結果的にgofpdf側でグリフが無い文字として扱われ、私用領域の文字と同じく
// 欠落する(パニックはしない)。
func (t *astralAliasTable) remap(text string) string {
	hasAstral := false
	for _, r := range text {
		if r > 0xFFFF {
			hasAstral = true
			break
		}
	}
	if !hasAstral {
		return text
	}

	runes := []rune(text)
	for i, r := range runes {
		if r <= 0xFFFF {
			continue
		}
		alias, ok := t.aliases[r]
		if !ok {
			for t.next <= privateUseAreaEnd && t.reserved[t.next] {
				t.next++
			}
			if t.next > privateUseAreaEnd {
				continue
			}
			alias = t.next
			t.aliases[r] = alias
			t.next++
		}
		runes[i] = alias
	}
	return string(runes)
}
