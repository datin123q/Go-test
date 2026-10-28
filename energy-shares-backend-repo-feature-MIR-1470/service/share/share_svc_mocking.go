package share

import (
	"time"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockShare1Id              = "SHARE_1"
	mockProject1Id            = int64(1000)
	mockUser1Id               = int64(2000)
	mockActor1Id              = int64(2001)
	mockDepositRequest1Id     = int64(3000)
	mockSpv1Id                = int64(4000)
	mockUserAccount1Id        = int64(5000)
	mockUserVirtualAccount1Id = int64(6000)
)

var (
	mockShare1 = domain.Share{
		Id:             mockShare1Id,
		ProjectId:      mockProject1Id,
		OwnerId:        mockUser1Id,
		Qty:            500,
		Amount:         10_000_000,
		Price:          500,
		CommitmentDate: jsonutils.JSONTime(time.Unix(0, 0)),
		Updated:        jsonutils.JSONTime(time.Unix(0, 0)),
	}
	mockSpv1 = domain.Spv{
		Id:   mockSpv1Id,
		Name: "수자원 공사",
	}

	mockProject1 = domain.Project{
		Id:                mockProject1Id,
		Spv:               &mockSpv1,
		Name:              "수자원 공사",
		IsInvestorLimit:   true,
		TargetShareAmount: 100_000_000,
		RaisedAmount:      50_000_000,
		ProjectType:       domain.EQUITY,
		OfferingType:      domain.OFFERING_PUBLIC,
	}

	mockUser1 = domain.User{
		Id:    mockUser1Id,
		Name:  "고길동",
		Email: "gildong@email.com",
	}

	mockActor1 = domain.AdminUser{
		Id:    mockActor1Id,
		Name:  "둘리",
		Email: "dool2@email.com",
	}

	mockDepositRequest1 = domain.DepositRequest{
		Id:            mockDepositRequest1Id,
		DepositDate:   jsonutils.JSONTime(time.Unix(0, 0)),
		DepositAmount: 100_000,
		UserId:        mockUser1Id,
		ShareId:       mockShare1Id,
	}

	mockUserAccount1 = domain.UserAccount{
		Id:      mockUserAccount1Id,
		UserId:  mockUser1Id,
		Balance: lo.ToPtr(int64(500_000)),
	}

	mockUserVirtualAccount1 = domain.UserVirtualAccount{
		Id:               mockUserVirtualAccount1Id,
		UserId:           mockUser1Id,
		VaHolderNameKana: "ひらがな",
	}
)
