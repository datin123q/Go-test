package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type ekycRequestLogRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.EkycRequestLogRepository {
	return &ekycRequestLogRepository{
		dbFactory: dbFactory,
	}
}

func (u ekycRequestLogRepository) Create(ctx context.Context, ekycRequestLog domain.EkycRequestLog) error {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(ekycRequestLog); err != nil {
		return errors.Wrap(err, "failed to insert ekycRequestLog")
	}

	return nil
}

func (u ekycRequestLogRepository) Update(ctx context.Context, ekycRequestLog domain.EkycRequestLog) error {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.ID(ekycRequestLog.Id).UseBool().Update(ekycRequestLog); err != nil {
		return errors.Wrap(err, "failed to update ekycRequestLog")
	}

	return nil
}

func (repo ekycRequestLogRepository) GetByEkycId(ctx context.Context, ekycId string) (domain.EkycRequestLog, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.EkycRequestLog{}, errors.Wrap(err, "err get db")
	}

	var ekycRequestLog domain.EkycRequestLog

	has, err := orm.Table(db, domain.EkycRequestLog{}).
		Equal(domain.EkycRequestLog{}.TableName()+".ekyc_id", ekycId).
		Get(&ekycRequestLog)
	if err != nil {
		return domain.EkycRequestLog{}, errors.Wrap(err, "failed to get ekycRequestLog")
	}
	if !has {
		return domain.EkycRequestLog{}, cdomain.ErrItemNotFound
	}

	return ekycRequestLog, nil
}

func (repo ekycRequestLogRepository) GetByUserId(ctx context.Context, userId int64) (domain.EkycRequestLog, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.EkycRequestLog{}, errors.Wrap(err, "err get db")
	}

	var ekycRequestLogs []domain.EkycRequestLog

	err = orm.Table(db, domain.EkycRequestLog{}).
		Equal(domain.EkycRequestLog{}.TableName()+".user_id", userId).
		SortOrder("created", "desc").
		Limit(1, 0).
		Find(&ekycRequestLogs)
	if err != nil {
		return domain.EkycRequestLog{}, errors.Wrap(err, "failed to get ekycRequestLog")
	}
	if len(ekycRequestLogs) == 0 {
		return domain.EkycRequestLog{}, cdomain.ErrItemNotFound
	}

	return ekycRequestLogs[0], nil
}
