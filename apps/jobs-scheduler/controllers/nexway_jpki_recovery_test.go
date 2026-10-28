package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	pkgerrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	userservice "github.com/illuminarean/energyshares-user-api/service"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

// fakeRetryNotifier は jpkiRetryNotifier の差し替え。呼ばれた内容を記録する。
type fakeRetryNotifier struct {
	err   error
	calls []fakeRetryNotifierCall
}

type fakeRetryNotifierCall struct {
	userId  int64
	reasons []string
}

func (f *fakeRetryNotifier) NotifyRetryVerify(ctx context.Context, userId int64, reasons []string) error {
	f.calls = append(f.calls, fakeRetryNotifierCall{userId: userId, reasons: reasons})
	return f.err
}

// echoJpkiService は、検証依頼IDごとに結果を返す IJpkiNexwayService の差し替え。複数の記録を1回の
// ジョブで処理するテスト用(検証結果の reference_id・request_id は、問い合わせた記録のものを返す)。
type echoJpkiService struct {
	results  map[string]*service.JpkiSignatureResult // 検証依頼ID -> 結果(reference_id・request_idは補う)
	errs     map[string]error
	register *service.JpkiBpoApplicantOutput

	signatureCalls []string
	registerCalls  int
}

func (f *echoJpkiService) GetSignatureResult(ctx context.Context, requestId string) (*service.JpkiSignatureResult, error) {
	f.signatureCalls = append(f.signatureCalls, requestId)
	if err := f.errs[requestId]; err != nil {
		return nil, err
	}
	r, ok := f.results[requestId]
	if !ok {
		return nil, service.ErrNexwayJpkiRequestNotFound
	}
	out := *r
	if out.RequestId == "" {
		out.RequestId = requestId
	}
	return &out, nil
}

func (f *echoJpkiService) RegisterBpoApplicant(ctx context.Context, input service.JpkiBpoApplicantInput) (*service.JpkiBpoApplicantOutput, error) {
	f.registerCalls++
	return f.register, nil
}

func newJpkiRecoveryController(jpki service.IJpkiNexwayService, notifier jpkiRetryNotifier) NexwayJpkiControllers {
	return NexwayJpkiControllers{
		privateUserService: userservice.NewPrivateUserService(),
		jpkiService:        jpki,
		notifier:           notifier,
	}
}

