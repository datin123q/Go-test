package code

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energyshares-user-api/domain"
)

type codeUsecase struct {
	code *config.Code
}

func New(code *config.Code) domain.CodeUsecase {
	return &codeUsecase{
		code: code,
	}
}

func (c codeUsecase) GetAll(ctx context.Context) string {
	return c.code.JsonOutput()
}
