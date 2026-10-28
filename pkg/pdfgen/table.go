package pdfgen

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
)

const (
	LEFT_COLUMN_SIZE  = 35
	LEFT_COLUMN_SIZE2 = 50
	RIGHT_COLUMN_SIZE = 46
	COLUMN_HIEGHT     = 13
	GAP_SIZE          = 3
)

// 投資情報
func InvestInfoTableForEquity(pdf gofpdf.Pdf, p ProductCatalogInfo) *gofpdf.Template {
	template := pdf.CreateTemplate(func(tpl *gofpdf.Tpl) {
		w, _ := tpl.GetPageSize()

		tpl.SetMargins(A4_LEFT_MARGIN, A4_TOP_MARGIN, A4_RIGHT_MARGIN)
		tpl.SetX(A4_LEFT_MARGIN)
		tpl.SetY(30)

		// 投資情報
		categoryToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.invest_info"), A4_LEFT_MARGIN)

		// Colors, line width and bold font
		tpl.SetFillColor(241, 241, 241)
		tpl.SetTextColor(0, 0, 0)
		tpl.SetDashPattern([]float64{0, 0.3}, 0)
		tpl.SetLineWidth(.1)

		width := w - (GAP_SIZE + GAP_SIZE + A4_LEFT_MARGIN + A4_RIGHT_MARGIN + LEFT_COLUMN_SIZE2)

		tpl.SetFont("noto", "D", COLUMN_HIEGHT)

		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.spv_developer"), p.SpvName, A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE2, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.power_plant"), i18nutils.DefaultTranslate("commoncode.PROJECT_CATEGORY."+p.CategoryType), A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE2, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.scale_fund"), p.Capacity, A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE2, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.interest_rate"), p.DividendRate+"%", A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE2, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.funding_date"), fmt.Sprintf("%s ~ %s", p.StatAt.Time().Format(time.DateOnly), p.EndAt.Time().Format(time.DateOnly)), A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE2, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.term_of_operation"), p.OperationTerm+i18nutils.DefaultTranslate("pdf.invest_info.month"), A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE2, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.funding_shares_qty"), p.TargetShareQty+" 口", A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE2, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.price_per_share"), "¥ "+p.SharePrice, A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE2, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.target_fundraising_amount"), "¥ "+p.TargetShareAmount, A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE2, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.minimum_fundraising_amount"), "¥ "+p.TargetMinShareAmount, A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE2, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.invest_info.minimum_number_of_investment_shares"), p.MinShareQty+" 口", A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE2, width)
	})

	return &template

}

// プロジェクト評価
func EvaluationTable(pdf gofpdf.Pdf, p ProductCatalogInfo) *gofpdf.Template {

	const (
		LEFT_COLUMN_SIZE = 75
	)
	template := pdf.CreateTemplate(func(tpl *gofpdf.Tpl) {

		w, _ := tpl.GetPageSize()

		tpl.SetMargins(A4_LEFT_MARGIN, A4_TOP_MARGIN, A4_RIGHT_MARGIN)
		tpl.SetX(A4_LEFT_MARGIN)
		tpl.SetY(215)

		// プロジェクト評価
		categoryToTable(tpl, i18nutils.DefaultTranslate("pdf.evaluation_project.evaluation_project"), A4_LEFT_MARGIN)

		// Colors, line width and bold font
		tpl.SetFillColor(241, 241, 241)
		tpl.SetTextColor(0, 0, 0)
		tpl.SetDashPattern([]float64{0, 0.3}, 0)

		tpl.SetLineWidth(.1)

		tpl.SetFont("noto", "D", 10)
		tpl.SetTextColor(155, 155, 155)
		tpl.CellFormat(w-(A4_LEFT_MARGIN+A4_RIGHT_MARGIN), COLUMN_HIEGHT, fmt.Sprintf("%s:%s / %s:%s / %s:%s / %s:%s",
			i18nutils.DefaultTranslate("pdf.evaluation_project.grade_a"), i18nutils.DefaultTranslate("pdf.evaluation_project.grade_a_risk"),
			i18nutils.DefaultTranslate("pdf.evaluation_project.grade_b"), i18nutils.DefaultTranslate("pdf.evaluation_project.grade_b_risk"),
			i18nutils.DefaultTranslate("pdf.evaluation_project.grade_c"), i18nutils.DefaultTranslate("pdf.evaluation_project.grade_c_risk"),
			i18nutils.DefaultTranslate("pdf.evaluation_project.grade_d"), i18nutils.DefaultTranslate("pdf.evaluation_project.grade_d_risk")),
			"0", 0, "R", false, 0, "")
	})

	return &template

}

