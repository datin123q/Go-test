package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockProject1 = domain.Project{
		Id:             1000,
		CommonId:       "0cb05069-19e1-4ea8-8cfd-a50388eec4c5",
		Name:           "水資源プロジェクト",
		CategoryType:   domain.SOLAR,
		OwnerId:        2000,
		TargetShareQty: 1000,
		IsWaitInvest:   false,
	}

	mockProjectImage1 = domain.ProjectImage{
		Id:            3000,
		ProjectId:     1000,
		FileName:      "test_report_13",
		OriginalName:  "test_report.pdf",
		FileType:      "PROJECT_REPORT",
		FileSize:      189117,
		FileExtension: "/project_file/project_report/",
		FilePath:      "https://test-cdn.dev.energyshares.jp/project_file/project_report/test_report.pdf",
		OwnerId:       2000,
		UpdateId:      2000,
		ImageType:     domain.MAIN_IMG,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.Project), new(domain.ProjectImage))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockProject1, mockProjectImage1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
