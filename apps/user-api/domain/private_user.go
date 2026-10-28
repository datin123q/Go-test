package domain

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/gmoaozorautils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type PrivateUserUsecase interface {
	GetSignUpStatus(ctx context.Context, request PrvUserGetSignUpStatusRequest) (PrvUserGetSignUpStatusResponse, error)
	RegisterProfile(ctx context.Context, request PrvUserRegisterProfileRequest) (PrvUserRegisterProfileResponse, error)
	RegisterMyNumber(ctx context.Context, request PrvUserRegisterMyNumberRequest) (PrvUserRegisterMyNumberResponse, error)
	VerifyWithEkyc(ctx context.Context, request PrvUserVerifyWithEkycRequest) (PrvUserVerifyWithEkycResponse, error)
	UpdateEkycStatus(ctx context.Context, request PrvUserUpdateEkycStatusRequest) error
	GetEkycStatus(ctx context.Context, request PrvUserGetEkycStatusRequest) (PrvUserGetEkycStatusResponse, error)
	CallEkycWebhook(ctx context.Context, request PrvUserCallEkycWebhookRequest) error
	GetOne(ctx context.Context, request PrvUserGetOneRequest) (PrvUserGetOneResponse, error)
	PatchOne(ctx context.Context, request PrvUserPatchOneRequest) error
}

type PrvUserBasicProfile struct {
	// B. 個人情報1
	FirstName      string                  `json:"first_name" swagger:"desc(姓)" validate:"required"`
	FamilyName     string                  `json:"family_name" swagger:"desc(名前)" validate:"required"`
	FirstNameKana  string                  `json:"first_name_kana" swagger:"desc(姓振り仮名)" validate:"required"`
	FamilyNameKana string                  `json:"family_name_kana" swagger:"desc(名前振り仮名)" validate:"required"`
	Birthday       string                  `json:"birthday" swagger:"desc(誕生日 YYYY-MM-DD)" validate:"required,datetime=2006-01-02"`
	Gender         domain.GENDER           `json:"gender" swagger:"desc(性別: [MALE | FEMALE])" validate:"required"`
	Nationality    domain.NATIONALITY_TYPE `json:"nationality" swagger:"desc(国籍: [JP | OTHERS])" validate:"required"`
}

func (p PrvUserBasicProfile) FromPrivateUser(privateUser domain.PrivateUser) PrvUserBasicProfile {
	birthday := ""
	if !privateUser.Birthday.IsEmpty() {
		birthday = privateUser.Birthday.ToString()
	}

	return PrvUserBasicProfile{
		FirstName:      privateUser.FirstName,
		FamilyName:     privateUser.FamilyName,
		FirstNameKana:  privateUser.FirstNameKana,
		FamilyNameKana: privateUser.FamilyNameKana,
		Birthday:       birthday,
		Gender:         privateUser.Gender,
		Nationality:    privateUser.Nationality,
	}
}

type PrvUserOtherProfile struct {
	// B. 個人情報2
	ZipCode     string               `json:"zip_code" swagger:"desc(郵便番号)" validate:"required"`
	Address     string               `json:"address" swagger:"desc(住所)" validate:"required"`
	Province    string               `json:"province" swagger:"desc(都道府県)" validate:"required"`
	Address2    *string              `json:"address2" swagger:"desc(詳細住所)"` // optional
	City        string               `json:"city" swagger:"desc(市区町村)" validate:"required"`
	MobilePhone string               `json:"mobile_phone" swagger:"desc(携帯電話)" validate:"required"`
	Job         domain.USER_JOB_TYPE `json:"job" swagger:"desc(職業: USER_JOB_TYPE)" validate:"required"`
}

func (p PrvUserOtherProfile) FromPrivateUser(privateUser domain.PrivateUser) PrvUserOtherProfile {
	return PrvUserOtherProfile{
		ZipCode:     privateUser.ZipCode,
		Address:     privateUser.Address,
		Province:    privateUser.Province,
		Address2:    privateUser.Address2,
		City:        privateUser.City,
		MobilePhone: privateUser.MobilePhone,
		Job:         privateUser.Job,
	}
}

