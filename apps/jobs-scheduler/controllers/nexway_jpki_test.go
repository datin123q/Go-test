package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	pkgerrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	userservice "github.com/illuminarean/energyshares-user-api/service"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"

	_ "github.com/mattn/go-sqlite3"
)

// fakeJpkiNexwayService は IJpkiNexwayService の差し替え。
type fakeJpkiNexwayService struct {
	signature    *service.JpkiSignatureResult
	signatureErr error
	register     *service.JpkiBpoApplicantOutput
	registerErr  error

	// onRegister は、BPOへの申請の直後(申請IDの保存の前)に呼ばれる。申請と保存の間に割り込む出来事
	// (DB障害、BPOの結果通知が先に届く等)を再現するための差し込み口。
	onRegister func()

	signatureCalls   []string
	registerCalls    []service.JpkiBpoApplicantInput
	ctxCancelledSeen []bool
}

func (f *fakeJpkiNexwayService) GetSignatureResult(ctx context.Context, requestId string) (*service.JpkiSignatureResult, error) {
	f.signatureCalls = append(f.signatureCalls, requestId)
	f.ctxCancelledSeen = append(f.ctxCancelledSeen, ctx.Err() != nil)
	if f.signatureErr != nil {
		return nil, f.signatureErr
	}
	return f.signature, nil
}

func (f *fakeJpkiNexwayService) RegisterBpoApplicant(ctx context.Context, input service.JpkiBpoApplicantInput) (*service.JpkiBpoApplicantOutput, error) {
	f.registerCalls = append(f.registerCalls, input)
	f.ctxCancelledSeen = append(f.ctxCancelledSeen, ctx.Err() != nil)
	if f.registerErr != nil {
		return nil, f.registerErr
	}
	if f.onRegister != nil {
		f.onRegister()
	}
	return f.register, nil
}

const (
	jpkiTestRefId = "ref-jpki-1"
	jpkiTestReqId = "req-jpki-1"
)

// jpkiTestFixture は、ある会員とそのeKYCリクエスト記録(進行中)を持つDBとctxを作る。
type jpkiTestFixture struct {
	ctx    context.Context
	engine *xorm.Engine
}

func setupJpkiWebhookTestDB(t *testing.T, mutate func(user *domain.User, profile *domain.PrivateProfile, log *domain.EkycRequestLog)) jpkiTestFixture {
	t.Helper()

	address2 := "ビル101"
	requestId := jpkiTestReqId
	user := domain.User{
		Id: 1, Uid: "uid-1", LoginType: domain.MAIL, UserType: domain.PRIVATE, CommonId: "common-1",
		InvestorType: domain.PRIVATE_INVESTOR, Name: "山田 太郎", Email: "yamada@example.com",
		Password: "pass", Salt: "salt", AuditStatus: domain.WAIT_AUDIT, SignupStatus: domain.SIGNUP_NONE,
		Province: "東京都", City: "千代田区", Address: "丸の内1-1-1", Address2: &address2,
	}
	profile := domain.PrivateProfile{
		Id: 1, UserId: 1, FamilyName: "山田", FirstName: "太郎", RequestId: "signup-req-1",
		Birthday: jsonutils.JSONTime(time.Date(1990, 1, 2, 0, 0, 0, 0, time.FixedZone("JST", 9*3600))),
	}
	ekycLog := domain.EkycRequestLog{
		Id: 1, Email: "yamada@example.com", UserId: 1, EkycId: jpkiTestRefId, RequestId: &requestId, Status: domain.EKYC_REQUEST_STATUS_NONE,
	}
	if mutate != nil {
		mutate(&user, &profile, &ekycLog)
	}

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	require.NoError(t, engine.Sync(new(domain.User), new(domain.PrivateProfile), new(domain.EkycRequestLog)))
	_, err = engine.Insert(&user)
	require.NoError(t, err)
	_, err = engine.Insert(&profile)
	require.NoError(t, err)
	_, err = engine.Insert(&ekycLog)
	require.NoError(t, err)

	return jpkiTestFixture{
		ctx:    context.WithValue(context.Background(), factory.ConextDBType, engine),
		engine: engine,
	}
}

func (f jpkiTestFixture) ekycLog(t *testing.T) domain.EkycRequestLog {
	t.Helper()
	var got domain.EkycRequestLog
	has, err := f.engine.ID(1).Get(&got)
	require.NoError(t, err)
	require.True(t, has)
	return got
}

func completedSignature(result string) *service.JpkiSignatureResult {
	return &service.JpkiSignatureResult{
		ReferenceId: jpkiTestRefId, RequestId: jpkiTestReqId, Status: service.JpkiStatusCompleted,
		Result: result, CompletedAt: "2026-10-15T10:00:00+09:00",
	}
}

