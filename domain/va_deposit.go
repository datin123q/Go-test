package domain

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

const VaDepositTransactionUniqueKeyName = "uq_va_deposit_account_item_key"

var (
	ErrVaDepositItemKeyRequired     = errors.New("va deposit item_key is required")
	ErrVaDepositTransactionConflict = errors.New("va deposit transaction conflicts with existing record")
)

func DepositAccountTypeCheck(depositType DEPOSIT_TYPE) string {
	switch depositType {
	case ORDINARY_ACCOUNT:
		return "1"
	case CHECKING_ACCOUNT:
		return "2"
	}
	return "3"
}

// Aozora仮想口座入金情報
type VaDeposit struct {
	Id                    int64                        `json:"id" xorm:"pk autoincr"`
	UserId                int64                        `json:"user_id" xorm:"user_id"`
	VaId                  string                       `json:"va_id" xorm:"va_id"`
	TransactionDate       string                       `json:"transaction_date" xorm:"transaction_date"`
	ValueDate             string                       `json:"value_date" xorm:"value_date"`
	VaBranchCode          string                       `json:"va_branch_code" xorm:"va_branch_code"`
	VaBranchNameKana      string                       `json:"va_branch_name_kana" `
	VaAccountNameKana     string                       `json:"vaAccountNameKana"`
	RemitterNameKana      string                       `json:"remitterNameKana"`
	PaymentBankName       string                       `json:"paymentBankName"`
	PaymentBranchName     string                       `json:"paymentBranchName"`
	PartnerName           string                       `json:"partnerName"`
	Remarks               string                       `json:"remarks,omitempty"`
	ItemKey               string                       `json:"itemKey"`
	VaAccountNumber       string                       `json:"va_account_number" xorm:"not null" swagger:"desc(口座番号)"` // 口座番号
	DepositAmount         int64                        `json:"deposit_amount" swagger:"desc(입금 금액)"`
	DepositDate           jsonutils.JSONTime           `json:"deposit_date" swagger:"desc(입금 일자 )"` // 입금 일자
	UseAmount             int64                        `json:"use_amount" swagger:"desc(투자 사용금)"`
	RefundAmount          int64                        `json:"refund_amount" swagger:"desc(강제출금)"`
	DepositStatus         DEPOSIT_STATUS               `json:"deposit_status" swagger:"desc(입금 상태: 현재 입금 금액기 투자에 사용되었는지 여부 )"`
	TransactionDetailType USER_TRANSACTION_DETAIL_TYPE `json:"transaction_type" swagger:"desc(거래종류)"`
	Created               jsonutils.JSONTime           `json:"created" xorm:"created" swagger:"desc(생성 시간)"`
	Updated               jsonutils.JSONTime           `json:"updated" xorm:"updated" swagger:"desc(업데이트 시간)"`
	Deleted               jsonutils.JSONTime           `json:"deleted" xorm:"deleted"`
}

