package mysql

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/samber/lo"
)

type projectFileMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.ProjectFileRepository {
	return &projectFileMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (p projectFileMysqlRepository) GetByProjectIdAndFileType(ctx context.Context, projectId int64, fileType domain.PROJECT_FILE_TYPE) (domain.ProjectFile, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.ProjectFile{}, errors.Wrap(err, "failed to get db")
	}

	projectFile := domain.ProjectFile{}

	has, err := orm.Table(db, domain.ProjectFile{}).
		Equal(domain.ProjectFile{}.TableName()+".project_id", lo.ToPtr(fmt.Sprintf("%v", projectId))).
		Equal(domain.ProjectFile{}.TableName()+".file_type", lo.ToPtr(string(fileType))).
		Get(&projectFile)
	if err != nil {
		return domain.ProjectFile{}, errors.Wrap(err, "failed to get project file")
	}

	if !has {
		return domain.ProjectFile{}, errors.Wrap(cdomain.ErrItemNotFound, "failed to get project file")
	}

	return projectFile, nil
}
