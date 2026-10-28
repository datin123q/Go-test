package controllers

import (
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQnaInput_Sanitize(t *testing.T) {
	tests := []struct {
		name        string
		input       QnaInput
		wantTitle   string
		wantContent string
	}{
		{
			name: "(成功) 通常のテキストはそのまま",
			input: QnaInput{
				Title:   "お問い合わせ件名",
				Content: "お問い合わせ内容です",
			},
			wantTitle:   "お問い合わせ件名",
			wantContent: "お問い合わせ内容です",
		},
		{
			name: "(成功) XSS攻撃: scriptタグをエスケープ",
			input: QnaInput{
				Title:   "<script>alert('XSS')</script>",
				Content: "<script>alert('XSS in content')</script>",
			},
			wantTitle:   "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
			wantContent: "&lt;script&gt;alert(&#39;XSS in content&#39;)&lt;/script&gt;",
		},
		{
			name: "(成功) HTMLインジェクション: imgタグをエスケープ",
			input: QnaInput{
				Title:   "画像を含む件名<img src=x onerror=alert('XSS')>",
				Content: "内容<img src=x onerror=alert('injected')>",
			},
			wantTitle:   "画像を含む件名&lt;img src=x onerror=alert(&#39;XSS&#39;)&gt;",
			wantContent: "内容&lt;img src=x onerror=alert(&#39;injected&#39;)&gt;",
		},
		{
			name: "(成功) HTMLインジェクション: iframeタグをエスケープ",
			input: QnaInput{
				Title:   "件名<iframe src=\"https://evil.com\"></iframe>",
				Content: "<iframe src=\"https://evil.com\">悪意のあるサイト</iframe>",
			},
			wantTitle:   "件名&lt;iframe src=&#34;https://evil.com&#34;&gt;&lt;/iframe&gt;",
			wantContent: "&lt;iframe src=&#34;https://evil.com&#34;&gt;悪意のあるサイト&lt;/iframe&gt;",
		},
		{
			name: "(成功) 特殊文字: 不等号とアンパサンドをエスケープ",
			input: QnaInput{
				Title:   "A > B",
				Content: "C & D について教えてください",
			},
			wantTitle:   "A &gt; B",
			wantContent: "C &amp; D について教えてください",
		},
		{
			name: "(成功) 複数のHTMLタグを含むコード例",
			input: QnaInput{
				Title:   "HTMLタグについて",
				Content: "ウェブサイトに<div class='container'>を追加したいのですが",
			},
			wantTitle:   "HTMLタグについて",
			wantContent: "ウェブサイトに&lt;div class=&#39;container&#39;&gt;を追加したいのですが",
		},
		{
			name: "(成功) 空文字列",
			input: QnaInput{
				Title:   "",
				Content: "",
			},
			wantTitle:   "",
			wantContent: "",
		},
		{
			name: "(成功) JavaScriptイベントハンドラー",
			input: QnaInput{
				Title:   "<button onclick=\"malicious()\">Click</button>",
				Content: "<a href=\"javascript:alert('XSS')\">リンク</a>",
			},
			wantTitle:   "&lt;button onclick=&#34;malicious()&#34;&gt;Click&lt;/button&gt;",
			wantContent: "&lt;a href=&#34;javascript:alert(&#39;XSS&#39;)&#34;&gt;リンク&lt;/a&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Sanitize メソッドを呼び出す
			tt.input.Sanitize()

			// 結果を検証
			assert.Equal(t, tt.wantTitle, tt.input.Title)
			assert.Equal(t, tt.wantContent, tt.input.Content)
		})
	}
}

func TestQnaInput_ToModel(t *testing.T) {
	tests := []struct {
		name    string
		input   QnaInput
		want    *domain.Qna
		wantErr bool
	}{
		{
			name: "(成功) 正常なデータをモデルに変換",
			input: QnaInput{
				Title:    "件名",
				Content:  "内容",
				CommonId: "common-123",
			},
			want: &domain.Qna{
				Title:    "件名",
				Content:  "内容",
				CommonId: "common-123",
			},
			wantErr: false,
		},
		{
			name: "(成功) サニタイズ後のデータをモデルに変換",
			input: QnaInput{
				Title:    "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
				Content:  "&lt;img src=x&gt;",
				CommonId: "common-456",
			},
			want: &domain.Qna{
				Title:    "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
				Content:  "&lt;img src=x&gt;",
				CommonId: "common-456",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.ToModel()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.want.Title, got.Title)
			assert.Equal(t, tt.want.Content, got.Content)
			assert.Equal(t, tt.want.CommonId, got.CommonId)
		})
	}
}
