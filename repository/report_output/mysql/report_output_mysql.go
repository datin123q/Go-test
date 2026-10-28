package mysql

import (
	"context"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type reportOutputRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.ReportOutputRepository {
	return &reportOutputRepository{
		dbFactory: dbFactory,
	}
}

func (repo reportOutputRepository) GetAll(ctx context.Context, options ...cdomain.ReportOutputQueryOption) (int64, []domain.ReportOutput, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.ReportOutputQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	var reportOutputs []domain.ReportOutput

	count, err := orm.Table(db, domain.ReportOutput{}).
		Select("*").
		Equal("submission_type", queryOpts.SubmissionType()).
		In("report_type", queryOpts.ReportTypes()).
		Equal("year", queryOpts.Year()).
		Equal("quarter", queryOpts.Quarter()).
		GroupBy(lo.FromPtr(queryOpts.GroupBy())).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&reportOutputs)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find and count")
	}

	return count, reportOutputs, nil
}

// GetById - Update テスト用に作成
func (repo reportOutputRepository) GetById(ctx context.Context, id int64) (domain.ReportOutput, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.ReportOutput{}, errors.Wrap(err, "err get db")
	}

	var reportOutput domain.ReportOutput
	has, err := db.ID(id).Get(&reportOutput)
	if err != nil {
		return domain.ReportOutput{}, errors.Wrap(err, "err get")
	}
	if !has {
		return domain.ReportOutput{}, cdomain.ErrItemNotFound
	}

	return reportOutput, nil
}

func (repo reportOutputRepository) Update(ctx context.Context, reportOutput domain.ReportOutput) error {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	_, err = db.ID(reportOutput.Id).UseBool().Update(&reportOutput)
	if err != nil {
		return errors.Wrap(err, "failed to update")
	}

	return nil
}
