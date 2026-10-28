package gmo_aozora

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type testSwaggerError struct {
	msg   string
	body  []byte
	model interface{}
}

func (e testSwaggerError) Error() string      { return e.msg }
func (e testSwaggerError) Body() []byte       { return e.body }
func (e testSwaggerError) Model() interface{} { return e.model }

type testErrorResponse struct {
	ErrorCode    string
	ErrorMessage string
}

// GetSwaggerErrorDetail は、err.(client.GenericSwaggerError) という無条件の型アサーションが
// CallWith429Retry のラップしたエラーに対して panic していた問題への対応(MIR-1127/MIR-1128)。
// errors.As でラップの有無に関わらず安全に取り出せることを確認する。
func TestGetSwaggerErrorDetail(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want interface{}
	}{
		{
			name: "swaggerErrorでなければnilを返す",
			err:  errors.New("network error"),
			want: nil,
		},
		{
			name: "CallWith429Retryにラップされていてもerrors.Asで取り出してModelを返す",
			err:  errors.Wrap(testSwaggerError{msg: "boom", model: "decoded-model"}, "wrapped by retry"),
			want: "decoded-model",
		},
		{
			name: "Modelがnilでもbodyがあれば文字列で返す",
			err:  testSwaggerError{msg: "boom", body: []byte("raw body")},
			want: "raw body",
		},
		{
			name: "Modelとbodyの両方があればModelを優先する",
			err:  testSwaggerError{msg: "boom", model: "decoded-model", body: []byte("raw body")},
			want: "decoded-model",
		},
		{
			name: "Modelもbodyも無ければnilを返す",
			err:  testSwaggerError{msg: "boom"},
			want: nil,
		},
		{
			name: "Modelがゼロ値の構造体ならbodyへフォールバックする",
			err:  testSwaggerError{msg: "boom", model: testErrorResponse{}, body: []byte("raw body")},
			want: "raw body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				assert.Equal(t, tt.want, GetSwaggerErrorDetail(tt.err))
			})
		})
	}
}

// GetSwaggerErrorDetailString は、GetSwaggerErrorDetail の戻り値をそのまま fmt.Sprintf("%+v", ...)
// すると nil のときにリテラル文字列 "<nil>" がログやエラーメッセージへ混入していた問題への対応(MIR-1128)。
func TestGetSwaggerErrorDetailString(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "swaggerErrorでなければ空文字を返す(<nil>にしない)",
			err:  errors.New("network error"),
			want: "",
		},
		{
			name: "detailがあれば%+vで文字列化する",
			err:  testSwaggerError{msg: "boom", model: "decoded-model"},
			want: "decoded-model",
		},
		{
			name: "Model経由でも上限を超えれば切り詰める",
			err:  testSwaggerError{msg: "boom", model: strings.Repeat("a", maxSwaggerErrorDetailLen+1)},
			want: strings.Repeat("a", maxSwaggerErrorDetailLen) + "...(truncated)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				assert.Equal(t, tt.want, GetSwaggerErrorDetailString(tt.err))
			})
		})
	}
}

// truncateSwaggerErrorDetailString は、中間装置(WAF/ALB等)が HTML のエラーページを丸ごと返した
// 場合や、GMOが明細単位のエラーを大量の明細に対して返した場合に、ログ・エラーメッセージへ全文が
// 載るのを防ぐための上限(MIR-1128)。上限以下は変更せず、超えた場合は切り詰めたことが分かる印を
// 末尾に付けることを確認する。
func TestTruncateSwaggerErrorDetailString(t *testing.T) {
	t.Run("上限以下ならそのまま返す", func(t *testing.T) {
		s := strings.Repeat("a", maxSwaggerErrorDetailLen)
		got := truncateSwaggerErrorDetailString(s)
		assert.Equal(t, s, got)
		assert.False(t, strings.HasSuffix(got, "...(truncated)"))
	})

	t.Run("上限を超えたら切り詰めて印を付ける", func(t *testing.T) {
		s := strings.Repeat("a", maxSwaggerErrorDetailLen+1)
		got := truncateSwaggerErrorDetailString(s)
		assert.Len(t, got, maxSwaggerErrorDetailLen+len("...(truncated)"))
		assert.True(t, strings.HasSuffix(got, "...(truncated)"))
	})

	t.Run("マルチバイト文字の途中で切れる場合は文字境界まで戻る", func(t *testing.T) {
		s := strings.Repeat("a", maxSwaggerErrorDetailLen-1) + "あ"
		got := truncateSwaggerErrorDetailString(s)
		assert.True(t, utf8.ValidString(got))
		assert.True(t, strings.HasSuffix(got, "...(truncated)"))
	})
}
