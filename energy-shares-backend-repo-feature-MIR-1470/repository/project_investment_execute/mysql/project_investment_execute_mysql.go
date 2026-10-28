package mysql

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type projectInvestmentExecuteMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.ProjectInvestmentExecuteRepository {
	return &projectInvestmentExecuteMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (p projectInvestmentExecuteMysqlRepository) Create(ctx context.Context, projectInvestmentExecute domain.ProjectInvestmentExecute) (domain.ProjectInvestmentExecute, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.ProjectInvestmentExecute{}, errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(&projectInvestmentExecute); err != nil {
		return domain.ProjectInvestmentExecute{}, errors.Wrap(err, "err insert ProjectInvestmentExecute")
	}
	return projectInvestmentExecute, nil

}

func (p projectInvestmentExecuteMysqlRepository) Update(ctx context.Context, projectInvestmentExecute domain.ProjectInvestmentExecute) error {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.ID(projectInvestmentExecute.Id).
		Cols("execute_status").
		Update(&projectInvestmentExecute); err != nil {
		return errors.Wrap(err, "err update ProjectInvestmentExecute")
	}

	return nil
}

// IncrementPublishFailedCount はSQLレベルで加算する(MIR-1202)。読み取ってから
// 書き戻す形にすると、同じレコードへの並行更新でカウントを取りこぼす。
func (p projectInvestmentExecuteMysqlRepository) IncrementPublishFailedCount(ctx context.Context, id int64, delta int64) error {
	if delta <= 0 {
		return nil
	}
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err := db.Exec("UPDATE project_investment_execute SET publish_failed_count = publish_failed_count + ? WHERE id = ?", delta, id); err != nil {
		return errors.Wrap(err, "err increment ProjectInvestmentExecute.publish_failed_count")
	}

	return nil
}

func (p projectInvestmentExecuteMysqlRepository) GetByProjectId(ctx context.Context, projectId int64) (domain.ProjectInvestmentExecute, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.ProjectInvestmentExecute{}, errors.Wrap(err, "failed to get db")
	}

	var execute domain.ProjectInvestmentExecute

	has, err := orm.Table(db, domain.ProjectInvestmentExecute{}).
		Equal(domain.ProjectInvestmentExecute{}.TableName()+".project_id", lo.ToPtr(fmt.Sprintf("%v", projectId))).
		Get(&execute)
	if err != nil {
		return domain.ProjectInvestmentExecute{}, errors.Wrap(err, "failed to get project investment execute")
	}

	if !has {
		return domain.ProjectInvestmentExecute{}, errors.Wrap(cdomain.ErrItemNotFound, "failed to get project investment execute")
	}

	return execute, nil
}

func (p projectInvestmentExecuteMysqlRepository) GetAllByProjectId(ctx context.Context, projectId int64) ([]domain.ProjectInvestmentExecute, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db")
	}

	var executes []domain.ProjectInvestmentExecute

	if err := orm.Table(db, domain.ProjectInvestmentExecute{}).
		Equal(domain.ProjectInvestmentExecute{}.TableName()+".project_id", lo.ToPtr(fmt.Sprintf("%v", projectId))).
		Find(&executes); err != nil {
		return nil, errors.Wrap(err, "failed to find project investment executes")
	}

	return executes, nil
}
