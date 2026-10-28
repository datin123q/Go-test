package service

import (
	"encoding/json"
	"regexp"
	"strings"
)

// 検証結果取得APIのレスポンスの errors(エラー情報)を読む(MIR-1470)。
//
// 署名の検証処理でのエラー(利用者が署名を拒否した・署名用電子証明書の期限切れ・ロックなど)は、
// 検証結果通知Webhookにも本人確認BPOサービスのAPIにも含まれず、このAPIのレスポンスにだけ入る
// (ネクスウェイ 2026-10-16 の回答)。エラーコードは仕様書 p.14。
//
// 仕様書のエラー情報は、項目表(p.20)とサンプル(p.28-29)で型の書き方が食い違う
// (errors が「文字列」とも「配列」とも読め、error_detail が「オブジェクト」とも「文字列」とも読める。
// p.29 のサンプルはJSONとして括弧も合っていない)。確認はネクスウェイへ依頼中のため、ここでは
// どの形で来ても、エラーコードだけを取り出す。エラー情報は補助の情報なので、形が想定と違っても、
// レスポンス全体(裏取りに使う reference_id・request_id・status・result)の読み取りは失敗させない。

// jpkiErrorCodePattern は、エラーコードとして受け付ける形。DB・メールへ入れる値なので、想定外の
// 文字列(長い・記号を含む)は捨てる。C01・E05 などが該当する。
var jpkiErrorCodePattern = regexp.MustCompile(`^[A-Za-z][0-9]{2,3}$`)

const (
	// maxJpkiErrorCount は、1件のレスポンスから読むエラーコードの上限。
	maxJpkiErrorCount = 10
	// maxJpkiErrorDecodeDepth は、文字列に入ったJSONを読み直す深さの上限。
	maxJpkiErrorDecodeDepth = 2
)

// JpkiSignatureErrors は、検証結果取得APIのレスポンスの errors から取り出したエラーコード(重複なし・出現順)。
type JpkiSignatureErrors []string

// UnmarshalJSON は、どんな形のJSONでも失敗せず、読めたエラーコードだけを保持する。
func (e *JpkiSignatureErrors) UnmarshalJSON(b []byte) error {
	*e = nil

	var raw interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil
	}

	seen := map[string]bool{}
	var codes []string
	collectJpkiErrorCodes(raw, 0, seen, &codes)
	if len(codes) > 0 {
		*e = codes
	}
	return nil
}

func collectJpkiErrorCodes(v interface{}, depth int, seen map[string]bool, codes *[]string) {
	if len(*codes) >= maxJpkiErrorCount {
		return
	}
	switch value := v.(type) {
	case []interface{}:
		for _, item := range value {
			collectJpkiErrorCodes(item, depth, seen, codes)
		}
	case map[string]interface{}:
		if code, ok := value["error_code"].(string); ok {
			addJpkiErrorCode(code, seen, codes)
		}
	case string:
		// 文字列の中にJSON(配列やオブジェクト)が入っている形、またはエラーコードそのもの。
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		if jpkiErrorCodePattern.MatchString(trimmed) {
			addJpkiErrorCode(trimmed, seen, codes)
			return
		}
		if depth >= maxJpkiErrorDecodeDepth {
			return
		}
		var inner interface{}
		if err := json.Unmarshal([]byte(trimmed), &inner); err == nil {
			collectJpkiErrorCodes(inner, depth+1, seen, codes)
		}
	}
}

func addJpkiErrorCode(code string, seen map[string]bool, codes *[]string) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if !jpkiErrorCodePattern.MatchString(code) || seen[code] || len(*codes) >= maxJpkiErrorCount {
		return
	}
	seen[code] = true
	*codes = append(*codes, code)
}

// jpkiErrorDetails は、仕様書 p.14 のエラーコード・エラー詳細の一覧。署名の検証処理でのエラー(C系)の
// エラー詳細は、利用者向けの文面になっている。E系(デジタル認証サービス側の通信エラーなど)は
// 利用者向けではないため、ここには入れない(JpkiErrorDetail を参照)。
var jpkiErrorDetails = map[string]string{
	"C01": "利用者が署名を拒否しました。",
	"C02": "処理の有効期限が切れました。",
	"C03": "署名トランザクションの有効期限が切れました。",
	"C04": "署名用電子証明書が有効期限切れです。マイナンバーカードの更新手続きが必要です。",
	"C05": "署名用電子証明書が失効しています。マイナンバーカードの更新手続きが必要です。",
	"C06": "署名用電子証明書がロックしています。マイナンバーカードのロック解除が必要です。",
	"C07": "署名用電子証明書が未搭載です。マイナンバーカードの発行手続きが必要です。",
	"C08": "署名用電子証明書の有効期間が開始前です。マイナンバーカードの有効期間を確認してください。",
	"C09": "署名用電子証明書が申請情報と一致しません。正しいマイナンバーカードを使用してください。",
}

