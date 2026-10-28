package domain

import (
	"context"
	"encoding/csv"
	"net/http"
	"reflect"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

const TRUST_INFO_QUERY string = "user_transaction.id as transaction_id, user_transaction.user_id as investor_id, user_transaction.amount as increase_amount, " +
	"user_transaction.transaction_date as deposit_date, user_transaction.project_type, user.name as investor_name, user.email"

type TrustInfo struct {
	TransactionId  int64              `json:"transaction_id" i18n:"model.trust_info.transaction_id"`
	DepositDate    jsonutils.JSONTime `json:"deposit_date" i18n:"model.trust_info.deposit_date"`
	InvestorId     int64              `json:"investor_id" i18n:"model.trust_info.investor_id"`
	InvestorName   string             `json:"investor_name" i18n:"model.trust_info.investor_name"`
	IncreaseAmount int64              `json:"increase_amount" i18n:"model.trust_info.deposit_amount"`
}

func (TrustInfo) DownloadCsv(ctx context.Context, w http.ResponseWriter, sortBy, order string, userId, name, email *string, startDate, endDate string) error {

	q := factory.DB(ctx).Table("user_transaction")
	q.Select(TRUST_INFO_QUERY).Join("INNER", "user", "user.id = user_transaction.user_id")

	// MIR-1113: 以前は無条件で "user_transaction."+sortBy と連結しており、
	// sortBy が空のとき "user_transaction." という不正な列名になっていた
	// (列名検証の追加後はこれが弾かれ、並び替えが無言で効かなくなる)。
	// QualifySortColumns は空の sortStr をそのまま返すため、
	// SetSortOrderForTable に置き換えて空のときは何も付けないようにする。
	if err := modelhelper.SetSortOrderForTable(q, "user_transaction", sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "user_transaction", "user_id", userId); err != nil {
		return err
	}
	if err := modelhelper.SearchLikeQuery(q, "user", "name", name); err != nil {
		return err
	}
	if err := modelhelper.SearchLikeQuery(q, "user", "email", email); err != nil {
		return err
	}
	if err := modelhelper.SetBetweenDate(q, "transaction_date", startDate, endDate); err != nil {
		return err
	}

	var items []TrustInfo

	_, err := q.Where("user_transaction.project_type = 'EQUITY' AND user_transaction.transaction_type = 'DEPOSIT_TRANSACTION'").FindAndCount(&items)
	if err != nil {
		return err
	}

	records := TrustInfo{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeader(w, "trust_info_list_"+time.Now().Format("2006_01_02_15_04_05")+".csv")

	return csv.NewWriter(w).WriteAll(records)
}

func (TrustInfo) PrepareAndReturnCsv(items []TrustInfo) [][]string {

	cols := factory.ReflectJsonFields(reflect.ValueOf(&TrustInfo{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range items {

		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}
