package domain

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/gmoaozorautils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type WITHDRAWAL_STATUS_TYPE string

const (
	WITHDRAWAL_STATUS_TYPE_CODE                         = "WITHDRAWAL_STATUS_TYPE"
	WITHDRAWAL_STATUS_COMPLETE   WITHDRAWAL_STATUS_TYPE = "COMPLETE"
	WITHDRAWAL_STATUS_INCOMPLETE WITHDRAWAL_STATUS_TYPE = "INCOMPLETE"
	WITHDRAWAL_STATUS_WAIT       WITHDRAWAL_STATUS_TYPE = "WAIT"
	WITHDRAWAL_STATUS_FAIL       WITHDRAWAL_STATUS_TYPE = "FAIL"
	WITHDRAWAL_STATUS_CANCEL     WITHDRAWAL_STATUS_TYPE = "CANCEL"
)

func (Withdrawal) GetWithdrawalStatusTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_STATUS_COMPLETE), Name: i18nutils.Translate("commoncode.WITHDRAWAL_STATUS_TYPE.WITHDRAWAL_STATUS_COMPLETE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_STATUS_INCOMPLETE), Name: i18nutils.Translate("commoncode.WITHDRAWAL_STATUS_TYPE.WITHDRAWAL_STATUS_INCOMPLETE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_STATUS_WAIT), Name: i18nutils.Translate("commoncode.WITHDRAWAL_STATUS_TYPE.WITHDRAWAL_STATUS_WAIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_STATUS_FAIL), Name: i18nutils.Translate("commoncode.WITHDRAWAL_STATUS_TYPE.WITHDRAWAL_STATUS_FAIL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WITHDRAWAL_STATUS_CANCEL), Name: i18nutils.Translate("commoncode.WITHDRAWAL_STATUS_TYPE.WITHDRAWAL_STATUS_CANCEL", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: WITHDRAWAL_STATUS_TYPE_CODE,
		Codes:     codes,
	}
}

type withdrawalStatusTypeDescription struct {
	english  string
	japanese string
}

var withdrawalStatusTypeDescriptions = map[WITHDRAWAL_STATUS_TYPE]withdrawalStatusTypeDescription{
	WITHDRAWAL_STATUS_COMPLETE:   {english: "Complete", japanese: "完了"},
	WITHDRAWAL_STATUS_INCOMPLETE: {english: "Incomplete", japanese: "未完了"},
	WITHDRAWAL_STATUS_WAIT:       {english: "Transfer Wait", japanese: "送金待ち"},
	WITHDRAWAL_STATUS_FAIL:       {english: "Transfer Fail", japanese: "送金失敗"},
	WITHDRAWAL_STATUS_CANCEL:     {english: "Cancel", japanese: "出金キャンセル"},
}

func (w WITHDRAWAL_STATUS_TYPE) Description() string {
	if desc, ok := withdrawalStatusTypeDescriptions[w]; ok {
		return desc.english
	}
	return "-"
}

func (w WITHDRAWAL_STATUS_TYPE) DescriptionJapanese() string {
	if desc, ok := withdrawalStatusTypeDescriptions[w]; ok {
		return desc.japanese
	}
	return "-"
}

type Withdrawal struct {
	Id           int64                  `json:"id"`                              // 出金番号
	InvestorId   string                 `json:"investor_id" xorm:"not null"`     // 投資家ID
	InvestorName string                 `json:"orderer" xorm:"orderer not null"` // 投資家名
	Amount       int64                  `json:"amount"`                          // 出金リクエスト金額
	Fee          int64                  `json:"fee"`                             // 振込手数料
	RequestDate  jsonutils.JSONTime     `json:"request_date"`                    // リクエスト日付
	CompleteDate jsonutils.JSONTime     `json:"complete_date"`                   // 完了日付
	CloseDate    jsonutils.JSONTime     `json:"close_date"`                      // 締切日付
	Status       WITHDRAWAL_STATUS_TYPE `json:"status"`                          // ステータス
	IsManual     bool                   `json:"is_manual"`                       // 手動出金かどうか（管理画面の出金依頼から作成した場合 true）
	Created      jsonutils.JSONTime     `json:"created" xorm:"created"`          // 作成日付
	Updated      jsonutils.JSONTime     `json:"updated" xorm:"updated"`          // 更新日付
	Deleted      jsonutils.JSONTime     `json:"deleted" xorm:"deleted"`          // 削除時にDeletedに時間が入力されると照会されません。1年後に削除できるようにします
}