// postJpkiWebhookWith は、Webhookを1回受信させた結果を、HTTPステータス相当(200・500)で返す。
func postJpkiWebhookWith(t *testing.T, ctx context.Context, c NexwayJpkiControllers, body string) int {
	t.Helper()
	var request NexwayJpkiWebhookRequest
	require.NoError(t, json.Unmarshal([]byte(body), &request))
	if err := c.handleSignatureResult(ctx, request, signatureResultFromWebhook); err != nil {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func signatureWithErrors(result string, codes ...string) *service.JpkiSignatureResult {
	r := completedSignature(result)
	r.Errors = service.JpkiSignatureErrors(codes)
	return r
}

// ---- Webhook: 署名が完了しなかったときの、理由の記録と会員への通知 ----

func TestNexwayJpkiControllers_署名失敗_理由を記録し会員へ通知する(t *testing.T) {
	setupServiceConfig()

	tests := []struct {
		name        string
		signature   *service.JpkiSignatureResult
		wantMessage string
		wantReasons []string
	}{
		{
			name:        "C04(証明書の有効期限切れ): コードと詳細を記録し、利用者向けの理由を通知する",
			signature:   signatureWithErrors("error", "C04"),
			wantMessage: "C04: 署名用電子証明書が有効期限切れです。マイナンバーカードの更新手続きが必要です。\n",
			wantReasons: []string{"署名用電子証明書が有効期限切れです。マイナンバーカードの更新手続きが必要です。"},
		},
		{
			name:        "C06(ロック)",
			signature:   signatureWithErrors("canceled", "C06"),
			wantMessage: "C06: 署名用電子証明書がロックしています。マイナンバーカードのロック解除が必要です。\n",
			wantReasons: []string{"署名用電子証明書がロックしています。マイナンバーカードのロック解除が必要です。"},
		},
		{
			name:        "エラー情報が無い失敗: 理由は記録せず、署名が完了しなかったとだけ通知する",
			signature:   completedSignature("failed"),
			wantMessage: "",
			wantReasons: []string{"署名が完了しませんでした。"},
		},
		{
			name:        "E05(サービス側のエラー): 記録には残すが、利用者へは内部の事情を出さない",
			signature:   signatureWithErrors("error", "E05"),
			wantMessage: "E05: Webhook通知が失敗しました。\n",
			wantReasons: []string{"署名が完了しませんでした。"},
		},
		{
			name:        "C04とE05: 利用者へは具体的な理由だけを通知する",
			signature:   signatureWithErrors("error", "C04", "E05"),
			wantMessage: "C04: 署名用電子証明書が有効期限切れです。マイナンバーカードの更新手続きが必要です。\nE05: Webhook通知が失敗しました。\n",
			wantReasons: []string{"署名用電子証明書が有効期限切れです。マイナンバーカードの更新手続きが必要です。"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := setupJpkiWebhookTestDB(t, nil)
			notifier := &fakeRetryNotifier{}
			controller := newJpkiRecoveryController(&fakeJpkiNexwayService{signature: tt.signature}, notifier)

			code := postJpkiWebhookWith(t, fx.ctx, controller, jpkiWebhookBody("failed"))

			require.Equal(t, http.StatusOK, code)
			got := fx.ekycLog(t)
			assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, got.Status)
			if tt.wantMessage == "" {
				assert.Nil(t, got.ErrorMessage)
			} else {
				require.NotNil(t, got.ErrorMessage)
				assert.Equal(t, tt.wantMessage, *got.ErrorMessage)
			}
			require.Len(t, notifier.calls, 1)
			assert.Equal(t, int64(1), notifier.calls[0].userId)
			assert.Equal(t, tt.wantReasons, notifier.calls[0].reasons)
		})
	}
}

func TestNexwayJpkiControllers_署名成功では通知しない(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	notifier := &fakeRetryNotifier{}
	controller := newJpkiRecoveryController(&fakeJpkiNexwayService{
		signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 9001},
	}, notifier)

	code := postJpkiWebhookWith(t, fx.ctx, controller, jpkiWebhookBody("succeeded"))

	require.Equal(t, http.StatusOK, code)
	assert.Empty(t, notifier.calls)
}

// Webhookの再送で同じ失敗を2回処理しても、メールは1通だけ(2回目は処理済みで飛ばされる)。
func TestNexwayJpkiControllers_同じ失敗を2回処理しても通知は1通(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	notifier := &fakeRetryNotifier{}
	controller := newJpkiRecoveryController(&fakeJpkiNexwayService{signature: signatureWithErrors("error", "C04")}, notifier)

	require.Equal(t, http.StatusOK, postJpkiWebhookWith(t, fx.ctx, controller, jpkiWebhookBody("error")))
	require.Equal(t, http.StatusOK, postJpkiWebhookWith(t, fx.ctx, controller, jpkiWebhookBody("error")))

	assert.Len(t, notifier.calls, 1)
}

// 会員が新しい署名用URLを発行したあとに、古い行の失敗が確定しても、会員へは知らせない
// (古いURLの有効期限切れなどで、署名の最中の会員を混乱させない)。
func TestNexwayJpkiControllers_新しい発行がある古い行の失敗は通知しない(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	newer := "req-newer"
	_, err := fx.engine.Insert(&domain.EkycRequestLog{
		Email: "yamada@example.com", UserId: 1, EkycId: "ref-newer", RequestId: &newer, Status: domain.EKYC_REQUEST_STATUS_NONE,
	})
	require.NoError(t, err)
	notifier := &fakeRetryNotifier{}
	controller := newJpkiRecoveryController(&fakeJpkiNexwayService{signature: signatureWithErrors("canceled", "C02")}, notifier)

	code := postJpkiWebhookWith(t, fx.ctx, controller, jpkiWebhookBody("canceled"))

	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, fx.ekycLog(t).Status, "古い行の失敗は記録する")
	assert.Empty(t, notifier.calls, "会員へは知らせない")
}

