package mysql

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type shareMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.ShareRepository {
	return &shareMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (s shareMysqlRepository) GetById(ctx context.Context, id string) (domain.Share, error) {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Share{}, errors.Wrap(err, "failed to get db")
	}

	share := domain.Share{}

	has, err := orm.Table(db, domain.Share{}).Id(id).Get(&share)
	if err != nil {
		return domain.Share{}, errors.Wrap(err, "failed to get share")
	}

	if !has {
		return domain.Share{}, errors.Wrap(cdomain.ErrItemNotFound, "share not found")
	}

	return share, nil
}

func (s shareMysqlRepository) LockById(ctx context.Context, id string) (domain.Share, error) {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Share{}, errors.Wrap(err, "failed to get db")
	}

	share := domain.Share{}
	has, err := db.
		SQL("SELECT share.* FROM share WHERE share.id = ? AND share.deleted IS NULL FOR UPDATE", id).
		Get(&share)
	if err != nil {
		return domain.Share{}, errors.Wrap(err, "failed to lock share")
	}
	if !has {
		return domain.Share{}, errors.Wrap(cdomain.ErrItemNotFound, "share not found")
	}

	return share, nil
}

func (s shareMysqlRepository) Update(ctx context.Context, share domain.Share) error {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	_, err = db.ID(share.Id).UseBool().Update(share)
	if err != nil {
		return errors.Wrap(err, "err update share")
	}

	return nil
}

func (s shareMysqlRepository) GetInvestor(ctx context.Context, options ...cdomain.ShareQueryOption) ([]domain.User, error) {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.ShareQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	var users []domain.User

	if opts.GroupBy() == nil {
		return nil, errors.New("group by is required. it can be 'owner_id' or 'email'")
	}

	if err = orm.Table(db, domain.Share{}).
		Select(domain.User{}).
		InnerJoin(domain.User{}, "share.owner_id = user.id").
		Equal(domain.Share{}.TableName()+".project_id", opts.ProjectId()).
		Equal("user.is_newsletter", opts.IsNewsLetter()).
		In(domain.Share{}.TableName()+".invest_status", opts.InvestStatuses()...).
		GroupBy(domain.Share{}.TableName() + "." + *opts.GroupBy()).
		Find(&users); err != nil {
		return nil, errors.Wrap(err, "failed to find users")
	}

	return users, nil
}
func (s shareMysqlRepository) GetInvestorCount(ctx context.Context, projectId int64, options ...cdomain.ShareQueryOption) (int64, error) {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.ShareQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	count, err := orm.Table(db, domain.Share{}).
		Equal(domain.Share{}.TableName()+".project_id", lo.ToPtr(fmt.Sprintf("%v", projectId))).
		In(domain.Share{}.TableName()+".invest_status", opts.InvestStatuses()...).
		GroupBy(domain.Share{}.TableName() + ".owner_id").
		Count()
	if err != nil {
		return 0, errors.Wrap(err, "failed to count")
	}

	return count, nil
}

// GetSum returns the sum of the share column
// options: project_id, user_id, invest_status, sum_column(required)
// Caution: it always returns float64. If you want to get int, you should convert it.
func (s shareMysqlRepository) GetSum(ctx context.Context, options ...cdomain.ShareQueryOption) (float64, error) {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.ShareQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	if opts.SumColumn() == nil {
		return 0, errors.New("sum column is required")
	}

	result, err := (orm.Table(db, domain.Share{}).
		Equal(domain.Share{}.TableName()+".project_id", opts.ProjectId())).
		Equal(domain.Share{}.TableName()+".owner_id", opts.UserId()).
		In(domain.Share{}.TableName()+".invest_status", opts.InvestStatuses()...).
		Sum(domain.Share{}.TableName() + "." + *opts.SumColumn())

	if err != nil {
		return 0, errors.Wrap(err, "failed to sum")
	}

	return result, nil
}

func (s shareMysqlRepository) GetAll(ctx context.Context, options ...cdomain.ShareQueryOption) ([]domain.Share, error) {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.ShareQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	var shares []domain.Share

	if err = orm.Table(db, domain.Share{}).
		Equal(domain.Share{}.TableName()+".project_id", opts.ProjectId()).
		Equal(domain.Share{}.TableName()+".owner_id", opts.UserId()).
		In(domain.Share{}.TableName()+".invest_status", opts.InvestStatuses()...).
		SortOrder(opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		Find(&shares); err != nil {
		return nil, errors.Wrap(err, "failed to find shares")
	}

	return shares, nil
}
