package gmoaozorautils

import "unicode/utf8"

// MaxDepositorLength はGMOあおぞらネット銀行の総合振込が受取人名に許容する
// 半角換算の最大文字数(MIR-1178)。corporateclient/model_bulk_transfer.go
// 「受取人名 半角文字 ... 30文字目まで有効（31文字目以降は切り捨てます）」より。
const MaxDepositorLength = 30

// DepositorTruncationWarning はdepositorがMaxDepositorLengthを超えている場合の
// 警告ログのメッセージ(MIR-1178)。送信リクエストを組み立てる時点で出すため、
// GMOへの送信自体が後続で失敗すれば実際には切り捨ても起きない。「送信されれば」
// の意味を込めている。
const DepositorTruncationWarning = "MIR-1178: depositor exceeds GMO's 30-char limit; will be truncated if this transfer is sent"

// HankakuLength はZenkakuToHankaku変換済みの文字列の文字数を返す。引数は
// 変換前ではなく変換後の文字列であること。呼び出し元はGMOへ送るBeneficiaryName
// 自体にもこの変換結果を使うため、ZenkakuToHankakuの呼び出しを1回で済ませられる
// よう、変換自体は関数の外(呼び出し元)で行う設計にしている。濁点・半濁点は
// 半角では2文字になる(例: "ガ" → "ｶﾞ")ため、変換前の見た目の文字数では
// 30文字以内でも変換後に超過することがある。逆に、ZenkakuToHankakuの変換表に
// 無い文字(ひらがな・漢字・全角ローマ字等)は変換されずそのまま1文字として
// 数えられるため、それらを含む名義はこの関数では実際より短く見積もられうる
// (文字種の検証自体はMIR-1171の責務で、このパッケージのスコープ外)。
func HankakuLength(hankaku string) int {
	return utf8.RuneCountInString(hankaku)
}