// メールが送れなくても、Webhookは200で応答し、行は失敗のまま(再送しても、行は処理済みで、メールは送れない)。
func TestNexwayJpkiControllers_通知に失敗してもWebhookは成功する(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	notifier := &fakeRetryNotifier{err: errors.New("ses down")}
	controller := newJpkiRecoveryController(&fakeJpkiNexwayService{signature: signatureWithErrors("error", "C01")}, notifier)
	hook := logrustest.NewGlobal()
	defer hook.Reset()

	code := postJpkiWebhookWith(t, fx.ctx, controller, jpkiWebhookBody("error"))

	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, fx.ekycLog(t).Status)
	require.Len(t, notifier.calls, 1)
	var logged bool
	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.ErrorLevel && strings.Contains(entry.Message, "failed to send the retry verification mail") {
			logged = true
		}
	}
	assert.True(t, logged, "送れなかったことは、エラーのログに残す")
}

// 新しい発行で置き換えられた行(succeededだが最新でない)は、これまでどおり、会員へは通知しない。
func TestNexwayJpkiControllers_置き換えられた行の失敗は通知しない(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	newer := "req-newer"
	_, err := fx.engine.Insert(&domain.EkycRequestLog{
		Email: "yamada@example.com", UserId: 1, EkycId: "ref-newer", RequestId: &newer, Status: domain.EKYC_REQUEST_STATUS_NONE,
	})
	require.NoError(t, err)
	notifier := &fakeRetryNotifier{}
	fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded")}
	controller := newJpkiRecoveryController(fake, notifier)

	code := postJpkiWebhookWith(t, fx.ctx, controller, jpkiWebhookBody("succeeded"))

	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, fx.ekycLog(t).Status)
	assert.Empty(t, notifier.calls)
}

// レビュー指摘(#655・#661): 署名に成功したあと、会員がホ方式の画面を開くと、新しい行(NONE)ができる。
// その行は申し込まれていないので、署名の結果を捨てる理由にしない(BPOへ申請する)。署名に失敗した場合も、
// 会員はやり直していないので、知らせる。
func TestNexwayJpkiControllers_画面を開いただけの新しい行があっても処理する(t *testing.T) {
	t.Run("署名成功: BPOへ申請する", func(t *testing.T) {
		setupServiceConfig()
		fx := setupJpkiWebhookTestDB(t, nil)
		_, err := fx.engine.Insert(&domain.EkycRequestLog{
			Email: "yamada@example.com", UserId: 1, EkycId: "ref-ho-opened", Status: domain.EKYC_REQUEST_STATUS_NONE,
		})
		require.NoError(t, err)
		fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 555}}
		controller := newJpkiRecoveryController(fake, &fakeRetryNotifier{})

		code := postJpkiWebhookWith(t, fx.ctx, controller, jpkiWebhookBody("succeeded"))

		require.Equal(t, http.StatusOK, code)
		assert.Equal(t, int64(555), fx.ekycLog(t).ApplicantId, "BPOへ申請し、申請IDを記録する")
		assert.NotEqual(t, domain.EKYC_REQUEST_STATUS_FAIL, fx.ekycLog(t).Status)
	})
	t.Run("署名失敗: 会員へ知らせる", func(t *testing.T) {
		setupServiceConfig()
		fx := setupJpkiWebhookTestDB(t, nil)
		_, err := fx.engine.Insert(&domain.EkycRequestLog{
			Email: "yamada@example.com", UserId: 1, EkycId: "ref-ho-opened", Status: domain.EKYC_REQUEST_STATUS_NONE,
		})
		require.NoError(t, err)
		notifier := &fakeRetryNotifier{}
		controller := newJpkiRecoveryController(&fakeJpkiNexwayService{signature: signatureWithErrors("canceled", "C02")}, notifier)

		code := postJpkiWebhookWith(t, fx.ctx, controller, jpkiWebhookBody("canceled"))

		require.Equal(t, http.StatusOK, code)
		assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, fx.ekycLog(t).Status)
		assert.Len(t, notifier.calls, 1)
	})
	t.Run("比較: 新しい行が申し込まれている(ホ方式で完了報告あり)なら、BPOへ申請しない", func(t *testing.T) {
		setupServiceConfig()
		fx := setupJpkiWebhookTestDB(t, nil)
		_, err := fx.engine.Insert(&domain.EkycRequestLog{
			Email: "yamada@example.com", UserId: 1, EkycId: "ref-ho-done", Status: domain.EKYC_REQUEST_STATUS_SUCCESS,
		})
		require.NoError(t, err)
		fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 556}}
		controller := newJpkiRecoveryController(fake, &fakeRetryNotifier{})

		code := postJpkiWebhookWith(t, fx.ctx, controller, jpkiWebhookBody("succeeded"))

		require.Equal(t, http.StatusOK, code)
		assert.Empty(t, fake.registerCalls, "新しい申請があるので、BPOへ申請しない")
		assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, fx.ekycLog(t).Status)
	})
}

