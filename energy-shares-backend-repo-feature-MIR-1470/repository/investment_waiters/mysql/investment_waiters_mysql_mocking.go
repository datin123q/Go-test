package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockOpenFundingStatus    = domain.OPEN_FUNDING
	mockSuccessFundingStatus = domain.SUCCESS_FUNDING
	mockSuccessInvestStatus  = domain.SUCCESS_INVEST

	mockInvestmentWaiters1 = domain.InvestmentWaiters{
		Id:            1000,
		UserId:        3000,
		ProjectId:     2000,
		ProjectStatus: &mockOpenFundingStatus,
	}

	mockInvestmentWaiters2 = domain.InvestmentWaiters{
		Id:        1001,
		UserId:    3000,
		ProjectId: 2000,
		// ShareId は xorm:"<-" のため Insert では書き込まれない。ここでは「クエリ経由で
		// 期待する結果値」のドキュメントとして書いているだけで、実際の値は
		// mockInvestmentWaitersWinning1(下記)から中間テーブル経由で解決される(MIR-1118)。
		ShareId:       "SHARE_ID1",
		InvestStatus:  &mockSuccessInvestStatus,
		ProjectStatus: &mockOpenFundingStatus,
	}

	mockInvestmentWaiters3 = domain.InvestmentWaiters{
		Id:            1002,
		UserId:        3000,
		ProjectId:     3000,
		ProjectStatus: &mockSuccessFundingStatus,
	}

	mockUser1 = domain.User{
		Id: 3000,
	}

	mockProject1 = domain.Project{
		Id:            2000,
		ProjectStatus: mockOpenFundingStatus,
	}

	mockProject2 = domain.Project{
		Id:            3000,
		ProjectStatus: mockSuccessFundingStatus,
	}

	mockShare1 = domain.Share{
		Id:           "SHARE_ID1",
		ProjectId:    2000,
		OwnerId:      3000,
		InvestStatus: mockSuccessInvestStatus,
	}

	// mockInvestmentWaiters2.ShareId は investment_waiters_winning 経由のサブクエリでのみ
	// 埋まる読み取り専用フィールドになったため(MIR-1118)、対応する中間テーブルの行を用意する。
	mockInvestmentWaitersWinning1 = domain.InvestmentWaitersWinning{
		InvestmentWaitersId: mockInvestmentWaiters2.Id,
		ShareId:             mockShare1.Id,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.InvestmentWaiters))
	xormEngine.Sync(new(domain.InvestmentWaitersWinning))
	xormEngine.Sync(new(domain.User))
	xormEngine.Sync(new(domain.Project))
	xormEngine.Sync(new(domain.Share))

	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockInvestmentWaiters1, mockInvestmentWaiters2, mockInvestmentWaiters3)
	xormEngine.Insert(mockInvestmentWaitersWinning1)
	xormEngine.Insert(mockUser1)
	xormEngine.Insert(mockProject1, mockProject2)
	xormEngine.Insert(mockShare1)
	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
