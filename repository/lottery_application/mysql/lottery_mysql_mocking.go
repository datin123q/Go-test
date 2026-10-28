package mysql

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
	"xorm.io/xorm"
)

type dbFactoryMock struct {
	db xorm.Interface
}

var (
	now        = time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)
	yesterday  = now.Add(-24 * time.Hour)
	tomorrow   = now.Add(24 * time.Hour)
	twoDaysAgo = now.Add(-48 * time.Hour)

	// Projects
	mockProject1 = domain.Project{
		Id:            1,
		Name:          "Project Alpha",
		ProjectStatus: domain.OPEN_FUNDING,
	}
	mockProject2 = domain.Project{
		Id:            2,
		Name:          "Project Beta",
		ProjectStatus: domain.END_FUNDING,
	}
	mockProject3 = domain.Project{
		Id:            3,
		Name:          "Project Gamma",
		ProjectStatus: domain.SUCCESS_FUNDING,
	}

	// Users
	mockUser1 = domain.User{
		Id:       100,
		Uid:      "uid-100",
		CommonId: "common-100",
		Name:     "Tanaka Taro",
		Email:    "tanaka@example.com",
		Password: "password",
		Salt:     "salt",
	}
	mockUserAccount1 = domain.UserAccount{
		Id:      1000,
		UserId:  100,
		Balance: lo.ToPtr(int64(10000)),
	}
	mockUser2 = domain.User{
		Id:       101,
		Uid:      "uid-101",
		CommonId: "common-101",
		Name:     "Suzuki Hanako",
		Email:    "suzuki@example.com",
		Password: "password",
		Salt:     "salt",
	}
	mockUserAccount2 = domain.UserAccount{
		Id:      1001,
		UserId:  101,
		Balance: lo.ToPtr(int64(20000)),
	}
	mockUser3 = domain.User{
		Id:       102,
		Uid:      "uid-102",
		CommonId: "common-102",
		Name:     "Sato Jiro",
		Email:    "sato@example.com",
		Password: "password",
		Salt:     "salt",
	}
	mockUserAccount3 = domain.UserAccount{
		Id:      1002,
		UserId:  102,
		Balance: lo.ToPtr(int64(30000)),
	}
	mockUser4 = domain.User{
		Id:       103,
		Uid:      "uid-103",
		CommonId: "common-103",
		Name:     "Yamada Ichiro",
		Email:    "yamada@example.com",
		Password: "password",
		Salt:     "salt",
	}
	mockUserAccount4 = domain.UserAccount{
		Id:      1003,
		UserId:  103,
		Balance: lo.ToPtr(int64(40000)),
	}
	mockUser5 = domain.User{
		Id:       104,
		Uid:      "uid-104",
		CommonId: "common-104",
		Name:     "Takahashi Yuki",
		Email:    "takahashi@example.com",
		Password: "password",
		Salt:     "salt",
	}
	mockUserAccount5 = domain.UserAccount{
		Id:      1004,
		UserId:  104,
		Balance: lo.ToPtr(int64(50000)),
	}
	mockUser6 = domain.User{
		Id:       105,
		Uid:      "uid-105",
		CommonId: "common-105",
		Name:     "Watanabe Kenji",
		Email:    "watanabe@example.com",
		Password: "password",
		Salt:     "salt",
	}
	mockUserAccount6 = domain.UserAccount{
		Id:      1005,
		UserId:  105,
		Balance: lo.ToPtr(int64(60000)),
	}

	// APPLIED: 全てのタイムスタンプがNULL
	mockAppApplied = domain.LotteryApplication{
		Id:         "app-001",
		ProjectId:  1,
		UserId:     100,
		Qty:        10,
		UnitPrice:  1000,
		Amount:     10000,
		AppliedAt:  jsonutils.JSONTime(twoDaysAgo),
		WinningAt:  jsonutils.JSONTime{},
		LostAt:     jsonutils.JSONTime{},
		CanceledAt: jsonutils.JSONTime{},
	}

	// WINNING: winning_atがセット、当選口数 = 申込口数（全口当選）
	mockAppWinning = domain.LotteryApplication{
		Id:         "app-002",
		ProjectId:  1,
		UserId:     101,
		Qty:        10,
		UnitPrice:  1000,
		Amount:     10000,
		AppliedAt:  jsonutils.JSONTime(twoDaysAgo),
		WinningAt:  jsonutils.JSONTime(yesterday),
		LostAt:     jsonutils.JSONTime{},
		CanceledAt: jsonutils.JSONTime{},
	}

	// PARTIAL_WINNING: winning_atがセット、当選口数 < 申込口数
	mockAppPartialWinning = domain.LotteryApplication{
		Id:         "app-003",
		ProjectId:  1,
		UserId:     102,
		Qty:        10,
		UnitPrice:  1000,
		Amount:     10000,
		AppliedAt:  jsonutils.JSONTime(twoDaysAgo),
		WinningAt:  jsonutils.JSONTime(yesterday),
		LostAt:     jsonutils.JSONTime{},
		CanceledAt: jsonutils.JSONTime{},
	}

	// LOST: lost_atがセット
	mockAppLost = domain.LotteryApplication{
		Id:         "app-004",
		ProjectId:  2,
		UserId:     100,
		Qty:        5,
		UnitPrice:  2000,
		Amount:     10000,
		AppliedAt:  jsonutils.JSONTime(twoDaysAgo),
		WinningAt:  jsonutils.JSONTime{},
		LostAt:     jsonutils.JSONTime(yesterday),
		CanceledAt: jsonutils.JSONTime{},
	}

	// CANCELED: canceled_atがセット
	mockAppCanceled = domain.LotteryApplication{
		Id:         "app-005",
		ProjectId:  2,
		UserId:     101,
		Qty:        8,
		UnitPrice:  1500,
		Amount:     12000,
		AppliedAt:  jsonutils.JSONTime(twoDaysAgo),
		WinningAt:  jsonutils.JSONTime{},
		LostAt:     jsonutils.JSONTime{},
		CanceledAt: jsonutils.JSONTime(yesterday),
	}

	// 金額フィルターテスト用（amount: 5000）
	mockAppSmallAmount = domain.LotteryApplication{
		Id:         "app-006",
		ProjectId:  1,
		UserId:     100,
		Qty:        5,
		UnitPrice:  1000,
		Amount:     5000,
		AppliedAt:  jsonutils.JSONTime(twoDaysAgo),
		WinningAt:  jsonutils.JSONTime{},
		LostAt:     jsonutils.JSONTime{},
		CanceledAt: jsonutils.JSONTime{},
	}

	// テストケース用: 同一ユーザー同一プロジェクトにキャンセル済み抽選がある
	mockAppCanceledForCreateTest = domain.LotteryApplication{
		Id:         "app-007",
		ProjectId:  3,
		UserId:     103,
		Qty:        10,
		UnitPrice:  1000,
		Amount:     10000,
		AppliedAt:  jsonutils.JSONTime(twoDaysAgo),
		WinningAt:  jsonutils.JSONTime{},
		LostAt:     jsonutils.JSONTime{},
		CanceledAt: jsonutils.JSONTime(yesterday),
	}

	// テストケース用: 同一ユーザー同一プロジェクトに申込中の抽選がある
	mockAppAppliedForCreateTest = domain.LotteryApplication{
		Id:         "app-008",
		ProjectId:  3,
		UserId:     104,
		Qty:        10,
		UnitPrice:  1000,
		Amount:     10000,
		AppliedAt:  jsonutils.JSONTime(twoDaysAgo),
		WinningAt:  jsonutils.JSONTime{},
		LostAt:     jsonutils.JSONTime{},
		CanceledAt: jsonutils.JSONTime{},
	}

	// テストケース用: 同一ユーザー同一プロジェクトにキャンセル済み抽選がある（User 105用）
	mockAppCanceledForMixedTest = domain.LotteryApplication{
		Id:         "app-009",
		ProjectId:  3,
		UserId:     105,
		Qty:        10,
		UnitPrice:  1000,
		Amount:     10000,
		AppliedAt:  jsonutils.JSONTime(twoDaysAgo),
		WinningAt:  jsonutils.JSONTime{},
		LostAt:     jsonutils.JSONTime{},
		CanceledAt: jsonutils.JSONTime(yesterday),
	}

	// テストケース用: 同一ユーザー同一プロジェクトに申込中の抽選がある（User 105用）
	mockAppAppliedForMixedTest = domain.LotteryApplication{
		Id:         "app-010",
		ProjectId:  3,
		UserId:     105,
		Qty:        5,
		UnitPrice:  1000,
		Amount:     5000,
		AppliedAt:  jsonutils.JSONTime(twoDaysAgo),
		WinningAt:  jsonutils.JSONTime{},
		LostAt:     jsonutils.JSONTime{},
		CanceledAt: jsonutils.JSONTime{},
	}

	// Shares for WINNING/PARTIAL_WINNING
	mockShare1 = domain.Share{
		Id:        "share-001",
		Qty:       10, // app-002全口当選
		OwnerId:   101,
		ProjectId: 1,
	}
	mockShare2 = domain.Share{
		Id:        "share-002",
		Qty:       5, // app-003一部当選（10口申込中5口）
		OwnerId:   102,
		ProjectId: 1,
	}

	// LotteryWinning (lottery_application <-> share の紐付け)
	mockWinning1 = domain.LotteryWinning{
		Id:                   1,
		LotteryApplicationId: "app-002",
		ShareId:              "share-001",
	}
	mockWinning2 = domain.LotteryWinning{
		Id:                   2,
		LotteryApplicationId: "app-003",
		ShareId:              "share-002",
	}
)

