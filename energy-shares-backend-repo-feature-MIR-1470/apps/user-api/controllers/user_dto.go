package controllers

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"

	"github.com/gofrs/uuid"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/otputils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/tool"
)

type UserSearchInput struct {
	Id           *string `query:"id"`
	Name         *string `query:"name"`
	Email        *string `query:"email"`
	InvestorType *string `query:"investor_type"`
	AuditStatus  *string `query:"audit_status"`
	StartDate    string  `query:"start_date"`
	EndDate      string  `query:"end_date"`
	InvestGrade  bool    `query:"investGrade"`
	BasicInfo    bool    `query:"basicInfo"`
	Withdrawal   bool    `query:"withdrawal"`
	Deposit      bool    `query:"deposit"`
	SearchInput
}

type AntiSocialUserSearchInput struct {
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
	SearchInput
}

type UserInput struct {
	FirstName               string                     `json:"first_name" form:"first_name" swagger:"desc(名前)" validate:"required"`
	FamilyName              string                     `json:"family_name"  form:"family_name" swagger:"desc(姓)"`
	FirstNameKana           string                     `json:"first_name_kana" form:"first_name_kana" swagger:"desc(名前ふりがな)"`
	FamilyNameKana          string                     `json:"family_name_kana"  form:"family_name_kana" swagger:"desc(姓ふりがな)" `
	AuditStatus             domain.AUDIT_STATUS        `json:"audit_status" form:"audit_status" swagger:"desc(登録審査ステータス 1: 審査待機中 2:審査完了適合 3:審査完了不適合),enum(NO_AUDIT|OK_AUDIT|FAIL_AUDIT),default(NO_AUDIT)"`
	BlockInvest             bool                       `json:"block_invest" form:"block_invest" swagger:"desc(投資制限 True:投資制限中 False: 投資許可)"`
	DepositType             domain.DEPOSIT_TYPE        `json:"deposit_type" form:"deposit_type" swagger:"desc(預金種類 共通コード : [DEPOSIT_TYPE])"`
	ZipCode                 string                     `json:"zip_code" form:"zip_code" swagger:"desc(郵便番号)"`
	Address                 string                     `json:"address" form:"address" swagger:"desc(住所)"`
	Province                string                     `json:"province" form:"province" swagger:"desc(都道府県)"`
	Address2                string                     `json:"address2" form:"address2" swagger:"desc(詳細住所)"`
	City                    string                     `json:"city" form:"city" swagger:"desc(市区町村)"`
	BankName                string                     `json:"bank_name" form:"bank_name" swagger:"desc(銀行名)"`
	BranchName              string                     `json:"branch_name" form:"branch_name" swagger:"desc(支店名)"`
	BankAccount             string                     `json:"bank_account" form:"bank_account" validate:"omitempty,len=7,number" swagger:"desc(銀行口座番号)"`
	BankCode                *string                    `json:"bank_code" form:"bank_code" swagger:"desc(銀行コード)"`
	BranchCode              *string                    `json:"branch_code" form:"branch_code" swagger:"desc(銀行支店コード)"`
	Depositor               string                     `json:"depositor" form:"depositor" swagger:"desc(預金者名)"`
	IsNewsletter            bool                       `json:"is_newsletter" form:"is_newsletter" swagger:"desc(ニュースレター受信希望)"`
	Asset                   domain.ASSETS_STATUS       `json:"asset" form:"asset" swagger:"desc(資産 共通コード : [ASSETS_STATUS])"`
	AnnualIncome            domain.ANNUALINCOME_STATUS `json:"annual_income" form:"annual_income" swagger:"desc(年収 共通コード : [ANNUALINCOME_STATUS])"`
	IsProInvestAsset        *bool                      `json:"is_pro_invest_asset" form:"is_pro_invest_asset" swagger:"desc(プロ投資資産保有有無)"`
	IsProInvestFinanceAsset *bool                      `json:"is_pro_invest_finance_asset" form:"is_pro_invest_finance_asset" swagger:"desc(プロ投資金融資産保有有無)"`
	InvestmentPurpose       string                     `json:"investment_purpose" form:"investment_purpose" swagger:"desc(投資目的)"`
	IsSurplusFund           *bool                      `json:"is_surplus_fund" form:"is_surplus_fund" swagger:"desc(余裕資金運用有無)"`
	PostalCodeStatus        bool                       `json:"postal_code_status" form:"postal_code_status" swagger:"desc(郵便コード認証有無 発送完了: True 未発送:False)"`
	Job                     string                     `json:"job" form:"job" swagger:"desc(職業)"`
	Gender                  domain.GENDER              `json:"gender" form:"gender" swagger:"desc(性別 : [GENDER])"`
}

