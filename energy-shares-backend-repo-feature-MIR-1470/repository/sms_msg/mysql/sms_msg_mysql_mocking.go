package mysql

import (
	"context"
	"time"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type dbFactoryMock struct{}

var (
	mockSmsMsg = domain.SmsMsg{
		Id:         1000,
		MsgId:      "MSG100",
		Title:      "タイトル",
		Name:       "ホン・ギルドン",
		Sender:     "test@email.com",
		Status:     domain.REQUEST_SMS,
		SmsType:    domain.PASSWORD_RESET_COMPLETE_SMS,
		SmsContent: "[EnergyShares]\nエナジーシェアーズをご利用頂き、誠にありがとうございます。\nお客様のパスワードが変更されましたので、ご案内致します。",
		Count:      1,
		SendDate:   jsonutils.JSONTime(time.Now()),
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.SmsMsg))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockSmsMsg)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
