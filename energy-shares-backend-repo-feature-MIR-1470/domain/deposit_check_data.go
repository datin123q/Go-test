package domain

import (
	"context"
	"encoding/csv"
	"net/http"
	"reflect"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type DEPOSIT_CHECK_STATUS string

const (
	DEPOSIT_CHECK_STATUS_CODE                      = "DEPOSIT_CHECK_STATUS"
	OK_DEPOSIT                DEPOSIT_CHECK_STATUS = "OK_DEPOSIT"     // OK
	NO_DEPOSIT                DEPOSIT_CHECK_STATUS = "NO_DEPOSIT"     // NO
	CANCEL_DEPOSIT            DEPOSIT_CHECK_STATUS = "CANCEL_DEPOSIT" // Cancel
)

func (DepositCheckData) GetDepositCheckStatus(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(OK_DEPOSIT), Name: i18nutils.Translate("commoncode."+DEPOSIT_CHECK_STATUS_CODE+"."+string(OK_DEPOSIT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(NO_DEPOSIT), Name: i18nutils.Translate("commoncode."+DEPOSIT_CHECK_STATUS_CODE+"."+string(NO_DEPOSIT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CANCEL_DEPOSIT), Name: i18nutils.Translate("commoncode."+DEPOSIT_CHECK_STATUS_CODE+"."+string(CANCEL_DEPOSIT), loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: DEPOSIT_CHECK_STATUS_CODE,
		Codes:     codes,
	}
}

const DEPOSIT_CHECK_COLUM = "share.id as share_id,deposit_request.user_id as investor_id, user.name as investor_name,share.amount as investment_amount, " +
	"share.invest_date as investment_date," +
	"deposit_request.deposit_amount as  deposit_amount, deposit_request.deposit_date as deposit_date," +
	"deposit_request.deposit_check_status as deposit_check_status,user_virtual_account.va_account_number,deposit_request.created as created"

// 입금 확인 - DB
// 배치 프로세스 진행후 입금 내역과 비교한 금액
type DepositCheckData struct {
	ShareId            string               `json:"share_id" swagger:"desc(Share ID)"`
	InvestorId         int64                `json:"investor_id" swagger:"desc(투자자 ID)"`
	InvestorName       string               `json:"investor_name" swagger:"desc(투자자 이름)"`
	InvestmentAmount   int64                `json:"investment_amount" swagger:"desc(투자 확정 금액)"`
	InvestmentDate     jsonutils.JSONTime   `json:"investment_date" swagger:"desc(투자 일시)"`
	DepositAmount      int64                `json:"deposit_amount" swagger:"desc(입금 금액)"`
	DepositDate        jsonutils.JSONTime   `json:"deposit_date" swagger:"desc(입금 일자 )"`
	DepositCheckStatus DEPOSIT_CHECK_STATUS `json:"deposit_check_status" swagger:"desc(입금 상태)"`
	VaAccountNumber    string               `json:"va_account_number" swagger:"desc(계좌 번호)"`
}

func (DepositCheckData) GetAll(ctx context.Context, sortby, order string, offset, limit int, investorId, investorName, depositCheckStatus *string, startInvestmentAmount, endInvestmentAmount, startDespositAmount, endDespositAmount *int) (int64, []DepositCheckData, error) {

	q := factory.DB(ctx).Table("deposit_request").Select(DEPOSIT_CHECK_COLUM).
		Join("INNER", "share", "deposit_request.share_id = share.id").
		Join("INNER", "user", "deposit_request.user_id = user.id").
		Join("INNER", "user_virtual_account", "deposit_request.deposit_branch_code = user_virtual_account.va_branch_code AND deposit_request.deposit_account_number = user_virtual_account.va_account_number")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "user", "id", investorId); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "user", "name", investorName); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "deposit_request", "deposit_check_status", depositCheckStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenRange(q, "share.amount", startInvestmentAmount, endInvestmentAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SetBetweenRange(q, "deposit_request.deposit_amount", startDespositAmount, endDespositAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]DepositCheckData, 0, 0)

	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)

	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (DepositCheckData) GetById(ctx context.Context, id int64) (*DepositCheckData, error) {
	var d DepositCheckData
	q := factory.DB(ctx).Table("deposit_request").Select(DEPOSIT_CHECK_COLUM).
		Join("INNER", "share", "deposit_request.share_id = share.id").
		Join("INNER", "user", "deposit_request.user_id = user.id").
		Join("INNER", "user_virtual_account", "deposit_request.deposit_branch_code = user_virtual_account.va_branch_code AND deposit_request.deposit_account_number = user_virtual_account.va_account_number")

	has, err := q.Where("deposit_request.id = ?", id).Get(&d)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &d, nil
}

func (DepositCheckData) PrepareAndReturnCsv(items []DepositCheckData) [][]string {

	type deposit_check_data struct {
		InvestorId         int64  `json:"investment_id" i18n:"model.deposit_check_data.investor_id"`
		InvestorName       string `json:"investor_name" i18n:"model.deposit_check_data.investor_name"`
		VaAccountNumber    string `json:"va_account" i18n:"model.deposit_check_data.va_account_number"`
		InvestmentAmount   int64  `json:"investment_amount" i18n:"model.deposit_check_data.investment_amount"`
		InvestmentDate     string `json:"investment_date" i18n:"model.deposit_check_data.invest_date"`
		DepositAmount      int64  `json:"deposit_amount" i18n:"model.deposit_check_data.deposit_amount"`
		DepositDate        string `json:"deposit_date" i18n:"model.deposit_check_data.deposit_date"`
		DepositCheckStatus string `json:"deposit_check_status" i18n:"model.deposit_check_data.deposit_check_status"`
	}

	var subscriptions []deposit_check_data
	for _, item := range items {
		subs := deposit_check_data{
			InvestorId:         item.InvestorId,
			InvestorName:       item.InvestorName,
			VaAccountNumber:    item.VaAccountNumber,
			InvestmentAmount:   item.InvestmentAmount,
			InvestmentDate:     item.InvestmentDate.ToString(),
			DepositAmount:      item.DepositAmount,
			DepositDate:        item.DepositDate.ToString(),
			DepositCheckStatus: i18nutils.DefaultTranslate("commoncode." + DEPOSIT_CHECK_STATUS_CODE + "." + string(item.DepositCheckStatus)),
		}

		subscriptions = append(subscriptions, subs)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&deposit_check_data{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range subscriptions {

		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (DepositCheckData) DownloadCsv(ctx context.Context, w http.ResponseWriter, sortby, order string, investorId, investorName, depositCheckStatus *string, startInvestmentAmount, endInvestmentAmount, startDespositAmount, endDespositAmount *int) error {

	q := factory.DB(ctx).Table("deposit_request").Select(DEPOSIT_CHECK_COLUM).
		Join("INNER", "share", "deposit_request.share_id = share.id").
		Join("INNER", "user", "deposit_request.user_id = user.id").
		Join("INNER", "user_virtual_account", "deposit_request.deposit_branch_code = user_virtual_account.va_branch_code AND deposit_request.deposit_account_number = user_virtual_account.va_account_number")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "user", "id", investorId); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "user", "name", investorName); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "deposit_request", "deposit_check_status", depositCheckStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenRange(q, "share.amount", startInvestmentAmount, endInvestmentAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SetBetweenRange(q, "deposit_request.deposit_amount", startDespositAmount, endDespositAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]DepositCheckData, 0, 0)

	err := q.Find(&items)

	if err != nil {
		return err
	}
	records := DepositCheckData{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeader(w, "energyshares_deposit_check_data_list.csv")

	return csv.NewWriter(w).WriteAll(records)
}
