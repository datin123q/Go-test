package http

import (
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/actorcontext"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type lotteryHttpDelivery struct {
	lotteryUsecase adomain.LotteryUseCase

	extractor jwthelper.UserExtractor
}

type lotteryCSVValidationErrorResponse struct {
	Result  interface{} `json:"result"`
	Success bool        `json:"success"`
	Error   struct {
		Code    string                                    `json:"code"`
		Details []adomain.LotteryCSVValidationErrorDetail `json:"details"`
	} `json:"error"`
}

func renderLotteryCSVValidationError(c echo.Context, details []adomain.LotteryCSVValidationErrorDetail) error {
	response := lotteryCSVValidationErrorResponse{
		Result:  nil,
		Success: false,
	}
	response.Error.Code = "CSV_VALIDATION_ERROR"
	response.Error.Details = details
	return c.JSON(http.StatusBadRequest, response)
}

func isMaxBytesError(err error) bool {
	if err == nil {
		return false
	}
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

// projectはControllerに置かれており古いのでLotteryで作成
// projectを新しくしたらlotteryはprojectに統合する

func New(
	api echoswagger.ApiRoot,
	lotteryUsecase adomain.LotteryUseCase,
	extractor jwthelper.UserExtractor,
) *lotteryHttpDelivery {
	handler := &lotteryHttpDelivery{
		lotteryUsecase: lotteryUsecase,
		extractor:      extractor,
	}

	g := api.Group("v1/lottery", "api/v1/lottery",
		echoutils.AllowRoles(domain.ROLE_OPERATOR))

	g.GET("/list",
		handler.GetAll,
	).AddResponse(http.StatusOK, "抽選申込一覧", adomain.LotteryGetAllResponse{}, nil).
		AddParamQueryNested(adomain.LotteryGetAllQueryParam{}).
		SetDescription("抽選申込一覧を取得").
		SetSummary("抽選申込一覧取得")

	g.GET("/list/download",
		handler.Download,
	).AddResponse(http.StatusOK, "抽選申込一覧CSV", adomain.LotteryDownloadResponse{}, nil).
		AddParamQueryNested(adomain.LotteryDownloadQueryParam{}).
		SetDescription("抽選申込一覧をCSV形式でダウンロード").
		SetSummary("抽選申込一覧CSVダウンロード")

	g.POST("/applications/:lottery_application_id/win",
		handler.Win,
	).AddParamPath("", "lottery_application_id", "抽選申込ID").
		AddParamBody(adomain.LotteryWinRequest{}, "body", "手動当選リクエスト", true).
		AddResponse(http.StatusOK, "手動当選成功", struct {
			Result  int  `json:"result" example:"1"`
			Success bool `json:"success" example:"true"`
		}{}, nil).
		SetDescription("抽選申込に対して手動で当選処理を実行").
		SetSummary("抽選申込手動当選")

	g.POST("/applications/:lottery_application_id/lose",
		handler.Lose,
	).AddParamPath("", "lottery_application_id", "抽選申込ID").
		AddParamBody(adomain.LotteryLoseRequest{}, "body", "手動落選リクエスト", true).
		AddResponse(http.StatusOK, "手動落選成功", struct {
			Result  int  `json:"result" example:"1"`
			Success bool `json:"success" example:"true"`
		}{}, nil).
		SetDescription("抽選申込に対して手動で落選処理を実行").
		SetSummary("抽選申込手動落選")

	g.POST("/applications/:lottery_application_id/cancel",
		handler.Cancel,
	).AddParamPath("", "lottery_application_id", "抽選申込ID").
		AddResponse(http.StatusOK, "抽選申込キャンセル", adomain.LotteryCancelResponse{}, nil).
		SetDescription("抽選申込をキャンセル").
		SetSummary("抽選申込キャンセル")

	projectGroup := api.Group("v1/projects", "api/v1/projects",
		echoutils.AllowRoles(domain.ROLE_PRODUCT))

	projectGroup.GET("/:id/lottery-applications/bulk-lottery/download",
		handler.DownloadForBulkLottery,
	).AddResponse(http.StatusOK, "一括抽選用 抽選申込CSV", adomain.LotteryDownloadResponse{}, nil).
		AddParamPath("", "id", "プロジェクトID").
		AddParamQueryNested(adomain.LotteryDownloadForBulkLotteryQueryParam{}).
		SetDescription("ファンドに紐づく一括抽選用の抽選申込をCSV形式でダウンロード").
		SetSummary("一括抽選用 抽選申込CSVダウンロード")

	g.POST("/applications/:project_id/csv-upload",
		handler.UploadCSV,
	).AddParamPath("", "project_id", "プロジェクトID").
		AddParamFile("file", "抽選CSVファイル", true).
		AddResponse(http.StatusOK, "抽選CSVアップロード受付", adomain.LotteryCSVUploadResponse{}, nil).
		SetRequestContentType(echo.MIMEMultipartForm).
		SetDescription("抽選CSVをアップロードして一括当落処理ジョブを受け付ける").
		SetSummary("抽選CSVアップロード")

	return handler
}

func (l lotteryHttpDelivery) GetAll(c echo.Context) error {
	// JWT から管理画面オペレーター情報を取り出し、一覧取得の実行主体として usecase に渡す。
	user, err := l.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	var request = adomain.LotteryGetAllRequest{
		JWTUser: user,
	}
	// 一覧系は query param の bind のみ行い、詳細な検索条件の解釈は usecase に寄せる。
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := l.lotteryUsecase.GetAll(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (l lotteryHttpDelivery) Download(c echo.Context) error {
	user, err := l.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	request := adomain.LotteryDownloadRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := l.lotteryUsecase.Download(ctx, request)
	if err != nil {
		if errors.Is(err, cdomain.ErrRequestLimitExceeded) {
			return factory.ErrorParameter.New(err)
		}
		return factory.ErrorFailDownloadCSV.New(err)
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename=energyshares_lottery_list.csv")
	return c.Blob(http.StatusOK, "application/csv", append([]byte{0xEF, 0xBB, 0xBF}, response.Data...))
}

func (l lotteryHttpDelivery) DownloadForBulkLottery(c echo.Context) error {
	user, err := l.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	request := adomain.LotteryDownloadForBulkLotteryRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := l.lotteryUsecase.DownloadForBulkLottery(ctx, request)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename=energyshares_lottery_bulk_lottery.csv")
	return c.Blob(http.StatusOK, "application/csv", append([]byte{0xEF, 0xBB, 0xBF}, response.Data...))
}

func (l lotteryHttpDelivery) Win(c echo.Context) error {
	// 手動当選は管理画面操作のため、まず JWT からオペレーター情報を解決する。
	user, err := l.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	request := adomain.LotteryWinRequest{
		JWTUser: user,
	}
	// path/body を 1 つの request DTO に詰めて usecase に渡す。
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	// 形式バリデーションのみ HTTP 層で処理し、業務バリデーションは usecase で行う。
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = l.lotteryUsecase.Win(ctx, request)
	if err != nil {
		// usecase の業務エラーを admin API 共通のレスポンス形式へ変換する。
		switch {
		case errors.Is(err, cdomain.ErrItemNotFound):
			return factory.ErrorNotFound.New(err)
		case errors.Is(err, cdomain.ErrBadRequest):
			return factory.ErrorParameter.New(err)
		default:
			return factory.ErrorDB.New(err)
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, 1)
}

func (l lotteryHttpDelivery) Lose(c echo.Context) error {
	// 手動落選も管理画面操作のため、JWT からオペレーター情報を解決する。
	user, err := l.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	request := adomain.LotteryLoseRequest{
		JWTUser: user,
	}
	// path/body を 1 つの request DTO に詰めて usecase に渡す。
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	// 形式バリデーションのみ HTTP 層で処理し、業務バリデーションは usecase で行う。
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = l.lotteryUsecase.Lose(ctx, request)
	if err != nil {
		// usecase の業務エラーを admin API 共通のレスポンス形式へ変換する。
		switch {
		case errors.Is(err, cdomain.ErrItemNotFound):
			return factory.ErrorNotFound.New(err)
		case errors.Is(err, cdomain.ErrBadRequest):
			return factory.ErrorParameter.New(err)
		default:
			return factory.ErrorDB.New(err)
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, 1)
}

func (l lotteryHttpDelivery) Cancel(c echo.Context) error {
	user, err := l.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := actorcontext.WithActor(c.Request().Context(), user)
	lotteryApplicationID := c.Param("lottery_application_id")

	request := adomain.LotteryCancelRequest{
		JWTUser:              user,
		LotteryApplicationID: lotteryApplicationID,
	}
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := l.lotteryUsecase.Cancel(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func findLotteryCSVUploadFile(c echo.Context) (*multipart.FileHeader, error) {
	file, err := c.FormFile("file")
	if err == nil {
		return file, nil
	}

	form, formErr := c.MultipartForm()
	if formErr != nil {
		return nil, formErr
	}

	var found *multipart.FileHeader
	fileCount := 0
	for _, files := range form.File {
		for _, fh := range files {
			fileCount++
			if found == nil {
				found = fh
			}
		}
	}

	if fileCount == 0 {
		return nil, err
	}
	if fileCount > 1 {
		return nil, errors.Wrap(cdomain.ErrBadRequest, "only one csv file can be uploaded")
	}

	return found, nil
}

func (l lotteryHttpDelivery) UploadCSV(c echo.Context) error {
	user, err := l.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	if c.Request().ContentLength > adomain.LotteryCSVUploadMaxRequestBodyBytes {
		return renderLotteryCSVValidationError(c, []adomain.LotteryCSVValidationErrorDetail{
			{
				Field:   "file",
				Message: "CSVファイルサイズは10MB以下にしてください",
			},
		})
	}
	c.Request().Body = http.MaxBytesReader(
		c.Response().Writer,
		c.Request().Body,
		adomain.LotteryCSVUploadMaxRequestBodyBytes,
	)

	projectID, err := strconv.ParseInt(c.Param("project_id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	file, err := findLotteryCSVUploadFile(c)
	if err != nil {
		if isMaxBytesError(err) {
			return renderLotteryCSVValidationError(c, []adomain.LotteryCSVValidationErrorDetail{
				{
					Field:   "file",
					Message: "CSVファイルサイズは10MB以下にしてください",
				},
			})
		}
		return factory.ErrorParameter.New(err)
	}
	if !strings.EqualFold(filepath.Ext(file.Filename), ".csv") {
		return factory.ErrorParameter.New(errors.Wrap(cdomain.ErrBadRequest, "uploaded file must have .csv extension"))
	}
	if file.Size > adomain.LotteryCSVUploadMaxFileSizeBytes {
		return renderLotteryCSVValidationError(c, []adomain.LotteryCSVValidationErrorDetail{
			{
				Field:   "file",
				Message: "CSVファイルサイズは10MB以下にしてください",
			},
		})
	}

	request := adomain.LotteryCSVUploadRequest{
		JWTUser:   user,
		ProjectID: projectID,
		File:      file,
	}
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := l.lotteryUsecase.UploadCSV(c.Request().Context(), request)
	if err != nil {
		var csvValidationErr *adomain.LotteryCSVValidationError
		switch {
		case errors.As(err, &csvValidationErr):
			return renderLotteryCSVValidationError(c, csvValidationErr.Details)
		case errors.Is(err, cdomain.ErrItemNotFound):
			return factory.ErrorNotFound.New(err)
		case errors.Is(err, cdomain.ErrBadRequest):
			return factory.ErrorParameter.New(err)
		default:
			return factory.ErrorDB.New(err)
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}