type WithdrawalDetail struct {
	Withdrawal `xorm:"extends"`
	User       `xorm:"extends"`
}

func (w Withdrawal) TableName() string {
	return "withdrawal"
}

func (w Withdrawal) SelectAll() string {
	return w.TableName() + ".*"
}

// WithdrawalBankInfoErrorはMIR-1256の対応。送金データ(Transfer/TransferInput)を
// 組み立てる前に、銀行コード・支店コード・口座番号・預金種別が揃っていることを
// 確認する。問題が無ければnilを返す。domain/private_user.go(522-524行目付近)の
// 既存の口座登録済み判定(4項目)と揃えている。以前は複数の呼び出し元
// (apps/admin-api/controllers/withdrawal.go、domain/transfer.goのRequestFee)が
// それぞれ個別にBankCode/BranchCodeをnilチェック無しで参照しており、未登録の
// 会員が1人いるだけで一括送金・締め処理がパニックで全体停止し、user-api側の
// RequestFeeでは会員本人の出金リクエスト1件でuser-apiプロセス全体が落ちる
// 可能性があった。この判断を1箇所へ集約する。
func WithdrawalBankInfoError(bankCode, branchCode *string, bankAccount string, depositType DEPOSIT_TYPE) error {
	if bankCode == nil || *bankCode == "" ||
		branchCode == nil || *branchCode == "" ||
		bankAccount == "" || string(depositType) == "" {
		return errors.New("bank account information is not fully registered")
	}
	return nil
}

func (p *WithdrawalDetail) MarshalJSON() ([]byte, error) {

	return json.Marshal(map[string]interface{}{
		"id":            p.Withdrawal.Id,
		"investor_id":   p.Withdrawal.InvestorId,
		"orderer":       p.Withdrawal.InvestorName,
		"amount":        p.Withdrawal.Amount,
		"fee":           p.Withdrawal.Fee,
		"request_date":  p.Withdrawal.RequestDate,
		"complete_date": p.Withdrawal.CompleteDate,
		"close_date":    p.Withdrawal.CloseDate,
		"status":        p.Withdrawal.Status,
		"is_manual":     p.Withdrawal.IsManual,
		"bank_name":     p.User.BankName,
		"bank_code":     p.User.BankCode,
		"branch_name":   p.User.BranchName,
		"deposit_type":  p.User.DepositType,
		"depositor":     p.User.Depositor,
		"bank_account":  p.User.BankAccount,
		"user_type":     p.User.UserType,
		"branch_code":   p.User.BranchCode,
	})
}

// 完了 - Create
func (n *Withdrawal) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(n)
}

