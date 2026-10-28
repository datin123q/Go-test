package mail

import (
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockEmailListResponse = emailutils.EmailListResponse{
		Content:       []emailutils.Email{mailEmail1, mailEmail2},
		TotalElements: 2,
	}

	mailEmail1 = emailutils.Email{
		ID:           1001,
		Subject:      "メール件名1",
		Receivers:    []string{"a@example.com", "b@example.com"},
		ServiceName:  "ESJP",
		ServiceEnv:   "DEV",
		Status:       "SENT",
		SendType:     "FAST",
		Sender:       "info@example.com",
		TemplateName: "NORMAL_MAIL",
	}

	mailEmail2 = emailutils.Email{
		ID:           1002,
		Subject:      "メール件名2",
		Receivers:    []string{"a@example.com", "b@example.com"},
		ServiceName:  "ESJP",
		ServiceEnv:   "DEV",
		Status:       "FAILED",
		SendType:     "NORMAL",
		Sender:       "info@example.com",
		TemplateName: "NORMAL_MAIL",
	}

	mockEmailResponse = emailutils.EmailResponse{
		ID:             1001,
		ServiceName:    "ESJP",
		ServiceEnv:     "DEV",
		Subject:        "メール件名1",
		Sender:         "info@example.com",
		Receivers:      []string{"a@example.com", "b@example.com"},
		Body:           "メール内容",
		TemplateName:   "NORMAL_MAIL",
		TemplateFields: "",
		SendType:       "FAST",
		Status:         "SENT",
		SendDateTime:   jsonutils.JSONTime{},
		CreatedDate:    nil,
		ModifiedDate:   nil,
	}

	mockMailFailed = domain.MailFailed{
		Id:           1003,
		EmailRequest: rawEmailRequester,
		Receivers:    rawReceiver,
		TemplateName: "NORMAL_MAIL",
		RetryStatus:  "FAILED",
		Created:      jsonutils.JSONTime{},
	}

	rawEmailRequester = []byte(`{
		"templateName": "NORMAL_MAIL",
		"mailData": [
			{
				"templateFields": {
					"field1": "value1",
					"field2": "value2"
				},
				"receivers": [
					"receiver1@email.com",
					"receiver2@email.com"
				],
				"attachmentFiles": [],
				"sendType": "NORMAL",
				"threading": true
			}
		],
		"serviceName": "ESJP",
		"serviceEnv": "DEV"
	}`)

	rawReceiver = []byte(`[
		"receiver1@email.com",
		"receiver2@email.com"
	]`)

	mockMailFailedItem = adomain.MailFailedItem{
		Id:           1003,
		TemplateName: "NORMAL_MAIL",
		Created:      jsonutils.JSONTime{},
		Receivers:    []string{"receiver1@email.com", "receiver2@email.com"},
	}

	mockEmailTemplateResponse = emailutils.MailTemplateResponse{
		Subject:      "[EnergyShares] {{Subject}}",
		Content:      "<!doctype html>\\n<html>\\n<head>\\n<meta charset=\\\"utf-8\\\">\\n<title>EnergyShares</title>\\n</head>\\n<body>\\n{{Content}}\\n</body>\\n</html>\\n",
		TemplateName: "NORMAL_MAIL",
	}
)
