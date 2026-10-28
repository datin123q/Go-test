package domain

import (
	"context"
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

func setupCheckAllPostalCodeTestDB(t *testing.T) (context.Context, *xorm.Engine) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	// User/CorporateProfileはxorm:"<-"(JOIN専用の読み取り列)を持たないため、
	// Sync2で構造体定義から素直にテーブルを作っても、本番に存在しない列を
	// 隠すことにはならない(CLAUDE.mdのレビュー観点)。base.schema.sqlを手で
	// 転記すると、その後のマイグレーションで足された列(has_experience等)が
	// 漏れて壊れるため、構造体から生成する方式にしている。
	require.NoError(t, engine.Sync2(new(User), new(CorporateProfile)))

	return context.WithValue(context.Background(), factory.ConextDBType, engine), engine
}

func insertCorporateUserFixture(t *testing.T, ctx context.Context, engine *xorm.Engine) *CorporateUser {
	t.Helper()

	// PostalCodeStatus/CorporatePostalStatusは「郵便コードを確認できた」ではなく
	// 「ハガキを発送した」を表すフラグ(apps/jobs-scheduler/jobs/
	// user_address_check_jobs.goが発送直後に両方trueにする)。利用者が実際に
	// コード入力画面へたどり着ける時点で、この2つは既にtrueになっている。
	// このtrue状態こそが本チケットの脆弱性を再現する前提条件なので、fixtureでも
	// 最初からtrueにしておく。
	user := User{
		Uid:              "test-uid",
		PostalCode:       "1000001",
		PostalCodeStatus: true,
		AuditStatus:      CHECK_ADDRESS_AUDIT,
	}
	_, err := engine.Insert(&user)
	require.NoError(t, err)

	profile := CorporateProfile{
		UserId:                       user.Id,
		CorpNo:                       "0000000000",
		CorpType:                     "test",
		EmployeeCount:                "1",
		CapitalAmount:                "1",
		Sales:                        "1",
		RepresentativeFamilyName:     "テスト",
		RepresentativeFirstName:      "テスト",
		RepresentativeFamilyNameKana: "テスト",
		RepresentativeFirstNameKana:  "テスト",
		RepresentativeMobilePhone:    "00000000000",
		RepresentativeEmail:          "rep@example.com",
		ManagerFamilyName:            lo.ToPtr("テスト"),
		ManagerFirstName:             lo.ToPtr("テスト"),
		ManagerFamilyNameKana:        lo.ToPtr("テスト"),
		ManagerFirstNameKana:         lo.ToPtr("テスト"),
		ManagerEmail:                 lo.ToPtr("manager@example.com"),
		ManagerZipCode:               "2000002",
		ManagerProvince:              "テスト",
		ManagerCity:                  "テスト",
		ManagerAddress:               "テスト",
		ManagerMobilePhone:           lo.ToPtr("00000000000"),
		ManagerBirthday:              jsonutils.JSONTime(time.Now()),
		Department:                   lo.ToPtr("テスト"),
		Position:                     lo.ToPtr("テスト"),
		CorporatePostalCode:          "2000002",
		CorporatePostalStatus:        true,
	}
	_, err = engine.Insert(&profile)
	require.NoError(t, err)

	return &CorporateUser{User: user, CorporateProfile: profile}
}

// reloadCorporateUserFixtureはテスト専用の軽量な再読み込み。domain.
// CorporateUser.GetByIdは本人確認書類・株主・バーチャル口座まで読みに行き
// このテストのsqliteスキーマでは賄いきれないため、User/CorporateProfileの
// 2テーブルだけを読み直す。本番は毎リクエストでDBから読み直す(インメモリの
// 構造体を使い回さない)ことを再現する目的。
func reloadCorporateUserFixture(t *testing.T, engine *xorm.Engine, userId, profileId int64) *CorporateUser {
	t.Helper()

	var user User
	ok, err := engine.Table("user").Where("id = ?", userId).Get(&user)
	require.NoError(t, err)
	require.True(t, ok)

	var profile CorporateProfile
	ok, err = engine.Table("corporate_profile").Where("id = ?", profileId).Get(&profile)
	require.NoError(t, err)
	require.True(t, ok)

	return &CorporateUser{User: user, CorporateProfile: profile}
}

