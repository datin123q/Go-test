package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockUser1 = domain.User{
		Id:           1000,
		Uid:          "UID1",
		CommonId:     "410d3741-7180-4024-a969-648723a57bbe",
		Email:        "gildong@email.com",
		IsNewsletter: true,
		IsSmsCert:    true,
		Roles:        []string{string(domain.ROLE_USER)},
	}

	mockUser2 = domain.User{
		Id:           2000,
		Uid:          "UID2",
		CommonId:     "410d3741-4321-1234-a969-648723a57bbe",
		Email:        "chulsu@email.com",
		IsNewsletter: false,
		IsSmsCert:    false,
		Roles:        []string{string(domain.ROLE_USER)},
	}
	mockUser2Role1 = domain.Role{Id: 2002, UserId: 2000, RoleCode: string(domain.ROLE_USER)}

	mockPrivateProfile1 = domain.PrivateProfile{
		UserId:       mockUser1.Id,
		FirstName:    "ギルドン",
		FamilyName:   "ホン",
		Asset:        domain.ASSET5,
		AnnualIncome: domain.ANNUALINCOME6,
		Nationality:  domain.JP,
		Job:          domain.USER_ORGANIZATION_EMPLOYEE,
		Gender:       domain.MALE,
	}

	mockPrivateProfile2 = domain.PrivateProfile{
		UserId:       mockUser2.Id,
		FirstName:    "ギルドン",
		FamilyName:   "ホン",
		Asset:        domain.ASSET5,
		AnnualIncome: domain.ANNUALINCOME6,
		Nationality:  domain.JP,
		Job:          domain.USER_ORGANIZATION_EMPLOYEE,
		Gender:       domain.MALE,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.User), new(domain.Role), new(domain.PrivateProfile))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	_, err = xormEngine.Insert(mockUser2, mockUser2Role1, mockPrivateProfile2)
	if err != nil {
		return nil, err
	}
	_, err = xormEngine.Insert(mockUser1, mockPrivateProfile1)
	if err != nil {
		return nil, err

	}

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
