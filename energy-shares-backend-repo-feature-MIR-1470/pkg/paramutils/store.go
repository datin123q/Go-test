package paramutils

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
)

type Store struct {
	Region  string
	Session *session.Session
}

func NewStore(region string) (*Store, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to session connect")
	}
	return &Store{
		Region:  region,
		Session: sess,
	}, nil
}

func (s *Store) getParameter(uri string, withDecryption bool) (*ssm.GetParameterOutput, error) {

	ssmClient := ssm.New(s.Session, aws.NewConfig().WithRegion(s.Region))
	output, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(uri),
		WithDecryption: aws.Bool(withDecryption),
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

// ParamRead はパラメータストアから値を読み込む関数です。
// 常にwithDecryption値を`true`に設定して、暗号化されたパラメータの場合は値を複号化して読み込むようにします。
func (s *Store) ParamRead(env, prefix, parameter string) (*ssm.GetParameterOutput, error) {
	uri := fmt.Sprintf("%s/%s", prefix, env)
	if parameter != "" {
		uri = fmt.Sprintf("%s/%s", uri, parameter)
	}
	result, err := s.getParameter(uri, true)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("parameter is empty")
	}
	return result, nil
}
