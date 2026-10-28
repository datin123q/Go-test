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

type LOGIN_RESULT string

const (
	LOGIN_RESULT_CODE              = "LOGIN_RESULT"
	SUCCESS_LOGIN     LOGIN_RESULT = "SUCCESS_LOGIN"
	FAIL_LOGIN        LOGIN_RESULT = "FAIL_LOGIN"
)

func (User) GetLoginResultCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)

	codes = append(codes, commoncode.Code{Code: string(SUCCESS_LOGIN), Name: i18nutils.Translate("commoncode.LOGIN_RESULT.SUCCESS_LOGIN", loc), Order: 1})
	codes = append(codes, commoncode.Code{Code: string(FAIL_LOGIN), Name: i18nutils.Translate("commoncode.LOGIN_RESULT.FAIL_LOGIN", loc), Order: 2})

	return commoncode.CodeGroup{
		GroupCode: LOGIN_RESULT_CODE,
		Codes:     codes,
	}
}

type LoginHistory struct {
	Id          int64              `json:"id"`
	UserId      int64              `json:"user_id" xorm:"index" swagger:"desc(사용자 ID)"`
	Name        string             `json:"name" xorm:"not null"`
	Email       string             `json:"email" xorm:"not null"`
	IpAddr      string             `json:"ip_addr" swagger:"desc(ClientIp)"`
	Agent       string             `json:"agent" swagger:"desc(브라우저)"`
	LoginResult LOGIN_RESULT       `json:"login_result" swagger:"desc(로그인 결과)"`
	LoginTime   jsonutils.JSONTime `json:"login_time" swagger:"desc(로그인 시간)"`
	Created     jsonutils.JSONTime `json:"created" xorm:"created"` // 作成日
}

func (l LoginHistory) TableName() string {
	return "login_history"
}

func (l LoginHistory) SelectAll() string {
	return l.TableName() + ".*"
}

func (LoginHistory) GetAll(ctx context.Context, sortby, order string, offset, limit int, name *string, startDate, endDate string) (int64, []LoginHistory, error) {
	q := factory.DB(ctx).Table("login_history").Select("login_history.*")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "login_history", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "login_time", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}
	var items []LoginHistory
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (lh LoginHistory) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(lh)
}

func (lh LoginHistory) FailCountCheck(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(lh)
}

func (LoginHistory) Delete(ctx context.Context, id int64) error {
	_, err := factory.DB(ctx).ID(id).Delete(&LoginHistory{})
	return err
}

func (LoginHistory) PrepareAndReturnCsv(items []LoginHistory) [][]string {

	//if items == nil {
	//	return nil
	//}

	type history struct {
		LoginTime jsonutils.JSONTime `json:"login_time" i18n:"model.login_history.login_time"` // アクセス日時
		NameEmail string             `json:"name_email" i18n:"model.login_history.email"`      // メール(ID)
		Ip        string             `json:"ip" i18n:"model.login_history.ip"`                 // アクセスIP
		Device    string             `json:"device" i18n:"model.login_history.device"`         // デバイス情報/位置情報/等
	}

	var histories []history
	for _, item := range items {
		h := history{
			LoginTime: item.LoginTime,
			NameEmail: item.Name + "(" + item.Email + ")",
			Ip:        item.IpAddr,
			Device:    item.Agent,
		}

		histories = append(histories, h)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&history{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range histories {

		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (LoginHistory) DownloadCsv(ctx context.Context, w http.ResponseWriter, sortby, order string, offset, limit int, name *string, startDate, endDate string) error {

	q := factory.DB(ctx).Table("login_history").Select("login_history.*")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "login_history", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "login_time", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []LoginHistory
	// TODO：今後、開始位置とリミット値を表現する必要がある場合は、次のようにq.Limit(offset, limit).FindAndCount(&items)修正が必要
	_, err := q.FindAndCount(&items)
	if err != nil {
		return err
	}

	records := LoginHistory{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeader(w, "energyshares_login_history.csv")

	return csv.NewWriter(w).WriteAll(records)
}
