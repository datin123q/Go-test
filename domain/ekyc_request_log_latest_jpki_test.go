package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// MIR-1470: 署名成功の通知を受けた行より新しく、実際に申し込まれた行が同じ会員にあるかを判断する
// (HasNewerAppliedRequest)。JPKI方式の行どうし、ホ方式の行との組み合わせ。ほかの会員の行は見ない。
func TestEkycRequestLog_HasNewerAppliedRequest_JPKIの通知(t *testing.T) {
	requestId := func(s string) *string { return &s }
	ctx, engine := newEkycJpkiTestDB(t,
		// 会員1: 署名したあと、新しいURLを発行し、署名してBPOへ申請済み
		EkycRequestLog{Id: 1, Email: "a@example.com", UserId: 1, EkycId: "a-old", RequestId: requestId("req-old"), Status: EKYC_REQUEST_STATUS_NONE},
		EkycRequestLog{Id: 2, Email: "a@example.com", UserId: 1, EkycId: "a-new", RequestId: requestId("req-new"), ApplicantId: 12, Status: EKYC_REQUEST_STATUS_NONE},
		// 会員2: 行が1つだけ
		EkycRequestLog{Id: 3, Email: "b@example.com", UserId: 2, EkycId: "b-only", RequestId: requestId("req-b"), Status: EKYC_REQUEST_STATUS_NONE},
		// 会員5: 署名したあと、ホ方式の画面を開いた(新しい行は申し込まれていない)
		EkycRequestLog{Id: 5, Email: "e@example.com", UserId: 5, EkycId: "e-jpki", RequestId: requestId("req-e"), Status: EKYC_REQUEST_STATUS_NONE},
		EkycRequestLog{Id: 6, Email: "e@example.com", UserId: 5, EkycId: "e-ho-opened", Status: EKYC_REQUEST_STATUS_NONE},
		// 会員7: 署名したあと、ホ方式で撮影を終えて完了報告をした(新しい行が申し込まれている)
		EkycRequestLog{Id: 7, Email: "g@example.com", UserId: 7, EkycId: "g-jpki", RequestId: requestId("req-g"), Status: EKYC_REQUEST_STATUS_NONE},
		EkycRequestLog{Id: 8, Email: "g@example.com", UserId: 7, EkycId: "g-ho-done", Status: EKYC_REQUEST_STATUS_SUCCESS},
		// 会員9: 新しいURLを発行したが、署名せずにやめた(署名前のNONE)
		EkycRequestLog{Id: 9, Email: "i@example.com", UserId: 9, EkycId: "i-old", RequestId: requestId("req-i-old"), Status: EKYC_REQUEST_STATUS_NONE},
		EkycRequestLog{Id: 10, Email: "i@example.com", UserId: 9, EkycId: "i-new", RequestId: requestId("req-i-new"), Status: EKYC_REQUEST_STATUS_NONE},
	)

	tests := []struct {
		name string
		id   int64
		want bool
	}{
		{"新しい署名でBPOへ申請済みの行がある: 置き換えられている", 1, true},
		{"会員の最新の行", 2, false},
		{"その会員の唯一の行", 3, false},
		{"あとからホ方式の画面を開いただけ: 置き換えられていない", 5, false},
		{"あとからホ方式で完了報告をした: 置き換えられている", 7, true},
		{"新しいURLを発行しただけ(署名前): 置き換えられていない(発行し直しは、user-api が古い行を失敗にする)", 9, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadEkycJpkiLog(t, engine, tt.id).HasNewerAppliedRequest(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

// MIR-1470: より新しい JPKI方式の行(署名用URLの発行し直し)があるかを判断する(HasNewerJpkiRequest)。
// 署名前でも数える。ホ方式の行は数えない。ほかの会員の行は見ない。
func TestEkycRequestLog_HasNewerJpkiRequest(t *testing.T) {
	requestId := func(s string) *string { return &s }
	ctx, engine := newEkycJpkiTestDB(t,
		// 会員1: 署名したあと、新しいURLを発行した(署名前)
		EkycRequestLog{Id: 1, Email: "a@example.com", UserId: 1, EkycId: "a-old", RequestId: requestId("req-old"), Status: EKYC_REQUEST_STATUS_NONE},
		EkycRequestLog{Id: 2, Email: "a@example.com", UserId: 1, EkycId: "a-new", RequestId: requestId("req-new"), Status: EKYC_REQUEST_STATUS_NONE},
		// 会員3: 署名したあと、ホ方式で完了報告をした(JPKI方式の行ではない)
		EkycRequestLog{Id: 3, Email: "c@example.com", UserId: 3, EkycId: "c-jpki", RequestId: requestId("req-c"), Status: EKYC_REQUEST_STATUS_NONE},
		EkycRequestLog{Id: 4, Email: "c@example.com", UserId: 3, EkycId: "c-ho", Status: EKYC_REQUEST_STATUS_SUCCESS},
		// 会員5: ほかの会員に、新しい JPKI方式の行がある
		EkycRequestLog{Id: 5, Email: "e@example.com", UserId: 5, EkycId: "e-jpki", RequestId: requestId("req-e"), Status: EKYC_REQUEST_STATUS_NONE},
		EkycRequestLog{Id: 6, Email: "f@example.com", UserId: 6, EkycId: "f-jpki", RequestId: requestId("req-f"), Status: EKYC_REQUEST_STATUS_NONE},
	)

	tests := []struct {
		name string
		id   int64
		want bool
	}{
		{"新しいURLを発行した(署名前でも数える)", 1, true},
		{"会員の最新の行", 2, false},
		{"新しい行がホ方式だけ: 数えない(申し込まれたかは HasNewerAppliedRequest で見る)", 3, false},
		{"ほかの会員の行は見ない", 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadEkycJpkiLog(t, engine, tt.id).HasNewerJpkiRequest(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
