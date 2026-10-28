package policy

import (
	"context"
	"testing"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/stretchr/testify/assert"
)

func TestPolicyService_CreatePolicyHistory(t *testing.T) {
	type args struct {
		ctx      context.Context
		before   domain.Policy
		after    domain.Policy
		modifier jwthelper.User
	}

	tests := []struct {
		name                    string
		args                    args
		adminUserRepository     func() *mocks.AdminUserRepository
		policyHistoryRepository func() *mocks.PolicyHistoryRepository
		wantErr                 error
	}{
		{
			name: "(성공) 약관 변경 이력 생성",
			args: args{
				ctx:    context.Background(),
				before: mockBeforePolicy,
				after:  mockAfterPolicy,
				modifier: jwthelper.User{
					Id:    mockAdminUser.Id,
					Name:  mockAdminUser.Name,
					Email: mockAdminUser.Email,
				},
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				adminUserRepository := &mocks.AdminUserRepository{}
				adminUserRepository.On("GetById", context.Background(), mockAdminUser.Id).
					Return(mockAdminUser, nil)
				return adminUserRepository
			},
			policyHistoryRepository: func() *mocks.PolicyHistoryRepository {
				policyHistoryRepository := &mocks.PolicyHistoryRepository{}
				policyHistoryRepository.On("Create", context.Background(),
					domain.PolicyInfoHistory{
						PolicyId:        mockAfterPolicy.Id,
						SubstanceType:   string(domain.POLICY_HISTORY_CONTENTS),
						ContentType:     string(domain.POLICY_CHANGED_CONTENTS),
						Name:            mockAdminUser.Name,
						Email:           mockAdminUser.Email,
						PrevContentHtml: mockBeforePolicy.ContentHtml,
						CurrContentHtml: mockAfterPolicy.ContentHtml,
					},
				).Return(int64(1), nil)
				policyHistoryRepository.On("Create", context.Background(),
					domain.PolicyInfoHistory{
						PolicyId:      mockAfterPolicy.Id,
						SubstanceType: string(domain.POLICY_HISTORY_CONTENTS),
						ContentType:   string(domain.POLICY_CHANGED_CONTENTS),
						Name:          mockAdminUser.Name,
						Email:         mockAdminUser.Email,
						CurrContent:   mockAfterPolicy.Attachment.OriginalName,
						CurrUrl:       mockAfterPolicy.Attachment.FileUrl,
						PrevContent:   mockBeforePolicy.Attachment.OriginalName,
						PrevUrl:       mockBeforePolicy.Attachment.FileUrl,
					},
				).Return(int64(2), nil)
				return policyHistoryRepository
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := policyService{
				adminUserRepository:     tt.adminUserRepository(),
				policyHistoryRepository: tt.policyHistoryRepository(),
			}

			err := service.CreatePolicyHistory(tt.args.ctx, cdomain.PolicyHistoryInput{After: mockAfterPolicy, Before: mockBeforePolicy, Modifier: tt.args.modifier})

			assert.Equalf(t, tt.wantErr, err, "PolicyService.CreatePolicyHistory() error = %v, wantErr %v", tt.wantErr, tt.wantErr)
		})
	}
}
