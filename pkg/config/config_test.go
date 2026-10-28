package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Code_loadCodes(t *testing.T) {
	// テスト用一時ディレクトリを作成
	tmpDir := t.TempDir()
	codePath := filepath.Join(tmpDir, "code")

	// テスト用ディレクトリ構造を作成
	domains := []string{"funding", "payment"}
	for _, domain := range domains {
		domainPath := filepath.Join(codePath, domain)
		err := os.MkdirAll(domainPath, 0755)
		assert.NoError(t, err)

		// テスト用YAMLファイルを作成
		yamlContent := []byte(`
FUNDING_REQUEST:
  REQUEST_STATUS:
    codes:
      - code: STANDBY_REQUEST
        descriptions:
          ko: "신청중"
          en: "Applying"
`)
		err = os.WriteFile(filepath.Join(domainPath, "code.yaml"), yamlContent, 0644)
		assert.NoError(t, err)
	}

	code := &Code{
		codeMap: make(map[string]interface{}),
	}

	// loadCodesのテスト
	err := code.loadCodes(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, code.codeMap, 2)
	assert.Contains(t, code.codeMap, "funding")
	assert.Contains(t, code.codeMap, "payment")
}

func Test_Code_generateJSON(t *testing.T) {
	tests := []struct {
		name    string
		codeMap map[string]interface{}
		wantErr bool
	}{
		{
			name: "マップを正常に変換",
			codeMap: map[string]interface{}{
				"domain1": map[string]interface{}{
					"key1": "value1",
				},
			},
			wantErr: false,
		},
		{
			name:    "空のマップを変換",
			codeMap: map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code := &Code{
				codeMap: tc.codeMap,
			}

			err := code.generateJSON()
			assert.NoErrorf(t, err, "error: %v", err)

			// JSONが正しく生成されたか確認
			var decoded map[string]interface{}
			err = json.Unmarshal([]byte(code.jsonOutput), &decoded)
			assert.NoError(t, err)

			// デコードされたマップが元のCodeMapのドメイン下の値を含むか確認
			for _, value := range tc.codeMap {
				if m, ok := value.(map[string]interface{}); ok {
					for k, v := range m {
						assert.Equal(t, v, decoded[k])
					}
				}
			}
		})
	}
}
