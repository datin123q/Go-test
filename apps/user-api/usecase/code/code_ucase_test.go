package code

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/stretchr/testify/assert"
)

func Test_commonUsecase_GetAll(t *testing.T) {
	tmpDir := t.TempDir()
	codePath := filepath.Join(tmpDir, "code")
	domainPath := filepath.Join(codePath, "funding")

	yamlContent := []byte(`
FUNDING_REQUEST:
  REQUEST_STATUS:
    codes:
      - code: STANDBY_REQUEST
        descriptions:
          ko: "申請中"
          en: "Applying"
`)
	err := os.MkdirAll(domainPath, 0755)
	assert.NoErrorf(t, err, "error creating directory: %v", err)

	err = os.WriteFile(filepath.Join(domainPath, "code.yaml"), yamlContent, 0644)
	assert.NoErrorf(t, err, "error writing file: %v", err)

	usecase := codeUsecase{
		code: config.NewCode(tmpDir),
	}

	got := usecase.GetAll(context.TODO())
	assert.Equalf(t, `{
  "FUNDING_REQUEST": {
    "REQUEST_STATUS": {
      "codes": [
        {
          "code": "STANDBY_REQUEST",
          "descriptions": {
            "en": "Applying",
            "ko": "申請中"
          }
        }
      ]
    }
  }
}`, got, "")
}