func (u *UserInput) ToModel() (*domain.PrivateUser, error) {
	var NewslettersAgreeAt jsonutils.JSONTime
	if u.IsNewsletter {
		NewslettersAgreeAt = jsonutils.JSONTime(time.Now())
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
			AuditStatus:       u.AuditStatus,
			BankAccount:       u.BankAccount,
			BankName:          u.BankName,
			BranchName:        u.BranchName,
			BankCode:          u.BankCode,
			BranchCode:        u.BranchCode,
			Depositor:         u.Depositor,
			DepositType:       u.DepositType,
			IsNewsletter:      u.IsNewsletter,
			PostalCodeStatus:  u.PostalCodeStatus,
			NewsletterAgreeAt: NewslettersAgreeAt,
			InvestmentPurpose: domain.INVESTMENT_PURPOSE_TYPE(u.InvestmentPurpose),
			IsSurplusFund:     u.IsSurplusFund,
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
			Gender:                  u.Gender,
		},
	}, nil
}

type UserWithdrawalAccountInput struct {
	DepositType domain.DEPOSIT_TYPE `json:"deposit_type" form:"deposit_type" swagger:"desc(DEPOSIT_TYPE_CODE: [ORDINARY_ACCOUNT|CHECKING_ACCOUNT])"`
	BankName    string              `json:"bank_name" form:"bank_name" swagger:"desc(銀行名)"`
	BranchName  string              `json:"branch_name" form:"branch_name" swagger:"desc(支店名)"`
	BankAccount string              `json:"bank_account" form:"bank_account" validate:"required,len=7,number" swagger:"desc(銀行口座番号)"`
	BankCode    *string             `json:"bank_code" form:"bank_code" swagger:"desc(銀行コード)"`
	BranchCode  *string             `json:"branch_code" form:"branch_code" swagger:"desc(銀行支店コード)"`
}

func (u *UserWithdrawalAccountInput) ToModel() (*domain.User, error) {
	return &domain.User{
		DepositType: u.DepositType,
		BankName:    u.BankName,
		BranchName:  u.BranchName,
		BankAccount: u.BankAccount,
		BankCode:    u.BankCode,
		BranchCode:  u.BranchCode,
	}, nil
}

type UserWithdrawalAccountOutput struct {
	Id          int64               `json:"id" swagger:"desc(ID)"`
	DepositType domain.DEPOSIT_TYPE `json:"deposit_type" form:"deposit_type" swagger:"desc(DEPOSIT_TYPE_CODE: [ORDINARY_ACCOUNT|CHECKING_ACCOUNT])"`
	BankName    string              `json:"bank_name" form:"bank_name" swagger:"desc(銀行名)"`
	BranchName  string              `json:"branch_name" form:"branch_name" swagger:"desc(支店名)"`
	BankAccount string              `json:"bank_account" form:"bank_account" swagger:"desc(銀行口座番号)"`
	BankCode    *string             `json:"bank_code" form:"bank_code" swagger:"desc(銀行コード)"`
	BranchCode  *string             `json:"branch_code" form:"branch_code" swagger:"desc(銀行支店コード)"`
	Depositor   string              `json:"depositor" form:"depositor" swagger:"desc(예금주)"`
}

func (*UserWithdrawalAccountOutput) New(user *domain.User) *UserWithdrawalAccountOutput {
	return &UserWithdrawalAccountOutput{
		Id:          user.Id,
		DepositType: user.DepositType,
		BankName:    user.BankName,
		BranchName:  user.BranchName,
		BankAccount: user.BankAccount,
		BankCode:    user.BankCode,
		BranchCode:  user.BranchCode,
		Depositor:   user.Depositor,
	}
}

