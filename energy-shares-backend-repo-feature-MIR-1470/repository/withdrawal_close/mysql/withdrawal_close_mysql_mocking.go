package mysql

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"xorm.io/xorm"
)

var (
	mockWithdrawalClose1 = domain.WithdrawalClose{
		Id:        int64(1),
		CloserId:  int64(1),
		CloseDate: jsonutils.JSONTime(time.Now().Local()),
		Total:     int64(100),
		Amount:    int64(100),
		Fee:       int64(0),
		Status:    domain.WITHDRAWAL_REQUEST_STATUS_WAIT,
	}

	mockWithdrawalClose2 = domain.WithdrawalClose{
		Id:           int64(2),
		CloserId:     int64(1),
		CloseDate:    jsonutils.JSONTime(time.Now().Local()),
		TransferDate: jsonutils.JSONTime(time.Now().Local()),
		Total:        int64(300),
		Amount:       int64(300),
		Fee:          int64(0),
		Status:       domain.WITHDRAWAL_REQUEST_STATUS_COMPLETE,
	}

	mockAdminUser1 = domain.AdminUser{
		Id:   int64(1),
		Name: "test1",
	}
)

type dbFactoryMock struct {
}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.ShowSQL(true)

	xormEngine.Sync(new(domain.WithdrawalClose))
	xormEngine.Sync(new(domain.AdminUser))

	xormEngine.Insert(mockWithdrawalClose1, mockWithdrawalClose2)
	xormEngine.Insert(mockAdminUser1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