// ---- 会員へ送るメール(送る必要の無い会員を除く) ----

func TestMailJpkiRetryNotifier(t *testing.T) {
	type sent struct {
		sender, email, name string
		reasons             []string
	}
	tests := []struct {
		name     string
		mutate   func(user *domain.User, profile *domain.PrivateProfile, log *domain.EkycRequestLog)
		userId   int64
		wantSent bool
	}{
		{name: "通常の会員(本人確認が未了)へ送る", userId: 1, wantSent: true},
		{
			name:   "本人確認が済んでいる会員(IsEkyc)へは送らない",
			mutate: func(u *domain.User, _ *domain.PrivateProfile, _ *domain.EkycRequestLog) { u.IsEkyc = true },
			userId: 1,
		},
		{
			name: "本登録が承認完了した会員へは送らない",
			mutate: func(u *domain.User, _ *domain.PrivateProfile, _ *domain.EkycRequestLog) {
				u.AuditStatus = domain.COMPLETED_AUDIT
			},
			userId: 1,
		},
		{
			name: "無効の会員へは送らない",
			mutate: func(u *domain.User, _ *domain.PrivateProfile, _ *domain.EkycRequestLog) {
				u.AuditStatus = domain.INVALIDITY_AUDIT
			},
			userId: 1,
		},
		{
			name:   "メールアドレスが空の会員へは送らない",
			mutate: func(u *domain.User, _ *domain.PrivateProfile, _ *domain.EkycRequestLog) { u.Email = " " },
			userId: 1,
		},
		{name: "会員がいなければ、何もしない(エラーにしない)", userId: 999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := setupJpkiWebhookTestDB(t, tt.mutate)
			var got []sent
			notifier := mailJpkiRetryNotifier{
				send: func(ctx context.Context, sender, email, name string, reasons []string) error {
					got = append(got, sent{sender, email, name, reasons})
					return nil
				},
				sender: func() string { return "info@example.com" },
			}

			err := notifier.NotifyRetryVerify(fx.ctx, tt.userId, []string{"署名が完了しませんでした。"})

			require.NoError(t, err)
			if !tt.wantSent {
				assert.Empty(t, got)
				return
			}
			require.Len(t, got, 1)
			assert.Equal(t, "info@example.com", got[0].sender)
			assert.Equal(t, "yamada@example.com", got[0].email)
			assert.Equal(t, "山田 太郎", got[0].name)
			assert.Equal(t, []string{"署名が完了しませんでした。"}, got[0].reasons)
		})
	}
}

func TestMailJpkiRetryNotifier_送信の失敗はエラーで返す(t *testing.T) {
	fx := setupJpkiWebhookTestDB(t, nil)
	notifier := mailJpkiRetryNotifier{
		send: func(ctx context.Context, sender, email, name string, reasons []string) error {
			return errors.New("ses down")
		},
		sender: func() string { return "info@example.com" },
	}

	err := notifier.NotifyRetryVerify(fx.ctx, 1, []string{"署名が完了しませんでした。"})

	assert.Error(t, err)
}

// ---- 定期ジョブ: 通知を取りこぼした記録を拾い直す ----

