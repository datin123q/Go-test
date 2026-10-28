package password_reset_request

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type passwordResetRequestService struct {
	cfg                            *config.Config
	logger                         factory.LoggerFactory
	passwordResetRequestRepository cdomain.PasswordResetRequestRepository
}

func New(
	cfg *config.Config,
	logger factory.LoggerFactory,
	passwordResetRequestRepository cdomain.PasswordResetRequestRepository,
) cdomain.PasswordResetRequestService {
	return &passwordResetRequestService{
		cfg:                            cfg,
		logger:                         logger,
		passwordResetRequestRepository: passwordResetRequestRepository,
	}
}

func (p passwordResetRequestService) HandlePasswordResetRequest(ctx context.Context, email string) (domain.PasswordResetRequest, error) {
	// (조건)
	// 1일 요청 횟수 제한 없음
	// 요청 당 유효 시간 20분, 최초 요청 이후 20분내 재요청 1회까지 가능

	passwordResetRequest, err := p.passwordResetRequestRepository.GetLatestByEmail(ctx, email)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return domain.PasswordResetRequest{}, errors.Wrap(err, "failed to get password_reset_request by email")
	}

	if err != nil && errors.Is(err, cdomain.ErrItemNotFound) || (passwordResetRequest.Expire.Unix() < time.Now().Unix()) {
		// 최초 발송: 조회된 것이 아무것도 없을 경우 또는 만료된 요청일 경우
		passwordResetRequest, err = p.FirstRequest(ctx, email)
		if err != nil {
			return domain.PasswordResetRequest{}, errors.Wrap(err, "failed to first request")
		}
	} else {
		// 재발송: 조회된 것이 만료되지 않은 요청일 경우
		passwordResetRequest, err = p.ReRequest(ctx, passwordResetRequest)
		if err != nil {
			return domain.PasswordResetRequest{}, errors.Wrap(err, "failed to re-request")
		}
	}

	return passwordResetRequest, nil
}

func (p passwordResetRequestService) FirstRequest(ctx context.Context, email string) (domain.PasswordResetRequest, error) {

	id, err := uuid.NewV4()
	if err != nil {
		return domain.PasswordResetRequest{}, errors.Wrap(err, "failed to generate request id")
	}
	reqId := id.String()

	prr := domain.PasswordResetRequest{
		Url:       p.cfg.ServiceConfig.Url.UserServiceDomain + "/reset/password/" + reqId,
		RequestId: reqId,
		Email:     email,
		Expire:    jsonutils.JSONTime(time.Now().Add(time.Minute * cdomain.PASSWORD_RESET_REQUEST_EXPIRE_MINUTES)),
	}

	_, err = p.passwordResetRequestRepository.Create(ctx, prr)
	if err != nil {
		return domain.PasswordResetRequest{}, errors.Wrap(err, "failed to create password_reset_request")
	}

	return prr, nil
}

func (p passwordResetRequestService) ReRequest(ctx context.Context, prr domain.PasswordResetRequest) (domain.PasswordResetRequest, error) {
	prr.SendCount++
	if prr.SendCount > cdomain.PASSWORD_RESET_REQUEST_MAX_SEND_COUNT {
		return domain.PasswordResetRequest{}, errors.Wrap(cdomain.ErrRequestLimitExceeded, "password_reset_request limit exceeded")
	}

	err := p.passwordResetRequestRepository.Update(ctx, prr)
	if err != nil {
		return domain.PasswordResetRequest{}, errors.Wrap(err, "failed to update password_reset_request")
	}

	return prr, nil
}
