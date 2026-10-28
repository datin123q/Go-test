package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type BlogPostRepository interface {
	Exist(ctx context.Context, id string) (bool, error)
	GetById(ctx context.Context, id string) (bool, domain.BlogPost, error)
	GetAll(ctx context.Context, offset, limit int) (int64, []domain.BlogPost, error)
	Create(ctx context.Context, p domain.BlogPost) (string, error)
	Update(ctx context.Context, p domain.BlogPost) error
	Delete(ctx context.Context, id string) error
}
