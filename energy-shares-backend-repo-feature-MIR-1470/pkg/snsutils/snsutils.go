package snsutils

import (
	"context"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

// Deprecated: Use Sns interface instead
func SendSMS(ctx context.Context, receiveNo string, object SnsObject) (string, error) {
	var num string
	receiveNo = strings.TrimSpace(receiveNo)
	if strings.HasPrefix(receiveNo, "+") {
		num = receiveNo
	} else {
		normalized := strings.TrimPrefix(receiveNo, "0")
		num = "+" + serviceutils.GetServiceCountryCode() + normalized
	}

	input := &sns.PublishInput{
		Message:     aws.String(object.Message()),
		PhoneNumber: aws.String(num),
		Subject:     aws.String(object.Subject()),
	}
	resp, err := factory.SmsMsgSender(ctx).Publish(input)
	if err != nil {
		return "", err
	}

	return *resp.MessageId, err
}

// Deprecated: Use Sns interface instead
func LoadTextTemplate(path string) (*template.Template, error) {
	t, err := template.ParseFiles(pathutils.GetPathOfAssets() + path)
	if err != nil {
		return nil, err
	}
	return t, nil
}
