package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type userMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.UserRepository {
	return &userMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (u userMysqlRepository) Create(ctx context.Context, user domain.User) (int64, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, err
	}

	_, err = db.UseBool().Insert(&user)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create user")
	}

	return user.Id, nil
}

func (u userMysqlRepository) Update(ctx context.Context, user domain.User) error {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return err
	}

	_, err = db.ID(user.Id).UseBool().Update(&user)
	if err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	return nil
}

func (u userMysqlRepository) UpdateLastLoginDate(ctx context.Context, id int64, lastLoginDate jsonutils.JSONTime) error {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return err
	}

	_, err = db.ID(id).Cols("last_login_date").Update(&domain.User{LastLoginDate: lastLoginDate})
	if err != nil {
		return errors.Wrap(err, "failed to update user last login date")
	}

	return nil
}

func (u userMysqlRepository) GetById(ctx context.Context, id int64) (domain.User, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.User{}, err
	}

	usersWithRoles := make(domain.UsersWithRoles, 0)

	err = orm.Table(db, domain.User{}).
		Select(domain.User{}.SelectAll()).
		Equal(domain.User{}.TableName()+".id", id).
		Find(&usersWithRoles)

	if err != nil {
		return domain.User{}, errors.Wrap(err, "failed to get user")
	}

	if usersWithRoles.IsEmpty() {
		return domain.User{}, errors.Wrap(cdomain.ErrItemNotFound, "user not found")
	}

	for i := range usersWithRoles {
		usersWithRoles[i].Roles = []string{string(domain.ROLE_USER)}
	}

	return usersWithRoles.First().User, nil
}

func (u userMysqlRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{}

	has, err := orm.Table(db, domain.User{}).
		Equal(domain.User{}.TableName()+".email", email).
		Get(&user)
	if err != nil {
		return domain.User{}, errors.Wrap(err, "failed to get user")
	}

	if !has {
		return domain.User{}, errors.Wrap(cdomain.ErrItemNotFound, "user not found")
	}

	return user, nil
}

func (u userMysqlRepository) GetByEmailIncludeDeleted(ctx context.Context, email string) (domain.User, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{}

	has, err := orm.Table(db, domain.User{}).
		Equal(domain.User{}.TableName()+".email", email).
		Unscoped().
		Get(&user)
	if err != nil {
		return domain.User{}, errors.Wrap(err, "failed to get user")
	}

	if !has {
		return domain.User{}, errors.Wrap(cdomain.ErrItemNotFound, "user not found")
	}

	return user, nil
}

func (u userMysqlRepository) GetAll(ctx context.Context, options ...cdomain.UserQueryOption) (int64, []domain.User, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, []domain.User{}, err
	}

	queryOpts := cdomain.UserQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	// find all users
	var users []domain.User
	query := orm.Table(db, domain.User{}).
		Select(domain.User{}.SelectAll()).
		Equal(domain.User{}.TableName()+".is_newsletter", queryOpts.NewsletterAgreement()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset())
	if queryOpts.InvestorType() != nil {
		query.In(domain.User{}.TableName()+".investor_type", *queryOpts.InvestorType())
	}
	if queryOpts.Ids() != nil {
		query.In(domain.User{}.TableName()+".id", queryOpts.Ids())
	}
	count, err := query.FindAndCount(&users)
	if err != nil {
		return 0, []domain.User{}, errors.Wrap(err, "failed to get users")
	}

	// set roles to user
	for i := range users {
		users[i].Roles = []string{string(domain.ROLE_USER)}
	}

	return count, users, nil
}