// insertStuckJpkiLogs は、会員ごとに、署名用URLを発行したままの記録を作る(作成から age 経過)。
// 会員は fx の会員(Id=1)とは別に作らない(ジョブの処理では、会員の読み込みは BPOへの申請の時だけ)。
func insertStuckJpkiLogs(t *testing.T, engine *xorm.Engine, age time.Duration, specs ...domain.EkycRequestLog) []int64 {
	t.Helper()
	var ids []int64
	for i := range specs {
		_, err := engine.Insert(&specs[i])
		require.NoError(t, err)
		_, err = engine.Exec("UPDATE ekyc_request_log SET created = ? WHERE id = ?",
			time.Now().UTC().Add(-age).Format("2006-01-02 15:04:05"), specs[i].Id)
		require.NoError(t, err)
		ids = append(ids, specs[i].Id)
	}
	return ids
}

func jpkiStuckLog(id, userId int64, ekycId, requestId string) domain.EkycRequestLog {
	return domain.EkycRequestLog{
		Id: id, Email: "u@example.com", UserId: userId, EkycId: ekycId, RequestId: &requestId, Status: domain.EKYC_REQUEST_STATUS_NONE,
	}
}

func loadLog(t *testing.T, engine *xorm.Engine, id int64) domain.EkycRequestLog {
	t.Helper()
	var got domain.EkycRequestLog
	has, err := engine.ID(id).Get(&got)
	require.NoError(t, err)
	require.True(t, has)
	return got
}

func TestRecoverStuckSignatureResults_署名成功ならBPOへ申請する(t *testing.T) {
	setupServiceConfig()
	// fx の記録(Id=1)を、1時間以上前に発行したものにする。
	fx := setupJpkiWebhookTestDB(t, nil)
	_, err := fx.engine.Exec("UPDATE ekyc_request_log SET created = ? WHERE id = 1", time.Now().UTC().Add(-2*time.Hour).Format("2006-01-02 15:04:05"))
	require.NoError(t, err)
	fake := &echoJpkiService{
		results:  map[string]*service.JpkiSignatureResult{jpkiTestReqId: {ReferenceId: jpkiTestRefId, Status: service.JpkiStatusCompleted, Result: "succeeded"}},
		register: &service.JpkiBpoApplicantOutput{ApplicantId: 9100},
	}
	notifier := &fakeRetryNotifier{}
	controller := newJpkiRecoveryController(fake, notifier)

	err = controller.recoverStuckSignatureResults(fx.ctx, time.Now())

	require.NoError(t, err)
	assert.Equal(t, []string{jpkiTestReqId}, fake.signatureCalls, "自社の記録の検証依頼IDで、ネクスウェイへ確認する")
	assert.Equal(t, 1, fake.registerCalls)
	got := fx.ekycLog(t)
	assert.Equal(t, int64(9100), got.ApplicantId, "申請IDを記録する(画面が「署名成功・BPO申請済み」を知る元)")
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, got.Status, "結果は確定させない(BPOの結果通知が確定する)")
	assert.Empty(t, notifier.calls)
}

func TestRecoverStuckSignatureResults_署名が失敗していたら失敗にして通知する(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	_, err := fx.engine.Exec("UPDATE ekyc_request_log SET created = ? WHERE id = 1", time.Now().UTC().Add(-2*time.Hour).Format("2006-01-02 15:04:05"))
	require.NoError(t, err)
	fake := &echoJpkiService{results: map[string]*service.JpkiSignatureResult{
		jpkiTestReqId: {ReferenceId: jpkiTestRefId, Status: service.JpkiStatusCompleted, Result: "canceled", Errors: service.JpkiSignatureErrors{"C02"}},
	}}
	notifier := &fakeRetryNotifier{}
	controller := newJpkiRecoveryController(fake, notifier)

	err = controller.recoverStuckSignatureResults(fx.ctx, time.Now())

	require.NoError(t, err)
	got := fx.ekycLog(t)
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, got.Status)
	require.NotNil(t, got.ErrorMessage)
	assert.Equal(t, "C02: 処理の有効期限が切れました。\n", *got.ErrorMessage)
	assert.Equal(t, 0, fake.registerCalls, "失敗した署名ではBPOへ申請しない")
	require.Len(t, notifier.calls, 1, "会員は画面を離れていて失敗に気付けないため、やり直しをお願いする")
	assert.Equal(t, []string{"処理の有効期限が切れました。"}, notifier.calls[0].reasons)
}

