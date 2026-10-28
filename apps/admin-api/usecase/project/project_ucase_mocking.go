package project

import "github.com/illuminarean/energy-shares-backend-repo/domain"

var (
	mockProject1 = domain.Project{
		Id:                148,
		TargetShareQty:    10,
		TargetShareAmount: 100000,
		OwnerId:           6,
	}

	mockShare1 = domain.Share{
		Id:           "2dj2Kjh9esSXCtB4gwPli8LYdBf",
		OwnerId:      6,
		Amount:       10000,
		Qty:          1,
		InvestStatus: domain.SUCCESS_INVEST,
	}
	mockShare2 = domain.Share{
		Id:           "2dj2OY1r00CSOFcpE9sBdFkA8bt",
		OwnerId:      6,
		Amount:       40000,
		Qty:          4,
		InvestStatus: domain.SUCCESS_INVEST,
	}
	mockShare3 = domain.Share{
		Id:           "2dj2SwBrEroEOsXsPhFYgZZ4VKo",
		OwnerId:      6,
		Amount:       40000,
		Qty:          4,
		InvestStatus: domain.SUCCESS_INVEST,
	}
	mockShare4 = domain.Share{
		Id:           "2dj2VbNOQZKJdHeC4Ex5nffs57H",
		OwnerId:      6,
		Amount:       10000,
		Qty:          1,
		InvestStatus: domain.SUCCESS_INVEST,
	}
)
