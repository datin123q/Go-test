package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type noticeMsgReadMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.NoticeMsgReadRepository {
	return &noticeMsgReadMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (repo noticeMsgReadMysqlRepository) GetAll(ctx context.Context, options ...cdomain.NoticeMsgReadQueryOption) (int64, []domain.NoticeMsgRead, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, []domain.NoticeMsgRead{}, errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.NoticeMsgReadQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	var noticeMsgReads []domain.NoticeMsgRead
	count, err := orm.Table(db, domain.NoticeMsgRead{}).
		Select(domain.NoticeMsgRead{}.SelectAll()).
		Equal(domain.NoticeMsgRead{}.TableName()+".msg_id", queryOpts.MsgId()).
		Equal(domain.NoticeMsgRead{}.TableName()+".user_id", queryOpts.UserId()).
		Equal(domain.NoticeMsgRead{}.TableName()+".is_read", queryOpts.IsRead()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&noticeMsgReads)
	if err != nil {
		return 0, []domain.NoticeMsgRead{}, errors.Wrap(err, "err find and count")
	}

	return count, noticeMsgReads, nil
}

func (repo noticeMsgReadMysqlRepository) Create(ctx context.Context, noticeMsgRead domain.NoticeMsgRead) (int64, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, err
	}

	_, err = db.UseBool().Insert(&noticeMsgRead)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create notice msg read")
	}

	return noticeMsgRead.Id, nil
}

func (repo noticeMsgReadMysqlRepository) Update(ctx context.Context, noticeMsgRead domain.NoticeMsgRead) error {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return err
	}

	_, err = db.ID(noticeMsgRead.Id).UseBool().Update(&noticeMsgRead)
	if err != nil {
		return errors.Wrap(err, "failed to update notice msg read")
	}

	return nil
}