func TestRecoverStuckSignatureResults_作成から1時間未満の記録は対象にしない(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil) // 作成直後(Webhookが届く見込みがある)
	fake := &echoJpkiService{results: map[string]*service.JpkiSignatureResult{
		jpkiTestReqId: {ReferenceId: jpkiTestRefId, Status: service.JpkiStatusCompleted, Result: "canceled"},
	}}
	notifier := &fakeRetryNotifier{}
	controller := newJpkiRecoveryController(fake, notifier)

	err := controller.recoverStuckSignatureResults(fx.ctx, time.Now())

	require.NoError(t, err)
	assert.Empty(t, fake.signatureCalls, "ネクスウェイへ問い合わせない")
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, fx.ekycLog(t).Status)
	assert.Empty(t, notifier.calls)
}

func TestRecoverStuckSignatureResults_ネクスウェイ側でまだ確定していなければ次の回へ持ち越す(t *testing.T) {
	setupServiceConfig()
	tests := []struct {
		name      string
		signature *service.JpkiSignatureResult
	}{
		{"処理中(processing)", &service.JpkiSignatureResult{ReferenceId: jpkiTestRefId, Status: "processing", Result: "unverified"}},
		{"完了しているが未検証(unverified)", &service.JpkiSignatureResult{ReferenceId: jpkiTestRefId, Status: service.JpkiStatusCompleted, Result: "unverified"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := setupJpkiWebhookTestDB(t, nil)
			_, err := fx.engine.Exec("UPDATE ekyc_request_log SET created = ? WHERE id = 1", time.Now().UTC().Add(-2*time.Hour).Format("2006-01-02 15:04:05"))
			require.NoError(t, err)
			fake := &echoJpkiService{results: map[string]*service.JpkiSignatureResult{jpkiTestReqId: tt.signature}}
			notifier := &fakeRetryNotifier{}
			controller := newJpkiRecoveryController(fake, notifier)
			hook := logrustest.NewGlobal()
			defer hook.Reset()

			err = controller.recoverStuckSignatureResults(fx.ctx, time.Now())

			require.NoError(t, err, "確定していないのは異常ではない。ジョブはエラーにしない")
			got := fx.ekycLog(t)
			assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, got.Status)
			assert.Zero(t, got.ApplicantId)
			assert.Empty(t, notifier.calls)
			for _, entry := range hook.AllEntries() {
				assert.NotEqual(t, logrus.ErrorLevel, entry.Level, "異常として扱わない: %s", entry.Message)
			}
		})
	}
}

// ネクスウェイが引けない記録: Webhookでは、なりすましの可能性があるため無視するが、ジョブの対象は
// 自社が発行した記録なので、保管期間が過ぎた等として、失敗にして会員へ知らせる。
func TestRecoverStuckSignatureResults_ネクスウェイが引けない記録は失敗にする(t *testing.T) {
	setupServiceConfig()

	t.Run("ジョブ: 失敗にして通知する", func(t *testing.T) {
		fx := setupJpkiWebhookTestDB(t, nil)
		_, err := fx.engine.Exec("UPDATE ekyc_request_log SET created = ? WHERE id = 1", time.Now().UTC().Add(-2*time.Hour).Format("2006-01-02 15:04:05"))
		require.NoError(t, err)
		fake := &echoJpkiService{errs: map[string]error{jpkiTestReqId: service.ErrNexwayJpkiRequestNotFound}}
		notifier := &fakeRetryNotifier{}
		controller := newJpkiRecoveryController(fake, notifier)

		err = controller.recoverStuckSignatureResults(fx.ctx, time.Now())

		require.NoError(t, err)
		assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, fx.ekycLog(t).Status)
		require.Len(t, notifier.calls, 1)
		assert.Equal(t, []string{"署名が完了しませんでした。"}, notifier.calls[0].reasons)
	})

	t.Run("Webhook: 同じ状況でも、これまでどおり無視する(行を失敗にしない)", func(t *testing.T) {
		fx := setupJpkiWebhookTestDB(t, nil)
		fake := &echoJpkiService{errs: map[string]error{jpkiTestReqId: service.ErrNexwayJpkiRequestNotFound}}
		notifier := &fakeRetryNotifier{}
		controller := newJpkiRecoveryController(fake, notifier)

		code := postJpkiWebhookWith(t, fx.ctx, controller, jpkiWebhookBody("canceled"))

		require.Equal(t, http.StatusOK, code)
		assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, fx.ekycLog(t).Status)
		assert.Empty(t, notifier.calls)
	})
}

