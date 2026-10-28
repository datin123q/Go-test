package domain

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// setupAcceptAntiSocialCheckTestDBは、AcceptAntiSocialCheckがCORPORATE会員に
// 対して内部で呼ぶCorporateUser.GetByIdが必要とするテーブル一式(株主・本人確認
// 書類・バーチャル口座・反社チェック・口座)を用意する。いずれも空のままでよく、
// GetByIdの各Find/Getはゼロ件を正常系として返すため、テーブルの存在だけで足りる。
func setupAcceptAntiSocialCheckTestDB(t *testing.T) (context.Context, *xorm.Engine) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	require.NoError(t, engine.Sync2(
		new(User),
		new(CorporateProfile),
		new(CorporateShareholder),
		new(UserFile),
		new(UserVirtualAccount),
		new(UserAntisocialCheck),
		new(UserAccount),
	))

	return context.WithValue(context.Background(), factory.ConextDBType, engine), engine
}

// TestAcceptAntiSocialCheck_ClearsStaleApprovalDatesForCorporateはMIR-1252の
// 検証。①敵対的検証(ZOLOME移植時)で見つかった穴: 一度本登録を完了した法人会員が
// 管理画面でCHECK_ANTISOCIAL_AUDITへ差し戻され、反社チェックを再承認されると、
// AcceptAntiSocialCheckが新しい郵便コードを発行するが、古いCodeApprovalDate/
// CorporatePostalCodeApprovalDateをクリアしていなかった。これにより、新しい
// コードを片方だけ確認しただけで(もう片方は古い承認日が残っているため)
// CheckAllPostalCodeの完了判定を満たしてしまう迂回が残っていた。
//
// さらに、AcceptAntiSocialCheckの呼び出し元(apps/admin-api/controllers/
// user.go)は、このメソッドのあとuを使って別途user.Update(ctx)を呼ぶ。corp側
// だけ承認日をクリアしてcorp.Update(ctx)でNULLを書いても、そのあとのuser.
// Update(ctx)がuに残ったままの古い非ゼロ値を書き戻してしまう(xormは非ゼロ値の
// フィールドをMustCols無しでも自動的に更新対象へ含めるため)。この呼び出し元の
// 挙動をここで再現し、uのCodeApprovalDateもクリアされていることを検証する。
func TestAcceptAntiSocialCheck_ClearsStaleApprovalDatesForCorporate(t *testing.T) {
	ctx, engine := setupAcceptAntiSocialCheckTestDB(t)

	staleApproval := jsonutils.JSONTime(time.Now().Add(-24 * time.Hour))
	user := User{
		Uid:              "test-uid",
		UserType:         CORPORATE,
		PostalCode:       "1000001",
		AuditStatus:      CHECK_ANTISOCIAL_AUDIT,
		CodeApprovalDate: staleApproval,
	}
	_, err := engine.Insert(&user)
	require.NoError(t, err)

	profile := CorporateProfile{
		UserId:                          user.Id,
		CorpNo:                          "0000000000",
		CorpType:                        "test",
		EmployeeCount:                   "1",
		CapitalAmount:                   "1",
		Sales:                           "1",
		RepresentativeFamilyName:        "テスト",
		RepresentativeFirstName:         "テスト",
		RepresentativeFamilyNameKana:    "テスト",
		RepresentativeFirstNameKana:     "テスト",
		RepresentativeMobilePhone:       "00000000000",
		RepresentativeEmail:             "rep@example.com",
		ManagerFamilyName:               lo.ToPtr("テスト"),
		ManagerFirstName:                lo.ToPtr("テスト"),
		ManagerFamilyNameKana:           lo.ToPtr("テスト"),
		ManagerFirstNameKana:            lo.ToPtr("テスト"),
		ManagerEmail:                    lo.ToPtr("manager@example.com"),
		ManagerZipCode:                  "2000002",
		ManagerProvince:                 "テスト",
		ManagerCity:                     "テスト",
		ManagerAddress:                  "テスト",
		ManagerMobilePhone:              lo.ToPtr("00000000000"),
		ManagerBirthday:                 jsonutils.JSONTime(time.Now()),
		Department:                      lo.ToPtr("テスト"),
		Position:                        lo.ToPtr("テスト"),
		CorporatePostalCode:             "2000002",
		CorporatePostalCodeApprovalDate: staleApproval,
	}
	_, err = engine.Insert(&profile)
	require.NoError(t, err)

	// CorporateUser.GetById(AcceptAntiSocialCheckが内部で呼ぶ)は
	// UserAccount.GetByUserIdで会員の口座が無いとエラーを返すため、必須。
	_, err = engine.Insert(&UserAccount{UserId: user.Id, Balance: lo.ToPtr(int64(0))})
	require.NoError(t, err)

	require.NoError(t, user.AcceptAntiSocialCheck(ctx, jsonutils.JSONTime(time.Now())))

	require.True(t, user.CodeApprovalDate.IsZero(),
		"呼び出し元が使い続けるuの側でも承認日がクリアされているはず(呼び出し元の後続user.Update(ctx)が古い値を書き戻さないため)")

	// 呼び出し元(apps/admin-api/controllers/user.go)は、AcceptAntiSocialCheck
	// のあとuser.Update(ctx)を呼ぶ。ここで再現する。
	require.NoError(t, user.Update(ctx))

	var reloadedUser User
	ok, err := engine.Table("user").Where("id = ?", user.Id).Cols("code_approval_date").Get(&reloadedUser)
	require.NoError(t, err)
	require.True(t, ok)
	require.True(t, reloadedUser.CodeApprovalDate.IsZero(),
		"DB上のCodeApprovalDateも、呼び出し元のuser.Update(ctx)後までクリアされたままであるはず")

	var reloadedProfile CorporateProfile
	ok, err = engine.Table("corporate_profile").Where("user_id = ?", user.Id).Cols("corporate_postal_code_approval_date").Get(&reloadedProfile)
	require.NoError(t, err)
	require.True(t, ok)
	require.True(t, reloadedProfile.CorporatePostalCodeApprovalDate.IsZero(),
		"DB上のCorporatePostalCodeApprovalDateもクリアされているはず")
}
