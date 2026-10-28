package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"
	userservice "github.com/illuminarean/energyshares-user-api/service"
	"github.com/illuminarean/energyshares-user-api/service/dto"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type UserApiController struct {
	myPageService          *userservice.MyPageService
	userInfoHistoryService service.IUserInfoHistoryService
}

func NewUserApiController() *UserApiController {
	return &UserApiController{
		myPageService:          userservice.NewMyPageService(),
		userInfoHistoryService: service.NewUserInfoHistoryService(),
	}
}

// 회원 탈퇴
// 비밀번호 확인

func (c UserApiController) Init(g echoswagger.ApiGroup) {
	// auth
	g.POST("/logout", c.Logout,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("회원 로그아웃 ").
		SetDescription("인증 토큰을 만료합니다.").
		AddParamBody(LogoutTokenRequest{}, "body", "인증 토큰 만료 ", true).
		SetSecurity("Authorization")

	g.POST("/passwordresetrequest", c.PasswordResetRequest).
		AddParamBody(PasswordResetRequestInput{}, "body", "패스워드 초기화 요청", true).
		SetSummary("패스워드 초기화 요청 생성").
		SetDescription("패스워드 초기화 법인 개인 동일").
		SetDeprecated() // Deprecated

	g.POST("/:id/passwordcheck", c.PasswordCheck,
		echoutils.AllowRoles(domain.ROLE_USER)).
		AddParamPath("", "id", "").
		SetSummary("패스워드 확인").
		AddParamBody(PasswordCheckInput{}, "body", "패스워드 확인", true).
		SetDescription("패스워드 동일 여부 확인").
		SetSecurity("Authorization")

	g.POST("/:id/mobile/verification-request", c.RequestSMSVerification,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("휴대전화번호 인증 요청 ").
		AddParamBody(MobilePhoneVerificationInput{}, "body", "패스워드 초기화 요청", true).
		AddParamPath("", "id", "").
		SetDescription("휴대 전화 번호 검증을 요청 합니다.").
		SetSecurity("Authorization")

	g.POST("/:id/mobile/verification-re-request", c.ReRequestSMSVerification,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("휴대전화번호 인증 재요청 ").
		AddParamBody(MobilePhoneVerificationInput{}, "body", "패스워드 초기화 요청", true).
		AddParamPath("", "id", "").
		SetDescription("휴대 전화 번호 검증을 요청 합니다.").
		SetSecurity("Authorization")

	g.POST("/:id/mobile/verification", c.ConfirmSmsVerification,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("휴대전화번호 인증 확인").
		AddParamBody(MobilePhoneVerificationResponse{}, "body", "패스워드 초기화 요청", true).
		AddParamPath("", "id", "").
		SetDescription("휴대폰 번호를 검증을 확인합니다.").
		SetSecurity("Authorization")

	g.POST("/email-code-request", c.RequestEmailVerification,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("이메일 코드 전송").
		AddParamBody(EmailCodeInput{}, "body", "이메일 코드 전송", true).
		SetDescription("이메일 코드 전송").
		SetSecurity("Authorization")

	g.POST("/email-code-re-request", c.ReRequestEmailVerification,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("이메일 코드 재전송").
		AddParamBody(EmailCodeInput{}, "body", "이메일 코드 재전송", true).
		SetDescription("이메일 코드 재전송").
		SetSecurity("Authorization")

	g.POST("/email-code-check", c.EmailCodeCheck,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("이메일 코드 확인").
		AddParamBody(EmailCodeCheckInput{}, "body", "이메일 코드 확인", true).
		SetDescription("이메일 코드 확인").
		SetSecurity("Authorization")

	g.GET("/:id/mypages", c.GetMyPage,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("마이페이지 요청").
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "마이페이지 정보", dto.MyPage{}, nil).
		SetSecurity("Authorization").
		SetDescription("마이페이지 요청")

	g.PUT("/:id/withdrawal-account", c.UpdateWithdrawalAccount,
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.COMPLETED_AUDIT)).
		SetSummary("유저의 출금계좌 정보 수정").
		SetDescription("유저의 출금계좌 정보 수정 (예금주 명은 수정 불가)").
		AddParamPath("", "id", "유저 ID").
		AddParamBody(UserWithdrawalAccountInput{}, "body", "", true).
		AddResponse(http.StatusOK, "사용자 출금 계좌 정보", UserWithdrawalAccountOutput{}, nil).
		SetSecurity("Authorization")

	g.GET("/:id/notice-messages", c.GetUserByNoticeMsg,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("사용자 소식 리스트").
		AddParamPath("", "id", "").
		SetDescription("사용자 예치금 정보 조회").
		SetSecurity("Authorization")

	// transactions
	g.GET("/:id/transactions", c.GetUserTransactions).
		SetSummary("사용자 입출금 내역 목록").
		AddParamPath("", "id", "사용자 ID").
		AddParamQuery(SearchInput{}, "", "", true).
		AddResponse(http.StatusOK, "사용자 입출금 내역 ", domain.UserTransactionDetail{}, nil).
		SetSecurity("Authorization").
		SetDescription("유저가 투자한 투자거래내역 요청")

	// investment
	g.GET("/:id/investments", c.GetInvestmentList).
		SetSummary("유저가 투자한 모든 투자거래내역 요청").
		AddParamPath("", "id", "").
		AddParamQuery(MyPageSearchInput{}, "", "", true).
		AddResponse(http.StatusOK, "유저가 투자한 모든 투자거래내역 정보", domain.Investment{}, nil).
		SetSecurity("Authorization").
		SetDescription("유저가 투자한 투자거래내역 요청")

	// investment
	g.GET("/:id/investments/:share_id", c.GetInvestment,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("투자거래내역 상세정보 요청").
		AddParamPath("", "id", "").
		AddParamPath("", "share_id", "").
		AddResponse(http.StatusOK, "투자거래내역 상세정보", domain.Investment{}, nil).
		SetSecurity("Authorization").
		SetDescription("투자거래내역 상세정보 요청")

	g.PUT("/:id/investments/:share_id/file-check-time", c.InvestmentFileCheck,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("거래 각종 서류 확인 시간 체크").
		AddParamPath("", "id", "").
		AddParamPath("", "share_id", "").
		AddParamBody(InvestmentFileType{}, "body", "거래 각종 서류 확인 시간 체크", true).
		AddResponse(http.StatusOK, "거래 각종 서류 확인 시간 체크", domain.Investment{}, nil).
		SetSecurity("Authorization").
		SetDescription("거래 각종 서류 확인 시간 체크")

	g.GET("/:id/waitlist", c.GetWaitlistList,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("ユーザーのキャンセル待ち一覧").
		AddParamPath("", "id", "").
		AddParamQuery(SearchInput{}, "", "", true).
		AddResponse(http.StatusOK, "ユーザーのキャンセル待ち一覧", UserWaitlistResponse{}, nil).
		SetSecurity("Authorization").
		SetDescription("ユーザーのキャンセル待ち一覧")

	g.GET("/:id/investment-dividends", c.GetUserByDividends,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("유저가 투자한 분배 내역 명세  목록").
		AddParamPath("", "id", "").
		AddParamQuery(UserByDividendsInput{}, "", "", true).
		AddResponse(http.StatusOK, "유저가 투자한 모든 투자거래내역 정보", domain.InvestmentDividendDetail{}, nil).
		SetSecurity("Authorization").
		SetDescription("유저가 투자한 분배 내역 명세  목록")

	g.GET("/:id/project-dividends", c.GetProjectByDividends,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("유저가 투자한 모든 상품별 분배 내역 합계 목록").
		AddParamPath("", "id", "").
		AddParamQuery(SearchInput{}, "", "", true).
		AddResponse(http.StatusOK, "유저가 투자한 모든 투자거래내역 정보", domain.ProjectDividendDetail{}, nil).
		SetSecurity("Authorization").
		SetDescription("유저가 투자한 투자거래내역 요청")

	// project

	g.GET("/:id/investment-dividend-projects", c.GetInvestmentDividendProjects,
		echoutils.AllowRoles(domain.ROLE_USER)).
		AddParamPath("", "id", "").
		AddParamQuery(SearchInput{}, "", "", true).
		SetSummary("유저 분배 내역 투자상품 목록 조회").
		SetDescription("유저 분배 내역 투자상품 목록 조회을 조회합니다.").
		SetSecurity("Authorization")

	// report
	g.GET("/me/reports/transaction-balance-reports", c.GetTransactionBalanceReports,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("사용자 장표출력 보고서 목록").
		AddParamQuery(SearchInput{}, "", "", true).
		AddResponse(http.StatusOK, "사용자 장표출력 보고서 목록", domain.ReportOutput{}, nil).
		SetSecurity("Authorization").
		SetDescription("사용자 장표출력 보고서 목록")

	g.GET("/me/reports/payment-statement", c.GetPaymentStatementReports,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("사용자 장표출력 보고서 목록").
		AddParamQuery(SearchInput{}, "", "", true).
		AddResponse(http.StatusOK, "사용자 장표출력 보고서 목록", domain.ReportOutput{}, nil).
		SetSecurity("Authorization").
		SetDescription("사용자 장표출력 보고서 목록")
}

var (
	MAX_FIRST_REQ  = int64(3)
	MAX_RESEND_REQ = int64(5)
)

func (UserApiController) Logout(c echo.Context) error {
	var v LogoutTokenRequest

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	_, err := domain.User{}.Logout(c.Request().Context(), v.RefreshToken)
	if err != nil {
		return err
	}
	return renderSucc(c, http.StatusOK, nil)
}

func (UserApiController) PasswordResetRequest(c echo.Context) error {
	var v PasswordResetRequestInput

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	user, err := domain.User{}.GetByEmail(c.Request().Context(), v.Email)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if user == nil {
		return factory.ErrorUserNotExists.New(errors.New("Not Found Email"))
	}
	if user.AuditStatus == domain.INVALIDITY_AUDIT {
		return factory.ErrorInvalidityAudit.New(err)
	}
	if user.Deleted.Unix() > 0 {
		return factory.ErrorDeletedUser.New(err)
	}
	if user.LoginType != domain.MAIL {
		return factory.ErrorSocialUser.New(err)
	}

	prr, err := domain.PasswordResetRequest{}.GetByEmail(c.Request().Context(), v.Email)
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if prr == nil {
		prr, err = v.ToModel()
		if err != nil {
			return factory.ErrorParameter.New(err)
		}

		err = prr.CreatePasswordResetRequest(c.Request().Context(), v.Email)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
	} else if prr != nil {
		prr.SendCount++
		if prr.SendCount > domain.MAX_SENT_PASSWORD_RESET_REQUEST_MAIL {
			return factory.ErrorSentEmail.New(nil)
		}
		_, err = prr.UpdateById(c.Request().Context())
		if err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	var mail domain.Mail
	err = mail.PasswordResetMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), prr.Email, user.Name, prr.Url)
	if err != nil {
		// MIR-1365: 以前はprr.Email、prr.Url(トークンを含むパスワード再設定リンク)
		// もそのままログへ出していた。IDから対象を特定できるため、出さない。
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"ID": prr.Id,
		}).Error("Fail Password Reset Email Send")
	}

	return renderSucc(c, http.StatusOK, true)
}

func (UserApiController) PasswordCheck(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	var v PasswordCheckInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	user, err := domain.User{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if !user.ValidatePassword(v.Password) {
		return factory.ErrorNotMatchPassword.New(nil)
	}
	return renderSucc(c, http.StatusOK, true)
}

func (a *UserApiController) GetMyPage(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	uid, _, err := tokenToUser(c)
	if err != nil {
		return err
	}

	if err := jwthelper.ValidateUserToken(c, id); err != nil {
		return err
	}

	myPage, err := a.myPageService.GetMyPageById(c.Request().Context(), uid)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	myPage.Roles, err = jwthelper.GetRoleFromToken(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	return renderSucc(c, http.StatusOK, myPage)
}

func (UserApiController) GetInvestmentList(c echo.Context) error {
	var v MyPageSearchInput
	_, _, err := tokenToUser(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := jwthelper.ValidateUserToken(c, userId); err != nil {
		return err
	}

	user, err := domain.User{}.GetById(c.Request().Context(), userId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if user == nil {
		return factory.ErrorNotFound.New(errors.New("Not Found User"))
	}

	totalCount, invests, err := domain.Investment{}.
		GetInvestmentsByOwnerId(c.Request().Context(), userId, v.ProjectType, v.ProjectStatus, v.InvestStatus, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, ArrayResult{
		Items:      invests,
		TotalCount: totalCount,
	})
}

func (UserApiController) GetInvestment(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	share_id := c.Param("share_id")

	share, err := domain.Share{}.GetById(c.Request().Context(), share_id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if share == nil {
		return factory.ErrorNotFound.New(err)
	}

	if share.OwnerId != id {
		return factory.ErrorNotHaveRole.New(errors.New("owner id of share and id are not equal"))
	}

	invest, err := domain.Investment{}.GetById(c.Request().Context(), share_id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if invest == nil {
		return factory.ErrorNotFound.New(err)
	}

	return renderSucc(c, http.StatusOK, invest)
}

func (UserApiController) GetWaitlistList(c echo.Context) error {
	var v SearchInput
	_, _, err := tokenToUser(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := jwthelper.ValidateUserToken(c, userId); err != nil {
		return err
	}

	user, err := domain.User{}.GetById(c.Request().Context(), userId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if user == nil {
		return factory.ErrorNotFound.New(errors.New("Not Found User"))
	}

	totalCount, waiters, err := domain.InvestmentWaiters{}.GetByUserId(c.Request().Context(), userId, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	var response []UserWaitlistResponseItem
	for _, waiter := range waiters {
		response = append(response, newUserWaitlistResponseItem(waiter))
	}

	return renderSucc(c, http.StatusOK, UserWaitlistResponse{
		TotalCount: totalCount,
		Items:      response,
	})
}

// RequestSMSVerification SMS 인증 코드 (최초)전송 API
// TODO: CorporateApiController와 중복되는 부분이 많음. 추후 리팩토링 필요
func (UserApiController) RequestSMSVerification(c echo.Context) error {
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

	if sendCnt >= MAX_FIRST_REQ && !serviceutils.EnvIsStg() {
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

// ReRequestSMSVerification SMS 인증코드 재전송 API
// TODO: CorporateApiController와 중복되는 부분이 많음. 추후 리팩토링 필요
func (UserApiController) ReRequestSMSVerification(c echo.Context) error {
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

func (a UserApiController) ConfirmSmsVerification(c echo.Context) error {
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

	before, err := a.userInfoHistoryService.GetDataForUser(c.Request().Context(), userId)
	if err != nil {
		return err
	}

	verification, err := domain.MobilePhoneVerificationRequest{}.GetByRequestId(c.Request().Context(), v.RequestId)
	if err != nil {
		return factory.ErrorParameter.New(err)
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
		user := *before.User.(*domain.User)
		user.MobilePhone = verification.MobilePhone
		err = user.MobilePhoneVerificationConfirm(c.Request().Context())
		if err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	after, err := a.userInfoHistoryService.GetDataForUser(c.Request().Context(), userId)
	if err != nil {
		return err
	}
	afterUser := after.User.(*domain.User)

	a.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, userId, afterUser.Name, afterUser.Email)

	var mail domain.Mail
	if err := mail.UpdateUserInformationMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), afterUser.Email, afterUser.Name, before.User.(*domain.User).MobilePhone, afterUser.MobilePhone, false); err != nil {
		return factory.ErrorAWSSendMail.New(err)
	}

	return renderSucc(c, http.StatusOK, verification)
}

func verifyEmailForChange(c echo.Context, old, new string) error {
	// verify different email
	if old == new {
		return factory.ErrorSameEmailExists.New(nil)
	}

	// verify unused email
	user, err := domain.User{}.GetByEmail(c.Request().Context(), new)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if user != nil {
		return factory.ErrorSameEmailExists.New(nil)
	}

	return nil
}

// RequestEmailVerification Email 인증 코드 (최초)발송 API
func (UserApiController) RequestEmailVerification(c echo.Context) error {
	var v EmailCodeInput
	ecr := &domain.EmailCodeRequest{}

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

	// verify email
	err = verifyEmailForChange(c, email, v.Email)
	if err != nil {
		return err
	}

	// check the number of requests (maximum 3 times within 24 hours)
	sendCnt, err := ecr.CountFirstReqByUserIdInToday(c.Request().Context(), userId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if sendCnt >= MAX_FIRST_REQ && !serviceutils.EnvIsStg() {
		return factory.ErrorMaxTransmissionsReached.New(nil)
	}

	// create email code
	v.UserId = userId
	v.FirstReq = true
	ecr, err = v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	_, err = ecr.Create(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// send email
	var mail domain.Mail
	err = mail.CertificationCodeSendMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), ecr.Email, v.Name, ecr.Code)
	if err != nil {
		return factory.ErrorAWSSendMail.New(err)
	}

	return renderSucc(c, http.StatusOK, true)
}

// ReRequestEmailVerification Email 인증 코드 재발송 API
func (UserApiController) ReRequestEmailVerification(c echo.Context) error {
	var v EmailCodeInput
	ecr := &domain.EmailCodeRequest{}

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

	// verify email
	err = verifyEmailForChange(c, email, v.Email)
	if err != nil {
		return err
	}

	// check the number of requests (maximum 5 times per first request)
	resendCnt, err := ecr.CountResendReqAfterFirstReqByUserId(c.Request().Context(), userId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if resendCnt >= MAX_RESEND_REQ {
		return factory.ErrorMaxTransmissionsReached.New(nil)
	}

	// check for unexpired request
	unexpired, err := ecr.GetNotExpireByUserIdAndEmail(c.Request().Context(), userId, v.Email)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if unexpired != nil {
		return factory.ErrorRequestTimeNotExpired.New(nil)
	}

	// create email code
	v.FirstReq = false
	v.UserId = userId
	ecr, err = v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	_, err = ecr.Create(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// send email
	var mail domain.Mail
	err = mail.CertificationCodeSendMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), ecr.Email, v.Name, ecr.Code)
	if err != nil {
		return factory.ErrorAWSSendMail.New(err)
	}

	return renderSucc(c, http.StatusOK, true)
}

func (UserApiController) EmailCodeCheck(c echo.Context) error {
	var v EmailCodeCheckInput

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	mc, err := domain.EmailCodeRequest{}.GetByEmail(c.Request().Context(), v.Email, v.Code)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if mc == nil {
		return factory.ErrorCodeMismatch.New(nil)
	}
	// メールコード有効期限チェック
	now := jsonutils.JSONTime(time.Now())
	if now.Unix() > mc.Expire.Unix() {
		return factory.ErrorExpireRequest.New(nil)
	}

	_, err = domain.EmailCodeRequest{
		Expire: jsonutils.JSONTime(time.Now()),
		Email:  v.Email,
	}.Update(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, true)
}

// GetUserTransactions 사용자 입출금 목록 조회기능
func (UserApiController) GetUserTransactions(c echo.Context) error {
	var v TransactionSearchInput
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := jwthelper.ValidateUserToken(c, id); err != nil {
		return err
	}

	user, err := domain.User{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	totalCount, items, err := domain.UserTransaction{}.GetByUserId(c.Request().Context(), user.Id, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.ProjectType)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

// UpdateWithdrawalAccount 사용자의 출금 계좌 정보 수정
func (UserApiController) UpdateWithdrawalAccount(c echo.Context) error {
	var v UserWithdrawalAccountInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	// check authority
	tokenId, _, err := tokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	if tokenId != id {
		return factory.ErrorNotHaveRole.New(err)
	}

	// update withdrawal account info
	user, err := v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	user.Id = tokenId
	if err := user.UpdateWithdrawalAccount(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	// get updated withdrawal account info
	userWithdrawalAccountOutput := &UserWithdrawalAccountOutput{}
	updated, err := domain.User{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, userWithdrawalAccountOutput.New(updated))
}

func (UserApiController) GetProjectByDividends(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := jwthelper.ValidateUserToken(c, id); err != nil {
		return err
	}

	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	totalCount, items, err := domain.ProjectDividendDetail{}.GetByUserProjectByDividends(c.Request().Context(), id, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return renderSucc(c, http.StatusOK, ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (UserApiController) GetUserByDividends(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := jwthelper.ValidateUserToken(c, id); err != nil {
		return err
	}

	var v UserByDividendsInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	totalCount, items, err := domain.InvestmentDividendDetail{}.GetDividendsByUserId(c.Request().Context(),
		id, v.StartDate, v.EndDate, v.TransactionDetailType, v.ProjectId, v.ProjectName,
		v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

func (UserApiController) GetPaymentStatementReports(c echo.Context) error {
	userId, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	totalCount, items, err := domain.ReportOutput{}.GetByUserReport(c.Request().Context(), userId, domain.REPORT_PAYMENT_STATEMENT, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	responses := make([]*ReportResponse, 0)
	for _, item := range items {
		responses = append(responses, ReportResponse{}.NewReportResponse(item.Id, item.Title, item.FileUrl))
	}
	return renderSucc(c, http.StatusOK, ArrayResult{
		Items:      responses,
		TotalCount: totalCount,
	})
}

func (UserApiController) GetTransactionBalanceReports(c echo.Context) error {
	userId, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	totalCount, items, err := domain.ReportOutput{}.GetByUserReport(c.Request().Context(), userId, domain.REPORT_TRANSACTION_BALANCE_REPORT, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	responses := make([]*ReportResponse, 0)
	for _, item := range items {
		responses = append(responses, ReportResponse{}.NewReportResponse(item.Id, item.Title, item.FileUrl))
	}
	return renderSucc(c, http.StatusOK, ArrayResult{
		Items:      responses,
		TotalCount: totalCount,
	})
}

// 공지 사항 조회
func (UserApiController) GetUserByNoticeMsg(c echo.Context) error {
	var v NoticeSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := jwthelper.ValidateUserToken(c, id); err != nil {
		return err
	}

	totalCount, items, err := domain.NoticeMsg{}.GetUserNoticeMsg(c.Request().Context(), v.Title, v.StartDate, v.EndDate, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (UserApiController) InvestmentFileCheck(c echo.Context) error {
	var v InvestmentFileType

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := jwthelper.ValidateUserToken(c, id); err != nil {
		return err
	}

	shareId := c.Param("share_id")
	share, err := domain.Share{}.GetById(c.Request().Context(), shareId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if share == nil {
		return factory.ErrorNotFound.New(err)
	}

	if share.OwnerId != id {
		return factory.ErrorParameter.New(errors.New("owner id of share and id are not equal"))
	}

	// CONTRACT_BEFORE_PAGE, SILENT_PARTNERSHIPS: 投資申込時確認日時保存
	// CONTRACT_AFTER_PAGE, PROJECT_REPORT: ダウンロードボタンクリック時確認日時保存
	switch v.ProjectFileType {
	// case domain.CONTRACT_BEFORE_PAGE:
	//	share.ContractBeforePageCheckDate = jsonutils.JSONTime(time.Now())
	case domain.CONTRACT_AFTER_PAGE:
		share.ContractAfterPageCheckDate = jsonutils.JSONTime(time.Now())
	// case domain.SILENT_PARTNERSHIPS:
	//	share.SilentPartnershipsCheckDate = jsonutils.JSONTime(time.Now())
	case domain.PROJECT_REPORT:
		share.ProjectReportCheckDate = jsonutils.JSONTime(time.Now())
	}

	share.UpdateId = id

	if err := share.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	invest, err := domain.Investment{}.GetById(c.Request().Context(), shareId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if invest == nil {
		return factory.ErrorNotFound.New(err)
	}

	return renderSucc(c, http.StatusOK, invest)
}

func (UserApiController) GetInvestmentDividendProjects(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := jwthelper.ValidateUserToken(c, id); err != nil {
		return err
	}

	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	user, err := domain.User{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if user == nil {
		return factory.ErrorUserNotExists.New(err)
	}

	project, err := domain.Project{}.GetInvestmentDividendProjects(c.Request().Context(), id, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, project)
}
