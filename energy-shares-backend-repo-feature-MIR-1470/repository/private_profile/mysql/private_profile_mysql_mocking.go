package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockPrivateProfile1 = domain.PrivateProfile{
		UserId:       2000,
		FirstName:    "ギルドン",
		FamilyName:   "ホン",
		Asset:        domain.ASSET5,
		AnnualIncome: domain.ANNUALINCOME6,
		Nationality:  domain.JP,
		Job:          domain.USER_OFFICIALS,
		Gender:       domain.MALE,
	}

	mockPrivateProfile2 = domain.PrivateProfile{
		UserId:       2001, // 重複したUserID - insertエラーが発生する必要がある
		FirstName:    "グァンスン",
		FamilyName:   "ユ",
		Asset:        domain.ASSET5,
		AnnualIncome: domain.ANNUALINCOME6,
		Nationality:  domain.JP,
		Job:          domain.USER_UNLISTED_COMPANY_EMPLOYEE,
		Gender:       domain.FEMALE,
	}

	mockPrivateProfile3 = domain.PrivateProfile{
		UserId:       1000,
		FirstName:    "グァンスン",
		FamilyName:   "ユ",
		Asset:        domain.ASSET5,
		AnnualIncome: domain.ANNUALINCOME6,
		Nationality:  domain.JP,
		Job:          domain.USER_UNLISTED_COMPANY_EMPLOYEE,
		Gender:       domain.FEMALE,
	}

	mockPrivateProfile4 = domain.PrivateProfile{
		UserId:       1001,
		FirstName:    "グァンスン",
		FamilyName:   "ユ",
		Asset:        domain.ASSET5,
		AnnualIncome: domain.ANNUALINCOME6,
		Nationality:  domain.OTHERS,
		Job:          domain.USER_UNLISTED_COMPANY_EMPLOYEE,
		Gender:       domain.FEMALE,
	}

	mockPrivateProfile5 = domain.PrivateProfile{
		Id:             1000,
		UserId:         2003,
		FirstName:      "怜",
		FamilyName:     "直井",
		FirstNameKana:  "レイ",
		FamilyNameKana: "ナオイ",
		Nationality:    domain.OTHERS,
		Job:            domain.USER_UNLISTED_COMPANY_EMPLOYEE,
		Gender:         domain.FEMALE,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.PrivateProfile))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	_, err = xormEngine.Insert(mockPrivateProfile2, mockPrivateProfile3, mockPrivateProfile4, mockPrivateProfile5)
	if err != nil {
		return nil, err
	}

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
