package project

import (
	"context"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type projectUsecase struct {
	shareRepository   cdomain.ShareRepository
	projectRepository cdomain.ProjectRepository
}

func New(
	shareRepository cdomain.ShareRepository,
	projectRepository cdomain.ProjectRepository) adomain.ProjectUsecase {
	return &projectUsecase{
		shareRepository:   shareRepository,
		projectRepository: projectRepository,
	}
}

func (p projectUsecase) GetStatistics(ctx context.Context, request adomain.ProjectGetStatisticsRequest) (adomain.ProjectGetStatisticsResponse, error) {
	project, err := p.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return adomain.ProjectGetStatisticsResponse{}, errors.Wrap(err, "failed to get project")
	}

	shares, err := p.shareRepository.GetAll(ctx, cdomain.ShareQueryOptionFunc().WithProjectId(request.ProjectId))
	if err != nil {
		return adomain.ProjectGetStatisticsResponse{}, errors.Wrap(err, "failed to get shares")
	}

	// 투자에 성공한 건을 필터링합니다.
	successShares := lo.FilterMap(shares, func(share domain.Share, _ int) (domain.Share, bool) {
		if lo.Contains([]domain.INVEST_STATUS{
			domain.SUCCESS_INVEST,
			domain.OPERATION_INVEST,
			domain.REFUND_INVEST,
		}, share.InvestStatus) {
			return share, true
		}
		return domain.Share{}, false
	})

	// 투자 요청 건을 필터링합니다.
	requestShares := lo.FilterMap(shares, func(share domain.Share, _ int) (domain.Share, bool) {
		if lo.Contains([]domain.INVEST_STATUS{
			domain.REQUEST_INVEST,
		}, share.InvestStatus) {
			return share, true
		}
		return domain.Share{}, false
	})

	// 취소된 투자 건을 필터링합니다.
	cancellationShares := lo.FilterMap(shares, func(share domain.Share, _ int) (domain.Share, bool) {
		if lo.Contains([]domain.INVEST_STATUS{
			domain.CANCEL_INVEST,
		}, share.InvestStatus) {
			return share, true
		}
		return domain.Share{}, false
	})

	// 투자에 성공한 유저들의 투자 금액을 합산합니다.
	successShareAmount := lo.Reduce(successShares, func(agg int64, share domain.Share, _ int) int64 {
		return agg + share.Amount
	}, int64(0))

	// 투자에 성공한 유저들의 투자 수량을 합산합니다.
	successShareQty := lo.Reduce(successShares, func(agg int64, share domain.Share, _ int) int64 {
		return agg + share.Qty
	}, int64(0))

	// 투자에 성공한 유저들의 투자 횟수를 값으로 가지는 맵을 만듭니다.
	investUserMap := lo.Reduce(successShares, func(userMap map[int64]int64, share domain.Share, _ int) map[int64]int64 {
		userMap[share.OwnerId]++
		return userMap
	}, map[int64]int64{})

	// 재투자 건수를 계산합니다. (재투자횟수 = 투자횟수 - 1)
	reinvestCount := lo.Reduce(lo.Values(investUserMap), func(agg int64, count int64, _ int) int64 {
		agg += count - 1
		return agg
	}, 0)

	// 투자 성공 건수
	successShareCount := len(successShares)

	return adomain.ProjectGetStatisticsResponse{
		TotalShareCount:        int64(len(shares)),                       // 총 투자 건수
		TargetShareAmount:      project.TargetShareAmount,                // 목표 투자 금액
		TargetShareQty:         project.TargetShareQty,                   // 목표 투자 수량
		SuccessShareAmount:     successShareAmount,                       // 투자 성공 금액
		RequestShareCount:      int64(len(requestShares)),                // 투자 요청 건수
		SuccessShareQty:        successShareQty,                          // 투자 성공 수량(구수)
		RemainingShareQty:      project.TargetShareQty - successShareQty, // 남은 수량 (목표 수량 - 투자 성공 수량)
		ReinvestCount:          reinvestCount,                            // 재투자 건수
		SuccessShareCount:      int64(successShareCount),                 // 투자 성공 건수
		CancellationShareCount: int64(len(cancellationShares)),           // 투자 취소 건수
		SuccessShareAvgQty: lo.If(successShareCount != 0, // 해당 조건은 0으로 나누는 것을 방지하기 위함
			float64(successShareQty)/float64(successShareCount)).Else(0), // 투자 성공 평균 수량 (투자 성공 수량 / 투자 성공 건수)
	}, nil
}
