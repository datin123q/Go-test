package domain

import (
	"errors"
	"time"

	"github.com/samber/lo"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

// ProfileSearchCondition は会員管理(個人・法人)の一覧・件数・CSV で使う、
// プロフィール側(private_profile / corporate_profile)の検索条件(MIR-534)。
// 一覧と CSV で結果が食い違わないよう、絞り込みはこのファイルの1か所にまとめる。
type ProfileSearchCondition struct {
	Kana         *string
	BirthdayFrom string // "YYYY-MM-DD"(JST の日付)
	BirthdayTo   string
}

var (
	ErrInvalidBirthdayFormat = errors.New("生年月日は YYYY-MM-DD 形式で指定してください")
	ErrInvalidBirthdayRange  = errors.New("生年月日の開始日が終了日より後になっています")
)

// profileSearchJST は生年月日の日付を解釈するタイムゾーン。
// 固定の UTC+9 ではなく Asia/Tokyo を使う。日本の夏時刻(1948〜1951年)の期間は UTC+10 で、
// その期間に生まれた人の生年月日は書き込み時(jsonutils.ParseJSONTimeJST)に 14:00:00 で保存されるため。
var profileSearchJST = lo.Must(time.LoadLocation("Asia/Tokyo"))

// birthdayFloor より前の生年月日は未設定(ゼロ値)とみなし、生年月日検索では常に除外する。
// 法人の取引担当者は生年月日が未設定のことがあり、下限を置かないと「まで」だけの検索で当たってしまう。
var birthdayFloor = time.Date(1900, 1, 1, 0, 0, 0, 0, profileSearchJST)

// Validate は生年月日の形式と前後関係を検査する。
// 逆順(から > まで)は入れ替えずにエラーにする。黙って入れ替えると、
// 入力ミスに気付かないまま別の範囲で検索した結果を見せてしまうため。
func (c ProfileSearchCondition) Validate() error {
	from, err := parseBirthday(c.BirthdayFrom)
	if err != nil {
		return err
	}
	to, err := parseBirthday(c.BirthdayTo)
	if err != nil {
		return err
	}
	if !from.IsZero() && !to.IsZero() && from.After(to) {
		return ErrInvalidBirthdayRange
	}
	return nil
}

// parseBirthday は "YYYY-MM-DD" を JST のその日の 00:00 として解釈する。
func parseBirthday(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.ParseInLocation(time.DateOnly, s, profileSearchJST)
	if err != nil {
		return time.Time{}, ErrInvalidBirthdayFormat
	}
	return t, nil
}

// 個人: 姓・名のフリガナと姓名(フルネーム) / 生年月日は birthday。
func applyPrivateProfileSearch(q xorm.Interface, c ProfileSearchCondition) error {
	return applyProfileSearch(q, c, "private_profile",
		[]modelhelper.KanaNameColumns{
			{Family: "family_name_kana", First: "first_name_kana"},
		},
		"birthday")
}

// 法人: 代表者・取引担当者それぞれの姓・名のフリガナと姓名(フルネーム)の OR / 生年月日は取引担当者(manager_birthday)。
func applyCorporateProfileSearch(q xorm.Interface, c ProfileSearchCondition) error {
	return applyProfileSearch(q, c, "corporate_profile",
		[]modelhelper.KanaNameColumns{
			{Family: "representative_family_name_kana", First: "representative_first_name_kana"},
			{Family: "manager_family_name_kana", First: "manager_first_name_kana"},
		},
		"manager_birthday")
}

func applyProfileSearch(q xorm.Interface, c ProfileSearchCondition, table string, kanaNames []modelhelper.KanaNameColumns, birthdayColumn string) error {
	if err := c.Validate(); err != nil {
		return err
	}
	session := q.(*xorm.Session)

	// xorm の論理削除の自動除外は FROM 側(user)にしか効かず、JOIN 先は見ない。
	session.And(table + ".deleted IS NULL")

	if err := modelhelper.SearchKanaLikeQuery(q, table, kanaNames, c.Kana); err != nil {
		return err
	}
	applyBirthdaySearch(session, table+"."+birthdayColumn, c.BirthdayFrom, c.BirthdayTo)
	return nil
}

// applyBirthdaySearch は生年月日の範囲条件を足す。
// 生年月日は JST の日付を UTC に直した日時で保存されている(例: 1985-03-12 生まれ → 1985-03-11 15:00:00)。
// そのため JST の日の境界(その日の 00:00 〜 翌日の 00:00)を、書き込み時と同じく
// DB のタイムゾーン(xorm の DatabaseTZ)に変換して半開区間で比べる。
// 文字列で渡すのは、書き込み時の xorm と同じ形式・同じタイムゾーンにそろえるため。
// 15:00:00 でない行(夏時刻の 14:00:00、古いデータの 00:00:00)も、JST のその日の範囲に入るので同じ日付で当たる。
func applyBirthdaySearch(session *xorm.Session, column, from, to string) {
	if from == "" && to == "" {
		return
	}
	dbTZ := session.Engine().DatabaseTZ
	format := func(t time.Time) string { return t.In(dbTZ).Format(time.DateTime) }

	lower := birthdayFloor
	if f, _ := parseBirthday(from); !f.IsZero() && f.After(lower) { // Validate 済み
		lower = f
	}
	session.And(column+" >= ?", format(lower))

	if to != "" {
		t, _ := parseBirthday(to) // Validate 済み
		session.And(column+" < ?", format(t.AddDate(0, 0, 1)))
	}
}
