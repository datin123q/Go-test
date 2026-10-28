package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const lotteryApi = "/api/v1/lottery"

func TestLotteryHttpDelivery_Cancel(t *testing.T) {
	tests := []struct {
		name           string
		echoCtx        func(e *echo.Echo) echo.Context
		lotteryUsecase func() *mocks.LotteryUseCase
		wantErr        bool
	}{
		{
			name: "(成功) POST /api/v1/lottery/applications/:lottery_application_id/cancel",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   lotteryApi + "/applications/:lottery_application_id/cancel",
				Method: http.MethodPost,
				Params: map[string]string{"lottery_application_id": "app-001"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
			}),
			lotteryUsecase: func() *mocks.LotteryUseCase {
				usecase := &mocks.LotteryUseCase{}
				usecase.On("Cancel", mock.Anything, adomain.LotteryCancelRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "operator@email.com",
						Name:  "オペレーター",
						Roles: map[string]int{"ROLE_OPERATOR": 1},
					},
					LotteryApplicationID: "app-001",
				}).Return(adomain.LotteryCancelResponse{}, nil)
				return usecase
			},
		},
		{
			name: "(失敗) POST /api/v1/lottery/applications/:lottery_application_id/cancel - 対象が存在しない",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   lotteryApi + "/applications/:lottery_application_id/cancel",
				Method: http.MethodPost,
				Params: map[string]string{"lottery_application_id": "app-999"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
			}),
			lotteryUsecase: func() *mocks.LotteryUseCase {
				usecase := &mocks.LotteryUseCase{}
				usecase.On("Cancel", mock.Anything, adomain.LotteryCancelRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "operator@email.com",
						Name:  "オペレーター",
						Roles: map[string]int{"ROLE_OPERATOR": 1},
					},
					LotteryApplicationID: "app-999",
				}).Return(adomain.LotteryCancelResponse{}, cdomain.ErrItemNotFound)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "(失敗) POST /api/v1/lottery/applications/:lottery_application_id/cancel - キャンセル不可",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   lotteryApi + "/applications/:lottery_application_id/cancel",
				Method: http.MethodPost,
				Params: map[string]string{"lottery_application_id": "app-002"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
			}),
			lotteryUsecase: func() *mocks.LotteryUseCase {
				usecase := &mocks.LotteryUseCase{}
				usecase.On("Cancel", mock.Anything, adomain.LotteryCancelRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "operator@email.com",
						Name:  "オペレーター",
						Roles: map[string]int{"ROLE_OPERATOR": 1},
					},
					LotteryApplicationID: "app-002",
				}).Return(adomain.LotteryCancelResponse{}, cdomain.ErrBadRequest)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			delivery := lotteryHttpDelivery{
				lotteryUsecase: tt.lotteryUsecase(),
				extractor:      jwthelper.NewUserExtractor(),
			}

			err := delivery.Cancel(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Cancel() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestLotteryHttpDelivery_UploadCSV(t *testing.T) {
	tests := []struct {
		name           string
		echoCtx        func(e *echo.Echo) echo.Context
		lotteryUsecase func() *mocks.LotteryUseCase
		wantErr        bool
	}{
		{
			name: "(成功) POST /api/v1/lottery/applications/:project_id/csv-upload",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   lotteryApi + "/applications/:project_id/csv-upload",
				Method: http.MethodPost,
				Params: map[string]string{"project_id": "101"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEMultipartForm,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
				Multipart: pkgtest.Multipart{
					MultipartFiles: map[string][]pkgtest.File{
						"file": {{
							Filename: "lottery.csv",
							File:     []byte("抽選ID,抽選ステータス,投資家名,メールアドレス,ファンド名,申込口数,申込金額,当選総額,当選後キャンセル額,投資可能金額,申込日時,当落,当選させたい金額\napp-001,申込中,投資家A,a@example.com,ファンドA,10,100000,0,0,100000,2026/5/1 10:00,落選,0\n"),
						}},
					},
				},
			}),
			lotteryUsecase: func() *mocks.LotteryUseCase {
				usecase := &mocks.LotteryUseCase{}
				usecase.On("UploadCSV", mock.Anything, mock.MatchedBy(func(req adomain.LotteryCSVUploadRequest) bool {
					return req.JWTUser.Id == 1000 &&
						req.JWTUser.Email == "operator@email.com" &&
						req.ProjectID == 101 &&
						req.File != nil &&
						req.File.Filename == "lottery.csv"
				})).Return(adomain.LotteryCSVUploadResponse{
					JobID:  "job-001",
					Status: "QUEUED",
				}, nil)
				return usecase
			},
		},
		{
			name: "(成功) 任意のmultipartフィールド名でも受け付ける",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   lotteryApi + "/applications/:project_id/csv-upload",
				Method: http.MethodPost,
				Params: map[string]string{"project_id": "101"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEMultipartForm,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
				Multipart: pkgtest.Multipart{
					MultipartFiles: map[string][]pkgtest.File{
						"lottery_application_upload_file": {{
							Filename: "lottery_csv_upload_template.csv",
							File:     []byte("抽選ID,抽選ステータス,投資家名,メールアドレス,ファンド名,申込口数,申込金額,当選総額,当選後キャンセル額,投資可能金額,申込日時,当落,当選させたい金額\napp-001,申込中,投資家A,a@example.com,ファンドA,10,100000,0,0,100000,2026/5/1 10:00,落選,0\n"),
						}},
					},
				},
			}),
			lotteryUsecase: func() *mocks.LotteryUseCase {
				usecase := &mocks.LotteryUseCase{}
				usecase.On("UploadCSV", mock.Anything, mock.MatchedBy(func(req adomain.LotteryCSVUploadRequest) bool {
					return req.JWTUser.Id == 1000 &&
						req.JWTUser.Email == "operator@email.com" &&
						req.ProjectID == 101 &&
						req.File != nil &&
						req.File.Filename == "lottery_csv_upload_template.csv"
				})).Return(adomain.LotteryCSVUploadResponse{
					JobID:  "job-002",
					Status: "QUEUED",
				}, nil)
				return usecase
			},
		},
		{
			name: "(失敗) file不足",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   lotteryApi + "/applications/:project_id/csv-upload",
				Method: http.MethodPost,
				Params: map[string]string{"project_id": "101"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEMultipartForm,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
			}),
			lotteryUsecase: func() *mocks.LotteryUseCase {
				return &mocks.LotteryUseCase{}
			},
			wantErr: true,
		},
		{
			name: "(失敗) csv以外の拡張子",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   lotteryApi + "/applications/:project_id/csv-upload",
				Method: http.MethodPost,
				Params: map[string]string{"project_id": "101"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEMultipartForm,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
				Multipart: pkgtest.Multipart{
					MultipartFiles: map[string][]pkgtest.File{
						"lottery_application_upload_file": {{
							Filename: "lottery.xlsx",
							File:     []byte("dummy"),
						}},
					},
				},
			}),
			lotteryUsecase: func() *mocks.LotteryUseCase {
				return &mocks.LotteryUseCase{}
			},
			wantErr: true,
		},
		{
			name: "(失敗) ファイルサイズ上限超過",
			echoCtx: func(e *echo.Echo) echo.Context {
				ctx := pkgtest.RequestContext(pkgtest.Request{
					Path:   lotteryApi + "/applications/:project_id/csv-upload",
					Method: http.MethodPost,
					Params: map[string]string{"project_id": "101"},
					Header: pkgtest.Header{
						Headers:   echo.HeaderContentType,
						MIMETypes: echo.MIMEMultipartForm,
					},
					Claims: map[string]interface{}{
						"user":  "1000",
						"email": "operator@email.com",
						"name":  "オペレーター",
						"role":  []string{"ROLE_OPERATOR"},
					},
				})(e)
				ctx.Request().ContentLength = adomain.LotteryCSVUploadMaxRequestBodyBytes + 1
				return ctx
			},
			lotteryUsecase: func() *mocks.LotteryUseCase {
				return &mocks.LotteryUseCase{}
			},
		},
		{
			name: "(失敗) CSVバリデーションエラーは構造化detailsで返す",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   lotteryApi + "/applications/:project_id/csv-upload",
				Method: http.MethodPost,
				Params: map[string]string{"project_id": "101"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEMultipartForm,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
				Multipart: pkgtest.Multipart{
					MultipartFiles: map[string][]pkgtest.File{
						"file": {{
							Filename: "lottery.csv",
							File:     []byte("dummy"),
						}},
					},
				},
			}),
			lotteryUsecase: func() *mocks.LotteryUseCase {
				usecase := &mocks.LotteryUseCase{}
				usecase.On("UploadCSV", mock.Anything, mock.Anything).Return(
					adomain.LotteryCSVUploadResponse{},
					adomain.NewLotteryCSVValidationError([]adomain.LotteryCSVValidationErrorDetail{
						{Row: 2, Field: "申込日時", Message: "申込日時はyyyy/M/d HH:mm形式で入力してください"},
						{Row: 2, Field: "抽選ステータス", Message: "抽選ステータスが現在の申込情報と一致しません"},
					}),
				)
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			delivery := lotteryHttpDelivery{
				lotteryUsecase: tt.lotteryUsecase(),
				extractor:      jwthelper.NewUserExtractor(),
			}

			ctx := tt.echoCtx(e)
			err := delivery.UploadCSV(ctx)
			assert.Equalf(t, tt.wantErr, err != nil, "UploadCSV() error = %v, wantErr %v", err, tt.wantErr)

			if tt.name == "(失敗) CSVバリデーションエラーは構造化detailsで返す" || tt.name == "(失敗) ファイルサイズ上限超過" {
				assert.NoError(t, err)

				var body struct {
					Result  interface{} `json:"result"`
					Success bool        `json:"success"`
					Error   struct {
						Code    string `json:"code"`
						Details []struct {
							Row     int    `json:"row"`
							Field   string `json:"field"`
							Message string `json:"message"`
						} `json:"details"`
					} `json:"error"`
				}

				rec := ctx.Response().Writer.(*httptest.ResponseRecorder)
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.False(t, body.Success)
				assert.Equal(t, "CSV_VALIDATION_ERROR", body.Error.Code)
				if tt.name == "(失敗) CSVバリデーションエラーは構造化detailsで返す" {
					if assert.Len(t, body.Error.Details, 2) {
						assert.Equal(t, 2, body.Error.Details[0].Row)
						assert.Equal(t, "申込日時", body.Error.Details[0].Field)
						assert.Equal(t, "申込日時はyyyy/M/d HH:mm形式で入力してください", body.Error.Details[0].Message)
					}
				}
				if tt.name == "(失敗) ファイルサイズ上限超過" {
					if assert.Len(t, body.Error.Details, 1) {
						assert.Equal(t, 0, body.Error.Details[0].Row)
						assert.Equal(t, "file", body.Error.Details[0].Field)
						assert.Equal(t, "CSVファイルサイズは10MB以下にしてください", body.Error.Details[0].Message)
					}
				}
			}
		})
	}
}
