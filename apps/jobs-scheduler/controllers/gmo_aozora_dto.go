package controllers

import (
	"strconv"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// GetLoginInput はGetLoginへのクエリパラメータ。stateはadmin-apiが発行した
// ワンタイム値(S)で、gmo_aozora_oauth_stateに保存済みでなければならない
// (MIR-1298)。
type GetLoginInput struct {
	State string `json:"state" query:"state"`
}

type RedirectUriInput struct {
	Code             *string `json:"code,omitempty"  query:"code"`                          // 最終承認されたコード
	State            *string `json:"state,omitempty" query:"state"`                         // リクエスト時に指定された値をそのまま返す
	Error            *string `json:"error,omitempty" query:"error"`                         // エラーコード
	ErrorDescription *string `json:"error_description,omitempty" query:"error_description"` // エラー内容
	ErrorUri         *string `json:"error_uri,omitempty" query:"error_uri"`                 // エラーに関する追加情報を含むWebページ
}

type Account struct {
	RaId             string `json:"raid"`
	RaBranchCode     string `json:"raBranchCode"`
	RaBranchNameKana string `json:"raBranchNameKana"`
	RaAccountNumber  string `json:"raAccountNumber"`
	RaHolderName     string `json:"raHolderName"`
	BaseDate         string `json:"baseDate"`
	BaseTime         string `json:"baseTime"`
}

type VaTransaction struct {
	VaId              string `json:"vaId,omitempty"`
	TransactionDate   string `json:"transactionDate,omitempty"`
	ValueDate         string `json:"valueDate,omitempty"`
	VaBranchCode      string `json:"vaBranchCode,omitempty"`
	VaBranchNameKana  string `json:"vaBranchNameKana,omitempty"`
	VaAccountNumber   string `json:"vaAccountNumber,omitempty"`
	VaAccountNameKana string `json:"vaAccountNameKana,omitempty"`
	DepositAmount     string `json:"depositAmount,omitempty"`
	RemitterNameKana  string `json:"remitterNameKana,omitempty"`
	PaymentBankName   string `json:"paymentBankName,omitempty"`
	PaymentBranchName string `json:"paymentBranchName,omitempty"`
	PartnerName       string `json:"partnerName,omitempty"`
	Remarks           string `json:"remarks,omitempty"`
	ItemKey           string `json:"itemKey,omitempty"`
}

type CsvVaDepositDto struct {
	TransactionDate   string
	VaId              string
	Desc              string
	VaBranchCode      string
	VaAccountNumber   string
	VaHolderNameKana  string
	PaymentBranchCode string
	RemitterNameKana  string
	DepositAmount     string
}

func (t VaTransaction) ToModel() (*domain.VaDeposit, error) {
	if err := domain.ValidateVaDepositItemKey(t.ItemKey); err != nil {
		return nil, err
	}

	depositAmount, err := strconv.ParseInt(t.DepositAmount, 10, 64)
	if err != nil {
		return nil, err
	}

	depositTime, err := jsonutils.ParseJSONTimeJST(time.DateOnly, t.TransactionDate)
	if err != nil {
		return nil, err
	}

	return &domain.VaDeposit{
		VaId:                  t.VaId,
		TransactionDate:       t.TransactionDate,
		ValueDate:             t.ValueDate,
		VaBranchCode:          t.VaBranchCode,
		VaBranchNameKana:      t.VaBranchNameKana,
		VaAccountNameKana:     t.VaAccountNameKana,
		RemitterNameKana:      t.RemitterNameKana,
		PaymentBankName:       t.PaymentBankName,
		PaymentBranchName:     t.PaymentBranchName,
		PartnerName:           t.PartnerName,
		Remarks:               t.Remarks,
		ItemKey:               t.ItemKey,
		VaAccountNumber:       t.VaAccountNumber,
		DepositAmount:         depositAmount,
		DepositDate:           jsonutils.JSONTime(depositTime),
		UseAmount:             0,
		DepositStatus:         domain.NOT_USE_DEPOSIT,
		TransactionDetailType: domain.DEPOSIT_VIRTUAL_ACCOUNT,
	}, nil
}

type AozoraWebHookMessageHeader struct {
	XAccessToken      string `json:"x-access-token"`
	XWebhookSignature string `json:"x-webhook-signature"`
	UserAgent         string `json:"User-UserAgent"`
	Authorization     string `json:"authorization"`
	XEventType        string `json:"x-eventType"`
}

type AozoraWebHookMessageInput struct {
	MessageId     string        `json:"messageId,omitempty"`
	TimeStamp     string        `json:"timestamp,omitempty"`
	Account       Account       `json:"account,omitempty"`
	VaTransaction VaTransaction `json:"vaTransaction,omitempty"`
}

type TransferInput struct {
	ItemId          string                         `json:"item_id,omitempty"`
	WithdrawalId    int64                          `json:"withdrawal_id,omitempty"`
	UserId          int64                          `json:"user_id,omitempty"`
	ProjectId       int64                          `json:"project_id,omitempty"`
	TransactionId   string                         `json:"transaction_id,omitempty"`
	RequestDate     string                         `json:"request_date,omitempty"`
	RequestAmount   int64                          `json:"request_amount,omitempty"`
	AccountNumber   string                         `json:"account_number,omitempty"`
	BankCode        string                         `json:"bank_code,omitempty"`
	BankName        string                         `json:"bank_name,omitempty"`
	BranchCode      string                         `json:"branch_code,omitempty"`
	BranchName      string                         `json:"branch_name,omitempty"`
	Depositor       string                         `json:"depositor,omitempty"`
	AccountTypeCode string                         `json:"account_type_code,omitempty"`
	RequestStatus   domain.TRANSFER_REQUEST_STATUS `json:"request_status,omitempty"`
}

func (t TransferInput) ToModel() (*domain.Transfer, error) {

	reqDate, err := jsonutils.ParseJSONTimeJST(time.DateTime, t.RequestDate)
	if err != nil {
		return nil, err
	}
	return &domain.Transfer{
		TransactionId:         t.TransactionId,
		UserId:                t.UserId,
		ProjectId:             t.ProjectId,
		TransferDate:          reqDate,
		TransferAmount:        t.RequestAmount,
		AccountNumber:         t.AccountNumber,
		BankCode:              t.BankCode,
		BankName:              t.BankName,
		BranchCode:            t.BranchCode,
		BranchName:            t.BranchName,
		Depositor:             t.Depositor,
		AccountTypeCode:       t.AccountTypeCode,
		TransactionDetailType: domain.WITHDRAWAL_REFUND,
	}, nil

}

type VaDepositInput struct {
	UserId            int64  `json:"user_id,omitempty"`
	ValueDate         string `json:"value_date,omitempty" swagger:"default(2021-06-01)"`
	VaId              string `json:"va_id,omitempty"`
	VaBranchCode      string `json:"va_branch_code" swagger:"default(502),required"`
	VaBranchNameKana  string `json:"va_branch_name_kana" swagger:"default(ｱｼﾞｻｲ),required"`
	VaAccountNameKana string `json:"va_account_name_kana" swagger:"default(ｶ)ﾊｺﾌﾞﾈｶﾊｺﾌﾞﾈ),required"`
	TransactionDate   string `json:"transaction_date" swagger:"default(2021-06-01),required"`
	RemitterNameKana  string `json:"remitter_name_kana" swagger:"default(ｺﾊﾞﾔｼ ｴﾘﾄ),required"`
	PaymentBankName   string `json:"payment_bank_name" swagger:"default(ｼﾞ-ｴﾑｵ-ｱｵｿﾞﾗﾈﾂﾄ)"`
	PaymentBranchName string `json:"payment_branch_name" swagger:"default(ﾆｼﾞ)"`
	PartnerName       string `json:"partner_name" swagger:"default(GMOあおぞらネット銀行)"`
	Remarks           string `json:"remarks"`
	ItemKey           string `json:"item_key" swagger:"required,desc(IDX 를 입력하세요.),default(1)"`
	VaAccountNumber   string `json:"va_account_number" swagger:"enum(5189966|5189974),desc(LOAN:5189966/EQUITY:5189974)"`
	DepositAmount     int64  `json:"deposit_amount" swagger:"required,desc(입금액을 입력하세요.)"`
	DepositDate       string `json:"deposit_date" swagger:"required,desc(입금일자를 입력하세요.),default(2021-01-01 00:00:03)"`
	UseAmount         int64  `json:"use_amount" swagger:"default(0)"`
	DepositStatus     string `json:"deposit_status" swagger:"enum(NOT_USE_DEPOSIT|FULL_USE_DEPOSIT|PART_USE_DEPOSIT) default(NOT_USE_DEPOSIT)"`
}

func (v VaDepositInput) ToModel() (*domain.VaDeposit, error) {
	if err := domain.ValidateVaDepositItemKey(v.ItemKey); err != nil {
		return nil, err
	}

	depositDate, err := jsonutils.ParseJSONTimeJST(time.DateTime, v.DepositDate)
	if err != nil {
		return nil, err
	}
	return &domain.VaDeposit{
		UserId:                v.UserId,
		VaId:                  v.VaId,
		TransactionDate:       v.TransactionDate,
		ValueDate:             v.ValueDate,
		VaBranchCode:          v.VaBranchCode,
		VaBranchNameKana:      v.VaBranchNameKana,
		VaAccountNameKana:     v.VaAccountNameKana,
		RemitterNameKana:      v.RemitterNameKana,
		PaymentBankName:       v.PaymentBankName,
		PaymentBranchName:     v.PaymentBranchName,
		PartnerName:           v.PartnerName,
		Remarks:               v.Remarks,
		ItemKey:               v.ItemKey,
		VaAccountNumber:       v.VaAccountNumber,
		DepositAmount:         v.DepositAmount,
		DepositDate:           depositDate,
		UseAmount:             v.UseAmount,
		DepositStatus:         domain.DEPOSIT_STATUS(v.DepositStatus),
		TransactionDetailType: domain.DEPOSIT_VIRTUAL_ACCOUNT,
	}, nil

}

type VaSearchInput struct {
	VaId              string `json:"va_id,omitempty" swagger:"desc(가상계좌 ID)"`
	VaStatusCode      string `json:"va_status_code,omitempty" swagger:"desc(가상계좌 상태 코드 1:중지 2:재개 3:삭제),enum(1|2|3)"`
	IssueDataFrom     string `json:"issue_data_from,omitempty" swagger:"desc(발행일 기준 조회 시작일)"`
	IssueDataTo       string `json:"issue_data_to,omitempty" swagger:"desc(발행일 기준 조회 종료일)"`
	VaContractAuthKey string `json:"va_contract_auth_key,omitempty" swagger:"desc(계약용 Key)"`
	// VaAccountNumber string `json:"va_account_number,omitempty" swagger:"desc(가상계좌 번호)"`
}

type BulkTransferCancelInput struct {
	ApplyNo              string `json:"apply_no,omitempty"`
	CancelTargetKeyClass string `json:"cancel_target_key_class,omitempty"`
}

type BulkTransferStatusInput struct {
	QueryKeyClass       string `json:"query_key_class,omitempty" swagger:"enum(1|2)"`
	DateFrom            string `json:"date_from,omitempty" swagger:"desc(조회 시작일)"`
	DateTo              string `json:"date_to,omitempty" swagger:"desc(조회 종료일)"`
	ApplyNo             string `json:"apply_no,omitempty" swagger:"desc(신청번호)"`
	DetailInfoNecessity string `json:"detail_info_necessity,omitempty" swagger:"desc(상세정보 필요여부),enum(Y|N)"`
}
