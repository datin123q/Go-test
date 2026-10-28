package domain

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type VIRTUAL_ACCOUNT_LINK_STATUS string

const (
	VIRTUAL_ACCOUNT_LINK_STATUS_NAME string                      = "VIRTUAL_ACCOUNT_LINK_STATUS"
	ON_VA_ACCOUNT                    VIRTUAL_ACCOUNT_LINK_STATUS = "ON_VA_ACCOUNT"
	OFF_VA_ACCOUNT                   VIRTUAL_ACCOUNT_LINK_STATUS = "OFF_VA_ACCOUNT"
	INVALID_VA_ACCOUNT               VIRTUAL_ACCOUNT_LINK_STATUS = "INVALID_VA_ACCOUNT"
)

const (
	VA_STATUS_STOP   = "1"
	VA_STATUS_START  = "2"
	VA_STATUS_DELETE = "3"
)

func (UserVirtualAccount) GetVaLinkStatusCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(ON_VA_ACCOUNT), Name: i18nutils.Translate("commoncode.VIRTUAL_ACCOUNT_LINK_STATUS.ON_VA_ACCOUNT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(OFF_VA_ACCOUNT), Name: i18nutils.Translate("commoncode.VIRTUAL_ACCOUNT_LINK_STATUS.OFF_VA_ACCOUNT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVALID_VA_ACCOUNT), Name: i18nutils.Translate("commoncode.VIRTUAL_ACCOUNT_LINK_STATUS.INVALID_VA_ACCOUNT", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: VIRTUAL_ACCOUNT_LINK_STATUS_NAME,
		Codes:     codes,
	}
}

type VaStatusChangeInput struct {
	VaId              string `json:"va_id,omitempty" swagger:"desc(가상계좌 ID)"`
	VaStatusCode      string `json:"va_status_code,omitempty" swagger:"desc(가상계좌 상태 코드 1:중지 2:재개 3:삭제),enum(1|2|3)"`
	IssueDataFrom     string `json:"issue_data_from,omitempty" swagger:"desc(발행일 기준 조회 시작일)"`
	IssueDataTo       string `json:"issue_data_to,omitempty" swagger:"desc(발행일 기준 조회 종료일)"`
	VaContractAuthKey string `json:"va_contract_auth_key,omitempty" swagger:"desc(계약용 Key)"`
}

// ユーザーごとに1つずつ仮想口座が割り当てられます。
type UserVirtualAccount struct {
	Id               int64                       `json:"id" xorm:"pk autoincr unique" swagger:"desc(고유 ID)"`
	UserId           int64                       `json:"user_id" xorm:"user_id" swagger:"desc(사용자 ID)"`
	VaBankName       string                      `json:"va_bank_name" xorm:"va_bank_name" swagger:"desc(은행명)"`
	VaBranchCode     string                      `json:"va_branch_code" xorm:"va_branch_code unique(va_branch_account)" swagger:"desc(支店コード)"`      // 支店コード
	VaBranchName     string                      `json:"va_branch_name" xorm:"va_branch_name" swagger:"desc(支店名)"`                                  // 支店名
	VaAccountNumber  string                      `json:"va_account_number" xorm:"va_account_number unique(va_branch_account)" swagger:"desc(口座番号)"` // 口座番号
	VaHolderNameKana string                      `json:"va_holder_name_kana" xorm:"va_holder_name_kana" swagger:"desc(口座名義人)"`                      // 口座名義人ひらがな
	ProjectType      PROJECT_TYPE                `json:"project_type" xorm:"project_type" swagger:"상품유형"`
	LinkedStatus     VIRTUAL_ACCOUNT_LINK_STATUS `json:"linked_status" xorm:"linked_status" swagger:"desc(仮想口座ステータス)"` // アカウント接続ステータス
	UpdateDate       jsonutils.JSONTime          `json:"update_date" xorm:"update_date" swagger:"desc(更新日付)"`
	Created          jsonutils.JSONTime          `json:"-" xorm:"created"`
	Updated          jsonutils.JSONTime          `json:"-" xorm:"updated"`
}

func (u UserVirtualAccount) IsEmpty() bool {
	return u.Id == 0
}

func (UserVirtualAccount) TableName() string {
	return "user_virtual_account"
}

func (u UserVirtualAccount) SelectAll() string {
	return u.TableName() + ".*"
}

