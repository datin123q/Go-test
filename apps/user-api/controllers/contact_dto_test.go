package controllers

import (
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContactInput_Sanitize(t *testing.T) {
	tests := []struct {
		name         string
		input        ContactInput
		wantName     string
		wantTitle    string
		wantContents string
	}{
		{
			name: "(成功) 通常のテキストはそのまま",
			input: ContactInput{
				Name:     "山田太郎",
				Email:    "test@example.com",
				Title:    "お問い合わせ件名",
				Contents: "お問い合わせ内容です",
			},
			wantName:     "山田太郎",
			wantTitle:    "お問い合わせ件名",
			wantContents: "お問い合わせ内容です",
		},
		{
			name: "(成功) XSS攻撃: scriptタグをエスケープ",
			input: ContactInput{
				Name:     "<script>alert('XSS')</script>",
				Email:    "attacker@example.com",
				Title:    "<script>alert('XSS in title')</script>",
				Contents: "<script>alert('XSS in contents')</script>",
			},
			wantName:     "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
			wantTitle:    "&lt;script&gt;alert(&#39;XSS in title&#39;)&lt;/script&gt;",
			wantContents: "&lt;script&gt;alert(&#39;XSS in contents&#39;)&lt;/script&gt;",
		},
		{
			name: "(成功) HTMLインジェクション: imgタグをエスケープ",
			input: ContactInput{
				Name:     "田中<img src=x onerror=alert('XSS')>",
				Email:    "test@example.com",
				Title:    "画像を含む件名<img src=x>",
				Contents: "内容<img src=x onerror=alert('injected')>",
			},
			wantName:     "田中&lt;img src=x onerror=alert(&#39;XSS&#39;)&gt;",
			wantTitle:    "画像を含む件名&lt;img src=x&gt;",
			wantContents: "内容&lt;img src=x onerror=alert(&#39;injected&#39;)&gt;",
		},
		{
			name: "(成功) HTMLインジェクション: iframeタグをエスケープ",
			input: ContactInput{
				Name:     "佐藤",
				Email:    "test@example.com",
				Title:    "件名<iframe src=\"https://evil.com\"></iframe>",
				Contents: "<iframe src=\"https://evil.com\">悪意のあるサイト</iframe>",
			},
			wantName:     "佐藤",
			wantTitle:    "件名&lt;iframe src=&#34;https://evil.com&#34;&gt;&lt;/iframe&gt;",
			wantContents: "&lt;iframe src=&#34;https://evil.com&#34;&gt;悪意のあるサイト&lt;/iframe&gt;",
		},
		{
			name: "(成功) 特殊文字: 不等号とアンパサンドをエスケープ",
			input: ContactInput{
				Name:     "鈴木 A > B",
				Email:    "test@example.com",
				Title:    "A > B について",
				Contents: "C & D について教えてください",
			},
			wantName:     "鈴木 A &gt; B",
			wantTitle:    "A &gt; B について",
			wantContents: "C &amp; D について教えてください",
		},
		{
			name: "(成功) 複数のHTMLタグを含むコード例",
			input: ContactInput{
				Name:     "高橋",
				Email:    "test@example.com",
				Title:    "HTMLタグについて",
				Contents: "ウェブサイトに<div class='container'>を追加したいのですが",
			},
			wantName:     "高橋",
			wantTitle:    "HTMLタグについて",
			wantContents: "ウェブサイトに&lt;div class=&#39;container&#39;&gt;を追加したいのですが",
		},
		{
			name: "(成功) 空文字列",
			input: ContactInput{
				Name:     "",
				Email:    "",
				Title:    "",
				Contents: "",
			},
			wantName:     "",
			wantTitle:    "",
			wantContents: "",
		},
		{
			name: "(成功) JavaScriptイベントハンドラー",
			input: ContactInput{
				Name:     "<button onclick=\"malicious()\">Click</button>",
				Email:    "test@example.com",
				Title:    "<a href=\"javascript:alert('XSS')\">リンク</a>",
				Contents: "<body onload=\"alert('XSS')\">本文</body>",
			},
			wantName:     "&lt;button onclick=&#34;malicious()&#34;&gt;Click&lt;/button&gt;",
			wantTitle:    "&lt;a href=&#34;javascript:alert(&#39;XSS&#39;)&#34;&gt;リンク&lt;/a&gt;",
			wantContents: "&lt;body onload=&#34;alert(&#39;XSS&#39;)&#34;&gt;本文&lt;/body&gt;",
		},
		{
			name: "(成功) SVGベースのXSS",
			input: ContactInput{
				Name:     "伊藤",
				Email:    "test@example.com",
				Title:    "<svg onload=alert('XSS')>",
				Contents: "<svg/onload=alert('XSS')>SVG攻撃</svg>",
			},
			wantName:     "伊藤",
			wantTitle:    "&lt;svg onload=alert(&#39;XSS&#39;)&gt;",
			wantContents: "&lt;svg/onload=alert(&#39;XSS&#39;)&gt;SVG攻撃&lt;/svg&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Sanitize メソッドを呼び出す
			tt.input.Sanitize()

			// 結果を検証（Email はサニタイズ対象外）
			assert.Equal(t, tt.wantName, tt.input.Name)
			assert.Equal(t, tt.wantTitle, tt.input.Title)
			assert.Equal(t, tt.wantContents, tt.input.Contents)
		})
	}
}

func TestContactInput_ToModel(t *testing.T) {
	tests := []struct {
		name    string
		input   ContactInput
		want    *domain.Contact
		wantErr bool
	}{
		{
			name: "(成功) 正常なデータをモデルに変換",
			input: ContactInput{
				Name:     "山田太郎",
				Email:    "yamada@example.com",
				Title:    "件名",
				Contents: "内容",
			},
			want: &domain.Contact{
				Name:     "山田太郎",
				Email:    "yamada@example.com",
				Title:    "件名",
				Contents: "内容",
			},
			wantErr: false,
		},
		{
			name: "(成功) サニタイズ後のデータをモデルに変換",
			input: ContactInput{
				Name:     "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
				Email:    "test@example.com",
				Title:    "&lt;img src=x&gt;",
				Contents: "&lt;iframe src=x&gt;&lt;/iframe&gt;",
			},
			want: &domain.Contact{
				Name:     "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
				Email:    "test@example.com",
				Title:    "&lt;img src=x&gt;",
				Contents: "&lt;iframe src=x&gt;&lt;/iframe&gt;",
			},
			wantErr: false,
		},
		{
			name: "(成功) 空のフィールドを含むデータ",
			input: ContactInput{
				Name:     "",
				Email:    "test@example.com",
				Title:    "",
				Contents: "",
			},
			want: &domain.Contact{
				Name:     "",
				Email:    "test@example.com",
				Title:    "",
				Contents: "",
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
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Email, got.Email)
			assert.Equal(t, tt.want.Title, got.Title)
			assert.Equal(t, tt.want.Contents, got.Contents)
		})
	}
}
