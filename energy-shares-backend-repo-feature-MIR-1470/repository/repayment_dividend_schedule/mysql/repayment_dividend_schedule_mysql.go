package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type repaymentDividendScheduleMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.RepaymentDividendScheduleRepository {
	return &repaymentDividendScheduleMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (r repaymentDividendScheduleMysqlRepository) Create(ctx context.Context, repaymentDividendSchedule domain.RepaymentDividendSchedule) (domain.RepaymentDividendSchedule, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.RepaymentDividendSchedule{}, errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(&repaymentDividendSchedule); err != nil {
		return domain.RepaymentDividendSchedule{}, errors.Wrap(err, "failed to insert repayment dividend schedule")
	}

	return repaymentDividendSchedule, nil
}

func (r repaymentDividendScheduleMysqlRepository) Update(ctx context.Context, repaymentDividendSchedule domain.RepaymentDividendSchedule) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.ID(repaymentDividendSchedule.Id).Update(repaymentDividendSchedule); err != nil {
		return errors.Wrap(err, "failed to update repayment dividend schedule")
	}

	return nil
}

func (r repaymentDividendScheduleMysqlRepository) GetAll(ctx context.Context, options ...cdomain.RepaymentDividendScheduleQueryOption,
) (int64, domain.RepaymentDividendSchedules, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.RepaymentDividendQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	repaymentDividendSchedules := make(domain.RepaymentDividendSchedules, 0)

	total, err := orm.Table(db, domain.RepaymentDividendSchedule{}).
		Equal(domain.RepaymentDividendSchedule{}.TableName()+".project_id", queryOpts.ProjectId()).
		Equal(domain.RepaymentDividendSchedule{}.TableName()+".status", queryOpts.Status()).
		Equal(domain.RepaymentDividendSchedule{}.TableName()+".return_type", queryOpts.ReturnType()).
		BeforeOrEqual(domain.RepaymentDividendSchedule{}.TableName()+".execute_date", queryOpts.BeforeOrEqualExecuteDate()).
		AfterOrEqual(domain.RepaymentDividendSchedule{}.TableName()+".execute_date", queryOpts.AfterOrEqualExecuteDate()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&repaymentDividendSchedules)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get repayment dividend schedules")
	}

	return total, repaymentDividendSchedules, nil
}

func (r repaymentDividendScheduleMysqlRepository) GetById(ctx context.Context, id int64) (domain.RepaymentDividendSchedule, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.RepaymentDividendSchedule{}, errors.Wrap(err, "failed to get db")
	}

	var schedule domain.RepaymentDividendSchedule

	has, err := orm.Table(db, domain.RepaymentDividendSchedule{}).
		Id(id).
		Get(&schedule)
	if err != nil {
		return domain.RepaymentDividendSchedule{}, errors.Wrap(err, "failed to get repayment dividend schedules")
	}
	if !has {
		return domain.RepaymentDividendSchedule{}, errors.Wrap(cdomain.ErrItemNotFound, "failed to get repayment dividend schedules")
	}

	return schedule, nil
}

func (r repaymentDividendScheduleMysqlRepository) Delete(ctx context.Context, id int64) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err := db.ID(id).Delete(&domain.RepaymentDividendSchedule{}); err != nil {
		return errors.Wrap(err, "failed to delete repayment dividend schedule")
	}

	return nil
}

// GetSum returns the sum of the column
// options: project_id, status, sum_column(required)
// Caution: it always returns float64. If you want to get int, you should convert it.
func (r repaymentDividendScheduleMysqlRepository) GetSum(ctx context.Context, options ...cdomain.RepaymentDividendScheduleQueryOption) (float64, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.RepaymentDividendQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	if opts.SumColumn() == nil {
		return 0, errors.New("sum column is required")
	}

	result, err := orm.Table(db, domain.RepaymentDividendSchedule{}).
		Equal(domain.RepaymentDividendSchedule{}.TableName()+".project_id", opts.ProjectId()).
		Equal(domain.RepaymentDividendSchedule{}.TableName()+".status", opts.Status()).
		Sum(domain.RepaymentDividendSchedule{}.TableName() + "." + *opts.SumColumn())

	if err != nil {
		return 0, errors.Wrap(err, "failed to sum")
	}

	return result, nil
}

// GetSums returns the sum of the share columns
// options: project_id, status, sum_columns(required)
// Caution: it always returns float64. If you want to get int, you should convert it.
func (r repaymentDividendScheduleMysqlRepository) GetSums(ctx context.Context, options ...cdomain.RepaymentDividendScheduleQueryOption) ([]float64, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.RepaymentDividendQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	sumColumns := *opts.SumColumns() // 複数列をサポートするようにメソッド変更が必要
	if len(sumColumns) == 0 {
		return nil, errors.New("at least one sum column is required")
	}

	var columnNames []string
	for _, column := range sumColumns {
		columnNames = append(columnNames, domain.RepaymentDividendSchedule{}.TableName()+"."+column)
	}

	results, err := orm.Table(db, domain.RepaymentDividendSchedule{}).
		Equal(domain.RepaymentDividendSchedule{}.TableName()+".project_id", opts.ProjectId()).
		Equal(domain.RepaymentDividendSchedule{}.TableName()+".status", opts.Status()).
		Sums(columnNames)

	if err != nil {
		return nil, errors.Wrap(err, "failed to sum")
	}

	return results, nil
}
