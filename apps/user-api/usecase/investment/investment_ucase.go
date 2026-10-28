package investment

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type investmentUsecase struct {
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository
	shareRepository              cdomain.ShareRepository
	projectRepository            cdomain.ProjectRepository
}

func New(
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository,
	shareRepository cdomain.ShareRepository,
	projectRepository cdomain.ProjectRepository,
) udomain.InvestmentUsecase {
	return &investmentUsecase{
		userVirtualAccountRepository: userVirtualAccountRepository,
		shareRepository:              shareRepository,
		projectRepository:            projectRepository,
	}
}

func (i investmentUsecase) GetResult(ctx context.Context, request udomain.InvestmentGetResultRequest) (udomain.InvestmentGetResultResponse, error) {
	share, err := i.shareRepository.GetById(ctx, request.InvestmentId)
	if err != nil {
		return udomain.InvestmentGetResultResponse{}, errors.Wrap(err, "failed to get share")
	}

	if share.OwnerId != request.JWTUser.Id {
		// 該当ユーザーでない場合はエラーを返します。
		return udomain.InvestmentGetResultResponse{}, errors.Wrap(cdomain.ErrItemNotFound, "err not owner")
	}

	project, err := i.projectRepository.GetById(ctx, share.ProjectId)
	if err != nil {
		return udomain.InvestmentGetResultResponse{}, errors.Wrap(err, "failed to get project")
	}

	userVirtualAccount, err := i.userVirtualAccountRepository.GetByUserIdAndLinkedStatus(ctx, request.JWTUser.Id, project.ProjectType, domain.ON_VA_ACCOUNT)
	if err != nil {
		return udomain.InvestmentGetResultResponse{}, errors.Wrap(err, "failed to get user virtual account")
	}

	return udomain.InvestmentGetResultResponse{
		InvestmentId:      share.Id,
		ProjectName:       project.Name,
		InvestmentDate:    share.InvestDate,
		BankName:          userVirtualAccount.VaBankName,
		BankAccountNumber: userVirtualAccount.VaAccountNumber,
		Depositor:         userVirtualAccount.VaHolderNameKana,
		Qty:               share.Qty,
		Price:             share.Price,
		Amount:            share.Qty * share.Price,
	}, nil
}
