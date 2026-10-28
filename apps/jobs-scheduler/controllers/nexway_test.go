package controllers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	userservice "github.com/illuminarean/energyshares-user-api/service"
	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"

	_ "github.com/mattn/go-sqlite3"
)

type fakeNexwayService struct {
	output *service.EkycResultOutput
	err    error
	// receivedCancelledCtx: 呼び出し時点でctxが既にDone()だったかを記録する。
	// MIR-1223: nexwayへの照会は受信リクエストのctxから独立させたため、
	// Webhookの受信リクエストが(攻撃者による切断等で)キャンセル済みでも
	// falseのままであるべき。
	receivedCancelledCtx bool
}

func (f *fakeNexwayService) GetEkycResultByApplicantId(ctx context.Context, applicantId string) (*service.EkycResultOutput, error) {
	select {
	case <-ctx.Done():
		f.receivedCancelledCtx = true
	default:
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.output, nil
}

type fakeUserInfoHistoryService struct {
	beforeCalls   int
	createCalled  bool
	getDataErrors []error
	getDataIndex  int
	// concurrentMutationは、GetDataForPrivateUser(ekycAfterProcessが会員を読んだ
	// 直後、更新の確定前に呼ぶ)の中で実行する。会員のログイン・本人情報の変更が
	// 同じ秒に割り込んだ場合を再現するためのフック(MIR-1373/1374、②レビューで
	// 指摘。旧実装(行を丸ごと書き戻す)と新実装(UpdateEkycResultで4列だけ書く)の
	// 違いを実際に検出できるテストにするため)。GetDataForPrivateUserはbefore用・
	// after用で計2回呼ばれるため、concurrentMutationDoneで最初の1回だけに絞る
	// (絞らないと、after用の呼び出しがUpdateの後で再度DBを書き換えてしまい、
	// Updateの実際の挙動をpostCheckが見えなくなる。実装時にこれで一度、旧実装でも
	// テストが通ってしまう誤りを踏んだ)。
	concurrentMutation     func()
	concurrentMutationDone bool
}

type failingReader struct {
	err error
}

func (r failingReader) Read(p []byte) (int, error) {
	return 0, r.err
}

func (f *fakeUserInfoHistoryService) GetDataForUser(ctx context.Context, userId int64) (*service.UserInfoHistoryData, error) {
	return nil, nil
}

func (f *fakeUserInfoHistoryService) GetDataForPrivateUser(ctx context.Context, userId int64) (*service.UserInfoHistoryData, error) {
	defer func() { f.beforeCalls++ }()
	if f.concurrentMutation != nil && !f.concurrentMutationDone {
		f.concurrentMutationDone = true
		f.concurrentMutation()
	}
	if f.getDataIndex < len(f.getDataErrors) {
		err := f.getDataErrors[f.getDataIndex]
		f.getDataIndex++
		if err != nil {
			return nil, err
		}
	} else {
		f.getDataIndex++
	}

	return &service.UserInfoHistoryData{
		User: &domain.PrivateUser{
			User:           domain.User{Id: userId},
			PrivateProfile: domain.PrivateProfile{Id: 1, UserId: userId},
		},
	}, nil
}

func (f *fakeUserInfoHistoryService) GetDataForCorporateUser(ctx context.Context, userId int64) (*service.UserInfoHistoryData, error) {
	return nil, nil
}

func (f *fakeUserInfoHistoryService) CreateUserInfoHistory(
	ctx context.Context, before, after service.UserInfoHistoryData, processId int64, adminName, adminEmail string,
) error {
	f.createCalled = true
	return nil
}

type emailTestServer struct {
	*httptest.Server
	mu       sync.Mutex
	requests int
	lastBody []byte
	status   int
	respBody []byte
}

func newEmailTestServer() *emailTestServer {
	tets := &emailTestServer{status: http.StatusOK, respBody: []byte("[1]")}
	tets.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/mails/send-request" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		tets.mu.Lock()
		tets.requests++
		tets.lastBody = body
		status := tets.status
		respBody := append([]byte(nil), tets.respBody...)
		tets.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(respBody)
	}))
	return tets
}

func (e *emailTestServer) SetResponse(status int, body []byte) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.status = status
	e.respBody = append([]byte(nil), body...)
}

func (e *emailTestServer) RequestCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.requests
}

func (e *emailTestServer) LastBody() []byte {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.lastBody
}

func setupServiceConfig() {
	serviceutils.SetEnv("dev")
	serviceutils.SetMailConfig(serviceutils.MailConfig{Info: "info@example.com"})
	serviceutils.SetServiceUrlConfig(serviceutils.UrlConfig{UserServiceDomain: "https://user.example.com"})
}

func setupEkycTestDB(t *testing.T, user domain.User, profile domain.PrivateProfile, log domain.EkycRequestLog) (context.Context, *xorm.Engine) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)

	err = engine.Sync(new(domain.User), new(domain.PrivateProfile), new(domain.EkycRequestLog), new(domain.UserVirtualAccount))
	require.NoError(t, err)

	_, err = engine.Insert(&user)
	require.NoError(t, err)

	_, err = engine.Insert(&profile)
	require.NoError(t, err)

	_, err = engine.Insert(&log)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	return ctx, engine
}

