package investment

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockShare1 = domain.Share{
		Id:          "SHARE-1",
		OwnerId:     1000,
		ProjectId:   2000,
		ProjectType: domain.EQUITY,
	}

	mockInvestment1 = domain.Investment{
		Share:             mockShare1,
		InvestmentId:      "SHARE-1",
		BankName:          "ジャパンネット銀行",
		BankAccountNumber: "2434551",
		Depositor:         "ペ·　ジヨンフン",
	}

	mockUser1 = domain.User{
		Id:          1000,
		BankAccount: "2434551",
		BankName:    "ジャパンネット銀行",
		Depositor:   "ペ·　ジヨンフン",
	}

	mockContract1 = domain.Contract{
		Id:                   5000,
		ShareId:              "SHARE-1",
		ContractAfterPageUrl: "https://cdn.dev.energyshares.jp.illuminarean.com/share_file/2022-01-27/c7p2l5v8s5nbrq5js5ug.pdf",
	}

	mockDepositRequest1 = domain.DepositRequest{
		Id:      3000,
		ShareId: "SHARE-1",
	}

	mockUserVirtualAccount1 = domain.UserVirtualAccount{
		UserId:      1000,
		ProjectType: domain.EQUITY,
	}

	mockRepaymentDividendSchedule1 = domain.RepaymentDividendSchedule{
		Id:          4000,
		ReturnType:  domain.REPAYMENT_RETURN,
		ProjectId:   2000,
		ExecuteDate: jsonutils.JSONTime(time.Unix(0, 0)),
	}
)
