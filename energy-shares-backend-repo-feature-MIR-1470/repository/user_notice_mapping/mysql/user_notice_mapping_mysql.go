package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type userNoticeMappingRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.UserNoticeMappingRepository {
	return &userNoticeMappingRepository{
		dbFactory: dbFactory,
	}
}

func (repo userNoticeMappingRepository) GetAll(ctx context.Context, options ...cdomain.UserNoticeMappingQueryOption) (int64, []domain.UserNoticeMapping, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, []domain.UserNoticeMapping{}, errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.UserNoticeMappingQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	// find all mappings
	var mappings []domain.UserNoticeMapping
	count, err := orm.Table(db, domain.UserNoticeMapping{}).
		Select(domain.UserNoticeMapping{}.SelectAll()).
		Equal("msg_id", queryOpts.MsgId()).
		Equal("user_id", queryOpts.UserId()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&mappings)
	if err != nil {
		return 0, []domain.UserNoticeMapping{}, errors.Wrap(err, "err find and count")
	}

	return count, mappings, nil
}

func (repo userNoticeMappingRepository) Create(ctx context.Context, mapping domain.UserNoticeMapping) (int64, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	// insert notice msg mapping
	_, err = db.InsertOne(&mapping)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert")
	}

	return mapping.MsgId, nil
}

func (repo userNoticeMappingRepository) CountByMsgIds(ctx context.Context, msgIds []int64) (int64, []domain.NoticeMsgUserCount, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, []domain.NoticeMsgUserCount{}, errors.Wrap(err, "failed to get db")
	}

	var counts []domain.NoticeMsgUserCount
	err = orm.Table(db, domain.UserNoticeMapping{}).
		Select("msg_id", "COUNT(msg_id) as user_count").
		In("msg_id", msgIds).
		GroupBy("msg_id").
		Find(&counts)
	if err != nil {
		return 0, []domain.NoticeMsgUserCount{}, errors.Wrap(err, "failed to find")
	}

	return int64(len(counts)), counts, nil
}
