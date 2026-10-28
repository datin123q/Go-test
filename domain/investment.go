package domain

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

const (
	INVESTOR_PROJECT_STRINGS string = `
        project.id,
        project.spv_id,
        project.name,
        project.start_at,
        project.end_at,
        project.project_status,
        project.capacity,
        project.main_image_url,
        project.project_report_date
    `

	INVESTOR_USER_STRINGS string = `
        user.id AS id,
        user.investor_type AS investor_type,
        user.name AS user_name,
        user.email AS email,
        user.mobile_phone AS mobile_phone,
        user.user_type AS user_type,
        user.address AS address,
        user.address2 AS address2,
        user.city AS city,
        user.zip_code AS zip_code,
        user.province AS province
    `

	INVESTOR_SPV_STRINGS string = `
        spv.id,
        spv.name AS spv_name
    `

	INVESTOR_PRIVATE_PROFILE_STRINGS string = `
        private_profile.first_name AS first_name,
        private_profile.family_name AS family_name,
        private_profile.first_name_kana AS first_name_kana,
        private_profile.family_name_kana AS family_name_kana
    `

	PROJECT_FILE_URL string = `
        (
            SELECT pf.file_url
            FROM project_file pf
            WHERE file_type = 'CONTRACT_BEFORE_PAGE'
              AND pf.project_id = project.id
              AND (deleted IS NULL OR deleted = '0001-01-01 00:00:00')
        ) AS contract_before_page_url,

        (
            SELECT pf.file_url
            FROM project_file pf
            WHERE file_type = 'SILENT_PARTNERSHIPS'
              AND pf.project_id = project.id
              AND (deleted IS NULL OR deleted = '0001-01-01 00:00:00')
        ) AS silent_partnerships_page_url,

        (
            SELECT pf.file_url
            FROM project_file pf
            WHERE file_type = 'PROJECT_REPORT'
              AND pf.project_id = project.id
              AND (deleted IS NULL OR deleted = '0001-01-01 00:00:00')
        ) AS project_report_url
    `

	INVESTMENT_QUERY string = `
        share.id AS investment_id,
        share.*
    `
)

// 투자현황 또는 투자목록
// <- 기호: 이 기호는 데이터의 흐름 방향을 나타내며, 데이터베이스에서 구조체로의 읽기 전용 필드
type Investment struct {
	Share             `xorm:"extends"`
	Project           `xorm:"extends"`
	InvestmentId      string        `json:"investment_id" xorm:"<-"`
	UserType          string        `json:"user_type" xorm:"<-"`
	FirstName         string        `json:"first_name" xorm:"<-"`
	FamilyName        string        `json:"family_name" xorm:"<-"`
	FirstNameKana     string        `json:"first_name_kana" xorm:"<-"`
	FamilyNameKana    string        `json:"family_name_kana" xorm:"<-"`
	BankName          string        `json:"bank_name"  xorm:"-"`
	BankAccountNumber string        `json:"bank_account_number" xorm:"-"`
	Depositor         string        `json:"depositor" xorm:"-"`
	SpvName           string        `json:"spv_name" xorm:"<-"`
	UserName          string        `json:"user_name" xorm:"<-"`
	InvestorType      INVESTOR_TYPE `json:"investor_type" xorm:"<-"`
	Email             string        `json:"email" xorm:"<-"`
	MobilePhone       string        `json:"mobile_phone" xorm:"<-"`

	Address  string `json:"address" xorm:"<-"`
	Address2 string `json:"address2" xorm:"<-"`
	City     string `json:"city" xorm:"<-"`
	ZipCode  string `json:"zip_code" xorm:"<-"`
	Province string `json:"province" xorm:"<-"`

	ContractStatus            string  `json:"contract_status" xorm:"-"`
	ContractAfterPageUrl      *string `json:"contract_after_page_url" xorm:"-"`
	ContractBeforePageUrl     *string `json:"contract_before_page_url" xorm:"<-"`
	SilentPartnershipsPageUrl string  `json:"silent_partnerships_page_url" xorm:"<-"`
	ProjectReportUrl          string  `json:"project_report_url" xorm:"<-"`
}

