package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// 実際のテストメソッドはGetAllですが、テスト用にGetAllPaymentStatementと名付けました
func Test_reportOutputRepository_GetAllPaymentStatement(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.ReportOutputQueryOption
	}
	tests := []struct {
		name string
		args args
		// want      []domain.ReportOutput
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "(成功) 投資家用支払調書照会(GetAllPaymentStatement)",
			args: args{
				ctx: context.Background(),
				options: []cdomain.ReportOutputQueryOption{
					cdomain.ReportOutputQueryOptionFunc().WithSubmissionType(domain.SUBMISSION_INVESTORS),
					cdomain.ReportOutputQueryOptionFunc().WithReportTypes(domain.REPORT_PAYMENT_STATEMENT),
					cdomain.ReportOutputQueryOptionFunc().WithGroupBy("year"),
					cdomain.ReportOutputQueryOptionFunc().WithSortBy(""),
					cdomain.ReportOutputQueryOptionFunc().WithOrder(""),
					cdomain.ReportOutputQueryOptionFunc().WithOffset(0),
					cdomain.ReportOutputQueryOptionFunc().WithLimit(30),
				},
			},
			wantTotal: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := reportOutputRepository{
				dbFactory: dbFactoryMock{},
			}
			count, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			assert.NoErrorf(t, err, "reportOutputRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.wantTotal, count, "reportOutputRepository.GetAll() total = %v, want %v", count, tt.wantTotal)
			for _, report := range got {
				assert.Equal(t, domain.SUBMISSION_INVESTORS, report.SubmissionType, "SubmissionType mismatch")
				assert.Equal(t, domain.REPORT_PAYMENT_STATEMENT, report.ReportType, "ReportType mismatch")
			}
		})
	}
}

// 実際のテストメソッドはGetAllですが、テスト用にGetAllTransactionBalanceReportと名付けました
func Test_reportOutputRepository_GetAllTransactionBalanceReport(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.ReportOutputQueryOption
	}
	tests := []struct {
		name string
		args args
		// want      []domain.ReportOutput
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "(成功) 投資家用取引残高報告書照会(GetAllTransactionBalanceReport)",
			args: args{
				ctx: context.Background(),
				options: []cdomain.ReportOutputQueryOption{
					cdomain.ReportOutputQueryOptionFunc().WithSubmissionType(domain.SUBMISSION_INVESTORS),
					cdomain.ReportOutputQueryOptionFunc().WithReportTypes(domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT),
					cdomain.ReportOutputQueryOptionFunc().WithGroupBy("year,quarter"),
					cdomain.ReportOutputQueryOptionFunc().WithSortBy(""),
					cdomain.ReportOutputQueryOptionFunc().WithOrder(""),
					cdomain.ReportOutputQueryOptionFunc().WithOffset(0),
					cdomain.ReportOutputQueryOptionFunc().WithLimit(30),
				},
			},
			wantTotal: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := reportOutputRepository{
				dbFactory: dbFactoryMock{},
			}
			count, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			assert.NoErrorf(t, err, "reportOutputRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.wantTotal, count, "reportOutputRepository.GetAll() total = %v, want %v", count, tt.wantTotal)
			for _, report := range got {
				assert.Equal(t, domain.SUBMISSION_INVESTORS, report.SubmissionType, "SubmissionType mismatch")
				assert.Equal(t, domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT, report.ReportType, "ReportType mismatch")
			}
		})
	}
}

func Test_reportOutputRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.ReportOutput
		wantErr bool
	}{
		{
			name: "(成功) ReportOutput ID照会",
			args: args{
				ctx: context.Background(),
				id:  mockReportPaymentStatementOutput1.Id,
			},
			want:    mockReportPaymentStatementOutput1,
			wantErr: false,
		},
		{
			name: "(失敗) 存在しないReportOutput ID 9999照会",
			args: args{
				ctx: context.Background(),
				id:  9999,
			},
			want:    domain.ReportOutput{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := reportOutputRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetById(tt.args.ctx, tt.args.id)

			ignoreType := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "reportOutputRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreType), "reportOutputRepository.GetById() = %v, want %v", got, tt.want)
		})
	}
}

func Test_reportOutputRepository_Update(t *testing.T) {

	// IsOpenのみが異なるモックデータを作成
	mockReportPaymentStatementOutputOpened := mockReportPaymentStatementOutputNotOpened
	mockReportPaymentStatementOutputOpened.IsOpen = true

	type args struct {
		ctx          context.Context
		reportOutput domain.ReportOutput
	}
	tests := []struct {
		name       string
		args       args
		wantIsOpen bool
		wantErr    bool
	}{
		{
			name: "(成功) ReportOutput更新",

			args: args{
				ctx:          context.Background(),
				reportOutput: mockReportPaymentStatementOutputOpened,
			},
			wantIsOpen: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			session, _ := dbFactoryMock{}.EnergySharesDB(tt.args.ctx)
			ctx := context.WithValue(tt.args.ctx, factory.ConextDBType, session)

			repo := reportOutputRepository{
				dbFactory: factory.NewDBFactory(),
			}

			// 更新前照会
			result, err := repo.GetById(ctx, mockReportPaymentStatementOutputNotOpened.Id)
			assert.Equal(t, mockReportPaymentStatementOutputNotOpened.IsOpen, result.IsOpen, "IsOpen mismatch")

			// 更新
			err = repo.Update(ctx, tt.args.reportOutput)
			assert.NoErrorf(t, err, "reportOutputRepository.Update() error = %v, wantErr %v", err, tt.wantErr)

			// 更新後照会
			result, err = repo.GetById(ctx, mockReportPaymentStatementOutputNotOpened.Id)
			assert.Equal(t, mockReportPaymentStatementOutputOpened.IsOpen, result.IsOpen, "IsOpen mismatch")

		})
	}
}
