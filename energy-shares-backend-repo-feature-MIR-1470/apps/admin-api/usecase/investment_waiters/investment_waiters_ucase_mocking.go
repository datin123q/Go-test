package investment_waiters

import (
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	mockSuccessInvestStatus       = domain.SUCCESS_INVEST
	mockCancelInvestStatus        = domain.CANCEL_INVEST
	mockInvestmentWaiters2ShareId = "SHARE_ID1"
	mockInvestmentWaiters3ShareId = "SHARE_ID2"

	mockInvestmentWaiters1 = domain.InvestmentWaiters{Id: 1000, ProjectId: 2000, UserId: 3000, Qty: 5, Status: domain.PROPOSAL}
	mockInvestmentWaiters2 = domain.InvestmentWaiters{Id: 1001, ProjectId: 2000, UserId: 3000, Status: domain.WAITING_END, ShareId: mockInvestmentWaiters2ShareId, InvestStatus: &mockSuccessInvestStatus}
	mockInvestmentWaiters3 = domain.InvestmentWaiters{Id: 1002, ProjectId: 2000, UserId: 3000, Status: domain.WAITING_END, ShareId: mockInvestmentWaiters3ShareId, InvestStatus: &mockCancelInvestStatus}

	mockUser1 = domain.User{Id: 3000, Name: "User 1", Email: "gildong@email.com"}

	mockInvestmentWaitersResponse1 = adomain.InvestmentWaitersResponse{Id: 1000, ProjectId: 2000, Qty: 5, Type: domain.PROPOSAL, Status: domain.INVESTMENT_WAITING_STATUS_PROPOSAL}
	mockInvestmentWaitersResponse2 = adomain.InvestmentWaitersResponse{Id: 1001, ProjectId: 2000, ShareId: &mockInvestmentWaiters2ShareId, Type: domain.WAITING_END, Status: domain.INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS}
	mockInvestmentWaitersResponse3 = adomain.InvestmentWaitersResponse{Id: 1002, ProjectId: 2000, ShareId: &mockInvestmentWaiters3ShareId, Type: domain.WAITING_END, Status: domain.INVESTMENT_WAITING_STATUS_WINNING_AFTER_CANCELED}

	mockInvestmentWaitersWaiting = domain.InvestmentWaiters{Id: 1003, ProjectId: 2000, UserId: 3000, Qty: 5, Status: domain.WAITING}

	// mockInvestmentWaitersWinningItem1 は繰上げ履歴(MIR-1068)の1明細のモック。
	mockInvestmentWaitersWinningItem1 = domain.InvestmentWaitersWinningItem{
		ShareId:            "SHARE_ID1",
		Qty:                3,
		Price:              10000,
		Amount:             30000,
		InvestStatus:       domain.SUCCESS_INVEST,
		DepositCheckStatus: domain.OK_DEPOSIT,
	}

	mockProject1 = domain.Project{
		Id:             2000,
		Name:           "Project 1",
		RemainingQty:   10,
		MinShareQty:    2,
		SharePrice:     10000,
		TargetShareQty: 100,
		IsAmountLimit:  true,
	}
)
