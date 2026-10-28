package private_user

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type privateUserUsecase struct {
	cfg                          *config.Config
	idGenerator                  idgenerator.IdGenerator
	logger                       factory.LoggerFactory
	mailService                  cdomain.MailService
	myNumberService              cdomain.MyNumberService
	nexwayService                cdomain.NexwayService
	userService                  cdomain.UserService
	userFileService              cdomain.UserFileService
	userVirtualAccountService    cdomain.UserVirtualAccountService
	ekycRequestLogRepository     cdomain.EkycRequestLogRepository
	myNumberRepository           cdomain.MyNumberRepository
	privateProfileRepository     cdomain.PrivateProfileRepository
	privateUserRepository        cdomain.PrivateUserRepository
	userRepository               cdomain.UserRepository
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository
}

func New(
	cfg *config.Config,
	idGenerator idgenerator.IdGenerator,
	logger factory.LoggerFactory,
	mailService cdomain.MailService,
	myNumberService cdomain.MyNumberService,
	nexwayService cdomain.NexwayService,
	userService cdomain.UserService,
	userFileService cdomain.UserFileService,
	userVirtualAccountService cdomain.UserVirtualAccountService,
	ekycRequestLogRepository cdomain.EkycRequestLogRepository,
	myNumberRepository cdomain.MyNumberRepository,
	privateProfileRepository cdomain.PrivateProfileRepository,
	privateUserRepository cdomain.PrivateUserRepository,
	userRepository cdomain.UserRepository,
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository,
) udomain.PrivateUserUsecase {
	return &privateUserUsecase{
		cfg:                          cfg,
		idGenerator:                  idGenerator,
		logger:                       logger,
		mailService:                  mailService,
		myNumberService:              myNumberService,
		nexwayService:                nexwayService,
		userService:                  userService,
		userFileService:              userFileService,
		userVirtualAccountService:    userVirtualAccountService,
		ekycRequestLogRepository:     ekycRequestLogRepository,
		myNumberRepository:           myNumberRepository,
		privateProfileRepository:     privateProfileRepository,
		privateUserRepository:        privateUserRepository,
		userRepository:               userRepository,
		userVirtualAccountRepository: userVirtualAccountRepository,
	}
}

var signUpStatusMap = map[domain.SIGNUP_STATUS]int{
	domain.SIGNUP_NONE:                   1,
	domain.SIGNUP_PRIVATE_PROFILE_BASIC:  2,
	domain.SIGNUP_PRIVATE_PROFILE_OTHER:  3,
	domain.SIGNUP_ACCOUNT:                4,
	domain.SIGNUP_INVESTMENT_SUITABILITY: 5,
	domain.SIGNUP_AGREE:                  6,
	// domain.SIGNUP_MYNUMBER:               7,
	domain.SIGNUP_KYC:      7,
	domain.SIGNUP_COMPLETE: 8,
}

// isSignupReentryAllowed は本登録の再入力を許す審査状態かを返す。
// 保留と仮登録への差し戻しは、ユーザーに入力し直させるための状態。
// 管理画面からこれらへ変更しても signup_status は据え置かれるため、
// signup_status だけで完了判定すると再登録できなくなる。
func isSignupReentryAllowed(auditStatus domain.AUDIT_STATUS) bool {
	return auditStatus == domain.HOLD_AUDIT || auditStatus == domain.PROVISIONAL_AUDIT
}

// isAuditStatusReentryBlocked は、審査が確定または反社確認等で進行中のため、
// signup_status の完了有無にかかわらずプロフィール再入力を許さない審査状態かを返す。
// エラーメッセージで審査状態そのものを開示しないよう、呼び出し側では共通の
// メッセージにまとめる(無効化・反社確認中であることが利用者に伝わらないようにする)。
//
// WAIT_AUDIT・WAIT_IDENTITY_REGISTER_AUDIT は対象外。後者は本登録フロー中に
// 自身が経由する正常な状態(このためだけに isSignupReentryAllowed が存在する)。
// 前者を対象に含めるかは要判断のため、対象外のまま残している。
func isAuditStatusReentryBlocked(auditStatus domain.AUDIT_STATUS) bool {
	switch auditStatus {
	case domain.INVALIDITY_AUDIT,
		domain.WAIT_FINAL_APPROVAL_AUDIT,
		domain.WAIT_ANTISOCIAL_AUDIT,
		domain.CHECK_ANTISOCIAL_AUDIT:
		return true
	default:
		return false
	}
}

var signUpStatusRMap = map[int]domain.SIGNUP_STATUS{
	1: domain.SIGNUP_NONE,
	2: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
	3: domain.SIGNUP_PRIVATE_PROFILE_OTHER,
	4: domain.SIGNUP_ACCOUNT,
	5: domain.SIGNUP_INVESTMENT_SUITABILITY,
	6: domain.SIGNUP_AGREE,
	// 7: domain.SIGNUP_MYNUMBER,
	7: domain.SIGNUP_KYC,
	8: domain.SIGNUP_COMPLETE,
}

var jstLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	return loc
}()

func (p privateUserUsecase) GetSignUpStatus(ctx context.Context, request udomain.PrvUserGetSignUpStatusRequest) (udomain.PrvUserGetSignUpStatusResponse, error) {

	// get private user
	privateUser, err := p.privateUserRepository.GetById(ctx, request.JWTUser.Id)
	if err != nil {
		return udomain.PrvUserGetSignUpStatusResponse{}, errors.Wrap(err, "failed to get private user by id")
	}

	if len(request.Status) == 0 {
		request.Status = privateUser.SignupStatus
	}
	signUpSequence := signUpStatusMap[request.Status]
	var response = udomain.PrvUserGetSignUpStatusResponse{
		Id:         privateUser.User.Id,
		Email:      privateUser.User.Email,
		PrevStatus: signUpStatusRMap[signUpSequence-1],
		NextStatus: signUpStatusRMap[signUpSequence+1],
		CurStatus:  request.Status,
	}

	switch request.Status {
	case domain.SIGNUP_NONE, domain.SIGNUP_AGREE, domain.SIGNUP_KYC, domain.SIGNUP_COMPLETE:
		return response, nil

	case domain.SIGNUP_PRIVATE_PROFILE_BASIC:
		data := udomain.PrvUserBasicProfile{}.FromPrivateUser(privateUser)
		response.BasicProfile = &data

	case domain.SIGNUP_PRIVATE_PROFILE_OTHER:
		data := udomain.PrvUserOtherProfile{}.FromPrivateUser(privateUser)
		response.OtherProfile = &data

	case domain.SIGNUP_ACCOUNT:
		data := udomain.PrvUserWithdrawalAccount{}.FromPrivateUser(privateUser)
		response.WithdrawalAccount = &data

	case domain.SIGNUP_INVESTMENT_SUITABILITY:
		data := udomain.PrvUserInvestmentSuitability{}.FromPrivateUser(privateUser)
		response.InvestmentSuitability = &data

	// TODO: マイナンバー登録復活後に元に戻す
	// case domain.SIGNUP_MYNUMBER:
	// 	myNumber, err := p.myNumberService.GetByUserId(ctx, privateUser.User.Id)
	// 	if err != nil {
	// 		return udomain.PrvUserGetSignUpStatusResponse{}, errors.Wrap(err, "failed to get my number")
	// 	}

	// 	if myNumber.IsEmpty() {
	// 		response.MyNumber = &udomain.PrvUserMyNumber{}
	// 	} else {
	// 		data := udomain.PrvUserMyNumber{}.FromMyNumber(myNumber)
	// 		response.MyNumber = &data
	// 	}

	default:
		return udomain.PrvUserGetSignUpStatusResponse{}, errors.Wrap(cdomain.ErrBadRequest, "invalid sign up status")
	}

	return response, nil
}