type EmailInput struct {
	Email string `json:"email" query:"email" form:"email" swagger:"dsec(Email)" validate:"required,email"`
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

type SignUpInput struct {
	Email                   string                     `json:"email" query:"email" form:"email" validate:"required,email"`
	Password                string                     `json:"password" query:"password" form:"password" validate:"required"`
	Birthday                string                     `json:"birthday" query:"birthday" form:"birthday" validate:"required"`
	FirstName               string                     `json:"first_name" query:"first_name" form:"first_name" validate:"required"`
	FamilyName              string                     `json:"family_name"  query:"family_name" form:"family_name" `
	FirstNameKana           string                     `json:"first_name_kana" query:"first_name_kana" form:"first_name_kana" `
	FamilyNameKana          string                     `json:"family_name_kana" query:"family_name_kana" form:"family_name_kana" swagger:"desc(名前姓ふりがな)" `
	DepositType             domain.DEPOSIT_TYPE        `json:"deposit_type" query:"deposit_type" form:"deposit_type" swagger:"desc(預金種類 共通コード : [DEPOSIT_TYPE])"`
	Asset                   domain.ASSETS_STATUS       `json:"asset" query:"asset" form:"asset" validate:"oneof=ASSET1 ASSET2 ASSET3 ASSET4 ASSET5 ASSET6 ASSET7" swagger:"desc(資産 共通コード : [ASSETS_STATUS])"`
	AnnualIncome            domain.ANNUALINCOME_STATUS `json:"annual_income" query:"annual_income" form:"annual_income" validate:"oneof=ANNUALINCOME1 ANNUALINCOME2 ANNUALINCOME3 ANNUALINCOME4 ANNUALINCOME5 ANNUALINCOME6"  swagger:"desc(年収 共通コード : [ANNUALINCOME_STATUS])"`
	ZipCode                 string                     `json:"zip_code" query:"zip_code" form:"zip_code" swagger:"desc(郵便番号)"`
	Address                 string                     `json:"address" query:"address" form:"address" swagger:"desc(住所)"`
	Province                string                     `json:"province" query:"province" form:"province" swagger:"desc(都道府県)"`
	Address2                string                     `json:"address2" query:"address2" form:"address2" swagger:"desc(詳細住所)"`
	City                    string                     `json:"city" query:"city" form:"city" swagger:"desc(市区町村)"`
	BankName                string                     `json:"bank_name" query:"bank_name" form:"bank_name" swagger:"desc(銀行名)"`
	BranchName              string                     `json:"branch_name" query:"branch_name" form:"branch_name" swagger:"desc(支店名)"`
	BankAccount             string                     `json:"bank_account" query:"bank_account" form:"bank_account" swagger:"desc(銀行口座番号)"`
	BankCode                *string                    `json:"bank_code" query:"bank_code" form:"bank_code" swagger:"desc(銀行コード)"`
	BranchCode              *string                    `json:"branch_code" query:"branch_code" form:"branch_code" swagger:"desc(銀行支店)"`
	Depositor               string                     `json:"depositor" query:"depositor" form:"depositor" swagger:"desc(預金者名)"`
	IsProInvestAsset        *bool                      `json:"is_pro_invest_asset" query:"is_pro_invest_asset" form:"is_pro_invest_asset" swagger:"desc(プロ投資資産保有有無)"`
	IsProInvestFinanceAsset *bool                      `json:"is_pro_invest_finance_asset" query:"is_pro_invest_finance_asset" form:"is_pro_invest_finance_asset" swagger:"desc(プロ投資金融資産保有有無)"`
	MobilePhone             string                     `json:"mobile_phone" query:"mobile_phone" form:"mobile_phone" swagger:"desc(携帯電話)"`
	Nationality             domain.NATIONALITY_TYPE    `json:"nationality" query:"nationality" form:"nationality" swagger:"desc(国籍 jp:日本, others: その他)"`
	InvestmentPurpose       string                     `json:"investment_purpose" query:"investment_purpose" form:"investment_purpose" swagger:"desc(投資目的)"`
	IsSurplusFund           *bool                      `json:"is_surplus_fund" query:"is_surplus_fund" form:"is_surplus_fund" swagger:"desc(余裕資金運用有無)"`
	IsNewsletter            bool                       `json:"is_newsletter" query:"is_newsletter" form:"is_newsletter" swagger:"enum(true|false),desc(ニュースレター受信希望)"`
	Job                     string                     `json:"job" query:"job" form:"job" swagger:"desc(職業)" validate:"required"`
	Gender                  domain.GENDER              `json:"gender" query:"gender" form:"gender" swagger:"desc(性別: [GENDER])"`
	IsEkyc                  bool                       `json:"is_ekyc" query:"is_ekyc" form:"is_ekyc" swagger:"desc(オンライン認証有無)"`
	RequestId               string                     `json:"request_id" query:"request_id" form:"request_id" swagger:"desc(会員登録リクエスト時に発行されたID)"`
	HasExperience           *bool                      `json:"has_experience" query:"has_experience" form:"has_experience" swagger:"desc(投資経験有無)"`
}

type PasswordResetRequestInput struct {
	Email string `json:"email" form:"email" validate:"required,email"`
}

type PasswordResetInput struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
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

// 文字認証リクエスト
type MobilePhoneVerificationInput struct {
	UserId      int64                    `json:"user_id" swagger:"description(userID)"`
	MobilePhone string                   `json:"mobile_phone" swagger:"description(携帯電話番号)" validate:"required"`
	FirstReq    bool                     `json:"first_request" swagger:"description(初回発送有無)"`
	PhoneType   domain.MOBILE_PHONE_TYPE `json:"type" swagger:"description(携帯番号タイプ: DEFAULT, CEO, MANAGER)"`
}

func (m *MobilePhoneVerificationInput) ToModel() (*domain.MobilePhoneVerificationRequest, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	requestId := id.String()
	optCode := otputils.EncodeToString(6)
	Expire := time.Now().Add(3 * time.Minute)
	return &domain.MobilePhoneVerificationRequest{
		RequestId:   requestId,
		UserId:      m.UserId,
		Code:        optCode,
		MobilePhone: m.MobilePhone,
		Expire:      jsonutils.JSONTime(Expire),
		FirstReq:    m.FirstReq,
		PhoneType:   m.PhoneType,
	}, nil
}

// 決済認証レスポンス
type MobilePhoneVerificationResponse struct {
	UserId    int64  `json:"user_id"`
	Code      string `json:"code"`
	RequestId string `json:"request_id"`
}

func (v *PasswordResetRequestInput) ToModel() (*domain.PasswordResetRequest, error) {

	return &domain.PasswordResetRequest{
		Email:  v.Email,
		Expire: jsonutils.JSONTime(time.Now().Add(time.Minute * 20)),
	}, nil
}

type SignUpRequestInput struct {
	Email             string `json:"email" form:"email" validate:"required,email"`
	MarketingCode     string `json:"marketing_code" form:"marketing_code"`
	MarketingCampaign string `json:"marketing_campaign" form:"marketing_campaign"`
	MarketingKey      string `json:"marketing_key" form:"marketing_key"`
}

type EmailCodeInput struct {
	Email    string `json:"email" form:"email" swagger:"description(Email)" validate:"required,email"`
	Name     string `json:"name" form:"name" swagger:"description(名前)" validate:"required"`
	UserId   int64  `json:"user_id" swagger:"description(userID)"`
	FirstReq bool   `json:"first_request" swagger:"description(初回発送有無)"`
}

func (v *EmailCodeInput) ToModel() (*domain.EmailCodeRequest, error) {

	code, err := tool.RandomMailCode(6)

	if err != nil {
		return nil, err
	}

	return &domain.EmailCodeRequest{
		Email:    v.Email,
		Code:     code,
		Expire:   jsonutils.JSONTime(time.Now().Add(time.Minute * 3)),
		UserId:   v.UserId,
		FirstReq: v.FirstReq,
	}, nil
}

type EmailCodeCheckInput struct {
	Email string `json:"email" form:"email" swagger:"dsec(Email)" validate:"required,email"`
	Code  string `json:"code" form:"code" swagger:"dsec(code)" validate:"required"`
}

type UserAntisocialCheckDto struct {
	CheckStatus   domain.ANTISOCIAL_CHECK_STATUS `json:"check_status" form:"check_status" swagger:"dsec(反社会的検査ステータス: 承認:true,非承認:false)"`
	Checker       string                         `json:"checker" form:"checker"  swagger:"dsec(反社会的チェッカー)"`
	InvalidReason string                         `json:"invalid_reason,omitempty" form:"invalid_reason" swagger:"dsec(無効理由)"`
}

func (d UserAntisocialCheckDto) ToModel() (domain.UserAntisocialCheck, error) {
	return domain.UserAntisocialCheck{
		CheckStatus: d.CheckStatus,
		Checker:     d.Checker,
	}, nil
}

type UserAuditStatusDto struct {
	AuditStatus   string  `json:"audit_status" form:"audit_status" swagger:"desc(登録承認ステータスコード値)"`
	InvalidReason *string `json:"invalid_reason,omitempty" form:"invalid_reason" swagger:"desc(無効理由)"`
}

type InvestmentFileType struct {
	ProjectFileType domain.PROJECT_FILE_TYPE `json:"project_file_type" form:"project_file_type"`
}

type UserByDividendsInput struct {
	StartDate             string  `query:"start_date"`
	EndDate               string  `query:"end_date"`
	TransactionDetailType *string `query:"transaction_detail_type"`
	ProjectId             *string `query:"project_id"`
	ProjectName           *string `query:"project_name"`
	SearchInput
}

type TransactionSearchInput struct {
	ProjectType *string `query:"project_type"`
	SearchInput
}

type ReportResponse struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	FileUrl string `json:"file_url"`
}

