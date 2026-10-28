package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	shareId1       = "shareId1"
	projectId1     = int64(1000)
	userId1        = int64(2000)
	spvId1         = int64(3000)
	projectFileId1 = int64(6000)
)

var (
	mockShare1 = domain.Share{
		Id:           shareId1,
		ProjectId:    projectId1,
		CategoryType: domain.SOLAR,
		ProjectType:  domain.EQUITY,
		OwnerId:      userId1,
		UpdateId:     userId1,
		Desc:         "太陽光プロジェクトの一部シェア",
		InvestStatus: domain.SUCCESS_INVEST,
		Amount:       5000000,
		Qty:          100,
		Price:        50000,
		ShareType:    domain.BASIC_SHARE,
	}

	mockProject1 = domain.Project{
		Id:              projectId1,
		CommonId:        "0ffbe80a-c628-11ee-a506-0242ac120002",
		OfferingType:    "公開", // 例示値
		SpvId:           spvId1,
		Name:            "太陽光発電プロジェクト",
		CategoryType:    domain.SOLAR, // 例示値
		AdminNote:       "プロジェクト管理者メモ",
		ProjectInfoHtml: "<p>プロジェクト情報</p>",
		Description:     "太陽光発電に関する詳細説明",
		ProjectType:     domain.EQUITY, // 例示値
		OwnerId:         userId1,
		UpdateId:        userId1,
		Capacity:        "100MW",
	}

	mockUser1 = domain.User{
		Id:                userId1,
		Uid:               "user-uuid-5678",
		LoginType:         "メール",
		UserType:          "一般",
		CommonId:          "common-uuid-4321",
		InvestorType:      domain.PRIVATE_PRO_INVESTOR,
		Name:              "ホン・ギルドン",
		Email:             "honggildong@example.com",
		MobilePhone:       "010-1234-5678",
		BlockInvest:       false,
		ZipCode:           "12345",
		Address:           "ソウル特別市江南区",
		City:              "ソウル",
		Province:          "江南区",
		AuditStatus:       domain.WAIT_ANTISOCIAL_AUDIT,
		InvestmentPurpose: "収益創出",
	}

	mockPrivateProfile1 = domain.PrivateProfile{
		UserId:       userId1,
		FirstName:    "ギルドン",
		FamilyName:   "ホン",
		Asset:        domain.ASSET5,
		AnnualIncome: domain.ANNUALINCOME6,
		Nationality:  "日本",
		Job:          "開発者",
		Gender:       "男性",
	}

	mockSpv1 = domain.Spv{
		Id:       spvId1,
		Name:     "エネルギー開発会社",
		Email:    "energydev@example.com",
		Address:  "ソウル特別市麻浦区",
		City:     "ソウル",
		ZipCode:  "121-220",
		Province: "麻浦区",
		CorpNo:   "123-45-67890",
		Phone:    "02-1234-5678",
		OwnerId:  userId1,
		UpdateId: userId1,
	}

	mockProjectFile1 = domain.ProjectFile{
		Id:        projectFileId1,
		ProjectId: projectId1,
		FileName:  "Hello.txt",
		OwnerId:   userId1,
		UpdateId:  userId1,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.Share), new(domain.Project), new(domain.User), new(domain.PrivateProfile), new(domain.Spv), new(domain.DepositRequest), new(domain.UserVirtualAccount), new(domain.ProjectFile))
	xormEngine.ShowSQL(true)

	xormEngine.Insert(mockShare1, mockProject1, mockUser1, mockPrivateProfile1, mockSpv1, mockProjectFile1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
