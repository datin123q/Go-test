package domain

import (
	"context"
	"encoding/csv"
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type ASSETS_STATUS string

const (
	ASSETS_STATUS_CODE               = "ASSETS_STATUS"
	ASSET1             ASSETS_STATUS = "ASSET1" // 150万円未満
	ASSET2             ASSETS_STATUS = "ASSET2" // 150万円以上～300万円未満
	ASSET3             ASSETS_STATUS = "ASSET3" // 300万円以上～500万円未満
	ASSET4             ASSETS_STATUS = "ASSET4" // 500万円以上～700万円未満
	ASSET5             ASSETS_STATUS = "ASSET5" // 700万円以上～1,000万円未満
	ASSET6             ASSETS_STATUS = "ASSET6" // 1,000万円以上～3,000万円未満
	ASSET7             ASSETS_STATUS = "ASSET7" // 3,000万円以上
)

type assetStatusDescription struct {
	english string
	japan   string
}

var assetStatusDescriptions = map[ASSETS_STATUS]assetStatusDescription{
	ASSET1: {"Less than 1.5 million yen", "150万円未満"},
	ASSET2: {"1.5 million yen ~ 3 million yen", "150万円以上 ~ 300万円未満"},
	ASSET3: {"3 million yen ~ 5 million yen", "300万円以上 ~ 500万円未満"},
	ASSET4: {"5 million yen ~ 7 million yen", "500万円以上 ~ 700万円未満"},
	ASSET5: {"7 million yen ~ 10 million yen", "700万円以上 ~ 1,000万円未満"},
	ASSET6: {"10 million yen ~ 30 million yen", "1,000万円以上 ~ 3,000万円未満"},
	ASSET7: {"More than 30 million yen", "3,000万円以上"},
}

func (a ASSETS_STATUS) Description() string {
	if desc, ok := assetStatusDescriptions[a]; ok {
		return desc.english
	}
	return "No description available"
}

func (a ASSETS_STATUS) DescriptionJapan() string {
	if desc, ok := assetStatusDescriptions[a]; ok {
		return desc.japan
	}
	return "説明なし"
}

func (PrivateProfile) GetAssetsCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(ASSET1), Name: i18nutils.Translate("commoncode.ASSETS_STATUS.ASSET1", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ASSET2), Name: i18nutils.Translate("commoncode.ASSETS_STATUS.ASSET2", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ASSET3), Name: i18nutils.Translate("commoncode.ASSETS_STATUS.ASSET3", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ASSET4), Name: i18nutils.Translate("commoncode.ASSETS_STATUS.ASSET4", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ASSET5), Name: i18nutils.Translate("commoncode.ASSETS_STATUS.ASSET5", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ASSET6), Name: i18nutils.Translate("commoncode.ASSETS_STATUS.ASSET6", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ASSET7), Name: i18nutils.Translate("commoncode.ASSETS_STATUS.ASSET7", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: ASSETS_STATUS_CODE,
		Codes:     codes,
	}
}

type ANNUALINCOME_STATUS string

const (
	ANNUALINCOME_STATUS_CODE                     = "ANNUALINCOME_STATUS"
	ANNUALINCOME1            ANNUALINCOME_STATUS = "ANNUALINCOME1" // 300万円未満
	ANNUALINCOME2            ANNUALINCOME_STATUS = "ANNUALINCOME2" // 300万円以上～500万円未満
	ANNUALINCOME3            ANNUALINCOME_STATUS = "ANNUALINCOME3" // 500万円以上～700万円未満
	ANNUALINCOME4            ANNUALINCOME_STATUS = "ANNUALINCOME4" // 700万円以上～1,000万円未満
	ANNUALINCOME5            ANNUALINCOME_STATUS = "ANNUALINCOME5" // 1,000万円以上～3,000万円未満
	ANNUALINCOME6            ANNUALINCOME_STATUS = "ANNUALINCOME6" // 3,000万円以上
)

type annualIncomeStatusDescription struct {
	english string
	japan   string
}

var annualIncomeStatusDescriptions = map[ANNUALINCOME_STATUS]annualIncomeStatusDescription{
	ANNUALINCOME1: {"Less than 3 million yen", "300万円未満"},
	ANNUALINCOME2: {"3 million yen ~ 5 million yen", "300万円以上 ~ 500万円未満"},
	ANNUALINCOME3: {"5 million yen ~ 7 million yen", "500万円以上 ~ 700万円未満"},
	ANNUALINCOME4: {"7 million yen ~ 10 million yen", "700万円以上 ~ 1,000万円未満"},
	ANNUALINCOME5: {"10 million yen ~ 30 million yen", "1,000万円以上 ~ 3,000万円未満"},
	ANNUALINCOME6: {"More than 30 million yen", "3,000万円以上"},
}

