package funding_request

import (
	"context"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
)

type fundingRequestUsecase struct {
	cfg *config.Config
}

func (f fundingRequestUsecase) GetAll(ctx context.Context, request adomain.FundingRequestGetAllRequest) (adomain.FundingRequestGetAllResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (f fundingRequestUsecase) GetOne(ctx context.Context, request adomain.FundingRequestGetOneRequest) (adomain.FundingRequestGetOneResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (f fundingRequestUsecase) Update(ctx context.Context, request adomain.FundingRequestUpdateRequest) error {
	// TODO implement me
	panic("implement me")
}

func New(cfg *config.Config) adomain.FundingRequestUsecase {
	return &fundingRequestUsecase{
		cfg: cfg,
	}
}
