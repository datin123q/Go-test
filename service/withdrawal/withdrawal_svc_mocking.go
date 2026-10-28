package withdrawal

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
)

var (
	mockWithdrawal1 = domain.WithdrawalDetail{
		Withdrawal: domain.Withdrawal{
			Id:           1,
			InvestorId:   "1",
			InvestorName: "テスト",
			Amount:       1000,
			Fee:          100,
			RequestDate: func() jsonutils.JSONTime {
				date, _ := time.ParseInLocation(time.DateTime, "2024-02-01 00:00:00", time.UTC)
				return jsonutils.JSONTime(date)
			}(),
			CompleteDate: func() jsonutils.JSONTime {
				date, _ := time.ParseInLocation(time.DateTime, "2024-03-01 00:00:00", time.UTC)
				return jsonutils.JSONTime(date)
			}(),
			CloseDate: func() jsonutils.JSONTime {
				date, _ := time.ParseInLocation(time.DateTime, "2024-02-15 00:00:00", time.UTC)
				return jsonutils.JSONTime(date)
			}(),
			Status: domain.WITHDRAWAL_STATUS_COMPLETE,
		},
		User: domain.User{
			Id:          1,
			BankCode:    lo.ToPtr("001"),
			BranchCode:  lo.ToPtr("031"),
			BankAccount: "1234", // 6桁未満
			Depositor:   "テスト",
			DepositType: domain.ORDINARY_ACCOUNT,
		},
	}

	mockWithdrawal2 = domain.WithdrawalDetail{
		Withdrawal: domain.Withdrawal{
			Id:           2,
			InvestorId:   "2",
			InvestorName: "テスト テスト",
			Amount:       5000,
			Fee:          500,
			RequestDate: func() jsonutils.JSONTime {
				date, _ := time.ParseInLocation(time.DateTime, "2024-02-01 00:00:00", time.UTC)
				return jsonutils.JSONTime(date)
			}(),
			CompleteDate: func() jsonutils.JSONTime {
				return jsonutils.JSONTime(time.Time{})
			}(),
			CloseDate: func() jsonutils.JSONTime {
				return jsonutils.JSONTime(time.Time{})
			}(),
			Status: domain.WITHDRAWAL_STATUS_WAIT,
		},
		User: domain.User{
			Id:          2,
			BankCode:    lo.ToPtr("322"),
			BranchCode:  lo.ToPtr("018"),
			BankAccount: "123456", // 6桁
			Depositor:   "テスト テスト",
			DepositType: domain.ORDINARY_ACCOUNT,
		},
	}
)
