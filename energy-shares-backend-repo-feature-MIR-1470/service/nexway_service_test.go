package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/nexway"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setNexwayApplicantsApiはグローバルなnexway設定を書き換える。他のテストへ
// 影響しないよう、t.Cleanupで元の値に戻す。
func setNexwayApplicantsApi(t *testing.T, url string) {
	t.Helper()
	original := nexway.Config{
		Id:                      serviceutils.GetServiceNexwayId(),
		Password:                serviceutils.GetServiceNexwayPwd(),
		FormRegistrationApi:     serviceutils.GetServiceNexwayRegistAPI(),
		FormFileRegistrationApi: serviceutils.GetServiceNexwayFileAPI(),
		FormOrderApi:            serviceutils.GetServiceNexwayOrderAPI(),
		ApplicantsApi:           serviceutils.GetServiceNexwayApplicantsAPI(),
		Token:                   serviceutils.GetServiceNexwayToken(),
	}
	t.Cleanup(func() { serviceutils.SetNexwayConfig(original) })

	serviceutils.SetNexwayConfig(nexway.Config{
		ApplicantsApi: url,
		Token:         "test-token",
	})
}

func TestNexwayService_GetEkycResultByApplicantId(t *testing.T) {
	t.Run("正常レスポンスならEkycResultOutputを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"applicant_id":100,"reference_id":"ekyc-1","result":"match"}`))
		}))
		defer server.Close()
		setNexwayApplicantsApi(t, server.URL)

		s := &NexwayService{}
		result, err := s.GetEkycResultByApplicantId(context.Background(), "100")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(100), result.ApplicantId)
		assert.Equal(t, "ekyc-1", result.ReferenceId)
		assert.Equal(t, "match", result.Result)
	})

	t.Run("HTTPステータス異常ならErrNexwayInvalidResponseを返す(フェイルクローズの判断材料)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("bad gateway"))
		}))
		defer server.Close()
		setNexwayApplicantsApi(t, server.URL)

		s := &NexwayService{}
		_, err := s.GetEkycResultByApplicantId(context.Background(), "100")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNexwayInvalidResponse)
	})

	t.Run("空ボディならErrNexwayInvalidResponseを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		setNexwayApplicantsApi(t, server.URL)

		s := &NexwayService{}
		_, err := s.GetEkycResultByApplicantId(context.Background(), "100")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNexwayInvalidResponse)
	})

	t.Run("不正なJSONならErrNexwayInvalidResponseを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not json"))
		}))
		defer server.Close()
		setNexwayApplicantsApi(t, server.URL)

		s := &NexwayService{}
		_, err := s.GetEkycResultByApplicantId(context.Background(), "100")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNexwayInvalidResponse)
	})

	t.Run("接続自体が失敗したらErrNexwayUnavailableを返す(フェイルオープンの判断材料)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		server.Close() // 起動直後に閉じ、接続できないURLにする

		setNexwayApplicantsApi(t, server.URL)

		s := &NexwayService{}
		_, err := s.GetEkycResultByApplicantId(context.Background(), "100")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNexwayUnavailable)
	})

	t.Run("呼び出し元のctxが既にキャンセルされていればErrNexwayUnavailableを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"applicant_id":100,"reference_id":"ekyc-1","result":"match"}`))
		}))
		defer server.Close()
		setNexwayApplicantsApi(t, server.URL)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		s := &NexwayService{}
		_, err := s.GetEkycResultByApplicantId(ctx, "100")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNexwayUnavailable)
	})

	t.Run("ApplicantsApiが未設定(空URL+制御文字)ならErrNexwayInvalidResponseを返す(設定不備を一時障害と混同しない)", func(t *testing.T) {
		setNexwayApplicantsApi(t, "http://[::1]:namedport/\x7f")

		s := &NexwayService{}
		_, err := s.GetEkycResultByApplicantId(context.Background(), "100")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNexwayInvalidResponse)
	})
}

func TestNexwayService_ErrorSentinels(t *testing.T) {
	// ErrNexwayUnavailable/ErrNexwayInvalidResponseはerrors.Wrapで包んでも
	// errors.Isで判定できる必要がある(呼び出し元のフェイルオープン/
	// フェイルクローズの判断がこれに依存するため)。
	wrapped := pkgerrors.Wrap(ErrNexwayUnavailable, "dial tcp: connection refused")
	assert.ErrorIs(t, wrapped, ErrNexwayUnavailable)
	assert.NotErrorIs(t, wrapped, ErrNexwayInvalidResponse)
}