func (p privateUserUsecase) RegisterProfile(ctx context.Context, request udomain.PrvUserRegisterProfileRequest) (udomain.PrvUserRegisterProfileResponse, error) {
	var err error

	// get private user
	privateUser, err := p.privateUserRepository.GetById(ctx, request.JWTUser.Id)
	if err != nil {
		return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(err, "failed to get private user by id")
	}

	// 本登録が完了している場合、プロフィール入力で状態を巻き戻さない。
	// 前の段階への再入力は登録途中の修正導線として許しているが、完了後に許すと
	// SIGNUP_AGREE で audit_status が本人確認待ちへ戻ってしまう。
	// ただし保留・仮登録への差し戻しは再入力させるための状態なので対象外にする。
	// 管理画面からこれらに変更しても signup_status は完了のまま残るため、
	// ここで弾くと再登録できなくなる。
	if !isSignupReentryAllowed(privateUser.AuditStatus) {
		if privateUser.SignupStatus == domain.SIGNUP_COMPLETE {
			return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(
				cdomain.ErrItemConflict,
				fmt.Sprintf("user's signup status is already %s", domain.SIGNUP_COMPLETE),
			)
		}
		if privateUser.AuditStatus == domain.COMPLETED_AUDIT {
			return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(
				cdomain.ErrItemConflict,
				"user's audit status is already completed",
			)
		}
		if isAuditStatusReentryBlocked(privateUser.AuditStatus) {
			return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(
				cdomain.ErrItemConflict,
				"user's audit status does not allow profile reentry",
			)
		}
	}

	signUpStatusSequence := signUpStatusMap[privateUser.SignupStatus]
	// 現在のプロフィール内で管理されていない状態の場合、現在より前の段階にあると仮定し、最大値を設定します。(INT16:32767)
	if signUpStatusSequence == 0 {
		signUpStatusSequence = math.MaxInt16
	}
	// プロフィール更新は前の段階または直後のステップリクエストのみ有効にリクエストできます。
	// より前の段階をリクエストされた場合、エラーを返す
	// ex1) 現在の状態がBASICの場合、OTHERリクエストは可能だがACCOUNTリクエストは不可
	// ex2) 現在の状態がOTHERの場合、BASIC、ACCOUNTリクエストは可能だがMY_NUMBERリクエストは不可
	if signUpStatusSequence+1 < signUpStatusMap[request.SignUpStatus] {
		return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(cdomain.ErrItemConflict, "invalid sign up status")
	}

	response := udomain.PrvUserRegisterProfileResponse{}
	switch request.SignUpStatus {
	case domain.SIGNUP_PRIVATE_PROFILE_BASIC:
		// 登録条件確認（申請日基準で18歳以上80歳未満）
		birthday, err := jsonutils.ParseDateTimeJST(time.DateOnly, request.BasicProfile.Birthday)
		if err != nil {
			return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(err, "failed to parse birthday")
		}

		applicationDate, err := jsonutils.ParseDateTimeJST(time.DateOnly, time.Now().In(jstLocation).Format(time.DateOnly))
		if err != nil {
			return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(err, "failed to parse application date")
		}

		youngerThan18Threshold := applicationDate.AddDate(-domain.MIN_INVESTOR_AGE_INCLUSIVE, 0, 0)
		olderThanOrEqual80Threshold := applicationDate.AddDate(-domain.MAX_INVESTOR_AGE_EXCLUSIVE, 0, 0)
		if birthday.After(youngerThan18Threshold) || !birthday.After(olderThanOrEqual80Threshold) {
			return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(cdomain.ErrServiceUnavailable, "birthday is not valid")
		}

		err = setBasicProfile(request.BasicProfile, &privateUser)

	case domain.SIGNUP_PRIVATE_PROFILE_OTHER:
		err = setOtherProfile(request.OtherProfile, &privateUser)

	case domain.SIGNUP_ACCOUNT:
		err = setWithdrawalAccount(request.WithdrawalAccount, &privateUser)

	case domain.SIGNUP_INVESTMENT_SUITABILITY:
		// 登録条件確認
		if request.InvestmentSuitability.Asset == domain.ASSET1 && request.InvestmentSuitability.AnnualIncome == domain.ANNUALINCOME1 {
			return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(cdomain.ErrServiceUnavailable, "asset and annual income is not valid")
		}
		err = setInvestmentSuitability(request.InvestmentSuitability, &privateUser)
	case domain.SIGNUP_AGREE:
		// MIR-1121: missingEkycProfileFieldsで検証せずに状態遷移(AuditStatusの
		// 変更・ekyc_request_logの作成)まで進めてしまうと、フロントエンド側の
		// 検証(resolveEkycSendValues)でeKYC送信自体はブロックできても、この
		// 会員はDB上「本人確認待ち」のまま既に確定しており、チケットが問題に
		// していた滞留と見分けが付かなくなる。状態を書き換える前にここで
		// 弾くことで、そもそも状態遷移が起きないようにする。
		if missing := missingEkycProfileFields(privateUser); len(missing) > 0 {
			return udomain.PrvUserRegisterProfileResponse{}, errors.Wrapf(
				cdomain.ErrProfileIncompleteForEkyc, "profile is incomplete for ekyc: missing %s", strings.Join(missing, ","),
			)
		}

		privateUser.SignupStatus = domain.SIGNUP_AGREE
		privateUser.AuditStatus = domain.WAIT_IDENTITY_REGISTER_AUDIT
		ekycReferenceId, err := p.verifyWithEkyc(ctx, privateUser)
		if err != nil {
			return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(err, "failed to verify with ekyc")
		}
		response.EKycReferenceId = ekycReferenceId
		// GetById直後の値をそのままフロントエンドへの一次ソースとして返す。
		response.FamilyName = privateUser.FamilyName
		response.FirstName = privateUser.FirstName
		response.Birthday = privateUser.Birthday.ToFormat(time.DateOnly)
		response.Address = buildEkycAddress(privateUser.Province, privateUser.City, privateUser.Address, privateUser.Address2)
	default:
		err = errors.Wrap(cdomain.ErrBadRequest, "invalid sign up status")
	}

	if err != nil {
		return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(err, "failed to set profile")
	}

	// update
	if err := p.userRepository.Update(ctx, privateUser.User); err != nil {
		return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(err, "failed to update user")
	}
	if err := p.privateProfileRepository.Update(ctx, privateUser.PrivateProfile); err != nil {
		return udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(err, "failed to update private profile")
	}

	return response, nil
}

// missingEkycProfileFields はSIGNUP_AGREE遷移直前のeKYC送信必須項目
// (氏名・生年月日・province/city/address、address2は任意)のうち空欄の
// ものをフィールド名で返す。欠落が無ければnilを返す(送信可能)。空白のみの
// 入力はsetOtherProfileがAddress/Address2しかTrimしないため、ここでも
// TrimSpaceして拾う。
func missingEkycProfileFields(privateUser domain.PrivateUser) []string {
	var missing []string
	if strings.TrimSpace(privateUser.FamilyName) == "" {
		missing = append(missing, "family_name")
	}
	if strings.TrimSpace(privateUser.FirstName) == "" {
		missing = append(missing, "first_name")
	}
	if privateUser.Birthday.IsZero() {
		missing = append(missing, "birthday")
	}
	if strings.TrimSpace(privateUser.Province) == "" {
		missing = append(missing, "province")
	}
	if strings.TrimSpace(privateUser.City) == "" {
		missing = append(missing, "city")
	}
	if strings.TrimSpace(privateUser.Address) == "" {
		missing = append(missing, "address")
	}
	return missing
}

