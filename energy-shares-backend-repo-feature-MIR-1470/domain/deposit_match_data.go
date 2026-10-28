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
	"xorm.io/xorm"
)

// 과부족금 관리 DEPOSIT_MATCH_STATUS
// 초과
// 일치
// 일치
type DEPOSIT_MATCH_STATUS string

const (
	DEPOSIT_MATCH_STATUS_CODE                      = "DEPOSIT_MATCH_STATUS"
	OVER_DEPOSIT              DEPOSIT_MATCH_STATUS = "OVER_DEPOSIT" // 초과
	SAME_DEPOSIT              DEPOSIT_MATCH_STATUS = "SAME_DEPOSIT" // 일치
	LACK_DEPOSIT              DEPOSIT_MATCH_STATUS = "LACK_DEPOSIT" // 부족
)

func (DepositMatchData) GetDepositMatchStatus(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(OVER_DEPOSIT), Name: i18nutils.Translate("commoncode."+DEPOSIT_MATCH_STATUS_CODE+"."+string(OVER_DEPOSIT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SAME_DEPOSIT), Name: i18nutils.Translate("commoncode."+DEPOSIT_MATCH_STATUS_CODE+"."+string(SAME_DEPOSIT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(LACK_DEPOSIT), Name: i18nutils.Translate("commoncode."+DEPOSIT_MATCH_STATUS_CODE+"."+string(LACK_DEPOSIT), loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: DEPOSIT_MATCH_STATUS_CODE,
		Codes:     codes,
	}
}

const DEPOSIT_MATCH_DATA_VIEW_SQL = "CREATE VIEW OR REPLACE `energysharesdb`.`deposit_match_data` AS " +
	"(SELECT `energysharesdb`.`user`.`id` AS `user_id`," +
	"`energysharesdb`.`user`.`name` AS `investor_name`," +
	"SUM(`energysharesdb`.`deposit_request`.`request_amount`) AS `investment_amount`," +
	"SUM(`energysharesdb`.`deposit_request`.`deposit_amount`) AS `deposit_total_amount`," +
	"IF (SUM(`energysharesdb`.`deposit_request`.`request_amount`)  < SUM(`energysharesdb`.`deposit_request`.`deposit_amount`),'OVER_DEPOSIT', IF( SUM(`energysharesdb`.`deposit_request`.`request_amount`)  >SUM(`energysharesdb`.`deposit_request`.`deposit_amount`),'LACK_DEPOSIT','SAME_DEPOSIT')) AS `deposit_match_status`," +
	"(SUM(`energysharesdb`.`deposit_request`.`request_amount`) - SUM(`energysharesdb`.`deposit_request`.`deposit_amount`)) AS `deposit_match_amount`," +
	"(SUM(`energysharesdb`.`deposit_request`.`deposit_amount`) - SUM(`energysharesdb`.`deposit_request`.`request_amount`)) AS `lack_amount`," +
	"MAX(`energysharesdb`.`deposit_request`.`update_date`) AS `update_date`" +
	"FROM" +
	"((`energysharesdb`.`user`" +
	"LEFT JOIN `energysharesdb`.`deposit_request` ON ((`energysharesdb`.`user`.`id` = `energysharesdb`.`deposit_request`.`user_id`)))" +
	"LEFT JOIN `energysharesdb`.`va_deposit` ON ((`energysharesdb`.`user`.`id` = `energysharesdb`.`va_deposit`.`user_id`)))" +
	"WHERE" +
	"(`energysharesdb`.`user`.`user_type` <> 'ADMIN')" +
	"GROUP BY `energysharesdb`.`user`.`id`)"

// 과부족금
// 입금 확인 배치시 입금 결과와 비교 하여 생성
// 부족금 의뢰는 페이즈3에 진행 예정
// 배치에서 데이터를 입력함
// 사용자 별 부족금 체크 전체 투자금과 입금 금액을 비교
type DepositMatchData struct {
	UserId             int64                `json:"user_id" swagger:"desc(투자자 ID)"`
	InvestorName       string               `json:"investor_name" swagger:"desc(투자자 이름)"`
	DepositMatchStatus DEPOSIT_MATCH_STATUS `json:"deposit_match_status" xorm:"not null" swagger:"desc(과부족 상태),enum(OVER_DEPOSIT|SAME_DEPOSIT|LACK_DEPOSIT)"` // 과부족 상태
	DepositMatchAmount int64                `json:"deposit_match_amount" swagger:"desc(과부족 금액)"`                                                              // 과부족 금액
	InvestmentAmount   int64                `json:"investment_amount" swagger:"desc(투자요청 금액)"`                                                                // 투자요청 금액
	DepositTotalAmount int64                `json:"deposit_total_amount" swagger:"desc(입금액 합계)"`                                                              // 입금액 합계
	LackAmount         int64                `json:"lack_amount" swagger:"desc(미확인금액 +초과금액 -부족금액)"`                                                            // 미확인금액 부족금액
	UpdateDate         jsonutils.JSONTime   `json:"update_date" swagger:"desc(최근 갱신일)"`                                                                       //최근 갱신일 은행 조회 시간
}

func setQuery(q xorm.Interface, sortby, order string, investorId, investorName, depositMatchStatus *string, startDepositMatchAmount, endDepositMatchAmount, startDepositAmount, endDepositAmount *int) error {
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		return err
	}
	if err := modelhelper.SearchEqQuery(q, "deposit_match_data", "user_id", investorId); err != nil {
		return err
	}
	if err := modelhelper.SearchLikeQuery(q, "deposit_match_data", "investor_name", investorName); err != nil {
		return err
	}
	if err := modelhelper.SearchEqQuery(q, "deposit_match_data", "deposit_match_status", depositMatchStatus); err != nil {
		return err
	}

	if err := modelhelper.SetBetweenRange(q, "deposit_match_amount", startDepositMatchAmount, endDepositMatchAmount); err != nil {
		return err
	}
	if err := modelhelper.SetBetweenRange(q, "deposit_total_amount", startDepositAmount, endDepositAmount); err != nil {
		return err
	}
	return nil
}

func (DepositMatchData) GetAll(ctx context.Context, sortby, order string, offset, limit int, investorId, investorName, depositMatchStatus *string, startDepositMatchAmount, endDepositMatchAmount, startDepositAmount, endDepositAmount *int) (int64, []DepositMatchData, error) {
	q := factory.DB(ctx).Table("deposit_match_data")
	if err := setQuery(q, sortby, order, investorId, investorName, depositMatchStatus, startDepositMatchAmount, endDepositMatchAmount, startDepositAmount, endDepositAmount); err != nil {
		return 0, nil, err
	}

	items := make([]DepositMatchData, 0, 0)
	var totalCount int64

	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)

	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (DepositMatchData) GetById(ctx context.Context, id int64) (*DepositMatchData, error) {
	var d DepositMatchData
	q := factory.DB(ctx).Table("deposit_match_data")

	has, err := q.Where("user_id= ?", id).Get(&d)

	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	return &d, nil
}

func (d *DepositMatchData) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(d)
}

func (d *DepositMatchData) Update(ctx context.Context) error {
	q := factory.DB(ctx).UseBool("withdrawal_status")
	_, err := q.Update(d)
	return err
}

func (DepositMatchData) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&DepositMatchData{})
	return
}

