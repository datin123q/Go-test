package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type blogPostMysqlRepository struct {
	dbFactory factory.DBFactory
}

func (b blogPostMysqlRepository) Exist(ctx context.Context, id string) (bool, error) {
	q, err := b.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return false, errors.Wrap(err, "err get db")
	}
	has, err := orm.Table(q, domain.BlogPost{}).
		Select(domain.BlogPost{}).
		Equal("id", id).
		Get(&domain.BlogPost{})
	if err != nil {
		return false, errors.Wrap(err, "err get blog post")
	}
	return has, nil
}

func (b blogPostMysqlRepository) GetById(ctx context.Context, id string) (bool, domain.BlogPost, error) {
	q, err := b.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return false, domain.BlogPost{}, errors.Wrap(err, "err get db")
	}

	var blogPosts domain.BlogPost

	has, err := orm.Table(q, domain.BlogPost{}).
		Select(domain.BlogPost{}).
		Equal("id", id).
		Get(&blogPosts)
	if err != nil {
		return has, domain.BlogPost{}, errors.Wrap(err, "err get blog post by id")
	}
	return has, blogPosts, nil
}

func (b blogPostMysqlRepository) GetAll(ctx context.Context, offset, limit int) (int64, []domain.BlogPost, error) {
	q, err := b.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	blogPosts := make([]domain.BlogPost, 0)

	err = orm.Table(q, domain.BlogPost{}).
		Select(domain.BlogPost{}).
		SortOrder("published_at", "DESC").
		Limit(limit, offset).
		Find(&blogPosts)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get blog posts")
	}

	return int64(len(blogPosts)), blogPosts, nil
}

func (b blogPostMysqlRepository) Create(ctx context.Context, blogPost domain.BlogPost) (string, error) {
	db, err := b.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return "", errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&blogPost); err != nil {
		return "", errors.Wrap(err, "err insert blog post")
	}

	return blogPost.Id, nil
}

func (b blogPostMysqlRepository) Update(ctx context.Context, p domain.BlogPost) error {
	db, err := b.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	_, err = db.ID(p.Id).Update(&p)
	if err != nil {
		return err
	}
	return nil
}

func (b blogPostMysqlRepository) Delete(ctx context.Context, id string) error {
	db, err := b.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	_, err = db.ID(id).Delete(&domain.BlogPost{})
	if err != nil {
		return err
	}
	return nil
}

func New(dbFactory factory.DBFactory) cdomain.BlogPostRepository {
	return &blogPostMysqlRepository{
		dbFactory: dbFactory,
	}
}