type PrvUserWithdrawalAccount struct {
	// C. 出金口座情報
	BankName    string              `json:"bank_name" swagger:"desc(銀行名)" validate:"required"`
	BranchName  string              `json:"branch_name" swagger:"desc(支店名)" validate:"required"`
	BankAccount string              `json:"bank_account" swagger:"desc(口座番号)" validate:"required,len=7,number"`
	BankCode    string              `json:"bank_code" swagger:"desc(銀行コード)" validate:"required"`
	BranchCode  string              `json:"branch_code" swagger:"desc(銀行支店コード)" validate:"required"`
	Depositor   string              `json:"depositor" swagger:"desc(預金者)" validate:"required"`
	DepositType domain.DEPOSIT_TYPE `json:"deposit_type" swagger:"desc(預金種別: [ORDINARY_ACCOUNT | CHECKING_ACCOUNT])" validate:"required"`
}

func (p PrvUserWithdrawalAccount) FromPrivateUser(privateUser domain.PrivateUser) PrvUserWithdrawalAccount {
	return PrvUserWithdrawalAccount{
		BankName:    privateUser.BankName,
		BranchName:  privateUser.BranchName,
		BankAccount: privateUser.BankAccount,
		BankCode:    lo.FromPtr(privateUser.BankCode),
		BranchCode:  lo.FromPtr(privateUser.BranchCode),
		Depositor:   privateUser.Depositor,
		DepositType: privateUser.DepositType,
	}
}

type PrvUserInvestmentSuitability struct {
	// D. 投資適合性確認
	Asset                   domain.ASSETS_STATUS           `json:"asset" swagger:"desc(資産コード: [ASSETS_STATUS])" validate:"required"`
	AnnualIncome            domain.ANNUALINCOME_STATUS     `json:"annual_income" swagger:"desc(年収コード: [ANNUALINCOME_STATUS])" validate:"required"`
	IsProInvestAsset        *bool                          `json:"is_pro_invest_asset" swagger:"desc(プロ投資資産保有の有無)" validate:"required"`
	IsProInvestFinanceAsset *bool                          `json:"is_pro_invest_finance_asset" swagger:"desc(プロ投資金融資産保有の有無)" validate:"required"`
	InvestmentPurpose       domain.INVESTMENT_PURPOSE_TYPE `json:"investment_purpose" swagger:"desc(投資目的コード: [INVESTMENT_PURPOSE_TYPE])" validate:"required"`
	IsSurplusFund           bool                           `json:"is_surplus_fund" swagger:"desc(余資運用の有無)" validate:"required"`
	HasExperience           bool                           `json:"has_experience" swagger:"desc(投資経験の有無)" validate:"required"`
}

func (p PrvUserInvestmentSuitability) FromPrivateUser(privateUser domain.PrivateUser) PrvUserInvestmentSuitability {
	return PrvUserInvestmentSuitability{
		Asset:                   privateUser.Asset,
		AnnualIncome:            privateUser.AnnualIncome,
		IsProInvestAsset:        privateUser.IsProInvestAsset,
		IsProInvestFinanceAsset: privateUser.IsProInvestFinanceAsset,
		InvestmentPurpose:       privateUser.InvestmentPurpose,
		IsSurplusFund:           lo.FromPtr(privateUser.IsSurplusFund),
		HasExperience:           lo.FromPtr(privateUser.HasExperience),
	}
}

type PrvUserMyNumber struct {
	Id        int64                   `json:"id" swagger:"desc(マイナンバーID)"`
	MyNumber  string                  `json:"my_number" swagger:"desc(マイナンバー)"`
	FrontFile PrvUserFile             `json:"front_file" swagger:"desc(マイナンバー正面)"`
	BackFile  PrvUserFile             `json:"back_file" swagger:"desc(マイナンバー背面)"`
	Status    domain.MY_NUMBER_STATUS `json:"status" swagger:"desc(マイナンバー状態)"`
}

