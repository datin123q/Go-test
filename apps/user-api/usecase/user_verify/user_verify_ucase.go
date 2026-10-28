package user_verify

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pwhelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

const (
	numericChars   = "0123456789"
	maxSendCount   = 3
	maxResendCount = 5

	emailCodeExpireTime = time.Minute * 20 // メール認証コード有効期限
)

type userVerifyUsecase struct {
	cfg                            *config.Config
	pwhelper                       pwhelper.PasswordHelper
	logger                         factory.LoggerFactory
	mailService                    cdomain.MailService
	passwordResetRequestService    cdomain.PasswordResetRequestService
	smsMsgService                  cdomain.SmsMsgService
	userService                    cdomain.UserService
	signUpRequestRepository        cdomain.SignUpRequestRepository
	passwordResetRequestRepository cdomain.PasswordResetRequestRepository
	emailCodeRequestRepository     cdomain.EmailCodeRequestRepository
	userRepository                 cdomain.UserRepository

	// 内部関数
	generateRandomString func(n int, charSet string) (string, error)
}

func New(
	cfg *config.Config,
	pwhelper pwhelper.PasswordHelper,
	logger factory.LoggerFactory,
	mailService cdomain.MailService,
	passwordResetRequestService cdomain.PasswordResetRequestService,
	smsMsgService cdomain.SmsMsgService,
	userService cdomain.UserService,
	signUpRequestRepository cdomain.SignUpRequestRepository,
	passwordResetRequestRepository cdomain.PasswordResetRequestRepository,
	userRepository cdomain.UserRepository,
	emailCodeRequestRepository cdomain.EmailCodeRequestRepository,
) udomain.UserVerifyUsecase {
	return &userVerifyUsecase{
		cfg:                            cfg,
		pwhelper:                       pwhelper,
		logger:                         logger,
		mailService:                    mailService,
		passwordResetRequestService:    passwordResetRequestService,
		smsMsgService:                  smsMsgService,
		userService:                    userService,
		signUpRequestRepository:        signUpRequestRepository,
		passwordResetRequestRepository: passwordResetRequestRepository,
		userRepository:                 userRepository,
		emailCodeRequestRepository:     emailCodeRequestRepository,
		generateRandomString:           generateRandomString,
	}
}

func (u userVerifyUsecase) CheckEmailExists(ctx context.Context, request udomain.UserCheckEmailExistsRequest) (bool, error) {

	// 削除されたメールアドレスを含めて重複チェック
	isDuplicatedEmail, err := u.userService.IsDuplicateEmailIncludeDeleted(ctx, request.Email)
	if err != nil {
		return false, errors.Wrap(err, "failed to get user by email")
	}
	if isDuplicatedEmail {
		return true, nil
	}

	// サインアップ要求中のメンバーメールの重複チェック
	checkUrlExpire, err := u.signUpRequestRepository.GetAll(ctx,
		cdomain.SignUpRequestQueryOptionFunc().WithEmail(request.Email),
		cdomain.SignUpRequestQueryOptionFunc().WithUrlExpire(jsonutils.JSONTime(time.Now())),
		cdomain.SignUpRequestQueryOptionFunc().WithSortBy("id"),
		cdomain.SignUpRequestQueryOptionFunc().WithOrder("desc"),
	)
	if err != nil {
		return false, errors.Wrap(err, "failed to get sign up requests")
	}
	if !checkUrlExpire.IsEmpty() {
		// 値あり == UrlExpireが有効期限を超えていないサインアップリクエストが存在する
		return true, nil
	}

	return false, nil
}

func (u userVerifyUsecase) RequestPasswordReset(ctx context.Context, request udomain.UserRequestPasswordResetRequest) (udomain.UserRequestPasswordResetResponse, error) {

	user, err := u.userRepository.GetByEmail(ctx, request.Email)
	if err != nil {
		return udomain.UserRequestPasswordResetResponse{}, errors.Wrap(err, "failed to get user by email")
	}

	if user.AuditStatus == domain.INVALIDITY_AUDIT {
		return udomain.UserRequestPasswordResetResponse{}, cdomain.ErrAuditInvalid
	}
	passwordResetRequest, err := u.passwordResetRequestService.HandlePasswordResetRequest(ctx, user.Email)
	if err != nil {
		return udomain.UserRequestPasswordResetResponse{}, errors.Wrap(err, "failed to get password reset request by email")
	}

	if err := u.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.PASSWORD_RESET_TEMPLATE,
		Sender:       u.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{user.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"ServiceUrl":       u.cfg.ServiceConfig.Url.UserServiceDomain,
			"Receiver":         user.Name,
			"PasswordResetUrl": passwordResetRequest.Url,
		},
	}); err != nil {
		return udomain.UserRequestPasswordResetResponse{}, errors.Wrap(err, "failed to send mail")
	}

	return udomain.UserRequestPasswordResetResponse{
		Email: user.Email,
	}, nil
}