// TestCheckAllPostalCode_OnlyOneCodeDoesNotCompleteAuditはMIR-1252の検証。
// 以前はcorpTypeを指定した個別呼び出しで、法人所在地・担当者のどちらか片方の
// 郵便コードを確認しただけでAuditStatus/InvestorTypeが無条件でCOMPLETED_AUDIT/
// CORPORATE_INVESTORへ進み、片方の住所しか確認していない状態で本登録完了・
// 投資可能になっていた。
//
// 敵対的検証で、修正の初版がPostalCodeStatus/CorporatePostalStatus(実際には
// 「ハガキ発送済み」を表すだけで、利用者がコード入力画面に来られる時点で
// 既に両方true)を条件にしており、迂回を防げていないことが判明した。この
// テストのfixtureは両方trueの状態(=実際に起こりうる、コード入力可能な状態)
// から始めており、法人側の正しいコードだけを送っても本登録が完了しない
// ことを検証する。
func TestCheckAllPostalCode_OnlyOneCodeDoesNotCompleteAudit(t *testing.T) {
	ctx, engine := setupCheckAllPostalCodeTestDB(t)
	cu := insertCorporateUserFixture(t, ctx, engine)

	corpType := string(CORP)
	err := cu.CheckAllPostalCode(ctx, "1000001", "2000002", &corpType)
	require.NoError(t, err, "法人側のコードだけの確認は成功するはず")

	var status string
	var investorType string
	ok, err := engine.Table("user").Where("id = ?", cu.User.Id).Cols("audit_status", "investor_type").Get(&status, &investorType)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, string(CHECK_ADDRESS_AUDIT), status,
		"法人側のコードだけを確認した時点では、審査中(CHECK_ADDRESS_AUDIT)のままであるはず"+
			"(修正前・および「発送済みフラグ」を条件にした誤った修正では、ここでCOMPLETED_AUDITになっていた)")
	assert.Empty(t, investorType, "InvestorTypeもまだ設定されないはず")

	var reloadedUser User
	ok, err = engine.Table("user").Where("id = ?", cu.User.Id).Cols("code_approval_date").Get(&reloadedUser)
	require.NoError(t, err)
	require.True(t, ok)
	assert.False(t, reloadedUser.CodeApprovalDate.IsZero(), "法人側は確認できたのでCodeApprovalDateがセットされるはず")

	var reloadedProfile CorporateProfile
	ok, err = engine.Table("corporate_profile").Where("id = ?", cu.CorporateProfile.Id).Cols("corporate_postal_code_approval_date").Get(&reloadedProfile)
	require.NoError(t, err)
	require.True(t, ok)
	assert.True(t, reloadedProfile.CorporatePostalCodeApprovalDate.IsZero(), "担当者側はまだ未確認のまま")
}

// TestCheckAllPostalCode_BothCodesCompleteAuditはMIR-1252の検証。両方の郵便
// コードを確認して初めて、AuditStatus/InvestorTypeが完了状態へ進むことを
// 検証する。
func TestCheckAllPostalCode_BothCodesCompleteAudit(t *testing.T) {
	ctx, engine := setupCheckAllPostalCodeTestDB(t)
	cu := insertCorporateUserFixture(t, ctx, engine)
	userId, profileId := cu.User.Id, cu.CorporateProfile.Id

	corpType := string(CORP)
	err := cu.CheckAllPostalCode(ctx, "1000001", "2000002", &corpType)
	require.NoError(t, err)

	// 1回目の呼び出しがDBへ実際に反映されたことを確かめるため、次の呼び出しは
	// 読み直した構造体に対して行う(本番は毎リクエストでDBから読み直す)。
	cu = reloadCorporateUserFixture(t, engine, userId, profileId)
	managerType := string(MANAGER)
	err = cu.CheckAllPostalCode(ctx, "1000001", "2000002", &managerType)
	require.NoError(t, err, "担当者側のコードの確認も成功するはず")

	var status string
	var investorType string
	ok, err := engine.Table("user").Where("id = ?", cu.User.Id).Cols("audit_status", "investor_type").Get(&status, &investorType)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, string(COMPLETED_AUDIT), status, "両方確認できたので本登録完了になるはず")
	assert.Equal(t, string(CORPORATE_INVESTOR), investorType)
}

