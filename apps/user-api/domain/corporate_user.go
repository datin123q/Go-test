package domain

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/gmoaozorautils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/samber/lo"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type CorporateUserUsecase interface {
	GetSignUpStatus(ctx context.Context, request CorporateUserGetSignUpStatusRequest) (CorporateUserGetSignUpStatusResponse, error)
	RegisterProfiles(ctx context.Context, request CorporateUserRegisterProfilesRequest) error
	VerifyWithKyc(ctx context.Context, request CorporateUserVerifyWithKycRequest) error
	VerifyPostalCode(ctx context.Context, request CorporateUserVerifyPostalCodeRequest) error
	GetOne(ctx context.Context, request CorporateUserGetOneRequest) (CorporateUserGetOneResponse, error)
	PatchOne(ctx context.Context, request CorporateUserPatchOneRequest) error
	GetDocuments(ctx context.Context, request CorporateUserGetDocumentsRequest) (CorporateUserGetDocumentsResponse, error)
	UpdateDocuments(ctx context.Context, request CorporateUserUpdateDocumentsRequest) error
}

type CorporateUserGetSignUpStatusRequest struct {
	jwthelper.User
	Status domain.SIGNUP_STATUS `query:"status" description:"会員登録ステータス" enum:"SIGNUP_NONE,SIGNUP_CORPORATE_PROFILE_BASIC,SIGNUP_CORPORATE_PROFILE_OTHER,SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE,SIGNUP_CORPORATE_PROFILE_MANAGER,SIGNUP_CORPORATE_PROFILE_SHAREHOLDER,SIGNUP_ACCOUNT,SIGNUP_INVESTMENT,SIGNUP_AGREE,SIGNUP_KYC,SIGNUP_COMPLETE"`
}

type CorporateUserGetSignUpStatusResponse struct {
	UserId                int64                                          `json:"user_id" description:"ユーザーID"`
	Email                 string                                         `json:"email" description:"メール"`
	PrevStatus            domain.SIGNUP_STATUS                           `json:"prev_status" description:"会員登録前のステータス"`
	NextStatus            domain.SIGNUP_STATUS                           `json:"next_status" description:"会員登録次のステータス"`
	CurStatus             domain.SIGNUP_STATUS                           `json:"cur_status" description:"会員登録現在のステータス"`
	BasicProfile          *CorporateUserBasicProfileItem                 `json:"basic_profile" description:"企業基本情報"`
	OtherProfile          *CorporateUserOtherProfileItem                 `json:"other_profile" description:"企業その他の情報"`
	ManagerProfile        *CorporateUserManagerProfileItem               `json:"manager_profile" description:"企業マネージャー情報"`
	RepresentativeProfile *CorporateUserRepresentativeProfileItem        `json:"representative_profile" description:"企業代表者情報"`
	ShareholderProfiles   []CorporateUserShareholderProfileItem          `json:"shareholder_profiles" description:"企業実質的支配者情報"`
	WithdrawalAccount     *CorporateUserRegisterWithdrawalAccountItem    `json:"withdrawal_account" description:"出金口座情報"`
	InvestmentSuitability *CorporateUserConfirmInvestmentSuitabilityItem `json:"investment_suitability" description:"投資適合性情報"`
	VerifyWithKyc         *CorporateUserVerifyWithKycItem                `json:"verify_with_kyc" description:"本人認証情報"`
}

type CorporateUserRegisterProfilesRequest struct {
	jwthelper.User
	SignUpStatus          domain.SIGNUP_STATUS                          `json:"sign_up_status" description:"会員登録ステータス" enum:"SIGNUP_NONE,SIGNUP_CORPORATE_PROFILE_BASIC,SIGNUP_CORPORATE_PROFILE_OTHER,SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE,SIGNUP_CORPORATE_PROFILE_MANAGER,SIGNUP_CORPORATE_PROFILE_SHAREHOLDER,SIGNUP_INVESTMENT_SUITABILITY,SIGNUP_KYC,SIGNUP_COMPLETE"`
	BasicProfile          CorporateUserBasicProfileItem                 `json:"basic_profile" description:"企業基本情報"`
	OtherProfile          CorporateUserOtherProfileItem                 `json:"other_profile" description:"企業その他の情報"`
	ManagerProfile        CorporateUserManagerProfileItem               `json:"manager_profile" description:"企業マネージャー情報"`
	RepresentativeProfile CorporateUserRepresentativeProfileItem        `json:"representative_profile" description:"企業代表者情報"`
	ShareholderProfiles   []CorporateUserShareholderProfileItem         `json:"shareholder_profiles" description:"企業実質的支配者情報"`
	WithdrawalAccount     CorporateUserRegisterWithdrawalAccountItem    `json:"withdrawal_account" description:"出金口座情報"`
	InvestmentSuitability CorporateUserConfirmInvestmentSuitabilityItem `json:"investment_suitability" description:"投資適合性情報"`
}

