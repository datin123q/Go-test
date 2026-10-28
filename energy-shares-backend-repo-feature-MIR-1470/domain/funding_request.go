package domain

import (
	"context"
	"encoding/csv"
	"net/http"
	"reflect"
	"time"

	"github.com/gofrs/uuid"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

type REQUEST_STATUS string

const (
	REQUEST_STATUS_CODE                = "REQUEST_STATUS"
	STANDBY_REQUEST     REQUEST_STATUS = "STANDBY_REQUEST"    // 심사대기중
	PROCESSING_REQUEST  REQUEST_STATUS = "PROCESSING_REQUEST" // 심사처리중
	PASS_REQUEST        REQUEST_STATUS = "PASS_REQUEST"       // [적합]심사완료
	FAIL_REQUEST        REQUEST_STATUS = "FAIL_REQUEST"       // [부적합]심사완료
)

func (FundingRequest) GetReqeustStatusCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(STANDBY_REQUEST), Name: i18nutils.Translate("commoncode.REQUEST_STATUS.STANDBY_REQUEST", loc), Order: 1})
	codes = append(codes, commoncode.Code{Code: string(PROCESSING_REQUEST), Name: i18nutils.Translate("commoncode.REQUEST_STATUS.PROCESSING_REQUEST", loc), Order: 2})
	codes = append(codes, commoncode.Code{Code: string(PASS_REQUEST), Name: i18nutils.Translate("commoncode.REQUEST_STATUS.PASS_REQUEST", loc), Order: 3})
	codes = append(codes, commoncode.Code{Code: string(FAIL_REQUEST), Name: i18nutils.Translate("commoncode.REQUEST_STATUS.FAIL_REQUEST", loc), Order: 4})

	return commoncode.CodeGroup{
		GroupCode: REQUEST_STATUS_CODE,
		Codes:     codes,
	}
}

type FundingRequest struct {
	Id                  int64              `json:"id"`
	CommonId            string             `json:"common_id" xorm:"not null UUID"`
	Name                string             `json:"name" xorm:"varchar(255) not null" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_NAME"` // 회사명
	ZipCode             string             `json:"zip_code" xorm:"not null" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_ZIP_CODE" swagger:"desc(우편번호)"`
	Address             string             `json:"address" xorm:"not null" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_ADDRESS" swagger:"desc(주소)"`
	Address2            string             `json:"address2" xorm:"varchar(255)" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_ADDRESS2" swagger:"desc(주소)"`
	CorpNo              string             `json:"corp_no" xorm:"not null" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_CORP_NO" swagger:"desc(법인번호)"` // 법인번호
	Province            string             `json:"province" xorm:"not null" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_PROVINCE" swagger:"desc(도, 시, 특별시)"`
	City                string             `json:"city" xorm:"not null" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_CITY" swagger:"desc(시 또는 구, 군)"`
	Email               string             `json:"email" xorm:"varchar(255) not null" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EMAIL" `
	Phone               string             `json:"phone" xorm:"varchar(30) not null" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_PHONE"`
	EquipmentName       string             `json:"equipment_name" xorm:"varchar(255)" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EQUIPMENT_NAME" swagger:"desc(설비명칭)"`
	EquipmentAddress    string             `json:"equipment_address" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EQUIPMENT_ADDRESS" swagger:"desc(설비 주소)"`
	EquipmentAddress2   string             `json:"equipment_address2" xorm:"varchar(255)" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EQUIPMENT_ADDRESS2" swagger:"desc(설비 상세 주소 )"`
	EquipmentInfo       string             `json:"equipment_info" xorm:"varchar(400)" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EQUIPMENT_INFO"`             // 프로젝트 정보
	EquipmentConsulting string             `json:"equipment_consulting" xorm:"varchar(400)" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_CONTENT_EQUIPMENT_CONSULTING"` // 상담 내용
	RequestStatus       REQUEST_STATUS     `json:"request_status" i18n:"commoncode.FUNDING_HISTORY_CONTENT_TYPE.FUNDING_REVIEW_STATE"`
	Created             jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated             jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted             jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
	Attachments         []*Attachment      `json:"attachments" xorm:"-"`
}

const FUNDING_REQUEST_FILE_TYPE = "funding_request_file"

type FundingRequestFile struct {
	Id            int64     `json:"id" swagger:"desc(프로젝트 이미지 고유 ID)"`
	FileName      string    `json:"file_name" xorm:"UNIQUE NOT NULL" swagger:"desc(파일명)"`
	OriginalName  string    `json:"original_name" xorm:"NOT NULL" swagger:"desc(오리지날 파일명)"`
	FileSize      int64     `json:"file_size" xorm:"NOT NULL" swagger:"desc(파일 사이즈)"`
	FileExtension string    `json:"file_extension" xorm:"NOT NULL" swagger:"desc(파일 확장자)"`
	FilePath      string    `json:"file_path" xorm:"NOT NULL" swagger:"desc(파일 경로)"`
	FileUrl       string    `json:"file_url" xorm:"NOT NULL" swagger:"desc(URL)"`
	Created       time.Time `json:"created" xorm:"created"`
	Updated       time.Time `json:"updated" xorm:"updated"`
	Deleted       time.Time `json:"deleted" xorm:"deleted"`
}