// buildEkycAddress は都道府県・市区町村・住所・建物名(任意)を1つの住所文字列に
// 組み立てる(MIR-1121)。組み立て自体は、カ方式(MIR-1470)と共有するため
// domain.BuildEkycAddress にある。
func buildEkycAddress(province, city, address string, address2 *string) string {
	return domain.BuildEkycAddress(province, city, address, address2)
}

// setBasicProfile 登録 - 本登録申請 - B. 個人情報1（SIGNUP_PRIVATE_PROFILE_BASIC）
func setBasicProfile(input udomain.PrvUserBasicProfile, privateUser *domain.PrivateUser) error {

	birthday, err := jsonutils.ParseJSONTimeJST(time.DateOnly, input.Birthday)
	if err != nil {
		return errors.Wrap(err, "failed to parse birthday")
	}

	privateUser.Name = input.FamilyName + " " + input.FirstName
	privateUser.FirstName = input.FirstName
	privateUser.FamilyName = input.FamilyName
	privateUser.FirstNameKana = input.FirstNameKana
	privateUser.FamilyNameKana = input.FamilyNameKana
	privateUser.Birthday = birthday
	privateUser.Gender = input.Gender
	privateUser.Nationality = input.Nationality
	privateUser.SignupStatus = domain.SIGNUP_PRIVATE_PROFILE_BASIC

	return nil
}

// setOtherProfile 登録 - 本登録申請 - B. 個人情報2（SIGNUP_PRIVATE_PROFILE_OTHER）
func setOtherProfile(input udomain.PrvUserOtherProfile, privateUser *domain.PrivateUser) error {

	if len(strings.TrimSpace(lo.FromPtr(input.Address2))) == 0 {
		privateUser.Address2 = nil
	} else {
		privateUser.Address2 = lo.ToPtr(strings.TrimSpace(lo.FromPtr(input.Address2)))
	}

	privateUser.ZipCode = input.ZipCode
	privateUser.Address = strings.TrimSpace(input.Address)
	privateUser.Province = input.Province
	privateUser.City = input.City
	privateUser.MobilePhone = input.MobilePhone
	privateUser.Job = input.Job
	privateUser.SignupStatus = domain.SIGNUP_PRIVATE_PROFILE_OTHER

	return nil
}

// setWithdrawalAccount 登録 - 本登録申請 - C. 出金口座情報（SIGNUP_ACCOUNT）
func setWithdrawalAccount(input udomain.PrvUserWithdrawalAccount, privateUser *domain.PrivateUser) error {

	privateUser.BankName = input.BankName
	privateUser.BranchName = input.BranchName
	privateUser.BankAccount = input.BankAccount
	privateUser.BankCode = &input.BankCode
	privateUser.BranchCode = &input.BranchCode
	privateUser.Depositor = input.Depositor
	privateUser.DepositType = input.DepositType
	privateUser.SignupStatus = domain.SIGNUP_ACCOUNT

	return nil
}

// setInvestmentSuitability 登録 - 本登録申請 - D. 投資適合性確認（SIGNUP_INVESTMENT_SUITABILITY）
func setInvestmentSuitability(input udomain.PrvUserInvestmentSuitability, privateUser *domain.PrivateUser) error {

	privateUser.Asset = input.Asset
	privateUser.AnnualIncome = input.AnnualIncome
	privateUser.IsProInvestAsset = input.IsProInvestAsset
	privateUser.IsProInvestFinanceAsset = input.IsProInvestFinanceAsset
	privateUser.InvestmentPurpose = input.InvestmentPurpose
	privateUser.IsSurplusFund = lo.ToPtr(input.IsSurplusFund)
	privateUser.HasExperience = lo.ToPtr(input.HasExperience)
	privateUser.SignupStatus = domain.SIGNUP_INVESTMENT_SUITABILITY

	return nil
}

