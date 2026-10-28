package controllers

import (
	"time"

	"github.com/gofrs/uuid"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/otputils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/tool"
)

type UserSearchInput struct {
	Id                  *string `query:"id"`
	Name                *string `query:"name"`
	Email               *string `query:"email"`
	InvestorType        *string `query:"investor_type"`
	AuditStatus         *string `query:"audit_status"`
	RequestId           *string `query:"request_id"`
	StartDate           string  `query:"start_date"`
	EndDate             string  `query:"end_date"`
	InvestGrade         bool    `query:"investGrade"`
	BasicInfo           bool    `query:"basicInfo"`
	Withdrawal          bool    `query:"withdrawal"`
	LoanDeposit         bool    `query:"loanDeposit"`
	EquityDeposit       bool    `query:"equityDeposit"`
	NewsletterAgreement bool    `query:"newsletterAgreement"`
	ProfileSearchInput
	SearchInput
}

type UserSearchSimpleInput struct {
	Key         *string `query:"key" validate:"required" swagger:"required,desc(検索キーex)email)"`
	Value       *string `query:"value" swagger:"desc(検索値)"`
	AuditStatus *string `query:"audit_status" swagger:"desc(登録状態)"`
	SearchInput
}

type UserSearchSimpleOutput struct {
	Id    int64  `json:"id" swagger:"desc(ユーザーID)"`
	Email string `query:"email" swagger:"desc(ユーザーメール)"`
}

func (p UserSearchSimpleOutput) New(user *domain.User) *UserSearchSimpleOutput {
	return &UserSearchSimpleOutput{
		Id:    user.Id,
		Email: user.Email,
	}
}

func (p UserSearchSimpleOutput) NewList(userList []domain.User) []*UserSearchSimpleOutput {
	var list []*UserSearchSimpleOutput
	for _, u := range userList {
		list = append(list, p.New(&u))
	}
	return list
}

type AntiSocialUserSearchInput struct {
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
	SearchInput
	QuerySearchInput
}

type UserInput struct {
	// WARNING: 管理者が「個人会員」情報を修正する際に使用するDTO。フィールドの追加・削除時に予期しない値が変更される可能性があるため注意が必要。
	FirstName               string                     `json:"first_name" form:"first_name" swagger:"desc(名前)" validate:"required"`
	FamilyName              string                     `json:"family_name"  form:"family_name" swagger:"desc(姓)"`
	FirstNameKana           string                     `json:"first_name_kana" form:"first_name_kana" swagger:"desc(名前フリガナ)"`
	FamilyNameKana          string                     `json:"family_name_kana"  form:"family_name_kana" swagger:"desc(姓フリガナ)" `
	DepositType             domain.DEPOSIT_TYPE        `json:"deposit_type" form:"deposit_type" swagger:"desc(預金種類共通コード:[DEPOSIT_TYPE])"`
	ZipCode                 string                     `json:"zip_code" form:"zip_code" swagger:"desc(郵便番号)"`
	Address                 string                     `json:"address" form:"address" swagger:"desc(住所)"`
	Province                string                     `json:"province" form:"province" swagger:"desc(都道府県)"`
	Address2                string                     `json:"address2" form:"address2" swagger:"desc(詳細住所)"`
	City                    string                     `json:"city" form:"city" swagger:"desc(市)"`
	BankName                string                     `json:"bank_name" form:"bank_name" swagger:"desc(銀行名)"`
	BranchName              string                     `json:"branch_name" form:"branch_name" swagger:"desc(支店名)"`
	BankAccount             string                     `json:"bank_account" form:"bank_account" validate:"omitempty,len=7,number" swagger:"desc(銀行口座番号)"`
	BankCode                *string                    `json:"bank_code" form:"bank_code" swagger:"desc(銀行コード)"`
	BranchCode              *string                    `json:"branch_code" form:"branch_code" swagger:"desc(銀行支店コード)"`
	Asset                   domain.ASSETS_STATUS       `json:"asset" form:"asset" swagger:"desc(資産共通コード:[ASSETS_STATUS])"`
	AnnualIncome            domain.ANNUALINCOME_STATUS `json:"annual_income" form:"annual_income" swagger:"desc(年収共通コード:[ASSETS_STATUS])"`
	IsProInvestAsset        *bool                      `json:"is_pro_invest_asset" form:"is_pro_invest_asset" swagger:"desc(プロ投資資産保有有無)"`
	IsProInvestFinanceAsset *bool                      `json:"is_pro_invest_finance_asset" form:"is_pro_invest_finance_asset" swagger:"desc(プロ投資金融資産保有有無)"`
	InvestmentPurpose       string                     `json:"investment_purpose" form:"investment_purpose" swagger:"desc(投資目的)"`
	IsSurplusFund           *bool                      `json:"is_surplus_fund" form:"is_surplus_fund" swagger:"desc(余裕資金運用有無)"`
	PostalCodeStatus        bool                       `json:"postal_code_status" form:"postal_code_status" swagger:"desc(郵便コード認証有無送信、完了:True | 未送信:False)"`
	HasExperience           *bool                      `json:"has_experience" form:"has_experience" swagger:"desc(投資経験有無)"`
	IsNewsletter            *bool                      `json:"is_newsletter" form:"is_newsletter" swagger:"desc(メルマガ購読有無)"`
	// MIR-1173: 本登録承認完了・無効の会員は変更を受け付けない
	// (private_user.go の Update ハンドラで beforeUser の値へ戻す)。
	MobilePhone string                  `json:"mobile_phone" form:"mobile_phone" swagger:"desc(電話番号。本登録承認完了・無効の会員は変更不可)"`
	Nationality domain.NATIONALITY_TYPE `json:"nationality" form:"nationality" swagger:"desc(国籍共通コード:[NATIONALITY_TYPE]。本登録承認完了・無効の会員は変更不可)"`
	Gender      domain.GENDER           `json:"gender" form:"gender" swagger:"desc(性別共通コード:[GENDER]。本登録承認完了・無効の会員は変更不可)"`
	Birthday    string                  `json:"birthday" form:"birthday" swagger:"desc(生年月日。本登録承認完了・無効の会員は変更不可)"`
}