func (d *dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	if d.db != nil {
		return d.db, nil
	}

	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.ShowSQL(false)

	// テーブル作成
	if err = xormEngine.Sync(
		new(domain.LotteryApplication),
		new(domain.LotteryWinning),
		new(domain.Share),
		new(domain.Project),
		new(domain.User),
		new(domain.UserAccount),
	); err != nil {
		return nil, err
	}

	// テストデータ挿入
	if _, err = xormEngine.Insert(mockProject1, mockProject2, mockProject3); err != nil {
		return nil, err
	}
	if _, err = xormEngine.Insert(mockUser1, mockUser2, mockUser3, mockUser4, mockUser5, mockUser6); err != nil {
		return nil, err
	}
	if _, err = xormEngine.Insert(
		mockUserAccount1,
		mockUserAccount2,
		mockUserAccount3,
		mockUserAccount4,
		mockUserAccount5,
		mockUserAccount6,
	); err != nil {
		return nil, err
	}
	if _, err = xormEngine.Insert(
		mockAppApplied,
		mockAppWinning,
		mockAppPartialWinning,
		mockAppLost,
		mockAppCanceled,
		mockAppSmallAmount,
		mockAppCanceledForCreateTest,
		mockAppAppliedForCreateTest,
		mockAppCanceledForMixedTest,
		mockAppAppliedForMixedTest,
	); err != nil {
		return nil, err
	}
	if _, err = xormEngine.Insert(mockShare1, mockShare2); err != nil {
		return nil, err
	}
	if _, err = xormEngine.Insert(mockWinning1, mockWinning2); err != nil {
		return nil, err
	}

	d.db = xormEngine
	return xormEngine, nil
}

func (d *dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