// InsertMany は渡されたitemsをまとめてINSERTする。以前はここで独自に
// session.Begin/Commit/defer Closeを呼んでいたため、呼び出し元のcronジョブ
// (va_account_jobs.go)がこの呼び出し1回でジョブ全体のセッションを閉じてしまい、
// 後続のジョブ実行時刻の更新・操作履歴の記録が失敗していた(MIR-1183)。
// Transfer.InsertMany(MIR-1181)と同じくトランザクション制御は呼び出し側に委ね、
// ctxが差すセッション/エンジンへそのままINSERTするだけにする。
func (UserVirtualAccount) InsertMany(ctx context.Context, items []UserVirtualAccount) (int64, error) {
	count, err := factory.DB(ctx).Insert(items)
	if err != nil {
		return 0, err
	}

	// 仮想口座アラートバッチ実行ログを適用
	adminUser, err := AdminUser{}.GetSuperAdmin(ctx)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		history := OperatorHistory{
			UserId: adminUser.Id,
			Email:  adminUser.Email,
			Menu:   "commoncode.LOG_CONTENT_TYPE.VA_BATCH_EXECUTE",
			Ip:     "::1",
			State:  OPERATOR_STATE("REGISTER"),
		}

		_, err = history.Create(ctx)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (v *UserVirtualAccount) Create(ctx context.Context) (int64, error) {
	q := factory.DB(ctx)
	return q.Insert(v)
}

func (v *UserVirtualAccount) Update(ctx context.Context) (int64, error) {
	q := factory.DB(ctx).Table("user_virtual_account")
	return q.ID(v.Id).Update(v)
}

func (UserVirtualAccount) GetVaByUserId(ctx context.Context, id int64, projectType PROJECT_TYPE) (*UserVirtualAccount, error) {
	var account UserVirtualAccount
	if has, err := factory.DB(ctx).Table("user_virtual_account").Select("user_virtual_account.*").Where("user_id = ? AND project_type = ? AND linked_status = ?", id, projectType, ON_VA_ACCOUNT).Get(&account); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &account, nil
}

// 支店コードと仮想口座番号に基づいて仮想口座を照会
func (UserVirtualAccount) GetUserVirtualAccountByBranchAndAccountNumber(ctx context.Context, branchCode, number string) (*UserVirtualAccount, error) {
	var account UserVirtualAccount
	if has, err := factory.DB(ctx).Table("user_virtual_account").Select("user_virtual_account.*").Where("va_branch_code = ? AND va_account_number = ? AND linked_status = ?", branchCode, number, ON_VA_ACCOUNT).Get(&account); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &account, nil
}

func (UserVirtualAccount) GetAll(ctx context.Context, vaBranchCode, vaAccountNumber, linkedStatus, userId, projectType *string, startDate, endDate string, sortby, order string, offset, limit int) (int64, []UserVirtualAccount, error) {
	query := factory.DB(ctx).Table("user_virtual_account")

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user_virtual_account", "va_branch_code", vaBranchCode); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "user_virtual_account", "va_account_number", vaAccountNumber); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "user_virtual_account", "linked_status", linkedStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user_virtual_account", "project_type", projectType); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "user_virtual_account", "user_id", userId); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(query, "user_virtual_account.update_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var userList []UserVirtualAccount
	var totalCount int64

	totalCount, err := query.Limit(limit, offset).FindAndCount(&userList)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, userList, nil
}

// リンクされていない仮想口座を照会
func (a UserVirtualAccount) GetVaByUnLinkedAccount(ctx context.Context, projectType PROJECT_TYPE) (*UserVirtualAccount, error) {
	var account UserVirtualAccount
	if has, err := factory.DB(ctx).Table("user_virtual_account").Where("linked_status = ? AND project_type = ?", OFF_VA_ACCOUNT, projectType).Limit(1).Get(&account); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &account, nil
}

func (a UserVirtualAccount) GetVaByUnLinkedAccountCount(ctx context.Context, projectType PROJECT_TYPE) (int64, error) {
	var account UserVirtualAccount
	var cnt int64
	cnt, err := factory.DB(ctx).Table("user_virtual_account").Where("linked_status = ? AND project_type = ?", OFF_VA_ACCOUNT, projectType).Count(&account)
	if err != nil {
		return 0, err
	}

	return cnt, nil
}

func (v *UserVirtualAccount) LinkUser(ctx context.Context, user_id int64) error {
	v.LinkedStatus = ON_VA_ACCOUNT
	v.UserId = user_id
	v.UpdateDate = jsonutils.JSONTime(time.Now())
	_, err := v.Update(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (v *UserVirtualAccount) UnLinkUser(ctx context.Context) error {
	v.LinkedStatus = OFF_VA_ACCOUNT
	v.UserId = 0
	v.UpdateDate = jsonutils.JSONTime(time.Now())
	_, err := v.Update(ctx)
	if err != nil {
		return err
	}
	return nil
}

func VaStatusChangeRequest(ctx context.Context, vaId, status string) error {
	v := VaStatusChangeInput{
		VaId:         vaId,
		VaStatusCode: status,
	}
	pbytes, _ := json.Marshal(v)
	buff := bytes.NewBuffer(pbytes)

	url := serviceutils.GetJobSchedulerUrl() + "/va-accounts/change-status"

	resp, err := http.Post(url, "application/json", buff)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = resp.Body.Close()

	factory.Logger(ctx).WithFields(logrus.Fields{
		"body": body,
	}).Info("success va_account status changed")

	return err
}

func (v *UserVirtualAccount) InvalidAccount(ctx context.Context) error {
	v.LinkedStatus = INVALID_VA_ACCOUNT
	v.UserId = 0
	v.UpdateDate = jsonutils.JSONTime(time.Now())
	vaId := v.VaBranchCode + v.VaAccountNumber
	err := VaStatusChangeRequest(ctx, vaId, VA_STATUS_DELETE)
	if err != nil {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"va_id":             vaId,
			"va_account_number": v.VaAccountNumber,
			"err":               err,
		}).Error("failed to va_account status change ")
	}
	_, err = v.Update(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to va_account update")
	}
	return nil
}

