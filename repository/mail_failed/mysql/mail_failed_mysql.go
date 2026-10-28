package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type mailFailedMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.MailFailedRepository {
	return &mailFailedMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (repo mailFailedMysqlRepository) Create(ctx context.Context, mailFailed domain.MailFailed) (int64, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&mailFailed); err != nil {
		return 0, errors.Wrap(err, "err insert mail failed to request")

	}

	return mailFailed.Id, nil
}

func (repo mailFailedMysqlRepository) GetAll(ctx context.Context, options ...cdomain.MailFailedQueryOption) (int64, []domain.MailFailed, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.MailFailedQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	var mailFailedList []domain.MailFailed

	total, err := orm.Table(db, domain.MailFailed{}).
		LikeJson(domain.MailFailed{}.TableName()+".receivers", queryOpts.Receiver()).
		Like(domain.MailFailed{}.TableName()+".template_name", queryOpts.TemplateName()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&mailFailedList)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find and count")
	}

	return total, mailFailedList, nil
}