// postJpkiWebhook はWebhookを1回受信させ、HTTPステータスを返す。
func postJpkiWebhook(t *testing.T, ctx context.Context, fake service.IJpkiNexwayService, body string) int {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/nexway/jpki/webhook", strings.NewReader(body)).WithContext(ctx)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	controller := NexwayJpkiControllers{privateUserService: userservice.NewPrivateUserService(), jpkiService: fake}
	require.NoError(t, controller.CheckSignatureResult(e.NewContext(req, rec)))
	return rec.Code
}

func jpkiWebhookBody(result string) string {
	return `{"reference_id":"` + jpkiTestRefId + `","request_id":"` + jpkiTestReqId + `","result":"` + result + `","completed_at":"2026-10-15T10:00:00+09:00"}`
}

func TestNexwayJpkiControllers_CheckSignatureResult(t *testing.T) {
	setupServiceConfig()

	tests := []struct {
		name                     string
		mutate                   func(user *domain.User, profile *domain.PrivateProfile, log *domain.EkycRequestLog)
		body                     string
		fake                     *fakeJpkiNexwayService
		wantCode                 int
		wantNexwaySignatureCalls int
		wantRegisterCalls        int
		wantStatus               domain.EKYC_REQUEST_STATUS
		wantApplicantId          int64
	}{
		{
			name:     "署名成功: BPOへ申請し、申請IDを記録する。結果は確定させない",
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 9001}},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 1,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE, wantApplicantId: 9001,
		},
		{
			name:     "Webhookはsucceededと主張しているが、Nexwayの記録ではfailed: 申請せず失敗にする(なりすまし)",
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("failed")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_FAIL,
		},
		{
			name:     "Webhookはfailedと主張しているが、Nexwayの記録ではsucceeded: Nexwayの記録に従って申請する",
			body:     jpkiWebhookBody("failed"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 9002}},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 1,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE, wantApplicantId: 9002,
		},
		{
			name:     "署名失敗(failed): 申請せず失敗にする",
			body:     jpkiWebhookBody("failed"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("failed")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_FAIL,
		},
		{
			name:     "処理エラー(error): 申請せず失敗にする",
			body:     jpkiWebhookBody("error"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("error")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_FAIL,
		},
		{
			name:     "中断(canceled): 申請せず失敗にする",
			body:     jpkiWebhookBody("canceled"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("canceled")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_FAIL,
		},
		{
			name:     "自社に無い連携ID: Nexwayへ問い合わせず200",
			body:     `{"reference_id":"unknown-ref","request_id":"req-x","result":"succeeded"}`,
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 0, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name: "失敗と記録済み(FAIL)の記録: 何もしない(再送の重複・新しい発行で閉じられた行)",
			mutate: func(_ *domain.User, _ *domain.PrivateProfile, l *domain.EkycRequestLog) {
				l.Status = domain.EKYC_REQUEST_STATUS_FAIL
			},
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 0, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_FAIL,
		},
		{
			// 画面からの完了報告(PUT)が、署名結果の通知より先に届いた場合。行はSUCCESSで申請IDが無い。
			// 「処理済み」として何もしないと、BPOへ一度も申請されず、結果も来ない。
			name: "完了報告が先に届いた行(SUCCESS・申請IDなし): 署名成功の通知でBPOへ申請する",
			mutate: func(_ *domain.User, _ *domain.PrivateProfile, l *domain.EkycRequestLog) {
				l.Status = domain.EKYC_REQUEST_STATUS_SUCCESS
			},
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 9003}},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 1,
			wantStatus: domain.EKYC_REQUEST_STATUS_SUCCESS, wantApplicantId: 9003,
		},
		{
			name: "完了報告が先に届いた行(SUCCESS・申請IDなし)に署名失敗の通知: 申請せず、確定済みの状態は書き換えない",
			mutate: func(_ *domain.User, _ *domain.PrivateProfile, l *domain.EkycRequestLog) {
				l.Status = domain.EKYC_REQUEST_STATUS_SUCCESS
			},
			body:     jpkiWebhookBody("failed"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("failed")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_SUCCESS,
		},
		{
			name: "完了報告済みでBPOへ申請済み(SUCCESS・申請IDあり): 何もしない(再送の重複)",
			mutate: func(_ *domain.User, _ *domain.PrivateProfile, l *domain.EkycRequestLog) {
				l.Status = domain.EKYC_REQUEST_STATUS_SUCCESS
				l.ApplicantId = 4243
			},
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 0, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_SUCCESS, wantApplicantId: 4243,
		},
		{
			name: "記録に検証依頼IDが無い(ホ方式の行): Nexwayへ問い合わせず、何もしない",
			mutate: func(_ *domain.User, _ *domain.PrivateProfile, l *domain.EkycRequestLog) {
				l.RequestId = nil
			},
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 0, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name: "記録の検証依頼IDが通知と違う(別の発行の通知): Nexwayへ問い合わせず、何もしない",
			mutate: func(_ *domain.User, _ *domain.PrivateProfile, l *domain.EkycRequestLog) {
				other := "another-request-id"
				l.RequestId = &other
			},
			body:     jpkiWebhookBody("failed"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("failed")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 0, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name: "既にBPOへ申請済み(申請IDあり): 再申請しない(再送の重複)",
			mutate: func(_ *domain.User, _ *domain.PrivateProfile, l *domain.EkycRequestLog) {
				l.ApplicantId = 4242
			},
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 0, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE, wantApplicantId: 4242,
		},
		{
			name:     "確認の通信に失敗: 500(再送を受ける)。申請も失敗確定もしない",
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signatureErr: pkgerrors.Wrap(service.ErrNexwayUnavailable, "timeout")},
			wantCode: http.StatusInternalServerError, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name:     "確認の応答が異常: 500。申請しない",
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signatureErr: pkgerrors.Wrap(service.ErrNexwayInvalidResponse, "status 503")},
			wantCode: http.StatusInternalServerError, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name:     "確認の設定が未設定: 500。申請しない",
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signatureErr: pkgerrors.Wrap(service.ErrNexwayJpkiNotConfigured, "empty")},
			wantCode: http.StatusInternalServerError, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name:     "検証依頼IDがNexwayに存在しない: 200で無視。申請も失敗確定もしない(なりすまし)",
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signatureErr: pkgerrors.Wrap(service.ErrNexwayJpkiRequestNotFound, "404")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name: "Nexwayの記録の連携IDがWebhookと違う: 200で無視。申請も失敗確定もしない(他人のrequest_idを騙った通知)",
			body: jpkiWebhookBody("succeeded"),
			fake: &fakeJpkiNexwayService{signature: &service.JpkiSignatureResult{
				ReferenceId: "someone-elses-ref", RequestId: jpkiTestReqId, Status: service.JpkiStatusCompleted, Result: "succeeded",
			}},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name: "Nexwayの記録の検証依頼IDがWebhookと違う: 200で無視",
			body: jpkiWebhookBody("succeeded"),
			fake: &fakeJpkiNexwayService{signature: &service.JpkiSignatureResult{
				ReferenceId: jpkiTestRefId, RequestId: "other-req", Status: service.JpkiStatusCompleted, Result: "succeeded",
			}},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name: "Nexwayの記録がまだ完了していない: 500(再送で確認し直す)",
			body: jpkiWebhookBody("succeeded"),
			fake: &fakeJpkiNexwayService{signature: &service.JpkiSignatureResult{
				ReferenceId: jpkiTestRefId, RequestId: jpkiTestReqId, Status: "verifying", Result: "unverified",
			}},
			wantCode: http.StatusInternalServerError, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name: "結果は成功だが処理ステータスが完了でない(食い違い): 500。申請しない",
			body: jpkiWebhookBody("succeeded"),
			fake: &fakeJpkiNexwayService{signature: &service.JpkiSignatureResult{
				ReferenceId: jpkiTestRefId, RequestId: jpkiTestReqId, Status: "verifying", Result: "succeeded",
			}},
			wantCode: http.StatusInternalServerError, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name: "完了だが結果が未検証・未知の値: 500。申請も失敗確定もしない",
			body: jpkiWebhookBody("succeeded"),
			fake: &fakeJpkiNexwayService{signature: &service.JpkiSignatureResult{
				ReferenceId: jpkiTestRefId, RequestId: jpkiTestReqId, Status: service.JpkiStatusCompleted, Result: "unverified",
			}},
			wantCode: http.StatusInternalServerError, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name:     "BPOが受付済みを返す(重複申請): 200。失敗にしない。申請IDは記録されず、BPOの結果通知(MIR-1468)で記録される",
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded"), registerErr: pkgerrors.Wrap(service.ErrNexwayJpkiAlreadyRegistered, "4000010")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 1,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name:     "BPOが申請内容の不備で拒否: 200で失敗にする(再送しても通らない)",
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded"), registerErr: pkgerrors.Wrap(service.ErrNexwayJpkiApplicantRejected, "4000009")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 1,
			wantStatus: domain.EKYC_REQUEST_STATUS_FAIL,
		},
		{
			name:     "BPOへ通信できない: 500。失敗にしない(再送で申請し直す)",
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded"), registerErr: pkgerrors.Wrap(service.ErrNexwayUnavailable, "timeout")},
			wantCode: http.StatusInternalServerError, wantNexwaySignatureCalls: 1, wantRegisterCalls: 1,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name:     "BPOの応答が異常(認証・署名・IP制限・5xx): 500。失敗にしない",
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded"), registerErr: pkgerrors.Wrap(service.ErrNexwayInvalidResponse, "401 4010003")},
			wantCode: http.StatusInternalServerError, wantNexwaySignatureCalls: 1, wantRegisterCalls: 1,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name:     "BPOの設定(署名用キー等)が未設定: 500。失敗にしない",
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded"), registerErr: pkgerrors.Wrap(service.ErrNexwayJpkiNotConfigured, "secret is empty")},
			wantCode: http.StatusInternalServerError, wantNexwaySignatureCalls: 1, wantRegisterCalls: 1,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name: "住所が空の会員: BPOへ送らず失敗にする",
			mutate: func(u *domain.User, _ *domain.PrivateProfile, _ *domain.EkycRequestLog) {
				u.Province, u.City, u.Address, u.Address2 = "", "", "", nil
			},
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_FAIL,
		},
		{
			name: "氏名が空の会員: BPOへ送らず失敗にする",
			mutate: func(_ *domain.User, p *domain.PrivateProfile, _ *domain.EkycRequestLog) {
				p.FamilyName, p.FirstName = "", ""
			},
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_FAIL,
		},
		{
			name: "生年月日が未設定の会員: BPOへ送らず失敗にする",
			mutate: func(_ *domain.User, p *domain.PrivateProfile, _ *domain.EkycRequestLog) {
				p.Birthday = jsonutils.JSONTime{}
			},
			body:     jpkiWebhookBody("succeeded"),
			fake:     &fakeJpkiNexwayService{signature: completedSignature("succeeded")},
			wantCode: http.StatusOK, wantNexwaySignatureCalls: 1, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_FAIL,
		},
		{
			name:     "本体がJSONでない: 400",
			body:     `not json`,
			fake:     &fakeJpkiNexwayService{},
			wantCode: http.StatusBadRequest, wantNexwaySignatureCalls: 0, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name:     "連携IDが無い: 400",
			body:     `{"request_id":"req-1","result":"succeeded"}`,
			fake:     &fakeJpkiNexwayService{},
			wantCode: http.StatusBadRequest, wantNexwaySignatureCalls: 0, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
		{
			name:     "検証依頼IDが無い: 400",
			body:     `{"reference_id":"` + jpkiTestRefId + `","result":"succeeded"}`,
			fake:     &fakeJpkiNexwayService{},
			wantCode: http.StatusBadRequest, wantNexwaySignatureCalls: 0, wantRegisterCalls: 0,
			wantStatus: domain.EKYC_REQUEST_STATUS_NONE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupJpkiWebhookTestDB(t, tt.mutate)

			code := postJpkiWebhook(t, f.ctx, tt.fake, tt.body)

			assert.Equal(t, tt.wantCode, code)
			assert.Len(t, tt.fake.signatureCalls, tt.wantNexwaySignatureCalls)
			assert.Len(t, tt.fake.registerCalls, tt.wantRegisterCalls)
			got := f.ekycLog(t)
			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, tt.wantApplicantId, got.ApplicantId)
		})
	}
}