func (c CorporateUserRegisterProfilesRequest) Validate() error {
	validate := validator.New()

	switch c.SignUpStatus {
	case domain.SIGNUP_CORPORATE_PROFILE_BASIC:
		if err := validate.Struct(c.BasicProfile); err != nil {
			return errors.Wrap(cdomain.ErrBadRequest, err.Error())
		}
	case domain.SIGNUP_CORPORATE_PROFILE_OTHER:
		if err := validate.Struct(c.OtherProfile); err != nil {
			return errors.Wrap(cdomain.ErrBadRequest, err.Error())
		}
	case domain.SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE:
		if err := validate.Struct(c.RepresentativeProfile); err != nil {
			return errors.Wrap(cdomain.ErrBadRequest, err.Error())
		}
	case domain.SIGNUP_CORPORATE_PROFILE_MANAGER:
		if err := validate.Struct(c.ManagerProfile); err != nil {
			return errors.Wrap(cdomain.ErrBadRequest, err.Error())
		}
	case domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER:
		if len(c.ShareholderProfiles) == 0 {
			return errors.Wrap(cdomain.ErrBadRequest, "shareholder profiles is empty")
		}

		for _, shareholderProfile := range c.ShareholderProfiles {
			if err := validate.Struct(shareholderProfile); err != nil {
				return errors.Wrap(cdomain.ErrBadRequest, err.Error())
			}
		}
	case domain.SIGNUP_ACCOUNT:
		if err := validate.Struct(c.WithdrawalAccount); err != nil {
			return errors.Wrap(cdomain.ErrBadRequest, err.Error())
		}
		// depositor(受取人名)に漢字が含まれると、振込バッチ全体がGMOあおぞらネット銀行に
		// 拒否される(MIR-1171)。ここは新規登録時の入力のみで既存値は無いため、
		// 管理画面側のような「変更時のみ検証」の考慮は不要で単純に検証する。
		if !gmoaozorautils.IsValidDepositorCharset(c.WithdrawalAccount.Depositor) {
			return errors.Wrap(cdomain.ErrBadRequest, "invalid depositor charset")
		}
	case domain.SIGNUP_INVESTMENT_SUITABILITY:
		if err := validate.Struct(c.InvestmentSuitability); err != nil {
			return errors.Wrap(cdomain.ErrBadRequest, err.Error())
		}
	case domain.SIGNUP_AGREE:
		// nothing to validate
	default:
		return errors.Wrap(cdomain.ErrBadRequest, "invalid sign up status")
	}

	return nil
}

type CorporateUserBasicProfileItem struct {
	CorpNo   string `json:"corp_no" description:"法人番号" required:"true" validate:"required,len=13,numeric"`
	CorpName string `json:"corp_name" description:"法人名" required:"true" validate:"required"`
	CorpType string `json:"corp_type" description:"事業内容" required:"true" validate:"required"`
}

type CorporateUserOtherProfileItem struct {
	ZipCode     string  `json:"zip_code" description:"郵便番号" required:"true" validate:"required,len=7,numeric"`
	Province    string  `json:"province" description:"都道府県" required:"true" validate:"required"`
	City        string  `json:"city" description:"市区郡" required:"true" validate:"required"`
	Address     string  `json:"address" description:"詳細住所" required:"true" validate:"required"`
	Address2    *string `json:"address2" description:"詳細住所2"`
	MobilePhone string  `json:"mobile_phone" description:"携帯電話番号" required:"true" validate:"required,numeric"`
}