// 完了 - GetDetailAll
func (Withdrawal) GetAll(ctx context.Context, sortby, order string, offset, limit int, investorId *string, orderer *string, amountStart *int, amountEnd *int, requestStartDate string, requestEndDate string, completeStartDate string, completeEndDate string, status WITHDRAWAL_STATUS_TYPE) (int64, []WithdrawalDetail, error) {
	q := factory.DB(ctx).Table("withdrawal").
		Join("INNER", "user", "user.id = withdrawal.investor_id").
		Select("withdrawal.id, withdrawal.investor_id, withdrawal.orderer, withdrawal.amount, withdrawal.fee, withdrawal.request_date, withdrawal.complete_date, withdrawal.close_date, withdrawal.status, withdrawal.is_manual, " +
			"user.user_type, user.bank_code, user.branch_code, user.deposit_type, user.bank_account, user.depositor")

	if err := modelhelper.SetSortOrderForTable(q, "withdrawal", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "withdrawal", "investor_id", investorId); err != nil {
		return 0, nil, err
	}
	if err := modelhelper.SearchLikeQuery(q, "withdrawal", "orderer", orderer); err != nil {
		return 0, nil, err
	}
	if err := modelhelper.SetBetweenRange(q, "amount", amountStart, amountEnd); err != nil {
		return 0, nil, err
	}
	if err := modelhelper.SetBetweenDate(q, "request_date", requestStartDate, requestEndDate); err != nil {
		return 0, nil, err
	}
	if err := modelhelper.SetBetweenDate(q, "complete_date", completeStartDate, completeEndDate); err != nil {
		return 0, nil, err
	}
	withdrawalStatus := string(status)
	if err := modelhelper.SearchEqQuery(q, "withdrawal", "status", &withdrawalStatus); err != nil {
		return 0, nil, err
	}

	var items []WithdrawalDetail
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

// 完了 - GetDetailAll
func (Withdrawal) GetAllByUserId(ctx context.Context, userId int64, sortBy, order string, offset, limit int) (int64, []WithdrawalDetail, error) {
	q := factory.DB(ctx).Table("withdrawal").
		Join("INNER", "user", "user.id = withdrawal.investor_id").
		Select("withdrawal.id, withdrawal.investor_id, withdrawal.orderer, withdrawal.amount, withdrawal.fee, withdrawal.request_date, withdrawal.complete_date, withdrawal.close_date, withdrawal.status, withdrawal.created, withdrawal.updated, "+
			"user.id, user.user_type, user.bank_code, user.branch_code, user.deposit_type, user.bank_account, user.depositor").
		Where("withdrawal.investor_id = ?", userId)

	if err := modelhelper.SetSortOrderForTable(q, "withdrawal", sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []WithdrawalDetail
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

// 完了 - GetDetailById
func (WithdrawalDetail) GetById(ctx context.Context, id int64) (*WithdrawalDetail, error) {
	var n WithdrawalDetail
	if has, err := factory.DB(ctx).Table("withdrawal").
		Join("INNER", "user", "user.id = withdrawal.investor_id").
		Select("withdrawal.id, withdrawal.investor_id, withdrawal.amount, withdrawal.orderer, withdrawal.fee, withdrawal.status, withdrawal.is_manual, "+
			" withdrawal.request_date, withdrawal.complete_date, withdrawal.close_date, withdrawal.created, withdrawal.updated, "+
			" user.id, user.email, user.bank_name, user.bank_code, user.branch_code, user.branch_name, user.deposit_type, user.bank_account, user.user_type, user.depositor").
		Where("withdrawal.id = ?", id).Get(&n); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &n, nil
}

// withdrawalDetailByCloseIdSelectColumns はGetByCloseId・GetAllByCloseIdで共有する
// SELECT列(MIR-1181)。
const withdrawalDetailByCloseIdSelectColumns = "withdrawal.id, withdrawal.investor_id, withdrawal.amount, withdrawal.orderer, withdrawal.fee, withdrawal.request_date, withdrawal.complete_date, withdrawal.close_date, withdrawal.status, withdrawal.is_manual," +
	" user.bank_name, user.bank_code, user.branch_code, user.branch_name, user.deposit_type, user.bank_account, user.user_type, user.depositor"

// 完了 - GetDetailAll
func (WithdrawalDetail) GetByCloseId(ctx context.Context, id int64, sortBy, order string, offset, limit int) (int64, []WithdrawalDetail, error) {

	q := factory.DB(ctx).Table("withdrawal").
		Join("INNER", "user", "user.id = withdrawal.investor_id").
		Join("INNER", "withdrawal_close_mapping", "withdrawal_close_mapping.withdrawal_id = withdrawal.id").
		Select(withdrawalDetailByCloseIdSelectColumns).
		Where("withdrawal_close_mapping.close_id = ?", id)

	if err := modelhelper.SetSortOrderForTable(q, "withdrawal", sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []WithdrawalDetail
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

// GetAllByCloseId は締めIDに紐づく出金依頼を全件取得する。GetByCloseIdは一覧画面の
// ページングに使うため上限を渡す前提だが、一括送金は対象を1件も取りこぼしては
// いけないため、上限を決め打ちせず全件返す別メソッドとして分けた(MIR-1181。
// 以前はRequestTransferByClosedIdがGetByCloseIdへlimit=1000を渡しており、締めが
// 1000件を超えると超過分が静かに送金対象から漏れていた)。
func (WithdrawalDetail) GetAllByCloseId(ctx context.Context, id int64) ([]WithdrawalDetail, error) {
	q := factory.DB(ctx).Table("withdrawal").
		Join("INNER", "user", "user.id = withdrawal.investor_id").
		Join("INNER", "withdrawal_close_mapping", "withdrawal_close_mapping.withdrawal_id = withdrawal.id").
		Select(withdrawalDetailByCloseIdSelectColumns).
		Where("withdrawal_close_mapping.close_id = ?", id)

	var items []WithdrawalDetail
	if err := q.Find(&items); err != nil {
		return nil, err
	}
	return items, nil
}

// 完了 - GetAll
func (Withdrawal) CountIncomplete(ctx context.Context) (int64, error) {
	cnt, err := factory.DB(ctx).Where("status = ?", WITHDRAWAL_STATUS_INCOMPLETE).Count(&Withdrawal{})
	if err != nil {
		return 0, err
	}

	return cnt, err
}

// 完了 - GetSum
func (Withdrawal) SumIncompleteAmountByInvestorId(ctx context.Context, id int64) (int64, error) {
	withdrawalSum, err := factory.DB(ctx).Where("withdrawal.investor_id = ?", id).Where("status = ?", WITHDRAWAL_STATUS_INCOMPLETE).SumInt(&Withdrawal{}, "amount")
	if err != nil {
		return 0, err
	}

	return withdrawalSum, err
}

// 完了 - GetSum
func (Withdrawal) SumWaitAmountByInvestorId(ctx context.Context, id int64) (int64, error) {
	withdrawalSum, err := factory.DB(ctx).Where("withdrawal.investor_id = ?", id).Where("status = ?", WITHDRAWAL_STATUS_WAIT).SumInt(&Withdrawal{}, "amount")
	if err != nil {
		return 0, err
	}

	return withdrawalSum, err
}

// 完了 - GetSum
func (Withdrawal) SumFailAmountByInvestorId(ctx context.Context, id int64) (int64, error) {
	withdrawalSum, err := factory.DB(ctx).Where("withdrawal.investor_id = ?", id).Where("status = ?", WITHDRAWAL_STATUS_FAIL).SumInt(&Withdrawal{}, "amount")
	if err != nil {
		return 0, err
	}

	return withdrawalSum, err
}

// 完了 - GetAll
func (Withdrawal) GetIncompleteWithdrawal(ctx context.Context, offset, limit int) ([]Withdrawal, int64, error) {
	q := factory.DB(ctx).Table("withdrawal").Select("withdrawal.id, withdrawal.investor_id, withdrawal.orderer, withdrawal.amount, withdrawal.fee,"+
		"withdrawal.request_date, withdrawal.complete_date, withdrawal.status").Where("status = ?", WITHDRAWAL_STATUS_INCOMPLETE).Desc("id")

	withdrawalList := make([]Withdrawal, 0, 0)
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&withdrawalList)
	if err != nil {
		return nil, 0, err
	}

	return withdrawalList, totalCount, nil
}

// 完了 - Update
func (n *Withdrawal) Update(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).ID(n.Id).Update(n)
	return
}

// UpdateStatusIfCurrentはMIR-1250の対応。呼び出し元(ConvertWithdrawalStatus)
// はステータス変更の可否をメモリ上のwithdrawal.Status(取得済みの古い値)で
// 判定しているため、同じ出金への変更(例: キャンセル)が並行して届くと、
// 両方とも判定を通過してしまう(片方がコミットする前に、もう片方が変更前の
// ステータスを読んでいる)。影響行数を見るだけでは防げない。二重目のUPDATEも
// 「idが一致する1行」には当たるため、条件を付けなければ影響行数は1のまま
// になる。判定と更新の間の隙間そのものをなくすため、WHEREへ「更新前に
// 読んだステータス」を条件として足し、DB側で不一致なら0行更新にする
// (条件付き更新)。0行だった場合はErrWithdrawalStatusChangedを返すので、
// 呼び出し元は「既に処理済み」として扱う。
func (n *Withdrawal) UpdateStatusIfCurrent(ctx context.Context, expectedCurrent WITHDRAWAL_STATUS_TYPE) error {
	affected, err := factory.DB(ctx).Where("id = ? AND status = ?", n.Id, expectedCurrent).Update(n)
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.Wrapf(ErrWithdrawalStatusChanged, "withdrawal_id=%d, expected=%s", n.Id, expectedCurrent)
	}
	return nil
}

// 完了 - GetByTransactionId
func (n *Withdrawal) GetByTransactionId(ctx context.Context, id string) (*Withdrawal, error) {
	var w Withdrawal

	q := factory.DB(ctx).Table("withdrawal").
		Join("INNER", "withdrawal_transfer_mapping", "withdrawal_transfer_mapping.withdrawal_id = withdrawal.id").
		Select("withdrawal.id, withdrawal.investor_id, withdrawal.amount, withdrawal.orderer, withdrawal.fee, withdrawal.request_date, withdrawal.complete_date, withdrawal.close_date, withdrawal.status").
		Where("withdrawal_transfer_mapping.transaction_id = ?", id)

	if has, err := q.Get(&w); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &w, nil
}

// TODO
func (n *Withdrawal) BulkIncompleteToWait(ctx context.Context) (int64, error) {
	updated, err := factory.DB(ctx).Table("withdrawal").
		Where("status = ?", WITHDRAWAL_STATUS_INCOMPLETE).
		Update(&Withdrawal{
			Status:    WITHDRAWAL_STATUS_WAIT,
			CloseDate: jsonutils.JSONTime(time.Now()),
		})
	if err != nil {
		return 0, err
	}
	return updated, nil
}

// 完了 - GetSum
func (n *Withdrawal) SumIncompleteAmount(ctx context.Context) (int64, error) {
	sum, err := factory.DB(ctx).Table("withdrawal").Where("status = ?", WITHDRAWAL_STATUS_INCOMPLETE).SumInt(&Withdrawal{}, "amount")
	if err != nil {
		return 0, err
	}
	return sum, nil
}

// 完了 - GetSum
func (n *Withdrawal) SumIncompleteFee(ctx context.Context) (int64, error) {
	sum, err := factory.DB(ctx).Table("withdrawal").Where("status = ?", WITHDRAWAL_STATUS_INCOMPLETE).SumInt(&Withdrawal{}, "fee")
	if err != nil {
		return 0, err
	}
	return sum, nil
}

func (Withdrawal) convertedDepositType(depositType DEPOSIT_TYPE) string {

	var dt string
	switch depositType {
	case ORDINARY_ACCOUNT:
		dt = "1"
	case CHECKING_ACCOUNT:
		dt = "2"
	}

	return dt
}

func (Withdrawal) PrepareAndReturnCsv(items []WithdrawalDetail) [][]string {

	if items == nil {
		return nil
	}

	// 銀行番号、支店番号、預金種目、口座番号（6桁の場合は先頭に0）、預金者名、振込金額、投資家ID
	type WithdrawalCsv struct {
		BankCode     string `json:"bank_code" i18n:"model.withdrawal.bank_code"`               // 銀行番号
		BranchCode   string `json:"branch_code" i18n:"model.withdrawal.branch_code"`           // 支店番号
		DepositType  string `json:"deposit_type" i18n:"model.withdrawal.deposit_type"`         // 預金種目
		BankAccount  string `json:"bank_account" i18n:"model.withdrawal.bank_account"`         // 口座番号
		Depositor    string `json:"depositor" xorm:"not null" i18n:"model.withdrawal.orderer"` // 預金者名
		Amount       int64  `json:"amount" i18n:"model.withdrawal.amount"`                     // 振込金額
		EmptyColumn1 string `json: "-"`
		EmptyColumn2 string `json: "-"`
		InvestorId   string `json:"investor_id" xorm:"not null" i18n:"model.withdrawal.investor_id"` // 投資家ID
		RequestDate  string `json:"request_date" i18n:"model.withdrawal.request_date"`               // リクエスト日付
		CompleteDate string `json:"complete_date" i18n:"model.withdrawal.complete_date"`             // 完了日付
		// CloseDate    string `json:"close_date" i18n:"model.withdrawal.close_date"`                   // 마감 날짜
		Status string `json:"status" i18n:"model.withdrawal.status"` // ステータス
	}

	var withdrawals []WithdrawalCsv
	for _, item := range items {
		w := WithdrawalCsv{
			DepositType:  string(item.DepositType),
			Depositor:    gmoaozorautils.ZenkakuToHankaku(item.User.Depositor),
			Amount:       item.Amount,
			EmptyColumn1: "",
			EmptyColumn2: "",
			InvestorId:   item.InvestorId,
			RequestDate:  item.RequestDate.ToString(),
			CompleteDate: item.CompleteDate.ToString(),
			// CloseDate:    item.CloseDate.ToString(),
		}

		val, err := strconv.ParseInt(item.BankAccount, 10, 64)
		if err != nil {

			return nil
		}
		w.BankAccount = fmt.Sprintf("%07d", val) // 6桁以下は先頭に0で埋める

		if item.BankCode != nil {
			w.BankCode = *item.BankCode
		}
		if item.BranchCode != nil {
			w.BranchCode = *item.BranchCode
		}

		w.DepositType = Withdrawal{}.convertedDepositType(item.DepositType)

		if item.Status == WITHDRAWAL_STATUS_COMPLETE {
			w.Status = i18nutils.DefaultTranslate("commoncode.WITHDRAWAL_STATUS_TYPE.WITHDRAWAL_STATUS_COMPLETE")
		} else if item.Status == WITHDRAWAL_STATUS_INCOMPLETE {
			w.Status = i18nutils.DefaultTranslate("commoncode.WITHDRAWAL_STATUS_TYPE.WITHDRAWAL_STATUS_INCOMPLETE")
		} else if item.Status == WITHDRAWAL_STATUS_FAIL {
			w.Status = i18nutils.DefaultTranslate("commoncode.WITHDRAWAL_STATUS_TYPE.WITHDRAWAL_STATUS_FAIL")
		} else {
			w.Status = i18nutils.DefaultTranslate("commoncode.WITHDRAWAL_STATUS_TYPE.WITHDRAWAL_STATUS_WAIT")
		}
		withdrawals = append(withdrawals, w)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&WithdrawalCsv{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	var records [][]string
	records = append(records, cols)

	for _, v := range withdrawals {
		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (Withdrawal) DownloadCsv(ctx context.Context, w http.ResponseWriter, sortby, order string, investorId *string, orderer *string, amountStart *int, amountEnd *int, requestStartDate string, requestEndDate string, completeStartDate string, completeEndDate string, status WITHDRAWAL_STATUS_TYPE) error {

	q := factory.DB(ctx).Table("withdrawal").
		Join("INNER", "user", "user.id = withdrawal.investor_id").
		Select("withdrawal.id, withdrawal.investor_id, withdrawal.orderer, withdrawal.amount, withdrawal.fee, withdrawal.request_date, withdrawal.complete_date, withdrawal.close_date, withdrawal.status, withdrawal.is_manual, " +
			"user.user_type, user.bank_code, user.branch_code, user.deposit_type, user.bank_account, user.depositor")

	if err := modelhelper.SetSortOrderForTable(q, "withdrawal", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "withdrawal", "investor_id", investorId); err != nil {
		return err
	}
	if err := modelhelper.SearchLikeQuery(q, "withdrawal", "orderer", orderer); err != nil {
		return err
	}
	if err := modelhelper.SetBetweenRange(q, "amount", amountStart, amountEnd); err != nil {
		return err
	}
	if err := modelhelper.SetBetweenDate(q, "request_date", requestStartDate, requestEndDate); err != nil {
		return err
	}
	if err := modelhelper.SetBetweenDate(q, "complete_date", completeStartDate, completeEndDate); err != nil {
		return err
	}
	withdrawalType := string(status)
	if err := modelhelper.SearchEqQuery(q, "withdrawal", "status", &withdrawalType); err != nil {
		return err
	}

	var items []WithdrawalDetail
	_, err := q.FindAndCount(&items)
	if err != nil {
		return err
	}

	records := Withdrawal{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeaderWithShiftJIS(w, "energyshares_withdrawal_request_"+time.Now().Format("2006_01_02_15_04_05")+".csv")
	writer := csv.NewWriter(transform.NewWriter(w, japanese.ShiftJIS.NewEncoder()))
	writer.UseCRLF = true

	return writer.WriteAll(records)
}

func (Withdrawal) DownloadCsvClosed(ctx context.Context, w http.ResponseWriter, id int64) error {

	q := factory.DB(ctx).Table("withdrawal").
		Join("INNER", "user", "user.id = withdrawal.investor_id").
		Join("INNER", "withdrawal_close_mapping", "withdrawal_close_mapping.withdrawal_id = withdrawal.id").
		Select("withdrawal.id, withdrawal.investor_id, withdrawal.orderer, withdrawal.amount, withdrawal.fee, withdrawal.request_date, withdrawal.complete_date, withdrawal.close_date, withdrawal.status, "+
			"user.user_type, user.bank_code, user.branch_code, user.deposit_type, user.bank_account, user.depositor").
		Where("withdrawal_close_mapping.close_id = ?", id)

	var items []WithdrawalDetail
	_, err := q.FindAndCount(&items)
	if err != nil {
		return err
	}

	records := Withdrawal{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeaderWithShiftJIS(w, "energyshares_closed_withdrawal_request_"+time.Now().Format("2006_01_02_15_04_05")+".csv")
	writer := csv.NewWriter(transform.NewWriter(w, japanese.ShiftJIS.NewEncoder()))
	writer.UseCRLF = true

	return writer.WriteAll(records)
}