func (i Investment) IsUserTypePrivate() bool {
	return i.UserType == string(PRIVATE)
}

func (i Investment) FullName() string {
	return i.FamilyName + " " + i.FirstName
}

func (i Investment) FullNameKana() string {
	return i.FamilyNameKana + " " + i.FirstNameKana
}

func (invest *Investment) MarshalJSON() ([]byte, error) {
	var name, nameKana string
	name = invest.UserName
	if invest.UserType == string(PRIVATE) {
		name = invest.FamilyName + " " + invest.FirstName
		nameKana = invest.FamilyNameKana + " " + invest.FirstNameKana
	}

	if invest.ContractStatus == "" {
		invest.ContractStatus = string(CONTRACT_NONE)
	}

	var projectReportCheckDate jsonutils.JSONTime
	if invest.ProjectReportCheckDate.Time().After(invest.ProjectReportDate.Time()) {
		projectReportCheckDate = invest.ProjectReportCheckDate
	}

	return json.Marshal(map[string]interface{}{
		"id":                              invest.InvestmentId,
		"owner_id":                        invest.Share.OwnerId,        // Share의 OwnerId 사용
		"share_id":                        invest.Share.Id,             // Share ID
		"project_id":                      invest.ProjectId,            // 연결된 프로젝트 ID
		"category_type":                   invest.Share.CategoryType,   // Project의 CategoryType 사용
		"project_type":                    invest.Share.ProjectType,    // Project의 ProjectType 사용
		"invest_status":                   invest.InvestStatus,         // 투자 상태
		"amount":                          invest.Amount,               // 총 구매 가격
		"qty":                             invest.Qty,                  // 구매 수량
		"price":                           invest.Price,                // 1주당 가격
		"invest_date":                     invest.InvestDate,           // 거래 일자
		"spv_id":                          invest.SpvId,                // 개발회사 아이디
		"spv_name":                        invest.SpvName,              // 개발회사 명
		"project_name":                    invest.Project.Name,         // 상품명
		"start_at":                        invest.StartAt,              // 모집 시작일
		"end_at":                          invest.EndAt,                // 모집 종료일
		"main_image_url":                  invest.Project.MainImageUrl, // 메인 이미지 주소
		"project_status":                  invest.ProjectStatus,        // 프로젝트 상태
		"capacity":                        invest.Capacity,             // 규모
		"investor_type":                   invest.InvestorType,         // 구매자 투자자 구분
		"name":                            name,                        // 구매자 이름
		"name_kana":                       nameKana,                    // 구매자 후리가나
		"first_name":                      invest.FirstName,
		"family_name":                     invest.FamilyName,
		"first_name_kana":                 invest.FirstNameKana,
		"family_name_kana":                invest.FamilyNameKana,
		"address":                         invest.Address,           // 주소1
		"address2":                        invest.Address2,          // 주소2
		"city":                            invest.City,              // 도시
		"zip_code":                        invest.ZipCode,           // 우편번호
		"province":                        invest.Province,          // 현 이름
		"email":                           invest.Email,             // 구매자 이메일
		"phone":                           invest.MobilePhone,       // 구매자 전화번호
		"bank_name":                       invest.BankName,          // 은행 명
		"bank_account_number":             invest.BankAccountNumber, // 은행 계좌번호
		"depositor":                       invest.Depositor,         // 예금주
		"investment_date":                 invest.InvestDate,        // 투자 일시
		"contract_after_page_url":         invest.ContractAfterPageUrl,
		"contract_before_page_url":        invest.ContractBeforePageUrl,
		"silent_partnerships_page_url":    invest.SilentPartnershipsPageUrl,
		"project_report_url":              invest.ProjectReportUrl,
		"contract_status":                 invest.ContractStatus,
		"contract_before_page_check_date": invest.ContractBeforePageCheckDate,
		"contract_after_page_check_date":  invest.ContractAfterPageCheckDate,
		"silent_partnerships_check_date":  invest.SilentPartnershipsCheckDate,
		"project_report_check_date":       projectReportCheckDate,
		"public_view_end_at":              invest.EndAt,
	})

}