type CorporateUserRepresentativeProfileItem struct {
	FamilyName     string `json:"family_name" description:"姓" required:"true" validate:"required"`
	FirstName      string `json:"first_name" description:"名" required:"true" validate:"required"`
	FamilyNameKana string `json:"family_name_kana" description:"フリガナ姓" required:"true" validate:"required"`
	FirstNameKana  string `json:"first_name_kana" description:"フリガナ名" required:"true" validate:"required"`
	Email          string `json:"email" description:"メール" required:"true" validate:"required,email"`
	MobilePhone    string `json:"mobile_phone" description:"携帯電話番号" required:"true" validate:"required,numeric"`
}

type CorporateUserManagerProfileItem struct {
	FamilyName     string  `json:"family_name" description:"姓" required:"true" validate:"required"`
	FirstName      string  `json:"first_name" description:"名" required:"true" validate:"required"`
	FamilyNameKana string  `json:"family_name_kana" description:"フリガナ姓" required:"true" validate:"required"`
	FirstNameKana  string  `json:"first_name_kana" description:"フリガナ名" required:"true" validate:"required"`
	Email          string  `json:"email" description:"メール" required:"true" validate:"required,email"`
	MobilePhone    string  `json:"mobile_phone" description:"携帯電話番号" required:"true" validate:"required,numeric"`
	Department     string  `json:"department" description:"部署名" required:"true" validate:"required"`
	Position       string  `json:"position" description:"職位" required:"true" validate:"required"`
	ZipCode        string  `json:"zip_code" description:"郵便番号" required:"true" validate:"required,len=7,numeric"`
	Province       string  `json:"province" description:"都道府県" required:"true" validate:"required"`
	City           string  `json:"city" description:"市区郡" required:"true" validate:"required"`
	Address        string  `json:"address" description:"詳細住所" required:"true" validate:"required"`
	Address2       *string `json:"address2" description:"詳細住所2"`
	Birthday       string  `json:"birthday" description:"生年月日" required:"true" validate:"required,datetime=2006-01-02"`
}

type CorporateUserShareholderProfileItem struct {
	FamilyName     string  `json:"family_name" description:"姓" required:"true" validate:"required"`
	FirstName      string  `json:"first_name" description:"名" required:"true" validate:"required"`
	FamilyNameKana string  `json:"family_name_kana" description:"フリガナ姓" required:"true" validate:"required"`
	FirstNameKana  string  `json:"first_name_kana" description:"フリガナ名" required:"true" validate:"required"`
	ZipCode        string  `json:"zip_code" description:"郵便番号" required:"true" validate:"required,len=7,numeric"`
	Province       string  `json:"province" description:"都道府県" required:"true" validate:"required"`
	City           string  `json:"city" description:"市区郡" required:"true" validate:"required"`
	Address        string  `json:"address" description:"詳細住所" required:"true" validate:"required"`
	Address2       *string `json:"address2" description:"詳細住所2"`
	Birthday       string  `json:"birthday" description:"生年月日" required:"true" validate:"required,datetime=2006-01-02"`
	Uid            string  `json:"uid" description:"株主ユニークID"`
}

type CorporateUserRegisterWithdrawalAccountItem struct {
	BankName    string              `json:"bank_name" required:"true" validate:"required" description:"銀行名"`
	BranchName  string              `json:"branch_name" required:"true" validate:"required" description:"支店名"`
	BankCode    *string             `json:"bank_code"  description:"銀行コード"`
	BranchCode  *string             `json:"branch_code"  description:"銀行支店コード"`
	DepositType domain.DEPOSIT_TYPE `json:"deposit_type" required:"true" validate:"oneof=ORDINARY_ACCOUNT CHECKING_ACCOUNT" description:"DEPOSIT_TYPE_CODE: [ORDINARY_ACCOUNT,CHECKING_ACCOUNT]"`
	BankAccount string              `json:"bank_account" required:"true" validate:"required,len=7,number" description:"銀行口座番号"`
	Depositor   string              `json:"depositor" required:"true" validate:"required" description:"預金者"`
}

