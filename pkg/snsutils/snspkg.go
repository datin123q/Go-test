package snsutils

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
)

type SnsObject interface {
	Message() string
	Subject() string
}

type Sns interface {
	SendSMS(ctx context.Context, receiveNo string, object SnsObject) (string, error)
	LoadTextTemplate(ctx context.Context, path string) (*template.Template, error)
	Subscribe(ctx context.Context, topicArn, protocol, endpoint string) (string, error)
	Unsubscribe(ctx context.Context, subscriptionArn string) error
}

type snsPkg struct {
	cfg        *config.Config
	snsFactory factory.SnsFactory
}

func New(
	cfg *config.Config,
	snsFactory factory.SnsFactory,
) Sns {
	return &snsPkg{
		cfg:        cfg,
		snsFactory: snsFactory,
	}
}

func (s snsPkg) SendSMS(ctx context.Context, receiveNo string, object SnsObject) (string, error) {
	var num string
	receiveNo = strings.TrimSpace(receiveNo)
	if strings.HasPrefix(receiveNo, "+") {
		num = receiveNo
	} else {
		normalized := strings.TrimPrefix(receiveNo, "0")
		num = "+" + s.cfg.CountryCode + normalized
	}

	input := &sns.PublishInput{
		Message:     aws.String(object.Message()),
		PhoneNumber: aws.String(num),
		Subject:     aws.String(object.Subject()),
	}
	resp, err := s.snsFactory.SmsMsgSender(ctx).Publish(input)
	if err != nil {
		return "", err
	}

	return *resp.MessageId, err
}

func (s snsPkg) LoadTextTemplate(ctx context.Context, path string) (*template.Template, error) {
	t, err := template.ParseFiles(pathutils.GetPathOfAssets() + path)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s snsPkg) Subscribe(ctx context.Context, topicArn, protocol, endpoint string) (string, error) {

	params := &sns.SubscribeInput{
		Protocol: aws.String(protocol),
		TopicArn: aws.String(topicArn),
		Endpoint: aws.String(endpoint),
	}

	output, err := s.snsFactory.SmsMsgSender(ctx).Subscribe(params)
	if err != nil {
		return "", err
	}

	return *output.SubscriptionArn, err
}

func (s snsPkg) Unsubscribe(ctx context.Context, subscriptionArn string) error {

	params := &sns.UnsubscribeInput{
		SubscriptionArn: aws.String(subscriptionArn),
	}

	output, err := s.snsFactory.SmsMsgSender(ctx).Unsubscribe(params)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return err
}
