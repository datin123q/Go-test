package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decodeJpkiResult(t *testing.T, body string) JpkiSignatureResult {
	t.Helper()
	var r JpkiSignatureResult
	require.NoError(t, json.Unmarshal([]byte(body), &r), "errorsの形が想定と違っても、全体のデコードは失敗させない")
	return r
}

// 仕様書 p.28 のサンプル(処理の中断による有効期限切れ)。completed_at は、仕様書の訂正(ネクスウェイ、
// 2026-10-23)の形式(YYYY-MM-DDThh:mm:ss+09:00)にしてある。値は読まない。
func TestJpkiSignatureResult_Errors_仕様書p28のサンプル(t *testing.T) {
	r := decodeJpkiResult(t, `{
		"reference_id": "ref_20251030155342052",
		"request_id": "01KE68FG200X2NKAQKWKPNT7AC",
		"status": "completed",
		"result": "canceled",
		"completed_at": "2026-02-02T17:49:42+09:00",
		"errors": [{"error_code": "C02", "error_detail": "処理の有効期限が切れました。"}]
	}`)

	assert.Equal(t, "canceled", r.Result)
	assert.Equal(t, []string{"C02"}, r.ErrorCodes())
}

// 仕様書 p.29 のサンプル(エラー2件)。サンプルはJSONとして括弧が合っていないため、正しい形に直した。
func TestJpkiSignatureResult_Errors_複数のエラー(t *testing.T) {
	r := decodeJpkiResult(t, `{
		"reference_id": "ref", "request_id": "req", "status": "completed", "result": "error",
		"errors": [
			{"error_code": "C04", "error_detail": "署名用電子証明書が有効期限切れです。"},
			{"error_code": "E05", "error_detail": "Webhook通知が失敗しました。"}
		]
	}`)

	assert.Equal(t, []string{"C04", "E05"}, r.ErrorCodes())
}

func TestJpkiSignatureErrors_形が違っても取り出せる(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []string
	}{
		{"配列(通常)", `{"errors":[{"error_code":"C01"}]}`, []string{"C01"}},
		{"オブジェクト1件(配列でない)", `{"errors":{"error_code":"C06"}}`, []string{"C06"}},
		{"項目表どおり error_detail がオブジェクト", `{"errors":[{"error_code":"C07","error_detail":{"message":"x"}}]}`, []string{"C07"}},
		{"文字列の中に配列のJSON", `{"errors":"[{\"error_code\":\"C05\"}]"}`, []string{"C05"}},
		{"文字列の中にオブジェクトのJSON", `{"errors":"{\"error_code\":\"C08\"}"}`, []string{"C08"}},
		{"エラーコードだけの文字列", `{"errors":"C09"}`, []string{"C09"}},
		{"小文字・前後の空白は正規化", `{"errors":[{"error_code":" c03 "}]}`, []string{"C03"}},
		{"重複は1件にまとめる", `{"errors":[{"error_code":"C02"},{"error_code":"C02"}]}`, []string{"C02"}},
		{"出現順を保つ", `{"errors":[{"error_code":"E01"},{"error_code":"C01"}]}`, []string{"E01", "C01"}},
		{"nullは空", `{"errors":null}`, nil},
		{"空配列は空", `{"errors":[]}`, nil},
		{"空文字は空", `{"errors":""}`, nil},
		{"項目なしは空", `{"status":"completed"}`, nil},
		{"数値は空", `{"errors":123}`, nil},
		{"真偽値は空", `{"errors":true}`, nil},
		{"error_codeが文字列でない要素は捨てる", `{"errors":[{"error_code":5},{"error_code":"C01"}]}`, []string{"C01"}},
		{"想定外の文字列は捨てる(DB・メールへ入れない)", `{"errors":[{"error_code":"<script>alert(1)</script>"},{"error_code":"C01"}]}`, []string{"C01"}},
		{"長すぎるコードは捨てる", `{"errors":[{"error_code":"C0123456789"}]}`, nil},
		{"壊れた文字列JSONは空", `{"errors":"[{not json"}`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := decodeJpkiResult(t, tt.body)
			assert.Equal(t, tt.want, r.ErrorCodes())
		})
	}
}

func TestJpkiSignatureErrors_件数の上限(t *testing.T) {
	var items []string
	for i := 0; i < 30; i++ {
		items = append(items, `{"error_code":"C`+string(rune('0'+i/10))+string(rune('0'+i%10))+`"}`)
	}
	r := decodeJpkiResult(t, `{"errors":[`+strings.Join(items, ",")+`]}`)

	assert.Len(t, r.ErrorCodes(), maxJpkiErrorCount)
}