func (p PrvUserMyNumber) FromMyNumber(myNumber domain.MyNumber) PrvUserMyNumber {
	return PrvUserMyNumber{
		Id:        myNumber.Id,
		MyNumber:  myNumber.MyNumber,
		FrontFile: PrvUserFile{}.FromUserFile(*myNumber.FrontFile),
		BackFile:  PrvUserFile{}.FromUserFile(*myNumber.BackFile),
		Status:    myNumber.Status,
	}
}

type PrvUserGetSignUpStatusRequest struct {
	JWTUser jwthelper.User       `json:"-" swaggerignore:"true"`
	Status  domain.SIGNUP_STATUS `query:"status" swagger:"desc(会員登録状態: [SIGNUP_NONE | SIGNUP_PRIVATE_PROFILE_BASIC | SIGNUP_PRIVATE_PROFILE_OTHER | SIGNUP_ACCOUNT | SIGNUP_INVESTMENT_SUITABILITY | SIGNUP_AGREE | SIGNUP_MYNUMBER | SIGNUP_KYC | SIGNUP_COMPLETE])"`
}

type PrvUserGetSignUpStatusResponse struct {
	Id                    int64                         `json:"id" swagger:"desc(ユーザーID)"`
	Email                 string                        `json:"email" swagger:"desc(メールアドレス)"`
	PrevStatus            domain.SIGNUP_STATUS          `json:"prev_status" swagger:"desc(会員登録状態: [SIGNUP_NONE | SIGNUP_PRIVATE_PROFILE_BASIC | SIGNUP_PRIVATE_PROFILE_OTHER | SIGNUP_ACCOUNT | SIGNUP_INVESTMENT_SUITABILITY | SIGNUP_AGREE | SIGNUP_MYNUMBER | SIGNUP_KYC | SIGNUP_COMPLETE])"`
	NextStatus            domain.SIGNUP_STATUS          `json:"next_status" swagger:"desc(会員登録状態: [SIGNUP_NONE | SIGNUP_PRIVATE_PROFILE_BASIC | SIGNUP_PRIVATE_PROFILE_OTHER | SIGNUP_ACCOUNT | SIGNUP_INVESTMENT_SUITABILITY | SIGNUP_AGREE | SIGNUP_MYNUMBER | SIGNUP_KYC | SIGNUP_COMPLETE])"`
	CurStatus             domain.SIGNUP_STATUS          `json:"cur_status" swagger:"desc(会員登録状態: [SIGNUP_NONE | SIGNUP_PRIVATE_PROFILE_BASIC | SIGNUP_PRIVATE_PROFILE_OTHER | SIGNUP_ACCOUNT | SIGNUP_INVESTMENT_SUITABILITY | SIGNUP_AGREE | SIGNUP_MYNUMBER | SIGNUP_KYC | SIGNUP_COMPLETE])"`
	BasicProfile          *PrvUserBasicProfile          `json:"basic_profile" swagger:"desc(個人情報入力段階(B-個人-1))"`
	OtherProfile          *PrvUserOtherProfile          `json:"other_profile" swagger:"desc(個人情報入力段階(B-個人-2))"`
	WithdrawalAccount     *PrvUserWithdrawalAccount     `json:"withdrawal_account" swagger:"desc(出金口座情報入力段階(C))"`
	InvestmentSuitability *PrvUserInvestmentSuitability `json:"investment_suitability" swagger:"desc(投資適合性確認段階(D))"`
	MyNumber              *PrvUserMyNumber              `json:"my_number" swagger:"desc(マイナンバー登録段階(F))"`
}