func (p privateUserUsecase) RegisterMyNumber(ctx context.Context, request udomain.PrvUserRegisterMyNumberRequest) (udomain.PrvUserRegisterMyNumberResponse, error) {
	// TODO: マイナンバー登録復活させるまではエラーとします。
	return udomain.PrvUserRegisterMyNumberResponse{}, errors.Wrap(cdomain.ErrServiceUnavailable, "Cannot register my number.")

	// // get user by id
	// privateUser, err := p.privateUserRepository.GetById(ctx, request.JWTUser.Id)
	// if err != nil {
	// 	return udomain.PrvUserRegisterMyNumberResponse{}, errors.Wrap(err, "failed to get private user by id")
	// }

	// // 유효한 회원가입 단계인지 확인
	// signUpStatusSequence := signUpStatusMap[privateUser.SignupStatus]
	// if signUpStatusSequence == 0 {
	// 	// 현재 프로필에서 관리하는 상태가 아닌 경우: 현재보다 더 앞 단계에 있다고 가정하고 최대값으로 설정 (INT16:32767)
	// 	signUpStatusSequence = math.MaxInt16
	// }
	// if signUpStatusSequence+1 < signUpStatusMap[domain.SIGNUP_MYNUMBER] {
	// 	// 프로필 업데이트는 바로 이전단계/다음단계 요청만 유효
	// 	return udomain.PrvUserRegisterMyNumberResponse{}, errors.Wrap(cdomain.ErrItemConflict, "invalid sign up status")
	// }

	// // check my number exists
	// myNumber, err := p.myNumberService.GetByUserId(ctx, request.JWTUser.Id)
	// if err != nil {
	// 	return udomain.PrvUserRegisterMyNumberResponse{}, errors.Wrap(err, "failed to get my number")
	// }
	// if myNumber.IsEmpty() {
	// 	myNumber.Status = domain.UNREGISTERED
	// }

	// if myNumber.Status == domain.APPROVE || myNumber.Status == domain.WAIT_APPROVE {
	// 	// 이미 승인된 마이넘버가 있는 경우 또는 승인 대기중인 경우 에러 반환
	// 	return udomain.PrvUserRegisterMyNumberResponse{}, errors.Wrap(cdomain.ErrAlreadyExists, "my number already exists")
	// }

	// // register my number
	// registeredMyNumber, err := p.myNumberService.RegisterWithFiles(ctx,
	// 	cdomain.MyNumberRegisterInput{
	// 		UserId:        request.JWTUser.Id,
	// 		MyNumberValue: request.MyNumber,
	// 		FrontFileUrl:  request.FrontFileUrl,
	// 		BackFileUrl:   request.BackFileUrl,
	// 		MyNumber:      myNumber,
	// 	})
	// if err != nil {
	// 	return udomain.PrvUserRegisterMyNumberResponse{}, errors.Wrap(err, "failed to register my number")
	// }

	// // update signup_status
	// privateUser.User.SignupStatus = domain.SIGNUP_MYNUMBER
	// privateUser.User.AuditStatus = domain.CHECK_MY_NUMBER_AUDIT
	// if err := p.userRepository.Update(ctx, privateUser.User); err != nil {
	// 	return udomain.PrvUserRegisterMyNumberResponse{}, errors.Wrap(err, "failed to update user")
	// }

	// // return response
	// return udomain.PrvUserRegisterMyNumberResponse{
	// 	PrvUserMyNumber: udomain.PrvUserMyNumber{}.FromMyNumber(registeredMyNumber),
	// }, nil
}

// MIR-1358: 最新のeKYC履歴が本人確認NG(UNMATCH)と判定済みかどうかだけを判断する。
// 保留(HOLD_AUDIT)かどうかでは判断しない。保留からの正しいやり直し(失敗画面の
// 再試行等)でも会員は保留のまま完了PUTが届くため、保留かどうかで断ると
// やり直しまで壊してしまう。
func isEkycResultUnmatch(log domain.EkycRequestLog) bool {
	return log.Result == domain.EKYC_RESULT_STATUS_UNMATCH
}

// MIR-1358: 直近のeKYC履歴を次のeKYC開始(verifyWithEkyc/VerifyWithEkyc)で使い
// 回してよいかだけを判断する。判定(Result)が既に入っている行を使い回すと、
// isEkycResultUnmatchがそのやり直し自体を断ってしまうため、新しい行を作る
// 条件はこの否定になる。履歴が見つからない場合はゼロ値(Status="")になり falseを返す。
func canReuseEkycRequestLog(log domain.EkycRequestLog) bool {
	return log.Status == domain.EKYC_REQUEST_STATUS_NONE && log.Result == ""
}