func TestNexwayControllers_ekycAfterProcess(t *testing.T) {
	setupServiceConfig()

	now := jsonutils.JSONTime(time.Now())

	// MIR-1223: nexwayへの照会は受信リクエストのctxから独立させたことを
	// 検証するためのfake。呼び出し時点でctxがDone()だったかを記録する。
	fakeCtxIndependentSvc := &fakeNexwayService{
		output: &service.EkycResultOutput{
			ApplicantId: 1801,
			ReferenceId: "ekyc-ctx-independent",
			Status:      "completed",
			Result:      "match",
			RiskDetect:  &domain.NexwayRiskDetect{Result: "cleared"},
		},
	}

	tests := []struct {
		name                      string
		user                      domain.User
		profile                   domain.PrivateProfile
		log                       domain.EkycRequestLog
		result                    NexwayEkycResult
		setup                     func(ctx context.Context, engine *xorm.Engine)
		postCheck                 func(t *testing.T, engine *xorm.Engine)
		expectAuditStatus         domain.AUDIT_STATUS
		expectSignupStatus        domain.SIGNUP_STATUS
		expectHoldNil             bool
		expectEmailCount          int
		expectErrorMessage        *string
		historyErrors             []error
		wantErr                   bool
		wantErrMessage            string
		wantErrContains           string
		expectBeforeCalls         int
		expectCreateHistory       bool
		checkUser                 bool
		emailServerSetup          func(server *emailTestServer)
		prepare                   func(t *testing.T, ctx context.Context, engine *xorm.Engine) context.Context
		historyConcurrentMutation func(t *testing.T, engine *xorm.Engine)
		verifyWithNexway          bool
		nexwayService             *fakeNexwayService
	}{
		{
			name: "本人確認OK",
			user: domain.User{
				Id:           1,
				Uid:          "uid-1",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-1",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Cleared User",
				Email:        "cleared@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
			},
			profile: domain.PrivateProfile{
				Id:         1,
				UserId:     1,
				FirstName:  "Cleared",
				FamilyName: "User",
				RequestId:  "req-cleared",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          1,
				Email:       "cleared@example.com",
				UserId:      1,
				EkycId:      "ekyc-cleared",
				ApplicantId: 100,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-cleared",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			// MIR-1373/1374(設計v1、10/8): 書き込みをuserの4項目だけに絞ったため、
			// private_profileとuser.emailが書き換わらないことを確認する。
			// GetDataForPrivateUser(会員を読んだ直後、更新の確定前に呼ばれる)の
			// 中で直接DBを書き換え、ekycAfterProcessと同じ秒に会員のログイン・
			// 本人情報の変更が割り込んだ状況を再現する。旧実装(行を丸ごと書き戻す
			// Update)であれば、ここで割り込んだ変更は読み込み時点の古い値へ
			// 戻されてしまう。postCheckでは割り込んだ側の値が残っていることを
			// 確認することで、その違いを実際に検出できるテストにする(②レビューで
			// 指摘。以前はpostCheckが「誰も変えていない値が変わらない」ことしか
			// 確認しておらず、旧実装でも新実装でも通ってしまっていた)。
			historyConcurrentMutation: func(t *testing.T, engine *xorm.Engine) {
				_, err := engine.Exec("UPDATE user SET email = ? WHERE id = ?", "changed-during-ekyc@example.com", int64(1))
				require.NoError(t, err)
				_, err = engine.Exec("UPDATE private_profile SET first_name = ? WHERE user_id = ?", "ChangedDuringEkyc", int64(1))
				require.NoError(t, err)
			},
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				var updatedUser domain.User
				has, err := engine.Table("user").Where("id = ?", int64(1)).Get(&updatedUser)
				require.NoError(t, err)
				require.True(t, has)
				require.Equal(t, "changed-during-ekyc@example.com", updatedUser.Email,
					"eKYC処理の途中で割り込んだemailの変更が、UpdateEkycResultによって古い値へ戻されないこと")

				var updatedProfile domain.PrivateProfile
				has, err = engine.Table("private_profile").Where("user_id = ?", int64(1)).Get(&updatedProfile)
				require.NoError(t, err)
				require.True(t, has)
				require.Equal(t, "ChangedDuringEkyc", updatedProfile.FirstName,
					"eKYC処理がprivate_profileへは一切書き込まないため、割り込んだ変更がそのまま残ること")
				require.Equal(t, "User", updatedProfile.FamilyName)
				require.Equal(t, "req-cleared", updatedProfile.RequestId)
			},
			expectAuditStatus:   domain.WAIT_FINAL_APPROVAL_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
		},
		{
			// MIR-1374: eKYC申込み後に会員がメールアドレスを変えたケース。
			// ekyc_request_log.Email(申込み時点のアドレス)がuser.Email(現在の
			// アドレス)と食い違っていても、UserIdで照会するため結果が
			// 正しく反映されること。
			name: "本人確認OK_申込み後にメールアドレスを変えた会員でも結果が反映される",
			user: domain.User{
				Id:           2,
				Uid:          "uid-2",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-2",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Changed Email User",
				Email:        "changed-new@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
			},
			profile: domain.PrivateProfile{
				Id:         2,
				UserId:     2,
				FirstName:  "Changed",
				FamilyName: "Email",
				RequestId:  "req-changed-email",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          2,
				Email:       "changed-old@example.com",
				UserId:      2,
				EkycId:      "ekyc-changed-email",
				ApplicantId: 200,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   200,
				ReferenceId:   "ekyc-changed-email",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			// MIR-1374: 修正前の本当の危険は「見つからない」ことよりも「別の
			// 会員に反映される」ことだった。旧アドレス(changed-old@example.com)
			// を、いま別の会員(Id=3、本人確認待ちの別の申請中)が使っている
			// 状態を作り、その会員が今回のWebhookの影響を受けないことまで
			// 確認する。
			setup: func(ctx context.Context, engine *xorm.Engine) {
				bystander := domain.User{
					Id:           3,
					Uid:          "uid-3",
					LoginType:    domain.MAIL,
					UserType:     domain.PRIVATE,
					CommonId:     "common-3",
					InvestorType: domain.PRIVATE_INVESTOR,
					Name:         "Bystander User",
					Email:        "changed-old@example.com",
					Password:     "pass",
					Salt:         "salt",
					SignUpDate:   now,
					AuditStatus:  domain.WAIT_AUDIT,
					SignupStatus: domain.SIGNUP_NONE,
					IsEkyc:       false,
				}
				_, err := engine.Insert(&bystander)
				require.NoError(t, err)

				bystanderProfile := domain.PrivateProfile{
					Id:         3,
					UserId:     3,
					FirstName:  "Bystander",
					FamilyName: "User",
					RequestId:  "req-bystander",
					Birthday:   now,
				}
				_, err = engine.Insert(&bystanderProfile)
				require.NoError(t, err)
			},
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				var bystander domain.User
				has, err := engine.Table("user").Where("id = ?", int64(3)).Get(&bystander)
				require.NoError(t, err)
				require.True(t, has)
				require.Equal(t, domain.WAIT_AUDIT, bystander.AuditStatus, "旧アドレスを持つ別会員の審査状態が変わらないこと")
				require.Equal(t, domain.SIGNUP_NONE, bystander.SignupStatus, "旧アドレスを持つ別会員のSignupStatusが変わらないこと")
				require.False(t, bystander.IsEkyc, "旧アドレスを持つ別会員にIsEkycが立たないこと")
			},
			expectAuditStatus:   domain.WAIT_FINAL_APPROVAL_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
		},
		{
			// MIR-1222: 本登録承認完了済みの会員にclearedのWebhookが再送されても、
			// 審査ステータスをWAIT_FINAL_APPROVAL_AUDITへ巻き戻さないこと。
			// IsEkyc・SignupStatusは本人確認結果として更新されることは変わらない。
			name: "本人確認OK_承認完了済みなら審査ステータスを戻さない",
			user: domain.User{
				Id:           101,
				Uid:          "uid-101",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-101",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Completed User",
				Email:        "completed@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.COMPLETED_AUDIT,
				SignupStatus: domain.SIGNUP_COMPLETE,
				IsEkyc:       false,
			},
			profile: domain.PrivateProfile{
				Id:         101,
				UserId:     101,
				FirstName:  "Completed",
				FamilyName: "User",
				RequestId:  "req-completed",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          101,
				Email:       "completed@example.com",
				UserId:      101,
				EkycId:      "ekyc-completed",
				ApplicantId: 10100,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   10100,
				ReferenceId:   "ekyc-completed",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			expectAuditStatus:   domain.COMPLETED_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
			// AuditStatusだけを据え置いていることを固定するため、IsEkycは初期値
			// falseから常に更新されてtrueになることも確認する(MIR-1222)。
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				var updatedUser domain.User
				has, err := engine.Table("user").Where("id = ?", int64(101)).Get(&updatedUser)
				require.NoError(t, err)
				require.True(t, has)
				require.True(t, updatedUser.IsEkyc, "AuditStatusを据え置いてもIsEkycは更新されること")
			},
		},
		{
			// MIR-1222: 無効(INVALIDITY_AUDIT)の会員も、承認完了済みと同様に
			// 審査ステータスを戻すべきではない。
			name: "本人確認OK_無効化済みなら審査ステータスを戻さない",
			user: domain.User{
				Id:           102,
				Uid:          "uid-102",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-102",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Invalidity User",
				Email:        "invalidity@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.INVALIDITY_AUDIT,
				SignupStatus: domain.SIGNUP_COMPLETE,
				IsEkyc:       false,
			},
			profile: domain.PrivateProfile{
				Id:         102,
				UserId:     102,
				FirstName:  "Invalidity",
				FamilyName: "User",
				RequestId:  "req-invalidity",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          102,
				Email:       "invalidity@example.com",
				UserId:      102,
				EkycId:      "ekyc-invalidity",
				ApplicantId: 10200,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   10200,
				ReferenceId:   "ekyc-invalidity",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			expectAuditStatus:   domain.INVALIDITY_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				var updatedUser domain.User
				has, err := engine.Table("user").Where("id = ?", int64(102)).Get(&updatedUser)
				require.NoError(t, err)
				require.True(t, has)
				require.True(t, updatedUser.IsEkyc, "AuditStatusを据え置いてもIsEkycは更新されること")
			},
		},
		{
			// MIR-1276①: 反社確認中(CHECK_ANTISOCIAL_AUDIT)の会員が、eKYCの
			// Webhookへ自分でCLEARED(cleared)を送るだけで、反社チェックの結果を
			// 待たずに最終承認待機まで進められてしまっていた。審査ステータスは
			// 据え置き、IsEkyc・SignupStatusは本人確認結果として更新される
			// ことを確認する。
			name: "本人確認OK_反社確認中なら審査ステータスを進めない",
			user: domain.User{
				Id:           1276,
				Uid:          "uid-1276",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-1276",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Check Antisocial User",
				Email:        "check-antisocial@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.CHECK_ANTISOCIAL_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
				IsEkyc:       false,
			},
			profile: domain.PrivateProfile{
				Id:         1276,
				UserId:     1276,
				FirstName:  "Check",
				FamilyName: "Antisocial",
				RequestId:  "req-check-antisocial",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          1276,
				Email:       "check-antisocial@example.com",
				UserId:      1276,
				EkycId:      "ekyc-check-antisocial",
				ApplicantId: 127600,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   127600,
				ReferenceId:   "ekyc-check-antisocial",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			expectAuditStatus:   domain.CHECK_ANTISOCIAL_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				var updatedUser domain.User
				has, err := engine.Table("user").Where("id = ?", int64(1276)).Get(&updatedUser)
				require.NoError(t, err)
				require.True(t, has)
				require.True(t, updatedUser.IsEkyc, "AuditStatusを据え置いてもIsEkycは更新されること")
			},
		},
		{
			// MIR-1276①: 反社確認待ち(WAIT_ANTISOCIAL_AUDIT)も同様に進めない。
			name: "本人確認OK_反社確認待ちなら審査ステータスを進めない",
			user: domain.User{
				Id:           1277,
				Uid:          "uid-1277",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-1277",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Wait Antisocial User",
				Email:        "wait-antisocial@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_ANTISOCIAL_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
				IsEkyc:       false,
			},
			profile: domain.PrivateProfile{
				Id:         1277,
				UserId:     1277,
				FirstName:  "Wait",
				FamilyName: "Antisocial",
				RequestId:  "req-wait-antisocial",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          1277,
				Email:       "wait-antisocial@example.com",
				UserId:      1277,
				EkycId:      "ekyc-wait-antisocial",
				ApplicantId: 127700,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   127700,
				ReferenceId:   "ekyc-wait-antisocial",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			expectAuditStatus:   domain.WAIT_ANTISOCIAL_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				var updatedUser domain.User
				has, err := engine.Table("user").Where("id = ?", int64(1277)).Get(&updatedUser)
				require.NoError(t, err)
				require.True(t, has)
				require.True(t, updatedUser.IsEkyc, "AuditStatusを据え置いてもIsEkycは更新されること")
			},
		},
		{
			// MIR-1222: 反社チェックNG(DETECTED)経路は、CLEARED・UNMATCHとは異なり
			// 承認完了済みの会員でも意図的にガードしない。確定済みだからと握り
			// つぶすと、管理者が別経路で先にCOMPLETED_AUDITへ進めた直後に届いた
			// 本物の初回反社検知まで一緒に捨ててしまうため。従来どおり
			// WAIT_ANTISOCIAL_AUDITへ進むことを固定する。
			name: "反社チェックNG_承認完了済みでも従来どおり反社待機へ進む",
			user: domain.User{
				Id:           103,
				Uid:          "uid-103",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-103",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Completed Detected User",
				Email:        "completed-detected@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.COMPLETED_AUDIT,
				SignupStatus: domain.SIGNUP_COMPLETE,
				IsEkyc:       false,
			},
			profile: domain.PrivateProfile{
				Id:         103,
				UserId:     103,
				FirstName:  "Completed",
				FamilyName: "Detected",
				RequestId:  "req-completed-detected",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          103,
				Email:       "completed-detected@example.com",
				UserId:      103,
				EkycId:      "ekyc-completed-detected",
				ApplicantId: 10300,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   10300,
				ReferenceId:   "ekyc-completed-detected",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_DETECTED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			expectAuditStatus:   domain.WAIT_ANTISOCIAL_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				var updatedUser domain.User
				has, err := engine.Table("user").Where("id = ?", int64(103)).Get(&updatedUser)
				require.NoError(t, err)
				require.True(t, has)
				require.True(t, updatedUser.IsEkyc, "IsEkycは更新されること")
			},
		},
		{
			// MIR-1222: DETECTED経路でも、INVALIDITY_AUDIT(既に反社で無効化済み)の
			// 会員はガードする。COMPLETED_AUDITとは逆に、無効化状態から動かさない
			// ことが安全側になる。user-api側の複数箇所がAuditStatus==
			// INVALIDITY_AUDITそのものをログイン・本人確認のブロック条件にして
			// いるため、ここでWAIT_ANTISOCIAL_AUDITへ進めてしまうとその専用の
			// ブロックが外れてしまう。
			name: "反社チェックNG_無効化済みなら審査ステータスを戻さない",
			user: domain.User{
				Id:           104,
				Uid:          "uid-104",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-104",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Invalidity Detected User",
				Email:        "invalidity-detected@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.INVALIDITY_AUDIT,
				SignupStatus: domain.SIGNUP_COMPLETE,
				IsEkyc:       false,
			},
			profile: domain.PrivateProfile{
				Id:         104,
				UserId:     104,
				FirstName:  "Invalidity",
				FamilyName: "Detected",
				RequestId:  "req-invalidity-detected",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          104,
				Email:       "invalidity-detected@example.com",
				UserId:      104,
				EkycId:      "ekyc-invalidity-detected",
				ApplicantId: 10400,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   10400,
				ReferenceId:   "ekyc-invalidity-detected",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_DETECTED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			expectAuditStatus:   domain.INVALIDITY_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				var updatedUser domain.User
				has, err := engine.Table("user").Where("id = ?", int64(104)).Get(&updatedUser)
				require.NoError(t, err)
				require.True(t, has)
				require.True(t, updatedUser.IsEkyc, "AuditStatusを据え置いてもIsEkycは更新されること")
			},
		},
		{
			name: "反社チェックNG",
			user: domain.User{
				Id:           2,
				Uid:          "uid-2",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-2",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Test User",
				Email:        "test-det@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
			},
			profile: domain.PrivateProfile{
				Id:         2,
				UserId:     2,
				FirstName:  "Test",
				FamilyName: "User",
				RequestId:  "req-det",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          2,
				Email:       "test-det@example.com",
				UserId:      2,
				EkycId:      "ekyc-det",
				ApplicantId: 100,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-det",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_DETECTED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			expectAuditStatus:   domain.WAIT_ANTISOCIAL_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
		},
		{
			name: "本人確認NG",
			user: domain.User{
				Id:           3,
				Uid:          "uid-3",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-3",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Retry User",
				Email:        "test-unmatch@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_COMPLETE,
			},
			profile: domain.PrivateProfile{
				Id:         3,
				UserId:     3,
				FirstName:  "Retry",
				FamilyName: "User",
				RequestId:  "req-unmatch",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          3,
				Email:       "test-unmatch@example.com",
				UserId:      3,
				EkycId:      "ekyc-unmatch",
				ApplicantId: 200,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_UNMATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:         200,
				ReferenceId:         "ekyc-unmatch",
				OverallResult:       string(domain.EKYC_OVERALL_RESULT_STATUS_UNMATCH),
				Result:              string(domain.EKYC_RESULT_STATUS_UNMATCH),
				VerificationErrors:  []NexwayEkycError{},
				OcrValidationErrors: []NexwayEkycError{},
			},
			expectAuditStatus:   domain.HOLD_AUDIT,
			expectSignupStatus:  domain.SIGNUP_NONE,
			expectHoldNil:       false,
			expectEmailCount:    1,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
		},
		{
			// MIR-1222: 本人確認NG(UNMATCH)経路でも、承認完了済みの会員は本登録保留
			// (HOLD_AUDIT)へ落とさず、登録状態のリセットも再本登録メールの送信も
			// 行わない。Webhookの再送(古いunmatch通知の再配信など)で、既に
			// 承認済みの会員が登録をやり直す羽目になる事象を防ぐ。
			name: "本人確認NG_承認完了済みなら巻き戻さない",
			user: domain.User{
				Id:           104,
				Uid:          "uid-104",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-104",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Completed Unmatch User",
				Email:        "completed-unmatch@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.COMPLETED_AUDIT,
				SignupStatus: domain.SIGNUP_COMPLETE,
			},
			profile: domain.PrivateProfile{
				Id:         104,
				UserId:     104,
				FirstName:  "Completed",
				FamilyName: "Unmatch",
				RequestId:  "req-completed-unmatch",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          104,
				Email:       "completed-unmatch@example.com",
				UserId:      104,
				EkycId:      "ekyc-completed-unmatch",
				ApplicantId: 10400,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   10400,
				ReferenceId:   "ekyc-completed-unmatch",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_UNMATCH),
				Result:        string(domain.EKYC_RESULT_STATUS_UNMATCH),
				VerificationErrors: []NexwayEkycError{
					{ErrorCode: "Q1", ErrorMessage: "顔照合結果がアンマッチ"},
				},
				OcrValidationErrors: []NexwayEkycError{
					{ErrorCode: "E1", ErrorMessage: "免許証番号の不整合"},
				},
			},
			expectAuditStatus:  domain.COMPLETED_AUDIT,
			expectSignupStatus: domain.SIGNUP_COMPLETE,
			expectHoldNil:      true,
			expectEmailCount:   0,
			// AuditStatusの巻き戻しはガードされるが、ekycRequestへのエラー
			// メッセージ記録は確定済みかどうかに関わらず行われることを、
			// 実際にエラー内容を入れて確認する(空スライスだと素通りしてしまい
			// この不変条件を検証できない)。
			expectErrorMessage: func() *string {
				s := "顔照合結果がアンマッチ\n免許証番号の不整合\n"
				return &s
			}(),
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
		},
		{
			name: "本人確認NG_エラーあり",
			user: domain.User{
				Id:           4,
				Uid:          "uid-4",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-4",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Retry User",
				Email:        "test-unmatch-error@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_COMPLETE,
			},
			profile: domain.PrivateProfile{
				Id:         4,
				UserId:     4,
				FirstName:  "Retry",
				FamilyName: "User",
				RequestId:  "req-unmatch-error",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          4,
				Email:       "test-unmatch-error@example.com",
				UserId:      4,
				EkycId:      "ekyc-unmatch-error",
				ApplicantId: 400,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_UNMATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   400,
				ReferenceId:   "ekyc-unmatch-error",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_UNMATCH),
				Result:        string(domain.EKYC_RESULT_STATUS_UNMATCH),
				VerificationErrors: []NexwayEkycError{
					{ErrorCode: "Q1", ErrorMessage: "顔照合結果がアンマッチ"},
				},
				OcrValidationErrors: []NexwayEkycError{
					{ErrorCode: "E1", ErrorMessage: "免許証番号の不整合"},
				},
			},
			expectAuditStatus:  domain.HOLD_AUDIT,
			expectSignupStatus: domain.SIGNUP_NONE,
			expectHoldNil:      false,
			expectEmailCount:   1,
			expectErrorMessage: func() *string {
				s := "顔照合結果がアンマッチ\n免許証番号の不整合\n"
				return &s
			}(),
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
			wantErr:             false,
		},
		{
			name: "eKYCリクエスト取得エラー",
			user: domain.User{
				Id:           9,
				Uid:          "uid-9",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-9",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Log Error User",
				Email:        "log-error@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
			},
			profile: domain.PrivateProfile{
				Id:         9,
				UserId:     9,
				FirstName:  "Log",
				FamilyName: "Error",
				RequestId:  "req-log-error",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          9,
				Email:       "log-error@example.com",
				UserId:      9,
				EkycId:      "ekyc-log-error",
				ApplicantId: 900,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   900,
				ReferenceId:   "ekyc-log-error",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			setup: func(ctx context.Context, engine *xorm.Engine) {
				_ = engine.Close()
			},
			expectAuditStatus:   domain.WAIT_AUDIT,
			expectSignupStatus:  domain.SIGNUP_NONE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			historyErrors:       nil,
			wantErr:             true,
			wantErrContains:     "database is closed",
			expectBeforeCalls:   0,
			expectCreateHistory: false,
			checkUser:           false,
		},
		{
			// MIR-1374: 以前はekycRequest.Emailで照会していたため、この
			// テストケースは「log.Emailがどの会員のuser.Emailとも一致しない」
			// ことで未登録状態を作っていた(user自体はId=10として存在していた)。
			// UserIdでの照会に直した後は、それでは未登録を再現できなくなった
			// ため、log.UserIdが実在しない会員IDを指す形に変えて、本来の
			// 「該当会員が存在しない」ケースとして維持する。
			// 設計v1(MIR-1373/1374、10/8):会員IDでも見つからない場合は、
			// ネクスウェイに再送させても状況が変わらないため、エラーではなく
			// nil(200扱い)を返すように変更した。wantErr:falseに合わせて
			// 更新済み(以前はtrue・"User not found")。
			name: "ユーザー未登録",
			user: domain.User{
				Id:           10,
				Uid:          "uid-10",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-10",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Exists User",
				Email:        "exists@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
			},
			profile: domain.PrivateProfile{
				Id:         10,
				UserId:     10,
				FirstName:  "Exists",
				FamilyName: "User",
				RequestId:  "req-exists",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          10,
				Email:       "missing@example.com",
				UserId:      9999,
				EkycId:      "ekyc-missing",
				ApplicantId: 1000,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   1000,
				ReferenceId:   "ekyc-missing",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			expectAuditStatus:   domain.WAIT_AUDIT,
			expectSignupStatus:  domain.SIGNUP_NONE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			historyErrors:       nil,
			wantErr:             false,
			expectBeforeCalls:   0,
			expectCreateHistory: false,
			checkUser:           false,
		},
		{
			name: "履歴取得エラー_before",
			user: domain.User{
				Id:           12,
				Uid:          "uid-12",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-12",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "History User",
				Email:        "history@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
			},
			profile: domain.PrivateProfile{
				Id:         12,
				UserId:     12,
				FirstName:  "History",
				FamilyName: "User",
				RequestId:  "req-history",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          12,
				Email:       "history@example.com",
				UserId:      12,
				EkycId:      "ekyc-history",
				ApplicantId: 1200,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   1200,
				ReferenceId:   "ekyc-history",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			expectAuditStatus:   domain.WAIT_AUDIT,
			expectSignupStatus:  domain.SIGNUP_NONE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			historyErrors:       []error{errors.New("before error")},
			wantErr:             true,
			wantErrMessage:      "before error",
			expectBeforeCalls:   1,
			expectCreateHistory: false,
			checkUser:           false,
		},
		{
			name: "本人確認NG_メール送信失敗",
			user: domain.User{
				Id:           14,
				Uid:          "uid-14",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-14",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Mail Fail User",
				Email:        "mail-fail@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_COMPLETE,
			},
			profile: domain.PrivateProfile{
				Id:         14,
				UserId:     14,
				FirstName:  "Mail",
				FamilyName: "Fail",
				RequestId:  "req-mail-fail",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          14,
				Email:       "mail-fail@example.com",
				UserId:      14,
				EkycId:      "ekyc-mail-fail",
				ApplicantId: 1400,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_UNMATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   1400,
				ReferenceId:   "ekyc-mail-fail",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_UNMATCH),
				Result:        string(domain.EKYC_RESULT_STATUS_UNMATCH),
				VerificationErrors: []NexwayEkycError{
					{ErrorCode: "Q1", ErrorMessage: "顔照合結果がアンマッチ"},
				},
				OcrValidationErrors: []NexwayEkycError{
					{ErrorCode: "E1", ErrorMessage: "免許証番号の不整合"},
				},
			},
			expectAuditStatus:   domain.WAIT_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    1,
			expectErrorMessage:  nil,
			historyErrors:       nil,
			wantErr:             true,
			wantErrContains:     "fail create mail retry",
			expectBeforeCalls:   1,
			expectCreateHistory: false,
			checkUser:           false,
			emailServerSetup: func(server *emailTestServer) {
				server.SetResponse(http.StatusInternalServerError, []byte("error"))
			},
		},
		{
			name: "本人確認NG_UUID生成失敗",
			user: domain.User{
				Id:           15,
				Uid:          "uid-15",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-15",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "UUID Fail User",
				Email:        "uuid-fail@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_COMPLETE,
			},
			profile: domain.PrivateProfile{
				Id:         15,
				UserId:     15,
				FirstName:  "UUID",
				FamilyName: "Fail",
				RequestId:  "req-uuid-fail",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          15,
				Email:       "uuid-fail@example.com",
				UserId:      15,
				EkycId:      "ekyc-uuid-fail",
				ApplicantId: 1500,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_UNMATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   1500,
				ReferenceId:   "ekyc-uuid-fail",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_UNMATCH),
				Result:        string(domain.EKYC_RESULT_STATUS_UNMATCH),
			},
			expectAuditStatus:   domain.WAIT_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			historyErrors:       nil,
			wantErr:             true,
			wantErrContains:     "unexpected EOF",
			expectBeforeCalls:   1,
			expectCreateHistory: false,
			checkUser:           false,
			prepare: func(t *testing.T, ctx context.Context, engine *xorm.Engine) context.Context {
				original := uuid.DefaultGenerator
				uuid.DefaultGenerator = uuid.NewGenWithOptions(uuid.WithRandomReader(failingReader{err: io.ErrUnexpectedEOF}))
				t.Cleanup(func() { uuid.DefaultGenerator = original })
				return ctx
			},
		},
		{
			name: "履歴取得エラー_after",
			user: domain.User{
				Id:           13,
				Uid:          "uid-13",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-13",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "History User",
				Email:        "history-after@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
			},
			profile: domain.PrivateProfile{
				Id:         13,
				UserId:     13,
				FirstName:  "History",
				FamilyName: "User",
				RequestId:  "req-history-after",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          13,
				Email:       "history-after@example.com",
				UserId:      13,
				EkycId:      "ekyc-history-after",
				ApplicantId: 1300,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   1300,
				ReferenceId:   "ekyc-history-after",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			expectAuditStatus:   domain.WAIT_AUDIT,
			expectSignupStatus:  domain.SIGNUP_NONE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			historyErrors:       []error{nil, errors.New("after error")},
			wantErr:             true,
			wantErrMessage:      "after error",
			expectBeforeCalls:   2,
			expectCreateHistory: false,
			checkUser:           false,
			setup: func(ctx context.Context, engine *xorm.Engine) {
				va := domain.UserVirtualAccount{
					VaBankName:       "TestBank",
					VaBranchCode:     "001",
					VaBranchName:     "TestBranch",
					VaAccountNumber:  "0000013",
					VaHolderNameKana: "ｴﾈﾙｷﾞｰｼｪｱｽ",
					ProjectType:      domain.EQUITY,
					LinkedStatus:     domain.OFF_VA_ACCOUNT,
					UpdateDate:       jsonutils.JSONTime(time.Now()),
					Created:          jsonutils.JSONTime(time.Now()),
					Updated:          jsonutils.JSONTime(time.Now()),
				}
				_, err := engine.Insert(&va)
				require.NoError(t, err)
			},
		},
		{
			// MIR-1223: verifyWithNexway=trueのとき、nexway照会結果と整合する
			// CLEARED Webhookは通常どおり処理される。
			name: "MIR-1223_検証あり_照会結果と整合するなら通常どおり進む",
			user: domain.User{
				Id:           16,
				Uid:          "uid-16",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-16",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Verified Cleared User",
				Email:        "verified-cleared@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
			},
			profile: domain.PrivateProfile{
				Id:         16,
				UserId:     16,
				FirstName:  "Verified",
				FamilyName: "Cleared",
				RequestId:  "req-verified-cleared",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          16,
				Email:       "verified-cleared@example.com",
				UserId:      16,
				EkycId:      "ekyc-verified-cleared",
				ApplicantId: 1600,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   1600,
				ReferenceId:   "ekyc-verified-cleared",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			verifyWithNexway: true,
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 1600,
					ReferenceId: "ekyc-verified-cleared",
					Status:      "completed",
					Result:      "match",
					RiskDetect:  &domain.NexwayRiskDetect{Result: "cleared"},
				},
			},
			expectAuditStatus:   domain.WAIT_FINAL_APPROVAL_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
		},
		{
			// MIR-1223: verifyWithNexway=trueのとき、照会結果と矛盾するCLEARED
			// Webhook(自己昇格の詐称)は弾かれ、後続のユーザー更新は行われない。
			name: "MIR-1223_検証あり_照会結果と矛盾するなら弾かれ後続処理は走らない",
			user: domain.User{
				Id:           17,
				Uid:          "uid-17",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-17",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Self Escalation User",
				Email:        "self-escalation@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_ANTISOCIAL_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
			},
			profile: domain.PrivateProfile{
				Id:         17,
				UserId:     17,
				FirstName:  "Self",
				FamilyName: "Escalation",
				RequestId:  "req-self-escalation",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          17,
				Email:       "self-escalation@example.com",
				UserId:      17,
				EkycId:      "ekyc-self-escalation",
				ApplicantId: 1700,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_UNMATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   1700,
				ReferenceId:   "ekyc-self-escalation",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED), // 詐称
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),           // 詐称
			},
			verifyWithNexway: true,
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 1700,
					ReferenceId: "ekyc-self-escalation",
					Status:      "completed",
					Result:      "unmatch", // nexway側の真の値は本人確認NG
					RiskDetect:  &domain.NexwayRiskDetect{Result: "detected"},
				},
			},
			wantErr:             true,
			wantErrContains:     "reject",
			expectBeforeCalls:   0,
			expectCreateHistory: false,
			checkUser:           false,
		},
		{
			// なりすましによる不当な降格(嫌がらせ・DoS)を防ぐケース。実際は
			// CLEARED済みの会員に対して偽のUNMATCH Webhookが送られても、
			// 照会結果と矛盾するため弾かれ、HOLD_AUDITへの降格・
			// SIGNUP_NONEへのリセット・再本登録メール送信は起きない。
			name: "MIR-1223_検証あり_UNMATCH詐称は弾かれ不当な降格は起きない",
			user: domain.User{
				Id:           18,
				Uid:          "uid-18",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-18",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Malicious Downgrade User",
				Email:        "malicious-downgrade@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_FINAL_APPROVAL_AUDIT,
				SignupStatus: domain.SIGNUP_COMPLETE,
			},
			profile: domain.PrivateProfile{
				Id:         18,
				UserId:     18,
				FirstName:  "Malicious",
				FamilyName: "Downgrade",
				RequestId:  "req-malicious-downgrade",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          18,
				Email:       "malicious-downgrade@example.com",
				UserId:      18,
				EkycId:      "ekyc-malicious-downgrade",
				ApplicantId: 1800,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   1800,
				ReferenceId:   "ekyc-malicious-downgrade",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_UNMATCH), // 詐称
				Result:        string(domain.EKYC_RESULT_STATUS_UNMATCH),         // 詐称
			},
			verifyWithNexway: true,
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 1800,
					ReferenceId: "ekyc-malicious-downgrade",
					Status:      "completed",
					Result:      "match", // nexway側の真の値は本人確認OK
				},
			},
			wantErr:             true,
			wantErrContains:     "reject",
			expectBeforeCalls:   0,
			expectCreateHistory: false,
			checkUser:           false,
		},
		{
			// 実際は本人確認に失敗した(unmatch)人物が、偽のDETECTEDを送って
			// IsEkyc=trueにされることで、マイナンバー承認画面
			// (apps/admin-api/controllers/my_number.go)経由の反社チェック
			// 未完了のままの本登録完了を狙おうとするケース。照会結果と矛盾
			// するため弾かれ、IsEkycは更新されない。
			name: "MIR-1223_検証あり_DETECTED詐称は弾かれ反社チェック未完了のままの本登録完了は起きない",
			user: domain.User{
				Id:           19,
				Uid:          "uid-19",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-19",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Address Check Bypass User",
				Email:        "address-check-bypass@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.HOLD_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
			},
			profile: domain.PrivateProfile{
				Id:         19,
				UserId:     19,
				FirstName:  "Address",
				FamilyName: "Bypass",
				RequestId:  "req-address-check-bypass",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          19,
				Email:       "address-check-bypass@example.com",
				UserId:      19,
				EkycId:      "ekyc-address-check-bypass",
				ApplicantId: 1900,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_UNMATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   1900,
				ReferenceId:   "ekyc-address-check-bypass",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_DETECTED), // 詐称
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),            // 詐称
			},
			verifyWithNexway: true,
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 1900,
					ReferenceId: "ekyc-address-check-bypass",
					Status:      "completed",
					Result:      "unmatch", // nexway側の真の値は本人確認NG
					RiskDetect:  &domain.NexwayRiskDetect{Result: "cleared"},
				},
			},
			wantErr:             true,
			wantErrContains:     "reject",
			expectBeforeCalls:   0,
			expectCreateHistory: false,
			checkUser:           false,
		},
		{
			// 実際は本人確認に失敗した(unmatch)人物が、本物の反社ヒットも
			// 併せ持つ(RiskDetect=detected)ケース。DETECTED主張の矛盾チェックは
			// RiskDetect.Result==detectedしか見ておらず本人確認結果自体は見て
			// いないため、この組み合わせでは弾かれずverified=trueになる
			// (レビュー指摘、2026-09-22)。「反社ヒットを見逃さない」という
			// 制約(MIR-1222)のため、AuditStatus=WAIT_ANTISOCIAL_AUDITへは
			// 正しく進めるが、本人確認結果(Result、照会結果でunmatchに補正
			// 済み)がmatchでないため、IsEkycは更新しない。
			name: "MIR-1223_検証あり_DETECTEDかつ本人確認結果はunmatchのときIsEkycを更新しない",
			user: domain.User{
				Id:           1901,
				Uid:          "uid-1901",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-1901",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Detected Unmatch User",
				Email:        "detected-unmatch@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
				IsEkyc:       false,
			},
			profile: domain.PrivateProfile{
				Id:         1901,
				UserId:     1901,
				FirstName:  "Detected",
				FamilyName: "Unmatch",
				RequestId:  "req-detected-unmatch",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          1901,
				Email:       "detected-unmatch@example.com",
				UserId:      1901,
				EkycId:      "ekyc-detected-unmatch",
				ApplicantId: 190100,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   190100,
				ReferenceId:   "ekyc-detected-unmatch",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_DETECTED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH), // Webhookの詐称。実際はunmatch
			},
			verifyWithNexway: true,
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 190100,
					ReferenceId: "ekyc-detected-unmatch",
					Status:      "completed",
					Result:      "unmatch", // 実際は本人確認NG(本物の反社ヒットと併発)
					RiskDetect:  &domain.NexwayRiskDetect{Result: "detected"},
				},
			},
			wantErr:             false,
			expectAuditStatus:   domain.WAIT_ANTISOCIAL_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				var updatedUser domain.User
				has, err := engine.Table("user").Where("id = ?", 1901).Get(&updatedUser)
				require.NoError(t, err)
				require.True(t, has)
				require.False(t, updatedUser.IsEkyc, "DETECTEDで本人確認結果がunmatchのときはIsEkycを更新しないこと")
			},
		},
		{
			// MIR-1223: nexwayの照会結果のResult(本人確認結果)が空の場合、
			// result.Resultは空に補正されIsEkycは更新されない(前のケースの
			// unmatchと同じ判定経路)。加えて、ekycRequest.Resultが実際に
			// DBへ空として反映されることまで確認する。verified=trueの分岐は
			// ゼロ値カラムをxormのデフォルトどおりUPDATE文から除外するため、
			// 対応しないとDBに前回の申請試行のResult(このテストではMATCH)が
			// 残ったままになり、「反社チェック待ちなのに本人確認済みと表示
			// される」矛盾表示が起きる(2026-09-24のレビューで指摘、MustCols
			// で"result"だけを対象にする対応で修正済み)。
			//
			// 既存ログのApplicantId(999999)とRiskDetectResult(cleared)は、
			// 今回の照会結果(190200・detected)とわざと異なる値にしてある。
			// MustColsを使うと"result"以外のカラムがUPDATE対象から外れて
			// しまう副作用が無いことを、これらが実際に新しい値へ更新される
			// ことで確認する(2026-09-24の追加レビューで指摘)。
			name: "MIR-1223_検証あり_DETECTEDかつ照会結果のResultが空のときIsEkycを更新せずekycRequest.Resultも空に反映される",
			user: domain.User{
				Id:           1902,
				Uid:          "uid-1902",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-1902",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Detected EmptyResult User",
				Email:        "detected-empty-result@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
				IsEkyc:       false,
			},
			profile: domain.PrivateProfile{
				Id:         1902,
				UserId:     1902,
				FirstName:  "Detected",
				FamilyName: "EmptyResult",
				RequestId:  "req-detected-empty-result",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:               1902,
				Email:            "detected-empty-result@example.com",
				UserId:           1902,
				EkycId:           "ekyc-detected-empty-result",
				ApplicantId:      999999, // 前回の申請試行の値(190200に更新されるべき)
				Status:           domain.EKYC_REQUEST_STATUS_NONE,
				Result:           domain.EKYC_RESULT_STATUS_MATCH,                // 前回の申請試行の値(残ってはいけない)
				RiskDetectResult: func() *string { s := "cleared"; return &s }(), // 前回の申請試行の値(detectedに更新されるべき)
				Created:          now,
			},
			result: NexwayEkycResult{
				ApplicantId:   190200,
				ReferenceId:   "ekyc-detected-empty-result",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_DETECTED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH), // Webhookの詐称
			},
			verifyWithNexway: true,
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 190200,
					ReferenceId: "ekyc-detected-empty-result",
					Status:      "completed",
					Result:      "", // 本人確認結果が空(反社チェックの実体はdetected)
					RiskDetect:  &domain.NexwayRiskDetect{Result: "detected"},
				},
			},
			wantErr:             false,
			expectAuditStatus:   domain.WAIT_ANTISOCIAL_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				var updatedUser domain.User
				has, err := engine.Table("user").Where("id = ?", 1902).Get(&updatedUser)
				require.NoError(t, err)
				require.True(t, has)
				require.False(t, updatedUser.IsEkyc, "DETECTEDで照会結果のResultが空のときはIsEkycを更新しないこと")

				var updatedLog domain.EkycRequestLog
				has, err = engine.Table("ekyc_request_log").Where("id = ?", 1902).Get(&updatedLog)
				require.NoError(t, err)
				require.True(t, has)
				require.Equal(t, domain.EKYC_RESULT_STATUS(""), updatedLog.Result, "ekycRequest.Resultが前回の申請試行の値(MATCH)のまま残らず空に更新されること")
				require.Equal(t, int64(190200), updatedLog.ApplicantId, "MustColsで\"result\"だけを対象にしても、ApplicantIdは通常どおり更新されること")
				require.NotNil(t, updatedLog.RiskDetectResult)
				require.Equal(t, "detected", *updatedLog.RiskDetectResult, "MustColsで\"result\"だけを対象にしても、RiskDetectResultは通常どおり更新されること")
			},
		},
		{
			// MIR-1223: DETECTED主張が通信失敗でフェイルオープンした場合、
			// 反社ヒットの記録(AuditStatus=WAIT_ANTISOCIAL_AUDIT)は残すが、
			// nexwayの裏付けが取れていないためIsEkycは更新しない。フェイル
			// オープンは本物のnexwayの一時的な不調でも起こり得るため、
			// IsEkycまで立てるとDETECTED詐称と同じ「反社チェック未完了の
			// ままの本登録完了到達」を再現できてしまう。また、フェイル
			// オープン後の「eKYC詳細結果取得」ブロックのスキップにより
			// 処理全体が正常終了することも合わせて確認する。resultの
			// ApplicantId/ResultをDB上の元の値とわざと異ならせており
			// (Webhookが別人9999のApplicantIdを騙る想定)、フェイルオープン
			// 時はこれらが検証未了として明示的にクリアされる(0・空文字)
			// ことも検証する(eKYCリクエスト行がStatusの更新されない使い
			// 回し設計であるため、「DBの既存値を維持する」実装では前回の
			// 申請試行の古い値を保ってしまうことが判明したための対応)。
			name: "MIR-1223_検証あり_DETECTEDフェイルオープン時はIsEkycを更新しない",
			user: domain.User{
				Id:           20,
				Uid:          "uid-20",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-20",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Fail Open User",
				Email:        "fail-open@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
				IsEkyc:       false,
			},
			profile: domain.PrivateProfile{
				Id:         20,
				UserId:     20,
				FirstName:  "Fail",
				FamilyName: "Open",
				RequestId:  "req-fail-open",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          20,
				Email:       "fail-open@example.com",
				UserId:      20,
				EkycId:      "ekyc-fail-open",
				ApplicantId: 2000,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_UNMATCH,
				// 前回の申請試行がUNMATCHだったときに保存されたエラー文言。
				// フェイルオープン時にResultだけクリアしてこれを残すと、
				// 「結果は空なのにエラー理由だけ表示される」矛盾が生まれる。
				ErrorMessage: func() *string { s := "顔照合結果がアンマッチ(前回試行分)"; return &s }(),
				// 詳細6フィールドも前回試行分の値として非nilを入れ、
				// クリア後にnilになることを検証する。
				EkycFrontMatch:       func() *string { s := "match"; return &s }(),
				EkycThicknessMatch:   func() *string { s := "match"; return &s }(),
				EkycLiveMatch:        func() *string { s := "match"; return &s }(),
				EkycEndAt:            &now,
				RiskDetectDetectedAt: &now,
				RiskDetectResult:     func() *string { s := "cleared"; return &s }(),
				Created:              now,
			},
			result: NexwayEkycResult{
				ApplicantId:   9999, // Webhookが別人のApplicantIdを騙る想定
				ReferenceId:   "ekyc-fail-open",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_DETECTED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			verifyWithNexway: true,
			nexwayService: &fakeNexwayService{
				err: pkgerrors.Wrap(service.ErrNexwayUnavailable, "connection refused"),
			},
			// verifyAndReconcileNexwayResultのフェイルオープン後、既存コード
			// 「eKYC詳細結果取得」ブロックはverified=falseのためスキップされる。
			// スキップせずこのブロックが同じnexwayServiceを再度呼び出す実装
			// だと、同じ通信エラーで失敗し、AuditStatus等は確定済みなのに
			// 処理全体がエラー終了してしまう。
			wantErr:             false,
			expectAuditStatus:   domain.WAIT_ANTISOCIAL_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				var updatedUser domain.User
				has, err := engine.Table("user").Where("id = ?", 20).Get(&updatedUser)
				require.NoError(t, err)
				require.True(t, has)
				require.False(t, updatedUser.IsEkyc, "フェイルオープン時はnexwayの裏付けが無いためIsEkycを更新しないこと")

				var updatedLog domain.EkycRequestLog
				hasLog, err := engine.Table("ekyc_request_log").Where("id = ?", 20).Get(&updatedLog)
				require.NoError(t, err)
				require.True(t, hasLog)
				require.Nil(t, updatedLog.RiskDetectResult, "「eKYC詳細結果取得」ブロックがスキップされ、裏付けの無い詳細フィールドは保存されないこと")
				require.Nil(t, updatedLog.RiskDetectDetectedAt, "前回試行分のRiskDetectDetectedAtが残らないこと")
				require.Nil(t, updatedLog.EkycFrontMatch, "前回試行分のEkycFrontMatchが残らないこと")
				require.Nil(t, updatedLog.EkycThicknessMatch, "前回試行分のEkycThicknessMatchが残らないこと")
				require.Nil(t, updatedLog.EkycLiveMatch, "前回試行分のEkycLiveMatchが残らないこと")
				require.Nil(t, updatedLog.EkycEndAt, "前回試行分のEkycEndAtが残らないこと")
				require.EqualValues(t, 0, updatedLog.ApplicantId, "フェイルオープン時は検証未了としてApplicantIdを明示的にクリアすること(前回試行の古い値を残さない)")
				require.Equal(t, domain.EKYC_RESULT_STATUS(""), updatedLog.Result, "フェイルオープン時は検証未了としてResultを明示的にクリアすること(前回試行の古い値を残さない)")
			},
		},
		{
			// MIR-1223: nexwayへの照会に渡るctxが、呼び出し元から渡されたctx
			// (このテストではDBセッションのctxで代用)のキャンセルを引き継がない
			// ことを検証する。このsqliteテスト環境ではDBセッションが受信ctxに
			// 紐づいていないため、ここでは処理全体が成功する。本番では
			// contextDBWithNameがDBセッションを受信ctxに紐づけているため、
			// 攻撃者が接続を切った場合、照会(独立ctx)は影響を受けず成功しても、
			// 後続のDB操作(受信ctx)は失敗しトランザクション全体がロールバック
			// される可能性がある(未実測)。それでもフェイルオープンを意図的に
			// 誘発できなくなる、という本来の目的は変わらない。
			name: "MIR-1223_検証あり_受信ctxが既にキャンセルされていても照会は影響を受けない",
			user: domain.User{
				Id:           18001,
				Uid:          "uid-18001",
				LoginType:    domain.MAIL,
				UserType:     domain.PRIVATE,
				CommonId:     "common-18001",
				InvestorType: domain.PRIVATE_INVESTOR,
				Name:         "Ctx Independent User",
				Email:        "ctx-independent@example.com",
				Password:     "pass",
				Salt:         "salt",
				SignUpDate:   now,
				AuditStatus:  domain.WAIT_AUDIT,
				SignupStatus: domain.SIGNUP_NONE,
			},
			profile: domain.PrivateProfile{
				Id:         18001,
				UserId:     18001,
				FirstName:  "Ctx",
				FamilyName: "Independent",
				RequestId:  "req-ctx-independent",
				Birthday:   now,
			},
			log: domain.EkycRequestLog{
				Id:          18001,
				Email:       "ctx-independent@example.com",
				UserId:      18001,
				EkycId:      "ekyc-ctx-independent",
				ApplicantId: 1801,
				Status:      domain.EKYC_REQUEST_STATUS_NONE,
				Result:      domain.EKYC_RESULT_STATUS_MATCH,
				Created:     now,
			},
			result: NexwayEkycResult{
				ApplicantId:   1801,
				ReferenceId:   "ekyc-ctx-independent",
				OverallResult: string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED),
				Result:        string(domain.EKYC_RESULT_STATUS_MATCH),
			},
			verifyWithNexway: true,
			nexwayService:    fakeCtxIndependentSvc,
			prepare: func(t *testing.T, ctx context.Context, engine *xorm.Engine) context.Context {
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()
				return cancelledCtx
			},
			wantErr:             false,
			expectAuditStatus:   domain.WAIT_FINAL_APPROVAL_AUDIT,
			expectSignupStatus:  domain.SIGNUP_COMPLETE,
			expectHoldNil:       true,
			expectEmailCount:    0,
			expectErrorMessage:  nil,
			expectBeforeCalls:   2,
			expectCreateHistory: true,
			checkUser:           true,
			postCheck: func(t *testing.T, engine *xorm.Engine) {
				require.False(t, fakeCtxIndependentSvc.receivedCancelledCtx, "nexwayへの照会は受信ctxのキャンセルから独立しているべき")
			},
		},
	}

	for _, tc := range tests {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			emailServer := newEmailTestServer()
			defer emailServer.Close()
			if testCase.emailServerSetup != nil {
				testCase.emailServerSetup(emailServer)
			}
			serviceutils.SetEmailServiceUrl(emailServer.URL)

			ctx, engine := setupEkycTestDB(t, testCase.user, testCase.profile, testCase.log)
			defer engine.Close()
			if testCase.prepare != nil {
				ctx = testCase.prepare(t, ctx, engine)
			}

			if testCase.setup != nil {
				setupFn := testCase.setup
				setupFn(ctx, engine)
			}

			fakeHistory := &fakeUserInfoHistoryService{getDataErrors: testCase.historyErrors}
			if testCase.historyConcurrentMutation != nil {
				fakeHistory.concurrentMutation = func() {
					testCase.historyConcurrentMutation(t, engine)
				}
			}
			controller := NexwayControllers{
				privateUserService:     &userservice.PrivateUserService{},
				userInfoHistoryService: fakeHistory,
			}
			// testCase.nexwayServiceがnilのまま代入すると、interfaceフィールドが
			// 「型は*fakeNexwayService・値はnil」の非nilインターフェースになり、
			// a.nexwayService != nilが誤ってtrueと評価されてpanicする(Goの
			// typed nilの罠)。nilのときはフィールドに触れず本当のnilのままにする。
			if testCase.nexwayService != nil {
				controller.nexwayService = testCase.nexwayService
			}

			err := controller.ekycAfterProcess(ctx, testCase.result, testCase.verifyWithNexway)
			if testCase.wantErr {
				require.Error(t, err)
				if testCase.wantErrMessage != "" {
					require.EqualError(t, err, testCase.wantErrMessage)
				}
				if testCase.wantErrContains != "" {
					require.Contains(t, err.Error(), testCase.wantErrContains)
				}
			} else {
				require.NoError(t, err)
			}

			if testCase.checkUser {
				var updatedUser domain.User
				has, err := engine.Table("user").Where("id = ?", testCase.user.Id).Get(&updatedUser)
				require.NoError(t, err)
				require.True(t, has)
				require.Equal(t, testCase.expectAuditStatus, updatedUser.AuditStatus)
				require.Equal(t, testCase.expectSignupStatus, updatedUser.SignupStatus)
				if testCase.expectHoldNil {
					require.Nil(t, updatedUser.HoldUserCheckCode)
				} else {
					require.NotNil(t, updatedUser.HoldUserCheckCode)
					require.NotEmpty(t, *updatedUser.HoldUserCheckCode)
				}

				updatedLog, err := (&domain.EkycRequestLog{}).GetByEkycId(ctx, testCase.log.EkycId)
				require.NoError(t, err)
				if testCase.expectErrorMessage == nil {
					require.Nil(t, updatedLog.ErrorMessage)
				} else {
					require.NotNil(t, updatedLog.ErrorMessage)
					require.Equal(t, *testCase.expectErrorMessage, *updatedLog.ErrorMessage)
				}
				if testCase.postCheck != nil {
					postFn := testCase.postCheck
					postFn(t, engine)
				}
			}

			require.Equal(t, testCase.expectBeforeCalls, fakeHistory.beforeCalls)
			require.Equal(t, testCase.expectCreateHistory, fakeHistory.createCalled)
			require.Equal(t, testCase.expectEmailCount, emailServer.RequestCount())
		})
	}
}