func (VaDeposit) GetNotUseDeposit(ctx context.Context) (int64, []VaDeposit, error) {
	items := make([]VaDeposit, 0, 0)
	q := factory.DB(ctx).Table("va_deposit").Select("va_deposit.*")
	q = q.Where("deposit_status = ? OR deposit_status = ?", NOT_USE_DEPOSIT, PARTIAL_USE_DEPOSIT)

	var totalCount int64
	totalCount, err := q.FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (VaDeposit) ExistCheck(ctx context.Context, accountNumber, itemKey string) (bool, error) {
	var d VaDeposit
	cnt, err := factory.DB(ctx).Where("va_account_number = ? AND item_key = ?", accountNumber, itemKey).Limit(1).Count(&d)
	if cnt > 0 {
		return true, err
	}
	return false, nil
}

func (VaDeposit) GetByTransactionKey(ctx context.Context, accountNumber, itemKey string) (*VaDeposit, error) {
	var d VaDeposit
	has, err := factory.DB(ctx).
		Unscoped().
		Where("va_account_number = ? AND item_key = ?", accountNumber, itemKey).
		Get(&d)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	return &d, nil
}

// ValidateVaDepositItemKey は銀行取引を一意に識別するitem_keyが設定されていることを確認する。
func ValidateVaDepositItemKey(itemKey string) error {
	if strings.TrimSpace(itemKey) == "" {
		return ErrVaDepositItemKeyRequired
	}

	return nil
}

// CreateOrConfirmExisting はVA入金を登録する。
// 事前確認または同時登録による一意制約違反で同じ銀行取引が見つかった場合は、
// 既存レコードとの内容一致を確認し、created=false で登録済みとして返す。
func (d *VaDeposit) CreateOrConfirmExisting(ctx context.Context) (created bool, err error) {
	if err := ValidateVaDepositItemKey(d.ItemKey); err != nil {
		return false, err
	}

	has, err := (VaDeposit{}).ExistCheck(ctx, d.VaAccountNumber, d.ItemKey)
	if err != nil {
		return false, errors.Wrap(err, "failed to check existing va deposit")
	}
	if has {
		return false, d.confirmExistingTransaction(ctx)
	}

	if _, err := d.Create(ctx); err != nil {
		if !IsVaDepositTransactionDuplicateError(err) {
			return false, errors.Wrap(err, "failed to create va deposit")
		}

		return false, d.confirmExistingTransaction(ctx)
	}

	return true, nil
}

// IsVaDepositTransactionDuplicateError は、VA入金取引キー用の一意制約違反だけを判定する。
// 他の制約で発生した1062を登録済みとして扱わないため、制約名まで確認する。
func IsVaDepositTransactionDuplicateError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if !stderrors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return false
	}

	const keyMarker = "for key '"
	markerIndex := strings.LastIndex(mysqlErr.Message, keyMarker)
	if markerIndex < 0 {
		return false
	}

	keyPart := mysqlErr.Message[markerIndex+len(keyMarker):]
	quoteIndex := strings.IndexByte(keyPart, '\'')
	if quoteIndex < 0 {
		return false
	}

	qualifiedKeyName := strings.Trim(keyPart[:quoteIndex], "`")
	if dotIndex := strings.LastIndexByte(qualifiedKeyName, '.'); dotIndex >= 0 {
		qualifiedKeyName = qualifiedKeyName[dotIndex+1:]
	}

	return strings.Trim(qualifiedKeyName, "`") == VaDepositTransactionUniqueKeyName
}

func (d *VaDeposit) confirmExistingTransaction(ctx context.Context) error {
	existing, err := (VaDeposit{}).GetByTransactionKey(ctx, d.VaAccountNumber, d.ItemKey)
	if err != nil {
		return errors.Wrap(err, "failed to get existing va deposit")
	}
	if existing == nil {
		return errors.New("existing va deposit not found after duplicate detection")
	}

	mismatchedFields := d.transactionMismatchedFields(existing)
	if len(mismatchedFields) > 0 {
		return errors.Wrapf(
			ErrVaDepositTransactionConflict,
			"existing_id=%d mismatched_fields=%s",
			existing.Id,
			strings.Join(mismatchedFields, ","),
		)
	}

	return nil
}

func (d *VaDeposit) transactionMismatchedFields(existing *VaDeposit) []string {
	mismatchedFields := make([]string, 0, 6)
	if d.VaAccountNumber != existing.VaAccountNumber {
		mismatchedFields = append(mismatchedFields, "va_account_number")
	}
	if d.ItemKey != existing.ItemKey {
		mismatchedFields = append(mismatchedFields, "item_key")
	}
	if d.VaId != existing.VaId {
		mismatchedFields = append(mismatchedFields, "va_id")
	}
	if d.DepositAmount != existing.DepositAmount {
		mismatchedFields = append(mismatchedFields, "deposit_amount")
	}
	if d.TransactionDate != existing.TransactionDate {
		mismatchedFields = append(mismatchedFields, "transaction_date")
	}
	if d.ValueDate != existing.ValueDate {
		mismatchedFields = append(mismatchedFields, "value_date")
	}

	return mismatchedFields
}

