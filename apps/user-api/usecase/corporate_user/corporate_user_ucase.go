package corporate_user

import (
	"context"
	"fmt"
	"math"
	"path"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type corporateUserUsecase struct {
	cfg                            *config.Config
	s3pkg                          s3utils.S3
	fileHelper                     filehelper.FileHelper
	logger                         factory.LoggerFactory
	userRepository                 cdomain.UserRepository
	userVirtualAccountRepository   cdomain.UserVirtualAccountRepository
	corporateProfileRepository     cdomain.CorporateProfileRepository
	corporateShareholderRepository cdomain.CorporateShareholderRepository
	userFileRepository             cdomain.UserFileRepository
	mailService                    cdomain.MailService
	userFileService                cdomain.UserFileService
}

func New(
	cfg *config.Config,
	s3pkg s3utils.S3,
	fileHelper filehelper.FileHelper,
	logger factory.LoggerFactory,
	userRepository cdomain.UserRepository,
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository,
	corporateProfileRepository cdomain.CorporateProfileRepository,
	corporateShareholderRepository cdomain.CorporateShareholderRepository,
	userFileRepository cdomain.UserFileRepository,
	mailService cdomain.MailService,
	userFileService cdomain.UserFileService,
) udomain.CorporateUserUsecase {
	return &corporateUserUsecase{
		cfg:                            cfg,
		s3pkg:                          s3pkg,
		fileHelper:                     fileHelper,
		logger:                         logger,
		userRepository:                 userRepository,
		userVirtualAccountRepository:   userVirtualAccountRepository,
		corporateProfileRepository:     corporateProfileRepository,
		corporateShareholderRepository: corporateShareholderRepository,
		userFileRepository:             userFileRepository,
		mailService:                    mailService,
		userFileService:                userFileService,
	}
}

var signUpStatusMap = map[domain.SIGNUP_STATUS]int{
	domain.SIGNUP_NONE:                             1,
	domain.SIGNUP_CORPORATE_PROFILE_BASIC:          2,
	domain.SIGNUP_CORPORATE_PROFILE_OTHER:          3,
	domain.SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE: 4,
	domain.SIGNUP_CORPORATE_PROFILE_MANAGER:        5,
	domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER:    6,
	domain.SIGNUP_ACCOUNT:                          7,
	domain.SIGNUP_INVESTMENT_SUITABILITY:           8,
	domain.SIGNUP_AGREE:                            9,
	domain.SIGNUP_KYC:                              10,
	domain.SIGNUP_COMPLETE:                         11,
}

// isAuditStatusReentryBlocked は、審査が確定または反社確認等で進行中のため、登録の完了有無にかかわらず再入力を
// 許さない審査状態かを返す(個人側と同じ4つ)。
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
	1:  domain.SIGNUP_NONE,
	2:  domain.SIGNUP_CORPORATE_PROFILE_BASIC,
	3:  domain.SIGNUP_CORPORATE_PROFILE_OTHER,
	4:  domain.SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE,
	5:  domain.SIGNUP_CORPORATE_PROFILE_MANAGER,
	6:  domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER,
	7:  domain.SIGNUP_ACCOUNT,
	8:  domain.SIGNUP_INVESTMENT_SUITABILITY,
	9:  domain.SIGNUP_AGREE,
	10: domain.SIGNUP_KYC,
	11: domain.SIGNUP_COMPLETE,
}

