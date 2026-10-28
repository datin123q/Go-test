package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type noticeMsgRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.NoticeMsgRepository {
	return &noticeMsgRepository{
		dbFactory: dbFactory,
	}
}

func (repo noticeMsgRepository) GetAll(ctx context.Context, options ...cdomain.NoticeMsgQueryOption) (int64, []domain.NoticeMsg, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, []domain.NoticeMsg{}, errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.NoticeMsgQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	// find all notice msgs
	var noticeMsgs []domain.NoticeMsg
	query := orm.Table(db, domain.NoticeMsg{}).
		Select(domain.NoticeMsg{}.SelectAll()).
		Like(domain.NoticeMsg{}.TableName()+".title", queryOpts.Title()).
		Between("notice_date", queryOpts.StartDate(), queryOpts.EndDate()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset())

	if queryOpts.MsgIds() != nil {
		query = query.In(domain.NoticeMsg{}.TableName()+".id", queryOpts.MsgIds())
	}

	count, err := query.FindAndCount(&noticeMsgs)
	if err != nil {
		return 0, []domain.NoticeMsg{}, errors.Wrap(err, "err find and count")
	}

	return count, noticeMsgs, nil
}

func (repo noticeMsgRepository) GetById(ctx context.Context, id int64) (domain.NoticeMsg, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.NoticeMsg{}, errors.Wrap(err, "failed to get db")
	}

	// find notice msg
	var noticeMsg domain.NoticeMsg
	has, err := orm.Table(db, domain.NoticeMsg{}).Id(id).Get(&noticeMsg)
	if err != nil {
		return domain.NoticeMsg{}, errors.Wrap(err, "failed to find by id")
	}
	if !has {
		return domain.NoticeMsg{}, cdomain.ErrItemNotFound
	}

	return noticeMsg, nil
}

func (repo noticeMsgRepository) Create(ctx context.Context, noticeMsg domain.NoticeMsg) (int64, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	// insert notice msg
	_, err = db.InsertOne(&noticeMsg)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert")
	}

	return noticeMsg.Id, nil
}
