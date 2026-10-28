package user_antisocial_check

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

type userAntisocialCheckService struct {
	userAntisocialCheckRepository cdomain.UserAntisocialCheckRepository
	antisocialCheckFileRepository cdomain.AntisocialCheckFileRepository
}

func New(
	userAntisocialCheckRepository cdomain.UserAntisocialCheckRepository,
	antisocialCheckFileRepository cdomain.AntisocialCheckFileRepository,
) cdomain.UserAntisocialCheckService {
	return &userAntisocialCheckService{
		userAntisocialCheckRepository,
		antisocialCheckFileRepository,
	}
}

// GetByUserIdはuser_idに対応するuser_antisocial_checkを取得します。
func (svc userAntisocialCheckService) GetByUserId(ctx context.Context, userId int64) (domain.UserAntisocialCheck, error) {

	userAntisocialCheck, err := svc.userAntisocialCheckRepository.GetByUserId(ctx, userId)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return domain.UserAntisocialCheck{}, nil
		}
		return domain.UserAntisocialCheck{}, err
	}

	_, files, err := svc.antisocialCheckFileRepository.GetAll(ctx,
		cdomain.AntisocialCheckFileQueryOptionFunc().WithCheckId(userId),
		cdomain.AntisocialCheckFileQueryOptionFunc().WithLimit(10), // (2024-04-19)1人あたりの最大アップロード数は10個です。したがって、制限を設けるとパフォーマンスに良いです。
		cdomain.AntisocialCheckFileQueryOptionFunc().WithSortBy("created"),
		cdomain.AntisocialCheckFileQueryOptionFunc().WithOrder("desc"),
	)
	for i := range files {
		files[i].FileUrl = s3utils.MakeCDNSignURL(files[i].FilePath + files[i].FileName + files[i].FileExtension)
	}
	userAntisocialCheck.CheckFiles = files.ToPtrSlice()

	return userAntisocialCheck, nil
}
