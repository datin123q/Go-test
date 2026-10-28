package controllers

import (
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFundingRequestInput_Sanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    FundingRequestInput
		wantName string
	}{
		{
			name: "(成功) 通常の企業名はそのまま",
			input: FundingRequestInput{
				Name: "株式会社サンプル",
			},
			wantName: "株式会社サンプル",
		},
		{
			name: "(成功) XSS攻撃: scriptタグをエスケープ",
			input: FundingRequestInput{
				Name: "<script>alert('XSS')</script>",
			},
			wantName: "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
		},
		{
			name: "(成功) オープンリダイレクト: URLをエスケープ",
			input: FundingRequestInput{
				Name: "https://evil-phishing-site.com",
			},
			wantName: "https://evil-phishing-site.com",
		},
		{
			name: "(成功) HTMLインジェクション: aタグをエスケープ",
			input: FundingRequestInput{
				Name: "<a href=\"https://evil.com\">クリック</a>",
			},
			wantName: "&lt;a href=&#34;https://evil.com&#34;&gt;クリック&lt;/a&gt;",
		},
		{
			name: "(成功) HTMLインジェクション: imgタグをエスケープ",
			input: FundingRequestInput{
				Name: "会社名<img src=x onerror=alert('XSS')>",
			},
			wantName: "会社名&lt;img src=x onerror=alert(&#39;XSS&#39;)&gt;",
		},
		{
			name: "(成功) 特殊文字: 不等号をエスケープ",
			input: FundingRequestInput{
				Name: "A > B 株式会社",
			},
			wantName: "A &gt; B 株式会社",
		},
		{
			name: "(成功) 特殊文字: アンパサンドをエスケープ",
			input: FundingRequestInput{
				Name: "A & B 合同会社",
			},
			wantName: "A &amp; B 合同会社",
		},
		{
			name: "(成功) 空文字列",
			input: FundingRequestInput{
				Name: "",
			},
			wantName: "",
		},
		{
			name: "(成功) iframeタグをエスケープ",
			input: FundingRequestInput{
				Name: "<iframe src=\"https://evil.com\"></iframe>",
			},
			wantName: "&lt;iframe src=&#34;https://evil.com&#34;&gt;&lt;/iframe&gt;",
		},
		{
			name: "(成功) 複合攻撃: URL + HTMLタグ",
			input: FundingRequestInput{
				Name: "<a href=\"https://evil.com\">https://fake-company.com</a>",
			},
			wantName: "&lt;a href=&#34;https://evil.com&#34;&gt;https://fake-company.com&lt;/a&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Sanitize()

			// 結果を検証
			assert.Equal(t, tt.wantName, tt.input.Name)
		})
	}
}

func TestFundingRequestInput_ToModel(t *testing.T) {
	tests := []struct {
		name    string
		input   FundingRequestInput
		want    *domain.FundingRequest
		wantErr bool
	}{
		{
			name: "(成功) 正常なデータをモデルに変換",
			input: FundingRequestInput{
				Name:     "株式会社サンプル",
				ZipCode:  "100-0001",
				Address:  "東京都千代田区千代田1-1",
				Address2: "サンプルビル1F",
				CorpNo:   "1234567890123",
				Province: "東京都",
				City:     "千代田区",
				Email:    "test@example.com",
				Phone:    "0312345678",
			},
			want: &domain.FundingRequest{
				Name:     "株式会社サンプル",
				ZipCode:  "100-0001",
				Address:  "東京都千代田区千代田1-1",
				Address2: "サンプルビル1F",
				CorpNo:   "1234567890123",
				Province: "東京都",
				City:     "千代田区",
				Email:    "test@example.com",
				Phone:    "0312345678",
			},
			wantErr: false,
		},
		{
			name: "(成功) サニタイズ後のデータをモデルに変換",
			input: FundingRequestInput{
				Name:  "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
				Email: "test@example.com",
				Phone: "0312345678",
			},
			want: &domain.FundingRequest{
				Name:  "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
				Email: "test@example.com",
				Phone: "0312345678",
			},
			wantErr: false,
		},
		{
			name: "(成功) エスケープされたURLをモデルに変換",
			input: FundingRequestInput{
				Name:  "&lt;a href=&#34;https://evil.com&#34;&gt;クリック&lt;/a&gt;",
				Email: "test@example.com",
				Phone: "0312345678",
			},
			want: &domain.FundingRequest{
				Name:  "&lt;a href=&#34;https://evil.com&#34;&gt;クリック&lt;/a&gt;",
				Email: "test@example.com",
				Phone: "0312345678",
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
			assert.Equal(t, tt.want.ZipCode, got.ZipCode)
			assert.Equal(t, tt.want.Address, got.Address)
			assert.Equal(t, tt.want.Address2, got.Address2)
			assert.Equal(t, tt.want.CorpNo, got.CorpNo)
			assert.Equal(t, tt.want.Province, got.Province)
			assert.Equal(t, tt.want.City, got.City)
			assert.Equal(t, tt.want.Email, got.Email)
			assert.Equal(t, tt.want.Phone, got.Phone)
		})
	}
}
