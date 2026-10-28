package user_account

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	mockDepositRequestId1 = int64(1000)
	mockUserId1           = int64(2000)
	mockProjectId1        = int64(3000)
	mockShareId1          = "SHARE_ID1"
)

var mockDepositRequest1 = domain.DepositRequest{
	Id:                   mockDepositRequestId1,
	UserId:               mockUserId1,
	ShareId:              mockShareId1,
	ProjectId:            mockProjectId1,
	ProjectType:          domain.EQUITY,
	RequestAmount:        30000,
	DepositAmount:        10000,
	DepositCheckStatus:   domain.NO_DEPOSIT,
	DepositAccountNumber: "123-123-123",
}

var mockUser1 = domain.User{
	Id:    mockUserId1,
	Name:  "홍길동",
	Email: "gildong@email.com",
}

var mockProject1 = domain.Project{
	Id:          mockProjectId1,
	ProjectType: domain.EQUITY,
	Name:        "태양광 프로젝트",
}

var mockShare1 = domain.Share{
	Id:     mockShareId1,
	Amount: 30000,
}

var mockUserVirtualAccount1 = domain.UserVirtualAccount{
	VaHolderNameKana: "가나다라",
}