func (u userVerifyUsecase) ResetPassword(ctx context.Context, request udomain.UserResetPasswordRequest) error {

	// get password_reset_request by request_id
	prr, err := u.passwordResetRequestRepository.GetByRequestId(ctx, request.RequestId)
	if err != nil {
		return errors.Wrap(err, "failed to get password_reset_request by request_id")
	}

	// return error if email is not matched
	if prr.Email != request.Email {
		return errors.Wrap(cdomain.ErrBadRequest, "email is not matched")
	}

	// return error if password_reset_request is already expired
	now := jsonutils.JSONTime(time.Now())
	if now.Unix() > prr.Expire.Unix() {
		return errors.Wrap(cdomain.ErrAlreadyExpired, "password_reset_request already expired")
	}

	// get user info by email
	user, err := u.userRepository.GetByEmail(ctx, prr.Email)
	if err != nil {
		return errors.Wrap(err, "failed to get user by email")
	}

	// 無効な会員はパスワード再設定不可。もともとこちらへのリクエストは困難だが、防御的なコード処理
	if user.AuditStatus == domain.INVALIDITY_AUDIT {
		return cdomain.ErrAuditInvalid
	}
	// change password
	if err := u.userService.ChangePassword(ctx, user, request.Password); err != nil {
		return errors.Wrap(err, "failed to change password")
	}

	// set password_reset_request expired
	prr.Expire = jsonutils.JSONTime(time.Now())
	if err := u.passwordResetRequestRepository.Update(ctx, prr); err != nil {
		return errors.Wrap(err, "failed to update password_reset_request")
	}

	// send mail
	if err := u.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.PASSWORD_RESET_COMPLETE_TEMPLATE,
		Sender:       u.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{user.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"Receiver":   user.Name,
			"Time":       jsonutils.JSONTime(time.Now()).ToString(),
			"ServiceUrl": u.cfg.ServiceConfig.Url.UserServiceDomain,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to send mail")
	}

	// send sms
	if err := u.smsMsgService.SendSms(ctx, cdomain.SendSmsRequest{
		Template: cdomain.SMS_PASSWORD_RESET_COMPLETE_TEMPLATE,
		Title:    cdomain.SMS_PASSWORD_RESET_COMPLETE_TITLE,
		Sender:   user.Name,
		Receiver: user.MobilePhone,
		SmsType:  domain.PASSWORD_RESET_COMPLETE_SMS,
		Details:  cdomain.SmsPasswordResetDetail{ServiceCall: "0000-0000-0000"},
	}); err != nil {
		return errors.Wrap(err, "failed to send sms")
	}

	return nil
}

func (u userVerifyUsecase) GetPasswordResetRequest(ctx context.Context, request udomain.UserGetPasswordResetRequestRequest) (udomain.UserGetPasswordResetRequestResponse, error) {

	// get password_reset_request by request_id
	prr, err := u.passwordResetRequestRepository.GetByRequestId(ctx, request.RequestId)
	if err != nil {
		return udomain.UserGetPasswordResetRequestResponse{}, errors.Wrap(err, "failed to get password_reset_request by request_id")
	}

	// check if password_reset_request is expired
	now := jsonutils.JSONTime(time.Now())
	if now.Unix() > prr.Expire.Unix() {
		return udomain.UserGetPasswordResetRequestResponse{}, errors.Wrap(cdomain.ErrAlreadyExpired, "password_reset_request already expired")
	}

	return udomain.UserGetPasswordResetRequestResponse{
		Email:     prr.Email,
		RequestId: prr.RequestId,
		Url:       prr.Url,
		Expire:    prr.Expire,
		SendCount: prr.SendCount,
	}, nil
}

func (u userVerifyUsecase) ChangePassword(ctx context.Context, request udomain.UserChangePasswordRequest) error {

	// get user info by ID
	user, err := u.userRepository.GetById(ctx, request.JWTUser.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get user by email")
	}

	// change password
	if err := u.userService.ChangePassword(ctx, user, request.Password); err != nil {
		return errors.Wrap(err, "failed to change password")
	}

	// send mail
	if err := u.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.PASSWORD_RESET_COMPLETE_TEMPLATE,
		Sender:       u.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{user.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"Receiver":   user.Name,
			"Time":       jsonutils.JSONTime(time.Now()).ToString(),
			"ServiceUrl": u.cfg.ServiceConfig.Url.UserServiceDomain,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to send mail")
	}

	// send sms
	if err := u.smsMsgService.SendSms(ctx, cdomain.SendSmsRequest{
		Template: cdomain.SMS_PASSWORD_RESET_COMPLETE_TEMPLATE,
		Title:    cdomain.SMS_PASSWORD_RESET_COMPLETE_TITLE,
		Sender:   user.Name,
		Receiver: user.MobilePhone,
		SmsType:  domain.PASSWORD_RESET_COMPLETE_SMS,
		Details:  cdomain.SmsPasswordResetDetail{ServiceCall: "0000-0000-0000"},
	}); err != nil {
		return errors.Wrap(err, "failed to send sms")
	}

	return nil
}

