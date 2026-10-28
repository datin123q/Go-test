package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type userAntisocialCheckRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.UserAntisocialCheckRepository {
	return &userAntisocialCheckRepository{
		dbFactory: dbFactory,
	}
}

func (repo userAntisocialCheckRepository) Create(ctx context.Context, userAntisocialCheck domain.UserAntisocialCheck) error {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&userAntisocialCheck); err != nil {
		return errors.Wrap(err, "err insert userAntisocialCheck")
	}

	return nil
}

// GetByUserId は userId を使用して UserAntisocialCheck を検索します。
// 注意: 複数ある場合、IDが最大のものを返します。
// 注意: UserAntisocialCheck は []*AntisocialCheckFile を保有していますが、これは別のクエリを通じて取得する必要があります。Serviceで処理します
func (repo userAntisocialCheckRepository) GetByUserId(ctx context.Context, userId int64) (domain.UserAntisocialCheck, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.UserAntisocialCheck{}, errors.Wrap(err, "err get db")
	}

	var userAntisocialCheck domain.UserAntisocialCheck
	has, err := orm.Table(db, domain.UserAntisocialCheck{}).
		Equal(domain.UserAntisocialCheck{}.TableName()+".check_user_id", userId).
		SortOrder("id", "desc").
		Get(&userAntisocialCheck)

	if err != nil {
		return domain.UserAntisocialCheck{}, errors.Wrap(err, "err find and count")
	}
	if !has {
		return domain.UserAntisocialCheck{}, cdomain.ErrItemNotFound
	}

	return userAntisocialCheck, nil
}
