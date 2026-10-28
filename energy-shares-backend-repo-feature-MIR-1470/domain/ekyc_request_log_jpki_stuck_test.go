package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jpkiReqIdPtr(s string) *string { return &s }

func TestEkycRequestLog_MarkFailedIfNoneWithMessage(t *testing.T) {
	tests := []struct {
		name        string
		status      EKYC_REQUEST_STATUS
		message     string
		wantStatus  EKYC_REQUEST_STATUS
		wantMessage string
		wantUpdated int64
	}{
		{"未確定なら、失敗にして理由を記録する", EKYC_REQUEST_STATUS_NONE, "C04: 期限切れ\n", EKYC_REQUEST_STATUS_FAIL, "C04: 期限切れ\n", 1},
		{"成功済みは、状態も理由も上書きしない", EKYC_REQUEST_STATUS_SUCCESS, "C04: 期限切れ\n", EKYC_REQUEST_STATUS_SUCCESS, "", 0},
		{"失敗済みは、理由を後から書き換えない", EKYC_REQUEST_STATUS_FAIL, "C04: 期限切れ\n", EKYC_REQUEST_STATUS_FAIL, "", 0},
		{"理由が空なら、状態だけ失敗にする", EKYC_REQUEST_STATUS_NONE, "", EKYC_REQUEST_STATUS_FAIL, "", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, engine := newEkycJpkiTestDB(t,
				EkycRequestLog{Email: "a@example.com", UserId: 1, EkycId: "ekyc-1", Status: tt.status},
				EkycRequestLog{Email: "b@example.com", UserId: 2, EkycId: "ekyc-2", Status: EKYC_REQUEST_STATUS_NONE},
			)

			target := loadEkycJpkiLog(t, engine, 1)
			updated, err := target.MarkFailedIfNoneWithMessage(ctx, tt.message)
			require.NoError(t, err)
			require.Equal(t, tt.wantUpdated, updated)

			got := loadEkycJpkiLog(t, engine, 1)
			assert.Equal(t, tt.wantStatus, got.Status)
			if tt.wantMessage == "" {
				assert.Nil(t, got.ErrorMessage)
			} else {
				require.NotNil(t, got.ErrorMessage)
				assert.Equal(t, tt.wantMessage, *got.ErrorMessage)
			}
			// 別の会員の行には触れない。
			other := loadEkycJpkiLog(t, engine, 2)
			assert.Equal(t, EKYC_REQUEST_STATUS_NONE, other.Status)
			assert.Nil(t, other.ErrorMessage)
		})
	}
}

// 同じ行へ、別のWebhookとジョブが順に失敗を書いても、先に書いた理由が残る。
func TestEkycRequestLog_MarkFailedIfNoneWithMessage_先に書いた理由が残る(t *testing.T) {
	ctx, engine := newEkycJpkiTestDB(t,
		EkycRequestLog{Email: "a@example.com", UserId: 1, EkycId: "ekyc-1", Status: EKYC_REQUEST_STATUS_NONE},
	)
	first := loadEkycJpkiLog(t, engine, 1)
	second := loadEkycJpkiLog(t, engine, 1)

	updated, err := first.MarkFailedIfNoneWithMessage(ctx, "C06: ロック\n")
	require.NoError(t, err)
	require.Equal(t, int64(1), updated)
	updated, err = second.MarkFailedIfNoneWithMessage(ctx, "C01: 拒否\n")
	require.NoError(t, err)
	assert.Equal(t, int64(0), updated, "2回目は書き換えない(通知メールを2通送らないための根拠)")

	got := loadEkycJpkiLog(t, engine, 1)
	require.NotNil(t, got.ErrorMessage)
	assert.Equal(t, "C06: ロック\n", *got.ErrorMessage)
}