func (u userVerifyUsecase) CheckPassword(ctx context.Context, request udomain.UserCheckPasswordRequest) (bool, error) {
	user, err := u.userRepository.GetById(ctx, request.JWTUser.Id)
	if err != nil {
		return false, errors.Wrap(err, "failed to get user by email")
	}

	encodedInput := u.pwhelper.EncodePassword(request.Password, user.Salt)
	if u.pwhelper.IsSamePassword(user.Password, encodedInput) {
		return true, nil
	} else {
		return false, nil
	}
}

func (u userVerifyUsecase) SendSmsCode(ctx context.Context, request udomain.UserSendSmsCodeRequest) (udomain.UserSendSmsCodeResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (u userVerifyUsecase) ResendSmsCode(ctx context.Context, request udomain.UserResendSmsCodeRequest) (udomain.UserResendSmsCodeResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (u userVerifyUsecase) CheckSmsCode(ctx context.Context, request udomain.UserCheckSmsCodeRequest) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (u userVerifyUsecase) SendEmailCode(ctx context.Context, request udomain.UserSendEmailCodeRequest) error {
	// 変更しようとするメールアドレスで既に登録されたアカウントがある場合はエラーを返します。(エラーがない場合は照会が正常に行われたことを意味します)
	isDuplicatedEmail, err := u.userService.IsDuplicateEmail(ctx, request.Email)
	if err != nil {
		return errors.Wrap(err, "failed to get user by email")
	}
	if isDuplicatedEmail {
		return cdomain.ErrAlreadyExists
	}

	today := time.Now().Format(time.DateOnly)

	// メールコードリクエストテーブルの確認
	emailCodeRequests, err := u.emailCodeRequestRepository.GetAll(ctx,
		cdomain.EmailCodeRequestQueryOptionFunc().WithUserId(request.JWTUser.Id),
		cdomain.EmailCodeRequestQueryOptionFunc().WithFirstReq(true),
		cdomain.EmailCodeRequestQueryOptionFunc().WithGTECreated(today),
	)
	if err != nil {
		return errors.Wrap(err, "failed to get email code requests")
	}

	// 3回以上申請した履歴がある場合はエラーを返す
	if len(emailCodeRequests) >= maxSendCount && !serviceutils.EnvIsStg() {
		return cdomain.ErrRequestLimitExceeded
	}

	certificationCode, err := u.generateRandomString(6, numericChars)
	if err != nil {
		return err
	}

	// メールコードリクエストを作成
	if err = u.emailCodeRequestRepository.Create(ctx, domain.EmailCodeRequest{
		UserId:   request.JWTUser.Id,
		Email:    request.Email,
		Code:     certificationCode,
		FirstReq: true,
		Expire:   jsonutils.JSONTime(time.Now().Add(emailCodeExpireTime)),
	}); err != nil {
		return errors.Wrap(err, "failed to create email code request")
	}

	// メール送信、正常に返す
	if err = u.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.CERTIFICATION_CODE_SEND_TEMPLATE,
		Sender:       u.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{request.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_FAST,
		Details: map[string]interface{}{
			"Receiver":          request.Email,
			"ServiceUrl":        u.cfg.ServiceConfig.Url.UserServiceDomain,
			"CertificationCode": certificationCode,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to send mail")
	}

	return nil
}

func (u userVerifyUsecase) ResendEmailCode(ctx context.Context, request udomain.UserResendEmailCodeRequest) error {
	// 変更しようとするメールアドレスで既に登録されたアカウントがある場合はエラーを返します。(エラーがない場合は照会が正常に行われたことを意味します)
	isDuplicatedEmail, err := u.userService.IsDuplicateEmail(ctx, request.Email)
	if err != nil {
		return errors.Wrap(err, "failed to get user by email")
	}
	if isDuplicatedEmail {
		return cdomain.ErrAlreadyExists
	}

	today := time.Now().Format(time.DateOnly)

	// メールコードリクエストテーブルの確認
	// 最初の送信照会（最新のもの）+時間逆順で並べ替え（最新に送信されたデータを最初に）
	// リクエスト自体に制限があるため（3件）、別途Limitは設定しない
	emailCodeRequests, err := u.emailCodeRequestRepository.GetAll(ctx,
		cdomain.EmailCodeRequestQueryOptionFunc().WithUserId(request.JWTUser.Id),
		cdomain.EmailCodeRequestQueryOptionFunc().WithGTECreated(today),
		cdomain.EmailCodeRequestQueryOptionFunc().WithSortBy("id"),
		cdomain.EmailCodeRequestQueryOptionFunc().WithOrder("desc"),
	)
	if err != nil {
		return errors.Wrap(err, "failed to get email code requests")
	}

	var (
		tryCount    int
		hasFirstReq bool
	)

	// 最初の送信履歴を探すループ
	for _, emailCodeRequest := range emailCodeRequests {

		// 最初のリクエストがあるとループを抜ける
		if emailCodeRequest.FirstReq {
			hasFirstReq = true
			break
		}

		// リクエスト数をカウント
		tryCount++
	}

	// 最大再送信回数を超えたらエラーを返す
	if tryCount >= maxResendCount {
		return cdomain.ErrRequestLimitExceeded
	}

	// firstReq（最初の送信リクエスト履歴）がない場合は再送信できません。
	if !hasFirstReq {
		return errors.New("err firstReq not found")
	}

	certificationCode, err := u.generateRandomString(6, numericChars)
	if err != nil {
		return err
	}

	// メールコードリクエストを作成
	if err = u.emailCodeRequestRepository.Create(ctx, domain.EmailCodeRequest{
		UserId:   request.JWTUser.Id,
		Email:    request.Email,
		Code:     certificationCode,
		FirstReq: false,
		Expire:   jsonutils.JSONTime(time.Now().Add(emailCodeExpireTime)),
	}); err != nil {
		return errors.Wrap(err, "failed to create email code request")
	}

	// メール送信、正常に返す
	if err = u.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.CERTIFICATION_CODE_SEND_TEMPLATE,
		Sender:       u.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{request.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_FAST,
		Details: map[string]interface{}{
			"Receiver":          request.Email,
			"ServiceUrl":        u.cfg.ServiceConfig.Url.UserServiceDomain,
			"CertificationCode": certificationCode,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to send mail")
	}

	return nil
}

func (u userVerifyUsecase) CheckEmailCode(ctx context.Context, request udomain.UserCheckEmailCodeRequest) error {
	// メールコードリクエストテーブルの確認
	emailCodeRequests, err := u.emailCodeRequestRepository.GetAll(ctx,
		cdomain.EmailCodeRequestQueryOptionFunc().WithEmail(request.Email),
		cdomain.EmailCodeRequestQueryOptionFunc().WithCode(request.Code),
	)
	if err != nil {
		return errors.Wrap(err, "failed to get email code requests")
	}

	// メールコードリクエストがない場合はエラーを返す
	if len(emailCodeRequests) == 0 {
		return cdomain.ErrItemNotFound
	}

	now := jsonutils.JSONTime(time.Now())

	// メールコードリクエストはあるが既に有効期限が切れている場合
	if now.Unix() > emailCodeRequests[0].Expire.Unix() {
		return cdomain.ErrAlreadyExpired
	}

	// 変更しようとするメールアドレスでアカウントが存在するかを確認
	isDuplicatedEmail, err := u.userService.IsDuplicateEmail(ctx, request.Email)
	if err != nil {
		return errors.Wrap(err, "failed to get user by email")
	}

	if isDuplicatedEmail {
		return cdomain.ErrAlreadyExists
	}

	// 有効期限前であり、正常に処理される認証の場合、有効期限を現在を基準として更新し、正常に返す
	if err = u.emailCodeRequestRepository.Update(ctx, domain.EmailCodeRequest{
		Id:     emailCodeRequests[0].Id,
		Expire: now,
	}); err != nil {
		return errors.Wrap(err, "failed to update email code request")
	}

	// ユーザー照会
	user, err := u.userRepository.GetById(ctx, request.JWTUser.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get user by id")
	}

	user.Email = request.Email

	if err = u.userRepository.Update(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	// 関連メールを送信
	if err = u.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.UPDATE_USER_INFORMATION_TEMPLATE,
		Sender:       u.cfg.ServiceConfig.Mail.Info,
		MailType:     domain.MAIL_SEND_PRIORITY_FAST,
		Receivers:    []string{user.Email},
		Details: map[string]interface{}{
			"Receiver":    user.Name,
			"ServiceUrl":  u.cfg.ServiceConfig.Url.UserServiceDomain,
			"Info":        "メールアドレス",
			"BeforePhone": request.JWTUser.Email,
			"AfterPhone":  user.Email,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to send mail")
	}

	return nil
}

func generateRandomString(n int, charSet string) (string, error) {
	buffer := make([]byte, n)
	charSetLength := big.NewInt(int64(len(charSet)))

	for i := 0; i < n; i++ {
		index, err := rand.Int(rand.Reader, charSetLength)
		if err != nil {
			return "", errors.Wrap(err, "failed to generate random int")
		}

		buffer[i] = charSet[index.Int64()]
	}

	return string(buffer), nil
}
