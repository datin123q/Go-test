package investor_report

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

func Test_investorReportUsecase_GetAllPaymentStatement(t *testing.T) {

	type args struct {
		ctx     context.Context
		request adomain.InvestorReportGetAllRequest
	}
	tests := []struct {
		name                   string
		args                   args
		s3pkg                  func() *mocks.S3
		reportOutputRepository func() *mocks.ReportOutputRepository
		want                   adomain.InvestorReportGetAllResponse
		wantErr                bool
	}{
		{
			name: "投資家用支払調書照会が正常に呼び出し",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestorReportGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",

					SearchInput: cdomain.SearchInput{
						SkipCount:      0,
						MaxResultCount: 10,
					},
				},
			},
			reportOutputRepository: func() *mocks.ReportOutputRepository {
				repo := new(mocks.ReportOutputRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.SubmissionType() == string(domain.SUBMISSION_INVESTORS)
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return reflect.DeepEqual(opt.ReportTypes(), []interface{}{string(domain.REPORT_PAYMENT_STATEMENT)})
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.GroupBy() == "year"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == ""
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Order() == ""
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}),
				).Return(int64(1), []domain.ReportOutput{
					mockReportPaymentStatementOutput1,
				}, nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3pkg := new(mocks.S3)
				s3pkg.On("CreateSignedUrl", context.Background(), mockReportPaymentStatementOutput1.FileUrl).
					Return(mockReportPaymentStatementOutput1.FileUrl, nil)
				return s3pkg
			},

			want: adomain.InvestorReportGetAllResponse{
				TotalCount: 1,
				Items: []domain.ReportOutput{
					mockReportPaymentStatementOutput1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := investorReportUsecase{
				s3pkg:                  tt.s3pkg(),
				reportOutputRepository: tt.reportOutputRepository(),
			}

			got, err := f.GetAllPaymentStatement(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "GetAllPaymentStatement() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetAllPaymentStatement() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investorReportUsecase_GetAllTransactionBalanceReport(t *testing.T) {

	type args struct {
		ctx     context.Context
		request adomain.InvestorReportGetAllRequest
	}
	tests := []struct {
		name                   string
		args                   args
		s3pkg                  func() *mocks.S3
		reportOutputRepository func() *mocks.ReportOutputRepository
		want                   adomain.InvestorReportGetAllResponse
		wantErr                bool
	}{
		{
			name: "投資家用取引残高報告書照会が正常に呼び出し",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestorReportGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",

					SearchInput: cdomain.SearchInput{
						SkipCount:      0,
						MaxResultCount: 10,
					},
				},
			},
			reportOutputRepository: func() *mocks.ReportOutputRepository {
				repo := new(mocks.ReportOutputRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.SubmissionType() == string(domain.SUBMISSION_INVESTORS)
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return reflect.DeepEqual(opt.ReportTypes(), []interface{}{string(domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT)})
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.GroupBy() == "year,quarter"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == ""
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Order() == ""
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}),
				).Return(int64(1), []domain.ReportOutput{
					mockReportTransactionBalanceReportOutput1,
				}, nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3pkg := new(mocks.S3)
				s3pkg.On("CreateSignedUrl", context.Background(), mockReportPaymentStatementOutput1.FileUrl).
					Return(mockReportTransactionBalanceReportOutput1.FileUrl, nil)
				return s3pkg
			},

			want: adomain.InvestorReportGetAllResponse{
				TotalCount: 1,
				Items: []domain.ReportOutput{
					mockReportTransactionBalanceReportOutput1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := investorReportUsecase{
				s3pkg:                  tt.s3pkg(),
				reportOutputRepository: tt.reportOutputRepository(),
			}

			got, err := f.GetAllTransactionBalanceReport(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "GetAllTransactionBalanceReport() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetAllTransactionBalanceReport() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investorReportUsecase_OpenPaymentStatement(t *testing.T) {

	type args struct {
		ctx     context.Context
		request adomain.InvestorReportOpenPaymentStatementRequest
	}
	tests := []struct {
		name                   string
		args                   args
		reportOutputRepository func() *mocks.ReportOutputRepository
		want                   error
	}{
		{
			name: "投資家用支払調書公開(成功)",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestorReportOpenPaymentStatementRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					Year:     "2024",
				},
			},
			reportOutputRepository: func() *mocks.ReportOutputRepository {
				repo := new(mocks.ReportOutputRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.SubmissionType() == string(domain.SUBMISSION_INVESTORS)
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return reflect.DeepEqual(opt.ReportTypes(), []interface{}{string(domain.REPORT_PAYMENT_STATEMENT)})
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Year() == "2024"
					}),
				).Return(int64(1), []domain.ReportOutput{
					mockReportPaymentStatementOutputOpened,
					mockReportPaymentStatementOutputNotOpened,
				}, nil)

				mockUpdated := mockReportPaymentStatementOutputNotOpened
				mockUpdated.IsOpen = true
				repo.On("Update", context.Background(), mockUpdated).Return(nil)

				return repo
			},

			want: nil,
		},
		{
			name: "投資家用支払調書公開(成功) すでに全て公開されている",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestorReportOpenPaymentStatementRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					Year:     "2024",
				},
			},
			reportOutputRepository: func() *mocks.ReportOutputRepository {
				repo := new(mocks.ReportOutputRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.SubmissionType() == string(domain.SUBMISSION_INVESTORS)
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return reflect.DeepEqual(opt.ReportTypes(), []interface{}{string(domain.REPORT_PAYMENT_STATEMENT)})
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Year() == "2024"
					}),
				).Return(int64(1), []domain.ReportOutput{
					mockReportPaymentStatementOutputOpened,
				}, nil)
				return repo
			},

			want: cdomain.ErrAlreadyOpenReport,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := investorReportUsecase{
				reportOutputRepository: tt.reportOutputRepository(),
			}

			err := f.OpenPaymentStatement(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.want, err, "OpenPaymentStatement() = %v, want %v", err, tt.want)
		})
	}
}