func (f *FundingRequestFile) GetKey() string {
	return f.FilePath + f.FileName + f.FileExtension
}

func (f *FundingRequestFile) GetCDNUrl() string {
	return s3utils.MakeCDNURL(f.GetKey())
}

func (f *FundingRequestFile) GetCDNSignUrl() string {
	return s3utils.MakeCDNSignURL(f.GetKey())
}

func (f *FundingRequestFile) GetFileType() string {
	return FUNDING_REQUEST_FILE_TYPE
}

func (f *FundingRequestFile) GetFileName() string {
	return f.OriginalName
}

func (f *FundingRequestFile) SetLocation(loc string) {
	f.FileUrl = loc
}

func (f *FundingRequestFile) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(f)
}

func (FundingRequestFile) GetById(ctx context.Context, id int64) (*FundingRequestFile, error) {

	var f FundingRequestFile
	if has, err := factory.DB(ctx).ID(id).Get(&f); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &f, nil
}

func (FundingRequest) TableName() string {
	return "funding_request"
}

func (FundingRequest) GetAll(ctx context.Context, sortby, order string, offset, limit int, id, email, phone, requestStatus *string, startDate, endDate string) (int64, []FundingRequest, error) {

	q := factory.DB(ctx).Table("funding_request").Select("funding_request.*")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "funding_request", "id", id); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "funding_request", "email", email); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "funding_request", "phone", phone); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "funding_request", "request_status", requestStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "created", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]FundingRequest, 0, 0)
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (FundingRequest) GetDashboardList(ctx context.Context, offset, limit int) (int64, []FundingRequest, error) {
	q := factory.DB(ctx).Table("funding_request").Select("funding_request.*")

	items := make([]FundingRequest, 0, 0)

	var totalCount int64
	totalCount, err := q.Where("request_status=?", STANDBY_REQUEST).Desc("created").Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	if items == nil {
		items = make([]FundingRequest, 0)
	}

	return totalCount, items, nil
}

func (fr *FundingRequest) Create(ctx context.Context) (int64, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return 0, err
	}
	fr.CommonId = uuid.String()

	if fr.RequestStatus == "" {
		fr.RequestStatus = STANDBY_REQUEST
	}

	return factory.DB(ctx).Insert(fr)
}

func (FundingRequest) GetById(ctx context.Context, id int64) (*FundingRequest, error) {
	var fr FundingRequest
	has, err := factory.DB(ctx).ID(id).Get(&fr)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	attachments, err := Attachment{}.FindByResourceId(ctx, modelhelper.MaxFile, fr.CommonId)
	if err != nil {
		return nil, err
	}
	fr.Attachments = attachments

	return &fr, nil
}

func (fr *FundingRequest) Update(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).UseBool("is_forest_development", "is_environmental_impact").ID(fr.Id).Update(fr)
	return
}

// 법인 번호 중복 체크
func (FundingRequest) GetByCorpNo(ctx context.Context, corp_no string) (*FundingRequest, error) {
	var fr FundingRequest

	if has, err := factory.DB(ctx).Table("funding_request").Where("corp_no = ?", corp_no).Get(&fr); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &fr, nil
}

func (FundingRequest) PrepareAndReturnCsv(items []FundingRequest) [][]string {

	type fundingReqeust struct {
		Id          int64              `json:"id" i18n:"model.funding_request.id"`                  // 신청 아이디
		Status      string             `json:"status" i18n:"model.funding_request.status"`          // 상태
		Name        string             `json:"name" i18n:"model.funding_request.corp_name"`         // 법인명
		Email       string             `json:"email" i18n:"model.funding_request.email"`            // 이메일
		PhoneNumber string             `json:"phone_number" i18n:"model.funding_request.phone"`     // 전화번호
		StartDate   jsonutils.JSONTime `json:"start_date"  i18n:"model.funding_request.start_date"` // 대출의뢰일
	}

	var requests []fundingReqeust
	for _, item := range items {
		c := fundingReqeust{
			Id:          item.Id,
			Status:      string(item.RequestStatus),
			Name:        item.Name,
			Email:       item.Email,
			PhoneNumber: item.Phone,
			StartDate:   item.Created,
		}

		if value, ok := reflect.ValueOf(&item.RequestStatus).Elem().Interface().(REQUEST_STATUS); ok {
			c.Status = "commoncode.REQUEST_STATUS." + string(value)
		}

		requests = append(requests, c)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&fundingReqeust{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range requests {

		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (FundingRequest) DownloadCsv(ctx context.Context, w http.ResponseWriter, sortby, order string, offset, limit int, id, email, phone, requestStatus *string, startDate, endDate string) error {
	var err error
	q := factory.DB(ctx)
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "funding_request", "id", id); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "funding_request", "email", email); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "funding_request", "phone", phone); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "funding_request", "request_status", requestStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "created", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []FundingRequest

	_, err = q.FindAndCount(&items)
	if err != nil {
		return err
	}

	records := FundingRequest{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeader(w, "energyshares_funding_request_list.csv")

	return csv.NewWriter(w).WriteAll(records)

}
