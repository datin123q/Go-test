package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type MyNumberRepository interface {
	Create(ctx context.Context, myNumber domain.MyNumber) (int64, error)
	GetByUserId(ctx context.Context, userId int64) (domain.MyNumber, error)
	GetById(ctx context.Context, id int64) (domain.MyNumber, error)
	Delete(ctx context.Context, id int64) error
}

type MyNumberService interface {
	RegisterWithFiles(ctx context.Context, input MyNumberRegisterInput) (domain.MyNumber, error)
	GetByUserId(context.Context, int64) (domain.MyNumber, error)
	CheckApproveMyNumberExists(context.Context, string) (bool, error)
}

type MyNumberRegisterInput struct {
	UserId        int64
	MyNumberValue string          // 入力されたマイナンバー値
	FrontFileUrl  string          // 入力されたマイナンバー表面ファイルURL
	BackFileUrl   string          // 入力されたマイナンバー裏面ファイルURL
	MyNumber      domain.MyNumber // userIDで照会したmyNumber
}