type PrvUserRegisterProfileRequest struct {
	JWTUser               jwthelper.User               `json:"-" swaggerignore:"true"`
	SignUpStatus          domain.SIGNUP_STATUS         `json:"sign_up_status" swagger:"desc(会員登録状態: [SIGNUP_PROFILE | SIGNUP_ACCOUNT | SIGNUP_INVESTMENT_SUITABILITY | SIGNUP_AGREE | SIGNUP_MYNUMBER | SIGNUP_KYC | SIGNUP_COMPLETE])"`
	BasicProfile          PrvUserBasicProfile          `json:"basic_profile" swagger:"desc(基本プロフィール)"`
	OtherProfile          PrvUserOtherProfile          `json:"other_profile" swagger:"desc(その他プロフィール)"`
	WithdrawalAccount     PrvUserWithdrawalAccount     `json:"withdrawal_account" swagger:"desc(出金口座情報)"`
	InvestmentSuitability PrvUserInvestmentSuitability `json:"investment_suitability" swagger:"desc(投資適合性確認)"`
}

func (p PrvUserRegisterProfileRequest) Validate() error {
	validate := validator.New()

	switch p.SignUpStatus {
	case domain.SIGNUP_PRIVATE_PROFILE_BASIC:
		if err := validate.Struct(p.BasicProfile); err != nil {
			return errors.Wrap(cdomain.ErrBadRequest, err.Error())
		}
	case domain.SIGNUP_PRIVATE_PROFILE_OTHER:
		if err := validate.Struct(p.OtherProfile); err != nil {
			return errors.Wrap(cdomain.ErrBadRequest, err.Error())
		}
	case domain.SIGNUP_ACCOUNT:
		if err := validate.Struct(p.WithdrawalAccount); err != nil {
			return errors.Wrap(cdomain.ErrBadRequest, err.Error())
		}
		// depositor(受取人名)に漢字が含まれると、振込バッチ全体がGMOあおぞらネット銀行に
		// 拒否される(MIR-1171)。個人投資家の画面ではdepositorはフリガナから自動生成され
		// 直接入力できないが、このAPI自体はリクエストの値をそのまま受け付けるため、
		// 法人と同様にここで検証する。
		if !gmoaozorautils.IsValidDepositorCharset(p.WithdrawalAccount.Depositor) {
			return errors.Wrap(cdomain.ErrBadRequest, "invalid depositor charset")
		}
	case domain.SIGNUP_INVESTMENT_SUITABILITY:
		if err := validate.Struct(p.InvestmentSuitability); err != nil {
			return errors.Wrap(cdomain.ErrBadRequest, err.Error())
		}
	case domain.SIGNUP_AGREE:
		// nothing to validate
	default:
		return errors.Wrap(cdomain.ErrBadRequest, "invalid sign up status")
	}

	return nil
}

type PrvUserRegisterProfileResponse struct {
	EKycReferenceId string `json:"e_kyc_reference_id" swagger:"desc(eKYC連携ID)"`
	// MIR-1121: eKYC送信(フロントエンドがこのレスポンス直後に外部のeKYC widgetへ
	// 遷移する)の値をブラウザのローカル状態やuser情報の非同期取得に依存させず、
	// このリクエスト内でDBから読んだ値をそのまま渡す。FamilyName・FirstName・
	// Birthday・Addressは SIGNUP_AGREE分岐でのみ設定する。氏名を結合済みの
	// 1フィールドではなく分けて返しているのは、家族名+名の結合ルール(区切り
	// 文字の有無)をフロントエンド側の既存実装1箇所に寄せ、バックエンドとの
	// 二重実装によるルールの食い違いを防ぐため(敵対的検証で、DB上のName列
	// (スペース区切り)をそのまま返すと、これまでスペース無しで送信できていた
	// 大多数のユーザーの送信値が変わってしまうことが判明した)。
	FamilyName string `json:"family_name" swagger:"desc(姓。SIGNUP_AGREEでのみ設定)"`
	FirstName  string `json:"first_name" swagger:"desc(名。SIGNUP_AGREEでのみ設定)"`
	Birthday   string `json:"birthday" swagger:"desc(生年月日 YYYY-MM-DD。SIGNUP_AGREEでのみ設定)"`
	Address    string `json:"address" swagger:"desc(都道府県+市区町村+住所+建物名。SIGNUP_AGREEでのみ設定)"`
}

