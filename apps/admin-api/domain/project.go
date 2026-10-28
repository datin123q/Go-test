package domain

import "context"

type ProjectUsecase interface {
	GetStatistics(ctx context.Context, request ProjectGetStatisticsRequest) (ProjectGetStatisticsResponse, error)
}

type ProjectGetStatisticsRequest struct {
	ProjectId int64 `param:"id" validate:"required" swagger:"desc(프로젝트 ID)"`
}

type ProjectGetStatisticsResponse struct {
	TotalShareCount        int64   `json:"total_share_count" swagger:"desc(총 투자 건수)"`
	TargetShareAmount      int64   `json:"target_share_amount" swagger:"desc(목표 투자 금액)"`
	TargetShareQty         int64   `json:"target_share_qty" swagger:"desc(목표 투자 수량)"`
	SuccessShareAmount     int64   `json:"success_share_amount" swagger:"desc(투자 성공 금액)"`
	RequestShareCount      int64   `json:"request_share_count" swagger:"desc(투자 요청 건수)"`
	SuccessShareQty        int64   `json:"success_share_qty" swagger:"desc(투자 성공 수량)"`
	RemainingShareQty      int64   `json:"remaining_share_qty" swagger:"desc(남은 수량)"`
	ReinvestCount          int64   `json:"reinvest_count" swagger:"desc(재투자 건수)"`
	SuccessShareCount      int64   `json:"success_share_count" swagger:"desc(투자 성공 건수)"`
	CancellationShareCount int64   `json:"cancellation_share_count" swagger:"desc(투자 취소 건수)"`
	SuccessShareAvgQty     float64 `json:"success_share_avg_qty" swagger:"desc(투자 성공 평균 수량)"`
}
