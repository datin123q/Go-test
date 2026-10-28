package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type attachmentMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.AttachmentRepository {
	return &attachmentMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (a attachmentMysqlRepository) Create(ctx context.Context, attachment domain.Attachment) (int64, error) {
	db, err := a.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	_, err = db.InsertOne(&attachment)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert")
	}

	return attachment.Id, nil
}

func (a attachmentMysqlRepository) GetAll(ctx context.Context, options ...cdomain.AttachmentQueryOption) ([]domain.Attachment, error) {
	db, err := a.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.AttachmentQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	attachments := make([]domain.Attachment, 0)

	if err = orm.Table(db, domain.Attachment{}).
		Select(domain.Attachment{}).
		Equal(domain.Attachment{}.TableName()+".resource_id", opts.ResourceId()).
		SortOrder(opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		Find(&attachments); err != nil {
		return nil, errors.Wrap(err, "failed to get attachment")
	}

	return attachments, nil
}

func (a attachmentMysqlRepository) DeleteByResourceId(ctx context.Context, resourceId string) error {
	db, err := a.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	_, err = db.Where(domain.Attachment{}.TableName()+".resource_id = ?", resourceId).Delete(domain.Attachment{})
	if err != nil {
		return errors.Wrap(err, "failed to delete attachment")
	}

	return nil
}

func (a attachmentMysqlRepository) Delete(ctx context.Context, id int64) error {
	db, err := a.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	_, err = db.ID(id).Delete(domain.Attachment{})
	if err != nil {
		return errors.Wrap(err, "failed to delete attachment")
	}

	return nil
}
