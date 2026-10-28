package http

import (
	"time"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
)

var (
	mockWithdrawalCreateResponse1 = adomain.WithdrawalCreateResponse{
		Id:           1,
		InvestorId:   "1",
		InvestorName: "ホン ギルドン",
		Amount:       1000000,
		Fee:          0,
		RequestDate:  jsonutils.JSONTime(time.Date(2025, 2, 9, 0, 0, 0, 0, time.Local)),
		Status:       domain.WITHDRAWAL_STATUS_INCOMPLETE,
		Created:      jsonutils.JSONTime(time.Date(2024, 12, 19, 0, 0, 0, 0, time.Local)),
		Updated:      jsonutils.JSONTime(time.Date(2024, 12, 19, 0, 0, 0, 0, time.Local)),
	}

	mockWithdrawalItem1 = adomain.WithdrawalItem{
		Id:           1,
		InvestorId:   "1",
		InvestorName: "ホン ギルドン",
		Amount:       1000000,
		Fee:          0,
		RequestDate:  jsonutils.JSONTime(time.Date(2025, 2, 9, 0, 0, 0, 0, time.Local)),
		CompleteDate: jsonutils.JSONTime{},
		Status:       domain.WITHDRAWAL_STATUS_INCOMPLETE,
		Created:      jsonutils.JSONTime(time.Date(2024, 12, 19, 0, 0, 0, 0, time.Local)),
	}

	mockWithdrawalDetailItem1 = adomain.WithdrawalDetailItem{
		WithdrawalItem: mockWithdrawalItem1,
		BankName:       "GMO Aozora",
		BankAccount:    "123456",
		BranchName:     "ﾆｼﾞｼﾃﾝ",
		DepositType:    domain.ORDINARY_ACCOUNT,
		UserType:       domain.PRIVATE,
		BranchCode:     lo.ToPtr("123"),
	}
)