func (p privateUserUsecase) VerifyWithEkyc(ctx context.Context, request udomain.PrvUserVerifyWithEkycRequest) (udomain.PrvUserVerifyWithEkycResponse, error) {
	if request.JWTUser.Id != request.UserId {
		return udomain.PrvUserVerifyWithEkycResponse{}, errors.Wrap(cdomain.ErrForbidden, "cannot verify other user's eKYC status")
	}

	var (
		ekycId   = p.idGenerator.GenKsuid().String()
		response udomain.PrvUserVerifyWithEkycResponse
	)

	// get private user
	privateUser, err := p.privateUserRepository.GetById(ctx, request.UserId)
	if err != nil {
		return response, errors.Wrap(err, "failed to get private user by id")
	}

	// 有効な登録段階かを確認
	signUpStatusSequence := signUpStatusMap[privateUser.SignupStatus]
	if signUpStatusSequence == 0 {
		// 現在のプロフィール内で管理されていない状態の場合：現在より前の段階にあると仮定し、最大値を設定（INT16:32767）
		signUpStatusSequence = math.MaxInt16
	}
	if signUpStatusSequence+1 < signUpStatusMap[domain.SIGNUP_KYC] {
		// プロフィール更新は直前/次のステップリクエストのみ有効
		return response, errors.Wrap(cdomain.ErrItemConflict, "invalid sign up status")
	}

	// eKYCリクエスト履歴
	ekycRequestLog, err := p.ekycRequestLogRepository.GetByUserId(ctx, request.UserId)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return response, errors.Wrap(err, "failed to get eKYC request log")
	}

	if canReuseEkycRequestLog(ekycRequestLog) {
		// 直近のeKYCリクエストが進行中(かつ未判定)だった場合は連携IDを再利用する
		ekycId = ekycRequestLog.EkycId
	}
	response.ReferenceId = ekycId

	if !canReuseEkycRequestLog(ekycRequestLog) {
		// 使い回せない(進行中でない、または既に判定済み)場合は新たにeKYCリクエストを記録する
		err := p.ekycRequestLogRepository.Create(ctx, domain.EkycRequestLog{
			Email:       privateUser.Email,
			UserId:      privateUser.User.Id,
			EkycId:      ekycId,
			ApplicantId: 0, // NOTE: 申請IDはWebhookの結果通知時に判明する
			Status:      domain.EKYC_REQUEST_STATUS_NONE,
		})
		if err != nil {
			return response, errors.Wrap(err, "failed to create ekyc request log")
		}
	}

	return response, nil
}

func (p privateUserUsecase) verifyWithEkyc(ctx context.Context, privateUser domain.PrivateUser) (string, error) {
	var (
		ekycId = p.idGenerator.GenKsuid().String()
	)

	// 有効な登録段階かを確認
	signUpStatusSequence := signUpStatusMap[privateUser.SignupStatus]
	if signUpStatusSequence == 0 {
		// 現在のプロフィール内で管理されていない状態の場合：現在より前の段階にあると仮定し、最大値を設定（INT16:32767）
		signUpStatusSequence = math.MaxInt16
	}
	if signUpStatusSequence+1 < signUpStatusMap[domain.SIGNUP_KYC] {
		// プロフィール更新は直前/次のステップリクエストのみ有効
		return "", errors.Wrap(cdomain.ErrItemConflict, "invalid sign up status")
	}

	// eKYCリクエスト履歴
	ekycRequestLog, err := p.ekycRequestLogRepository.GetByUserId(ctx, privateUser.User.Id)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return "", errors.Wrap(err, "failed to get eKYC request log")
	}

	if canReuseEkycRequestLog(ekycRequestLog) {
		// 直近のeKYCリクエストが進行中(かつ未判定)だった場合は連携IDを再利用する
		ekycId = ekycRequestLog.EkycId
	}

	if !canReuseEkycRequestLog(ekycRequestLog) {
		// 使い回せない(進行中でない、または既に判定済み)場合は新たにeKYCリクエストを記録する
		err := p.ekycRequestLogRepository.Create(ctx, domain.EkycRequestLog{
			Email:       privateUser.Email,
			UserId:      privateUser.User.Id,
			EkycId:      ekycId,
			ApplicantId: 0, // NOTE: 申請IDはWebhookの結果通知時に判明する
			Status:      domain.EKYC_REQUEST_STATUS_NONE,
		})
		if err != nil {
			return "", errors.Wrap(err, "failed to create ekyc request log")
		}
	}

	return ekycId, nil
}

func (p privateUserUsecase) UpdateEkycStatus(ctx context.Context, request udomain.PrvUserUpdateEkycStatusRequest) error {
	if request.JWTUser.Id != request.UserId {
		return errors.Wrap(cdomain.ErrForbidden, "cannot update other user's eKYC status")
	}

	ekycReq, err := p.ekycRequestLogRepository.GetByUserId(ctx, request.UserId)
	if err != nil {
		return errors.Wrap(err, "failed to get eKYC request log")
	}

	switch domain.EKYC_REQUEST_STATUS(strings.ToUpper(request.Status)) {
	case domain.EKYC_REQUEST_STATUS_SUCCESS:
		// MIR-1358: 最新の履歴が本人確認NG(UNMATCH)と判定済みなら、画面からの
		// 完了報告(成功画面の開き直し等)を断る。保留へ戻った会員が、古い完了
		// 報告で本登録承認待機へ戻ってしまう不具合の対策。会員も履歴も更新しない。
		if isEkycResultUnmatch(ekycReq) {
			p.logger.Entry(ctx).WithFields(logrus.Fields{
				"user_id":             ekycReq.UserId,
				"ekyc_request_log_id": ekycReq.Id,
			}).Info("rejected eKYC completion report because the latest eKYC request log is already judged as UNMATCH")
			return errors.Wrap(cdomain.ErrItemConflict, "latest eKYC request is already judged as UNMATCH")
		}

		ekycReq.Status = domain.EKYC_REQUEST_STATUS_SUCCESS

		// get user by id
		user, err := p.userRepository.GetById(ctx, ekycReq.UserId)
		if err != nil {
			return errors.Wrap(err, "failed to get user by id")
		}
		// signup_statusがSIGNUP_COMPLETEの時はすでにeKYC提出完了状態のため変更しない。
		if user.SignupStatus == domain.SIGNUP_COMPLETE {
			return errors.Wrap(cdomain.ErrItemConflict, fmt.Sprintf("user's signup status is already %s", domain.SIGNUP_COMPLETE))
		}
		// audit_statusがCOMPLETED_AUDITの時はすでに本登録完了しているため変更しない。
		if user.AuditStatus == domain.COMPLETED_AUDIT {
			return errors.Wrap(cdomain.ErrItemConflict, "user's audit status is already completed")
		}

		// update user
		user.SignupStatus = domain.SIGNUP_COMPLETE
		user.AuditStatus = domain.WAIT_AUDIT
		if err = p.userRepository.Update(ctx, user); err != nil {
			return errors.Wrap(err, "failed to update user")
		}

	case domain.EKYC_REQUEST_STATUS_FAIL:
		ekycReq.Status = domain.EKYC_REQUEST_STATUS_FAIL
	default:
		return errors.Wrap(cdomain.ErrBadRequest, "invalid eKYC status")
	}

	if err := p.ekycRequestLogRepository.Update(ctx, ekycReq); err != nil {
		return errors.Wrap(err, "failed to update eKYC request log")
	}

	return nil
}

