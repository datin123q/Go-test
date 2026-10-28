package withdrawal

import (
	"time"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
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
			Status:   domain.WITHDRAWAL_STATUS_WAIT,
			IsManual: true,
		},
		User: domain.User{
			Id:          1,
			BankCode:    lo.ToPtr("322"),
			BranchCode:  lo.ToPtr("018"),
			BankAccount: "123456", // 6桁
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
			Status:   domain.WITHDRAWAL_STATUS_COMPLETE,
			IsManual: true,
		},
		User: domain.User{
			Id:          2,
			BankCode:    lo.ToPtr("001"),
			BranchCode:  lo.ToPtr("031"),
			BankAccount: "1234", // 6桁未満
			Depositor:   "テスト",
			DepositType: domain.ORDINARY_ACCOUNT,
		},
	}

	mockWithdrawalDetail3 = domain.WithdrawalDetail{
		Withdrawal: domain.Withdrawal{
			Id:           3,
			InvestorId:   "2",
			InvestorName: "テスト",
			Amount:       3000,
			Fee:          350,
			RequestDate: func() jsonutils.JSONTime {
				date, _ := time.ParseInLocation(time.DateTime, "2024-02-01 00:00:00", time.Local)
				return jsonutils.JSONTime(date)
			}(),
			CompleteDate: func() jsonutils.JSONTime {
				return jsonutils.JSONTime(time.Time{})
			}(),
			CloseDate: func() jsonutils.JSONTime {
				date, _ := time.ParseInLocation(time.DateTime, "2024-02-15 00:00:00", time.Local)
				return jsonutils.JSONTime(date)
			}(),
			Status: domain.WITHDRAWAL_STATUS_FAIL,
		},
		User: domain.User{
			Id:          2,
			BankCode:    lo.ToPtr("001"),
			BranchCode:  lo.ToPtr("031"),
			BankAccount: "1234", // 6桁未満
			Depositor:   "テスト",
			DepositType: domain.ORDINARY_ACCOUNT,
		},
	}

	mockWithdrawalItem1 = adomain.WithdrawalItem{
		Id:           mockWithdrawalDetail1.Withdrawal.Id,
		InvestorId:   mockWithdrawalDetail1.Withdrawal.InvestorId,
		InvestorName: mockWithdrawalDetail1.Withdrawal.InvestorName,
		Amount:       mockWithdrawalDetail1.Withdrawal.Amount,
		Fee:          mockWithdrawalDetail1.Withdrawal.Fee,
		RequestDate:  mockWithdrawalDetail1.Withdrawal.RequestDate,
		CompleteDate: mockWithdrawalDetail1.Withdrawal.CompleteDate,
		Status:       mockWithdrawalDetail1.Withdrawal.Status,
		IsManual:     mockWithdrawalDetail1.Withdrawal.IsManual,
		Created:      mockWithdrawalDetail1.Withdrawal.Created,
	}
	mockWithdrawalItem2 = adomain.WithdrawalItem{
		Id:           mockWithdrawalDetail2.Withdrawal.Id,
		InvestorId:   mockWithdrawalDetail2.Withdrawal.InvestorId,
		InvestorName: mockWithdrawalDetail2.Withdrawal.InvestorName,
		Amount:       mockWithdrawalDetail2.Withdrawal.Amount,
		Fee:          mockWithdrawalDetail2.Withdrawal.Fee,
		RequestDate:  mockWithdrawalDetail2.Withdrawal.RequestDate,
		CompleteDate: mockWithdrawalDetail2.Withdrawal.CompleteDate,
		Status:       mockWithdrawalDetail2.Withdrawal.Status,
		IsManual:     mockWithdrawalDetail2.Withdrawal.IsManual,
		Created:      mockWithdrawalDetail2.Withdrawal.Created,
	}
	mockWithdrawalItem3 = adomain.WithdrawalItem{
		Id:           mockWithdrawalDetail3.Withdrawal.Id,
		InvestorId:   mockWithdrawalDetail3.Withdrawal.InvestorId,
		InvestorName: mockWithdrawalDetail3.Withdrawal.InvestorName,
		Amount:       mockWithdrawalDetail3.Withdrawal.Amount,
		Fee:          mockWithdrawalDetail3.Withdrawal.Fee,
		RequestDate:  mockWithdrawalDetail3.Withdrawal.RequestDate,
		CompleteDate: mockWithdrawalDetail3.Withdrawal.CompleteDate,
		Status:       mockWithdrawalDetail3.Withdrawal.Status,
		IsManual:     mockWithdrawalDetail3.Withdrawal.IsManual,
		Created:      mockWithdrawalDetail3.Withdrawal.Created,
	}

	mockTransfer1 = adomain.WithdrawalTransferItem{
		Id:             1,
		RequestId:      1,
		TransactionId:  "transaction-id-1",
		TransferDate:   jsonutils.JSONTime(time.Now()),
		TransferAmount: 5000,
		TransferFee:    500,
		TransferStatus: domain.TRANSFER_STATUS_SUCCESS,
		ErrorCode:      "",
		ErrorMsg:       "",
	}
	mockTransfer2 = adomain.WithdrawalTransferItem{
		Id:             2,
		RequestId:      2,
		TransactionId:  "transaction-id-1",
		TransferDate:   jsonutils.JSONTime(time.Now()),
		TransferAmount: 5000,
		TransferFee:    500,
		TransferStatus: domain.TRANSFER_STATUS_FAIL,
		ErrorCode:      "error-code",
		ErrorMsg:       "error-msg",
	}
)