// 사용자의 가상계좌 무효화
func (UserVirtualAccount) InvalidAccounts(ctx context.Context, userId int64, projecType PROJECT_TYPE) error {
	var items []UserVirtualAccount
	var err error
	q := factory.DB(ctx).Table("user_virtual_account")
	q = q.Where("user_id = ? AND project_type = ? ", userId, projecType)
	err = q.Find(&items)
	if err != nil {
		return errors.Wrap(err, "failed to va_account find")
	}
	for i := range items {
		err2 := items[i].InvalidAccount(ctx)
		if err2 != nil {
			return errors.Wrap(err, "failed to va_account invalid")
		}
	}
	return nil
}

func (UserVirtualAccount) DownloadCsv(ctx context.Context, w http.ResponseWriter, vaBranchCode, vaAccountNumber, projectType, linkedStatus, userId *string, startDate, endDate string, sortby, order string) error {
	query := factory.DB(ctx).Table("user_virtual_account")

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user_virtual_account", "va_branch_code", vaBranchCode); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "user_virtual_account", "va_account_number", vaAccountNumber); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "user_virtual_account", "linked_status", linkedStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "user_virtual_account", "project_type", projectType); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "user_virtual_account", "user_id", userId); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(query, "user_virtual_account.update_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var vaList []UserVirtualAccount
	_, err := query.FindAndCount(&vaList)
	if err != nil {
		return err
	}

	records := UserVirtualAccount{}.PrepareAndReturnCsv(vaList)
	csvutils.WriteHeader(w, "energyshares_virtual_account_list.csv")

	return csv.NewWriter(w).WriteAll(records)
}

func (UserVirtualAccount) PrepareAndReturnCsv(items []UserVirtualAccount) [][]string {
	type va struct {
		VaBranchCode    string `json:"va_branch_code" i18n:"model.virtual_account.va_branch_code" swagger:"desc(仮想口座支店番号)"`     // 支店コード
		VaAccountNumber string `json:"va_account_number" i18n:"model.virtual_account.va_account_number" swagger:"desc(仮想口座番号)"` // 口座番号
		ProjectType     string `json:"project_type" i18n:"model.virtual_account.project_type" swagger:"desc(投資タイプ)"`            // 口座番号
		LinkedStatus    string `json:"linked_status" i18n:"model.virtual_account.linked_status" swagger:"desc(仮想口座ステータス)"`      // アカウント接続ステータス
		UserId          int64  `json:"user_id" i18n:"model.virtual_account.user_id" swagger:"desc(투자자 ID)"`
		UpdateDate      string `json:"update_date" i18n:"model.virtual_account.update_date" swagger:"desc(更新日時)"`
	}

	var vas []va
	for _, item := range items {
		c := va{
			VaBranchCode:    item.VaBranchCode,
			VaAccountNumber: item.VaAccountNumber,
			ProjectType:     string(item.ProjectType),
			UserId:          item.UserId,
			UpdateDate:      item.UpdateDate.ToString(),
		}

		if value, ok := reflect.ValueOf(&item.LinkedStatus).Elem().Interface().(VIRTUAL_ACCOUNT_LINK_STATUS); ok {
			c.LinkedStatus = "commoncode.VIRTUAL_ACCOUNT_LINK_STATUS." + string(value)
		}

		if value, ok := reflect.ValueOf(&item.ProjectType).Elem().Interface().(PROJECT_TYPE); ok {
			c.ProjectType = "commoncode.PROJECT_TYPE." + string(value)
		}

		vas = append(vas, c)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&va{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range vas {
		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (UserVirtualAccount) CheckIfNotSufficient(ctx context.Context) (bool, error) {
	var account UserVirtualAccount
	totalCnt, err := factory.DB(ctx).Table("user_virtual_account").Count(&account)
	if err != nil {
		return false, err
	}

	offCnt, err := factory.DB(ctx).Table("user_virtual_account").Where("linked_status = ? ", OFF_VA_ACCOUNT).Count(&account)
	if err != nil {
		return false, err
	}

	invalidCnt, err := factory.DB(ctx).Table("user_virtual_account").Where("linked_status = ? ", INVALID_VA_ACCOUNT).Count(&account)
	if err != nil {
		return false, err
	}

	factory.Logger(ctx).WithFields(logrus.Fields{
		"total_count":   totalCnt,
		"off_count":     offCnt,
		"invalid_count": invalidCnt,
	}).Info("current va_account status count")

	// 仮想口座が100個以下の場合は不足している（not sufficient, true）
	// 100個はテスト用の一時設定
	// https://illuminarean.atlassian.net/browse/ESJP-561
	if offCnt <= 100 {
		return true, nil
	}
	return false, nil
}
