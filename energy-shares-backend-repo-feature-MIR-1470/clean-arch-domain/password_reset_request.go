package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

const (
	PASSWORD_RESET_REQUEST_EXPIRE_MINUTES = 20
	PASSWORD_RESET_REQUEST_MAX_SEND_COUNT = 1
)

type PasswordResetRequestService interface {
	HandlePasswordResetRequest(ctx context.Context, email string) (domain.PasswordResetRequest, error)
	FirstRequest(ctx context.Context, email string) (domain.PasswordResetRequest, error)
	ReRequest(ctx context.Context, prr domain.PasswordResetRequest) (domain.PasswordResetRequest, error)
}

type PasswordResetRequestRepository interface {
	Create(ctx context.Context, passwordResetRequest domain.PasswordResetRequest) (int64, error)
	Update(ctx context.Context, passwordResetRequest domain.PasswordResetRequest) error
	GetLatestByEmail(ctx context.Context, email string) (domain.PasswordResetRequest, error)
	GetByRequestId(ctx context.Context, requestId string) (domain.PasswordResetRequest, error)
}
