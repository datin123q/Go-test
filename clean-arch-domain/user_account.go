package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type UserAccountRepository interface {
	Create(ctx context.Context, userAccount domain.UserAccount) error
	Update(ctx context.Context, userAccount domain.UserAccount) error
	GetByUserId(ctx context.Context, id int64) (domain.UserAccount, error)
}

type UserAccountService interface {
	DepositCheckProcess(ctx context.Context, depositRequest domain.DepositRequest) error
}
