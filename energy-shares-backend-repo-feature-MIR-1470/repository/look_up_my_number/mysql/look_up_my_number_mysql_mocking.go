package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockLookUpMyNumber1 = domain.LookUpMyNumber{
		Id:         1000,
		MyNumberId: 2000,
		MyNumber:   "bBBbNyUqDZTowZNAcdPDgzlj9GsnJhMR/sxC4HGT5EoLcK2omgsgOXWZLOkZPun5", // encrypted "410d3741-7180-4024-a969-648723a57bbe",
	}

	mockLookUpMyNumber2 = domain.LookUpMyNumber{
		Id:         2,
		MyNumberId: 2,
		MyNumber:   "abcd",
	}

	mockLookUpMyNumberDeleted1 = domain.LookUpMyNumber{
		Id:         1001,
		MyNumberId: 2001,
		MyNumber:   "encrypted_my_number",
	}

	mockLookUpMyNumberCreate1 = domain.LookUpMyNumber{
		Id:         3000,
		MyNumberId: 3000,
		MyNumber:   "encrypted_my_number",
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.LookUpMyNumber))
	xormEngine.ShowSQL(true)

	// テストデータ挿入
	xormEngine.Insert(mockLookUpMyNumber1, mockLookUpMyNumber2, mockLookUpMyNumberDeleted1)
	xormEngine.Delete(mockLookUpMyNumberDeleted1)

	return xormEngine, nil
}
