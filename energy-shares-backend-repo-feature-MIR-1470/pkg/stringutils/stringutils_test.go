package stringutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "(成功) 通常のテキストはそのまま",
			input: "これは通常のテキストです",
			want:  "これは通常のテキストです",
		},
		{
			name:  "(成功) XSS攻撃: scriptタグをエスケープ",
			input: "<script>alert('XSS')</script>",
			want:  "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
		},
		{
			name:  "(成功) HTMLインジェクション: imgタグをエスケープ",
			input: "<img src=x onerror=alert('injected')>",
			want:  "&lt;img src=x onerror=alert(&#39;injected&#39;)&gt;",
		},
		{
			name:  "(成功) HTMLインジェクション: iframeタグをエスケープ",
			input: "<iframe src=\"https://evil.com\"></iframe>",
			want:  "&lt;iframe src=&#34;https://evil.com&#34;&gt;&lt;/iframe&gt;",
		},
		{
			name:  "(成功) 属性インジェクション: ダブルクォートをエスケープ",
			input: "\" onload=\"alert('attr injection')\"",
			want:  "&#34; onload=&#34;alert(&#39;attr injection&#39;)&#34;",
		},
		{
			name:  "(成功) 特殊文字: 不等号をエスケープ",
			input: "A > B かつ C < D",
			want:  "A &gt; B かつ C &lt; D",
		},
		{
			name:  "(成功) 特殊文字: アンパサンドをエスケープ",
			input: "A & B",
			want:  "A &amp; B",
		},
		{
			name:  "(成功) 複数のHTMLタグを含む文章",
			input: "<div class='container'><p>段落</p></div>",
			want:  "&lt;div class=&#39;container&#39;&gt;&lt;p&gt;段落&lt;/p&gt;&lt;/div&gt;",
		},
		{
			name:  "(成功) 空文字列",
			input: "",
			want:  "",
		},
		{
			name:  "(成功) JavaScriptイベントハンドラー",
			input: "<button onclick=\"malicious()\">Click</button>",
			want:  "&lt;button onclick=&#34;malicious()&#34;&gt;Click&lt;/button&gt;",
		},
		{
			name:  "(成功) SVGベースのXSS",
			input: "<svg onload=alert('XSS')>",
			want:  "&lt;svg onload=alert(&#39;XSS&#39;)&gt;",
		},
		{
			name:  "(成功) <br>タグもエスケープする",
			input: "テストです<br>改行後",
			want:  "テストです&lt;br&gt;改行後",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeHTML(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeHTMLKeepBR(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "(成功) 通常のテキストはそのまま",
			input: "これは通常のテキストです",
			want:  "これは通常のテキストです",
		},
		{
			name:  "(成功) XSS攻撃: scriptタグをエスケープ",
			input: "<script>alert('XSS')</script>",
			want:  "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
		},
		{
			name:  "(成功) <br>タグはエスケープしない",
			input: "テストです<br>改行後",
			want:  "テストです<br />改行後",
		},
		{
			name:  "(成功) <br/>タグはエスケープしない",
			input: "テストです<br/>改行後",
			want:  "テストです<br />改行後",
		},
		{
			name:  "(成功) <br />タグはエスケープしない",
			input: "テストです<br />改行後",
			want:  "テストです<br />改行後",
		},
		{
			name:  "(成功) 複数の<br>タグを含む文章",
			input: "1行目<br>2行目<br />3行目<br/>4行目",
			want:  "1行目<br />2行目<br />3行目<br />4行目",
		},
		{
			name:  "(成功) <br>タグと他のHTMLタグが混在",
			input: "<script>alert('XSS')</script><br>改行後",
			want:  "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;<br />改行後",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeHTMLKeepBR(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