func (c corporateUserUsecase) GetSignUpStatus(ctx context.Context, request udomain.CorporateUserGetSignUpStatusRequest) (udomain.CorporateUserGetSignUpStatusResponse, error) {
	user, err := c.userRepository.GetById(ctx, request.User.Id)
	if err != nil {
		return udomain.CorporateUserGetSignUpStatusResponse{}, errors.Wrap(err, "failed to get user")
	}

	corporateProfile, err := c.corporateProfileRepository.GetByUserId(ctx, request.User.Id)
	if err != nil {
		return udomain.CorporateUserGetSignUpStatusResponse{}, errors.Wrap(err, "failed to get corporate profile")
	}

	// 만약 요청이 없을 경우 현재 조회된 유저의 현재상태를 반환합니다.
	if lo.IsEmpty(request.Status) {
		request.Status = signUpStatusRMap[signUpStatusMap[user.SignupStatus]]
	}

	signUpSequence := signUpStatusMap[request.Status]

	var response = udomain.CorporateUserGetSignUpStatusResponse{
		UserId:     user.Id,
		Email:      user.Email,
		PrevStatus: signUpStatusRMap[signUpSequence-1],
		NextStatus: signUpStatusRMap[signUpSequence+1],
		CurStatus:  request.Status,
	}

	if response.CurStatus == domain.SIGNUP_NONE {
		response.PrevStatus = domain.SIGNUP_NONE
	}
	if response.CurStatus == domain.SIGNUP_COMPLETE {
		response.PrevStatus = domain.SIGNUP_AGREE
		response.NextStatus = domain.SIGNUP_COMPLETE
	}

	switch request.Status {
	case domain.SIGNUP_NONE, domain.SIGNUP_AGREE:
		return response, nil
	case domain.SIGNUP_CORPORATE_PROFILE_BASIC:
		response.BasicProfile = &udomain.CorporateUserBasicProfileItem{
			CorpNo:   corporateProfile.CorpNo,
			CorpName: user.Name,
			CorpType: corporateProfile.CorpType,
		}
	case domain.SIGNUP_CORPORATE_PROFILE_OTHER:
		response.OtherProfile = &udomain.CorporateUserOtherProfileItem{
			ZipCode:     user.ZipCode,
			Province:    user.Province,
			City:        user.City,
			Address:     user.Address,
			Address2:    user.Address2,
			MobilePhone: user.MobilePhone,
		}
	case domain.SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE:
		response.RepresentativeProfile = &udomain.CorporateUserRepresentativeProfileItem{
			FamilyName:     corporateProfile.RepresentativeFamilyName,
			FirstName:      corporateProfile.RepresentativeFirstName,
			FamilyNameKana: corporateProfile.RepresentativeFamilyNameKana,
			FirstNameKana:  corporateProfile.RepresentativeFirstNameKana,
			Email:          corporateProfile.RepresentativeEmail,
			MobilePhone:    corporateProfile.RepresentativeMobilePhone,
		}
	case domain.SIGNUP_CORPORATE_PROFILE_MANAGER:
		birthday := ""

		// 생년월일이 존재하지 않을 경우 빈 문자열로 설정합니다. (현재 따로 설정하지 않으면 0001-01-01 값으로 설정됩니다.)
		if !corporateProfile.ManagerBirthday.EqualDate(time.Time{}) {
			birthday = corporateProfile.ManagerBirthday.ToFormat(time.DateOnly)
		}

		response.ManagerProfile = &udomain.CorporateUserManagerProfileItem{
			FamilyName:     lo.FromPtr(corporateProfile.ManagerFamilyName),
			FirstName:      lo.FromPtr(corporateProfile.ManagerFirstName),
			FamilyNameKana: lo.FromPtr(corporateProfile.ManagerFamilyNameKana),
			FirstNameKana:  lo.FromPtr(corporateProfile.ManagerFirstNameKana),
			Email:          lo.FromPtr(corporateProfile.ManagerEmail),
			MobilePhone:    lo.FromPtr(corporateProfile.ManagerMobilePhone),
			Department:     lo.FromPtr(corporateProfile.Department),
			Position:       lo.FromPtr(corporateProfile.Position),
			ZipCode:        corporateProfile.ManagerZipCode,
			Province:       corporateProfile.ManagerProvince,
			City:           corporateProfile.ManagerCity,
			Address:        corporateProfile.ManagerAddress,
			Address2:       corporateProfile.ManagerAddress2,
			Birthday:       birthday,
		}
	case domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER:
		shareholderProfiles := make([]udomain.CorporateUserShareholderProfileItem, 0)

		_, corporateShareholders, err := c.corporateShareholderRepository.GetAll(ctx,
			cdomain.CorporateShareholderQueryOptionFunc().WithUserId(user.Id))
		if err != nil {
			return udomain.CorporateUserGetSignUpStatusResponse{}, errors.Wrap(err, "failed to get corporate shareholders")
		}

		for _, shareholder := range corporateShareholders {
			birthday := ""

			// 생년월일이 존재하지 않을 경우 빈 문자열로 설정합니다. (현재 따로 설정하지 않으면 0001-01-01 값으로 설정됩니다.)
			if !shareholder.Birthday.EqualDate(time.Time{}) {
				birthday = shareholder.Birthday.ToFormat(time.DateOnly)
			}

			shareholderProfiles = append(shareholderProfiles, udomain.CorporateUserShareholderProfileItem{
				FamilyName:     shareholder.FamilyName,
				FirstName:      shareholder.FirstName,
				FamilyNameKana: shareholder.FamilyNameKana,
				FirstNameKana:  shareholder.FirstNameKana,
				ZipCode:        shareholder.ZipCode,
				Province:       shareholder.Province,
				City:           shareholder.City,
				Address:        shareholder.Address,
				Address2:       &shareholder.Address2,
				Birthday:       birthday,
				Uid:            shareholder.Uid,
			})
		}
		response.ShareholderProfiles = shareholderProfiles
	case domain.SIGNUP_ACCOUNT:
		response.WithdrawalAccount = &udomain.CorporateUserRegisterWithdrawalAccountItem{
			BankName:    user.BankName,
			BranchName:  user.BranchName,
			BankCode:    user.BankCode,
			BranchCode:  user.BranchCode,
			DepositType: user.DepositType,
			BankAccount: user.BankAccount,
			Depositor:   user.Depositor,
		}
	case domain.SIGNUP_INVESTMENT_SUITABILITY:
		response.InvestmentSuitability = &udomain.CorporateUserConfirmInvestmentSuitabilityItem{
			EmployeeCount:     corporateProfile.EmployeeCount,
			CapitalAmount:     corporateProfile.CapitalAmount,
			Sales:             corporateProfile.Sales,
			InvestmentPurpose: user.InvestmentPurpose,
			HasExperience:     user.HasExperience,
			IsSurplusFund:     user.IsSurplusFund,
			InvestorType:      user.InvestorType,
		}
	case domain.SIGNUP_KYC, domain.SIGNUP_COMPLETE: // 실제로 SIGNUP_KYC 상태가 되는 경우는 없다.
		kycFiles, err := c.userFileService.GetCorporateUserIdentificationFiles(ctx, user.Id)
		if err != nil {
			return udomain.CorporateUserGetSignUpStatusResponse{}, errors.Wrap(err, "failed to get corporate user identification files")
		}
		response.VerifyWithKyc = &udomain.CorporateUserVerifyWithKycItem{
			Identification1FrontFile: kycFiles.Front1,
			Identification1BackFile:  kycFiles.Back1,
			Identification2FrontFile: kycFiles.Front2,
			Identification2BackFile:  kycFiles.Back2,
			CertifiedCopyFile:        kycFiles.CertifiedCopy,
			PoaFile:                  kycFiles.POA,
		}
	default:
		return udomain.CorporateUserGetSignUpStatusResponse{}, errors.Wrap(cdomain.ErrBadRequest, "invalid sign up status")
	}

	return response, nil
}

