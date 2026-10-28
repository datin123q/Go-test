package referral_code

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func TestReferralCodeUsecase_Create(t *testing.T) {
	ctx := context.Background()

	baseRequest := adomain.ReferralCodeCreateRequest{
		ReferralCode: "TESTCODE",
		Notes:        "memo",
	}

	tests := []struct {
		name      string
		request   adomain.ReferralCodeCreateRequest
		repoSetup func() *mocks.CodeOwnershipRepository
		wantErr   bool
	}{
		{
			name:    "creates referral code",
			request: baseRequest,
			repoSetup: func() *mocks.CodeOwnershipRepository {
				repo := &mocks.CodeOwnershipRepository{}
				repo.On("GetByCode", ctx, domain.CodeTypeReferral, baseRequest.ReferralCode).Return((*domain.CodeOwnership)(nil), nil).Once()
				repo.On("Create", ctx, mock.MatchedBy(func(arg domain.CodeOwnership) bool {
					return arg.Code == baseRequest.ReferralCode && arg.OwnerType == domain.SYSTEM
				})).Return(nil).Once()
				repo.On("GetByCode", ctx, domain.CodeTypeReferral, baseRequest.ReferralCode).Return(&domain.CodeOwnership{Id: 10, OwnerType: domain.SYSTEM, OwnerId: 0, CodeType: domain.CodeTypeReferral, Code: baseRequest.ReferralCode}, nil).Once()
				return repo
			},
		},
		{
			name:    "duplicate code",
			request: baseRequest,
			repoSetup: func() *mocks.CodeOwnershipRepository {
				repo := &mocks.CodeOwnershipRepository{}
				repo.On("GetByCode", ctx, domain.CodeTypeReferral, baseRequest.ReferralCode).Return(&domain.CodeOwnership{}, nil)
				return repo
			},
			wantErr: true,
		},
		{
			name:    "create error",
			request: baseRequest,
			repoSetup: func() *mocks.CodeOwnershipRepository {
				repo := &mocks.CodeOwnershipRepository{}
				repo.On("GetByCode", ctx, domain.CodeTypeReferral, baseRequest.ReferralCode).Return((*domain.CodeOwnership)(nil), nil)
				repo.On("Create", ctx, mock.AnythingOfType("domain.CodeOwnership")).Return(errors.New("insert error"))
				return repo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := referralCodeUsecase{
				codeOwnershipRepository: tt.repoSetup(),
			}

			resp, err := uc.Create(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.request.ReferralCode, resp.Code)
		})
	}
}