func (ReportResponse) NewReportResponse(id int64, title, fileUrl string) *ReportResponse {
	return &ReportResponse{
		Id:      id,
		Title:   title,
		FileUrl: fileUrl,
	}
}

type UserWaitlistResponse struct {
	TotalCount int64                      `json:"total_count" swagger:"desc(総数)"`
	Items      []UserWaitlistResponseItem `json:"items" swagger:"desc(キャンセル待ち一覧)"`
}

type UserWaitlistResponseItem struct {
	Id            int64                            `json:"id"`
	ProjectName   string                           `json:"project_name"`
	Status        domain.INVESTMENT_WAITING_STATUS `json:"status" swagger:"desc(表示用キャンセル待ちステータス),enum(WAITING|END_FUNDING|CANCELLED|WINNING|WINNING_AFTER_SUCCESS|WINNING_AFTER_CANCELED|WINNING_AFTER_FAILED|PROPOSAL)"`
	InvestStatus  *domain.INVEST_STATUS            `json:"invest_status" swagger:"desc(投資ステータス。Share未作成またはステータス未設定時はnull),enum(REQUEST_INVEST|FAIL_INVEST|CANCEL_INVEST|SUCCESS_INVEST|OPERATION_INVEST|REFUND_INVEST)"`
	ProjectStatus *domain.PROJECT_STATUS           `json:"project_status" swagger:"desc(プロジェクトステータス。ステータス未設定時はnull),enum(STANDBY_FUNDING|OPEN_FUNDING|END_FUNDING|SUCCESS_FUNDING|FAIL_FUNDING|OPERATION_FUNDING|OPERATION_END_FUNDING|CLOSE_FUNDING|INVALID_FUNDING)"`
	ProjectId     int64                            `json:"project_id" swagger:"desc(プロジェクトID)"`
	Qty           int64                            `json:"qty" swagger:"desc(投資待機申請数量)"`
	Amount        int64                            `json:"amount" swagger:"desc(投資待機申請金額)"`
	WaitingDate   jsonutils.JSONTime               `json:"waiting_date" swagger:"desc(投資待機申請時間)"`
	ShareId       string                           `json:"share_id" swagger:"desc(取引ID)"`
}

func newUserWaitlistResponseItem(waiter domain.InvestmentWaiters) UserWaitlistResponseItem {
	return UserWaitlistResponseItem{
		Id:            waiter.Id,
		ProjectId:     waiter.ProjectId,
		Qty:           waiter.Qty,
		Status:        waiter.GetDisplayStatus(),
		InvestStatus:  waiter.InvestStatus,
		ProjectStatus: waiter.ProjectStatus,
		ProjectName:   waiter.ProjectName,
		Amount:        waiter.Amount,
		WaitingDate:   waiter.WaitingDate,
		ShareId:       waiter.ShareId,
	}
}
