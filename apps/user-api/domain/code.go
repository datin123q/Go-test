package domain

import "context"

type CodeUsecase interface {
	GetAll(ctx context.Context) string
}