type CorporateUserConfirmInvestmentSuitabilityItem struct {
	EmployeeCount     string                         `json:"employee_count" required:"true" validate:"required" description:"社員数 企画チーム要求により文字列値"`
	CapitalAmount     string                         `json:"capital_amount" required:"true" validate:"required" description:"資本金 企画チーム要求により文字列値"`
	Sales             string                         `json:"sales" required:"true" validate:"required" description:"売上額 企画チーム要求により文字列値"`
	InvestmentPurpose domain.INVESTMENT_PURPOSE_TYPE `json:"investment_purpose" required:"true" validate:"oneof=VARIANCE_PORTFOLIO RAISING_FUTURE_MONEY MICRO_INVESTMENT RENEWABLE_ENERGY_INVESTMENT ETC" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PURPOSE" description:"投資目的"`
	HasExperience     *bool                          `json:"has_experience" required:"true" validate:"required" description:"投資経験の有無"`
	IsSurplusFund     *bool                          `json:"is_surplus_fund" required:"true" validate:"required" description:"余裕資金運用の有無"`
	InvestorType      domain.INVESTOR_TYPE           `json:"investor_type" default:"CORPORATE_INVESTOR" validate:"omitempty,oneof=CORPORATE_INVESTOR" description:"投資家タイプ"`
}

type CorporateUserVerifyWithKycItem struct {
	CertifiedCopyFile        *domain.UserFile `json:"certified_copy_file" description:"経歴書全部証明書"`
	Identification1FrontFile *domain.UserFile `json:"identification1_front_file" description:"取引担当者本人確認書類1表面"`
	Identification1BackFile  *domain.UserFile `json:"identification1_back_file" description:"取引担当者本人確認書類1裏面"`
	Identification2FrontFile *domain.UserFile `json:"identification2_front_file" description:"取引担当者本人確認書類2表面"`
	Identification2BackFile  *domain.UserFile `json:"identification2_back_file" description:"取引担当者本人確認書類2裏面"`
	PoaFile                  *domain.UserFile `json:"poa_file" description:"法人代表者委任状"`
	AoiFile                  *domain.UserFile `json:"aoi_file" description:"定款"`
	ShareholdersFile         *domain.UserFile `json:"shareholders_file" description:"株主名簿"`
	OthersFile               *domain.UserFile `json:"others_file" description:"その他"`
}

type CorporateUserVerifyWithKycRequest struct {
	JWTUser jwthelper.User
	CorporateUserFileUrlItem
}

func (c CorporateUserVerifyWithKycRequest) Validate() error {
	validate := validator.New()

	if err := validate.Struct(c); err != nil {
		return errors.Wrap(err, "validation error")
	}

	if len(lo.FromPtr(c.CorporateUserFileUrlItem.CertifiedCopyFileUrl)) == 0 {
		return errors.New("CertifiedCopyFileUrl is required")
	}
	if len(lo.FromPtr(c.CorporateUserFileUrlItem.Identification1FrontFileUrl)) == 0 {
		return errors.New("Identification1FrontFileUrl is required")
	}
	if len(lo.FromPtr(c.CorporateUserFileUrlItem.Identification2FrontFileUrl)) == 0 {
		return errors.New("Identification2FrontFileUrl is required")
	}

	return nil
}

type CorporateUserVerifyPostalCodeRequest struct {
	jwthelper.User
	CorporatePostalCode string `json:"corporate_postal_code" required:"true" validate:"required" description:"法人郵便認証コード"`
	ManagerPostalCode   string `json:"manager_postal_code" required:"true" validate:"required" description:"マネージャー郵便認証コード"`
}

type CorporateUserGetOneRequest struct {
	jwthelper.User
}