type PrvUserRegisterMyNumberRequest struct {
	JWTUser      jwthelper.User `json:"-" swaggerignore:"true"`
	MyNumber     string         `json:"my_number" swagger:"desc(マイナンバー)" validate:"required"`
	FrontFileUrl string         `json:"front_file_url" swagger:"desc(マイナンバー正面)" validate:"required"`
	BackFileUrl  string         `json:"back_file_url" swagger:"desc(マイナンバー背面)" validate:"required"`
}

type PrvUserRegisterMyNumberResponse struct {
	PrvUserMyNumber `json:"my_number" swagger:"desc(マイナンバー)"`
}

type PrvUserVerifyWithEkycRequest struct {
	JWTUser jwthelper.User `json:"-" swaggerignore:"true"`
	UserId  int64          `json:"user_id" swagger:"desc(user ID)" validate:"required"`
}

type PrvUserVerifyWithEkycResponse struct {
	ReferenceId string `json:"reference_id" swagger:"desc(eKYC reference id)"`
}

type PrvUserUpdateEkycStatusRequest struct {
	JWTUser jwthelper.User `json:"-" swaggerignore:"true"`
	Status  string         `json:"status" swagger:"desc(eKYCリクエスト結果: success, fail)" validate:"required"`
	UserId  int64          `json:"user_id" swagger:"desc(ユーザーID)" validate:"required"`
}

type PrvUserGetEkycStatusRequest struct {
	UserId int64 `query:"user_id" swagger:"desc(ユーザーID)" validate:"required"`
}

type PrvUserGetEkycStatusResponse struct {
	Status *bool `json:"status" swagger:"desc(eKYCリクエスト結果: null | true | false)"`
}

type PrvUserCallEkycWebhookRequest struct {
	Status string `json:"status" swagger:"desc(eKYCリクエスト結果: success, fail)" validate:"required"`
	UserId int64  `json:"user_id" swagger:"desc(ユーザーID)" validate:"required"`
}

type PrvUserGetOneRequest struct {
	JWTUser jwthelper.User `json:"-" swaggerignore:"true"`
}

type PrvUserGetOneResponse struct {
	UserId       int64                   `json:"user_id" swagger:"desc(ユーザーID)"`
	Email        string                  `json:"email" swagger:"desc(メールアドレス)"`
	IsNewsletter bool                    `json:"is_newsletter" swagger:"desc(ニュースレター受信の有無)"`
	Created      jsonutils.JSONTime      `json:"created" swagger:"desc(登録日)"`
	Name         string                  `json:"name" swagger:"desc(名前)"`
	NameKana     string                  `json:"name_kana" swagger:"desc(名前振り仮名)"`
	Nationality  domain.NATIONALITY_TYPE `json:"nationality" swagger:"desc(国籍)"`
	Birthday     string                  `json:"birthday" swagger:"desc(誕生日 YYYY-MM-DD)"`
	ZipCode      string                  `json:"zip_code" swagger:"desc(郵便番号)"`
	Address      string                  `json:"address" swagger:"desc(住所)"`
	Province     string                  `json:"province" swagger:"desc(都道府県)"`
	Address2     *string                 `json:"address2" swagger:"desc(詳細住所)"`
	City         string                  `json:"city" swagger:"desc(市区町村)"`
	MobilePhone  string                  `json:"mobile_phone" swagger:"desc(携帯電話)"`
	Job          domain.USER_JOB_TYPE    `json:"job" swagger:"desc(職業: USER_JOB_TYPE)"`
	AuditStatus  domain.AUDIT_STATUS     `json:"audit_status" swagger:"desc(会員ステータス)"`

	InvestmentSuitability PrvUserGetOneInvestmentSuitability `json:"investment_suitability" swagger:"desc(投資適合性情報)"`
	WithdrawalAccount     PrvUserGetOneWithdrawalAccount     `json:"withdrawal_account" swagger:"desc(出金口座情報)"`
	VirtualAccount        PrvUserGetOneVirtualAccount        `json:"virtual_account" swagger:"desc(仮想口座情報)"`
}

