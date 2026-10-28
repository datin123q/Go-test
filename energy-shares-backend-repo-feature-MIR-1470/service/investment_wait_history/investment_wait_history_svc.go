package investment_wait_history

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type investmentWaitHistoryService struct {
	adminUserRepository             cdomain.AdminUserRepository
	investmentWaitHistoryRepository cdomain.InvestmentWaitHistoryRepository
}

func New(
	adminUserRepository cdomain.AdminUserRepository,
	investmentWaitHistoryRepository cdomain.InvestmentWaitHistoryRepository,
) cdomain.InvestmentWaitHistoryService {
	return &investmentWaitHistoryService{
		adminUserRepository:             adminUserRepository,
		investmentWaitHistoryRepository: investmentWaitHistoryRepository,
	}
}

func (i investmentWaitHistoryService) CreateStatusHistory(ctx context.Context, comparer domain.InvestmentWaitersComparer, processMethod domain.USER_HISTORY_PROCESS_TYPE, updateId int64) error {
	actor, err := i.adminUserRepository.GetById(ctx, updateId)
	if err != nil {
		return errors.Wrap(err, "failed to get user by id")
	}
	// 変更前/後のステータス値が異なるもののみ、履歴を記録します。
	if comparer.IsStatusEqual() {
		return nil
	}

	if err = i.investmentWaitHistoryRepository.Create(ctx, domain.InvestmentWaitHistory{
		InvestmentWaitId: comparer.AfterId(),
		SubstanceType:    string(domain.INVESTMENT_WAIT_HISTORY_STATUS),
		Name:             actor.Name,
		Email:            actor.Email,
		PrevStatus:       comparer.BeforeStatus(),
		CurrStatus:       comparer.AfterStatus(),
		ProcessMethod:    processMethod,
	}); err != nil {
		return errors.Wrap(err, "failed to create investment wait history")
	}

	return nil
}
