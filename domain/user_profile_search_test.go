package domain

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// setupProfileSearchTest は会員一覧(個人・法人)の検索を通すための sqlite を用意する(MIR-534)。
// user・private_profile・corporate_profile は既存テストにならって Sync2 で作る。
//
// 生年月日はアプリと同じ経路(jsonutils.ParseJSONTimeJST → xorm の Insert)で保存する。
// 本番と同じく「JST の日付 00:00 を UTC に直した日時」(1985-03-12 生まれ → 1985-03-11 15:00:00)
// で入るため、日付の境界がずれていればここで落ちる。
func setupProfileSearchTest(t *testing.T) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	require.NoError(t, engine.Sync2(new(User), new(PrivateProfile), new(CorporateProfile)))

	insertUser := func(id int64, name string, userType USER_TYPE) {
		_, err := engine.Insert(&User{
			Id: id, Uid: fmt.Sprintf("uid-%d", id), CommonId: fmt.Sprintf("common-%d", id),
			Email: fmt.Sprintf("u%d@example.com", id), Password: "x", Salt: "x",
			Name: name, UserType: userType,
		})
		require.NoError(t, err)
	}
	birthday := func(date string) jsonutils.JSONTime {
		if date == "" {
			return jsonutils.JSONTime{}
		}
		b, err := jsonutils.ParseJSONTimeJST(time.DateOnly, date)
		require.NoError(t, err)
		return b
	}

	// 個人会員
	privates := []struct {
		id                  int64
		name, family, first string
		birthday            string
		deleted             bool
	}{
		{1, "山田太郎", "ヤマダ", "タロウ", "1985-03-12", false},
		{2, "佐藤花子", "サトウ", "ハナコ", "1985-03-13", false},
		{3, "鈴木一郎", "スズキ", "イチロウ", "1990-01-01", false},
		// プロフィールだけ論理削除済み。どの条件でも出てはいけない
		{4, "山本次郎", "ヤマモト", "ジロウ", "1985-03-12", true},
		// 日本の夏時刻(1948〜1951年)の期間に生まれた人。JST の 00:00 が UTC+10 なので 14:00:00 で入る
		{5, "田中誠", "タナカ", "マコト", "1950-07-01", false},
		{6, "野村亜希", "ノムラ", "アキ", "", false}, // 生年月日は下で生SQLの旧形式を入れる
	}
	for _, p := range privates {
		insertUser(p.id, p.name, PRIVATE)
		_, err := engine.Insert(&PrivateProfile{
			Id: p.id, UserId: p.id, FamilyName: "x", FirstName: "x",
			FamilyNameKana: p.family, FirstNameKana: p.first,
			Birthday: birthday(p.birthday),
		})
		require.NoError(t, err)
		if p.deleted {
			_, err = engine.Exec(`UPDATE private_profile SET deleted = '2026-01-01 00:00:00' WHERE id = ?`, p.id)
			require.NoError(t, err)
		}
	}
	// 本番には JST の日付がそのまま 00:00:00 で入った古い行もある。その日付で当たること
	_, err = engine.Exec(`UPDATE private_profile SET birthday = '1970-07-07 00:00:00' WHERE id = 6`)
	require.NoError(t, err)

	// 法人会員。代表者と取引担当者でフリガナを変えておき、どちらからでも当たることを見る。
	// 13 は取引担当者の生年月日が未設定(ゼロ値)
	corporates := []struct {
		id                        int64
		name, repFamily, repFirst string
		mgrFamily, mgrFirst       string
		managerBirthday           string
	}{
		{11, "株式会社アルファ", "タナカ", "マコト", "キムラ", "サトシ", "1980-05-05"},
		{12, "株式会社ベータ", "イノウエ", "ケン", "タナベ", "ユミ", "1975-12-31"},
		{13, "株式会社ガンマ", "オカダ", "ジュン", "ハヤシ", "ミキ", ""},
	}
	for _, c := range corporates {
		insertUser(c.id, c.name, CORPORATE)
		_, err := engine.Insert(&CorporateProfile{
			Id: c.id, UserId: c.id, CorpNo: fmt.Sprintf("corp-%d", c.id),
			RepresentativeFamilyNameKana: c.repFamily, RepresentativeFirstNameKana: c.repFirst,
			ManagerFamilyNameKana: lo.ToPtr(c.mgrFamily), ManagerFirstNameKana: lo.ToPtr(c.mgrFirst),
			ManagerBirthday: birthday(c.managerBirthday),
			// 以下は NOT NULL を満たすためのダミー値
			ManagerFamilyName: lo.ToPtr("x"), ManagerFirstName: lo.ToPtr("x"),
			ManagerEmail: lo.ToPtr("x"), ManagerMobilePhone: lo.ToPtr("x"),
			Department: lo.ToPtr("x"), Position: lo.ToPtr("x"),
		})
		require.NoError(t, err)
	}

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}

