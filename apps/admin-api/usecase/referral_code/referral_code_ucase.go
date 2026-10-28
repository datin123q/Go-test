package referral_code

import (
	"context"

	"github.com/pkg/errors"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type referralCodeUsecase struct {
	codeOwnershipRepository cdomain.CodeOwnershipRepository
}

func New(codeOwnershipRepository cdomain.CodeOwnershipRepository) adomain.ReferralCodeUsecase {
	return &referralCodeUsecase{codeOwnershipRepository: codeOwnershipRepository}
}

func (u referralCodeUsecase) Create(ctx context.Context, request adomain.ReferralCodeCreateRequest) (adomain.ReferralCodeCreateResponse, error) {
	existing, err := u.codeOwnershipRepository.GetByCode(ctx, domain.CodeTypeReferral, request.ReferralCode)
	if err != nil {
		return adomain.ReferralCodeCreateResponse{}, errors.Wrap(err, "failed to get code_ownership")
	}
	if existing != nil {
		return adomain.ReferralCodeCreateResponse{}, errors.Wrap(cdomain.ErrAlreadyExists, "this referral code is already exists")
	}

	ownership := request.ToModel()

	if err := u.codeOwnershipRepository.Create(ctx, ownership); err != nil {
		return adomain.ReferralCodeCreateResponse{}, errors.Wrap(err, "failed to create code ownership")
	}

	created, err := u.codeOwnershipRepository.GetByCode(ctx, domain.CodeTypeReferral, request.ReferralCode)
	if err != nil {
		return adomain.ReferralCodeCreateResponse{}, errors.Wrap(err, "failed to reload code ownership")
	}

	response := adomain.NewReferralCodeCreateResponse(created)
	return response, nil
}
