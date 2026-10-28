package contract

import (
	"os"
	"testing"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpumodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readTestContext(t *testing.T, path string) *pdfcpumodel.Context {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	ctx, err := pdfcpuapi.ReadValidateAndOptimize(f, pdfcpumodel.NewDefaultConfiguration())
	require.NoError(t, err)
	return ctx
}

func TestReadAcroFormFields(t *testing.T) {
	ctx := readTestContext(t, "../../assets/testdata/contract/契約締結時交付書面_20220401_FORM_NEW.pdf")

	fields, err := readAcroFormFields(ctx)
	require.NoError(t, err)

	tests := []struct {
		name     string
		wantPage int
		wantLL   [2]float64 // 矩形左下(x, y)
		wantUR   [2]float64 // 矩形右上(x, y)
	}{
		{name: "invest_name", wantPage: 1, wantLL: [2]float64{72.72, 679.68}, wantUR: [2]float64{185.76, 694.80}},
		{name: "date1", wantPage: 1, wantLL: [2]float64{383.04, 709.92}, wantUR: [2]float64{496.08, 725.04}},
		{name: "date2", wantPage: 1, wantLL: [2]float64{89.28, 198.00}, wantUR: [2]float64{202.32, 213.12}},
		{name: "spv_name", wantPage: 1, wantLL: [2]float64{91.44, 320.40}, wantUR: [2]float64{237.60, 334.08}},
		{name: "project_name", wantPage: 2, wantLL: [2]float64{87.84, 214.56}, wantUR: [2]float64{447.12, 231.12}},
		{name: "shares_count", wantPage: 2, wantLL: [2]float64{87.84, 157.68}, wantUR: [2]float64{146.16, 174.96}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			widgets, ok := fields[tt.name]
			require.True(t, ok, "field %s not found", tt.name)
			require.Len(t, widgets, 1)

			w := widgets[0]
			assert.Equal(t, tt.wantPage, w.page)
			assert.InDelta(t, tt.wantLL[0], w.rect.LL.X, 0.01)
			assert.InDelta(t, tt.wantLL[1], w.rect.LL.Y, 0.01)
			assert.InDelta(t, tt.wantUR[0], w.rect.UR.X, 0.01)
			assert.InDelta(t, tt.wantUR[1], w.rect.UR.Y, 0.01)
		})
	}
}

func TestReadAcroFormFields_NoForm(t *testing.T) {
	// I-1ESKK201224.pdfはAcroFormを持たない出力済みサンプル。エラーにせず
	// 空のmapを返すこと(フォーム項目の無いテンプレートは何も書き込まず
	// 出力する、という既存挙動を保つため。MIR-1201)。
	ctx := readTestContext(t, "../../assets/testdata/contract/I-1ESKK201224.pdf")

	fields, err := readAcroFormFields(ctx)
	require.NoError(t, err)
	assert.Empty(t, fields)
}