func Test_investorReportUsecase_OpenTransactionBalanceReport(t *testing.T) {

	type args struct {
		ctx     context.Context
		request adomain.InvestorReportOpenTransactionBalanceRequest
	}
	tests := []struct {
		name                   string
		args                   args
		reportOutputRepository func() *mocks.ReportOutputRepository
		want                   error
	}{
		{
			name: "投資家用取引残高報告書公開(成功)",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestorReportOpenTransactionBalanceRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					Year:     "2024",
					Quarter:  "QUARTER_FIRST",
				},
			},
			reportOutputRepository: func() *mocks.ReportOutputRepository {
				repo := new(mocks.ReportOutputRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.SubmissionType() == string(domain.SUBMISSION_INVESTORS)
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return reflect.DeepEqual(opt.ReportTypes(),
							[]interface{}{
								string(domain.REPORT_TRANSACTION_BALANCE_REPORT),
								string(domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT),
							})
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Year() == "2024"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Quarter() == "QUARTER_FIRST"
					}),
				).Return(int64(1), []domain.ReportOutput{
					mockReportTransactionBalanceReportOpened,
					mockReportTransactionBalanceReportNotOpened,
				}, nil)

				mockUpdated := mockReportTransactionBalanceReportNotOpened
				mockUpdated.IsOpen = true
				repo.On("Update", context.Background(), mockUpdated).Return(nil)

				return repo
			},

			want: nil,
		},
		{
			name: "投資家用取引残高報告書公開(成功) すでに全て公開されている",
			args: args{
				ctx: context.Background(),
				request: adomain.InvestorReportOpenTransactionBalanceRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					Year:     "2024",
					Quarter:  "QUARTER_FIRST",
				},
			},
			reportOutputRepository: func() *mocks.ReportOutputRepository {
				repo := new(mocks.ReportOutputRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.SubmissionType() == string(domain.SUBMISSION_INVESTORS)
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return reflect.DeepEqual(opt.ReportTypes(),
							[]interface{}{
								string(domain.REPORT_TRANSACTION_BALANCE_REPORT),
								string(domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT),
							})
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Year() == "2024"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Quarter() == "QUARTER_FIRST"
					}),
				).Return(int64(1), []domain.ReportOutput{
					mockReportTransactionBalanceReportOpened,
				}, nil)
				return repo
			},

			want: cdomain.ErrAlreadyOpenReport,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := investorReportUsecase{
				reportOutputRepository: tt.reportOutputRepository(),
			}

			err := f.OpenTransactionBalanceReport(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.want, err, "OpenTransactionBalanceReport() = %v, want %v", err, tt.want)
		})
	}
}

