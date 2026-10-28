package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

type CorporateApiController struct {
	userInfoHistoryService service.IUserInfoHistoryService
}

func NewCorporateApiController() *CorporateApiController {
	return &CorporateApiController{
		userInfoHistoryService: service.NewUserInfoHistoryService(),
	}
}

func (c *CorporateApiController) Init(g echoswagger.ApiGroup) {

	g.POST("/:id/mobile/verification-request", c.RequestSMSVerification,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("대표자/담당자 휴대전화번호 인증 요청 ").
		AddParamBody(MobilePhoneVerificationInput{}, "body", "대표자/담당자 휴대전화번호 인증 요청", true).
		AddParamPath("", "id", "").
		SetDescription("대표자/담당자 휴대전화번호 인증을 요청 합니다.").
		SetSecurity("Authorization")

	g.POST("/:id/mobile/verification-re-request", c.ReRequestSMSVerification,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("대표자/담당자 휴대전화번호 인증 요청 ").
		AddParamBody(MobilePhoneVerificationInput{}, "body", "대표자/담당자 휴대전화번호 인증 재요청", true).
		AddParamPath("", "id", "").
		SetDescription("대표자/담당자 휴대전화번호 인증을 재요청 합니다.").
		SetSecurity("Authorization")

	g.POST("/:id/mobile/verification/:corporation_type", c.MobilePhoneVerificationConfirm,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("대표자/담당자 휴대전화번호 인증 확인").
		AddParamBody(MobilePhoneVerificationResponse{}, "body", "대표자/담당자 인증번호 확인", true).
		AddParamPath("", "id", "").
		AddParamPath("", "corporation_type", "").
		SetDescription("대표자/담당자 휴대전화번호 인증을 요청합니다.").
		SetSecurity("Authorization")

	// FIXME: deprecated (/v1/corporate-users/verify-postal-code 사용)
	g.POST("/postal-code-check", c.PostalCodeCheck,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("법인회원 우편코드 확인").
		AddParamBody(CorpPostalCodeInput{}, "body", "법인회원 우편코드 확인", true).
		SetDescription("법인회원 우편코드  확인").
		SetSecurity("Authorization").
		SetDeprecated()
}

// RequestSMSVerification SMS 인증 코드 (최초)전송 API
// TODO: CorporateApiController와 중복되는 부분이 많음. 추후 리팩토링 필요
func (CorporateApiController) RequestSMSVerification(c echo.Context) error {
	var v MobilePhoneVerificationInput
	mvr := &domain.MobilePhoneVerificationRequest{}

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// get user info from token
	userId, email, err := tokenToUser(c)
	if err != nil {
		return err
	}

	// check the number of requests (maximum 3 times within 24 hours)
	sendCnt, err := mvr.CountFirstReqByUserIdInToday(c.Request().Context(), userId, v.PhoneType)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if sendCnt >= MAX_FIRST_REQ {
		return factory.ErrorMaxTransmissionsReached.New(nil)
	}

	// create sms code
	v.UserId = userId
	v.FirstReq = true
	mvr, err = v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	_, err = mvr.CreateMobilePhoneVerificationRequest(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// send sms
	_, err = domain.SmsMsg{}.SendCertSms(c.Request().Context(), mvr.Code, email, v.MobilePhone)
	if err != nil {
		return factory.ErrorAWSSendSMS.New(err)
	}

	return renderSucc(c, http.StatusOK, map[string]string{"request_id": mvr.RequestId})
}

// ReRequestSMSVerification SMS認証コード再送信API
// TODO: CorporateApiControllerと重複する部分が多い。後でリファクタリング必要
func (CorporateApiController) ReRequestSMSVerification(c echo.Context) error {
	var v MobilePhoneVerificationInput
	mvr := &domain.MobilePhoneVerificationRequest{}

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// get user info from token
	userId, email, err := tokenToUser(c)
	if err != nil {
		return err
	}

	// check the number of requests (maximum 5 times per first request)
	resendCnt, err := mvr.CountResendReqAfterFirstReqByUserId(c.Request().Context(), userId, v.PhoneType)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if resendCnt >= MAX_RESEND_REQ {
		return factory.ErrorMaxTransmissionsReached.New(nil)
	}

	// check for unexpired request
	unexpired, err := mvr.GetNotExpireByUserIdAndMobilePhone(c.Request().Context(), userId, v.PhoneType, v.MobilePhone)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if unexpired != nil {
		return factory.ErrorRequestTimeNotExpired.New(nil)
	}

	// create sms code
	v.UserId = userId
	v.FirstReq = false
	mvr, err = v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	_, err = mvr.CreateMobilePhoneVerificationRequest(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// send sms
	_, err = domain.SmsMsg{}.SendCertSms(c.Request().Context(), mvr.Code, email, v.MobilePhone)
	if err != nil {
		return factory.ErrorAWSSendSMS.New(err)
	}

	return renderSucc(c, http.StatusOK, map[string]string{"request_id": mvr.RequestId})
}

func (a CorporateApiController) MobilePhoneVerificationConfirm(c echo.Context) error {
	var v MobilePhoneVerificationResponse
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// MIR-1420: 書き換える相手と照合する会員は、トークンの会員にする。本文の user_id で決めると、他の会員の電話番号を
	// 書き換えられた。
	userId, _, err := tokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	before, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), userId)
	if err != nil {
		return err
	}
	beforeUser := before.User.(*domain.CorporateUser)

	verification, err := domain.MobilePhoneVerificationRequest{}.GetByRequestId(c.Request().Context(), v.RequestId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if verification == nil {
		return factory.ErrorNotFoundRequest.New(err)
	}

	// 有効期限チェック
	now := jsonutils.JSONTime(time.Now())
	if now.Unix() > verification.Expire.Unix() {
		return factory.ErrorExpireRequest.New(nil)
	}

	res, err := verification.VerifyRequest(c.Request().Context(), userId, v.RequestId, v.Code, time.Now())
	if err != nil {
		return factory.ErrorFailSmsCert.New(err)
	}
	if res {
		corpUser := *beforeUser
		t := c.Param("corporation_type")
		switch domain.CORP_TYPE(t) {
		case domain.REPRESENTATIVE:
			corpUser.RepresentativeMobilePhone = verification.MobilePhone
		case domain.MANAGER:
			corpUser.ManagerMobilePhone = &verification.MobilePhone
		}

		err = corpUser.RepresentativeVerificationConfirm(c.Request().Context())
		if err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	after, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), userId)
	if err != nil {
		return err
	}

	a.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, userId, beforeUser.Name, beforeUser.Email)

	return renderSucc(c, http.StatusOK, verification)
}

func (a CorporateApiController) PostalCodeCheck(c echo.Context) error {
	var v CorpPostalCodeInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	uid, _, err := tokenToUser(c)
	if err != nil {
		return err
	}

	before, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), uid)
	if err != nil {
		return err
	}

	user := *before.User.(*domain.CorporateUser)

	// 郵便コード認証は住所確認中のときだけ許す。
	// コードは DB に残り続けるため、保留・無効・承認済みから同じコードを再送されると
	// audit_status が承認完了へ戻り、承認日の更新と承認メールの再送まで起きる。
	// NOTE: このエンドポイントは deprecated だがルートは生きているため、
	// /v1/corporate-users/verify-postal-code と同じガードを入れて迂回を防ぐ。
	if user.AuditStatus != domain.CHECK_ADDRESS_AUDIT {
		return factory.ErrorInvalidAuditStatus.New(nil)
	}

	err = user.CheckAllPostalCode(c.Request().Context(), v.PostalCode, v.CorporatePostalCode, v.CorpType)
	if err != nil {
		return err
	}

	// MIR-1252: CheckAllPostalCodeは両方の郵便コードが確認できるまでAuditStatusを
	// COMPLETED_AUDITへ進めない。片方だけの呼び出し(corpType指定)ではまだ
	// 審査中のままなので、完了メールの送信とバーチャル口座の割り当ては
	// 実際に完了した場合だけ行う。ここを無条件のままにすると、審査未完了なのに
	// 「本登録完了」メールが届き、口座だけ先に割り当てられてしまう。
	// 完了メール→履歴作成→口座割り当て、という既存の順序自体は変えていない
	// (口座割り当てを履歴作成より前に動かすと、割り当て失敗時にreturnで
	// 抜けてしまい、既にDBへコミット済みの変更が履歴に一切残らなくなるため)。
	auditCompletedNow := user.AuditStatus == domain.COMPLETED_AUDIT

	if auditCompletedNow {
		var mail domain.Mail
		err = mail.CodeAuditCompleteForCorporateEmail(c.Request().Context(), serviceutils.GetServiceInfoMail(), user.Email, user.Name)
		if err != nil {
			factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
				"userId": user.User.Id,
				"sender": serviceutils.GetServiceInfoMail(),
				"err":    err,
			}).Info("CodeAuditCompleteForCorporateEmail")
		}
	}

	after, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), uid)
	if err != nil {
		return err
	}

	a.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, uid, user.Name, user.Email)

	if auditCompletedNow && serviceutils.EnvIsProd() {
		// バーチャル口座割り当て
		// NOTE: 本番環境のみ自動割り当て。その他環境は手動で割り当てる。
		err = user.LinkVaAccounts(c.Request().Context(), domain.EQUITY)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"message": "Failed to link va accounts",
				"error":   err,
				"time":    time.Now().String(),
			}).Error("VerifyPostalCode: failed to link va accounts")

			return err
		}
	}

	token, err := domain.CreateToken(c.Request().Context(), user.User)
	if err != nil {
		return err
	}

	resultToken := domain.ResultToken{
		AccessToken: token,
	}
	return renderSucc(c, http.StatusOK, resultToken)
}
