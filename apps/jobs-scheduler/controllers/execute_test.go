package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"energyshares-job-scheduler/jobs"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

func Test_ExecuteApiControllers_ContextHandling(t *testing.T) {
	var request jobs.ReportDateInfo

	// Echoインスタンス作成
	e := echo.New()

	// テストするJSONデータの準備
	requestBody := `{
        "year": "2024",
        "quarter": "QUARTER_FOURTH"
    }`

	// HTTPリクエスト作成
	req := httptest.NewRequest(http.MethodPost, "/api/report", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	// レスポンスレコーダー作成
	rec := httptest.NewRecorder()

	// Echoコンテキスト作成
	c := e.NewContext(req, rec)

	err := c.Bind(&request)
	assert.NoErrorf(t, err, "Failed to bind request body: %v", err)

	ctxWithValue := context.WithValue(c.Request().Context(), "request", request)
	c.SetRequest(c.Request().WithContext(ctxWithValue))

	// Do Something

	// contextから値を確認
	ctx := c.Request().Context()
	requestFromCtx, ok := ctx.Value("request").(jobs.ReportDateInfo)

	assert.True(t, ok, "Context should contain request value of correct type")

	// contextに保存された値が元のリクエストと一致するかを確認
	expectedRequest := jobs.ReportDateInfo{
		Year:    "2024",
		Quarter: "QUARTER_FOURTH",
	}

	assert.Equal(t, expectedRequest, requestFromCtx)
}

// ctx.Bind の失敗自体は *echo.HTTPError(400)を返すため、errors.Wrap で包んでも
// errors.As がラップを辿って検出でき、この5箇所は実際には「空の200」を起こしていなかった
// (実測で確認済み。この点は当初MIR-1133のチケット記載を誤って適用していた)。実際に
// 空の200を広く引き起こしていたのは同ファイルの executeJob 側(jobFunc失敗・本番ガード)。
// この5箇所を factory.ErrorParameter.New(err) に置き換えたのは、ctx.Bind失敗の扱いとして
// リポジトリ全体で確立している慣習(294箇所)にexecute.goだけが従っていなかったのを揃える
// ためで、レスポンス本文の形が変わる(トップレベルの message から EchoApiResult.error.details
// へ)。Content-Type不正時のステータスも415から400に変わる(echo.HTTPErrorのコードを
// 引き継がなくなるため)。これはリポジトリの他294箇所と同じ挙動であり、退行ではない。
func Test_ExecuteApiControllers_BindFailure(t *testing.T) {
	controller := ExecuteApiControllers{}

	tests := []struct {
		name    string
		handler func(echo.Context) error
	}{
		{"投資家別残高報告書ジョブ: 不正なJSONは400のfactory.Errorになる", controller.CreateInvestorTransactionBalanceReportJob},
		{"投資家別支払調書ジョブ: 不正なJSONは400のfactory.Errorになる", controller.CreateInvestorPaymentStatementJob},
		{"税務署提出用支払調書ジョブ: 不正なJSONは400のfactory.Errorになる", controller.CreateTaxOfficePaymentStatementJob},
		{"税務署提出用支払調書合計ジョブ: 不正なJSONは400のfactory.Errorになる", controller.CreateTaxOfficePaymentStatementSumJob},
		{"抽選CSVアップロードジョブ: 不正なJSONは400のfactory.Errorになる", controller.LotteryCSVUploadJob},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/execute", strings.NewReader("{invalid json"))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := tt.handler(c)

			require.Error(t, err)
			var apiError *factory.Error
			require.ErrorAs(t, err, &apiError)
			assert.Equal(t, http.StatusBadRequest, apiError.Status())
		})
	}
}

// jobFuncが既に*factory.Errorを返している場合、executeJobが一律ErrorSystem(500)で
// 包み直すと元のステータス・コードが失われる(errors.Asは外側から先に一致するため)。
// report_jobsのErrorNotFoundReport(404)のような、意図した業務エラーが500に化けない
// ことを固定する。
func Test_ExecuteJob_PreservesFactoryError(t *testing.T) {
	tests := []struct {
		name       string
		jobFunc    func(context.Context) error
		wantStatus int
		wantCode   int
	}{
		{
			name: "jobFuncがfactory.Errorを返した場合はそのステータス・コードを維持する",
			jobFunc: func(context.Context) error {
				return factory.ErrorNotFoundReport.New(nil)
			},
			wantStatus: http.StatusNotFound,
			wantCode:   factory.ErrorNotFoundReport.Code,
		},
		{
			name: "jobFuncが素のエラーを返した場合はErrorSystem(500)として返す",
			jobFunc: func(context.Context) error {
				return errors.New("gmo api error")
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   factory.ErrorSystem.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/execute", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := executeJob(c, tt.jobFunc)

			require.Error(t, err)
			var apiError *factory.Error
			require.ErrorAs(t, err, &apiError)
			assert.Equal(t, tt.wantStatus, apiError.Status())
			assert.Equal(t, tt.wantCode, apiError.Code)
		})
	}
}