func TestJpkiSignatureErrors_文字列の入れ子の深さに上限がある(t *testing.T) {
	// 文字列の中の文字列の中の…と深くしても、読み直しは上限までで止まる(無限に読み込まない)。
	r := decodeJpkiResult(t, `{"errors":"\"\\\"\\\\\\\"[{\\\\\\\\\\\\\\\"error_code\\\\\\\\\\\\\\\":\\\\\\\\\\\\\\\"C01\\\\\\\\\\\\\\\"}]\\\\\\\"\\\"\""}`)

	assert.Empty(t, r.ErrorCodes())
}

// 基本4情報(jpki_result)は、構造体に項目が無いので、デコードしても保持されない。
func TestJpkiSignatureResult_基本4情報は保持しない(t *testing.T) {
	r := decodeJpkiResult(t, `{
		"reference_id": "ref", "request_id": "req", "status": "completed", "result": "succeeded",
		"completed_at": "2026-02-02T17:49:42+09:00",
		"jpki_result": {"name": "公的 花子", "birthday": "420000101", "address": "東京都清瀬市松田町1-1-1", "gender": "female"}
	}`)

	assert.Equal(t, "succeeded", r.Result)
	assert.Empty(t, r.ErrorCodes())
	// 構造体全体を文字列にしても、個人情報は入っていない。
	dumped, err := json.Marshal(r)
	require.NoError(t, err)
	for _, personal := range []string{"公的", "420000101", "清瀬", "female"} {
		assert.NotContains(t, string(dumped), personal)
	}
}

// errorsの形が想定外でも、裏取りに使う項目(reference_id・request_id・status・result)は読める。
func TestJpkiSignatureResult_errorsが想定外でも裏取りの項目は読める(t *testing.T) {
	r := decodeJpkiResult(t, `{
		"reference_id": "ref", "request_id": "req", "status": "completed", "result": "error",
		"errors": [[[["x"]]], 5, null, "???"]
	}`)

	assert.Equal(t, "ref", r.ReferenceId)
	assert.Equal(t, "req", r.RequestId)
	assert.Equal(t, "completed", r.Status)
	assert.Equal(t, "error", r.Result)
	assert.Empty(t, r.ErrorCodes())
}

func TestJpkiErrorDetail(t *testing.T) {
	assert.Equal(t, "署名用電子証明書が有効期限切れです。マイナンバーカードの更新手続きが必要です。", JpkiErrorDetail("C04"))
	assert.Equal(t, "デジタル認証サービスを開始できませんでした。", JpkiErrorDetail("E01"), "E系は記録用の詳細を返す")
	assert.Equal(t, "Webhook通知が失敗しました。", JpkiErrorDetail("E05"))
	assert.Empty(t, JpkiErrorDetail("Z99"), "未知のコードは空")
}

// 仕様書 p.14 のC01〜C09が、すべて詳細を持つ(増えたときに、メール文面の漏れに気付けるように)。
// メールの理由は、意味が確認中のコード(C09)を除いて、具体的な文面になる。
func TestJpkiErrorDetails_C01からC09がそろっている(t *testing.T) {
	for _, code := range []string{"C01", "C02", "C03", "C04", "C05", "C06", "C07", "C08", "C09"} {
		assert.NotEmpty(t, JpkiErrorDetail(code), code)
		if jpkiUnconfirmedMemberReasons[code] {
			continue
		}
		assert.NotEqual(t, jpkiGenericMemberReason, JpkiErrorMemberReason(code), code)
	}
}

// C09は、何と照合するかについてネクスウェイから2通りの回答があり確認中(2026-10-23)。確定するまで、
// 会員へのメールには具体的な理由を載せない。記録(error_message)には仕様書の詳細を残す。
func TestJpkiErrorMemberReason_C09は確定までメールに具体的な理由を載せない(t *testing.T) {
	assert.Equal(t, jpkiGenericMemberReason, JpkiErrorMemberReason("C09"))

	onlyC09 := JpkiSignatureResult{Errors: JpkiSignatureErrors{"C09"}}
	assert.Equal(t, []string{jpkiGenericMemberReason}, onlyC09.MemberReasons())
	for _, reason := range onlyC09.MemberReasons() {
		assert.NotContains(t, reason, "正しいマイナンバーカード", "確認中の案内を会員へ送らない")
	}
	assert.Equal(t, "C09: 署名用電子証明書が申請情報と一致しません。正しいマイナンバーカードを使用してください。\n",
		onlyC09.ErrorMessage(), "記録には仕様書の詳細を残す")

	withC04 := JpkiSignatureResult{Errors: JpkiSignatureErrors{"C09", "C04"}}
	assert.Equal(t, []string{JpkiErrorDetail("C04")}, withC04.MemberReasons(), "ほかに具体的な理由があれば、それだけを載せる")
}

