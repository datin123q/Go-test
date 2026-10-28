package pdfgen

import (
	"path/filepath"

	"github.com/jung-kurt/gofpdf"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

const PAGE_SIZE = "A4"
const (
	A4_LEFT_MARGIN     = 10
	A4_TOP_MARGIN      = 10
	A4_RIGHT_MARGIN    = 10
	COVER_LEFT_MARGIN  = 15
	COVER_RIGHT_MARGIN = 15

	FONT_ROOT_PATH = "font"
	FONT_SANS_CJK  = "NotoSansCJK"
)

var assetRootPath string

func SetDefaultFont(path string) {
	assetRootPath = path
}

type ProductCatalogInfo struct {
	ProjectName  string `json:"project_name"`
	ProjectType  string `json:"project_type"`
	MainImageUrl string `json:"main_image_url"`
	Description  string `json:"description"`
	SpvName      string `json:"spv_name"`
	SpvCorpNo    string `json:"spv_corp_no"`
	SpvPhone     string `json:"spv_phone"`
	SpvMail      string `json:"spv_mail"`
	SpvZipCode   string `json:"spv_zip_code"`
	SpvProvince  string `json:"spv_province"`
	SpvCity      string `json:"spv_city"`
	SpvAddr      string `json:"spv_addr"`
	SpvAddr2     string `json:"spv_addr2"`

	StatAt jsonutils.JSONTime `json:"stat_at"`
	EndAt  jsonutils.JSONTime `json:"end_at"`

	CategoryType string `json:"category_type"`
	Capacity     string `json:"capacity"`

	DividendRate              string  `json:"dividend_rate"`                                     // 配当率
	OperationTerm             string  `json:"operation_term" swagger:"desc(貸出期間, 運用期間)"`         // 貸出期間 運用期間
	MinShareQty               string  `json:"min_share_qty" swagger:"desc(最小購入数量)"`            // 最小購入数量
	SharePrice                string  `json:"share_price" swagger:"desc(1口当たり価格)"`               // 1口当たり価格
	TargetMinShareAmount      string  `json:"target_min_share_amount" swagger:"desc(目標募金額)"`    // 目標募金額
	RatingRateProfitability   float32 `json:"rating_rate_profitability" swagger:"desc(収益性)"`     // 収益性
	RatingDevelopmentStatus   float32 `json:"rating_development_status" swagger:"desc(開発状況)"`    // 開発状況
	RatingDevelopmentRisk     float32 `json:"rating_development_risk" swagger:"desc(開発リスク)"`    // 開発リスク
	RatingTechnicalAssessment float32 `json:"rating_technical_assessment" swagger:"desc(技術評価)"` // 技術評価
	TargetShareQty            string  `json:"target_share_qty" swagger:"desc(目標株式数 総発行株式)"`   // 目標株式数
	TargetShareAmount         string  `json:"target_share_amount" swagger:"desc(上限募金額)"`
	LoanAmount                string  `json:"loan_amount" swagger:"desc(合計貸出額)"` // 合計貸出額
}

// PDF生成関数 新たに生成される商品関連PDFのDefault設定を指定する。
func CreateCatelogPDF() (*gofpdf.Fpdf, error) {
	rfontStr := "NotoSansCJKtc-Regular.ttf"
	bfontStr := "NotoSansCJKtc-Bold.ttf"

	pdf := gofpdf.New("P", "mm", PAGE_SIZE, filepath.Join(assetRootPath, FONT_ROOT_PATH, FONT_SANS_CJK))
	pdf.AddUTF8Font("noto", "D", rfontStr)
	pdf.AddUTF8Font("noto", "B", bfontStr)
	pdf.SetCompression(false)
	pdf.SetFont("Times", "", 12)
	pdf.SetMargins(A4_LEFT_MARGIN, A4_TOP_MARGIN, A4_RIGHT_MARGIN)
	if pdf.Err() {
		return nil, pdf.Error()
	}
	return pdf, nil
}

func SplitChunks(s string, chunkSize int) []string {
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks []string
	chunk := make([]rune, chunkSize)
	len := 0
	for _, r := range s {
		chunk[len] = r
		len++
		if len == chunkSize {
			chunks = append(chunks, string(chunk))
			len = 0
		}
	}
	if len > 0 {
		chunks = append(chunks, string(chunk[:len]))
	}
	return chunks
}

func AutoTextSize(s string, strSize float64, maxLength float64) float64 {

	length := float64(len(s))

	if maxLength-length >= 0 {
		return strSize
	}

	return strSize * (1 / (length / maxLength))
}
