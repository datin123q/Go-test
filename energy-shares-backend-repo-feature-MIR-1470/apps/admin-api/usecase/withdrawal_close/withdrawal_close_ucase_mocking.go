package withdrawal_close

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
)

var (
	mockAdmin = domain.AdminUser{
		Id:            int64(1),
		CommonId:      "commonId-admin-user",
		Email:         "admin-user@energyshares.jp",
		Name:          "admin-user",
		Password:      "password1234",
		Salt:          "salt",
		MobilePhone:   "01011112222",
		LastLoginDate: jsonutils.JSONTime(time.Now()),
		LoginFailCnt:  0,
		Roles:         []string{"ROLE_PAYMENT", "ROLE_ADMIN"},
		IsSuperAdmin:  false,
	}

	mockWithdrawalCloseDetail1 = domain.WithdrawalCloseDetail{
		WithdrawalClose: domain.WithdrawalClose{
			Id:           1,
			Total:        100000,
			Amount:       100000,
			Fee:          0,
			CloserId:     1,
			CloseDate:    jsonutils.JSONTime(time.Date(2024, 12, 19, 0, 0, 0, 0, time.Local)),
			TransferDate: jsonutils.JSONTime(time.Date(2025, 2, 9, 0, 0, 0, 0, time.Local)),
			Status:       domain.WITHDRAWAL_REQUEST_STATUS_COMPLETE,
		},
		CloserName: "ホン ギルドン",
	}

	mockWithdrawalDetail1 = domain.WithdrawalDetail{
		Withdrawal: domain.Withdrawal{
			Id:           1,
			InvestorId:   "1",
			InvestorName: "テスト テスト",
			Amount:       5000,
			Fee:          500,
			RequestDate: func() jsonutils.JSONTime {
				date, _ := time.ParseInLocation(time.DateTime, "2024-02-01 00:00:00", time.Local)
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
			Id:          1,
			BankCode:    lo.ToPtr("322"),
			BranchCode:  lo.ToPtr("018"),
			BankAccount: "123456", // 6文字
			Depositor:   "テスト テスト",
			DepositType: domain.ORDINARY_ACCOUNT,
		},
	}

	mockWithdrawalDetail2 = domain.WithdrawalDetail{
		Withdrawal: domain.Withdrawal{
			Id:           2,
			InvestorId:   "2",
			InvestorName: "テスト",
			Amount:       1000,
			Fee:          100,
			RequestDate: func() jsonutils.JSONTime {
				date, _ := time.ParseInLocation(time.DateTime, "2024-02-01 00:00:00", time.Local)
				return jsonutils.JSONTime(date)
			}(),
			CompleteDate: func() jsonutils.JSONTime {
				date, _ := time.ParseInLocation(time.DateTime, "2024-03-01 00:00:00", time.Local)
				return jsonutils.JSONTime(date)
			}(),
			CloseDate: func() jsonutils.JSONTime {
				date, _ := time.ParseInLocation(time.DateTime, "2024-02-15 00:00:00", time.Local)
				return jsonutils.JSONTime(date)
			}(),
			Status: domain.WITHDRAWAL_STATUS_COMPLETE,
		},
		User: domain.User{
			Id:          2,
			BankCode:    lo.ToPtr("001"),
			BranchCode:  lo.ToPtr("031"),
			BankAccount: "1234", // 6文字未満
			Depositor:   "テスト",
			DepositType: domain.ORDINARY_ACCOUNT,
		},
	}
)
