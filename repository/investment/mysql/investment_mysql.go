package mysql

import (
	"context"

	"github.com/samber/lo"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type investmentRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.InvestmentRepository {
	return &investmentRepository{
		dbFactory: dbFactory,
	}
}

func (i investmentRepository) GetAll(ctx context.Context, options ...cdomain.InvestmentQueryOption) (int64, []domain.Investment, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.InvestmentQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	invests := make([]domain.Investment, 0)

	totalCount, err := orm.Table(db, domain.Share{}).Select(
		domain.INVESTMENT_QUERY,
		domain.INVESTOR_PROJECT_STRINGS,
		domain.INVESTOR_USER_STRINGS,
		domain.INVESTOR_SPV_STRINGS,
		domain.INVESTOR_PRIVATE_PROFILE_STRINGS).
		InnerJoin(domain.Project{}, "project.id = share.project_id").
		InnerJoin(domain.User{}, "share.owner_id = user.id").
		LeftJoin(domain.PrivateProfile{}, "share.owner_id = private_profile.user_id").
		InnerJoin(domain.Spv{}, "project.spv_id = spv.id").
		SortOrderForTable(domain.Share{}.TableName(), queryOpts.SortBy(), queryOpts.Order()).
		Equal(domain.Share{}.TableName()+".invest_status", queryOpts.InvestStatus()).
		Equal(domain.Share{}.TableName()+".project_type", queryOpts.ProjectType()).
		Like(domain.User{}.TableName()+".name", queryOpts.InvestorName()).
		Like(domain.Spv{}.TableName()+".name", queryOpts.SpvName()).
		Like(domain.Project{}.TableName()+".name", queryOpts.ProjectName()).
		Between("amount", queryOpts.StartAmount(), queryOpts.EndAmount()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&invests)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find and count")
	}

	return totalCount, invests, nil
}

func (i investmentRepository) GetById(ctx context.Context, id string) (domain.Investment, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Investment{}, errors.Wrap(err, "err get db")
	}

	investment := domain.Investment{}

	has, err := orm.Table(db, domain.Share{}).Select(
		domain.INVESTMENT_QUERY,
		domain.INVESTOR_PROJECT_STRINGS,
		domain.INVESTOR_USER_STRINGS,
		domain.INVESTOR_SPV_STRINGS,
		domain.INVESTOR_PRIVATE_PROFILE_STRINGS,
		domain.PROJECT_FILE_URL,
	).
		InnerJoin(domain.Project{}, "project.id = share.project_id").
		InnerJoin(domain.User{}, "share.owner_id = user.id").
		LeftJoin(domain.PrivateProfile{}, "share.owner_id = private_profile.user_id").
		InnerJoin(domain.Spv{}, "project.spv_id = spv.id").
		Equal(domain.Share{}.TableName()+".id", lo.ToPtr(id)).
		Get(&investment)
	if err != nil {
		return domain.Investment{}, errors.Wrap(err, "err get investment detail")
	}

	if !has {
		return domain.Investment{}, errors.New("err investment detail not found")
	}

	return investment, nil
}

func (i investmentRepository) GroupByOwner(ctx context.Context, options ...cdomain.InvestmentQueryOption) ([]domain.Investment, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.InvestmentQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	investments := make([]domain.Investment, 0)

	if err = orm.Table(db, domain.Share{}).
		Select(domain.Share{}, domain.User{}, "user.name as user_name").
		InnerJoin(domain.User{}, "share.owner_id = user.id").
		Equal(domain.Share{}.TableName()+".project_id", queryOpts.ProjectId()).
		Equal(domain.Share{}.TableName()+".invest_status", queryOpts.InvestStatus()).
		GroupBy(domain.Share{}.TableName() + ".owner_id").
		Find(&investments); err != nil {
		return nil, errors.Wrap(err, "err get investment detail")
	}

	return investments, nil
}
