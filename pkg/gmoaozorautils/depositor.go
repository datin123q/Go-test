package gmoaozorautils

import "regexp"

// depositorCharsetPattern は受取人名(預金者名)に使用できる文字種を表す。全角英数・半角/全角
// カタカナ・句読点等のみを許可し、漢字・ひらがなを除外する。GMOあおぞらネット銀行は振込CSVの
// 受取人名に漢字が含まれると、その行だけでなくバッチ全体を拒否するため(MIR-1171)、
// energy-share-next の depositorValidation(src/utils/validation.ts)と同じ文字種にそろえている。
var depositorCharsetPattern = regexp.MustCompile(`^[ァ-ヶー｡-ﾟA-Za-zＡ-Ｚａ-ｚ0-9０-９()（）.．,、。「」・/／\-－ 　]*$`)

// IsValidDepositorCharset は depositor(預金者名)がGMOあおぞらネット銀行の振込に使用できる
// 文字種のみで構成されているかを返す。
func IsValidDepositorCharset(s string) bool {
	return depositorCharsetPattern.MatchString(s)
}
