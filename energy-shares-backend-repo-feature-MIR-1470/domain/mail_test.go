package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateDataToMap_AllowsEscapedNewlines(t *testing.T) {
	templateData := "{\"ServiceUrl\":\"http://example.com\",\"Subject\":\"抽選結果のお知らせ\",\"Content\":\"山田太郎様\\n抽選ファンドテスト の抽選について今回は落選となりました。\"}"

	got, err := TemplateDataToMap(templateData)

	assert.NoError(t, err)
	assert.Equal(t, "抽選結果のお知らせ", got["Subject"])
	assert.Equal(t, "山田太郎様\n抽選ファンドテスト の抽選について今回は落選となりました。", got["Content"])
}

// MIR-1470: JPKI方式の本人確認のやり直しメールは、再本登録メールとは別のテンプレートを使い、
// 変数(Receiver・ServiceUrl・Reasons)は同じ形で渡す。
func TestJpkiRetryVerifyTemplate(t *testing.T) {
	assert.Equal(t, "JPKI_RETRY_VERIFY", string(JPKI_RETRY_VERIFY_TEMPLATE))
	assert.NotEqual(t, RETRY_SIGN_UP_TEMPLATE, JPKI_RETRY_VERIFY_TEMPLATE, "再本登録メールのテンプレートを流用しない")
	assert.Equal(t, "JPKI_RETRY_VERIFY", Mail{}.ConvertTemplateName(JPKI_RETRY_VERIFY_TEMPLATE))
}

func TestJpkiRetryVerifyTemplateData(t *testing.T) {
	got, err := jpkiRetryVerifyTemplateData("山田 太郎", []string{"署名用電子証明書が有効期限切れです。", "署名が完了しませんでした。"})
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(got), &decoded))
	assert.Equal(t, "山田 太郎", decoded["Receiver"])
	assert.Contains(t, decoded, "ServiceUrl")
	assert.Equal(t, []interface{}{"署名用電子証明書が有効期限切れです。", "署名が完了しませんでした。"}, decoded["Reasons"])
}

// MIR-942: ファンド名に二重引用符・バックスラッシュが入っても、組み立てた
// JSONが壊れず、元のファンド名がそのまま復元できることをテーブル駆動で確認する。
var mir942ProjectNameCases = []struct {
	name        string
	projectName string
}{
	{"通常のファンド名", "テストファンド1号"},
	{"二重引用符を含む", `"特別"ファンド`},
	{"バックスラッシュを含む", `ファンド\テスト`},
	{"引用符とバックスラッシュの組み合わせ", `"ファンド\テスト"`},
	{"空文字", ""},
}

func TestBuildProjectStatusMailWithProductViewTemplateData(t *testing.T) {
	for _, tt := range mir942ProjectNameCases {
		t.Run(tt.name, func(t *testing.T) {
			td, err := buildProjectStatusMailWithProductViewTemplateData("山田太郎", "https://example.com/product/view/1", tt.projectName)
			require.NoError(t, err)

			var got map[string]string
			require.NoError(t, json.Unmarshal([]byte(td), &got))

			assert.Equal(t, tt.projectName, got["ProjectName"])
			assert.Equal(t, "山田太郎", got["Receiver"])
			assert.Equal(t, "https://example.com/product/view/1", got["ProductView"])
		})
	}
}

func TestBuildProjectStatusMailTemplateData(t *testing.T) {
	for _, tt := range mir942ProjectNameCases {
		t.Run(tt.name, func(t *testing.T) {
			td, err := buildProjectStatusMailTemplateData("山田太郎", tt.projectName)
			require.NoError(t, err)

			var got map[string]string
			require.NoError(t, json.Unmarshal([]byte(td), &got))

			assert.Equal(t, tt.projectName, got["ProjectName"])
			assert.Equal(t, "山田太郎", got["Receiver"])
		})
	}
}

func TestBuildInvestRecruitmentSuccessMailTemplateData(t *testing.T) {
	for _, tt := range mir942ProjectNameCases {
		t.Run(tt.name, func(t *testing.T) {
			td, err := buildInvestRecruitmentSuccessMailTemplateData("山田太郎", tt.projectName)
			require.NoError(t, err)

			var got map[string]string
			require.NoError(t, json.Unmarshal([]byte(td), &got))

			assert.Equal(t, tt.projectName, got["ProjectName"])
			assert.Equal(t, "山田太郎", got["Receiver"])
		})
	}
}

func TestBuildInvestStandbyImmediateSuccessMailTemplateData(t *testing.T) {
	for _, tt := range mir942ProjectNameCases {
		t.Run(tt.name, func(t *testing.T) {
			td, err := buildInvestStandbyImmediateSuccessMailTemplateData("山田太郎", tt.projectName, "1000000")
			require.NoError(t, err)

			var got map[string]string
			require.NoError(t, json.Unmarshal([]byte(td), &got))

			assert.Equal(t, tt.projectName, got["ProjectName"])
			assert.Equal(t, "山田太郎", got["Receiver"])
			// MIR-971: modelhelper.ConvertNumberToPriceでカンマ区切りの表記になる。
			assert.Equal(t, "1,000,000", got["InvestmentAmount"])
		})
	}

	t.Run("投資金額は桁区切り後の値がそのまま入る", func(t *testing.T) {
		td, err := buildInvestStandbyImmediateSuccessMailTemplateData("山田太郎", "テストファンド", "500000")
		require.NoError(t, err)

		var got map[string]string
		require.NoError(t, json.Unmarshal([]byte(td), &got))

		assert.Equal(t, "500,000", got["InvestmentAmount"])
	})
}