func Test_investorReportUsecase_DownloadTransactionBalanceReport(t *testing.T) {

	type args struct {
		ctx     context.Context
		request adomain.DownloadTransactionBalanceReportRequest
	}
	tests := []struct {
		name                   string
		args                   args
		s3pkg                  func() *mocks.S3
		reportOutputRepository func() *mocks.ReportOutputRepository
		want                   adomain.DownloadTransactionBalanceReportResponse
		wantErr                bool
	}{
		{
			name: "投資家用取引残高報告書ダウンロードリンク要求(成功)",
			args: args{
				ctx: context.Background(),
				request: adomain.DownloadTransactionBalanceReportRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					Year:     "2024",
					Quarter:  "QUARTER_FIRST",
				},
			},
			reportOutputRepository: func() *mocks.ReportOutputRepository {
				repo := new(mocks.ReportOutputRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.SubmissionType() == string(domain.SUBMISSION_INVESTORS)
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return reflect.DeepEqual(opt.ReportTypes(), []interface{}{string(domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT)})
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Year() == "2024"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Quarter() == "QUARTER_FIRST"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 1
					}),
				).Return(int64(1), []domain.ReportOutput{
					mockReportTransactionBalanceReportOutput4,
				}, nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3pkg := new(mocks.S3)
				s3pkg.On("CreateSignedUrl", context.Background(), mockReportTransactionBalanceReportOutput4.FileUrl).
					Return(mockReportTransactionBalanceReportOutput4.FileUrl, nil)
				return s3pkg
			},

			want: adomain.DownloadTransactionBalanceReportResponse{
				Url: mockReportTransactionBalanceReportOutput4.FileUrl,
			},
			wantErr: false,
		},
		{
			name: "投資家用取引残高報告書ダウンロードリンク要求(失敗) - 投資家用取引残高報告書がない",
			args: args{
				ctx: context.Background(),
				request: adomain.DownloadTransactionBalanceReportRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					Year:     "2024",
					Quarter:  "QUARTER_FIRST",
				},
			},
			reportOutputRepository: func() *mocks.ReportOutputRepository {
				repo := new(mocks.ReportOutputRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.SubmissionType() == string(domain.SUBMISSION_INVESTORS)
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return reflect.DeepEqual(opt.ReportTypes(), []interface{}{string(domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT)})
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Year() == "2024"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Quarter() == "QUARTER_FIRST"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 1
					}),
				).Return(int64(1), []domain.ReportOutput{}, cdomain.ErrItemNotFound)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3pkg := new(mocks.S3)
				s3pkg.On("CreateSignedUrl", context.Background(), mockReportPaymentStatementOutput4.FileUrl).
					Return(mockReportPaymentStatementOutput4.FileUrl, nil)
				return s3pkg
			},

			want:    adomain.DownloadTransactionBalanceReportResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := investorReportUsecase{
				s3pkg:                  tt.s3pkg(),
				reportOutputRepository: tt.reportOutputRepository(),
			}

			got, err := f.DownloadTransactionBalanceReport(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "DownloadTransactionBalanceReport() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "DownloadTransactionBalanceReport() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investorReportUsecase_DownloadPaymentStatement(t *testing.T) {

	type args struct {
		ctx     context.Context
		request adomain.DownloadPaymentStatementRequest
	}
	tests := []struct {
		name                   string
		args                   args
		s3pkg                  func() *mocks.S3
		reportOutputRepository func() *mocks.ReportOutputRepository
		want                   adomain.DownloadPaymentStatementResponse
		wantErr                bool
	}{
		{
			name: "投資家用支払調書ダウンロードリンク要求(成功)",
			args: args{
				ctx: context.Background(),
				request: adomain.DownloadPaymentStatementRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					Year:     "2024",
				},
			},
			reportOutputRepository: func() *mocks.ReportOutputRepository {
				repo := new(mocks.ReportOutputRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.SubmissionType() == string(domain.SUBMISSION_INVESTORS)
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return reflect.DeepEqual(opt.ReportTypes(), []interface{}{string(domain.REPORT_PAYMENT_STATEMENT)})

					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Year() == "2024"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 1
					}),
				).Return(int64(1), []domain.ReportOutput{
					mockReportPaymentStatementOutput4,
				}, nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3pkg := new(mocks.S3)
				s3pkg.On("CreateSignedUrl", context.Background(), mockReportPaymentStatementOutput4.FileUrl).
					Return(mockReportPaymentStatementOutput4.FileUrl, nil)
				return s3pkg
			},

			want: adomain.DownloadPaymentStatementResponse{
				Url: mockReportPaymentStatementOutput4.FileUrl,
			},
			wantErr: false,
		},
		{
			name: "投資家用支払調書ダウンロードリンク要求(失敗) - 投資家用支払調書がない",
			args: args{
				ctx: context.Background(),
				request: adomain.DownloadPaymentStatementRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					Year:     "2024",
				},
			},
			reportOutputRepository: func() *mocks.ReportOutputRepository {
				repo := new(mocks.ReportOutputRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(fields *cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.SubmissionType() == string(domain.SUBMISSION_INVESTORS)
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return reflect.DeepEqual(opt.ReportTypes(), []interface{}{string(domain.REPORT_PAYMENT_STATEMENT)})
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return *opt.Year() == "2024"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc"
					}),
					mock.MatchedBy(func(field func(*cdomain.ReportOutputQueryOptionFields)) bool {
						opt := cdomain.ReportOutputQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 1
					}),
				).Return(int64(1), []domain.ReportOutput{}, cdomain.ErrItemNotFound)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3pkg := new(mocks.S3)
				s3pkg.On("CreateSignedUrl", context.Background(), mockReportPaymentStatementOutput4.FileUrl).
					Return(mockReportPaymentStatementOutput4.FileUrl, nil)
				return s3pkg
			},

			want:    adomain.DownloadPaymentStatementResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := investorReportUsecase{
				s3pkg:                  tt.s3pkg(),
				reportOutputRepository: tt.reportOutputRepository(),
			}

			got, err := f.DownloadPaymentStatement(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "DownloadPaymentStatement() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "DownloadPaymentStatement() = %v, want %v", got, tt.want)
		})
	}
}
