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

type userVirtualAccountMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.UserVirtualAccountRepository {
	return &userVirtualAccountMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (u userVirtualAccountMysqlRepository) GetByUserIdAndLinkedStatus(ctx context.Context, userId int64, projectType domain.PROJECT_TYPE, linkedStatus domain.VIRTUAL_ACCOUNT_LINK_STATUS) (domain.UserVirtualAccount, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.UserVirtualAccount{}, errors.Wrap(err, "failed to get db")
	}

	userVirtualAccount := domain.UserVirtualAccount{}

	has, err := orm.Table(db, domain.UserVirtualAccount{}).
		Select(domain.UserVirtualAccount{}).
		Equal(domain.UserVirtualAccount{}.TableName()+".user_id", userId).
		Equal(domain.UserVirtualAccount{}.TableName()+".project_type", lo.ToPtr(fmt.Sprintf("%v", projectType))).
		Equal(domain.UserVirtualAccount{}.TableName()+".linked_status", lo.ToPtr(fmt.Sprintf("%v", linkedStatus))).
		Get(&userVirtualAccount)
	if err != nil {
		return domain.UserVirtualAccount{}, errors.Wrap(err, "failed to get user virtual account")
	}

	if !has {
		return domain.UserVirtualAccount{}, cdomain.ErrItemNotFound
	}

	return userVirtualAccount, nil
}

func (u userVirtualAccountMysqlRepository) Update(ctx context.Context, userVirtualAccount domain.UserVirtualAccount) error {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return err
	}

	_, err = db.ID(userVirtualAccount.Id).Update(&userVirtualAccount)
	if err != nil {
		return errors.Wrap(err, "failed to update user virtual account")
	}

	return nil
}
