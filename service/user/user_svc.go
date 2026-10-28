package user

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pwhelper"
)

type userService struct {
	logger                       factory.LoggerFactory
	pwhelper                     pwhelper.PasswordHelper
	myNumberService              cdomain.MyNumberService
	userAntisocialCheckService   cdomain.UserAntisocialCheckService
	userFileService              cdomain.UserFileService
	privateProfileRepository     cdomain.PrivateProfileRepository
	privateUserRepository        cdomain.PrivateUserRepository
	userAccountRepository        cdomain.UserAccountRepository
	userRepository               cdomain.UserRepository
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository
}

func New(
	logger factory.LoggerFactory,
	pwhelper pwhelper.PasswordHelper,
	myNumberService cdomain.MyNumberService,
	userAntisocialCheckService cdomain.UserAntisocialCheckService,
	userFileService cdomain.UserFileService,
	privateProfileRepository cdomain.PrivateProfileRepository,
	privateUserRepository cdomain.PrivateUserRepository,
	userAccountRepository cdomain.UserAccountRepository,
	userRepository cdomain.UserRepository,
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository,
) cdomain.UserService {
	return &userService{
		logger:                       logger,
		pwhelper:                     pwhelper,
		myNumberService:              myNumberService,
		userAntisocialCheckService:   userAntisocialCheckService,
		userFileService:              userFileService,
		privateProfileRepository:     privateProfileRepository,
		privateUserRepository:        privateUserRepository,
		userAccountRepository:        userAccountRepository,
		userRepository:               userRepository,
		userVirtualAccountRepository: userVirtualAccountRepository,
	}
}

func (u userService) IsDuplicateEmail(ctx context.Context, email string) (bool, error) {
	_, err := u.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return false, nil
		}
		return false, errors.Wrap(err, "failed to get user by email")
	}

	return true, nil
}

func (u userService) IsDuplicateEmailIncludeDeleted(ctx context.Context, email string) (bool, error) {
	_, err := u.userRepository.GetByEmailIncludeDeleted(ctx, email)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return false, nil
		}
		return false, errors.Wrap(err, "failed to get user by email")
	}

	return true, nil
}