func TestFindStuckJpkiRequestLogs(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	old := now.Add(-2 * time.Hour)
	recent := now.Add(-10 * time.Minute)

	type row struct {
		log     EkycRequestLog
		created time.Time
	}
	rows := []row{
		{EkycRequestLog{Email: "1@example.com", UserId: 1, EkycId: "stuck-old", RequestId: jpkiReqIdPtr("req-1"), Status: EKYC_REQUEST_STATUS_NONE}, old},
		{EkycRequestLog{Email: "2@example.com", UserId: 2, EkycId: "stuck-older", RequestId: jpkiReqIdPtr("req-2"), Status: EKYC_REQUEST_STATUS_NONE}, now.Add(-5 * time.Hour)},
		{EkycRequestLog{Email: "3@example.com", UserId: 3, EkycId: "too-recent", RequestId: jpkiReqIdPtr("req-3"), Status: EKYC_REQUEST_STATUS_NONE}, recent},
		{EkycRequestLog{Email: "4@example.com", UserId: 4, EkycId: "ho-method", Status: EKYC_REQUEST_STATUS_NONE}, old},
		{EkycRequestLog{Email: "5@example.com", UserId: 5, EkycId: "empty-request-id", RequestId: jpkiReqIdPtr(""), Status: EKYC_REQUEST_STATUS_NONE}, old},
		{EkycRequestLog{Email: "6@example.com", UserId: 6, EkycId: "already-failed", RequestId: jpkiReqIdPtr("req-6"), Status: EKYC_REQUEST_STATUS_FAIL}, old},
		{EkycRequestLog{Email: "7@example.com", UserId: 7, EkycId: "already-success", RequestId: jpkiReqIdPtr("req-7"), Status: EKYC_REQUEST_STATUS_SUCCESS, ApplicantId: 9}, old},
		{EkycRequestLog{Email: "8@example.com", UserId: 8, EkycId: "bpo-applied", RequestId: jpkiReqIdPtr("req-8"), Status: EKYC_REQUEST_STATUS_NONE, ApplicantId: 55}, old},
	}
	logs := make([]EkycRequestLog, 0, len(rows))
	for _, r := range rows {
		logs = append(logs, r.log)
	}
	ctx, engine := newEkycJpkiTestDB(t, logs...)
	// createdは自動で入るため、行ごとに作成時刻を書き換える(UTCで保存する)。
	for i, r := range rows {
		_, err := engine.Exec("UPDATE ekyc_request_log SET created = ? WHERE id = ?", r.created.UTC().Format("2006-01-02 15:04:05"), i+1)
		require.NoError(t, err)
	}

	got, err := FindStuckJpkiRequestLogs(ctx, now.Add(-1*time.Hour), now.Add(-14*24*time.Hour), 10)
	require.NoError(t, err)

	var ekycIds []string
	for _, l := range got {
		ekycIds = append(ekycIds, l.EkycId)
	}
	// 対象は、検証依頼IDがあり・未確定で・BPOへ未申請で・1時間以上前の行だけ。新しい(Idの大きい)順。
	assert.Equal(t, []string{"stuck-older", "stuck-old"}, ekycIds)
}

func TestFindStuckJpkiRequestLogs_件数の上限と順序(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	var logs []EkycRequestLog
	for i := 1; i <= 5; i++ {
		logs = append(logs, EkycRequestLog{Email: "u@example.com", UserId: int64(i), EkycId: "e-" + string(rune('0'+i)), RequestId: jpkiReqIdPtr("req"), Status: EKYC_REQUEST_STATUS_NONE})
	}
	ctx, engine := newEkycJpkiTestDB(t, logs...)
	_, err := engine.Exec("UPDATE ekyc_request_log SET created = ?", now.Add(-3*time.Hour).Format("2006-01-02 15:04:05"))
	require.NoError(t, err)

	got, err := FindStuckJpkiRequestLogs(ctx, now.Add(-1*time.Hour), now.Add(-14*24*time.Hour), 3)
	require.NoError(t, err)

	require.Len(t, got, 3, "1回に処理する件数に上限がある(ネクスウェイの処理上限は1分に30件)")
	assert.Equal(t, int64(5), got[0].Id)
	assert.Equal(t, int64(3), got[2].Id, "新しい(Idの大きい)ものから処理する。古い行が枠を使い切って新しい行が処理されなくなるのを防ぐ")
}

func TestFindStuckJpkiRequestLogs_対象が無ければ空(t *testing.T) {
	ctx, _ := newEkycJpkiTestDB(t)

	got, err := FindStuckJpkiRequestLogs(ctx, time.Now(), time.Now().Add(-14*24*time.Hour), 10)
	require.NoError(t, err)

	assert.NotNil(t, got)
	assert.Empty(t, got)
}

// ネクスウェイが結果を保管している期間より古い行は、対象にしない(検証結果取得APIで引けない)。
func TestFindStuckJpkiRequestLogs_保管期間より古い行は対象にしない(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	ctx, engine := newEkycJpkiTestDB(t,
		EkycRequestLog{Email: "1@example.com", UserId: 1, EkycId: "within-window", RequestId: jpkiReqIdPtr("req-1"), Status: EKYC_REQUEST_STATUS_NONE},
		EkycRequestLog{Email: "2@example.com", UserId: 2, EkycId: "beyond-window", RequestId: jpkiReqIdPtr("req-2"), Status: EKYC_REQUEST_STATUS_NONE},
	)
	for id, created := range map[int64]time.Time{1: now.Add(-3 * 24 * time.Hour), 2: now.Add(-20 * 24 * time.Hour)} {
		_, err := engine.Exec("UPDATE ekyc_request_log SET created = ? WHERE id = ?", created.Format("2006-01-02 15:04:05"), id)
		require.NoError(t, err)
	}

	got, err := FindStuckJpkiRequestLogs(ctx, now.Add(-1*time.Hour), now.Add(-14*24*time.Hour), 10)
	require.NoError(t, err)

	require.Len(t, got, 1)
	assert.Equal(t, "within-window", got[0].EkycId)
}