func TestRecoverStuckSignatureResults_対象外の記録には触れない(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	// fxの記録(Id=1)は対象外にする: BPOへ申請済み(申請IDあり)。
	_, err := fx.engine.Exec("UPDATE ekyc_request_log SET created = ?, applicant_id = 55 WHERE id = 1", time.Now().UTC().Add(-2*time.Hour).Format("2006-01-02 15:04:05"))
	require.NoError(t, err)
	hoLog := domain.EkycRequestLog{Id: 2, Email: "ho@example.com", UserId: 2, EkycId: "ho-ekyc", Status: domain.EKYC_REQUEST_STATUS_NONE} // ホ方式(検証依頼IDなし)
	failed := jpkiStuckLog(3, 3, "failed-ekyc", "req-failed")
	failed.Status = domain.EKYC_REQUEST_STATUS_FAIL
	success := jpkiStuckLog(4, 4, "success-ekyc", "req-success")
	success.Status = domain.EKYC_REQUEST_STATUS_SUCCESS
	success.ApplicantId = 77
	insertStuckJpkiLogs(t, fx.engine, 2*time.Hour, hoLog, failed, success)
	fake := &echoJpkiService{results: map[string]*service.JpkiSignatureResult{}}
	notifier := &fakeRetryNotifier{}
	controller := newJpkiRecoveryController(fake, notifier)

	err = controller.recoverStuckSignatureResults(fx.ctx, time.Now())

	require.NoError(t, err)
	assert.Empty(t, fake.signatureCalls, "ホ方式の行・確定済みの行・BPO申請済みの行は、ネクスウェイへ問い合わせない")
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, loadLog(t, fx.engine, 2).Status)
}

func TestRecoverStuckSignatureResults_1回の件数に上限がある(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	// 記録(Id=1)を対象外にし、30件の対象を作る。
	_, err := fx.engine.Exec("UPDATE ekyc_request_log SET applicant_id = 1 WHERE id = 1")
	require.NoError(t, err)
	results := map[string]*service.JpkiSignatureResult{}
	var specs []domain.EkycRequestLog
	for i := 0; i < 30; i++ {
		reqId := "req-" + string(rune('A'+i/26)) + string(rune('a'+i%26))
		specs = append(specs, jpkiStuckLog(int64(100+i), int64(100+i), "ekyc-"+reqId, reqId))
		results[reqId] = &service.JpkiSignatureResult{ReferenceId: "ekyc-" + reqId, Status: service.JpkiStatusCompleted, Result: "canceled"}
	}
	insertStuckJpkiLogs(t, fx.engine, 2*time.Hour, specs...)
	fake := &echoJpkiService{results: results}
	controller := newJpkiRecoveryController(fake, &fakeRetryNotifier{})

	err = controller.recoverStuckSignatureResults(fx.ctx, time.Now())

	require.NoError(t, err)
	assert.Len(t, fake.signatureCalls, jpkiRecoveryBatchSize, "ネクスウェイの処理上限(1分に30件)を超えない")
	var failedCount int
	for i := 0; i < 30; i++ {
		if loadLog(t, fx.engine, int64(100+i)).Status == domain.EKYC_REQUEST_STATUS_FAIL {
			failedCount++
		}
	}
	assert.Equal(t, jpkiRecoveryBatchSize, failedCount, "残りは次の回で処理する")
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, loadLog(t, fx.engine, 129).Status, "新しい記録から処理する")
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, loadLog(t, fx.engine, 100).Status)
}