type CorporateUserGetOneResponse struct {
	UserId            int64               `json:"user_id" description:"ユーザーID"`
	Email             string              `json:"email" description:"メール"`
	IsNewsletter      bool                `json:"is_newsletter" description:"ニュースレター受信の有無"`
	InvestmentPurpose string              `json:"investment_purpose" description:"投資目的"`
	Created           jsonutils.JSONTime  `json:"created" description:"加入日"`
	AuditStatus       domain.AUDIT_STATUS `json:"audit_status" swagger:"desc(会員ステータス)"`

	CorporateUserBasicProfileItem
	CorporateUserOtherProfileItem
	Representative        CorporateUserRepresentativeProfileItem        `json:"representative" description:"代表取締役情報"`
	Manager               CorporateUserManagerProfileItem               `json:"manager" description:"取引担当者情報"`
	Shareholders          CorporateUserShareholderProfileItems          `json:"shareholders" description:"実質的支配者情報"`
	WithdrawalAccount     CorporateUserRegisterWithdrawalAccountItem    `json:"withdrawal_account" description:"出金口座情報"`
	InvestmentSuitability CorporateUserConfirmInvestmentSuitabilityItem `json:"investment_suitability" description:"投資適合性情報"`
	VirtualAccount        CorporateUserVirtualAccountItem               `json:"virtual_account" description:"仮想口座情報"`
}

type CorporateUserShareholderProfileItems []CorporateUserShareholderProfileItem

func (c CorporateUserShareholderProfileItems) WithCorporateUserShareholders(corporateShareholders []domain.CorporateShareholder) CorporateUserShareholderProfileItems {
	response := make([]CorporateUserShareholderProfileItem, 0)

	for _, corporateShareholder := range corporateShareholders {
		response = append(response, CorporateUserShareholderProfileItem{
			FamilyName:     corporateShareholder.FamilyName,
			FirstName:      corporateShareholder.FirstName,
			FamilyNameKana: corporateShareholder.FamilyNameKana,
			FirstNameKana:  corporateShareholder.FirstNameKana,
			ZipCode:        corporateShareholder.ZipCode,
			Province:       corporateShareholder.Province,
			City:           corporateShareholder.City,
			Address:        corporateShareholder.Address,
			Address2:       &corporateShareholder.Address2,
			Birthday:       corporateShareholder.Birthday.ToFormat(time.DateOnly),
			Uid:            corporateShareholder.Uid,
		})
	}

	return response
}

type CorporateUserVirtualAccountItem struct {
	BankName       string `json:"bank_name" description:"銀行名"`
	BranchName     string `json:"branch_name" description:"支店名"`
	AccountNumber  string `json:"account_number" description:"口座番号"`
	HolderNameKana string `json:"holder_name_kana" description:"預金者フリガナ"`
}

type CorporateUserPatchOneRequest struct {
	JwtUser      jwthelper.User
	IsNewsletter *bool `json:"is_newsletter" description:"ニュースレター受信の有無"`
}

type CorporateUserGetDocumentsRequest struct {
	JwtUser jwthelper.User
}

type CorporateUserGetDocumentsResponse struct {
	CorporateUserVerifyWithKycItem
}

type CorporateUserUpdateDocumentsRequest struct {
	JwtUser jwthelper.User
	CorporateUserFileUrlItem
	// DeleteFileIds は削除する既存ファイルのIDの一覧。管理画面(admin-api)の
	// attachments_delete と同様に、差し替え前に不要なファイルを明示的に
	// 削除できるようにする。自分が所有するファイルのみ削除できる。
	DeleteFileIds []int64 `json:"delete_file_ids" description:"削除するファイルIDの一覧"`
}

type CorporateUserFileUrlItem struct {
	CertifiedCopyFileUrl        *string `json:"certified_copy_file_url" description:"経歴書全部証明書"`
	Identification1FrontFileUrl *string `json:"identification1_front_file_url" description:"本人確認書類1表面"`
	Identification1BackFileUrl  *string `json:"identification1_back_file_url" description:"本人確認書類1裏面"`
	Identification2FrontFileUrl *string `json:"identification2_front_file_url" description:"本人確認書類2表面"`
	Identification2BackFileUrl  *string `json:"identification2_back_file_url" description:"本人確認書類2裏面"`
	PoaFileUrl                  *string `json:"poa_file_url" description:"法人代表者委任状"`
	AoiFileUrl                  *string `json:"aoi_file_url" description:"定款"`
	ShareholdersFileUrl         *string `json:"shareholders_file_url" description:"株主名簿"`
	OthersFileUrl               *string `json:"others_file_url" description:"その他ファイル"`
}