func (a ANNUALINCOME_STATUS) Description() string {
	if desc, ok := annualIncomeStatusDescriptions[a]; ok {
		return desc.english
	}
	return "No description available"
}

func (a ANNUALINCOME_STATUS) DescriptionJapan() string {
	if desc, ok := annualIncomeStatusDescriptions[a]; ok {
		return desc.japan
	}
	return "説明なし"
}

func (PrivateProfile) GetsAnnualincomeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(ANNUALINCOME1), Name: i18nutils.Translate("commoncode.ANNUALINCOME_STATUS.ANNUALINCOME1", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ANNUALINCOME2), Name: i18nutils.Translate("commoncode.ANNUALINCOME_STATUS.ANNUALINCOME2", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ANNUALINCOME3), Name: i18nutils.Translate("commoncode.ANNUALINCOME_STATUS.ANNUALINCOME3", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ANNUALINCOME4), Name: i18nutils.Translate("commoncode.ANNUALINCOME_STATUS.ANNUALINCOME4", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ANNUALINCOME5), Name: i18nutils.Translate("commoncode.ANNUALINCOME_STATUS.ANNUALINCOME5", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ANNUALINCOME6), Name: i18nutils.Translate("commoncode.ANNUALINCOME_STATUS.ANNUALINCOME6", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: ANNUALINCOME_STATUS_CODE,
		Codes:     codes,
	}
}

type GENDER string

const (
	GENDER_CODE        = "GENDER"
	MALE        GENDER = "MALE"
	FEMALE      GENDER = "FEMALE"
)

type genderDescription struct {
	english string
	japan   string
}

var genderDescriptions = map[GENDER]genderDescription{
	MALE:   {"Male", "男性"},
	FEMALE: {"Female", "女性"},
}

func (g GENDER) Description() string {
	if desc, ok := genderDescriptions[g]; ok {
		return desc.english
	}
	return "No description available"
}

func (g GENDER) DescriptionJapan() string {
	if desc, ok := genderDescriptions[g]; ok {
		return desc.japan
	}
	return "説明なし"
}

func (PrivateProfile) GetGenderCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(MALE), Name: i18nutils.Translate("commoncode.GENDER.MALE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FEMALE), Name: i18nutils.Translate("commoncode.GENDER.FEMALE", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: GENDER_CODE,
		Codes:     codes,
	}
}

type PrivateProfile struct {
	Id                      int64               `json:"-"`
	UserId                  int64               `json:"user_id" xorm:"index unique"`
	Birthday                jsonutils.JSONTime  `json:"birthday" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PERSONAL_BIRTHDAY" swagger:"desc(생년월일)"`
	FirstName               string              `json:"first_name" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_NAME" swagger:"desc(이름)" `
	FamilyName              string              `json:"family_name"  i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_FAMILY_NAME" swagger:"desc(이름 성)" `
	FirstNameKana           string              `json:"first_name_kana" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_NAME2" swagger:"desc(이름 후리가나)" `
	FamilyNameKana          string              `json:"family_name_kana" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_FAMILY_NAME2" swagger:"desc(이름 성 후리가나)" `
	Asset                   ASSETS_STATUS       `json:"asset" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ASSET" swagger:"desc(자산 공통코드 : [ASSETS_STATUS])"`
	AnnualIncome            ANNUALINCOME_STATUS `json:"annual_income" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ANNUAL_INCOME" swagger:"desc(연봉 공통코드 : [ANNUALINCOME_STATUS])"`
	IsProInvestAsset        *bool               `json:"is_pro_invest_asset" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PRO_INVEST_ASSET" swagger:"desc(전문투자 자산 보유 여부)"`
	IsProInvestFinanceAsset *bool               `json:"is_pro_invest_finance_asset" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PRO_INVEST_FINANCE_ASSET" swagger:"desc(전문투자 금융자산 보유 여부)"`
	Nationality             NATIONALITY_TYPE    `json:"nationality" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_NATIONALITY" swagger:"desc(국적 공통 코드: [NATIONALITY_TYPE])"`
	Job                     USER_JOB_TYPE       `json:"job" swagger:"desc(직업)"`
	Gender                  GENDER              `json:"gender" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_GENDER" swagger:"desc(성별: [GENDER])"`
	RequestId               string              `json:"request_id" swagger:"desc(회원 가입 요청 시 발급 받은 ID)"`
	Created                 jsonutils.JSONTime  `json:"-" xorm:"created"`
	Updated                 jsonutils.JSONTime  `json:"-" xorm:"updated"`
	Deleted                 jsonutils.JSONTime  `json:"-" xorm:"deleted"`
}