func TestNexwayJpkiControllers_CheckSignatureResult_BpoApplicantContent(t *testing.T) {
	setupServiceConfig()

	tests := []struct {
		name   string
		mutate func(user *domain.User, profile *domain.PrivateProfile, log *domain.EkycRequestLog)
		want   service.JpkiBpoApplicantInput
	}{
		{
			name: "氏名は姓名をスペースなしで連結、住所は都道府県から建物名まで、生年月日はYYYY-MM-DD",
			want: service.JpkiBpoApplicantInput{
				ReferenceId: jpkiTestRefId, RequestId: jpkiTestReqId,
				Name: "山田太郎", Birthday: "1990-01-02", Address: "東京都千代田区丸の内1-1-1 ビル101",
			},
		},
		{
			name: "建物名なし・前後に空白がある項目",
			mutate: func(u *domain.User, p *domain.PrivateProfile, _ *domain.EkycRequestLog) {
				u.Address2 = nil
				u.Province, u.City = " 東京都 ", " 千代田区 "
				p.FamilyName, p.FirstName = " 山田 ", " 太郎 "
			},
			want: service.JpkiBpoApplicantInput{
				ReferenceId: jpkiTestRefId, RequestId: jpkiTestReqId,
				Name: "山田太郎", Birthday: "1990-01-02", Address: "東京都千代田区丸の内1-1-1",
			},
		},
		{
			name: "生年月日は日本時間の日付(UTCで前日になる時刻でも日本の日付)",
			mutate: func(_ *domain.User, p *domain.PrivateProfile, _ *domain.EkycRequestLog) {
				// 1990-01-02 00:30 JST = 1990-01-01 15:30 UTC
				p.Birthday = jsonutils.JSONTime(time.Date(1990, 1, 1, 15, 30, 0, 0, time.UTC))
			},
			want: service.JpkiBpoApplicantInput{
				ReferenceId: jpkiTestRefId, RequestId: jpkiTestReqId,
				Name: "山田太郎", Birthday: "1990-01-02", Address: "東京都千代田区丸の内1-1-1 ビル101",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupJpkiWebhookTestDB(t, tt.mutate)
			fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 1}}

			code := postJpkiWebhook(t, f.ctx, fake, jpkiWebhookBody("succeeded"))

			require.Equal(t, http.StatusOK, code)
			require.Len(t, fake.registerCalls, 1)
			assert.Equal(t, tt.want, fake.registerCalls[0])
		})
	}
}