func (DepositMatchData) PrepareAndReturnCsv(items []DepositMatchData) [][]string {

	type deposit_match_data struct {
		InvestorId         int64  `json:"investment_id" i18n:"model.deposit_match_data.investor_id"`
		InvestorName       string `json:"investor_name" i18n:"model.deposit_match_data.investor_name"`
		DepositMatchStatus string `json:"deposit_match_status" i18n:"model.deposit_match_data.deposit_match_status"`
		DepositMatchAmount int64  `json:"deposit_match_amount" i18n:"model.deposit_match_data.deposit_match_amount"`
		DepositTotalAmount int64  `json:"deposit_total_amount" i18n:"model.deposit_match_data.deposit_total_amount"`
		InvestmentAmount   int64  `json:"investment_amount" i18n:"model.deposit_match_data.investment_amount"`
		LackAmount         int64  `json:"lack_amount" i18n:"model.deposit_match_data.lack_amount"`
		UpdateDate         string `json:"update_date" i18n:"model.deposit_match_data.update_date"`
	}
	var depositMatchData []deposit_match_data
	for _, item := range items {
		data := deposit_match_data{
			InvestorId:         item.UserId,
			InvestorName:       item.InvestorName,
			DepositMatchStatus: i18nutils.DefaultTranslate("commoncode.DEPOSIT_MATCH_STATUS." + string(item.DepositMatchStatus)),
			DepositMatchAmount: item.DepositMatchAmount,
			DepositTotalAmount: item.DepositTotalAmount,
			InvestmentAmount:   item.InvestmentAmount,
			LackAmount:         item.LackAmount,
			UpdateDate:         item.UpdateDate.ToString(),
		}

		depositMatchData = append(depositMatchData, data)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&deposit_match_data{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range depositMatchData {

		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}
func (DepositMatchData) DownloadCsv(ctx context.Context, w http.ResponseWriter, sortby, order string, investorId, investorName, depositMatchStatus *string, startDepositMatchAmount, endDepositMatchAmount, startDepositAmount, endDepositAmount *int) error {

	q := factory.DB(ctx).Table("deposit_match_data").Select("deposit_match_data.*")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "deposit_match_data", "user_id", investorId); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "deposit_match_data", "investor_name", investorName); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "deposit_match_data", "deposit_match_status", depositMatchStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SetBetweenRange(q, "deposit_match_amount", startDepositMatchAmount, endDepositMatchAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenRange(q, "deposit_amount", startDepositAmount, endDepositAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]DepositMatchData, 0, 0)

	_, err := q.FindAndCount(&items)
	if err != nil {
		return err
	}
	records := DepositMatchData{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeader(w, "energyshares_deposit_match_data_list.csv")

	return csv.NewWriter(w).WriteAll(records)
}
