package mysql

import (
	"context"

	"github.com/samber/lo"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockUserId1 = int64(1000)
	mockUserId2 = int64(1001)
	mockUserId3 = int64(1002)

	mockCorporateProfile1 = domain.CorporateProfile{
		Id:                    1,
		UserId:                mockUserId1,
		CorpNo:                "1234567890",
		CorpType:              "フィルム暖房製造",
		ManagerFamilyName:     lo.ToPtr("김"),
		ManagerFirstName:      lo.ToPtr("민수"),
		ManagerFamilyNameKana: lo.ToPtr("キム"),
		ManagerFirstNameKana:  lo.ToPtr("ミンス"),
		ManagerEmail:          lo.ToPtr("minsu.kim@example.com"),
		ManagerMobilePhone:    lo.ToPtr("010-1234-5678"),
		Department:            lo.ToPtr("인사부"),
		Position:              lo.ToPtr("팀장"),
	}

	mockCorporateProfile2 = domain.CorporateProfile{
		Id:                    2,
		UserId:                mockUserId2,
		CorpNo:                "1234567891",
		CorpType:              "フィルム暖房製造",
		ManagerFamilyName:     lo.ToPtr("Kim"),
		ManagerFirstName:      lo.ToPtr("Jinsoo"),
		ManagerFamilyNameKana: lo.ToPtr("キム"),
		ManagerFirstNameKana:  lo.ToPtr("ジンス"),
		ManagerEmail:          lo.ToPtr("jinsoo.kim@example.com"),
		ManagerMobilePhone:    lo.ToPtr("010-9876-5432"),
		Department:            lo.ToPtr("HR Department"),
		Position:              lo.ToPtr("Manager"),
	}

	mockCorporateProfile3 = domain.CorporateProfile{
		Id:                    3,
		UserId:                mockUserId3,
		CorpNo:                "1234567892",
		CorpType:              "フィルム暖房製造",
		ManagerFamilyName:     lo.ToPtr("Kim"),
		ManagerFirstName:      lo.ToPtr("Binsoo"),
		ManagerFamilyNameKana: lo.ToPtr("キム"),
		ManagerFirstNameKana:  lo.ToPtr("ジンス"),
		ManagerEmail:          lo.ToPtr("binsoo.kim@example.com"),
		ManagerMobilePhone:    lo.ToPtr("010-9876-5432"),
		Department:            lo.ToPtr("HR Department"),
		Position:              lo.ToPtr("Manager"),
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.CorporateProfile))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockCorporateProfile1, mockCorporateProfile2, mockCorporateProfile3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
