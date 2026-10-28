package mysql

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockRepaymentDividendSchedule1 = domain.RepaymentDividendSchedule{
		Id:              1000,
		ProjectId:       2000,
		ExecuteDate:     jsonutils.JSONTime(time.Date(2024, 2, 7, 0, 0, 0, 0, time.UTC)),
		ReturnType:      domain.REPAYMENT_RETURN,
		RepaymentAmount: 100,
	}

	mockRepaymentDividendSchedule2 = domain.RepaymentDividendSchedule{
		Id:              1001,
		ProjectId:       2000,
		ExecuteDate:     jsonutils.JSONTime(time.Date(2024, 2, 8, 0, 0, 0, 0, time.UTC)),
		ReturnType:      domain.REPAYMENT_RETURN,
		RepaymentAmount: 200,
	}

	mockRepaymentDividendSchedule3 = domain.RepaymentDividendSchedule{
		Id:             1002,
		ProjectId:      2000,
		ExecuteDate:    jsonutils.JSONTime(time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC)),
		ReturnType:     domain.DIVIDEND_RETURN,
		DividendAmount: 300,
	}

	mockRepaymentDividendSchedule4 = domain.RepaymentDividendSchedule{
		Id:          1003,
		ProjectId:   2001,
		ExecuteDate: jsonutils.JSONTime(time.Date(2024, 2, 8, 0, 0, 0, 0, time.UTC)),
		ReturnType:  domain.REPAYMENT_RETURN,
	}

	mockRepaymentDividendScheduleCreate = domain.RepaymentDividendSchedule{
		ProjectId:   2000,
		ExecuteDate: jsonutils.JSONTime(time.Date(2024, 2, 7, 0, 0, 0, 0, time.UTC)),
		ReturnType:  domain.REPAYMENT_RETURN,
	}

	mockRepaymentDividendSchedule1Update = domain.RepaymentDividendSchedule{
		Id:          1000,
		ProjectId:   3000,
		ExecuteDate: jsonutils.JSONTime(time.Date(2024, 2, 7, 0, 0, 0, 0, time.UTC)),
		ReturnType:  domain.REPAYMENT_RETURN,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.RepaymentDividendSchedule))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockRepaymentDividendSchedule1, mockRepaymentDividendSchedule2, mockRepaymentDividendSchedule3, mockRepaymentDividendSchedule4)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