func (p PrivateProfile) TableName() string {
	return "private_profile"
}

func (p PrivateProfile) SelectAll() string {
	return p.TableName() + ".*"
}

// PrivateProfileを変更する場合は、対応するクエリを変更する必要があります
const PRIVATE_USER_COLUMN_STRING = "private_profile.user_id as user_id, private_profile.birthday as birthday," +
	"private_profile.first_name as first_name, private_profile.family_name as family_name, private_profile.first_name_kana," +
	"private_profile.family_name_kana as family_name_kana, private_profile.asset as asset,private_profile.annual_income as annual_income," +
	"private_profile.is_pro_invest_asset as is_pro_invest_asset, private_profile.is_pro_invest_finance_asset as is_pro_invest_finance_asset," +
	"private_profile.nationality as nationality, private_profile.job as job, private_profile.gender as gender, private_profile.request_id as request_id"

type PrivateUserWithRoles struct {
	PrivateUser `xorm:"extends"`
	RolesString string `json:"-" xorm:"roles_string <-"`
}

type PrivateUsersWithRoles []PrivateUserWithRoles

func (u PrivateUsersWithRoles) First() PrivateUserWithRoles {
	if u.IsEmpty() {
		return PrivateUserWithRoles{}
	}

	return u[0]
}

func (u PrivateUsersWithRoles) IsEmpty() bool {
	return len(u) == 0
}

type PrivateUser struct {
	User           `xorm:"extends"`
	PrivateProfile `xorm:"extends"`
}

// 会員登録
func (u *PrivateUser) SignUp(ctx context.Context) (*AuthToken, error) {
	if _, err := u.Create(ctx); err != nil {
		return nil, err
	}

	token, err := AuthToken{}.Create(ctx, u.User)
	if err != nil {
		return nil, errors.New("Fail Create Token")
	}
	return token, nil
}

// プロファイルモデルを作成
func (p *PrivateProfile) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(p)
}

