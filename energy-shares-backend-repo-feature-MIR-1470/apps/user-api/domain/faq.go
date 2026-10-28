package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type FaqUsecase interface {
	GetAll(ctx context.Context, request FaqGetAllRequest) (FaqGetAllResponse, error)
}

type FaqGetAllRequest struct {
	Title        *string `query:"title" description:"タイトル"`
	CategoryType *string `query:"category_type" description:"カテゴリ\n* FAQ_CATEGORY_ACCOUNT_REGISTRATION(会員登録)\n* FAQ_CATEGORY_PUT_TAKE_MONEY(入金・出金)\n* FAQ_CATEGORY_INVESTING(投資)\n* FAQ_CATEGORY_FUNDS(ファンド)\n* FAQ_CATEGORY_REPAY_DISTRIBUTION(返済・分配)" enum:"FAQ_CATEGORY_ACCOUNT_REGISTRATION,FAQ_CATEGORY_PUT_TAKE_MONEY,FAQ_CATEGORY_INVESTING,FAQ_CATEGORY_FUNDS,FAQ_CATEGORY_REPAY_DISTRIBUTION"`
	cdomain.SearchInput
}

type FaqGetAllResponse struct {
	TotalCount int64        `json:"total_count" description:"総件数" format:"int64" example:"1"`
	Items      []domain.Faq `json:"items" description:"FAQリスト"`
}
