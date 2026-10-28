package mysql

import (
	"context"
	"time"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
	"xorm.io/xorm"
)

type dbFactoryMock struct{}

var (
	mockUser1 = domain.User{
		Id:                 1,
		Uid:                "UID1",
		CommonId:           "111d3741-7180-4024-a969-648723a57bbe",
		Email:              "user1@test.com",
		Name:               "user1",
		IsNewsletter:       true,
		AuditStatus:        domain.COMPLETED_AUDIT,
		UserType:           domain.PRIVATE,
		SignUpApprovalDate: jsonutils.JSONTime(time.Date(2024, 3, 1, 0, 0, 0, 0, time.Local)),
		LastLoginDate:      jsonutils.JSONTime(time.Date(2024, 10, 15, 0, 0, 0, 0, time.Local)),
		Created:            jsonutils.JSONTime(time.Date(2024, 3, 1, 0, 0, 0, 0, time.Local)),
	}

	mockUser2 = domain.User{
		Id:                 2,
		Uid:                "UID2",
		CommonId:           "222d3741-4321-1234-a969-648723a57bbe",
		Email:              "user2@test.com",
		Name:               "user2",
		IsNewsletter:       true,
		AuditStatus:        domain.CHECK_ANTISOCIAL_AUDIT, // TODO: マイナンバー登録復活後にdomain.CHECK_MY_NUMBER_AUDITに戻す
		UserType:           domain.PRIVATE,
		SignUpApprovalDate: jsonutils.JSONTime(time.Date(2024, 4, 2, 0, 0, 0, 0, time.Local)),
		LastLoginDate:      jsonutils.JSONTime(time.Date(2024, 10, 1, 0, 0, 0, 0, time.Local)),
		Created:            jsonutils.JSONTime(time.Date(2024, 4, 1, 0, 0, 0, 0, time.Local)),
	}

	mockUser3 = domain.User{
		Id:                 3,
		Uid:                "UID3",
		CommonId:           "333d3741-4321-1234-a969-648723a57bba",
		Email:              "user3@test.com",
		Name:               "user3",
		IsNewsletter:       true,
		AuditStatus:        domain.COMPLETED_AUDIT,
		UserType:           domain.CORPORATE,
		SignUpApprovalDate: jsonutils.JSONTime(time.Date(2024, 5, 3, 0, 0, 0, 0, time.Local)),
		LastLoginDate:      jsonutils.JSONTime(time.Date(2024, 12, 19, 0, 0, 0, 0, time.Local)),
		Created:            jsonutils.JSONTime(time.Date(2024, 5, 1, 0, 0, 0, 0, time.Local)),
	}

	mockPrivateProfile1 = domain.PrivateProfile{
		Id:           1,
		UserId:       1,
		Nationality:  domain.JP,
		Gender:       domain.MALE,
		Job:          domain.USER_UNLISTED_COMPANY_EMPLOYEE,
		Asset:        domain.ASSET5,
		AnnualIncome: domain.ANNUALINCOME5,
		Birthday:     jsonutils.JSONTime(time.Date(1990, 12, 19, 0, 0, 0, 0, time.Local)),
	}

	mockPrivateProfile2 = domain.PrivateProfile{
		Id:           2,
		UserId:       2,
		Nationality:  domain.JP,
		Gender:       domain.MALE,
		Job:          domain.USER_UNLISTED_COMPANY_EMPLOYEE,
		Asset:        domain.ASSET2,
		AnnualIncome: domain.ANNUALINCOME3,
		Birthday:     jsonutils.JSONTime(time.Date(1996, 2, 9, 0, 0, 0, 0, time.Local)),
	}

	mockUserAssetView1 = domain.UserAssetView{
		UserId:                  1,
		Deposit:                 3000,
		InvestOperationAmount:   1000,
		InvestSuccessAmount:     2000,
		InvestRequestAmount:     1000,
		UnpaidAmount:            1000,
		WithdrawalRequestAmount: 500,
		TotalAsset:              7500,
	}

	mockUserAssetView2 = domain.UserAssetView{
		UserId:                  2,
		Deposit:                 5000,
		InvestOperationAmount:   1000,
		InvestSuccessAmount:     1000,
		InvestRequestAmount:     1000,
		UnpaidAmount:            1000,
		WithdrawalRequestAmount: 500,
		TotalAsset:              8500,
	}

	mockUserAssetView3 = domain.UserAssetView{
		UserId:                  3,
		Deposit:                 1000,
		InvestOperationAmount:   3000,
		InvestSuccessAmount:     3000,
		InvestRequestAmount:     3000,
		UnpaidAmount:            3000,
		WithdrawalRequestAmount: 3000,
		TotalAsset:              13000,
	}

	mockMyNumber1 = domain.MyNumber{
		Id:     1,
		UserId: 1,
		Status: domain.DISAPPROVED,
	}

	mockMyNumber2 = domain.MyNumber{
		Id:     2,
		UserId: 1,
		Status: domain.APPROVE,
	}

	mockMyNumber3 = domain.MyNumber{
		Id:     3,
		UserId: 2,
		Status: domain.DISAPPROVED,
	}

	mockMyNumber4 = domain.MyNumber{
		Id:     4,
		UserId: 2,
		Status: domain.WAIT_APPROVE,
	}

	compositeUser1 = cdomain.CompositeUser{
		Id:                    1,
		Email:                 "user1@test.com",
		Name:                  "user1",
		AuditStatus:           lo.ToPtr(domain.COMPLETED_AUDIT),
		UserType:              lo.ToPtr(domain.PRIVATE),
		Nationality:           lo.ToPtr(domain.JP),
		Gender:                lo.ToPtr(domain.MALE),
		Job:                   lo.ToPtr(domain.USER_UNLISTED_COMPANY_EMPLOYEE),
		Asset:                 lo.ToPtr(domain.ASSET5),
		AnnualIncome:          lo.ToPtr(domain.ANNUALINCOME5),
		AllowsNewsletter:      lo.ToPtr(true),
		MyNumberStatus:        lo.ToPtr(domain.APPROVE),
		Birthday:              lo.ToPtr(jsonutils.JSONTime(time.Date(1990, 12, 19, 0, 0, 0, 0, time.Local))),
		SignUpRequestDate:     lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 3, 1, 0, 0, 0, 0, time.Local))),
		SignUpApprovalDate:    lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 3, 1, 0, 0, 0, 0, time.Local))),
		LastLoginDate:         lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 10, 15, 0, 0, 0, 0, time.Local))),
		TotalAsset:            lo.ToPtr(int64(7500)),
		InvestOperationAmount: lo.ToPtr(int64(1000)),
		Deposit:               lo.ToPtr(int64(3000)),
		InvestSuccessAmount:   lo.ToPtr(int64(2000)),
		InvestRequestAmount:   lo.ToPtr(int64(1000)),
		UnpaidAmount:          lo.ToPtr(int64(1000)),
	}

	compositeUser2 = cdomain.CompositeUser{
		Id:                    2,
		Email:                 "user2@test.com",
		Name:                  "user2",
		AuditStatus:           lo.ToPtr(domain.CHECK_ANTISOCIAL_AUDIT), // TODO: マイナンバー登録復活後にdomain.CHECK_MY_NUMBER_AUDITに戻す
		UserType:              lo.ToPtr(domain.PRIVATE),
		Nationality:           lo.ToPtr(domain.JP),
		Gender:                lo.ToPtr(domain.MALE),
		Job:                   lo.ToPtr(domain.USER_UNLISTED_COMPANY_EMPLOYEE),
		Asset:                 lo.ToPtr(domain.ASSET2),
		AnnualIncome:          lo.ToPtr(domain.ANNUALINCOME3),
		AllowsNewsletter:      lo.ToPtr(true),
		MyNumberStatus:        lo.ToPtr(domain.WAIT_APPROVE),
		Birthday:              lo.ToPtr(jsonutils.JSONTime(time.Date(1996, 2, 9, 0, 0, 0, 0, time.Local))),
		SignUpRequestDate:     lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 4, 1, 0, 0, 0, 0, time.Local))),
		SignUpApprovalDate:    lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 4, 2, 0, 0, 0, 0, time.Local))),
		LastLoginDate:         lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 10, 1, 0, 0, 0, 0, time.Local))),
		TotalAsset:            lo.ToPtr(int64(8500)),
		InvestOperationAmount: lo.ToPtr(int64(1000)),
		Deposit:               lo.ToPtr(int64(5000)),
		InvestSuccessAmount:   lo.ToPtr(int64(1000)),
		InvestRequestAmount:   lo.ToPtr(int64(1000)),
		UnpaidAmount:          lo.ToPtr(int64(1000)),
	}

	compositeUser3 = cdomain.CompositeUser{
		Id:                    3,
		Email:                 "user3@test.com",
		Name:                  "user3",
		AuditStatus:           lo.ToPtr(domain.COMPLETED_AUDIT),
		UserType:              lo.ToPtr(domain.CORPORATE),
		Nationality:           nil,
		Gender:                nil,
		Job:                   nil,
		Asset:                 nil,
		AnnualIncome:          nil,
		AllowsNewsletter:      lo.ToPtr(true),
		MyNumberStatus:        nil,
		Birthday:              nil,
		SignUpRequestDate:     lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 5, 1, 0, 0, 0, 0, time.Local))),
		SignUpApprovalDate:    lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 5, 3, 0, 0, 0, 0, time.Local))),
		LastLoginDate:         lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 12, 19, 0, 0, 0, 0, time.Local))),
		TotalAsset:            lo.ToPtr(int64(13000)),
		InvestOperationAmount: lo.ToPtr(int64(3000)),
		Deposit:               lo.ToPtr(int64(1000)),
		InvestSuccessAmount:   lo.ToPtr(int64(3000)),
		InvestRequestAmount:   lo.ToPtr(int64(3000)),
		UnpaidAmount:          lo.ToPtr(int64(3000)),
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.User), new(domain.PrivateProfile), new(domain.UserAssetView), new(domain.MyNumber))
	xormEngine.ShowSQL(true)

	xormEngine.Insert(mockUser1, mockUser2, mockUser3)
	xormEngine.Insert(mockPrivateProfile1, mockPrivateProfile2)
	xormEngine.Insert(mockUserAssetView1, mockUserAssetView2, mockUserAssetView3)
	xormEngine.Insert(mockMyNumber1, mockMyNumber2, mockMyNumber3, mockMyNumber4)
	xormEngine.Delete(mockMyNumber1)
	xormEngine.Delete(mockMyNumber3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