func TestJpkiErrorMemberReason(t *testing.T) {
	assert.Equal(t, "署名用電子証明書がロックしています。マイナンバーカードのロック解除が必要です。", JpkiErrorMemberReason("C06"))
	// サービス側の事情(E系)は、利用者へ具体的な内容を出さない。
	assert.Equal(t, jpkiGenericMemberReason, JpkiErrorMemberReason("E05"))
	assert.Equal(t, jpkiGenericMemberReason, JpkiErrorMemberReason("E01"))
	assert.Equal(t, jpkiGenericMemberReason, JpkiErrorMemberReason("Z99"))
	assert.NotContains(t, JpkiErrorMemberReason("E05"), "Webhook", "内部の用語は利用者へ出さない")
}

func TestJpkiSignatureResult_ErrorMessage(t *testing.T) {
	t.Run("C系とE系を1件ずつ改行し、末尾も改行する", func(t *testing.T) {
		r := JpkiSignatureResult{Errors: JpkiSignatureErrors{"C04", "E05"}}

		assert.Equal(t, "C04: 署名用電子証明書が有効期限切れです。マイナンバーカードの更新手続きが必要です。\nE05: Webhook通知が失敗しました。\n", r.ErrorMessage())
	})
	t.Run("未知のコードは、コードだけを残す", func(t *testing.T) {
		r := JpkiSignatureResult{Errors: JpkiSignatureErrors{"Z99"}}

		assert.Equal(t, "Z99\n", r.ErrorMessage())
	})
	t.Run("エラー情報が無ければ空", func(t *testing.T) {
		assert.Empty(t, JpkiSignatureResult{}.ErrorMessage())
	})
	t.Run("最大件数・最長の詳細でも、1000文字に収まる", func(t *testing.T) {
		// 列(longtext)に収まる長さで、上限を設けなくてよいことの根拠。詳細の文面を増やして、
		// これを超えるようになったら、気付けるようにする。
		var longest string
		for code, detail := range jpkiErrorDetails {
			if len([]rune(detail)) > len([]rune(jpkiErrorDetails[longest])) {
				longest = code
			}
		}
		codes := make(JpkiSignatureErrors, 0, maxJpkiErrorCount)
		for i := 0; i < maxJpkiErrorCount; i++ {
			codes = append(codes, longest)
		}

		got := JpkiSignatureResult{Errors: codes}.ErrorMessage()

		assert.Less(t, len([]rune(got)), 1000)
		assert.True(t, strings.HasPrefix(got, longest+": "))
	})
}

func TestJpkiSignatureResult_MemberReasons(t *testing.T) {
	t.Run("C系はエラー詳細をそのまま使う", func(t *testing.T) {
		r := JpkiSignatureResult{Errors: JpkiSignatureErrors{"C04"}}

		assert.Equal(t, []string{"署名用電子証明書が有効期限切れです。マイナンバーカードの更新手続きが必要です。"}, r.MemberReasons())
	})
	t.Run("C系とE系が混ざるときは、具体的な理由だけを載せる", func(t *testing.T) {
		r := JpkiSignatureResult{Errors: JpkiSignatureErrors{"C04", "E05"}}

		assert.Equal(t, []string{"署名用電子証明書が有効期限切れです。マイナンバーカードの更新手続きが必要です。"}, r.MemberReasons())
	})
	t.Run("E系だけのときは、署名が完了しなかったとだけ伝える", func(t *testing.T) {
		r := JpkiSignatureResult{Errors: JpkiSignatureErrors{"E01"}}

		assert.Equal(t, []string{jpkiGenericMemberReason}, r.MemberReasons())
	})
	t.Run("エラー情報が無くても、空にはしない", func(t *testing.T) {
		assert.Equal(t, []string{jpkiGenericMemberReason}, JpkiSignatureResult{}.MemberReasons())
	})
	t.Run("同じ理由は1件にまとめる", func(t *testing.T) {
		r := JpkiSignatureResult{Errors: JpkiSignatureErrors{"C05", "C05"}}

		assert.Len(t, r.MemberReasons(), 1)
	})
}
