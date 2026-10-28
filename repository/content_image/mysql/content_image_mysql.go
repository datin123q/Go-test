package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type contentImageMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.ContentImageRepository {
	return &contentImageMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (c contentImageMysqlRepository) Create(ctx context.Context, contentImage domain.ContentImage) (int64, error) {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	_, err = db.InsertOne(&contentImage)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert")
	}

	return contentImage.Id, nil
}

func (c contentImageMysqlRepository) GetAll(ctx context.Context, options ...cdomain.ContentImageQueryOption) ([]domain.ContentImage, error) {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.ContentImageQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	ContentImages := make([]domain.ContentImage, 0)

	if err = orm.Table(db, domain.ContentImage{}).
		Select(domain.ContentImage{}).
		Equal(domain.ContentImage{}.TableName()+".resource_id", opts.ResourceId()).
		SortOrder(opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		Find(&ContentImages); err != nil {
		return nil, errors.Wrap(err, "failed to get ContentImage")
	}

	return ContentImages, nil
}

func (c contentImageMysqlRepository) DeleteByResourceId(ctx context.Context, resourceId string) error {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	_, err = db.Where(domain.ContentImage{}.TableName()+".resource_id = ?", resourceId).Delete(domain.ContentImage{})
	if err != nil {
		return errors.Wrap(err, "failed to delete ContentImage")
	}

	return nil
}

func (c contentImageMysqlRepository) Delete(ctx context.Context, id int64) error {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	_, err = db.ID(id).Delete(domain.ContentImage{})
	if err != nil {
		return errors.Wrap(err, "failed to delete ContentImage")
	}

	return nil
}
