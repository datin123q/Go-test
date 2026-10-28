package changelog

import (
	"context"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
)

type changelogUsecase struct {
	cfg                 *config.Config
	changelogRepository cdomain.ChangelogRepository
}

func New(
	cfg *config.Config,
	changelogRepository cdomain.ChangelogRepository,
) adomain.ChangelogUsecase {
	return &changelogUsecase{
		cfg:                 cfg,
		changelogRepository: changelogRepository,
	}
}

func (u changelogUsecase) GetAll(ctx context.Context, request adomain.ChangelogGetAllRequest) (adomain.ChangelogGetAllResponse, error) {
	options := []cdomain.ChangelogQueryOption{
		cdomain.ChangelogQueryOptionFunc().WithStartDate(request.StartDate),
		cdomain.ChangelogQueryOptionFunc().WithEndDate(request.EndDate),
		cdomain.ChangelogQueryOptionFunc().WithLimit(request.MaxResultCount),
		cdomain.ChangelogQueryOptionFunc().WithId(request.Id),
	}

	events, err := u.changelogRepository.GetAll(ctx, options...)
	if err != nil {
		return adomain.ChangelogGetAllResponse{}, err
	}

	return adomain.ChangelogGetAllResponse{Items: events}, nil
}
