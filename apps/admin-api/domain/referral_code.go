package domain

import (
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type ReferralCodeUsecase interface {
	Create(ctx context.Context, request ReferralCodeCreateRequest) (ReferralCodeCreateResponse, error)
}

type ReferralCodeCreateRequest struct {
	JWTUser      jwthelper.User `json:"-"`
	ReferralCode string         `json:"referral_code" description:"紹介コード" validate:"required"`
	Notes        string         `json:"notes,omitempty" description:"備考"`
}

func (r *ReferralCodeCreateRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if strings.TrimSpace(r.ReferralCode) == "" {
		return factory.ErrorParameter.New(errors.New("referral_code is required"))
	}

	return nil
}

func (r *ReferralCodeCreateRequest) ToModel() domain.CodeOwnership {
	var notes *string
	if trimmed := strings.TrimSpace(r.Notes); trimmed != "" {
		notes = &trimmed
	}

	return domain.CodeOwnership{
		OwnerType: domain.SYSTEM,
		OwnerId:   0,
		CodeType:  domain.CodeTypeReferral,
		Code:      r.ReferralCode,
		Notes:     notes,
	}
}

type ReferralCodeCreateResponse struct {
	ID        int64            `json:"id"`
	OwnerType domain.USER_TYPE `json:"owner_type"`
	OwnerID   int64            `json:"owner_id"`
	CodeType  domain.CodeType  `json:"code_type"`
	Code      string           `json:"code"`
	Notes     *string          `json:"notes,omitempty"`
}

func NewReferralCodeCreateResponse(model *domain.CodeOwnership) ReferralCodeCreateResponse {
	if model == nil {
		return ReferralCodeCreateResponse{}
	}

	return ReferralCodeCreateResponse{
		ID:        model.Id,
		OwnerType: model.OwnerType,
		OwnerID:   model.OwnerId,
		CodeType:  model.CodeType,
		Code:      model.Code,
		Notes:     model.Notes,
	}
}
