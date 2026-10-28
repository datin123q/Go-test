package egift

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/sirupsen/logrus"
)

type egiftClientStub struct {
	logger factory.LoggerFactory
}

func (e egiftClientStub) RewardSignUp(ctx context.Context, marketingCode string, signUpRequestId string) error {
	e.logger.Entry(ctx).
		WithFields(logrus.Fields{
			"marketingCode":   marketingCode,
			"signUpRequestId": signUpRequestId,
		}).
		Info("egiftClientStub RewardSignUp")

	return nil
}
