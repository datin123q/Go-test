package user

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"mime"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pwhelper"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/egift"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

const (
	userFilePath          = "user_file"
	signUpEmailExpireTime = time.Minute * 20
	signUpUrlExpireTime   = time.Minute * 20
	alphanumericChars     = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	numericChars          = "0123456789"
	maxSendCount          = 3
	maxResendCount        = 5

	// 紹介コード関連
	referralCodeLength         = 10                                // 紹介コードの文字数
	referralCodeMaxGenerateTry = 5                                 // 紹介コード生成リトライ回数
	referralCodeChars          = "0123456789ABCDEFGHJKLMNPRTUVWXY" // 紹介コードに使用する文字(I O Q S Z は視覚的に他の文字と混同しやすいので除外する)
)

type userUsecase struct {
	cfg                           *config.Config
	egift                         egift.Egift
	idGenerator                   idgenerator.IdGenerator
	s3                            s3utils.S3
	pwhelper                      pwhelper.PasswordHelper
	txManager                     factory.TxManager
	corporateProfileRepository    cdomain.CorporateProfileRepository
	loginHistoryRepository        cdomain.LoginHistoryRepository
	marketingCodeRepository       cdomain.MarketingCodeRepository
	privateProfileRepository      cdomain.PrivateProfileRepository
	roleRepository                cdomain.RoleRepository
	signUpRequestRepository       cdomain.SignUpRequestRepository
	userRepository                cdomain.UserRepository
	userAntisocialCheckRepository cdomain.UserAntisocialCheckRepository
	userAccountRepository         cdomain.UserAccountRepository
	userLoginLockRepository       cdomain.UserLoginLockRepository
	emailCodeRequestRepository    cdomain.EmailCodeRequestRepository
	authService                   cdomain.AuthService
	mailService                   cdomain.MailService
	userService                   cdomain.UserService
	codeOwnershipRepository       cdomain.CodeOwnershipRepository
	codeUsageRepository           cdomain.CodeUsageRepository

	// 内部関数
	generateRandomString func(n int, charSet string) (string, error)
}

func New(
	cfg *config.Config,
	egift egift.Egift,
	idGenerator idgenerator.IdGenerator,
	s3 s3utils.S3,
	pwhelper pwhelper.PasswordHelper,
	txManager factory.TxManager,
	corporateProfileRepository cdomain.CorporateProfileRepository,
	loginHistoryRepository cdomain.LoginHistoryRepository,
	marketingCodeRepository cdomain.MarketingCodeRepository,
	privateProfileRepository cdomain.PrivateProfileRepository,
	roleRepository cdomain.RoleRepository,
	signUpRequestRepository cdomain.SignUpRequestRepository,
	codeOwnershipRepository cdomain.CodeOwnershipRepository,
	codeUsageRepository cdomain.CodeUsageRepository,
	userRepository cdomain.UserRepository,
	userAntisocialCheckRepository cdomain.UserAntisocialCheckRepository,
	userAccountRepository cdomain.UserAccountRepository,
	userLoginLockRepository cdomain.UserLoginLockRepository,
	emailCodeRequestRepository cdomain.EmailCodeRequestRepository,
	authService cdomain.AuthService,
	mailService cdomain.MailService,
	userService cdomain.UserService,

) udomain.UserUsecase {
	return &userUsecase{
		cfg:                           cfg,
		egift:                         egift,
		idGenerator:                   idGenerator,
		s3:                            s3,
		pwhelper:                      pwhelper,
		txManager:                     txManager,
		corporateProfileRepository:    corporateProfileRepository,
		loginHistoryRepository:        loginHistoryRepository,
		marketingCodeRepository:       marketingCodeRepository,
		privateProfileRepository:      privateProfileRepository,
		roleRepository:                roleRepository,
		signUpRequestRepository:       signUpRequestRepository,
		codeOwnershipRepository:       codeOwnershipRepository,
		codeUsageRepository:           codeUsageRepository,
		userRepository:                userRepository,
		userAntisocialCheckRepository: userAntisocialCheckRepository,
		userAccountRepository:         userAccountRepository,
		userLoginLockRepository:       userLoginLockRepository,
		emailCodeRequestRepository:    emailCodeRequestRepository,
		authService:                   authService,
		mailService:                   mailService,
		userService:                   userService,

		generateRandomString: generateRandomString,
	}
}

