package domain

import (
	"context"
	"encoding/csv"
	"net/http"
	"reflect"
	"strconv"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type USER_TRANSACTION_TYPE string

const (
	USER_TRANSACTION_TYPE_NAME string                = "USER_TRANSACTION_TYPE"
	DEPOSIT_TRANSACTION        USER_TRANSACTION_TYPE = "DEPOSIT_TRANSACTION"
	WITHDRAWAL_TRANSACTION     USER_TRANSACTION_TYPE = "WITHDRAWAL_TRANSACTION"
)

type TRANSACTION_STATUS string

const (
	TRANSACTION_OK   TRANSACTION_STATUS = "TRANSACTION_OK"
	TRANSACTION_NO   TRANSACTION_STATUS = "TRANSACTION_NO"
	TRANSACTION_FAIL TRANSACTION_STATUS = "TRANSACTION_FAIL"
)

func (UserTransaction) GetTransactionTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)

	codes = append(codes, commoncode.Code{Code: string(DEPOSIT_TRANSACTION), Name: i18nutils.Translate("commoncode."+string(USER_TRANSACTION_TYPE_NAME)+"."+string(DEPOSIT_TRANSACTION), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_TRANSACTION), Name: i18nutils.Translate("commoncode."+string(USER_TRANSACTION_TYPE_NAME)+"."+string(WITHDRAWAL_TRANSACTION), loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: USER_TRANSACTION_TYPE_NAME,
		Codes:     codes,
	}
}

type USER_TRANSACTION_DETAIL_TYPE string

const (
	USER_TRANSACTION_DETAIL_TYPE_NAME string                       = "USER_TRANSACTION_DETAIL_TYPE"
	DEPOSIT_VIRTUAL_ACCOUNT           USER_TRANSACTION_DETAIL_TYPE = "DEPOSIT_VIRTUAL_ACCOUNT"    // 仮想口座入金
	DEPOSIT_DIVIDEND                  USER_TRANSACTION_DETAIL_TYPE = "DEPOSIT_DIVIDEND"           // 投資配当金
	DEPOSIT_REPAYMENT                 USER_TRANSACTION_DETAIL_TYPE = "DEPOSIT_REPAYMENT"          // 投資償還金
	WITHDRAWAL_INVESTMENT             USER_TRANSACTION_DETAIL_TYPE = "WITHDRAWAL_INVESTMENT"      // 出資金
	WITHDRAWAL_INVESTOR               USER_TRANSACTION_DETAIL_TYPE = "WITHDRAWAL_INVESTOR"        // 投資家出金（出金リクエスト）
	WITHDRAWAL_REFUND                 USER_TRANSACTION_DETAIL_TYPE = "WITHDRAWAL_REFUND"          // 強制出金
	DEPOSIT_WITHDRAWAL_CANCEL         USER_TRANSACTION_DETAIL_TYPE = "DEPOSIT_WITHDRAWAL_CANCEL"  // 投資金額出金キャンセル
	WITHDRAWAL_INVESTOR_CANCEL        USER_TRANSACTION_DETAIL_TYPE = "WITHDRAWAL_INVESTOR_CANCEL" // 投資家出金キャンセル（出金リクエストキャンセル）
)

func (UserTransaction) GetTransactionDetailTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)

	codes = append(codes, commoncode.Code{Code: string(DEPOSIT_VIRTUAL_ACCOUNT), Name: i18nutils.Translate("commoncode."+string(USER_TRANSACTION_DETAIL_TYPE_NAME)+"."+string(DEPOSIT_VIRTUAL_ACCOUNT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(DEPOSIT_DIVIDEND), Name: i18nutils.Translate("commoncode."+string(USER_TRANSACTION_DETAIL_TYPE_NAME)+"."+string(DEPOSIT_DIVIDEND), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(DEPOSIT_REPAYMENT), Name: i18nutils.Translate("commoncode."+string(USER_TRANSACTION_DETAIL_TYPE_NAME)+"."+string(DEPOSIT_REPAYMENT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_INVESTMENT), Name: i18nutils.Translate("commoncode."+string(USER_TRANSACTION_DETAIL_TYPE_NAME)+"."+string(WITHDRAWAL_INVESTMENT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_INVESTOR), Name: i18nutils.Translate("commoncode."+string(USER_TRANSACTION_DETAIL_TYPE_NAME)+"."+string(WITHDRAWAL_INVESTOR), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_REFUND), Name: i18nutils.Translate("commoncode."+string(USER_TRANSACTION_DETAIL_TYPE_NAME)+"."+string(WITHDRAWAL_REFUND), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(DEPOSIT_WITHDRAWAL_CANCEL), Name: i18nutils.Translate("commoncode."+string(USER_TRANSACTION_DETAIL_TYPE_NAME)+"."+string(DEPOSIT_WITHDRAWAL_CANCEL), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_INVESTOR_CANCEL), Name: i18nutils.Translate("commoncode."+string(USER_TRANSACTION_DETAIL_TYPE_NAME)+"."+string(WITHDRAWAL_INVESTOR_CANCEL), loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: USER_TRANSACTION_DETAIL_TYPE_NAME,
		Codes:     codes,
	}
}

// ユーザー取引履歴
type UserTransaction struct {
	Id                    int64                        `json:"id" xorm:"pk autoincr"`
	UserId                int64                        `json:"user_id" swagger:"desc(사용자 ID)"`
	UserType              USER_TYPE                    `json:"user_type" xorm:"<-" swagger:"desc(사용자 type)"`
	UserName              string                       `json:"user_name" xorm:"<-" swagger:"desc(사용자 이름)"`
	ProjectId             int64                        `json:"project_id" xorm:"project_id" swagger:"desc(프로젝트 ID - 상품 ID)"`
	AccountNumber         string                       `json:"account_number" swagger:"desc(계좌 번호)"`
	ProjectType           PROJECT_TYPE                 `json:"project_type" swagger:"desc(프로젝트 종류),enum(LOAN|EQUITY)"`
	Amount                int64                        `json:"amount" swagger:"desc(금액)"`
	Balance               int64                        `json:"balance" swagger:"desc(예치금잔액)"`
	TransactionType       USER_TRANSACTION_TYPE        `json:"transaction_type" swagger:"desc(입금/출금),enum(DEPOSIT_TRANSACTION|WITHDRAWAL_TRANSACTION)"`
	TransactionDetailType USER_TRANSACTION_DETAIL_TYPE `json:"transaction_detail_type" swagger:"desc(종류),enum(DEPOSIT_VIRTUAL_ACCOUNT|DEPOSIT_DIVIDEND|DEPOSIT_REPAYMENT|WITHDRAWAL_INVESTMENT|WITHDRAWAL_INVESTOR_REQUEST)"`
	TransactionDate       jsonutils.JSONTime           `json:"transaction_date" swagger:"desc(거래일시)"`
	Created               jsonutils.JSONTime           `json:"-" xorm:"created"`
	Updated               jsonutils.JSONTime           `json:"-" xorm:"updated"`
}

// ユーザー取引履歴
type UserTransactionDetail struct {
	Id                    int64                        `json:"id" xorm:"id"`
	UserId                int64                        `json:"user_id" swagger:"desc(사용자 ID)"`
	ProjectId             int64                        `json:"project_id" xorm:"project_id" swagger:"desc(프로젝트 ID - 상품 ID)"`
	ProjectName           string                       `json:"project_name" xorm:"<-" swagger:"desc(프로젝 이름)"`
	AccountNumber         string                       `json:"account_number" swagger:"desc(계좌 번호)"`
	ProjectType           PROJECT_TYPE                 `json:"project_type" swagger:"desc(프로젝트 종류),enum(LOAN|EQUITY)"`
	Amount                int64                        `json:"amount" swagger:"desc(금액)"`
	Balance               int64                        `json:"balance" swagger:"desc(예치금잔액)"`
	TransactionType       USER_TRANSACTION_TYPE        `json:"transaction_type" swagger:"desc(입금/출금),enum(DEPOSIT_TRANSACTION|WITHDRAWAL_TRANSACTION)"`
	TransactionDetailType USER_TRANSACTION_DETAIL_TYPE `json:"transaction_detail_type" swagger:"desc(종류),enum(DEPOSIT_VIRTUAL_ACCOUNT|DEPOSIT_DIVIDEND|DEPOSIT_REPAYMENT|WITHDRAWAL_INVESTMENT|WITHDRAWAL_INVESTOR_REQUEST)"`
	TransactionDate       jsonutils.JSONTime           `json:"transaction_date" swagger:"desc(거래일시)"`
	Created               jsonutils.JSONTime           `json:"-" xorm:"created"`
	Updated               jsonutils.JSONTime           `json:"-" xorm:"updated"`
}

// DONE: ユーザー取引履歴モデル実装
func (UserTransaction) GetAll(ctx context.Context, sortby, order string, offset, limit int, userId, transactionType, userName *string, startAmount, endAmount *int) (int64, []UserTransaction, error) {
	// 取引履歴を追加する。

	query := factory.DB(ctx).Table("user_transaction").Select("user.name as user_name,user.user_type as user_type, user_transaction.*").
		Join("LEFT", "user", "user_transaction.user_id = user.id")

	if err := modelhelper.SetSortOrderForTable(query, "user_transaction", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user_transaction", "user_id", userId); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user_transaction", "transaction_type", transactionType); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user", "name", userName); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SetBetweenRange(query, "amount", startAmount, endAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []UserTransaction

	totalCount, err := query.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (UserTransaction) GetByUserId(ctx context.Context, userId int64, sortBy, order string, offset, limit int, projectType *string) (int64, []UserTransactionDetail, error) {
	// 取引履歴を追加する。
	query := factory.DB(ctx).Table("user_transaction")
	query.Select("user_transaction.*,project.name as project_name").
		Join("LEFT", "project", "user_transaction.project_id = project.id").
		Where("user_id = ?", userId)

	if err := modelhelper.SetSortOrderForTable(query, "user_transaction", sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user_transaction", "project_type", projectType); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []UserTransactionDetail
	var totalCount int64
	totalCount, err := query.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (UserTransaction) GetByUserIdAndType(ctx context.Context, userId int64, projectType PROJECT_TYPE, sortby, order string, offset, limit int) (int64, []UserTransaction, error) {
	q := factory.DB(ctx).Table("user_transaction")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []UserTransaction
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).Where("user_id = ? AND project_type = ?", userId, projectType).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	if items == nil {
		items = make([]UserTransaction, 0, 0)
	}

	return totalCount, items, nil
}

func (UserTransaction) GetAllDownloadCsv(ctx context.Context, w http.ResponseWriter, sortby, order string, offset, limit int, userId, transactionType, userName *string, startAmount, endAmount *int) error {
	query := factory.DB(ctx).Table("user_transaction").Select("user.name as user_name, user_transaction.* ").
		Join("LEFT", "user", "user_transaction.user_id = user.id")

	if err := modelhelper.SetSortOrderForTable(query, "user_transaction", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user_transaction", "user_id", userId); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user_transaction", "transaction_type", transactionType); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user", "name", userName); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SetBetweenRange(query, "amount", startAmount, endAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []UserTransaction

	_, err := query.FindAndCount(&items)
	if err != nil {
		return err
	}

	records := UserTransaction{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeader(w, "user_transaction.csv")

	return csv.NewWriter(w).WriteAll(records)
}

func (UserTransaction) PrepareAndReturnCsv(items []UserTransaction) [][]string {
	type usertransactioncsv struct {
		UserId                string `json:"user_id" i18n:"model.deposit_match_data.investor_id"`                                          // 投資家ID
		UserName              string `json:"user_name" i18n:"model.user_transaction.invest_name"`                                          // 投資家名
		AccountNumber         string `json:"account_number" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_BANK_ACCOUNT"` // 口座番号
		ProjectType           string `json:"project_type" i18n:"model.project.project_type"`                                               // 投資タイプ
		TransactionType       string `json:"transaction_type" i18n:"model.user_transaction.deposit_withdrawal"`                            // 입금/출금
		TransactionDetailType string `json:"transaction_detail_type" i18n:"model.user_transaction.kind"`                                   // 種類
		Amount                string `json:"amount" i18n:"model.user_transaction.amount"`                                                  // 金額
		TransactionDate       string `json:"transaction_date" i18n:"model.user_transaction.date"`                                          // 日時
	}

	var usertransactions []usertransactioncsv
	for _, item := range items {

		us := usertransactioncsv{
			UserId:        strconv.FormatInt(item.UserId, 10),
			UserName:      item.UserName,
			AccountNumber: item.AccountNumber,
			Amount:        strconv.FormatInt(item.Amount, 10),
		}
		//
		if value, ok := reflect.ValueOf(&item.ProjectType).Elem().Interface().(PROJECT_TYPE); ok {
			us.ProjectType = "commoncode.PROJECT_TYPE." + string(value)
		}
		if value, ok := reflect.ValueOf(&item.TransactionType).Elem().Interface().(USER_TRANSACTION_TYPE); ok {
			us.TransactionType = "commoncode.USER_TRANSACTION_TYPE." + string(value)
		}
		if value, ok := reflect.ValueOf(&item.TransactionDetailType).Elem().Interface().(USER_TRANSACTION_DETAIL_TYPE); ok {
			us.TransactionDetailType = "commoncode.USER_TRANSACTION_DETAIL_TYPE." + string(value)
		}
		if item.TransactionDate.Unix() > 0 {
			us.TransactionDate = item.TransactionDate.ToString()
		}

		usertransactions = append(usertransactions, us)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&usertransactioncsv{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range usertransactions {
		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (u *UserTransaction) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(u)
}

func (UserTransaction) DepositUserTransaction(ctx context.Context, userId, amount, projectId int64, projectType PROJECT_TYPE, accountNumber string, transactionDetailType USER_TRANSACTION_DETAIL_TYPE, transactionDate jsonutils.JSONTime) error {
	// 入金
	balance, err := CalculateCustomerLedgerBalance(ctx, userId)
	if err != nil {
		return errors.Wrap(err, "failed to calculate customer ledger balance")
	}
	ut := UserTransaction{
		UserId:                userId,
		ProjectId:             projectId,
		AccountNumber:         accountNumber, // 입금일 때는 필수, 출금일 때 없어도 됨
		ProjectType:           projectType,
		Amount:                amount,
		Balance:               balance,
		TransactionType:       DEPOSIT_TRANSACTION,
		TransactionDetailType: transactionDetailType,
		TransactionDate:       transactionDate,
	}
	if _, err := ut.Create(ctx); err != nil {
		return err
	}
	return nil
}

func (UserTransaction) WithdrawalUserTransaction(ctx context.Context, userId, amount, projectId int64, projectType PROJECT_TYPE, accountNumber string, transactionDetailType USER_TRANSACTION_DETAIL_TYPE, transactionDate jsonutils.JSONTime) error {
	// 出金
	balance, err := CalculateCustomerLedgerBalance(ctx, userId)
	if err != nil {
		return errors.Wrap(err, "failed to calculate customer ledger balance")
	}
	return UserTransaction{}.WithdrawalUserTransactionWithBalance(ctx, userId, balance, amount, projectId, projectType, accountNumber, transactionDetailType, transactionDate)
}

// WithdrawalUserTransactionWithBalance は WithdrawalUserTransaction と同じ処理だが、
// 呼び出し元が既に計算済みのbalanceをそのまま使う(MIR-1202)。出資成立の一括処理
// (repayment_dividend_ucase.go)では、この関数を呼ぶ直前に同じユーザーのbalanceを
// 既に算出済みで、その間にuser_asset_viewへ影響する書き込みは無いため、
// CalculateCustomerLedgerBalanceをもう一度呼ぶのは同じ結果を得るためだけの
// 二重計算になっていた(1口あたりuser_asset_viewの実測が約58msかかり、1口259msの
// 約22%を占めていた)。balanceの計算をこの関数の外へ出す以上、呼び出し元は
// CalculateCustomerLedgerBalanceの後にuser_asset_viewへ影響する処理を挟まないこと。
func (UserTransaction) WithdrawalUserTransactionWithBalance(ctx context.Context, userId, balance, amount, projectId int64, projectType PROJECT_TYPE, accountNumber string, transactionDetailType USER_TRANSACTION_DETAIL_TYPE, transactionDate jsonutils.JSONTime) error {
	ut := UserTransaction{
		UserId:                userId,
		ProjectId:             projectId,
		AccountNumber:         accountNumber,
		ProjectType:           projectType,
		Amount:                amount * -1,
		Balance:               balance,
		TransactionType:       WITHDRAWAL_TRANSACTION,
		TransactionDetailType: transactionDetailType,
		TransactionDate:       transactionDate,
	}
	if _, err := ut.Create(ctx); err != nil {
		return errors.Wrap(err, "fail to create user transaction")
	}
	return nil
}
