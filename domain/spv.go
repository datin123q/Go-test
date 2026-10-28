package domain

import (
	"context"
	"encoding/csv"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type Spv struct {
	Id       int64              `json:"id"`
	Name     string             `json:"name" xorm:"not null" i18n:"commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_CORP_NAME" swagger:"desc(회사명)"` // 회사명
	Email    string             `json:"email" xorm:"not null" i18n:"commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_EMAIL" swagger:"desc(메일)"`     // 이메일
	Address  string             `json:"address" i18n:"commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_ADDRESS" swagger:"desc(주소1)"`                // 주소1
	Address2 string             `json:"address2" i18n:"commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_ADDRESS2" swagger:"desc(주소2)"`              // 주소2
	City     string             `json:"city" i18n:"commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_CITY" swagger:"desc(도시)"`                       // 도시
	ZipCode  string             `json:"zip_code" swagger:"desc(우편번호)"`
	Province string             `json:"province" i18n:"commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_PROVINCE" swagger:"desc(현 이름)"`                      // 현 이름
	CorpNo   string             `json:"corp_no" xorm:"UNIQUE not null" i18n:"commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_CORPNO" swagger:"desc(법인 번호)"` // 사업자등록번호
	Phone    string             `json:"phone" xorm:"not null" i18n:"commoncode.SPV_CHANGED_CONTENT_TYPE.SPV_CHANGED_CONTENT_PHONE" swagger:"desc(전화 번호)"`           // 전화번호
	OwnerId  int64              `json:"owner_id" xorm:"index"`
	UpdateId int64              `json:"update_id" xorm:"index"`
	Projects []*Project         `json:"projects" xorm:"-" swagger:"desc(펀드 목록)"` // 연결된 프로젝트
	Created  jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated  jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted  jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (s Spv) TableName() string {
	return "spv"
}

func (s Spv) SelectAll() string {
	return s.TableName() + ".*"
}

func (Spv) GetAll(ctx context.Context, id, name, email, phone *string, sortby, order string, offset, limit int) (int64, []Spv, error) {
	q := factory.DB(ctx).Table("spv").Select("spv.*")
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "spv", "id", id); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "spv", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "spv", "email", email); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "spv", "phone", phone); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []Spv
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func init() {

}

func (c *Spv) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(c)
}

// 법인 번호 체크
func (Spv) GetByCropNo(ctx context.Context, corpNo string) (*Spv, error) {
	var company Spv

	if has, err := factory.DB(ctx).Where("corp_no = ?", strings.ToLower(corpNo)).Get(&company); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &company, nil
}

func (Spv) GetById(ctx context.Context, id int64) (*Spv, error) {
	var v Spv
	if has, err := factory.DB(ctx).ID(id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &v, nil
}

func (c *Spv) Update(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).ID(c.Id).Cols("name", "email", "address", "address2", "city", "zip_code", "province", "phone", "corp_no").Update(c)
	return err
}

func (Spv) PrepareAndReturnCsv(items []Spv) [][]string {

	type spv struct {
		Id          string             `json:"id" i18n:"model.spv.id"`                // 개발회사 아이디
		Name        string             `json:"name" i18n:"model.spv.corp_name"`       // 법인명
		Email       string             `json:"email" i18n:"model.spv.email"`          // 이메일
		PhoneNumber string             `json:"phone_number" i18n:"model.spv.phone"`   // 전화번호
		Address     string             `json:"address" i18n:"model.spv.address"`      // 주소
		PostDate    jsonutils.JSONTime `json:"post_date"  i18n:"model.spv.post_date"` // 등록일자
	}

	var spvs []spv
	for _, item := range items {
		c := spv{
			Id:          strconv.FormatInt(item.Id, 10),
			Name:        item.Name,
			Email:       item.Email,
			PhoneNumber: item.Phone,
			Address:     item.ZipCode + " " + item.Province + " " + item.City + " " + item.Address + " " + item.Address2,
			PostDate:    item.Created,
		}

		spvs = append(spvs, c)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&spv{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range spvs {

		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (Spv) DownloadCsv(ctx context.Context, w http.ResponseWriter, id, name, email, phone *string, sortby, order string, offset, limit int) error {

	q := factory.DB(ctx).Table("spv").Select("spv.*")
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "spv", "id", id); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "spv", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "spv", "email", email); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "spv", "phone", phone); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []Spv

	_, err := q.FindAndCount(&items)
	if err != nil {
		return err
	}

	records := Spv{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeader(w, "energyshares_spv_list.csv")

	return csv.NewWriter(w).WriteAll(records)

}
