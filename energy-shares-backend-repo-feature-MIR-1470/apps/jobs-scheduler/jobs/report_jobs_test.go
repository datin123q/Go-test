package jobs

import (
	"context"
	"errors"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestReportDateInfo_FromContext(t *testing.T) {
	t.Run("context 값이 비어있을 때", func(t *testing.T) {
		got := ReportDateInfo{}.FromContext(nil)
		assert.Equalf(t, got.IsEmpty(), true, "got: %v", got)
	})

	t.Run("context에 값이 없을 때", func(t *testing.T) {
		got := ReportDateInfo{}.FromContext(context.Background())
		assert.Equalf(t, got.IsEmpty(), true, "got: %v", got)
	})

	t.Run("context 값이 비어있지 않을 때", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request", ReportDateInfo{
			Year:    "2024",
			Quarter: "QUARTER_FOURTH",
		})

		got := ReportDateInfo{}.FromContext(ctx)
		assert.Equalf(t, got.IsEmpty(), false, "got: %v", got)
	})
}

func TestTransactionBalanceReportErrorFields(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantErrorType string
		wantErrorCode string
	}{
		{
			name:          "PDF_ERROR",
			err:           &domain.ReportServiceError{ErrorCode: "s003", Message: "normalize pdf text"},
			wantErrorType: "PDF_ERROR",
			wantErrorCode: "s003",
		},
		{
			name:          "S3_ERROR",
			err:           &domain.ReportServiceError{ErrorCode: "s002", Message: "fail s3 upload"},
			wantErrorType: "S3_ERROR",
			wantErrorCode: "s002",
		},
		{
			name:          "NOT_FOUND",
			err:           factory.ErrorNotFoundReport.New(errors.New("s001:not found report data")),
			wantErrorType: "NOT_FOUND",
		},
		{
			name:          "wrapped HTTP_ERROR",
			err:           pkgerrors.Wrap(errors.New("fail http transaction balance report status=500 error_code=s003 error=pdf error"), "fail request transaction balance report"),
			wantErrorType: "HTTP_ERROR",
		},
		{
			name:          "wrapped JSON_PARSE_ERROR",
			err:           pkgerrors.Wrap(errors.New("fail http transaction balance report Unmarshal"), "fail request transaction balance report"),
			wantErrorType: "JSON_PARSE_ERROR",
		},
		{
			name:          "DB_INSERT_ERROR",
			err:           errors.New("Insert report_output failed"),
			wantErrorType: "DB_INSERT_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transactionBalanceReportErrorFields(tt.err)
			assert.Equal(t, tt.wantErrorType, got["error_type"])
			if tt.wantErrorCode != "" {
				assert.Equal(t, tt.wantErrorCode, got["error_code"])
			}
		})
	}
}

func TestIsNotFoundReportError(t *testing.T) {
	assert.True(t, isNotFoundReportError(factory.ErrorNotFoundReport.New(errors.New("s001:not found report data"))))
	assert.False(t, isNotFoundReportError(&domain.ReportServiceError{ErrorCode: "s003", Message: "pdf error"}))
}
