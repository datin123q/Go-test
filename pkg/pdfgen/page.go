package pdfgen

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
)

const (
	TITLE_AREA_TOP  = 185
	TITLE_AREA_LEFT = 0
	TITLE_FONT_SIZE = 30
	BASIC_FONT_SIZE = 20
)

func CreateCoverPage(pdf gofpdf.Pdf, pname, spv, endAt string) (*gofpdf.Template, error) {

	template := pdf.CreateTemplate(func(tpl *gofpdf.Tpl) {

		_, h := tpl.GetPageSize()

		tpl.Image(assetRootPath+"/img/cover.png", 0, 0, 210, 0, false, "", 0, "")

		tpl.SetMargins(COVER_LEFT_MARGIN, A4_TOP_MARGIN, COVER_RIGHT_MARGIN)
		tpl.SetFont("noto", "D", TITLE_FONT_SIZE)
		tpl.SetTextColor(0, 0, 139)
		tpl.SetX(TITLE_AREA_LEFT)
		tpl.SetY(TITLE_AREA_TOP)
		tpl.CellFormat(0, 0, pname, "0", 0, "L", false, 0, "")

		tpl.SetFont("noto", "D", BASIC_FONT_SIZE)
		tpl.SetTextColor(0, 0, 0)
		tpl.SetY(h - 80)
		tpl.CellFormat(0, 0, spv, "0", 0, "L", false, 0, "")
		tpl.Ln(8)
		tpl.SetFont("noto", "D", 11)
		tpl.CellFormat(0, 0, i18nutils.DefaultTranslate("pdf.spv_info.end_at")+" ~ "+endAt, "0", 0, "L", false, 0, "")
	})
	if !pdf.Ok() {
		return nil, pdf.Error()
	}
	return &template, nil
}

func CreateInfoPage(pdf gofpdf.Pdf, projectinfo ProductCatalogInfo) (*gofpdf.Template, error) {
	var (
		rsp *http.Response
		err error
		tp  string
	)
	template := pdf.CreateTemplate(func(tpl *gofpdf.Tpl) {

		var lastY float64
		w, _ := tpl.GetPageSize()

		tpl.SetMargins(A4_LEFT_MARGIN, A4_TOP_MARGIN, A4_RIGHT_MARGIN)
		tpl.SetFont("noto", "B", BASIC_FONT_SIZE)
		tpl.Ln(30)

		var xPos float64
		xPos = 10 // 商品タイプを表示できないため仮指定
		tpl.SetX(xPos)
		tpl.SetFont("noto", "D", BASIC_FONT_SIZE)
		tpl.CellFormat(0, 0, projectinfo.ProjectName, "0", 0, "L", false, 0, "")
		lastY = tpl.GetY() + 8

		width := w - (A4_LEFT_MARGIN + A4_RIGHT_MARGIN)
		height := width * 0.5
		rsp, err = http.Get(projectinfo.MainImageUrl)
		if err == nil {
			tp = strings.Replace(filepath.Ext(projectinfo.MainImageUrl), ".", "", 1)
			infoPtr := tpl.RegisterImageReader(projectinfo.MainImageUrl, tp, rsp.Body)
			if tpl.Ok() {
				_, _ = infoPtr.Extent()
				tpl.Image(projectinfo.MainImageUrl, A4_LEFT_MARGIN, lastY, width, height, false, tp, 0, "")

			}
		} else {
			tpl.SetError(err)
		}

		tpl.Ln(height + 15)
		tpl.SetFont("noto", "D", 14)
		tpl.SetX(A4_LEFT_MARGIN)
		chunks := SplitChunks(projectinfo.Description, 35)
		for _, v := range chunks {
			tpl.CellFormat(w, 0, v, "0", 1, "L", false, 0, "")
			tpl.Ln(7)
		}
	})

	return &template, nil
}