type EmailInput struct {
	Email string `json:"email" form:"email" swagger:"dsec(Email)" validate:"required,email"`
}

func (u *UserInput) ToModel() (*domain.PrivateUser, error) {
	// MIR-1173: 生年月日未設定の古い会員(本登録承認後にトラッキングを始めた
	// ため、既存データにゼロ値のまま残っているケースがある)を編集する際、
	// 空文字のままエラーにすると本登録承認完了・無効の会員は生年月日欄が
	// disabledで直せないため、更新自体が一切できなくなる。空文字はゼロ値の
	// まま通し、本登録承認前後の判定はrestrictPersonalInfoEditsAfterFinalAudit
	// (呼び出し元のprivate_user.go)に委ねる。
	var birthday jsonutils.JSONTime
	if u.Birthday != "" {
		parsed, err := jsonutils.ParseJSONTimeJST(time.DateTime, u.Birthday)
		if err != nil {
			return nil, err
		}
		birthday = parsed
	}

	return &domain.PrivateUser{
		User: domain.User{
			UserType:          domain.PRIVATE,
			Name:              u.FamilyName + " " + u.FirstName,
			ZipCode:           u.ZipCode,
			Address:           u.Address,
			Province:          u.Province,
			Address2:          &u.Address2,
			City:              u.City,
			BankAccount:       u.BankAccount,
			BankName:          u.BankName,
			BranchName:        u.BranchName,
			BankCode:          u.BankCode,
			BranchCode:        u.BranchCode,
			Depositor:         u.FamilyNameKana + "　" + u.FirstNameKana, // スペース文字注意（全角スペース）
			DepositType:       u.DepositType,
			PostalCodeStatus:  u.PostalCodeStatus,
			InvestmentPurpose: domain.INVESTMENT_PURPOSE_TYPE(u.InvestmentPurpose),
			IsSurplusFund:     u.IsSurplusFund,
			HasExperience:     u.HasExperience,
			MobilePhone:       u.MobilePhone,
		},
		PrivateProfile: domain.PrivateProfile{
			Asset:                   u.Asset,
			AnnualIncome:            u.AnnualIncome,
			IsProInvestAsset:        u.IsProInvestAsset,
			IsProInvestFinanceAsset: u.IsProInvestFinanceAsset,
			FirstName:               u.FirstName,
			FamilyName:              u.FamilyName,
			FirstNameKana:           u.FirstNameKana,
			FamilyNameKana:          u.FamilyNameKana,
			Nationality:             u.Nationality,
			Gender:                  u.Gender,
			Birthday:                birthday,
		},
	}, nil
}

type LoginInput struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required"`
}