// TestCheckAllPostalCode_StaleApprovalDateAfterRevertDoesNotCompleteAuditは
// MIR-1252の検証。①敵対的検証で見つかった穴: 一度本登録を完了したユーザーを
// 管理画面(apps/admin-api/controllers/corporate.goのUpdateAuditStatus、
// case domain.CHECK_ADDRESS_AUDIT)から住所確認中へ差し戻して新しいコードを
// 発行しても、CodeApprovalDate/CorporatePostalCodeApprovalDateがクリアされない
// ままだと、新しいコードを片方だけ確認しただけで(古い承認日が残っているため)
// 完了判定を満たしてしまう。CorporateUser.Updateの修正(code_approval_dateを
// MustColsでゼロ値でも更新対象に含める)により、差し戻し時に承認日をゼロへ
// 戻せば正しくDBへ反映され、この迂回が塞がることを検証する。
func TestCheckAllPostalCode_StaleApprovalDateAfterRevertDoesNotCompleteAudit(t *testing.T) {
	ctx, engine := setupCheckAllPostalCodeTestDB(t)
	cu := insertCorporateUserFixture(t, ctx, engine)
	userId, profileId := cu.User.Id, cu.CorporateProfile.Id

	// 1巡目: 両方確認して本登録完了。以降は毎回DBから読み直し、本番と同じく
	// インメモリの構造体を使い回さない(DB書き込みが実際に効いているかを
	// 検証するため)。
	corpType := string(CORP)
	require.NoError(t, cu.CheckAllPostalCode(ctx, "1000001", "2000002", &corpType))

	cu = reloadCorporateUserFixture(t, engine, userId, profileId)
	managerType := string(MANAGER)
	require.NoError(t, cu.CheckAllPostalCode(ctx, "1000001", "2000002", &managerType))

	cu = reloadCorporateUserFixture(t, engine, userId, profileId)
	require.Equal(t, string(COMPLETED_AUDIT), string(cu.AuditStatus), "前提: 1巡目で本登録完了しているはず")

	// 管理画面が住所確認中へ差し戻し、新しいコードを発行する操作を再現。
	cu.AuditStatus = CHECK_ADDRESS_AUDIT
	cu.PostalCode = "3000003"
	cu.CorporatePostalCode = "4000004"
	cu.CodeApprovalDate = jsonutils.JSONTime{}
	cu.CorporatePostalCodeApprovalDate = jsonutils.JSONTime{}
	require.NoError(t, cu.Update(ctx))

	// 2巡目: 法人側の新しいコードだけを確認(担当者側はまだ未確認)。
	cu = reloadCorporateUserFixture(t, engine, userId, profileId)
	require.True(t, cu.CodeApprovalDate.IsZero() && cu.CorporatePostalCodeApprovalDate.IsZero(),
		"前提: 差し戻し後、DBから読み直した時点で両方の承認日がクリアされているはず")
	corpType2 := string(CORP)
	require.NoError(t, cu.CheckAllPostalCode(ctx, "3000003", "4000004", &corpType2))

	cu = reloadCorporateUserFixture(t, engine, userId, profileId)
	assert.Equal(t, string(CHECK_ADDRESS_AUDIT), string(cu.AuditStatus),
		"差し戻し後、片方だけの再確認では完了しないはず(修正前は古い承認日がDBに残っており、ここでCOMPLETED_AUDITになっていた)")
}
