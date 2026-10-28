package awsutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	awsv1 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-faster/errors"
)

func InitAws(region string) (*session.Session, error) {

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.

	sess, err := session.NewSession(&awsv1.Config{
		Region: aws.String(region)},
	)
	return sess, err
}

func InitAwsV2(ctx context.Context, region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return aws.Config{}, errors.Wrap(err, "fail aws loadDefaultConfig")
	}
	return cfg, nil
}

func InitSQSClient(cfg aws.Config, isLocal bool, baseEndPoint string) *sqs.Client {
	if isLocal {
		return sqs.NewFromConfig(cfg, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(baseEndPoint)
		})
	}
	return sqs.NewFromConfig(cfg)
}