type PrvUserGetOneInvestmentSuitability struct {
	Asset                   domain.ASSETS_STATUS           `json:"asset" swagger:"desc(資産コード: [ASSETS_STATUS])"`
	AnnualIncome            domain.ANNUALINCOME_STATUS     `json:"annual_income" swagger:"desc(年収コード: [ANNUALINCOME_STATUS])"`
	IsProInvestAsset        *bool                          `json:"is_pro_invest_asset" swagger:"desc(プロ投資資産保有の有無)"`
	IsProInvestFinanceAsset *bool                          `json:"is_pro_invest_finance_asset" swagger:"desc(プロ投資金融資産保有の有無)"`
	InvestmentPurpose       domain.INVESTMENT_PURPOSE_TYPE `json:"investment_purpose" swagger:"desc(投資目的コード: [INVESTMENT_PURPOSE_TYPE])"`
	IsSurplusFund           *bool                          `json:"is_surplus_fund" swagger:"desc(余資運用の有無)"`
	HasExperience           *bool                          `json:"has_experience" swagger:"desc(投資経験の有無)"`
}

type PrvUserGetOneWithdrawalAccount struct {
	BankName    string              `json:"bank_name" swagger:"desc(銀行名)"`
	BranchName  string              `json:"branch_name" swagger:"desc(支店名)"`
	BankAccount string              `json:"bank_account" swagger:"desc(口座番号)"`
	BankCode    string              `json:"bank_code" swagger:"desc(銀行コード)"`
	BranchCode  string              `json:"branch_code" swagger:"desc(銀行支店コード)"`
	Depositor   string              `json:"depositor" swagger:"desc(預金者)"`
	DepositType domain.DEPOSIT_TYPE `json:"deposit_type" swagger:"desc(預金種別: [ORDINARY_ACCOUNT | CHECKING_ACCOUNT])"`
}

type PrvUserGetOneVirtualAccount struct {
	BankName       string `json:"bank_name" swagger:"desc(銀行名)"`
	BranchName     string `json:"branch_name" swagger:"desc(支店名)"`
	AccountNumber  string `json:"account_number" swagger:"desc(口座番号)"`
	HolderNameKana string `json:"holder_name_kana" swagger:"desc(預金者振り仮名)"`
}

type PrvUserPatchOneRequest struct {
	JWTUser      jwthelper.User `json:"-" swaggerignore:"true"`
	IsNewsletter *bool          `json:"is_newsletter" swagger:"desc(ニュースレター受信の有無)"`
}

type PrvUserFile struct {
	Id            int64                 `json:"id" swagger:"desc(ID)"`
	OriginalName  string                `json:"original_name" swagger:"desc(元のファイル名)"`
	FileType      domain.USER_FILE_TYPE `json:"file_type" swagger:"desc(添付ファイルタイプ 共通コード [USER_FILE_TYPE])"`
	FileSize      int64                 `json:"file_size" swagger:"desc(ファイルサイズ)"`
	FileExtension string                `json:"file_extension" swagger:"desc(ファイル拡張子)"`
	FilePath      string                `json:"file_path" swagger:"desc(ファイルパス)"`
	FileUrl       string                `json:"file_url" swagger:"desc(URL)"`
	OwnerId       int64                 `json:"owner_id" swagger:"desc(所有者ID)"`
	UpdateId      int64                 `json:"update_id" swagger:"desc(最後の更新ユーザー)"`
	Created       jsonutils.JSONTime    `json:"created" swagger:"desc(作成日)"`
	Updated       jsonutils.JSONTime    `json:"updated" swagger:"desc(更新日)"`
}

func (p PrvUserFile) FromUserFile(file domain.UserFile) PrvUserFile {
	return PrvUserFile{
		Id:            file.Id,
		OriginalName:  file.OriginalName,
		FileType:      file.FileType,
		FileSize:      file.FileSize,
		FileExtension: file.FileExtension,
		FilePath:      file.FilePath,
		FileUrl:       file.FileUrl,
		OwnerId:       file.OwnerId,
		UpdateId:      file.UpdateId,
		Created:       jsonutils.JSONTime(file.Created),
		Updated:       jsonutils.JSONTime(file.Updated),
	}
}
