package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type privateUserMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.PrivateUserRepository {
	return &privateUserMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (u privateUserMysqlRepository) GetById(ctx context.Context, id int64) (domain.PrivateUser, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.PrivateUser{}, err
	}

	privateUsersWithRoles := make(domain.PrivateUsersWithRoles, 0)

	err = orm.Table(db, domain.User{}).
		Select(domain.User{}, domain.PrivateProfile{}).
		InnerJoin(domain.PrivateProfile{}).
		Equal(domain.User{}.TableName()+".id", id).
		Find(&privateUsersWithRoles)

	if err != nil {
		return domain.PrivateUser{}, errors.Wrap(err, "failed to get private user")
	}
	if privateUsersWithRoles.IsEmpty() {
		return domain.PrivateUser{}, cdomain.ErrItemNotFound
	}

	for i := range privateUsersWithRoles {
		privateUsersWithRoles[i].Roles = []string{string(domain.ROLE_USER)}
	}

	return privateUsersWithRoles.First().PrivateUser, nil
}

func (u privateUserMysqlRepository) GetByEmail(ctx context.Context, email string) (domain.PrivateUser, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.PrivateUser{}, err
	}

	privateUsersWithRoles := make(domain.PrivateUsersWithRoles, 0)

	err = orm.Table(db, domain.User{}).
		Select(domain.User{}, domain.PrivateProfile{}).
		InnerJoin(domain.PrivateProfile{}).
		Equal(domain.User{}.TableName()+".email", email).
		Find(&privateUsersWithRoles)

	if err != nil {
		return domain.PrivateUser{}, errors.Wrap(err, "failed to get private user")
	}
	if privateUsersWithRoles.IsEmpty() {
		return domain.PrivateUser{}, cdomain.ErrItemNotFound
	}

	for i := range privateUsersWithRoles {
		privateUsersWithRoles[i].Roles = []string{string(domain.ROLE_USER)}
	}

	return privateUsersWithRoles.First().PrivateUser, nil
}
