package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type UserVirtualAccountRepository interface {
	GetByUserIdAndLinkedStatus(ctx context.Context, userId int64, projectType domain.PROJECT_TYPE, linkedStatus domain.VIRTUAL_ACCOUNT_LINK_STATUS) (domain.UserVirtualAccount, error)
	Update(ctx context.Context, userVirtualAccount domain.UserVirtualAccount) error
}

type UserVirtualAccountService interface {
	LinkVirtualAccount(ctx context.Context, userId int64, projectType domain.PROJECT_TYPE) error
}
