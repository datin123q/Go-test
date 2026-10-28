package private_user

import (
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
)

type privateUserUsecase struct {
	cfg            *config.Config
	userRepository cdomain.UserRepository
}

func New(
	cfg *config.Config,
	userRepository cdomain.UserRepository,
) adomain.PrivateUserUsecase {
	return &privateUserUsecase{
		cfg:            cfg,
		userRepository: userRepository,
	}
}
