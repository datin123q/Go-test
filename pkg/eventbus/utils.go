package eventbus

import (
	"github.com/sirupsen/logrus"
)

func InitializeEventBus(cfg Config, timeout int) IEventBus {
	var (
		bus IEventBus
		err error
	)

	switch cfg.Authentication {
	case AuthenticationAWSMskIAM:
		bus, err = NewMskEventBus(cfg.Region, cfg.Brokers, cfg.Topic, timeout)
	default:
		bus, err = NewEventBus(cfg.Brokers, cfg.Topic, timeout)
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"brokers": cfg.Brokers,
			"topic":   cfg.Topic,
			"timeout": timeout,
			"auth":    cfg.Authentication,
		}).Error("fail new event bus")
		panic(err)
	}

	return bus
}