func TestNexwayControllers_verifyAndReconcileNexwayResult(t *testing.T) {
	tests := []struct {
		name          string
		result        NexwayEkycResult
		nexwayService *fakeNexwayService
		wantErr       bool
		wantResultVal string
		wantVerified  bool
	}{
		{
			name: "CLEARED主張・照会結果と整合するなら通過しResultを照会結果で上書きする",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "CLEARED",
				Result:        "unmatch", // Webhookが詐称した値
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "completed",
					Result:      "match", // nexway側の真の値
					RiskDetect:  &domain.NexwayRiskDetect{Result: "cleared"},
				},
			},
			wantErr:       false,
			wantResultVal: "match",
			wantVerified:  true,
		},
		{
			name: "CLEARED主張・ReferenceId不一致なら弾く(なりすましの可能性)",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "victim-ekyc-id", // 他人のReferenceIdを騙る
				OverallResult: "CLEARED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "attacker-ekyc-id", // 実際にはApplicantId=100自身の申請
					Result:      "match",
					RiskDetect:  &domain.NexwayRiskDetect{Result: "cleared"},
				},
			},
			wantErr: true,
		},
		{
			name: "CLEARED主張・照会結果はunmatchで矛盾するなら弾く(自己昇格の防止)",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "CLEARED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "completed",
					Result:      "unmatch", // 実際は本人確認NG
					RiskDetect:  &domain.NexwayRiskDetect{Result: "cleared"},
				},
			},
			wantErr: true,
		},
		{
			name: "CLEARED主張・反社チェックがdetectedで矛盾するなら弾く",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "CLEARED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "completed",
					Result:      "match",
					RiskDetect:  &domain.NexwayRiskDetect{Result: "detected"}, // 実際は反社ヒット
				},
			},
			wantErr: true,
		},
		{
			name: "UNMATCH主張・照会結果と整合するなら通過する",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "UNMATCH",
				Result:        "unmatch",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "completed",
					Result:      "unmatch",
				},
			},
			wantErr:       false,
			wantResultVal: "unmatch",
			wantVerified:  true,
		},
		{
			// なりすましによる不当な降格(HOLD_AUDITへの降格・SIGNUP_NONEへの
			// リセット・再本登録メール送信という嫌がらせ・DoS)を防ぐ。
			name: "UNMATCH主張・照会結果は実際にはmatchで矛盾するなら弾く(不当な降格の防止)",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "UNMATCH",
				Result:        "unmatch",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "completed",
					Result:      "match", // 実際は本人確認OK
				},
			},
			wantErr: true,
		},
		{
			name: "DETECTED主張・照会結果と整合するなら通過する(本物の反社ヒットは必ず通る)",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "DETECTED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "completed",
					Result:      "match",
					RiskDetect:  &domain.NexwayRiskDetect{Result: "detected"},
				},
			},
			wantErr:       false,
			wantResultVal: "match",
			wantVerified:  true,
		},
		{
			// DETECTED主張の矛盾チェックはRiskDetect.Result==detectedしか
			// 見ておらず、本人確認結果(output.Result)は見ていない。「反社
			// ヒットを見逃さない」という制約(MIR-1222)のため、本人確認結果が
			// unmatchでも反社チェックの実体がdetectedならこのチェック自体は
			// 通過しverified=trueになる(レビュー指摘、2026-09-22)。ただし
			// Resultはoutput.Result(実際にはunmatch)に補正されるため、
			// 呼び出し元(ekycAfterProcess)がverifiedだけでなくResultも見て
			// IsEkycを制御できるようになっている。
			name: "DETECTED主張・本人確認結果がunmatchでも反社チェックがdetectedなら通過する(verified=trueだがResultはunmatchに補正される)",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "DETECTED",
				Result:        "match", // Webhookの詐称。実際はunmatch
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "completed",
					Result:      "unmatch", // 実際は本人確認NG
					RiskDetect:  &domain.NexwayRiskDetect{Result: "detected"},
				},
			},
			wantErr:       false,
			wantResultVal: "unmatch",
			wantVerified:  true,
		},
		{
			// output.Result(本人確認結果)が空でも、本物の反社ヒット
			// (riskResult=detected)を握りつぶさないよう矛盾チェックは通す。
			// その代わりresult.Resultは無条件にoutput.Result(空)で上書き
			// されるため、Webhookの詐称値(match)は残らない。これにより
			// 呼び出し元(ekycAfterProcess)のresult.Result==matchという
			// IsEkyc判定は自動的にfalseになる(2026-09-22のレビューで指摘)。
			name: "DETECTED主張・照会結果のResultが空でも通過するがResultは空に補正される(詐称値を残さない)",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "DETECTED",
				Result:        "match", // Webhookの詐称
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "completed",
					Result:      "", // 反社チェックの実体はdetectedだが本人確認結果が空
					RiskDetect:  &domain.NexwayRiskDetect{Result: "detected"},
				},
			},
			wantErr:       false,
			wantResultVal: "",
			wantVerified:  true,
		},
		{
			// 実際は本人確認に失敗した(unmatch)人物が、反社チェック未完了の
			// まま本登録完了に到達する目的で偽のDETECTEDを送るケース
			// (マイナンバー承認画面apps/admin-api/controllers/my_number.goは
			// IsEkyc==trueなら会員の審査状態を見ずCOMPLETED_AUDITへ進めるため、
			// IsEkyc=trueにされること自体が実害になる)。
			name: "DETECTED主張・反社チェックの実体がclearedで矛盾するなら弾く(反社チェック未完了のままの本登録完了の防止)",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "DETECTED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "completed",
					Result:      "unmatch", // 実際は本人確認NG
					RiskDetect:  &domain.NexwayRiskDetect{Result: "cleared"},
				},
			},
			wantErr: true,
		},
		{
			// nexway側の仕様追加等でOverallResultに未知の値が来ても、弾くと
			// 本人確認処理が止まるため矛盾チェックをスキップして素通りする。
			name: "未知のOverallResultは矛盾チェックをスキップして素通りする",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "SOME_NEW_STATUS",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "completed",
					Result:      "unmatch", // 矛盾していても弾かれない
				},
			},
			wantErr:       false,
			wantResultVal: "unmatch",
			wantVerified:  true,
		},
		{
			// output.Statusがcompleted以外(または空)の場合は弾く(フェイルクローズ)。
			// 素通りにすると、攻撃者が「未完了の時間帯を狙って送信する」だけで
			// 矛盾チェック全体を回避できてしまうため。
			name: "CLEARED主張・照会結果のStatusが完了前なら弾く",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "CLEARED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "processing", // まだ完了していない
					Result:      "match",
					RiskDetect:  &domain.NexwayRiskDetect{Result: "cleared"},
				},
			},
			wantErr: true,
		},
		{
			// Statusが空(表示されないケース)も同様に弾く。
			name: "CLEARED主張・照会結果のStatusが空なら弾く",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "CLEARED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "",
					Result:      "match",
					RiskDetect:  &domain.NexwayRiskDetect{Result: "cleared"},
				},
			},
			wantErr: true,
		},
		{
			// DETECTED主張・Status未完了も同様に弾く(通信失敗時のフェイル
			// オープンとは別の経路であることを確認する)。
			name: "DETECTED主張・照会結果のStatusが完了前なら弾く",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "DETECTED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "processing",
					Result:      "match",
					RiskDetect:  &domain.NexwayRiskDetect{Result: "detected"},
				},
			},
			wantErr: true,
		},
		{
			// CLEARED/UNMATCH主張は通信失敗でもフェイルオープンにしない。
			// フェイルオープンにすると、nexwayが落ちている間、無認証のこの
			// エンドポイントに偽の結果を送るだけで裏取り無しに通ってしまう。
			name: "CLEARED主張・通信自体の失敗でもフェイルクローズで弾く",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "CLEARED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				err: pkgerrors.Wrap(service.ErrNexwayUnavailable, "connection refused"),
			},
			wantErr: true,
		},
		{
			name: "UNMATCH主張・通信自体の失敗でもフェイルクローズで弾く",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "UNMATCH",
				Result:        "unmatch",
			},
			nexwayService: &fakeNexwayService{
				err: pkgerrors.Wrap(service.ErrNexwayUnavailable, "connection refused"),
			},
			wantErr: true,
		},
		{
			// DETECTED主張だけが唯一の例外。通信失敗時は従来どおり
			// フェイルオープンでWebhookの値をそのまま使う。ただしverifiedは
			// falseを返し、呼び出し元(ekycAfterProcess)がIsEkycを更新しない
			// よう伝える。
			name: "DETECTED主張・通信自体の失敗はフェイルオープンでWebhookの値をそのまま使うがverified=false",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "DETECTED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				err: pkgerrors.Wrap(service.ErrNexwayUnavailable, "connection refused"),
			},
			wantErr:       false,
			wantResultVal: "match",
			wantVerified:  false,
		},
		{
			name: "応答はあるが不正(ErrNexwayInvalidResponse)ならフェイルクローズで弾く",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "CLEARED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				err: pkgerrors.Wrap(service.ErrNexwayInvalidResponse, "unexpected status code 502"),
			},
			wantErr: true,
		},
		{
			name: "照会結果にErrorCodeが含まれるなら弾く",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "CLEARED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					ErrorCode:   "E999",
				},
			},
			wantErr: true,
		},
		{
			name: "照会結果のReferenceIdが空なら弾く",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "CLEARED",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "",
					Result:      "match",
				},
			},
			wantErr: true,
		},
		{
			// output.Resultが(Status=completedであっても)何らかの理由で空に
			// なるケース。空のまま上書きするとekycRequest.Resultが空文字列で
			// 保存されてしまうため、空ならWebhookのbodyの値を保持する。矛盾
			// チェックをスキップする未知のOverallResultを使い、Result補正だけを
			// 切り分けて検証する。
			name: "照会結果のResultが空ならWebhookの値を保持する(空で上書きしない)",
			result: NexwayEkycResult{
				ApplicantId:   100,
				ReferenceId:   "ekyc-1",
				OverallResult: "SOME_NEW_STATUS",
				Result:        "match",
			},
			nexwayService: &fakeNexwayService{
				output: &service.EkycResultOutput{
					ApplicantId: 100,
					ReferenceId: "ekyc-1",
					Status:      "completed",
					Result:      "",
				},
			},
			wantErr:       false,
			wantResultVal: "match",
			wantVerified:  true,
		},
	}

	for _, tc := range tests {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			controller := NexwayControllers{nexwayService: testCase.nexwayService}

			got, verified, err := controller.verifyAndReconcileNexwayResult(context.Background(), testCase.result)
			require.Equal(t, testCase.wantVerified, verified, "verifiedがerrに関わらず改変で誤ってtrueに戻っていないことを確認する")
			if testCase.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, testCase.wantResultVal, got.Result)
		})
	}
}

