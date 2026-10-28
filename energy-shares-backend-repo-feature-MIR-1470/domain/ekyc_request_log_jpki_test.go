package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"

	_ "github.com/mattn/go-sqlite3"
)

func newEkycJpkiTestDB(t *testing.T, logs ...EkycRequestLog) (context.Context, *xorm.Engine) {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	require.NoError(t, engine.Sync(new(EkycRequestLog)))
	for i := range logs {
		_, err := engine.Insert(&logs[i])
		require.NoError(t, err)
	}
	return context.WithValue(context.Background(), factory.ConextDBType, engine), engine
}

func loadEkycJpkiLog(t *testing.T, engine *xorm.Engine, id int64) EkycRequestLog {
	t.Helper()
	var got EkycRequestLog
	has, err := engine.ID(id).Get(&got)
	require.NoError(t, err)
	require.True(t, has)
	return got
}

func TestEkycRequestLog_MarkFailedIfNone(t *testing.T) {
	tests := []struct {
		name        string
		status      EKYC_REQUEST_STATUS
		wantStatus  EKYC_REQUEST_STATUS
		wantUpdated int64
	}{
		{"結果が未確定(NONE)なら失敗にする", EKYC_REQUEST_STATUS_NONE, EKYC_REQUEST_STATUS_FAIL, 1},
		{"成功済み(SUCCESS)は上書きしない", EKYC_REQUEST_STATUS_SUCCESS, EKYC_REQUEST_STATUS_SUCCESS, 0},
		{"失敗済み(FAIL)はそのまま", EKYC_REQUEST_STATUS_FAIL, EKYC_REQUEST_STATUS_FAIL, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, engine := newEkycJpkiTestDB(t,
				EkycRequestLog{Email: "a@example.com", UserId: 1, EkycId: "ekyc-1", ApplicantId: 77, Status: tt.status, Result: EKYC_RESULT_STATUS_MATCH},
				EkycRequestLog{Email: "b@example.com", UserId: 2, EkycId: "ekyc-2", Status: EKYC_REQUEST_STATUS_NONE},
			)

			target := loadEkycJpkiLog(t, engine, 1)
			updated, err := target.MarkFailedIfNone(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.wantUpdated, updated)

			got := loadEkycJpkiLog(t, engine, 1)
			require.Equal(t, tt.wantStatus, got.Status)
			// status以外の列は変えない(申請ID・結果は残る)。
			require.Equal(t, int64(77), got.ApplicantId)
			require.Equal(t, EKYC_RESULT_STATUS_MATCH, got.Result)

			// 他の行には影響しない。
			other := loadEkycJpkiLog(t, engine, 2)
			require.Equal(t, EKYC_REQUEST_STATUS_NONE, other.Status)
		})
	}
}

func TestEkycRequestLog_SetApplicantIdIfUnset(t *testing.T) {
	tests := []struct {
		name            string
		status          EKYC_REQUEST_STATUS
		applicantId     int64
		wantApplicantId int64
		wantUpdated     int64
	}{
		{"未設定・未確定なら記録する", EKYC_REQUEST_STATUS_NONE, 0, 555, 1},
		{"既に申請IDがあれば上書きしない(結果通知が先に届いた場合)", EKYC_REQUEST_STATUS_NONE, 300, 300, 0},
		// 画面からの完了報告(PUT)が先に届いて行がSUCCESSでも、申請IDが無ければ記録する。状態を条件に
		// すると、申請したのに申請IDが残らない。
		{"完了報告が先に届いた行(SUCCESS)でも、申請IDが無ければ記録する", EKYC_REQUEST_STATUS_SUCCESS, 0, 555, 1},
		{"新しい発行で閉じられた行(FAIL)でも、申請済みなら申請IDを残す", EKYC_REQUEST_STATUS_FAIL, 0, 555, 1},
		{"完了報告済みで申請IDがあれば上書きしない", EKYC_REQUEST_STATUS_SUCCESS, 300, 300, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, engine := newEkycJpkiTestDB(t,
				EkycRequestLog{Email: "a@example.com", UserId: 1, EkycId: "ekyc-1", ApplicantId: tt.applicantId, Status: tt.status},
				EkycRequestLog{Email: "b@example.com", UserId: 2, EkycId: "ekyc-2", Status: EKYC_REQUEST_STATUS_NONE},
			)

			target := loadEkycJpkiLog(t, engine, 1)
			updated, err := target.SetApplicantIdIfUnset(ctx, 555)
			require.NoError(t, err)
			require.Equal(t, tt.wantUpdated, updated)

			got := loadEkycJpkiLog(t, engine, 1)
			require.Equal(t, tt.wantApplicantId, got.ApplicantId)
			require.Equal(t, tt.status, got.Status, "状態は変えない")

			other := loadEkycJpkiLog(t, engine, 2)
			require.Equal(t, int64(0), other.ApplicantId)
		})
	}
}
