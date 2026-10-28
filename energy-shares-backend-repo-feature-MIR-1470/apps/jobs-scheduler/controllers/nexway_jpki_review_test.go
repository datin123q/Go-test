package controllers

import (
	"net/http"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

// MIR-1470 レビュー(rkouchi-hub)で指摘された点の回帰テスト。

// 同じ会員がURLを2回発行して両方で署名した場合に、古い発行(この行)をBPOへ申請しない。
// 最新の発行だけを申請する(有料のBPO申請を無駄にしない。古い行の結果は会員の状態を動かさない)。
func TestNexwayJpkiControllers_CheckSignatureResult_より新しい行がある古い行は申請しない(t *testing.T) {
	setupServiceConfig()
	f := setupJpkiWebhookTestDB(t, nil)
	newerRequestId := "newer-request-id"
	_, err := f.engine.Insert(&domain.EkycRequestLog{
		Id: 2, Email: "yamada@example.com", UserId: 1, EkycId: "ref-jpki-newer", RequestId: &newerRequestId,
		Status: domain.EKYC_REQUEST_STATUS_NONE,
	})
	require.NoError(t, err)
	fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 9001}}

	code := postJpkiWebhook(t, f.ctx, fake, jpkiWebhookBody("succeeded"))

	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, fake.signatureCalls, 1, "署名結果の確認はする")
	assert.Empty(t, fake.registerCalls, "より新しい行があるのに、BPOへ申請している")
	got := f.ekycLog(t)
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, got.Status, "古い行は閉じる")
	assert.Zero(t, got.ApplicantId)
	var newer domain.EkycRequestLog
	has, err := f.engine.ID(2).Get(&newer)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, newer.Status, "新しい行には触れない")
}

// 記録の取得でDB障害: 500(Nexwayの再送を受ける)。Nexwayへは問い合わせない。
func TestNexwayJpkiControllers_CheckSignatureResult_記録の取得でDB障害なら500(t *testing.T) {
	setupServiceConfig()
	f := setupJpkiWebhookTestDB(t, nil)
	require.NoError(t, f.engine.Close()) // 以降のDBアクセスはエラーになる
	fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded")}

	code := postJpkiWebhook(t, f.ctx, fake, jpkiWebhookBody("succeeded"))

	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Empty(t, fake.signatureCalls)
	assert.Empty(t, fake.registerCalls)
}

// 会員が見つからない(退会・削除等): 200で申請しない。再送しても状況が変わらないため。
func TestNexwayJpkiControllers_CheckSignatureResult_会員が見つからなければ200で申請しない(t *testing.T) {
	setupServiceConfig()
	f := setupJpkiWebhookTestDB(t, func(_ *domain.User, _ *domain.PrivateProfile, l *domain.EkycRequestLog) {
		l.UserId = 999 // 存在しない会員
	})
	fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 9001}}

	code := postJpkiWebhook(t, f.ctx, fake, jpkiWebhookBody("succeeded"))

	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, fake.registerCalls)
	got := f.ekycLog(t)
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, got.Status)
	assert.Zero(t, got.ApplicantId)
}

// 申請は受け付けられたが、申請IDの保存でDB障害 → 500 → Nexwayの再送 → BPOが「受付済み」を返す → 200。
// 重複して申請されないが、この経路では申請IDは記録されない(BPOの結果通知で記録される)。
func TestNexwayJpkiControllers_CheckSignatureResult_申請IDの保存に失敗しても再送で重複申請しない(t *testing.T) {
	setupServiceConfig()
	f := setupJpkiWebhookTestDB(t, nil)
	fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 9001}}
	// 申請の直後(保存の前)にDBが使えなくなる状況を、テーブル名の変更で再現する。
	fake.onRegister = func() {
		_, err := f.engine.Exec("ALTER TABLE ekyc_request_log RENAME TO ekyc_request_log_unavailable")
		require.NoError(t, err)
	}

	first := postJpkiWebhook(t, f.ctx, fake, jpkiWebhookBody("succeeded"))

	assert.Equal(t, http.StatusInternalServerError, first, "保存に失敗したら500にして、Nexwayの再送を受ける")
	require.Len(t, fake.registerCalls, 1)

	// DBが戻り、Nexwayが再送してくる。BPOは「受付済み」を返す。
	_, err := f.engine.Exec("ALTER TABLE ekyc_request_log_unavailable RENAME TO ekyc_request_log")
	require.NoError(t, err)
	fake.onRegister = nil
	fake.registerErr = pkgerrors.Wrap(service.ErrNexwayJpkiAlreadyRegistered, "4000010")

	second := postJpkiWebhook(t, f.ctx, fake, jpkiWebhookBody("succeeded"))

	assert.Equal(t, http.StatusOK, second)
	require.Len(t, fake.registerCalls, 2, "再送でもう一度申請を試みる(BPOが受付済みを返すので重複しない)")
	got := f.ekycLog(t)
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, got.Status, "失敗にはしない")
	assert.Zero(t, got.ApplicantId, "受付済みの応答には申請IDが無い。BPOの結果通知で記録される")
}