// CDN 주소 Mapping
func ContractDetailMapper(ctx context.Context, item *InvestmentDetail) error {
	if item != nil {
		cnt, contracts, err := Contract{}.GetByShareId(ctx, "created", "asc", item.Share.Id, 0, 1)
		if err != nil {
			return err
		}
		if cnt == 0 {
			return nil
		}

		item.ContractStatus = string(contracts[0].Status)
		if contracts[0].ContractAfterPageUrl != "" {
			item.ContractAfterPageUrl = &contracts[0].ContractAfterPageUrl
		}

		if item.ContractAfterPageUrl != nil {
			url := s3utils.GetUrlSign(*item.ContractAfterPageUrl)
			item.ContractAfterPageUrl = &url
		}
	}
	return nil
}
func ContractMapper(item *Investment) {
	if item != nil {
		if item.ContractAfterPageUrl != nil {
			url := s3utils.GetUrlSign(*item.ContractAfterPageUrl)
			item.ContractAfterPageUrl = &url
		}
	}
}

// CDN 주소 Mapping
func ContractsUrlMapper(ctx context.Context, items []Investment) error {
	for i := range items {
		cnt, contracts, err := Contract{}.GetByShareId(ctx, "created", "desc", items[i].Share.Id, 0, 1)
		if err != nil {
			return errors.Wrap(err, "failed to get contracts")
		}
		if cnt == 0 {
			continue
		}
		items[i].ContractStatus = string(contracts[0].Status)
		items[i].ContractAfterPageUrl = &contracts[0].ContractAfterPageUrl
		ContractMapper(&items[i])
	}
	return nil
}

type InvestmentDetail struct {
	Investment        `xorm:"extends"`
	DepositRequest    DepositRequest     `json:"deposit_request" xorm:"extends"`
	VaAccount         UserVirtualAccount `json:"va_account" xorm:"extends"`
	LastRepaymentDate *string            `json:"last_repayment_date" xorm:"-"`
}

func (invest *InvestmentDetail) MarshalJSON() ([]byte, error) {

	invBytes, err := invest.Investment.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(invBytes, &data)
	if err != nil {
		return nil, err
	}

	data["deposit_request"] = invest.DepositRequest
	data["va_account"] = invest.VaAccount
	data["last_repayment_date"] = invest.LastRepaymentDate

	return json.Marshal(data)
}

// 거래 상태/ 상품명 / 거래금액 /거래 일시
type InvestmentDashboardResponseDto struct {
	Id           string        `json:"id" xorm:"id <-"`
	InvestStatus INVEST_STATUS `json:"invest_status" xorm:"invest_status <-"`
	ProjectName  string        `json:"project_name" xorm:"name <-"`
	InvestAmount int64         `json:"invest_amount" xorm:"amount <-"`
	InvestDate   string        `json:"invest_date" xorm:"invest_date <-"`
}

func (Investment) GetDashboardList(ctx context.Context, offset, limit int) (int64, []InvestmentDashboardResponseDto, error) {
	query := factory.DB(ctx).Table("share").Select(INVESTMENT_QUERY+","+INVESTOR_PROJECT_STRINGS).
		Join("INNER", "project", "project.id = share.project_id")

	var items []InvestmentDashboardResponseDto

	var totalCount int64
	totalCount, err := query.Desc("share.created").Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	if items == nil {
		items = make([]InvestmentDashboardResponseDto, 0, 0)
	}
	return totalCount, items, nil
}