func (p privateUserUsecase) GetEkycStatus(ctx context.Context, request udomain.PrvUserGetEkycStatusRequest) (udomain.PrvUserGetEkycStatusResponse, error) {

	ekycReq, err := p.ekycRequestLogRepository.GetByUserId(ctx, request.UserId)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return udomain.PrvUserGetEkycStatusResponse{Status: nil}, nil
		}
		return udomain.PrvUserGetEkycStatusResponse{}, errors.Wrap(err, "failed to get eKYC request log")
	}

	if ekycReq.Status == domain.EKYC_REQUEST_STATUS_SUCCESS {
		return udomain.PrvUserGetEkycStatusResponse{Status: lo.ToPtr(true)}, nil
	} else if ekycReq.Status == domain.EKYC_REQUEST_STATUS_FAIL {
		return udomain.PrvUserGetEkycStatusResponse{Status: lo.ToPtr(false)}, nil
	} else {
		return udomain.PrvUserGetEkycStatusResponse{Status: nil}, nil
	}
}

func (p privateUserUsecase) CallEkycWebhook(ctx context.Context, request udomain.PrvUserCallEkycWebhookRequest) error {

	// DEV、QA環境でのみ使用可能
	if strings.ToUpper(p.cfg.CommonConfig.Env) != "DEV" && strings.ToUpper(p.cfg.CommonConfig.Env) != "QA" {
		return errors.Wrap(cdomain.ErrServiceUnavailable, "only available in DEV, QA environment")
	}

	// get eKYC request log
	ekycReq, err := p.ekycRequestLogRepository.GetByUserId(ctx, request.UserId)
	if err != nil {
		return errors.Wrap(err, "failed to get eKYC request log")
	}

	// get user
	user, err := p.userRepository.GetById(ctx, request.UserId)
	if err != nil {
		return errors.Wrap(err, "failed to get user by email")
	}

	// get my number
	myNumber, err := p.myNumberRepository.GetByUserId(ctx, request.UserId)
	if err != nil {
		return errors.Wrap(err, "failed to get my number by user id")
	}

	switch domain.EKYC_REQUEST_STATUS(strings.ToUpper(request.Status)) {
	case domain.EKYC_REQUEST_STATUS_SUCCESS:
		// eKYC結果ステータス変更
		ekycReq.Result = domain.EKYC_RESULT_STATUS_MATCH

		// 事業型仮想口座を接続
		err = p.userVirtualAccountService.LinkVirtualAccount(ctx, user.Id, domain.EQUITY)
		if err != nil {
			return errors.Wrap(err, "failed to link virtual account")
		}

		// ステータス変更
		user.IsEkyc = true
		if myNumber.Status == domain.APPROVE {
			user.AuditStatus = domain.COMPLETED_AUDIT
			user.SignUpApprovalDate = jsonutils.JSONTime(time.Now())

			// 本登録承認完了メール送信
			if err := p.mailService.SendMail(ctx, cdomain.SendMailRequest{
				TemplateType: domain.COMPLETED_AUDIT_STATUS_FOR_PRIVATE_TEMPLATE,
				Sender:       p.cfg.ServiceConfig.Mail.Info,
				Receivers:    []string{user.Email},
				MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
				Details: map[string]interface{}{
					"Receiver":   user.Name,
					"ServiceUrl": p.cfg.ServiceConfig.Url.UserServiceDomain,
					"ProductUrl": p.cfg.ServiceConfig.Url.UserServiceDomain + p.cfg.ServiceConfig.Url.Product,
				},
			}); err != nil {
				return errors.Wrap(err, "failed to send mail")
			}

		} else {
			// TODO: マイナンバー登録時に復活させます
			// NOTE: このエンドポイントは現在は使われていなさそう
			// user.AuditStatus = domain.CHECK_MY_NUMBER_AUDIT
		}

	case domain.EKYC_REQUEST_STATUS_FAIL:
		// eKYC結果ステータス変更
		ekycReq.Result = domain.EKYC_RESULT_STATUS_UNMATCH
		ekycReq.ErrorMessage = lo.ToPtr("顔照合結果がアンマッチ(同一人物でない)\n指定した確認書類の画像でない\n免許証番号の不整合を検出しました。\n公安委員会と住所の不整合を検出しました。\n")

		// HOLD_AUDITステータスに変更
		user.AuditStatus = domain.HOLD_AUDIT
		holdUserCheckCode := p.idGenerator.GenUuidV4().String()
		user.HoldUserCheckCode = &holdUserCheckCode
	default:
		return errors.Wrap(cdomain.ErrBadRequest, "invalid eKYC status")
	}

	// update
	if err := p.ekycRequestLogRepository.Update(ctx, ekycReq); err != nil {
		return errors.Wrap(err, "failed to update eKYC request log")
	}
	if err := p.userRepository.Update(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	// TODO: user info history更新追加

	return nil
}

func (p privateUserUsecase) GetOne(ctx context.Context, request udomain.PrvUserGetOneRequest) (udomain.PrvUserGetOneResponse, error) {

	privateUser, err := p.privateUserRepository.GetById(ctx, request.JWTUser.Id)
	if err != nil {
		return udomain.PrvUserGetOneResponse{}, errors.Wrap(err, "failed to get private user by id")
	}

	userVirtualAccount, err := p.userVirtualAccountRepository.GetByUserIdAndLinkedStatus(ctx, request.JWTUser.Id, domain.EQUITY, domain.ON_VA_ACCOUNT)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return udomain.PrvUserGetOneResponse{}, errors.Wrap(err, "failed to get virtual account")
	}

	response := udomain.PrvUserGetOneResponse{
		UserId:       privateUser.User.Id,
		Email:        privateUser.User.Email,
		IsNewsletter: privateUser.User.IsNewsletter,
		Created:      privateUser.User.Created,
		Name:         privateUser.PrivateProfile.FamilyName + " " + privateUser.PrivateProfile.FirstName,
		NameKana:     privateUser.PrivateProfile.FamilyNameKana + " " + privateUser.PrivateProfile.FirstNameKana,
		Nationality:  privateUser.PrivateProfile.Nationality,
		Birthday:     privateUser.PrivateProfile.Birthday.ToFormat(time.DateOnly),
		ZipCode:      privateUser.User.ZipCode,
		Address:      privateUser.User.Address,
		Address2:     privateUser.User.Address2,
		Province:     privateUser.User.Province,
		City:         privateUser.User.City,
		Job:          privateUser.PrivateProfile.Job,
		MobilePhone:  privateUser.User.MobilePhone,
		AuditStatus:  privateUser.AuditStatus,

		InvestmentSuitability: udomain.PrvUserGetOneInvestmentSuitability{
			Asset:                   privateUser.PrivateProfile.Asset,
			AnnualIncome:            privateUser.PrivateProfile.AnnualIncome,
			IsProInvestAsset:        privateUser.PrivateProfile.IsProInvestAsset,
			IsProInvestFinanceAsset: privateUser.PrivateProfile.IsProInvestFinanceAsset,
			InvestmentPurpose:       privateUser.User.InvestmentPurpose,
			HasExperience:           privateUser.User.HasExperience,
			IsSurplusFund:           privateUser.User.IsSurplusFund,
		},

		WithdrawalAccount: udomain.PrvUserGetOneWithdrawalAccount{
			BankName:    privateUser.User.BankName,
			BranchName:  privateUser.User.BranchName,
			BankAccount: privateUser.User.BankAccount,
			BankCode:    lo.FromPtr(privateUser.User.BankCode),
			BranchCode:  lo.FromPtr(privateUser.User.BranchCode),
			Depositor:   privateUser.User.Depositor,
			DepositType: privateUser.User.DepositType,
		},

		VirtualAccount: udomain.PrvUserGetOneVirtualAccount{
			BankName:       userVirtualAccount.VaBankName,
			BranchName:     userVirtualAccount.VaBranchName,
			AccountNumber:  userVirtualAccount.VaAccountNumber,
			HolderNameKana: userVirtualAccount.VaHolderNameKana,
		},
	}

	return response, nil
}

func (p privateUserUsecase) PatchOne(ctx context.Context, request udomain.PrvUserPatchOneRequest) error {
	user, err := p.userRepository.GetById(ctx, request.JWTUser.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	if user.UserType != domain.PRIVATE {
		return errors.Wrap(cdomain.ErrItemConflict, fmt.Sprintf("invalid user type:%s", user.UserType))
	}

	// ニュースレター受信同意の変更
	user.IsNewsletter = lo.FromPtrOr(request.IsNewsletter, user.IsNewsletter)

	// TODO 今後追加されるフィールドはここに追加します。

	if err = p.userRepository.Update(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	return nil
}