func (u userUsecase) Login(ctx context.Context, request udomain.UserLoginRequest) (udomain.UserLoginResponse, error) {

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// メールアドレスでユーザーを照会
	user, err := u.userRepository.GetByEmail(ctx, request.Email)
	if err != nil {
		return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to get user by email")
	}

	// 「無効」状態のユーザーはログインできない
	if user.AuditStatus == domain.INVALIDITY_AUDIT {
		return udomain.UserLoginResponse{}, cdomain.ErrAuditInvalid
	}

	// 今日のログイン失敗記録を確認。MIR-1233: WithSortBy/WithOrderの引数が入れ替わっていた
	// (列名を渡すべきWithSortByへ"desc"を、並び順を渡すべきWithOrderへ"id"を渡していた)。
	// pkg/modelhelper.SetSortOrderは列名と並び順の組が不正だとエラーを返さずその組を読み飛ばし
	// (MIR-1113でこの方式になった、別チケットのスコープ)、さらにpkg/orm.SortOrderは戻り値の
	// エラーを`_ =`で捨てているため、たとえエラーが返っても表には出ない。この二重の理由で
	// ORDER BYが一切積まれないままLIMIT 1で不定の1行を取得していた。「今日の最新の失敗記録」を
	// 取るにはWithSortBy("id")・WithOrder("desc")が正しい。
	var failCount int64
	total, userLoginLocks, err := u.userLoginLockRepository.GetAll(ctx,
		cdomain.UserLoginLockQueryOptionFunc().WithUserId(user.Id),
		cdomain.UserLoginLockQueryOptionFunc().WithStartDate(startOfDay.Format(time.DateTime)),
		cdomain.UserLoginLockQueryOptionFunc().WithSortBy("id"),
		cdomain.UserLoginLockQueryOptionFunc().WithOrder("desc"),
		cdomain.UserLoginLockQueryOptionFunc().WithLimit(1),
		cdomain.UserLoginLockQueryOptionFunc().WithOffset(0),
	)
	if err != nil {
		return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to get user login locks")
	}

	if total == 0 {
		failCount = int64(0)
	} else {
		failCount = userLoginLocks[0].LoginFailCnt
	}

	// 7回以上失敗したユーザーは30分経過後にログイン試行可能
	if failCount >= domain.LOGIN_FAIL_LOCK_CNT {
		if userLoginLocks[0].LockAt.Time().After(now.Add(-domain.LOGIN_LOCK_MINUTES * time.Minute)) {
			return udomain.UserLoginResponse{}, cdomain.ErrResourceLocked
		}
	}

	// パスワード確認
	inputPassword := u.pwhelper.EncodePassword(request.Password, user.Salt)
	if !u.pwhelper.IsSamePassword(user.Password, inputPassword) {
		// ログイン失敗記録はメインセッションに関わらずDBに保存する必要があるため、WithTxを使用して別のセッションにまとめる

		failCount++

		// パスワード不一致 - ログイン失敗記録を作成
		if err := u.txManager.WithTx(ctx, func(txCtx context.Context) error {
			if _, err := u.loginHistoryRepository.Create(ctx, domain.LoginHistory{
				UserId:      user.Id,
				Name:        user.Name,
				Email:       request.Email,
				IpAddr:      request.ClientIp,
				Agent:       request.UserAgent,
				LoginResult: domain.FAIL_LOGIN,
				LoginTime:   jsonutils.JSONTime(now),
			}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to create login history")
		}

		// パスワード不一致 - ログイン失敗回数チェック
		if total == 0 {
			if err := u.txManager.WithTx(ctx, func(txCtx context.Context) error {
				if _, err := u.userLoginLockRepository.Create(txCtx, domain.UserLoginLock{
					UserId:       user.Id,
					LoginFailCnt: failCount,
				}); err != nil {
					return err
				}
				return nil

			}); err != nil {
				return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to create user login lock")
			}

		} else {
			userLoginLocks[0].LoginFailCnt = failCount
			if userLoginLocks[0].LoginFailCnt >= domain.LOGIN_FAIL_LOCK_CNT {
				userLoginLocks[0].LockAt = jsonutils.JSONTime(now)
			}

			if err := u.txManager.WithTx(ctx, func(txCtx context.Context) error {
				if err := u.userLoginLockRepository.Update(txCtx, userLoginLocks[0]); err != nil {
					return err
				}
				return nil

			}); err != nil {
				return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to update user login lock")
			}
		}

		if failCount >= domain.LOGIN_FAIL_LOCK_CNT {
			// ログイン30分ロック案内メール送信
			if err := u.mailService.SendMail(ctx, cdomain.SendMailRequest{
				TemplateType: domain.USER_LOGIN_FAIL_LOCK_TEMPLATE,
				Sender:       u.cfg.ServiceConfig.Mail.Info,
				Receivers:    []string{user.Email},
				MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
				Details: map[string]interface{}{
					"ServiceUrl": u.cfg.ServiceConfig.Url.UserServiceDomain,
					"Receiver":   user.Name,
				},
			}); err != nil {
				return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to send mail")
			}
			// 7回以上: ログインロックエラーを返す
			return udomain.UserLoginResponse{}, cdomain.ErrResourceLocked
		} else if failCount >= domain.LOGIN_FAIL_WARN_CNT {
			// 4回以上: ログイン失敗警告エラーを返す
			return udomain.UserLoginResponse{}, cdomain.ErrTooManyRequests
		} else {
			// 4回未満: パスワード誤りエラーを返す
			return udomain.UserLoginResponse{}, cdomain.ErrItemInvalid
		}
	}

	// パスワード一致 - 既存ログイン失敗カウント初期化
	if total != 0 {
		if err := u.userLoginLockRepository.Delete(ctx, userLoginLocks[0].Id); err != nil {
			return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to delete user login lock")
		}
	}

	// パスワード一致 - アクセストークンを作成
	accessToken, err := u.authService.CreateUserAccessToken(ctx, user)
	if err != nil {
		return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to create access token")
	}

	// パスワード一致 - リフレッシュトークンを作成
	refreshToken, err := u.authService.CreateUserRefreshToken(ctx, cdomain.UserRefreshTokenInput{
		User:  user,
		IP:    request.ClientIp,
		Agent: request.UserAgent,
	})
	if err != nil {
		return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to create refresh token")
	}

	// パスワード一致 - CDNキャッシュ制御用Cookie作成
	cookies, err := u.s3.CreateCookieSign()
	if err != nil {
		return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to get cookie sign")
	}

	// 初回ログイン判定
	var isFirstTimeLoginAfterAudit bool
	// 会員ステータスが本登録承認完了かつ、最終ログイン日が本登録承認日前の場合は初回ログイン
	if user.AuditStatus == domain.COMPLETED_AUDIT && user.LastLoginDate.Time().Before(user.SignUpApprovalDate.Time()) {
		isFirstTimeLoginAfterAudit = true
	}

	// パスワード一致 - ユーザーログイン情報を更新。last_login_date 以外の列は書かない。
	// Update(全列書き戻し)だと、冒頭でGetByEmailした時点から書き込みまでの間に
	// eKYC Webhookなどが書いた is_ekyc・audit_status を古い値へ戻してしまう(MIR-1373)。
	if err := u.userRepository.UpdateLastLoginDate(ctx, user.Id, jsonutils.JSONTime(now)); err != nil {
		return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to update user")
	}

	// パスワード一致 - ログイン成功履歴を作成
	if _, err := u.loginHistoryRepository.Create(ctx, domain.LoginHistory{
		UserId:      user.Id,
		Name:        user.Name,
		Email:       request.Email,
		IpAddr:      request.ClientIp,
		Agent:       request.UserAgent,
		LoginResult: domain.SUCCESS_LOGIN,
		LoginTime:   jsonutils.JSONTime(now),
	}); err != nil {
		return udomain.UserLoginResponse{}, errors.Wrap(err, "failed to create login history")
	}

	return udomain.UserLoginResponse{
		AccessToken:                accessToken,
		RefreshToken:               refreshToken,
		Cookies:                    cookies,
		IsFirstTimeLoginAfterAudit: isFirstTimeLoginAfterAudit,
	}, nil
}

func (u userUsecase) RequestSignUp(ctx context.Context, request udomain.UserRequestSignUpRequest) (udomain.UserRequestSignUpResponse, error) {
	var (
		requestId  = u.idGenerator.GenUuidV4().String()
		serviceUrl = u.cfg.ServiceConfig.Url.UserServiceDomain
	)

	// 既に非アクティブ化されたアカウントのメールアドレスを含めて確認します。
	isDuplicatedEmail, err := u.userService.IsDuplicateEmailIncludeDeleted(ctx, request.Email)
	if err != nil {
		return udomain.UserRequestSignUpResponse{}, errors.Wrap(err, "failed to get user by email")
	}
	if isDuplicatedEmail {
		return udomain.UserRequestSignUpResponse{}, cdomain.ErrAlreadyExists
	}

	signUpRequests, err := u.signUpRequestRepository.GetAll(ctx,
		cdomain.SignUpRequestQueryOptionFunc().WithEmail(request.Email),
		cdomain.SignUpRequestQueryOptionFunc().WithEmailExpire(jsonutils.JSONTime(time.Now())),
		cdomain.SignUpRequestQueryOptionFunc().WithSortBy("id"),
		cdomain.SignUpRequestQueryOptionFunc().WithOrder("desc"),
	)
	if err != nil {
		return udomain.UserRequestSignUpResponse{}, errors.Wrap(err, "failed to get sign up requests")
	}

	// メール送信制限を超えた場合はエラーを返します。
	if signUpRequests.IsSentMailLimitExceeded() {
		return udomain.UserRequestSignUpResponse{}, cdomain.ErrRequestLimitExceeded
	}

	// sign up request がない場合は新規作成し、ある場合は send count を増加させます。
	if signUpRequests.IsEmpty() {
		if _, err = u.signUpRequestRepository.Create(ctx, domain.SignUpRequest{
			RequestId:    requestId,
			Email:        request.Email,
			Url:          serviceUrl + "/sign-up/" + requestId,
			UserType:     request.UserType,
			EmailExpire:  jsonutils.JSONTime(time.Now().Add(signUpEmailExpireTime)), // 20분
			UrlExpire:    jsonutils.JSONTime(time.Now().Add(signUpUrlExpireTime)),   // 20분
			ReferralCode: request.ReferralCode,
		}); err != nil {
			return udomain.UserRequestSignUpResponse{}, errors.Wrap(err, "failed to create sign up request")
		}

		// マーケティング関連データがある場合は、e-giftにリワードを要求します。
		if request.HasMarketing() {
			if err = u.egift.RewardSignUp(ctx, request.MarketingCode, requestId); err != nil {
				return udomain.UserRequestSignUpResponse{}, errors.Wrap(err, "failed to reward sign up")
			}

			if _, err = u.marketingCodeRepository.Create(ctx, domain.MarketingCode{
				Code:            request.MarketingCode,
				Campaign:        request.MarketingCampaign,
				SignUpRequestId: requestId,
				Email:           request.Email,
			}); err != nil {
				return udomain.UserRequestSignUpResponse{}, errors.Wrap(err, "failed to create marketing code")
			}
		}
	} else {
		signUpRequests[0].SendCount++
		if request.ReferralCode != "" {
			signUpRequests[0].ReferralCode = request.ReferralCode
		}
		requestId = signUpRequests[0].RequestId
		if err = u.signUpRequestRepository.Update(ctx, signUpRequests[0]); err != nil {
			return udomain.UserRequestSignUpResponse{}, errors.Wrap(err, "failed to update sign up request")
		}
	}

	// SIGN_UP_JAメール送信
	if err = u.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.SIGN_UP_TEMPLATE,
		Sender:       u.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{request.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_FAST,
		Details: map[string]interface{}{
			"SignUpUrl":  serviceUrl + "/sign-up/" + requestId,
			"ServiceUrl": serviceUrl,
		},
	}); err != nil {
		return udomain.UserRequestSignUpResponse{}, errors.Wrap(err, "failed to send mail")
	}

	return udomain.UserRequestSignUpResponse{RequestId: requestId}, nil
}

func (u userUsecase) CheckSignupRequestValidity(ctx context.Context, request udomain.UserCheckSignupRequestValidityRequest) (udomain.UserCheckSignupRequestValidityResponse, error) {
	signUpRequest, err := u.signUpRequestRepository.GetByRequestId(ctx, request.RequestId)
	if err != nil {
		return udomain.UserCheckSignupRequestValidityResponse{}, err
	}

	// 有効期限を確認（email_expire、url_expireは一緒に登録されるため、url有効期限のみ確認すれば問題なし）
	if time.Now().After(signUpRequest.UrlExpire.Time()) {
		return udomain.UserCheckSignupRequestValidityResponse{}, cdomain.ErrAlreadyExpired
	}

	return udomain.UserCheckSignupRequestValidityResponse{
		Email:    signUpRequest.Email,
		UserType: signUpRequest.UserType,
	}, nil
}

func (u userUsecase) VerifyEmail(ctx context.Context, request udomain.UserVerifyEmailRequest) error {
	signUpRequest, err := u.signUpRequestRepository.GetByRequestId(ctx, request.RequestId)
	if err != nil {
		return err
	}

	// 有効期限を確認（email_expire、url_expireは一緒に登録されるため、url有効期限のみ確認すれば問題なし）
	if time.Now().After(signUpRequest.UrlExpire.Time()) {
		return cdomain.ErrAlreadyExpired
	}

	// 有効期限処理：メール有効期限を現在時刻より前に設定
	signUpRequest.EmailExpire = jsonutils.JSONTime(time.Now().Add(-time.Second))

	if err = u.signUpRequestRepository.Update(ctx, signUpRequest); err != nil {
		return errors.Wrap(err, "failed to update sign up request")
	}

	return nil
}

func (u userUsecase) SignUp(ctx context.Context, request udomain.UserSignUpRequest) error {
	var (
		uid      = u.idGenerator.GenUuidV4()
		commonId = u.idGenerator.GenUuidV5(uid, request.Email).String()
		name     string
	)

	// 個人の場合はFamilyName + FirstName、法人の場合はCorpNameを使用します。
	if request.IsUserTypePrivate() {
		name = fmt.Sprintf("%s %s", request.FamilyName, request.FirstName)
	} else {
		name = request.CorpName
	}

	signUpRequest, err := u.signUpRequestRepository.GetByRequestId(ctx, request.RequestId)
	if err != nil {
		return errors.Wrap(err, "failed to get sign up request")
	}

	// サインアップリクエストメールと異なるメールアドレスで登録しようとするとエラーが返されます。
	if signUpRequest.Email != request.Email {
		return cdomain.ErrItemConflict
	}

	salt := u.pwhelper.GenerateSalt()
	encodePassword := u.pwhelper.EncodePassword(request.Password, salt)
	user := domain.User{
		Uid:          commonId, // id 발급 전 임시로 사용
		UserType:     domain.USER_TYPE(request.UserType),
		CommonId:     commonId,
		Email:        request.Email,
		Name:         name,
		Password:     encodePassword,
		Salt:         salt,
		SignupStatus: domain.SIGNUP_NONE,
		SignUpDate:   jsonutils.JSONTime(time.Now()),
		LoginType:    domain.MAIL,
		AuditStatus:  domain.PROVISIONAL_AUDIT,
	}

	// ニュースレター受信同意の有無に応じたデータ追加
	if request.IsNewsLetter {
		user.NewsletterAgreeAt = jsonutils.JSONTime(time.Now())
		user.IsNewsletter = request.IsNewsLetter
	}

	userId, err := u.userRepository.Create(ctx, user)
	if err != nil {
		return errors.Wrap(err, "failed to create user")
	}

	user.Id = userId
	user.Uid = u.idGenerator.GenUID(userId)

	// 自身の紹介コード生成
	var referralCode string
	for i := 0; i < referralCodeMaxGenerateTry; i++ {
		referralCode, err = u.generateRandomString(referralCodeLength, referralCodeChars)
		if err != nil {
			return errors.Wrap(err, "failed to generate referral code")
		}

		existingCodeOwnership, err := u.codeOwnershipRepository.GetByCode(ctx, domain.CodeTypeReferral, referralCode)
		if err != nil {
			return errors.Wrap(err, "failed to get code ownership by code")
		}
		if existingCodeOwnership == nil {
			break
		}

		referralCode = ""
	}

	if referralCode == "" {
		return errors.Wrap(cdomain.ErrServiceUnavailable, "failed to generate unique referral code")
	}

	if err := u.codeOwnershipRepository.Create(ctx, domain.CodeOwnership{
		OwnerType: domain.USER_TYPE(request.UserType),
		OwnerId:   userId,
		CodeType:  domain.CodeTypeReferral,
		Code:      referralCode,
	}); err != nil {
		return errors.Wrap(err, "failed to create code ownership")
	}

	// 紹介コードの利用を記録
	if signUpRequest.ReferralCode != "" {
		existingCodeOwnership, err := u.codeOwnershipRepository.GetByCode(ctx, domain.CodeTypeReferral, signUpRequest.ReferralCode)
		if err != nil {
			return errors.Wrap(err, "failed to get code ownership")
		} else if existingCodeOwnership == nil {
			// TODO: フロントエンドにバリデーションチェックを追加するときに合わせてバックエンド側も対応する。
			// 存在しない紹介コードが入力された場合は警告ログだけ出力し、処理を継続する。
			factory.Logger(ctx).WithField("referral_code", signUpRequest.ReferralCode).Warn("referral code not found")
		} else {
			codeInputAt := jsonutils.JSONTime(time.Now())
			if err := u.codeUsageRepository.Create(ctx, domain.CodeUsage{
				UsedOwnerType:   domain.USER_TYPE(request.UserType),
				UsedOwnerId:     userId,
				CodeOwnershipId: existingCodeOwnership.Id,
				CodeInputAt:     &codeInputAt,
			}); err != nil {
				return errors.Wrap(err, "failed to create code usage")
			}
		}
	}

	// 基本テーブル設定 - 反社会勢力確認作成
	if err = u.userAntisocialCheckRepository.Create(ctx, domain.UserAntisocialCheck{
		CheckUserId: userId,
	}); err != nil {
		return errors.Wrap(err, "failed to create user antisocial check")
	}

	// 基本テーブル設定 - ユーザーアカウント作成
	if err = u.userAccountRepository.Create(ctx, domain.UserAccount{
		UserId: userId,
	}); err != nil {
		return errors.Wrap(err, "failed to create user account")
	}

	// 基本テーブル設定 - 個人/法人プロフィール
	switch domain.USER_TYPE(request.UserType) {
	case domain.PRIVATE:
		user.InvestorType = domain.PRIVATE_INVESTOR

		if err = u.privateProfileRepository.Create(ctx, domain.PrivateProfile{
			UserId:         userId,
			FirstName:      request.FirstName,
			FamilyName:     request.FamilyName,
			FirstNameKana:  request.FirstNameKana,
			FamilyNameKana: request.FamilyNameKana,
			RequestId:      request.RequestId,
		}); err != nil {
			return errors.Wrap(err, "failed to create private profile")
		}

	case domain.CORPORATE:
		user.InvestorType = domain.CORPORATE_INVESTOR

		if _, err = u.corporateProfileRepository.Create(ctx, domain.CorporateProfile{
			UserId:                       userId,
			ManagerFirstName:             lo.EmptyableToPtr(request.ManagerFirstName),
			ManagerFamilyName:            lo.EmptyableToPtr(request.ManagerFamilyName),
			ManagerFirstNameKana:         lo.EmptyableToPtr(request.ManagerFirstNameKana),
			ManagerFamilyNameKana:        lo.EmptyableToPtr(request.ManagerFamilyNameKana),
			RepresentativeFirstName:      request.FirstName,
			RepresentativeFamilyName:     request.FamilyName,
			RepresentativeFirstNameKana:  request.FirstNameKana,
			RepresentativeFamilyNameKana: request.FamilyNameKana,
		}); err != nil {
			return errors.Wrap(err, "failed to create corporate profile")
		}

	default:
		return cdomain.ErrBadRequest
	}

	if err = u.userRepository.Update(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	// メール有効期限を現在時刻より前に設定
	signUpRequest.UrlExpire = jsonutils.JSONTime(time.Now().Add(-time.Second))

	if err = u.signUpRequestRepository.Update(ctx, signUpRequest); err != nil {
		return errors.Wrap(err, "failed to update sign up request")
	}

	// 会員登録完了メール送信
	templateType := domain.PROVISIONAL_PRIVATE_SIGN_UP_COMPLETE_TEMPLATE
	if domain.USER_TYPE(request.UserType) == domain.CORPORATE {
		templateType = domain.PROVISIONAL_CORPORATE_SIGN_UP_COMPLETE_TEMPLATE
	}

	if err = u.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: templateType,
		Sender:       u.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{request.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"Receiver":   name,
			"ServiceUrl": u.cfg.ServiceConfig.Url.UserServiceDomain,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to send mail")
	}

	return nil
}

func (u userUsecase) CreatePreSignedUrl(ctx context.Context, request udomain.UserCreatePreSignedUrlRequest) (udomain.UserCreatePreSignedUrlResponse, error) {
	now := time.Now()
	filePath := userFilePath + "/" + now.Format(time.DateOnly)

	escapedFilename := url.QueryEscape(request.FileName)
	escapedFilename = strings.ReplaceAll(escapedFilename, "+", " ")
	contentType := mime.TypeByExtension(path.Ext(request.FileName))

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"", escapedFilename)

	urlString, err := u.s3.CreatePreSignedUrl(ctx, s3utils.CreatePreSignedUrlRequest{
		BucketName:         u.cfg.AWS.S3_PRIVATE_BUCKET,
		FilePath:           filePath,
		FileName:           request.FileName,
		ContentType:        aws.String(contentType),
		ContentDisposition: aws.String(contentDisposition),
	})
	if err != nil {
		return udomain.UserCreatePreSignedUrlResponse{}, err
	}

	return udomain.UserCreatePreSignedUrlResponse{
		PreSignedUrl:       urlString,
		ExpiredAt:          jsonutils.JSONTime(now.Add(time.Minute * 30)),
		ContentType:        contentType,
		ContentDisposition: contentDisposition,
	}, nil
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
