package mysql

import (
	"context"
	"encoding/json"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockMailRetry1 = domain.MailRetry{
		Id: 1001,
		EmailRequest: json.RawMessage(`{
			"templateName": "welcome_template",
			"mailData": [
				{
					"templateFields": {
						"firstName": "John",
						"lastName": "Doe"
					},
					"receivers": ["john.doe@example.com"],
					"attachmentFiles": [],
					"sendType": "FAST",
					"threading": false
				}
			],
			"serviceName": "MyService",
			"serviceEnv": "production"
		}`),
		Receivers: json.RawMessage(`[
			"john.doe@example.com",
			"jane.doe@example.com"
		]`),
		TemplateName: "TEST_TEMPLATE",
		RetryStatus:  "PENDING",
	}
	mockMailRetry2 = domain.MailRetry{
		Id: 1002,
		EmailRequest: json.RawMessage(`{
			"templateName": "welcome_template",
			"mailData": [
				{
					"templateFields": {
						"firstName": "Jane",
						"lastName": "Doe"
					},
					"receivers": ["jane.doe@example.com"],
					"attachmentFiles": [],
					"sendType": "NORMAL",
					"threading": false
				}
			],
			"serviceName": "MyService",
			"serviceEnv": "production"
		}`),
		Receivers: json.RawMessage(`[
			"john.doe@example.com",
			"jane.doe@example.com"
		]`),
		TemplateName: "TEST_TEMPLATE",
		RetryStatus:  "PENDING",
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.MailRetry))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockMailRetry2)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