func TestShouldAdvanceOnCleared(t *testing.T) {
	tests := []struct {
		name   string
		status domain.AUDIT_STATUS
		want   bool
	}{
		{"確定済み(COMPLETED_AUDIT)はfalse", domain.COMPLETED_AUDIT, false},
		{"確定済み(INVALIDITY_AUDIT)はfalse", domain.INVALIDITY_AUDIT, false},
		{"反社確認中(CHECK_ANTISOCIAL_AUDIT)はfalse", domain.CHECK_ANTISOCIAL_AUDIT, false},
		{"反社確認待ち(WAIT_ANTISOCIAL_AUDIT)はfalse", domain.WAIT_ANTISOCIAL_AUDIT, false},
		{"保留(HOLD_AUDIT)は今回は対象外のためtrue", domain.HOLD_AUDIT, true},
		{"最終承認待機はtrue", domain.WAIT_FINAL_APPROVAL_AUDIT, true},
		{"本登録承認待機はtrue", domain.WAIT_AUDIT, true},
		{"仮登録はtrue", domain.PROVISIONAL_AUDIT, true},
		{"本人確認待ちはtrue", domain.WAIT_IDENTITY_REGISTER_AUDIT, true},
		{"住所確認中(法人限定、eKYCの処理対象外だが網羅のため)はtrue", domain.CHECK_ADDRESS_AUDIT, true},
	}

	for _, tc := range tests {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.want, shouldAdvanceOnCleared(testCase.status))
		})
	}
}