// jpkiInternalErrorDetails は、利用者へは見せない、サービス側のエラー(E系)の詳細。記録にだけ残す。
var jpkiInternalErrorDetails = map[string]string{
	"E01": "デジタル認証サービスを開始できませんでした。",
	"E05": "Webhook通知が失敗しました。",
}

// JpkiErrorDetail は、エラーコードに対応する詳細(仕様書 p.14)を返す。未知のコードは、コードだけを返す。
// 管理者・開発者が記録を読むための文面で、利用者への案内には JpkiErrorMemberReason を使う。
func JpkiErrorDetail(code string) string {
	if detail, ok := jpkiErrorDetails[code]; ok {
		return detail
	}
	if detail, ok := jpkiInternalErrorDetails[code]; ok {
		return detail
	}
	return ""
}

// jpkiGenericMemberReason は、利用者へ具体的な理由を示せない(E系・未知のコード・エラー情報なし)ときの文面。
const jpkiGenericMemberReason = "署名が完了しませんでした。"

// jpkiUnconfirmedMemberReasons は、エラーの意味がネクスウェイに確認中で、利用者へ具体的な理由を
// 案内できないコード。記録(ErrorMessage)には仕様書の詳細を残し、メールでは署名が完了しなかったと
// だけ伝える。意味が確定したら、ここから外す(または文面を直す)。
//   - C09: 仕様書は「申請情報と一致しない」だが、何と照合するか(会員が入力した基本4情報か、
//     デジタル認証アプリ内で完結するものか)について、ネクスウェイから2通りの回答があり確認中(2026-10-23)。
//     誤った案内(「正しいマイナンバーカードを使用してください」)を会員へ送らないようにする。
var jpkiUnconfirmedMemberReasons = map[string]bool{
	"C09": true,
}

// JpkiErrorMemberReason は、利用者へのメールに載せる理由を返す。C系は、仕様書のエラー詳細
// (利用者向けの文面)をそのまま使う。E系・未知のコード・意味が確認中のコードは、具体的な内容は
// 出さず、署名が完了しなかったとだけ伝える。
func JpkiErrorMemberReason(code string) string {
	if jpkiUnconfirmedMemberReasons[code] {
		return jpkiGenericMemberReason
	}
	if detail, ok := jpkiErrorDetails[code]; ok {
		return detail
	}
	return jpkiGenericMemberReason
}

// ErrorCodes は、検証結果に含まれるエラーコードを返す(ログ用。コード以外は含まない)。
func (r JpkiSignatureResult) ErrorCodes() []string {
	return []string(r.Errors)
}

// ErrorMessage は、eKYCリクエスト記録のエラーメッセージ(管理者向けの記録)を返す。
// 「C04: 署名用電子証明書が有効期限切れです。…」の形で1件ずつ改行し、末尾も改行する
// (BPOの不一致の理由と同じ形)。エラー情報が無ければ空文字。
//
// 長さの上限は設けない。コードは最大 maxJpkiErrorCount 件で、1件は50文字ほどのため、最大でも500文字ほど
// (eKYCリクエスト記録のエラーメッセージの列は longtext)。詳細の文面を増やすときは、テストで確かめる。
func (r JpkiSignatureResult) ErrorMessage() string {
	if len(r.Errors) == 0 {
		return ""
	}
	var b strings.Builder
	for _, code := range r.Errors {
		b.WriteString(code)
		if detail := JpkiErrorDetail(code); detail != "" {
			b.WriteString(": ")
			b.WriteString(detail)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// MemberReasons は、利用者へのメールに載せる理由(重複なし・出現順)を返す。エラー情報が無い、または
// 具体的な理由を示せないときは、署名が完了しなかったとだけ伝える1件を返す(空にはしない)。
func (r JpkiSignatureResult) MemberReasons() []string {
	var reasons []string
	seen := map[string]bool{}
	for _, code := range r.Errors {
		reason := JpkiErrorMemberReason(code)
		// 具体的な理由が示せないコード(E系・未知・意味が確認中)は、ほかに具体的な理由があるときは載せない。
		if reason == jpkiGenericMemberReason || seen[reason] {
			continue
		}
		seen[reason] = true
		reasons = append(reasons, reason)
	}
	if len(reasons) == 0 {
		return []string{jpkiGenericMemberReason}
	}
	return reasons
}