func privateIds(t *testing.T, ctx context.Context, name *string, cond ProfileSearchCondition) []int64 {
	t.Helper()
	_, users, err := PrivateUser{}.GetAll(ctx, nil, name, nil, nil, nil, nil, "user.id", "asc", 0, 100, "", "", cond)
	require.NoError(t, err)
	return lo.Map(users, func(u PrivateUser, _ int) int64 { return u.User.Id })
}

func corporateIds(t *testing.T, ctx context.Context, cond ProfileSearchCondition) []int64 {
	t.Helper()
	_, users, err := CorporateUser{}.GetAll(ctx, nil, nil, nil, nil, nil, "user.id", "asc", 0, 100, "", "", cond)
	require.NoError(t, err)
	return lo.Map(users, func(u CorporateUser, _ int) int64 { return u.User.Id })
}

// 前提の確認: 生年月日が本番と同じ形(JST 00:00 を UTC にした 15:00:00)で入っていること
func TestProfileSearch_BirthdayStoredAsUTC(t *testing.T) {
	ctx := setupProfileSearchTest(t)
	var stored string
	_, err := factory.DB(ctx).SQL(`SELECT birthday FROM private_profile WHERE id = 1`).Get(&stored)
	require.NoError(t, err)
	assert.Contains(t, stored, "1985-03-11")
	assert.Contains(t, stored, "15:00:00")

	// 夏時刻(UTC+10)の期間は JST の 00:00 が 14:00:00 になる。本番の 15:00:00 でない行の1つ目の形
	_, err = factory.DB(ctx).SQL(`SELECT birthday FROM private_profile WHERE id = 5`).Get(&stored)
	require.NoError(t, err)
	assert.Contains(t, stored, "1950-06-30")
	assert.Contains(t, stored, "14:00:00")

	// 2つ目の形: JST の日付が 00:00:00 のまま入った古い行
	_, err = factory.DB(ctx).SQL(`SELECT birthday FROM private_profile WHERE id = 6`).Get(&stored)
	require.NoError(t, err)
	assert.Contains(t, stored, "1970-07-07")
	assert.Contains(t, stored, "00:00:00")
}