func TestRecoverStuckSignatureResults_一部が失敗してもほかの記録は処理する(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	_, err := fx.engine.Exec("UPDATE ekyc_request_log SET applicant_id = 1 WHERE id = 1")
	require.NoError(t, err)
	insertStuckJpkiLogs(t, fx.engine, 2*time.Hour,
		jpkiStuckLog(10, 10, "ekyc-bad", "req-bad"),
		jpkiStuckLog(11, 11, "ekyc-good", "req-good"),
	)
	fake := &echoJpkiService{
		results: map[string]*service.JpkiSignatureResult{
			"req-good": {ReferenceId: "ekyc-good", Status: service.JpkiStatusCompleted, Result: "canceled"},
		},
		errs: map[string]error{"req-bad": pkgerrors.Wrap(service.ErrNexwayUnavailable, "timeout")},
	}
	controller := newJpkiRecoveryController(fake, &fakeRetryNotifier{})

	err = controller.recoverStuckSignatureResults(fx.ctx, time.Now())

	require.NoError(t, err, "1件でも処理できていれば、ジョブとしては動いている")
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, loadLog(t, fx.engine, 11).Status)
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, loadLog(t, fx.engine, 10).Status, "失敗した記録は変えない(次の回で確認し直す)")
}

func TestRecoverStuckSignatureResults_全件が失敗したらエラーにする(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	_, err := fx.engine.Exec("UPDATE ekyc_request_log SET created = ? WHERE id = 1", time.Now().UTC().Add(-2*time.Hour).Format("2006-01-02 15:04:05"))
	require.NoError(t, err)
	fake := &echoJpkiService{errs: map[string]error{jpkiTestReqId: service.ErrNexwayJpkiNotConfigured}}
	controller := newJpkiRecoveryController(fake, &fakeRetryNotifier{})

	err = controller.recoverStuckSignatureResults(fx.ctx, time.Now())

	assert.Error(t, err, "設定の不備・ネクスウェイの障害など全体の問題に気付けるようにする")
	assert.Equal(t, domain.EKYC_REQUEST_STATUS_NONE, fx.ekycLog(t).Status)
}

func TestRecoverStuckSignatureResults_対象が無ければ何もしない(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	fake := &echoJpkiService{}
	controller := newJpkiRecoveryController(fake, &fakeRetryNotifier{})

	err := controller.recoverStuckSignatureResults(fx.ctx, time.Now())

	require.NoError(t, err)
	assert.Empty(t, fake.signatureCalls)
}

// ジョブには「通知が言っている結果」が無い。ネクスウェイの結果だけで判断し、結果の食い違いの
// 警告(Webhookのなりすましを疑うためのログ)を出さない。
func TestRecoverStuckSignatureResults_結果の食い違いの警告を出さない(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil)
	_, err := fx.engine.Exec("UPDATE ekyc_request_log SET created = ? WHERE id = 1", time.Now().UTC().Add(-2*time.Hour).Format("2006-01-02 15:04:05"))
	require.NoError(t, err)
	fake := &echoJpkiService{results: map[string]*service.JpkiSignatureResult{
		jpkiTestReqId: {ReferenceId: jpkiTestRefId, Status: service.JpkiStatusCompleted, Result: "canceled"},
	}}
	controller := newJpkiRecoveryController(fake, &fakeRetryNotifier{})
	hook := logrustest.NewGlobal()
	defer hook.Reset()

	require.NoError(t, controller.recoverStuckSignatureResults(fx.ctx, time.Now()))

	for _, entry := range hook.AllEntries() {
		assert.NotContains(t, entry.Message, "differs from nexway", "ジョブでは食い違いの警告を出さない")
	}
}

// ジョブの対象を選ぶ時間は、nowから決まる(現在時刻に依存しない)。
func TestRecoverStuckSignatureResults_対象の時間はnowで決まる(t *testing.T) {
	setupServiceConfig()
	fx := setupJpkiWebhookTestDB(t, nil) // 作成直後
	fake := &echoJpkiService{results: map[string]*service.JpkiSignatureResult{
		jpkiTestReqId: {ReferenceId: jpkiTestRefId, Status: service.JpkiStatusCompleted, Result: "canceled"},
	}}
	controller := newJpkiRecoveryController(fake, &fakeRetryNotifier{})

	// 2時間後の時点で実行すれば、対象になる。
	require.NoError(t, controller.recoverStuckSignatureResults(fx.ctx, time.Now().Add(2*time.Hour)))

	assert.Equal(t, domain.EKYC_REQUEST_STATUS_FAIL, fx.ekycLog(t).Status)
}
