package contract

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pdfgen"
	"github.com/pkg/errors"
)

var contractFontDir string

// SetDefaultFontは、契約書面の記入に使うCJKフォントの置き場所を記憶する。
// 実際の読み込みはPDF生成のたびにgofpdfが行う(MIR-1201、以前のUniPDF実装は
// ここでフォントを一度だけ読み込みグローバル変数に保持していたが、gofpdfは
// フォントディレクトリを都度指定する作りのため、ここではパスの存在確認だけ
// 行う)。
func SetDefaultFont(rootPath string) error {
	dir := filepath.Join(rootPath, pdfgen.FONT_ROOT_PATH, pdfgen.FONT_SANS_CJK)
	if _, err := os.Stat(filepath.Join(dir, contractFontFile)); err != nil {
		return errors.Wrap(err, "Fail Read Contract Noto Font")
	}
	contractFontDir = dir
	return nil
}

func addRightBlankToNumber(number string) string {
	return number + " "
}

func FillFields(templateData []byte, price, amount, sharesNum, name, spvName, projectName, address, managerName string, invest_date jsonutils.JSONTime) (*bytes.Buffer, error) {
	if contractFontDir == "" {
		return nil, errors.New("contract default font is nil")
	}
	year, month, day := time.Now().Date()
	investYear, investMonth, investDay := invest_date.Time().Date()

	values := map[string]string{
		"invest_name":  name,
		"address":      address,
		"manager_name": managerName,
		"spv_name":     spvName,
		"date1":        fmt.Sprintf("%d年 %d月 %d日", year, int(month), day),
		"invest_date":  fmt.Sprintf("%d年 %d月 %d日", investYear, int(investMonth), investDay),
		"project_name": projectName,
		"share_count":  addRightBlankToNumber(sharesNum),
		"share_price":  addRightBlankToNumber(price),
		"share_amount": addRightBlankToNumber(amount),
	}

	return fillContractTemplate(contractFontDir, templateData, values)
}
