package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type UserAssetView struct {
	UserId                  int64 `json:"user_id"`
	Deposit                 int64 `json:"deposit"`                   // 예치금, 보유금, 투자 가능 금액
	InvestOperationAmount   int64 `json:"invest_operation_amount"`   // 운용 중 금액
	InvestSuccessAmount     int64 `json:"invest_success_amount"`     // 투자 완료 금액
	InvestRequestAmount     int64 `json:"invest_request_amount"`     // 투자 신청 금액
	WithdrawalRequestAmount int64 `json:"withdrawal_request_amount"` // 출금 요청 금액 (= 미완료 + 마감 + 송금 실패)
	UnpaidAmount            int64 `json:"unpaid_amount"`             // 입금 필요 금액
	TotalAsset              int64 `json:"total_asset"`               // 총 자산
}

func (UserAssetView) GetByUserId(ctx context.Context, userId int64) (*UserAssetView, error) {
	ua := UserAssetView{UserId: userId}
	_, err := factory.DB(ctx).Get(&ua)
	if err != nil {
		return nil, err
	}
	return &ua, nil
}

func (UserAssetView) TableName() string {
	return "user_asset_view"
}

func (v UserAssetView) SelectAll() string {
	return v.TableName() + ".*"
}

func (v UserAssetView) IsEmpty() bool {
	return v.UserId == 0
}
