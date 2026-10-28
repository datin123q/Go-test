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
	mockMailFailed1 = domain.MailFailed{
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
			"john.doe@example.com"
		]`),
		TemplateName: "NORMAL_MAIL",
		RetryStatus:  "FAILED",
	}
	mockMailFailed2 = domain.MailFailed{
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
			"jane.doe@example.com"
		]`),
		TemplateName: "NORMAL_MAIL",
		RetryStatus:  "FAILED",
	}
	mockMailFailed3 = domain.MailFailed{
		Id: 1003,
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
			"song.doe@example.com"
		]`),
		TemplateName: "NORMAL_MAIL",
		RetryStatus:  "FAILED",
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.MailFailed))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockMailFailed2, mockMailFailed3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
