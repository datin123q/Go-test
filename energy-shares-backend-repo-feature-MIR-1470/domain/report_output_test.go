package domain

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func Test_isNewerReportOutput(t *testing.T) {
	t1 := jsonutils.JSONTime(time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC))
	t2 := jsonutils.JSONTime(time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name      string
		current   ReportOutput
		candidate ReportOutput
		want      bool
	}{
		{
			name:      "candidateのCreatedが新しければtrue",
			current:   ReportOutput{Id: 1, Created: t1},
			candidate: ReportOutput{Id: 2, Created: t2},
			want:      true,
		},
		{
			name:      "candidateのCreatedが古ければfalse",
			current:   ReportOutput{Id: 1, Created: t2},
			candidate: ReportOutput{Id: 2, Created: t1},
			want:      false,
		},
		{
			name:      "Createdが同一でcandidateのIdが大きければtrue",
			current:   ReportOutput{Id: 1, Created: t1},
			candidate: ReportOutput{Id: 2, Created: t1},
			want:      true,
		},
		{
			name:      "Createdが同一でcandidateのIdが小さければfalse",
			current:   ReportOutput{Id: 2, Created: t1},
			candidate: ReportOutput{Id: 1, Created: t1},
			want:      false,
		},
		{
			name:      "完全に同一ならfalse",
			current:   ReportOutput{Id: 1, Created: t1},
			candidate: ReportOutput{Id: 1, Created: t1},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNewerReportOutput(tt.current, tt.candidate)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_latestReportOutputIDsByUser(t *testing.T) {
	t1 := jsonutils.JSONTime(time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC))
	t2 := jsonutils.JSONTime(time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC))
	t3 := jsonutils.JSONTime(time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name  string
		items []ReportOutput
		want  map[int64]int64
	}{
		{
			name:  "空スライスは空mapを返す",
			items: []ReportOutput{},
			want:  map[int64]int64{},
		},
		{
			name: "1件のみの場合はそのIDを返す",
			items: []ReportOutput{
				{Id: 100, UserId: 1, Created: t1},
			},
			want: map[int64]int64{
				1: 100,
			},
		},
		{
			name: "user_idごとにcreatedが最新の1件を選ぶ",
			items: []ReportOutput{
				{Id: 10, UserId: 1, Created: t1},
				{Id: 11, UserId: 1, Created: t2},
				{Id: 20, UserId: 2, Created: t2},
				{Id: 21, UserId: 2, Created: t1},
			},
			want: map[int64]int64{
				1: 11,
				2: 20,
			},
		},
		{
			name: "createdが同時刻ならidが大きい方を選ぶ",
			items: []ReportOutput{
				{Id: 30, UserId: 3, Created: t2},
				{Id: 31, UserId: 3, Created: t2},
			},
			want: map[int64]int64{
				3: 31,
			},
		},
		{
			name: "3件以上でも最新の1件を正しく選ぶ",
			items: []ReportOutput{
				{Id: 40, UserId: 1, Created: t1},
				{Id: 41, UserId: 1, Created: t3},
				{Id: 42, UserId: 1, Created: t2},
			},
			want: map[int64]int64{
				1: 41,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := latestReportOutputIDsByUser(tt.items)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSearchErrClassifiesReportServiceError(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		err := searchErr("s001", "not found report data")
		var apiErr *factory.Error
		assert.True(t, pkgerrors.As(err, &apiErr))
		assert.Equal(t, factory.ErrorNotFoundReport.Code, apiErr.Code)
	})

	t.Run("pdf error", func(t *testing.T) {
		err := searchErr("s003", "normalize pdf text")
		var reportErr *ReportServiceError
		assert.True(t, errors.As(err, &reportErr))
		assert.Equal(t, "s003", reportErr.ErrorCode)
		assert.True(t, reportErr.IsPDFError())
	})

	t.Run("s3 error", func(t *testing.T) {
		err := searchErr("s002", "fail s3 upload")
		var reportErr *ReportServiceError
		assert.True(t, errors.As(err, &reportErr))
		assert.Equal(t, "s002", reportErr.ErrorCode)
		assert.True(t, reportErr.IsS3Error())
	})
}

func TestRequestReportServiceReturnsHTTPStatusError(t *testing.T) {
	originalURL := serviceutils.GetStatutoryBooksServiceUrl()
	originalClient := reportServiceHTTPClient
	t.Cleanup(func() {
		serviceutils.SetStatutoryBooksServiceUrl(originalURL)
		reportServiceHTTPClient = originalClient
	})
	serviceutils.SetStatutoryBooksServiceUrl("http://statutory-books.example")
	reportServiceHTTPClient = http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`{"success":false,"error":"pdf error","error_code":"s003"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	_, err := RequestReportService("/api/download-quarterly-balance-report", url.Values{})

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "status=500")
		assert.Contains(t, err.Error(), "error_code=s003")
	}
}

func TestRequestReportServiceReturnsHTTPStatusErrorWithNonJSONBody(t *testing.T) {
	originalURL := serviceutils.GetStatutoryBooksServiceUrl()
	originalClient := reportServiceHTTPClient
	t.Cleanup(func() {
		serviceutils.SetStatutoryBooksServiceUrl(originalURL)
		reportServiceHTTPClient = originalClient
	})
	serviceutils.SetStatutoryBooksServiceUrl("http://statutory-books.example")
	reportServiceHTTPClient = http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       io.NopCloser(strings.NewReader(`bad gateway`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	_, err := RequestReportService("/api/download-quarterly-balance-report", url.Values{})

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "status=502")
		assert.Contains(t, err.Error(), "bad gateway")
	}
}

func TestRequestReportServiceReturnsJSONParseError(t *testing.T) {
	originalURL := serviceutils.GetStatutoryBooksServiceUrl()
	originalClient := reportServiceHTTPClient
	t.Cleanup(func() {
		serviceutils.SetStatutoryBooksServiceUrl(originalURL)
		reportServiceHTTPClient = originalClient
	})
	serviceutils.SetStatutoryBooksServiceUrl("http://statutory-books.example")
	reportServiceHTTPClient = http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`invalid json`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	_, err := RequestReportService("/api/download-quarterly-balance-report", url.Values{})

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Unmarshal")
	}
}
