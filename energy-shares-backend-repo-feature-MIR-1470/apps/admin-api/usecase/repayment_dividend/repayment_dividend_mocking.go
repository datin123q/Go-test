package repayment_dividend

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockUser1 = domain.User{Id: 3000, Name: "User 1", Email: "gildong@email.com"}

	mockProject1 = domain.Project{
		Id:            2000,
		Name:          "Project 1",
		RemainingQty:  10,
		MinShareQty:   2,
		SharePrice:    10000,
		RaisedAmount:  10000,
		ProjectType:   domain.EQUITY,
		ProjectStatus: domain.OPERATION_FUNDING,
	}

	mockProjectInvestmentExecute1 = domain.ProjectInvestmentExecute{
		Id:            1,
		ProjectId:     mockProject1.Id,
		ExecuteStatus: domain.COMPLETED,
		Amount:        2000,
		ExecuteDate:   jsonutils.JSONTime{},
		Created:       jsonutils.JSONTime{},
	}

	mockSchedule1 = domain.RepaymentDividendSchedule{
		Id:                        1000,
		ProjectId:                 2000,
		ProjectType:               domain.EQUITY,
		ProjectShareQty:           10,
		ProjectShareAmount:        1000,
		DecideDate:                jsonutils.JSONTime{},
		ExecuteDate:               jsonutils.JSONTime(time.Date(2024, 11, 10, 0, 0, 0, 0, time.Local)),
		ReturnType:                domain.REPAYMENT_RETURN,
		Status:                    "",
		RepaymentAmount:           1000,
		RepaymentPerShare:         0,
		PreviousRepaymentReserve:  0,
		RepaymentReserve:          0,
		InvestmentBalance:         0,
		DividendAmount:            0,
		NetDividendAfterBrokerFee: 0,
		BrokerFee:                 0,
		ReturnRate:                0,
		DividendPerShare:          0,
		WithholdingTax:            0,
		ProjectSpvId:              0,
		NumInvestors:              0,
		Created:                   jsonutils.JSONTime{},
		Updated:                   jsonutils.JSONTime{},
		Deleted:                   jsonutils.JSONTime{},
	}

	mockSchedule2 = domain.RepaymentDividendSchedule{
		Id:                        1002,
		ProjectId:                 2000,
		ProjectType:               domain.EQUITY,
		ProjectShareQty:           10,
		ProjectShareAmount:        1000,
		DecideDate:                jsonutils.JSONTime{},
		ExecuteDate:               jsonutils.JSONTime(time.Date(2024, 12, 10, 0, 0, 0, 0, time.Local)),
		ReturnType:                domain.REPAYMENT_RETURN,
		Status:                    "",
		RepaymentAmount:           1000,
		RepaymentPerShare:         0,
		PreviousRepaymentReserve:  0,
		RepaymentReserve:          0,
		InvestmentBalance:         0,
		DividendAmount:            0,
		NetDividendAfterBrokerFee: 0,
		BrokerFee:                 0,
		ReturnRate:                0,
		DividendPerShare:          0,
		WithholdingTax:            0,
		ProjectSpvId:              0,
		NumInvestors:              0,
		Created:                   jsonutils.JSONTime{},
		Updated:                   jsonutils.JSONTime{},
		Deleted:                   jsonutils.JSONTime{},
	}

	mockSchedule3 = domain.RepaymentDividendSchedule{
		Id:                        1002,
		ProjectId:                 2000,
		ProjectType:               domain.EQUITY,
		ProjectShareQty:           10,
		ProjectShareAmount:        1000,
		DecideDate:                jsonutils.JSONTime{},
		ExecuteDate:               jsonutils.JSONTime(time.Date(2024, 9, 10, 0, 0, 0, 0, time.Local)),
		ReturnType:                domain.REPAYMENT_RETURN,
		Status:                    domain.RETURN_COMPLETED,
		RepaymentAmount:           1000,
		RepaymentPerShare:         0,
		PreviousRepaymentReserve:  0,
		RepaymentReserve:          0,
		InvestmentBalance:         0,
		DividendAmount:            0,
		NetDividendAfterBrokerFee: 0,
		BrokerFee:                 0,
		ReturnRate:                0,
		DividendPerShare:          0,
		WithholdingTax:            0,
		ProjectSpvId:              0,
		NumInvestors:              0,
		Created:                   jsonutils.JSONTime{},
		Updated:                   jsonutils.JSONTime{},
		Deleted:                   jsonutils.JSONTime{},
	}

	mockShare1 = domain.Share{
		Id:           "2dj2Kjh9esSXCtB4gwPli8LYdBf",
		OwnerId:      mockUser1.Id,
		ProjectId:    mockSchedule1.ProjectId,
		Amount:       10000,
		Qty:          1,
		InvestStatus: domain.SUCCESS_INVEST,
	}
	mockShare2 = domain.Share{
		Id:           "2dj2OY1r00CSOFcpE9sBdFkA8bt",
		OwnerId:      mockUser1.Id,
		ProjectId:    mockSchedule1.ProjectId,
		Amount:       40000,
		Qty:          4,
		InvestStatus: domain.SUCCESS_INVEST,
	}
	mockShare3 = domain.Share{
		Id:           "2dj2SwBrEroEOsXsPhFYgZZ4VKo",
		OwnerId:      mockUser1.Id,
		ProjectId:    mockSchedule1.ProjectId,
		Amount:       40000,
		Qty:          4,
		InvestStatus: domain.SUCCESS_INVEST,
	}
	mockShare4 = domain.Share{
		Id:           "2dj2VbNOQZKJdHeC4Ex5nffs57H",
		ProjectId:    mockSchedule1.ProjectId,
		OwnerId:      mockUser1.Id,
		Amount:       10000,
		Qty:          1,
		InvestStatus: domain.SUCCESS_INVEST,
	}
)
