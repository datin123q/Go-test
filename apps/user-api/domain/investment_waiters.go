package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type InvestmentWaitersUsecase interface {
	Cancel(ctx context.Context, request InvestmentWaitersCancelRequest) (InvestmentWaitersCancelResponse, error)
}

type InvestmentWaitersCancelRequest struct {
	JWTUser            jwthelper.User
	ClientIp           string `json:"-" query:"-" form:"-"`
	UserID             int64  `param:"id" path:"id" json:"-" query:"-" form:"-" validate:"required,gt=0" description:"ユーザーID" example:"1"`
	InvestmentWaiterID int64  `param:"waitlist_id" path:"waitlist_id" json:"-" query:"-" form:"-" validate:"required,gt=0" description:"キャンセル待ち申請ID" example:"1"`
}

type InvestmentWaitersCancelResponse struct{}