// 申請IDが既に記録されていれば(BPOの結果通知が先に届いた等)、上書きせず、Warnで残す。
func TestNexwayJpkiControllers_CheckSignatureResult_申請IDが既にあれば上書きせずWarnで残す(t *testing.T) {
	setupServiceConfig()
	hook := logrustest.NewGlobal()
	t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks)) })
	f := setupJpkiWebhookTestDB(t, nil)
	fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 9001}}
	fake.onRegister = func() {
		_, err := f.engine.Exec("UPDATE ekyc_request_log SET applicant_id = 555 WHERE id = 1")
		require.NoError(t, err)
	}

	code := postJpkiWebhook(t, f.ctx, fake, jpkiWebhookBody("succeeded"))

	assert.Equal(t, http.StatusOK, code)
	assert.EqualValues(t, 555, f.ekycLog(t).ApplicantId, "先に記録された申請IDを上書きしてはいけない")
	var warned bool
	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.WarnLevel && entry.Data["updated"] == int64(0) {
			warned = true
		}
	}
	assert.True(t, warned, "0件更新をWarnで残していない(調査の手掛かりにならない)")
}

// 申請IDを保存できたとき(1件更新)は、Infoで更新件数を残す。
func TestNexwayJpkiControllers_CheckSignatureResult_申請IDを保存できたらInfoで件数を残す(t *testing.T) {
	setupServiceConfig()
	hook := logrustest.NewGlobal()
	t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks)) })
	f := setupJpkiWebhookTestDB(t, nil)
	fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 9001}}

	postJpkiWebhook(t, f.ctx, fake, jpkiWebhookBody("succeeded"))

	var logged bool
	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.InfoLevel && entry.Data["updated"] == int64(1) && entry.Data["applicant_id"] == int64(9001) {
			logged = true
		}
	}
	assert.True(t, logged, "更新件数が残っていない")
}

// 認証の無い公開口のため、BPOの結果通知と同じく、本文の大きさに上限(1MB)を設ける。超えれば413で止め、
// 記録の取得やネクスウェイへの問い合わせはしない。上限ちょうどまでは受け付ける(Bindへ渡す)。
func TestNexwayJpkiControllers_CheckSignatureResult_本文の上限(t *testing.T) {
	setupServiceConfig()

	t.Run("1MBを超える本文は413", func(t *testing.T) {
		f := setupJpkiWebhookTestDB(t, nil)
		fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded")}
		body := `{"reference_id":"` + jpkiTestRefId + `","request_id":"` + jpkiTestReqId + `","pad":"` +
			strings.Repeat("x", nexwayWebhookMaxBodyBytes) + `"}`

		code := postJpkiWebhook(t, f.ctx, fake, body)

		assert.Equal(t, http.StatusRequestEntityTooLarge, code)
		assert.Empty(t, fake.signatureCalls, "ネクスウェイへ問い合わせない")
		assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, f.ekycLog(t).Status)
	})

	t.Run("上限ちょうどの本文は受け付ける", func(t *testing.T) {
		f := setupJpkiWebhookTestDB(t, nil)
		fake := &fakeJpkiNexwayService{signature: completedSignature("failed")}
		head := `{"reference_id":"` + jpkiTestRefId + `","request_id":"` + jpkiTestReqId + `","result":"failed","pad":"`
		tail := `"}`
		body := head + strings.Repeat("x", nexwayWebhookMaxBodyBytes-len(head)-len(tail)) + tail
		require.Len(t, body, nexwayWebhookMaxBodyBytes)

		code := postJpkiWebhook(t, f.ctx, fake, body)

		assert.Equal(t, http.StatusOK, code)
		assert.Len(t, fake.signatureCalls, 1, "通常どおり、ネクスウェイの記録で確かめる")
	})
}
