package user_virtual_account

import (
	"context"
	"time"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type userVirtualAccountService struct {
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository
}

func New(
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository,
) cdomain.UserVirtualAccountService {
	return &userVirtualAccountService{
		userVirtualAccountRepository: userVirtualAccountRepository,
	}
}

func (u userVirtualAccountService) LinkVirtualAccount(ctx context.Context, userId int64, projectType domain.PROJECT_TYPE) error {
	linkedAccount, err := u.userVirtualAccountRepository.GetByUserIdAndLinkedStatus(ctx, userId, projectType, domain.ON_VA_ACCOUNT)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return errors.Wrap(err, "failed to get user virtual account")
	}

	// 既に接続された仮想口座がある場合は接続解除
	if !linkedAccount.IsEmpty() {
		linkedAccount.LinkedStatus = domain.OFF_VA_ACCOUNT
		linkedAccount.UserId = 0
		linkedAccount.UpdateDate = jsonutils.JSONTime(time.Now())
		if err := u.userVirtualAccountRepository.Update(ctx, linkedAccount); err != nil {
			return errors.Wrap(err, "failed to update user virtual account")
		}
	}

	unlinkedAccount, err := u.userVirtualAccountRepository.GetByUserIdAndLinkedStatus(ctx, 0, projectType, domain.OFF_VA_ACCOUNT)
	if err != nil {
		// ErrItemNotFoundの場合はfactory.ErrorNoVirtualAccountを返す必要があります
		return errors.Wrap(err, "failed to get user virtual account")
	}

	// 仮想口座接続
	unlinkedAccount.LinkedStatus = domain.ON_VA_ACCOUNT
	unlinkedAccount.UserId = userId
	unlinkedAccount.UpdateDate = jsonutils.JSONTime(time.Now())
	if err := u.userVirtualAccountRepository.Update(ctx, unlinkedAccount); err != nil {
		return errors.Wrap(err, "failed to update user virtual account")
	}

	return nil
}