func (PrivateUser) GetAll(ctx context.Context, id, name, email, investorType, auditStatus, requestId *string, sortBy, order string, offset, limit int, startDate, endDate string, profileCond ProfileSearchCondition) (int64, []PrivateUser, error) {

	query := factory.DB(ctx).Table("user").
		Select("user.* ,"+PRIVATE_USER_COLUMN_STRING).
		Join("INNER", "private_profile", "private_profile.user_id = user.id").
		Where("user.user_type = ?", PRIVATE)

	if err := modelhelper.SetSortOrder(query, sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "user", "id", id); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(query, "user", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(query, "user", "email", email); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user", "investor_type", investorType); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "user", "audit_status", auditStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "private_profile", "request_id", requestId); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(query, "user.sign_up_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	// フリガナ・生年月日(MIR-534)。CSV もこの関数を通るので一覧と同じ条件になる。
	if err := applyPrivateProfileSearch(query, profileCond); err != nil {
		return 0, nil, err
	}

	var userList []PrivateUser
	var totalCount int64
	// Noに行カウンター適用が必要
	totalCount, err := query.Limit(limit, offset).FindAndCount(&userList)

	if err != nil {
		return 0, nil, err
	}

	return totalCount, userList, nil
}

func (PrivateUser) GetById(ctx context.Context, id int64) (*PrivateUser, error) {
	var user PrivateUser
	query := factory.DB(ctx).Table("user").Select("user.*, private_profile.*").Join("INNER", "private_profile", "private_profile.user_id = user.id")
	if has, err := query.Where("user.id = ?", id).Get(&user); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &user, nil
}

func (PrivateUser) GetByEmail(ctx context.Context, email string) (*PrivateUser, error) {
	var user PrivateUser
	query := factory.DB(ctx).Table("user").Select("user.*, private_profile.*").Join("INNER", "private_profile", "private_profile.user_id = user.id")

	if has, err := query.Where("user.email = ?", email).Get(&user); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &user, nil
}

func (PrivateUser) GetDefaultInfoById(ctx context.Context, id int64) (*PrivateUser, error) {
	var user PrivateUser
	query := factory.DB(ctx).Table("user").Select("user.*, private_profile.*").Join("INNER", "private_profile", "private_profile.user_id = user.id")
	if has, err := query.Where("user.id = ?", id).Get(&user); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	user.Roles = []string{string(ROLE_USER)}

	return &user, nil
}

func (up PrivateUser) Create(ctx context.Context) (cnt int64, err error) {
	up.AuditStatus = WAIT_AUDIT
	up.UserType = PRIVATE

	if cnt, err = up.User.Create(ctx); err != nil {
		return 0, err
	}
	return cnt, err

}
func (u *PrivateUser) Update(ctx context.Context) (err error) {
	q := factory.DB(ctx).UseBool()
	if _, err := q.ID(u.User.Id).Update(&u.User); err != nil {
		return err
	}

	q2 := factory.DB(ctx).UseBool()
	if _, err := q2.ID(u.PrivateProfile.Id).Update(&u.PrivateProfile); err != nil {
		return err
	}

	return nil
}

// UpdateEkycResult は、eKYCの受信処理(jobs-scheduler・ekycAfterProcess)がWebhook結果を
// 反映するときに使う。is_ekyc・signup_status・audit_status・hold_user_check_codeの
// 4項目だけをUPDATE文へ含め(Cols指定)、Userのそれ以外の列・PrivateProfileには触れない。
// 以前はUpdate(ctx)で会員の行を丸ごと書き戻しており、受信処理と同じ秒に別の処理
// (会員のログイン・本人情報の変更等)が行われると、互いの結果を消し合っていた
// (MIR-1373/1374)。
func (u *PrivateUser) UpdateEkycResult(ctx context.Context) error {
	_, err := factory.DB(ctx).UseBool().ID(u.User.Id).
		Cols("is_ekyc", "signup_status", "audit_status", "hold_user_check_code").
		Update(&u.User)
	return err
}

func (PrivateUser) AntisocialWaitAndReturnCsv(items []PrivateUser) [][]string {

	type private struct {
		Name     string `json:"name" i18n:"model.user.name"`            // 名前
		Birthday string `json:"birthday" i18n:"model.private.birthday"` // 生年月日
	}

	var privates []private
	for _, item := range items {
		p := private{
			Name:     item.Name,
			Birthday: item.Birthday.ToFormat("20060102"),
		}

		privates = append(privates, p)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&private{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range privates {

		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (PrivateUser) GetDashboardList(ctx context.Context, offset, limit int) (int64, []PrivateUser, error) {
	query := factory.DB(ctx).Table("user").Select("user.*, private_profile.*").Join("INNER", "private_profile", "private_profile.user_id = user.id")

	var userList []PrivateUser
	var totalCount int64
	totalCount, err := query.Where("user.audit_status=?", WAIT_AUDIT).Desc("user.sign_up_date").Limit(limit, offset).FindAndCount(&userList)
	if err != nil {
		return 0, nil, err
	}
	if userList == nil {
		userList = make([]PrivateUser, 0)
	}
	return totalCount, userList, nil
}

func (PrivateUser) GetAntisocialCheckUsers(ctx context.Context, offset, limit int) (int64, []PrivateUser, error) {
	query := factory.DB(ctx).Table("user").Select("user.*, private_profile.*").Join("INNER", "private_profile", "private_profile.user_id = user.id")

	var userList []PrivateUser
	var totalCount int64
	totalCount, err := query.Where("user.audit_status=?", CHECK_ANTISOCIAL_AUDIT).Desc("user.sign_up_date").Limit(limit, offset).FindAndCount(&userList)
	if err != nil {
		return 0, nil, err
	}
	if userList == nil {
		userList = make([]PrivateUser, 0)
	}
	return totalCount, userList, nil
}
func (PrivateUser) AntiSocialWaitUserList(ctx context.Context, w http.ResponseWriter, searchkeys, searchvalues, sortby, order string, offset, limit int, startDate, endDate string, processId int64) error {
	query := factory.DB(ctx).Table("user").Select("user.*,"+PRIVATE_USER_COLUMN_STRING).Join("INNER", "private_profile", "private_profile.user_id = user.id").Desc("user.created")

	var userList []PrivateUser

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	// MIR-1232: 管理画面はこのAPIへsearchKeysを一度も送っていない(確認済み)ため、
	// 許可する列名を空にしている。エラーは握りつぶさず呼び出し元(コントローラー)まで
	// 返し、フィルタが無視されたまま全件が返ることを防ぐ。
	if err := modelhelper.SearchQuery(query, "", searchkeys, searchvalues, map[string]struct{}{}); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := modelhelper.SetBetweenDate(query, "user.sign_up_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	_, err := query.Where("audit_status = ?", WAIT_ANTISOCIAL_AUDIT).Limit(limit, offset).FindAndCount(&userList)

	if err != nil {
		return nil
	}

	records := PrivateUser{}.AntisocialWaitAndReturnCsv(userList)

	for _, item := range userList {
		// before := item

		item.AuditStatus = CHECK_ANTISOCIAL_AUDIT
		item.AntiResearchDate = jsonutils.JSONTime(time.Now())
		if item.Update(ctx) != nil {
			return err
		}

		// _, err = UserInfoHistory{}.Create(ctx, *before, &item, processId)
		if err != nil {
			return err
		}
	}
	query.Commit()

	csvutils.WriteHeader(w, "energyShares_anti_member_list.csv")

	return csv.NewWriter(w).WriteAll(records)
}

type DuplicateUserResult struct {
	CheckedUserID   int64
	DuplicateUserID int64
	DuplicateReason string
}

// FindDuplicates は指定ユーザーと重複する可能性のある個人ユーザーを検索する
func (PrivateUser) FindDuplicates(ctx context.Context, user *PrivateUser) ([]DuplicateUserResult, error) {
	var results []DuplicateUserResult

	// 1. 同姓同名かつ誕生日が同じ
	if user.FirstName != "" && user.FamilyName != "" && !time.Time(user.Birthday).IsZero() {
		var matched []PrivateUser
		err := factory.DB(ctx).Table("user").
			Select("user.id").
			Join("INNER", "private_profile", "private_profile.user_id = user.id").
			Where("user.user_type = ?", PRIVATE).
			And("user.id != ?", user.User.Id).
			And("private_profile.first_name = ?", user.FirstName).
			And("private_profile.family_name = ?", user.FamilyName).
			And("private_profile.birthday = ?", time.Time(user.Birthday)).
			Find(&matched)
		if err != nil {
			return nil, err
		}
		for _, m := range matched {
			results = append(results, DuplicateUserResult{
				CheckedUserID:   user.User.Id,
				DuplicateUserID: m.User.Id,
				DuplicateReason: "同姓同名かつ誕生日が同じ",
			})
		}
	}

	// 2. 同じ電話番号
	if user.MobilePhone != "" {
		var matched []User
		err := factory.DB(ctx).Table("user").
			Select("user.id").
			Where("user.user_type = ?", PRIVATE).
			And("user.id != ?", user.User.Id).
			And("user.mobile_phone = ?", user.MobilePhone).
			Find(&matched)
		if err != nil {
			return nil, err
		}
		for _, m := range matched {
			results = append(results, DuplicateUserResult{
				CheckedUserID:   user.User.Id,
				DuplicateUserID: m.Id,
				DuplicateReason: "同じ電話番号",
			})
		}
	}

	// 3. 出金先口座が同じ
	if user.BankCode != nil && *user.BankCode != "" &&
		user.BranchCode != nil && *user.BranchCode != "" &&
		user.BankAccount != "" && string(user.DepositType) != "" {
		var matched []User
		err := factory.DB(ctx).Table("user").
			Select("user.id").
			Where("user.user_type = ?", PRIVATE).
			And("user.id != ?", user.User.Id).
			And("user.bank_code = ?", *user.BankCode).
			And("user.branch_code = ?", *user.BranchCode).
			And("user.bank_account = ?", user.BankAccount).
			And("user.deposit_type = ?", string(user.DepositType)).
			Find(&matched)
		if err != nil {
			return nil, err
		}
		for _, m := range matched {
			results = append(results, DuplicateUserResult{
				CheckedUserID:   user.User.Id,
				DuplicateUserID: m.Id,
				DuplicateReason: "出金先口座が同じ",
			})
		}
	}

	return results, nil
}

func IsPrivateSuitability(asset ASSETS_STATUS, annualincome ANNUALINCOME_STATUS) bool {
	if asset == ASSET1 && annualincome == ANNUALINCOME1 {
		return false
	}
	return true
}