func (u *LoginInput) ToModel() (*domain.User, error) {

	return &domain.User{
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

type PasswordResetRequestInput struct {
	Email string `json:"email" form:"email" validate:"required,email"`
}

type PasswordResetInput struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required"`
}

type PasswordChangeInput struct {
	Password string `json:"password" form:"password" validate:"required"`
}

type PasswordCheckInput struct {
	Password string `json:"password" form:"password" validate:"required"`
}

type PostalCodeInput struct {
	PostalCode string `json:"postal_code" form:"postal_code" validate:"required"`
}

type CorpPostalCodeInput struct {
	PostalCode          string  `json:"postal_code" form:"postal_code" validate:"required"`
	CorporatePostalCode string  `json:"corporate_postal_code" form:"corporate_postal_code" validate:"required"`
	CorpType            *string `json:"corp_type" form:"corp_type"`
}

// SMS認証リクエスト
type MobilePhoneVerificationInput struct {
	UserId      int64  `json:"user_id"`
	MobilePhone string `json:"mobile_phone"`
}

func (m *MobilePhoneVerificationInput) ToModel() (*domain.MobilePhoneVerificationRequest, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	requstId := id.String()
	optCode := otputils.EncodeToString(6)
	Expire := time.Now().Add(60 * time.Second)
	return &domain.MobilePhoneVerificationRequest{
		RequestId:   requstId,
		UserId:      m.UserId,
		Code:        optCode,
		MobilePhone: m.MobilePhone,
		Expire:      jsonutils.JSONTime(Expire),
	}, nil
}

// 決済認証応答
type MobilePhoneVerificationResponse struct {
	UserId    int64  `json:"user_id"`
	Code      string `json:"code"`
	RequestId string `json:"request_id"`
}

func (v *PasswordResetRequestInput) ToModel() (*domain.PasswordResetRequest, error) {

	return &domain.PasswordResetRequest{
		Email: v.Email,
	}, nil
}

type SignUpRequestInput struct {
	Email string `json:"email" form:"email" validate:"required,email"`
}

func (v *SignUpRequestInput) ToModel() (*domain.SignUpRequest, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	requstId := id.String()
	url := serviceutils.GetUserServiceDomain() + "/signUp/basic" + "/" + requstId

	return &domain.SignUpRequest{
		RequestId:   requstId,
		Email:       v.Email,
		Url:         string(url),
		EmailExpire: jsonutils.JSONTime(time.Now().Add(time.Minute * 20)),
	}, nil
}

type EkycRequestInput struct {
	RequestId string `json:"request_id" form:"request_id" validate:"required"`
}

type EmailCodeInput struct {
	Email string `json:"email" form:"email" swagger:"dsec(Email)" validate:"required,email"`
	Name  string `json:"name" form:"name" swagger:"dsec(名前)" validate:"required"`
}

func (v *EmailCodeInput) ToModel() (*domain.EmailCodeRequest, error) {

	code, err := tool.RandomMailCode(6)

	if err != nil {
		return nil, err
	}

	return &domain.EmailCodeRequest{
		Email:  v.Email,
		Code:   code,
		Expire: jsonutils.JSONTime(time.Now().Add(time.Minute * 20)),
	}, nil
}

type EmailCodeCheckInput struct {
	Email string `json:"email" form:"email" swagger:"dsec(Email)" validate:"required,email"`
	Code  string `json:"code" form:"code" swagger:"dsec(コード)" validate:"required"`
}

type UserAntisocialCheckDto struct {
	CheckStatus   domain.ANTISOCIAL_CHECK_STATUS `json:"check_status" form:"check_status" swagger:"dsec(反社会検査状態:未承認:0 承認:1,非承認:2)"`
	Checker       string                         `json:"checker" form:"checker" validate:"min=1,max=200" swagger:"dsec(反社会チェック者)"`
	InvalidReason string                         `json:"invalid_reason,omitempty" form:"invalid_reason" swagger:"dsec(無効事由)"`
}

func (d UserAntisocialCheckDto) ToModel() (domain.UserAntisocialCheck, error) {
	return domain.UserAntisocialCheck{
		CheckStatus: d.CheckStatus,
		Checker:     d.Checker,
	}, nil
}

type UserAuditStatusDto struct {
	AuditStatus   string  `json:"audit_status" form:"audit_status" param:"audit_status" swagger:"desc(登録承認状態コード値)"`
	InvalidReason *string `json:"invalid_reason,omitempty" form:"invalid_reason" swagger:"desc(無効事由)"`
}

type InvestorTypeDto struct {
	InvestorType domain.INVESTOR_TYPE `json:"investor_type" form:"investor_type" swagger:"desc(投資家タイプコード値)"`
}

type LinkVaAccountProjectType struct {
	ProjectType domain.PROJECT_TYPE `json:"project_type" form:"project_type"`
}

type UserByDividendsInput struct {
	StartDate             string  `query:"start_date"`
	EndDate               string  `query:"end_date"`
	TransactionDetailType *string `query:"transaction_detail_type"`
	ProjectId             *string `query:"project_id"`
	SearchInput
}