// 会社情報
func SpvInfoTable(pdf gofpdf.Pdf, p ProductCatalogInfo) *gofpdf.Template {

	template := pdf.CreateTemplate(func(tpl *gofpdf.Tpl) {

		w, _ := tpl.GetPageSize()

		tpl.SetMargins(A4_LEFT_MARGIN, A4_TOP_MARGIN, A4_RIGHT_MARGIN)
		tpl.SetX(A4_LEFT_MARGIN)
		tpl.SetY(185)

		// 会社情報
		categoryToTable(tpl, i18nutils.DefaultTranslate("pdf.spv_info.corp_info"), A4_LEFT_MARGIN)

		// Colors, line width and bold font
		tpl.SetFillColor(241, 241, 241)
		tpl.SetTextColor(0, 0, 0)
		tpl.SetDashPattern([]float64{0, 0.3}, 0)

		tpl.SetLineWidth(.1)

		width := w - (GAP_SIZE + GAP_SIZE + A4_LEFT_MARGIN + A4_RIGHT_MARGIN + LEFT_COLUMN_SIZE)

		tpl.SetFont("noto", "D", COLUMN_HIEGHT)

		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.spv_info.corp_name"), p.SpvName, A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.spv_info.corp_no"), p.SpvCorpNo, A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE, width)
		writeToTable(tpl, i18nutils.DefaultTranslate("pdf.spv_info.addr"), fmt.Sprintf("%s %s %s %s %s", p.SpvZipCode, p.SpvProvince, p.SpvCity, p.SpvAddr, p.SpvAddr2), A4_LEFT_MARGIN, GAP_SIZE, COLUMN_HIEGHT, LEFT_COLUMN_SIZE, width)
	})
	return &template
}

func categoryToTable(tpl *gofpdf.Tpl, category string, leftMargin float64) {

	tpl.SetFont("noto", "D", 7)
	tpl.SetTextColor(0, 0, 139)
	tpl.CellFormat(5, 0, "●", "0", 0, "L", false, 0, "")
	tpl.SetFont("noto", "B", BASIC_FONT_SIZE)
	tpl.CellFormat(leftMargin, 0, category, "0", 0, "L", false, 0, "")
	tpl.Ln(6)
}

func writeToTable(tpl *gofpdf.Tpl, leftStr, rightStr string, leftMargin, gabSize, columnHeight, leftColumnSize, width float64) {

	strs := SplitChunks(rightStr, 35)
	length := float64(len(strs))

	tpl.SetTextColor(155, 155, 155)
	tpl.CellFormat(gabSize, columnHeight*length, "", "TB", 0, "L", true, 0, "")
	tpl.CellFormat(leftColumnSize, columnHeight*length, leftStr, "TB", 0, "L", true, 0, "")
	tpl.SetTextColor(0, 0, 0)
	for index, v := range strs {
		if index > 0 {
			tpl.SetX(leftMargin + gabSize + leftColumnSize)
		}
		tpl.CellFormat(gabSize, columnHeight, "", "TB", 0, "L", false, 0, "")
		tpl.CellFormat(width, columnHeight, v, "TB", 0, "L", false, 0, "")
		tpl.Ln(-1)
	}
}

func writeToTable2(tpl *gofpdf.Tpl, leftStr1, leftStr2, rightStr1, rightStr2 string, gabSize, columnHeight, leftColumnSize, width, fontSize float64) {

	tpl.SetFont("noto", "D", fontSize)
	tpl.SetTextColor(155, 155, 155)
	tpl.CellFormat(gabSize, columnHeight, "", "TB", 0, "L", true, 0, "")
	tpl.CellFormat(leftColumnSize, columnHeight, leftStr1, "TB", 0, "L", true, 0, "")

	tpl.SetFont("noto", "D", fontSize)
	tpl.SetTextColor(0, 0, 0)
	tpl.CellFormat(gabSize, columnHeight, "", "TB", 0, "L", false, 0, "")
	tpl.CellFormat((width*0.5)-(leftColumnSize*0.5+gabSize), columnHeight, leftStr2, "TB", 0, "L", false, 0, "")

	tpl.SetFont("noto", "D", fontSize)
	tpl.SetTextColor(155, 155, 155)
	tpl.CellFormat(gabSize, columnHeight, "", "TB", 0, "L", true, 0, "")
	tpl.CellFormat(leftColumnSize, columnHeight, rightStr1, "TB", 0, "L", true, 0, "")

	tpl.SetFont("noto", "D", fontSize)
	tpl.SetTextColor(0, 0, 0)
	tpl.CellFormat(gabSize, columnHeight, "", "TB", 0, "L", false, 0, "")
	tpl.CellFormat((width*0.5)-(leftColumnSize*0.5+gabSize), columnHeight, rightStr2, "TB", 0, "L", false, 0, "")
	tpl.Ln(-1)
}