func (d *VaDeposit) Create(ctx context.Context) (int64, error) {
	if err := ValidateVaDepositItemKey(d.ItemKey); err != nil {
		return 0, err
	}

	return factory.DB(ctx).Insert(d)
}

func (d *VaDeposit) AvailableAmount() int64 {
	return d.DepositAmount - (d.UseAmount + d.RefundAmount)
}

func (d *VaDeposit) Update(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(d.Id).Update(d)
	return err
}

func (VaDeposit) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&VaDeposit{})
	return
}

func (d *VaDeposit) CheckUseAmount(ctx context.Context) error {
	if d.DepositAmount == d.UseAmount {
		d.DepositStatus = FULL_USE_DEPOSIT
	} else if d.UseAmount == 0 {
		d.DepositStatus = NOT_USE_DEPOSIT
	} else {
		d.DepositStatus = PARTIAL_USE_DEPOSIT
	}
	return d.Update(ctx)
}

func (d *VaDeposit) DepositEquityProcess(ctx context.Context) error {
	account, err := UserAccount{}.GetByUserId(ctx, d.UserId)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get user account by userId %d", d.UserId))
	}

	_, err = account.DepositBalance(ctx, d.AvailableAmount())
	if err != nil {
		return errors.Wrap(err, "fail Deposit To User Account")
	}

	d.UseAmount = d.DepositAmount
	d.TransactionDetailType = DEPOSIT_VIRTUAL_ACCOUNT
	d.DepositStatus = FULL_USE_DEPOSIT

	if err := d.Update(ctx); err != nil {
		return errors.Wrap(err, "failed to update VaDeposit")
	}

	// 入金履歴を保存
	now := time.Now()
	if err = (UserTransaction{}).DepositUserTransaction(ctx, d.UserId, d.DepositAmount, 0, EQUITY, d.VaAccountNumber, DEPOSIT_VIRTUAL_ACCOUNT, jsonutils.JSONTime(now)); err != nil {
		return errors.Wrap(err, "failed to create user transaction")
	}

	err = CreateCustomerLedgerDepositEvent(ctx, d.UserId, account.UserName, d.DepositAmount, now)
	if err != nil {
		return errors.Wrap(err, "failed to create customer ledger deposit event")
	}

	drs, err := DepositRequest{}.GetNoDepositRequestByUserId(ctx, d.UserId)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get deposit request by userId %d", d.UserId))
	}

	for _, dr := range drs {
		logrus.WithFields(logrus.Fields{
			"share_id": dr.ShareId,
			"status":   dr.DepositCheckStatus,
			"amount":   dr.DepositAmount,
			"need":     dr.NeedAmount(),
		}).Info("DepositEquityProcess: Start UserAccount.DepositCheckProcess()")

		if err := account.DepositCheckProcess(ctx, &dr, NOT_IMMEDIATE); err != nil {
			return errors.Wrap(err, "fail Deposit Check process share_id:"+dr.ShareId)
		}

		logrus.WithFields(logrus.Fields{
			"share_id": dr.ShareId,
			"status":   dr.DepositCheckStatus,
			"amount":   dr.DepositAmount,
			"need":     dr.NeedAmount(),
		}).Info("DepositEquityProcess: End UserAccount.DepositCheckProcess()")
	}

	logrus.WithFields(logrus.Fields{
		"user_id":                   d.UserId,
		"deposit_account_number":    d.VaAccountNumber,
		"deposit_account_name_kana": d.VaAccountNameKana,
		"deposit_date":              d.DepositDate,
		"deposit_amount":            d.DepositAmount,
	}).Info("DepositEquityProcess: Finish")

	// 入金投資金額が不足している場合
	// 仮想口座のお金が余っている場合はオーバー
	return nil
}
