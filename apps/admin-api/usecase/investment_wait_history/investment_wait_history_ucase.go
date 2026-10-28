package investment_wait_history

import (
	"context"
	adomain "github.com/illuminarean/energy-share-core-api/domain"

	"github.com/samber/lo"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
)

type investmentWaitHistoryUsecase struct {
	investmentWaitHistoryRepository cdomain.InvestmentWaitHistoryRepository
}

// Deprecated : v1.17.0 이후 사용하지 않습니다.
func New(investmentWaitHistoryRepository cdomain.InvestmentWaitHistoryRepository) adomain.InvestmentWaitHistoryUsecase {
	return &investmentWaitHistoryUsecase{
		investmentWaitHistoryRepository: investmentWaitHistoryRepository,
	}
}

func (i investmentWaitHistoryUsecase) GetAll(ctx context.Context, request adomain.InvestmentWaitHistoryGetAllRequest) (adomain.InvestmentWaitHistoryGetAllResponse, error) {
	var response adomain.InvestmentWaitHistoryGetAllResponse

	opts := cdomain.InvestmentWaitHistoryQueryOptionFunc()

	totalCount, investmentWaitHistories, err := i.investmentWaitHistoryRepository.GetAll(
		ctx,
		opts.WithInvestmentWaitId(lo.FromPtr(request.InvestmentWaitId)),
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return response, errors.Wrap(err, "failed to get investment wait histories")
	}

	return response.WithTotalCount(totalCount).WithItems(investmentWaitHistories), nil
}