func (c corporateUserUsecase) RegisterProfiles(ctx context.Context, request udomain.CorporateUserRegisterProfilesRequest) error {
	// 업데이트를 위한 현재 상태조회
	user, err := c.userRepository.GetById(ctx, request.User.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	// MIR-1421: 登録が完了した法人は、登録用のルートで登録内容(法人名・法人番号・代表者・担当者・出金口座の名義など)を
	// 書き換えられない。このルートは状態を戻さず審査にも回さないため、完了後に許すと審査を通さずに書き換えられた。
	// MIR-940では状態が巻き戻らないことを理由に法人側のガードを取り消した(2e79aab0)が、書き換えの穴として塞ぐ。
	// 個人側と違い、保留・仮登録へ差し戻されていても弾く。法人の画面は登録完了の法人を本人確認ページへ入れず、
	// 差し戻しからの再入力に使っていないため(保留からの再提出用には書類の差し替えのAPIがある)。審査が確定・進行中の
	// 状態は完了前でも弾く(エラーで審査状態を開示しない)。
	if user.SignupStatus == domain.SIGNUP_COMPLETE {
		return errors.Wrap(cdomain.ErrItemConflict, fmt.Sprintf("user's signup status is already %s", domain.SIGNUP_COMPLETE))
	}
	if user.AuditStatus == domain.COMPLETED_AUDIT {
		return errors.Wrap(cdomain.ErrItemConflict, "user's audit status is already completed")
	}
	if isAuditStatusReentryBlocked(user.AuditStatus) {
		return errors.Wrap(cdomain.ErrItemConflict, "user's audit status does not allow profile reentry")
	}

	signUpStatusSequence := signUpStatusMap[user.SignupStatus]

	// 현재 프로필에서 관리하는 상태가 아닌 경우에는 현재보다 더 앞 단계에 있다고 가정을하고 최대값으로 설정합니다. (INT16:32767)
	if signUpStatusSequence == 0 {
		signUpStatusSequence = math.MaxInt16
	}

	// 프로필 업데이트는 이전단계 혹은 바로 다음 단계요청만 유효하게 요청할 수 있습니다.
	// 만약 더 앞 단계를 요청하게 될 경우 에러를 반환합니다.
	// ex1) 현재 상태가 BASIC인 경우 OTHER 요청은 가능하지만 MANAGER 요청은 불가능합니다.
	// ex2) 현재 상태가 OTHER인 경우 BASIC, REPRESENTATIVE 요청은 가능하지만 MANAGER 요청은 불가능합니다.
	if signUpStatusSequence+1 < signUpStatusMap[request.SignUpStatus] {
		return errors.Wrap(cdomain.ErrItemConflict, "invalid sign up status")
	}

	corporateProfile, err := c.corporateProfileRepository.GetByUserId(ctx, request.User.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get corporate profile")
	}

	switch request.SignUpStatus {
	case domain.SIGNUP_CORPORATE_PROFILE_BASIC:
		// 업데이트
		// 법인번호  : corporate_profile.corp_no (13자리 숫자만 허용, 이미 있는 법인 번호일 경우 허용하지 않음)
		// 법인명 : user.corp_name (비어있으면 안됨)
		// 사업내용 : corporate_profile.corp_type (비어있으면 안됨)

		// 요청 받은 법인번호가 기존에 저장된 법인번호와 다를 때 이미 존재하는 법인번호인지 확인합니다.
		if corporateProfile.CorpNo != request.BasicProfile.CorpNo {
			_, err = c.corporateProfileRepository.GetByCorpNo(ctx, request.BasicProfile.CorpNo)

			// 에러가 존재하지 않는다는 것은 이미 존재하는 법인번호가 있음을 의미합니다.
			if err == nil {
				return errors.Wrap(cdomain.ErrAlreadyExists, "already existing corporate number")
			}
		}

		corporateProfile.CorpNo = request.BasicProfile.CorpNo
		corporateProfile.CorpType = request.BasicProfile.CorpType

		user.Name = request.BasicProfile.CorpName
	case domain.SIGNUP_CORPORATE_PROFILE_OTHER:
		// 업데이트
		// 우편번호 : user.zip_code (7자리 숫자만 허용, 유효한 zip_code는 어떻게 판단하는지 확인필요)
		// 도도부현 : user.province (비어있으면 안됨, 우편번호에 따른 자동입력)
		// 시구정촌 : user.city (비어있으면 안됨, 우편번호에 따른 자동입력)
		// 주소 : user.address (비어있으면 안됨)
		// 상세주소 : user.address2 (비어있어도 됨)
		// 전화번호 : user.mobile_phone (비어있으면 안됨, 숫자만 허용)

		user.ZipCode = request.OtherProfile.ZipCode
		user.Province = request.OtherProfile.Province
		user.City = request.OtherProfile.City
		user.Address = request.OtherProfile.Address
		user.Address2 = request.OtherProfile.Address2
		user.MobilePhone = request.OtherProfile.MobilePhone

	case domain.SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE:
		// 업데이트
		// 대표임원 성 : corporate_profile.representative_family_name (비어있으면 안됨)
		// 대표임원 이름 : corporate_profile.representative_first_name (비어있으면 안됨)
		// 대표임원 후리가나 성 : corporate_profile.representative_family_name_kana (비어있으면 안됨)
		// 대표임원 후리가나 이름 : corporate_profile.representative_first_name_kana (비어있으면 안됨)
		// 대표임원 이메일주소 : corporate_profile.representative_email (비어있으면 안됨, 이메일 형식)
		// 대표임원 전화번호 : corporate_profile.representative_mobile_phone (비어있으면 안됨, 숫자이외의 문자 허용하지 않음)

		corporateProfile.RepresentativeFamilyName = request.RepresentativeProfile.FamilyName
		corporateProfile.RepresentativeFirstName = request.RepresentativeProfile.FirstName
		corporateProfile.RepresentativeFamilyNameKana = request.RepresentativeProfile.FamilyNameKana
		corporateProfile.RepresentativeFirstNameKana = request.RepresentativeProfile.FirstNameKana
		corporateProfile.RepresentativeEmail = request.RepresentativeProfile.Email
		corporateProfile.RepresentativeMobilePhone = request.RepresentativeProfile.MobilePhone

	case domain.SIGNUP_CORPORATE_PROFILE_MANAGER:
		// 업데이트
		// 법인 담당자 성 : corporate_profile.manager_family_name (비어있으면 안됨)
		// 법인 담당자 이름 : corporate_profile.manager_first_name (비어있으면 안됨)
		// 법인 담당자 후리가나 성 : corporate_profile.manager_family_name_kana (비어있으면 안됨)
		// 법인 담당자 후리가나 이름 : corporate_profile.manager_first_name_kana (비어있으면 안됨)
		// 법인 담당자 이메일주소 : corporate_profile.manager_email (비어있으면 안됨, 이메일 형식)
		// 법인 담당자 전화번호 : corporate_profile.manager_mobile_phone (비어있으면 안됨, 숫자이외의 문자 허용하지 않음)
		// 법인 담당자 생년월일 : corporate_profile.manager_birthday (비어있으면 안됨, 날짜 형식)
		// 법인 담당자 부서명 : corporate_profile.department (비어있으면 안됨)
		// 법인 담당자 직책 : corporate_profile.position (비어있으면 안됨)
		// 법인 담당자 우편번호 : corporate_profile.manager_zip_code (7자리 숫자만 허용, 유효한 zip_code는 어떻게 판단하는지 확인필요)
		// 법인 담당자 도도부현 : corporate_profile.manager_province (비어있으면 안됨, 우편번호에 따른 자동입력)
		// 법인 담당자 시구정촌 : corporate_profile.manager_city (비어있으면 안됨, 우편번호에 따른 자동입력)
		// 법인 담당자 주소 : corporate_profile.manager_address (비어있으면 안됨)
		// 법인 담당자 상세주소 : corporate_profile.manager_address2 (비어있어도 됨)

		corporateProfile.ManagerFamilyName = lo.ToPtr(request.ManagerProfile.FamilyName)
		corporateProfile.ManagerFirstName = lo.ToPtr(request.ManagerProfile.FirstName)
		corporateProfile.ManagerFamilyNameKana = lo.ToPtr(request.ManagerProfile.FamilyNameKana)
		corporateProfile.ManagerFirstNameKana = lo.ToPtr(request.ManagerProfile.FirstNameKana)
		corporateProfile.ManagerEmail = lo.ToPtr(request.ManagerProfile.Email)
		corporateProfile.ManagerMobilePhone = lo.ToPtr(request.ManagerProfile.MobilePhone)
		corporateProfile.Department = lo.ToPtr(request.ManagerProfile.Department)
		corporateProfile.Position = lo.ToPtr(request.ManagerProfile.Position)
		corporateProfile.ManagerZipCode = request.ManagerProfile.ZipCode
		corporateProfile.ManagerProvince = request.ManagerProfile.Province
		corporateProfile.ManagerCity = request.ManagerProfile.City
		corporateProfile.ManagerAddress = request.ManagerProfile.Address
		corporateProfile.ManagerAddress2 = request.ManagerProfile.Address2

		birthday, err := jsonutils.ParseJSONTimeJST(time.DateOnly, request.ManagerProfile.Birthday)
		if err != nil {
			return errors.Wrap(err, "failed to parse birthday")
		}

		corporateProfile.ManagerBirthday = birthday

	case domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER:
		// 기존 유저의 ShareHolder 삭제 (hard delete)
		err = c.corporateShareholderRepository.DeleteByUserId(ctx, user.Id)
		if err != nil && !errors.Is(err, cdomain.ErrItemNotAffected) {
			return errors.Wrap(err, "failed to delete corporate shareholder")
		}

		for _, shareholder := range request.ShareholderProfiles {
			// 법인 주주 성 : corporate_shareholder.family_name (비어있으면 안됨)
			// 법인 주주 이름 : corporate_shareholder.first_name (비어있으면 안됨)
			// 법인 주주 후리가나 성 : corporate_shareholder.family_name_kana (비어있으면 안됨)
			// 법인 주주 후리가나 이름 : corporate_shareholder.first_name_kana (비어있으면 안됨)
			// 법인 주주 우편번호 : corporate_shareholder.zip_code (비어있으면 안됨, 7자리 숫자만 허용)
			// 법인 주주 도도부현 : corporate_shareholder.province (비어있으면 안됨, 우편번호에 따른 자동입력)
			// 법인 주주 시구정촌 : corporate_shareholder.city (비어있으면 안됨, 우편번호에 따른 자동입력)
			// 법인 주주 주소 : corporate_shareholder.address (비어있으면 안됨)
			// 법인 주주 상세주소 : corporate_shareholder.address2 (비어있어도 됨)
			birthday, err := jsonutils.ParseJSONTimeJST(time.DateOnly, shareholder.Birthday)
			if err != nil {
				return errors.Wrap(err, "failed to parse birthday")
			}

			if _, err = c.corporateShareholderRepository.Create(ctx, domain.CorporateShareholder{
				UserId:         user.Id,
				FamilyName:     shareholder.FamilyName,
				FirstName:      shareholder.FirstName,
				FamilyNameKana: shareholder.FamilyNameKana,
				FirstNameKana:  shareholder.FirstNameKana,
				ZipCode:        shareholder.ZipCode,
				Province:       shareholder.Province,
				City:           shareholder.City,
				Address:        shareholder.Address,
				Address2:       lo.FromPtr(shareholder.Address2),
				Birthday:       birthday,
				Uid:            shareholder.Uid,
			}); err != nil {
				return errors.Wrap(err, "failed to create corporate shareholder")
			}

		}
	case domain.SIGNUP_ACCOUNT:
		user.BankName = request.WithdrawalAccount.BankName
		user.BranchName = request.WithdrawalAccount.BranchName
		user.BankCode = request.WithdrawalAccount.BankCode
		user.BranchCode = request.WithdrawalAccount.BranchCode
		user.DepositType = request.WithdrawalAccount.DepositType
		user.BankAccount = request.WithdrawalAccount.BankAccount
		user.Depositor = request.WithdrawalAccount.Depositor

	case domain.SIGNUP_INVESTMENT_SUITABILITY:
		// update corporate_profile
		corporateProfile.EmployeeCount = request.InvestmentSuitability.EmployeeCount
		corporateProfile.CapitalAmount = request.InvestmentSuitability.CapitalAmount
		corporateProfile.Sales = request.InvestmentSuitability.Sales

		// update user
		user.InvestmentPurpose = request.InvestmentSuitability.InvestmentPurpose
		user.HasExperience = request.InvestmentSuitability.HasExperience
		user.IsSurplusFund = request.InvestmentSuitability.IsSurplusFund

	case domain.SIGNUP_AGREE:
		// do nothing
	default:
		return errors.Wrap(cdomain.ErrBadRequest, "invalid sign up status")
	}

	if signUpStatusMap[user.SignupStatus] < signUpStatusMap[request.SignUpStatus] {
		user.SignupStatus = request.SignUpStatus
	}
	if err = c.userRepository.Update(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	if err = c.corporateProfileRepository.Update(ctx, corporateProfile); err != nil {
		return errors.Wrap(err, "failed to update corporate profile")
	}

	return nil
}

func (c corporateUserUsecase) VerifyWithKyc(ctx context.Context, request udomain.CorporateUserVerifyWithKycRequest) error {

	// update user, corporate_profile
	user, err := c.userRepository.GetById(ctx, request.JWTUser.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get user by email")
	}

	signUpStatusSequence := signUpStatusMap[user.SignupStatus]

	// 현재 프로필에서 관리하는 상태가 아닌 경우에는 현재보다 더 앞 단계에 있다고 가정을하고 최대값으로 설정합니다. (INT16:32767)
	if signUpStatusSequence == 0 {
		signUpStatusSequence = math.MaxInt16
	}

	// 프로필 업데이트는 이전단계 혹은 바로 다음 단계요청만 유효하게 요청할 수 있습니다.
	// 만약 더 앞 단계를 요청하게 될 경우 에러를 반환합니다.
	// ex1) 현재 상태가 BASIC인 경우 OTHER 요청은 가능하지만 MANAGER 요청은 불가능합니다.
	// ex2) 현재 상태가 OTHER인 경우 BASIC, REPRESENTATIVE 요청은 가능하지만 MANAGER 요청은 불가능합니다.
	// 또한, SIGNUP_KYC 상태는 SignupStatus로 설정되는 경우는 없지만, 개념적인 단계이다.
	if signUpStatusSequence+1 < signUpStatusMap[domain.SIGNUP_KYC] {
		return errors.Wrap(cdomain.ErrItemConflict, "invalid sign up status")
	}

	// 파일 등록을 위한 정보 셋팅
	userFiles, _, err := c.buildUserFiles(ctx, request.JWTUser.Id, request.CorporateUserFileUrlItem)
	if err != nil {
		return errors.Wrap(err, "failed to init user file")
	}

	// 二重送信などで全て同名だった場合、追加対象が無くなる。
	// xorm は空スライスの Insert で ErrNoElementsOnSlice を返すため、先に弾く。
	if len(userFiles) == 0 {
		return errors.Wrap(cdomain.ErrBadRequest, "no document to register")
	}

	// update audit status
	user.SignupStatus = domain.SIGNUP_COMPLETE
	user.AuditStatus = domain.WAIT_AUDIT
	if err = c.userRepository.Update(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	// create file info
	if err = c.userFileRepository.CreateMany(ctx, userFiles); err != nil {
		return errors.Wrap(err, "failed to update corporate profile")
	}

	return nil
}

func (c corporateUserUsecase) VerifyPostalCode(ctx context.Context, request udomain.CorporateUserVerifyPostalCodeRequest) error {
	user, err := c.userRepository.GetById(ctx, request.User.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	// 郵便コード認証は住所確認中のときだけ許す。
	// コードは DB に残り続けるため、保留・無効・承認済みから同じコードを再送されると
	// audit_status が承認完了へ戻り、承認日の更新と承認メールの再送まで起きる。
	if user.AuditStatus != domain.CHECK_ADDRESS_AUDIT {
		// ErrBadRequest は delivery でコード不一致(20015)に変換されるため、
		// 状態エラーが「コードが違う」と見えてしまう。状態の衝突として返す。
		return errors.Wrap(cdomain.ErrItemConflict, "invalid audit status")
	}

	corporateProfile, err := c.corporateProfileRepository.GetByUserId(ctx, request.User.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get corporate profile")
	}

	// 法人郵便認証コードが異なる場合
	if user.PostalCode != request.CorporatePostalCode {
		return errors.Wrap(cdomain.ErrBadRequest, "invalid corporate postal code")
	}

	// マネージャーメール認証コードが異なる場合
	if corporateProfile.CorporatePostalCode != request.ManagerPostalCode {
		return errors.Wrap(cdomain.ErrBadRequest, "invalid manager postal code")
	}

	// 法人/マネージャー郵便番号認証完了
	user.PostalCodeStatus = true
	user.CodeApprovalDate = jsonutils.JSONTime(time.Now())

	corporateProfile.CorporatePostalStatus = true
	corporateProfile.CorporatePostalCodeApprovalDate = jsonutils.JSONTime(time.Now())

	user.AuditStatus = domain.COMPLETED_AUDIT
	user.SignUpApprovalDate = jsonutils.JSONTime(time.Now())

	if serviceutils.EnvIsProd() {
		// バーチャル口座割り当て
		// NOTE: 本番環境のみ自動割り当て。その他環境は手動で割り当てる。
		err = user.LinkVaAccounts(ctx, domain.EQUITY)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"message": "Failed to link va accounts",
				"error":   err,
				"time":    time.Now().String(),
			}).Error("VerifyPostalCode: failed to link va accounts")

			return err
		}
	}

	// ユーザー/プロフィール更新
	if err = c.userRepository.Update(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	if err = c.corporateProfileRepository.Update(ctx, corporateProfile); err != nil {
		return errors.Wrap(err, "failed to update corporate profile")
	}

	if err = c.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.CODE_AUDIT_COMPLETE_FOR_CORPORATE_TEMPLATE,
		Sender:       c.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{user.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"Receiver":   user.Name,
			"ServiceUrl": c.cfg.ServiceConfig.Url.UserServiceDomain,
			"ProductUrl": c.cfg.ServiceConfig.Url.UserServiceDomain + c.cfg.ServiceConfig.Url.Product,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to send mail")
	}

	return nil
}

func (c corporateUserUsecase) GetOne(ctx context.Context, request udomain.CorporateUserGetOneRequest) (udomain.CorporateUserGetOneResponse, error) {
	user, err := c.userRepository.GetById(ctx, request.User.Id)
	if err != nil {
		return udomain.CorporateUserGetOneResponse{}, errors.Wrap(err, "failed to get user")
	}

	corporateProfile, err := c.corporateProfileRepository.GetByUserId(ctx, request.User.Id)
	if err != nil {
		return udomain.CorporateUserGetOneResponse{}, errors.Wrap(err, "failed to get corporate profile")
	}

	_, corporateShareholders, err := c.corporateShareholderRepository.GetAll(ctx,
		cdomain.CorporateShareholderQueryOptionFunc().WithUserId(user.Id),
	)
	if err != nil {
		return udomain.CorporateUserGetOneResponse{}, errors.Wrap(err, "failed to get corporate shareholders")
	}

	// 유저의 가상계좌 조회 (사업형, 가상계좌 상태가 정상인 것)
	userVirtualAccount, err := c.userVirtualAccountRepository.GetByUserIdAndLinkedStatus(ctx, user.Id, domain.EQUITY, domain.ON_VA_ACCOUNT)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return udomain.CorporateUserGetOneResponse{}, errors.Wrap(err, "failed to get virtual account")
	}

	response := udomain.CorporateUserGetOneResponse{
		UserId:            user.Id,
		Email:             user.Email,
		IsNewsletter:      user.IsNewsletter,
		InvestmentPurpose: string(user.InvestmentPurpose),
		Created:           user.Created,
		AuditStatus:       user.AuditStatus,
		CorporateUserBasicProfileItem: udomain.CorporateUserBasicProfileItem{
			CorpNo:   corporateProfile.CorpNo,
			CorpName: user.Name,
			CorpType: corporateProfile.CorpType,
		},
		CorporateUserOtherProfileItem: udomain.CorporateUserOtherProfileItem{
			ZipCode:     user.ZipCode,
			Province:    user.Province,
			City:        user.City,
			Address:     user.Address,
			Address2:    user.Address2,
			MobilePhone: user.MobilePhone,
		},
		Representative: udomain.CorporateUserRepresentativeProfileItem{
			FamilyName:     corporateProfile.RepresentativeFamilyName,
			FirstName:      corporateProfile.RepresentativeFirstName,
			FamilyNameKana: corporateProfile.RepresentativeFamilyNameKana,
			FirstNameKana:  corporateProfile.RepresentativeFirstNameKana,
			Email:          corporateProfile.RepresentativeEmail,
			MobilePhone:    corporateProfile.RepresentativeMobilePhone,
		},
		Manager: udomain.CorporateUserManagerProfileItem{
			FamilyName:     lo.FromPtr(corporateProfile.ManagerFamilyName),
			FirstName:      lo.FromPtr(corporateProfile.ManagerFirstName),
			FamilyNameKana: lo.FromPtr(corporateProfile.ManagerFamilyNameKana),
			FirstNameKana:  lo.FromPtr(corporateProfile.ManagerFirstNameKana),
			Email:          lo.FromPtr(corporateProfile.ManagerEmail),
			MobilePhone:    lo.FromPtr(corporateProfile.ManagerMobilePhone),
			Department:     lo.FromPtr(corporateProfile.Department),
			Position:       lo.FromPtr(corporateProfile.Position),
			ZipCode:        corporateProfile.ManagerZipCode,
			Province:       corporateProfile.ManagerProvince,
			City:           corporateProfile.ManagerCity,
			Address:        corporateProfile.ManagerAddress,
			Address2:       corporateProfile.ManagerAddress2,
			Birthday:       corporateProfile.ManagerBirthday.ToFormat(time.DateOnly),
		},
		Shareholders: udomain.CorporateUserShareholderProfileItems{}.WithCorporateUserShareholders(corporateShareholders),
		WithdrawalAccount: udomain.CorporateUserRegisterWithdrawalAccountItem{
			BankName:    user.BankName,
			BranchName:  user.BranchName,
			BankCode:    user.BankCode,
			BranchCode:  user.BranchCode,
			DepositType: user.DepositType,
			BankAccount: user.BankAccount,
			Depositor:   user.Depositor,
		},
		InvestmentSuitability: udomain.CorporateUserConfirmInvestmentSuitabilityItem{
			EmployeeCount:     corporateProfile.EmployeeCount,
			CapitalAmount:     corporateProfile.CapitalAmount,
			Sales:             corporateProfile.Sales,
			InvestmentPurpose: user.InvestmentPurpose,
			HasExperience:     user.HasExperience,
			IsSurplusFund:     user.IsSurplusFund,
			InvestorType:      user.InvestorType,
		},
		VirtualAccount: udomain.CorporateUserVirtualAccountItem{
			BankName:       userVirtualAccount.VaBankName,
			BranchName:     userVirtualAccount.VaBranchName,
			AccountNumber:  userVirtualAccount.VaAccountNumber,
			HolderNameKana: userVirtualAccount.VaHolderNameKana,
		},
	}
	return response, nil
}

func (c corporateUserUsecase) PatchOne(ctx context.Context, request udomain.CorporateUserPatchOneRequest) error {
	user, err := c.userRepository.GetById(ctx, request.JwtUser.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	if user.UserType != domain.CORPORATE {
		return errors.Wrap(cdomain.ErrItemConflict, fmt.Sprintf("invalid user type:%s", user.UserType))
	}

	// 뉴스레터 수신 동의 변경
	user.IsNewsletter = lo.FromPtrOr(request.IsNewsletter, user.IsNewsletter)

	// TODO 향후에 추가될 필드들은 여기에 추가합니다.

	if err = c.userRepository.Update(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	return nil
}

func (c corporateUserUsecase) GetDocuments(ctx context.Context, request udomain.CorporateUserGetDocumentsRequest) (udomain.CorporateUserGetDocumentsResponse, error) {

	identificationFiles, err := c.userFileService.GetCorporateUserIdentificationFiles(ctx, request.JwtUser.Id)
	if err != nil {
		return udomain.CorporateUserGetDocumentsResponse{}, errors.Wrap(err, "failed to get corporate user identification files")
	}

	return udomain.CorporateUserGetDocumentsResponse{
		CorporateUserVerifyWithKycItem: udomain.CorporateUserVerifyWithKycItem{
			Identification1FrontFile: identificationFiles.Front1,
			Identification1BackFile:  identificationFiles.Back1,
			Identification2FrontFile: identificationFiles.Front2,
			Identification2BackFile:  identificationFiles.Back2,
			CertifiedCopyFile:        identificationFiles.CertifiedCopy,
			PoaFile:                  identificationFiles.POA,
			AoiFile:                  identificationFiles.AOI,
			ShareholdersFile:         identificationFiles.Shareholders,
			OthersFile:               identificationFiles.Others,
		},
	}, nil
}

func (c corporateUserUsecase) UpdateDocuments(ctx context.Context, request udomain.CorporateUserUpdateDocumentsRequest) error {

	// get user info
	user, err := c.userRepository.GetById(ctx, request.JwtUser.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get user by email")
	}

	// 본등록보류 상태일 때만 파일 업로드 가능
	if user.AuditStatus != domain.HOLD_AUDIT {
		return errors.Wrap(cdomain.ErrBadRequest, "invalid audit status")
	}

	// 管理画面(admin-api)の attachments_delete と同様、差し替え前に不要な
	// 既存ファイルを削除できるようにする。他人のファイルは削除できない。
	// ここでは検証のみ行い、実際の削除は新規ファイルの組み立てが成功した後で行う
	// (buildUserFiles が失敗した場合に、削除だけが先に反映されるのを防ぐため)。
	if err = c.ensureOwnsFiles(ctx, request.JwtUser.Id, request.DeleteFileIds); err != nil {
		return err
	}

	// 파일 등록을 위한 정보 셋팅
	// 同名の既存ファイルは buildUserFiles 側で除くため、ここでの取得は不要になった。
	userFiles, hadFileUrlInput, err := c.buildUserFiles(ctx, request.JwtUser.Id, request.CorporateUserFileUrlItem)
	if err != nil {
		return errors.Wrap(err, "failed to init user file")
	}

	// 追加も削除も無いなら、何も変わらないリクエストなので弾く。
	// ファイルURLが指定されていたのに全て重複でスキップされた場合も同様に弾く。
	// 削除だけは成立させて「差し替えたつもりが差し替わっていないのに成功」を
	// 返してしまうと、このチケットが直そうとした不具合より気付きにくくなる。
	if len(userFiles) == 0 && (hadFileUrlInput || len(request.DeleteFileIds) == 0) {
		return errors.Wrap(cdomain.ErrBadRequest, "no document to update")
	}

	for _, fileId := range request.DeleteFileIds {
		if err = c.userFileRepository.Delete(ctx, fileId); err != nil {
			return errors.Wrap(err, "failed to delete user file")
		}
	}

	// 削除だけのリクエストなら、まだ何も再提出されていないので審査状態は動かさない。
	if len(userFiles) == 0 {
		return nil
	}

	// update audit status
	user.AuditStatus = domain.WAIT_AUDIT
	if err = c.userRepository.Update(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	// create file info
	if err = c.userFileRepository.CreateMany(ctx, userFiles); err != nil {
		return errors.Wrap(err, "failed to create user files")
	}

	return nil
}

// corporateDocumentFileTypes はこの画面(法人書類の差し戻し対応)が扱う
// 書類の種別。buildUserFiles の fileConfigs と対応させること。
// user_file はマイナンバーやeKYCなど他の機能のファイルも同じテーブル・
// 同じ user_id で持つため、種別を絞らずに削除を許すと画面の範囲外の
// ファイルまで消せてしまう。
var corporateDocumentFileTypes = map[domain.USER_FILE_TYPE]bool{
	domain.CERTIFIED_COPY:        true,
	domain.IDENTIFICATION1_FRONT: true,
	domain.IDENTIFICATION2_FRONT: true,
	domain.POA:                   true,
	domain.IDENTIFICATION1_BACK:  true,
	domain.IDENTIFICATION2_BACK:  true,
	domain.AOI:                   true,
	domain.SHAREHOLDERS:          true,
	domain.USER_FILE_OTHERS:      true,
}

// ensureOwnsFiles は削除対象のファイルがすべて自分の所有物であり、かつ
// この画面が扱う法人書類の種別であることを確認する。存在しない、他人の
// ファイル、または対象外の種別が含まれていればエラーを返す。
func (c corporateUserUsecase) ensureOwnsFiles(ctx context.Context, userId int64, fileIds []int64) error {
	for _, fileId := range fileIds {
		file, err := c.userFileRepository.GetById(ctx, fileId)
		if err != nil {
			return err
		}
		if file.UserId != userId {
			return errors.Wrap(cdomain.ErrForbidden, "not owner of file")
		}
		if !corporateDocumentFileTypes[file.FileType] {
			return errors.Wrap(cdomain.ErrBadRequest, "file type is not a corporate document")
		}
	}
	return nil
}

// buildUserFiles は新しく追加する UserFile を組み立てる。戻り値の hadInput は
// ファイルURLが1件でも指定されていたか(結果として重複でスキップされ
// userFiles が空になった場合と、そもそも指定が無かった場合を区別するため)。
func (c corporateUserUsecase) buildUserFiles(ctx context.Context, userId int64, fileUrlInput udomain.CorporateUserFileUrlItem) (userFiles []domain.UserFile, hadInput bool, err error) {

	// このリクエストで既に組み立てた file_name
	seen := make(map[string]bool)

	// 파일 설정 배열
	fileConfigs := []struct {
		fileType domain.USER_FILE_TYPE
		fileUrl  string
	}{
		{domain.CERTIFIED_COPY, lo.FromPtr(fileUrlInput.CertifiedCopyFileUrl)},
		{domain.IDENTIFICATION1_FRONT, lo.FromPtr(fileUrlInput.Identification1FrontFileUrl)},
		{domain.IDENTIFICATION2_FRONT, lo.FromPtr(fileUrlInput.Identification2FrontFileUrl)},
		{domain.POA, lo.FromPtr(fileUrlInput.PoaFileUrl)},
		{domain.IDENTIFICATION1_BACK, lo.FromPtr(fileUrlInput.Identification1BackFileUrl)},
		{domain.IDENTIFICATION2_BACK, lo.FromPtr(fileUrlInput.Identification2BackFileUrl)},
		{domain.AOI, lo.FromPtr(fileUrlInput.AoiFileUrl)},
		{domain.SHAREHOLDERS, lo.FromPtr(fileUrlInput.ShareholdersFileUrl)},
		{domain.USER_FILE_OTHERS, lo.FromPtr(fileUrlInput.OthersFileUrl)},
	}

	for _, config := range fileConfigs {
		if len(config.fileUrl) == 0 {
			continue
		}
		hadInput = true

		file := domain.UserFile{
			FileType: config.fileType,
			FileUrl:  config.fileUrl,
			UserId:   userId,
			OwnerId:  userId,
			UpdateId: userId,
			UserType: domain.CORPORATE,
		}

		// 파일 메타정보 설정
		if err := c.setFileInfoFromUrl(&file); err != nil {
			return nil, hadInput, err
		}

		// 同じファイルの再送信なら行を作らない。
		// user_file.file_name はテーブル全体で一意で、制約は deleted を見ないため、
		// 論理削除済みの行も含めて同名がないか調べる。種別ごとの最新1件だけを見ると、
		// 管理画面で添付を削除したあとの再送信や、過去に差し替えたファイルの
		// 再送信で duplicate key になる。
		//
		// NOTE: URL では判定できない。取得した既存ファイルの FileUrl は
		// 署名付きURLへ差し替わるため、リクエストの URL とは一致しない。
		// 同一リクエスト内で同じファイルを複数の枠に入れている場合も弾く。
		// DB を見るだけでは組み立て中の重複に気付けず、同名の行を2件作ってしまう。
		if seen[file.FileName] {
			c.logger.Entry(ctx).WithFields(logrus.Fields{
				"user_id":   userId,
				"file_type": file.FileType,
				"file_name": file.FileName,
			}).Info("buildUserFiles: skipped duplicated file in the same request")
			continue
		}
		seen[file.FileName] = true

		exists, err := c.userFileRepository.ExistsByFileName(ctx, file.FileName)
		if err != nil {
			return nil, hadInput, errors.Wrap(err, "failed to check existing file name")
		}
		if exists {
			// 同名の行が既にある。差し替えではないので作らない。file_name はアップロード
			// のたびに CreatePreSignedUrl が xid で新規発行する一意な値で、元のファイル名や
			// 内容とは無関係なため、選び直しての再アップロードでは通常ここに来ない
			// (ID の衝突や、同じ署名付き URL の使い回しといった異常系でのみ発生する)。
			// 論理削除済みの行に当たった場合もここに入る。unique 制約は deleted を
			// 見ないため作れず、無言で欠けたことが分かるようログに残す。
			c.logger.Entry(ctx).WithFields(logrus.Fields{
				"user_id":   userId,
				"file_type": file.FileType,
				"file_name": file.FileName,
			}).Info("buildUserFiles: skipped already existing file name")
			continue
		}

		// 別のファイルなら差し替えとして新しい行を作る。user_file は同じ種別を
		// 複数持てる作りで、参照側は created の降順で最新1件を見るため、
		// 追加するだけで差し替えになる。
		userFiles = append(userFiles, file)
	}

	return userFiles, hadInput, nil
}

// setFileInfoFromUrl 파일 URL 에서 파일 정보 추출 및 S3 에서 파일 정보 가져오기
func (c corporateUserUsecase) setFileInfoFromUrl(userFile *domain.UserFile) error {

	// 파일 경로 및 파일 이름 추출
	filePath, err := c.fileHelper.ExtractFilePath(userFile.FileUrl)
	if err != nil {
		return errors.Wrap(err, "failed to extract file path")
	}
	userFile.FilePath = path.Dir(filePath) + "/"
	userFile.FileExtension = path.Ext(filePath)
	userFile.FileName = c.fileHelper.ExtractFileName(filePath)

	// S3 에서 파일 메타정보 가저오기
	header, err := c.s3pkg.GetHeadObject(s3utils.BUCKET_PRIVATE_NAME, userFile.GetKey())
	if err != nil {
		return errors.Wrap(err, "failed to get head object")
	}
	userFile.FileSize = header.FileSize
	userFile.OriginalName = header.OriginalFileName

	return nil
}