func TestNexwayJpkiControllers_CheckSignatureResult_NexwayCallsAreIndependentOfRequestContext(t *testing.T) {
	setupServiceConfig()
	f := setupJpkiWebhookTestDB(t, nil)
	fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded"), register: &service.JpkiBpoApplicantOutput{ApplicantId: 1}}

	// 送信元が接続を切ると、受信リクエストのctxはキャンセルされる。Nexwayへの通信は影響を受けない。
	ctx, cancel := context.WithCancel(f.ctx)
	cancel()
	require.Error(t, ctx.Err())

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/nexway/jpki/webhook", strings.NewReader(jpkiWebhookBody("succeeded"))).WithContext(ctx)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	controller := NexwayJpkiControllers{privateUserService: userservice.NewPrivateUserService(), jpkiService: fake}
	require.NoError(t, controller.CheckSignatureResult(e.NewContext(req, rec)))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, fake.ctxCancelledSeen, 2)
	assert.Equal(t, []bool{false, false}, fake.ctxCancelledSeen)
}

func TestNexwayJpkiControllers_CheckSignatureResult_DoesNotLogPersonalInfo(t *testing.T) {
	setupServiceConfig()
	hook := logrustest.NewGlobal()
	t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks)) })

	for _, registerErr := range []error{
		nil,
		pkgerrors.Wrap(service.ErrNexwayJpkiApplicantRejected, "4000009"),
		errors.New("boom"),
	} {
		hook.Reset()
		f := setupJpkiWebhookTestDB(t, nil)
		fake := &fakeJpkiNexwayService{
			signature:   completedSignature("succeeded"),
			register:    &service.JpkiBpoApplicantOutput{ApplicantId: 1},
			registerErr: registerErr,
		}
		postJpkiWebhook(t, f.ctx, fake, jpkiWebhookBody("succeeded"))

		require.NotEmpty(t, hook.AllEntries())
		for _, entry := range hook.AllEntries() {
			line := entry.Message
			for k, v := range entry.Data {
				line += " " + k + "=" + toLogString(v)
			}
			for _, pii := range []string{"山田", "太郎", "千代田", "丸の内", "1990", "yamada@example.com"} {
				assert.NotContains(t, line, pii, "ログに個人情報が含まれている: %s", line)
			}
		}
	}
}

func toLogString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case error:
		return s.Error()
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func TestNexwayJpkiControllers_CheckSignatureResult_UnknownReferenceIsLogged(t *testing.T) {
	setupServiceConfig()
	hook := logrustest.NewGlobal()
	t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks)) })
	f := setupJpkiWebhookTestDB(t, nil)
	fake := &fakeJpkiNexwayService{signature: completedSignature("succeeded")}

	code := postJpkiWebhook(t, f.ctx, fake, `{"reference_id":"unknown-ref","request_id":"req-x","result":"succeeded"}`)

	require.Equal(t, http.StatusOK, code)
	var warned bool
	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.WarnLevel && strings.Contains(entry.Message, "unknown reference_id") {
			warned = true
		}
	}
	assert.True(t, warned, "自社に無い連携IDは、状態が確定済みの通知とは区別して警告を残す")
}
