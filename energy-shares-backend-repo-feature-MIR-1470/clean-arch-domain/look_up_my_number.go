package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type LookUpMyNumberRepository interface {
	Create(ctx context.Context, lookUpMyNumber domain.LookUpMyNumber) (int64, error)
	GetByMyNumberId(ctx context.Context, myNumberId int64) (domain.LookUpMyNumber, error)
	GetByMyNumberIdIncludeDeleted(ctx context.Context, myNumberId int64) (domain.LookUpMyNumber, error)
	GetByMyNumber(ctx context.Context, myNumber string) (domain.LookUpMyNumber, error)
	GetByMyNumberIncludeDeleted(ctx context.Context, myNumber string) (domain.LookUpMyNumber, error)
	DeleteByMyNumberId(ctx context.Context, myNumberId int64) error
}

type LookUpMyNumberService interface {
	GetDecryptedMyNumberByMyNumberId(ctx context.Context, myNumberId int64) (string, error)
}