func TestPrivateUser_GetAll_ProfileSearch(t *testing.T) {
	ctx := setupProfileSearchTest(t)

	tests := []struct {
		name string
		user *string
		cond ProfileSearchCondition
		want []int64
	}{
		{name: "条件なしでは削除済みプロフィールだけが除かれる", want: []int64{1, 2, 3, 5, 6}},
		{name: "姓フリガナの部分一致", cond: ProfileSearchCondition{Kana: lo.ToPtr("ヤマ")}, want: []int64{1}},
		{name: "名フリガナの部分一致", cond: ProfileSearchCondition{Kana: lo.ToPtr("ハナ")}, want: []int64{2}},
		{name: "前後の空白は無視する", cond: ProfileSearchCondition{Kana: lo.ToPtr("　サトウ ")}, want: []int64{2}},
		{name: "空のフリガナは条件なし扱い", cond: ProfileSearchCondition{Kana: lo.ToPtr("")}, want: []int64{1, 2, 3, 5, 6}},
		{name: "ひらがなの入力でカタカナに当たる", cond: ProfileSearchCondition{Kana: lo.ToPtr("やまだ")}, want: []int64{1}},
		{name: "姓名をつなげた入力で当たる", cond: ProfileSearchCondition{Kana: lo.ToPtr("ヤマダタロウ")}, want: []int64{1}},
		{name: "姓名の間に空白があっても当たる", cond: ProfileSearchCondition{Kana: lo.ToPtr("サトウ ハナコ")}, want: []int64{2}},
		{name: "半角カタカナの入力で全角に当たる", cond: ProfileSearchCondition{Kana: lo.ToPtr("ﾔﾏﾀﾞ")}, want: []int64{1}},
		{name: "_ はワイルドカードにならない", cond: ProfileSearchCondition{Kana: lo.ToPtr("タ_")}, want: []int64{}},
		{name: "% はワイルドカードにならない", cond: ProfileSearchCondition{Kana: lo.ToPtr("%")}, want: []int64{}},
		{name: "フリガナと氏名は AND で結ぶ(両方当たる)", user: lo.ToPtr("山田"), cond: ProfileSearchCondition{Kana: lo.ToPtr("タロウ")}, want: []int64{1}},
		{name: "フリガナと氏名は AND で結ぶ(片方だけ当たる)", user: lo.ToPtr("佐藤"), cond: ProfileSearchCondition{Kana: lo.ToPtr("タロウ")}, want: []int64{}},
		{name: "同じ日を入れると完全一致(翌日生まれは出ない)", cond: ProfileSearchCondition{BirthdayFrom: "1985-03-12", BirthdayTo: "1985-03-12"}, want: []int64{1}},
		{name: "翌日の完全一致", cond: ProfileSearchCondition{BirthdayFrom: "1985-03-13", BirthdayTo: "1985-03-13"}, want: []int64{2}},
		{name: "生年月日 from のみはその日以降", cond: ProfileSearchCondition{BirthdayFrom: "1985-03-13"}, want: []int64{2, 3}},
		{name: "生年月日 to のみはその日以前", cond: ProfileSearchCondition{BirthdayTo: "1985-03-12"}, want: []int64{1, 5, 6}},
		{name: "範囲の両端を含む", cond: ProfileSearchCondition{BirthdayFrom: "1985-03-12", BirthdayTo: "1985-03-13"}, want: []int64{1, 2}},
		{name: "夏時刻の期間に生まれた人(14:00:00 で保存)も完全一致で当たる", cond: ProfileSearchCondition{BirthdayFrom: "1950-07-01", BirthdayTo: "1950-07-01"}, want: []int64{5}},
		{name: "夏時刻の期間に生まれた人は前日の完全一致では出ない", cond: ProfileSearchCondition{BirthdayFrom: "1950-06-30", BirthdayTo: "1950-06-30"}, want: []int64{}},
		{name: "00:00:00 で入った古い行もその日付の完全一致で当たる", cond: ProfileSearchCondition{BirthdayFrom: "1970-07-07", BirthdayTo: "1970-07-07"}, want: []int64{6}},
		{name: "00:00:00 で入った古い行は前日の完全一致では出ない", cond: ProfileSearchCondition{BirthdayFrom: "1970-07-06", BirthdayTo: "1970-07-06"}, want: []int64{}},
		{name: "00:00:00 で入った古い行は翌日の完全一致では出ない", cond: ProfileSearchCondition{BirthdayFrom: "1970-07-08", BirthdayTo: "1970-07-08"}, want: []int64{}},
		{name: "フリガナと生年月日の組み合わせ", cond: ProfileSearchCondition{Kana: lo.ToPtr("ウ"), BirthdayFrom: "1985-03-13"}, want: []int64{2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := privateIds(t, ctx, tt.user, tt.cond)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestCorporateUser_GetAll_ProfileSearch(t *testing.T) {
	ctx := setupProfileSearchTest(t)

	tests := []struct {
		name string
		cond ProfileSearchCondition
		want []int64
	}{
		{name: "条件なし", want: []int64{11, 12, 13}},
		{name: "代表者の姓フリガナで当たる", cond: ProfileSearchCondition{Kana: lo.ToPtr("タナカ")}, want: []int64{11}},
		{name: "代表者の名フリガナで当たる", cond: ProfileSearchCondition{Kana: lo.ToPtr("ケン")}, want: []int64{12}},
		{name: "取引担当者の姓フリガナで当たる", cond: ProfileSearchCondition{Kana: lo.ToPtr("キムラ")}, want: []int64{11}},
		{name: "取引担当者の名フリガナで当たる", cond: ProfileSearchCondition{Kana: lo.ToPtr("ユミ")}, want: []int64{12}},
		{name: "代表者・取引担当者の両方を横断する", cond: ProfileSearchCondition{Kana: lo.ToPtr("タナ")}, want: []int64{11, 12}},
		{name: "ひらがなで取引担当者に当たる", cond: ProfileSearchCondition{Kana: lo.ToPtr("きむら")}, want: []int64{11}},
		{name: "代表者の姓名で当たる", cond: ProfileSearchCondition{Kana: lo.ToPtr("イノウエケン")}, want: []int64{12}},
		{name: "取引担当者の姓名(空白あり)で当たる", cond: ProfileSearchCondition{Kana: lo.ToPtr("キムラ サトシ")}, want: []int64{11}},
		{name: "代表者の姓と取引担当者の名を混ぜても当たらない", cond: ProfileSearchCondition{Kana: lo.ToPtr("タナカサトシ")}, want: []int64{}},
		{name: "_ はワイルドカードにならない(タナカ・タナベに当たらない)", cond: ProfileSearchCondition{Kana: lo.ToPtr("タナ_")}, want: []int64{}},
		{name: "取引担当者の生年月日で完全一致", cond: ProfileSearchCondition{BirthdayFrom: "1975-12-31", BirthdayTo: "1975-12-31"}, want: []int64{12}},
		{name: "範囲に誰もいなければ0件", cond: ProfileSearchCondition{BirthdayFrom: "1976-01-01", BirthdayTo: "1979-12-31"}, want: []int64{}},
		{name: "to のみでも生年月日未設定の法人は出ない", cond: ProfileSearchCondition{BirthdayTo: "2000-01-01"}, want: []int64{11, 12}},
		{name: "from のみでも生年月日未設定の法人は出ない", cond: ProfileSearchCondition{BirthdayFrom: "1900-01-01"}, want: []int64{11, 12}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := corporateIds(t, ctx, tt.cond)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestProfileSearchCondition_Validate(t *testing.T) {
	tests := []struct {
		name string
		cond ProfileSearchCondition
		want error
	}{
		{name: "未指定", cond: ProfileSearchCondition{}, want: nil},
		{name: "同じ日", cond: ProfileSearchCondition{BirthdayFrom: "1985-03-12", BirthdayTo: "1985-03-12"}, want: nil},
		{name: "逆順はエラー", cond: ProfileSearchCondition{BirthdayFrom: "1985-03-13", BirthdayTo: "1985-03-12"}, want: ErrInvalidBirthdayRange},
		{name: "形式違い(スラッシュ)", cond: ProfileSearchCondition{BirthdayFrom: "1985/03/12"}, want: ErrInvalidBirthdayFormat},
		{name: "存在しない日付", cond: ProfileSearchCondition{BirthdayTo: "1985-02-30"}, want: ErrInvalidBirthdayFormat},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, tt.cond.Validate(), tt.want)
		})
	}
}

// 不正な条件はクエリを投げずにエラーで返す(一覧が条件を無視して全件出すことはない)
func TestPrivateUser_GetAll_InvalidBirthdayRange(t *testing.T) {
	ctx := setupProfileSearchTest(t)
	_, _, err := PrivateUser{}.GetAll(ctx, nil, nil, nil, nil, nil, nil, "", "", 0, 100, "", "",
		ProfileSearchCondition{BirthdayFrom: "1985-03-13", BirthdayTo: "1985-03-12"})
	require.ErrorIs(t, err, ErrInvalidBirthdayRange)
}