func (Investment) GetAll(ctx context.Context, sortby, order string, offset, limit int, investStatus, investorName, spvName, projectName, projectType *string, startAmount, endAmount *int) (int64, []Investment, error) {
	q := factory.DB(ctx)

	q.Table("share").Select(INVESTMENT_QUERY+","+INVESTOR_PROJECT_STRINGS+","+INVESTOR_USER_STRINGS+","+INVESTOR_SPV_STRINGS+","+INVESTOR_PRIVATE_PROFILE_STRINGS).
		Join("INNER", "project", "project.id = share.project_id").
		Join("INNER", "user", "share.owner_id = user.id").
		Join("LEFT", "private_profile", "share.owner_id = private_profile.user_id").
		Join("INNER", "spv", "project.spv_id = spv.id")

	if err := modelhelper.SetSortOrderForTable(q, "share", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	// 투자 상태 검색
	if err := modelhelper.SearchEqQuery(q, "share", "invest_status", investStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	// 투자 상태 검색
	if err := modelhelper.SearchEqQuery(q, "share", "project_type", projectType); err != nil {
		factory.Logger(ctx).Error(err)
	}

	// 투자자 명 검색
	if err := modelhelper.SearchLikeQuery(q, "user", "name", investorName); err != nil {
		factory.Logger(ctx).Error(err)
	}

	// 개발회사  검색
	if err := modelhelper.SearchLikeQuery(q, "spv", "name", spvName); err != nil {
		factory.Logger(ctx).Error(err)
	}
	// 상품명
	if err := modelhelper.SearchLikeQuery(q, "project", "name", projectName); err != nil {
		factory.Logger(ctx).Error(err)
	}

	// 기간 검색 기능
	if err := modelhelper.SetBetweenRange(q, "amount", startAmount, endAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var invests []Investment
	totalCount, err := q.Limit(limit, offset).FindAndCount(&invests)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, invests, nil
}

func (Investment) GetInvestmentsByOwnerId(ctx context.Context, id int64, projectType, projectStatus, investStatus *string, sortby, order string, offset, limit int) (int64, []Investment, error) {
	q := factory.DB(ctx)

	q.Table("share").Select(INVESTMENT_QUERY+","+INVESTOR_PROJECT_STRINGS+","+INVESTOR_USER_STRINGS+","+INVESTOR_PRIVATE_PROFILE_STRINGS+","+INVESTOR_SPV_STRINGS+","+PROJECT_FILE_URL).
		Join("INNER", "project", "project.id = share.project_id").
		Join("INNER", "user", "share.owner_id = user.id").
		Join("LEFT", "private_profile", "share.owner_id = private_profile.user_id").
		Join("INNER", "spv", "project.spv_id = spv.id").
		Where("share.owner_id = ?", id)

	if err := modelhelper.SearchEqQuery(q, "share", "project_type", projectType); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "project", "project_status", projectStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "share", "invest_status", investStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetSortOrderForTable(q, "share", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var invests []Investment
	totalCount, err := q.Limit(limit, offset).FindAndCount(&invests)
	if err != nil {
		return 0, nil, err
	}

	if err := ContractsUrlMapper(ctx, invests); err != nil {
		return 0, nil, err
	}

	return totalCount, invests, nil
}

func (Investment) GetInvestmentsByProjectId(ctx context.Context, id int64, sortby, order string, offset, limit int) (int64, []Investment, error) {
	q := factory.DB(ctx)

	q.Table("share").Select(INVESTMENT_QUERY+","+INVESTOR_PROJECT_STRINGS+","+INVESTOR_USER_STRINGS+","+INVESTOR_PRIVATE_PROFILE_STRINGS+","+INVESTOR_SPV_STRINGS+","+PROJECT_FILE_URL).
		Join("INNER", "project", "project.id = share.project_id").
		Join("INNER", "user", "share.owner_id = user.id").
		Join("LEFT", "private_profile", "share.owner_id = private_profile.user_id").
		Join("INNER", "spv", "project.spv_id = spv.id").
		Where("project.id = ?", id)

	if err := modelhelper.SetSortOrderForTable(q, "share", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var invests []Investment
	totalCount, err := q.Limit(limit, offset).FindAndCount(&invests)
	if err != nil {
		return 0, nil, err
	}

	if err := ContractsUrlMapper(ctx, invests); err != nil {
		return 0, nil, err
	}

	return totalCount, invests, nil
}

func (Investment) GetById(ctx context.Context, id string) (*InvestmentDetail, error) {

	q := factory.DB(ctx)

	var investDetail InvestmentDetail
	if has, err := q.Table("share").Select(INVESTMENT_QUERY+","+INVESTOR_PROJECT_STRINGS+","+INVESTOR_USER_STRINGS+","+INVESTOR_SPV_STRINGS+","+INVESTOR_SPV_STRINGS+","+INVESTOR_PRIVATE_PROFILE_STRINGS+","+PROJECT_FILE_URL+",deposit_request.*, user_virtual_account.*").
		Join("INNER", "project", "project.id = share.project_id").
		Join("INNER", "user", "share.owner_id = user.id").
		Join("LEFT", "private_profile", "share.owner_id = private_profile.user_id").
		Join("LEFT", "spv", "project.spv_id = spv.id").
		Join("INNER", "deposit_request", "deposit_request.share_id = share.id").
		Join("INNER", "user_virtual_account", "user_virtual_account.user_id = share.owner_id and user_virtual_account.linked_status = 'ON_VA_ACCOUNT'").
		Where("share.id = ?", id).Get(&investDetail); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	var user User
	if has, err := q.Table("user").Where("id = ?", investDetail.Share.OwnerId).Get(&user); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	investDetail.BankName = user.BankName
	investDetail.BankAccountNumber = user.BankAccount
	investDetail.Depositor = user.Depositor

	var err error

	// investDetail.LastRepaymentDate, err = Settlement{}.GetLastRepaymentDateProjectId(ctx, investDetail.Project.Id)
	// if err != nil {
	// 	return nil, err
	// // }
	lastSchedule, err := RepaymentDividendSchedule{}.GetLastSchedultByProjectIdCompleted(ctx, investDetail.Project.Id)
	if err != nil {
		return nil, err
	}
	if lastSchedule != nil {
		executeDate := lastSchedule.ExecuteDate.ToString()
		investDetail.LastRepaymentDate = &executeDate
	}

	if err := ContractDetailMapper(ctx, &investDetail); err != nil {
		return nil, err
	}
	return &investDetail, nil
}

func (Investment) PrepareAndReturnCsv(items []Investment) [][]string {

	type (
		invest struct {
			InvestNumber string             `json:"invest_number" i18n:"model.mypage.invest_number"`             // 거래 번호
			InvestStatus string             `json:"invest_status" i18n:"model.mypage.invest_status"`             // 거래 상태
			Name         string             `json:"name" i18n:"model.mypage.name"`                               // 투자자명
			SpvName      string             `json:"spv_name" i18n:"model.mypage.spv_name"`                       // 개발회사명
			ProjectType  string             `json:"project_type" i18n:"model.mypage.project_type"`               // 투자유형
			ProjectName  string             `json:"project_name" i18n:"model.mypage.project_name"`               // 상품명
			Pay          int64              `json:"pay" i18n:"model.mypage.pay"`                                 // 거래 금액
			InvestDate   jsonutils.JSONTime `json:"invest_date" i18n:"model.mypage.invest_date" xorm:"not null"` // 거래 일시
		}
	)

	var invests []invest
	for _, item := range items {
		// 거래번호, 거래상태, 투자자명, 개발회사명, 투자유형, 상품명, 거래금액, 거래일시
		inv := invest{
			InvestNumber: item.Share.Id,
			InvestStatus: string(item.InvestStatus),
			Name:         item.UserName,
			SpvName:      item.SpvName,
			ProjectType:  "commoncode.PROJECT_TYPE." + string(item.Share.ProjectType),
			ProjectName:  item.Project.Name,
			Pay:          item.Qty * item.Price,
			InvestDate:   item.InvestDate,
		}

		if value, ok := reflect.ValueOf(&item.InvestStatus).Elem().Interface().(INVEST_STATUS); ok {
			inv.InvestStatus = "commoncode.INVEST_STATUS." + string(value)
		}

		invests = append(invests, inv)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&invest{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range invests {

		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (Investment) DownloadCsv(ctx context.Context, w http.ResponseWriter, sortby, order string, investStatus, investorName, spvName, projectName *string, startAmount, endAmount *int) error {

	q := factory.DB(ctx)

	q.Table("share").Select("share.*,"+INVESTOR_PROJECT_STRINGS+","+INVESTOR_USER_STRINGS+","+INVESTOR_SPV_STRINGS+","+INVESTOR_PRIVATE_PROFILE_STRINGS).
		Join("INNER", "project", "project.id = share.project_id").
		Join("INNER", "user", "share.owner_id = user.id").
		Join("LEFT", "private_profile", "share.owner_id = private_profile.user_id").
		Join("INNER", "spv", "project.spv_id = spv.id")

	if err := modelhelper.SetSortOrderForTable(q, "share", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	// 투자 상태 검색
	if err := modelhelper.SearchEqQuery(q, "share", "invest_status", investStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	// 투자자 명 검색
	if err := modelhelper.SearchLikeQuery(q, "user", "name", investorName); err != nil {
		factory.Logger(ctx).Error(err)
	}

	// 개발회사  검색
	if err := modelhelper.SearchLikeQuery(q, "spv", "name", spvName); err != nil {
		factory.Logger(ctx).Error(err)
	}
	// 상품명
	if err := modelhelper.SearchLikeQuery(q, "project", "name", projectName); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenRange(q, "amount", startAmount, endAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var invests []Investment
	_, err := q.FindAndCount(&invests)
	if err != nil {
		return err
	}

	records := Investment{}.PrepareAndReturnCsv(invests)
	csvutils.WriteHeader(w, "energyshares_invest_list.csv")

	return csv.NewWriter(w).WriteAll(records)
}

func (Investment) InvestmentsByProjectIdPrepareAndReturnCsv(items []Investment) [][]string {

	type (
		invest struct {
			InvestorType string             `json:"investor_type" i18n:"model.mypage.investor_type"` // 투자 등급
			Email        string             `json:"email" i18n:"model.mypage.email"`
			Name         string             `json:"name" i18n:"model.mypage.name"`
			MobilePhone  string             `json:"mobile_phone" i18n:"model.mypage.mobile_phone"`
			Qty          int64              `json:"qty" i18n:"model.mypage.qty" swagger:"desc(수량)"`              // 주식 수량
			Pay          int64              `json:"pay" i18n:"model.mypage.pay"`                                 // 결제 금액
			InvestStatus string             `json:"invest_status" i18n:"model.mypage.invest_status"`             // 거래 상태
			InvestDate   jsonutils.JSONTime `json:"invest_date" i18n:"model.mypage.invest_date" xorm:"not null"` // 거래 일시
		}
	)

	var invests []invest
	for _, item := range items {
		inv := invest{
			InvestorType: "commoncode.INVESTOR_TYPE." + string(item.InvestorType),
			Email:        item.Email,
			Name:         item.UserName,
			MobilePhone:  item.MobilePhone,
			Qty:          item.Qty,
			Pay:          item.Qty * item.Price,
			InvestStatus: string(item.InvestStatus),
			InvestDate:   item.InvestDate,
		}

		if value, ok := reflect.ValueOf(&item.InvestStatus).Elem().Interface().(INVEST_STATUS); ok {
			inv.InvestStatus = "commoncode.INVEST_STATUS." + string(value)
		}

		invests = append(invests, inv)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&invest{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range invests {

		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (Investment) InvestmentsDownloadCsv(ctx context.Context, w http.ResponseWriter, id int64, sortby, order string, offset, limit int) error {

	q := factory.DB(ctx)

	q.Table("share").Select("share.*,"+INVESTOR_PROJECT_STRINGS+","+INVESTOR_USER_STRINGS+","+INVESTOR_SPV_STRINGS).
		Join("INNER", "project", "project.id = share.project_id").
		Join("INNER", "user", "share.owner_id = user.id").
		Join("LEFT", "private_profile", "share.owner_id = private_profile.user_id").
		Join("INNER", "spv", "project.spv_id = spv.id").
		Where("project.id = ?", id)

	if err := modelhelper.SetSortOrderForTable(q, "share", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var invests []Investment
	_, err := q.FindAndCount(&invests)
	if err != nil {
		return nil
	}

	records := Investment{}.InvestmentsByProjectIdPrepareAndReturnCsv(invests)

	fileName := fmt.Sprintf("%s_%d_%s", "energyshares_member", id, "invest_list.csv")
	csvutils.WriteHeader(w, fileName)

	return csv.NewWriter(w).WriteAll(records)
}
