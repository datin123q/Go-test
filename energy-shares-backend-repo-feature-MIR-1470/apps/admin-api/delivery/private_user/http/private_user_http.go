package http

import (
	"github.com/pangpanglabs/echoswagger/v2"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type privateUserHttpDelivery struct {
	privateUserUsecase adomain.PrivateUserUsecase
	logger             factory.LoggerFactory
	extractor          jwthelper.UserExtractor
}

func New(
	api echoswagger.ApiRoot,
	privateUserUsecase adomain.PrivateUserUsecase,
	logger factory.LoggerFactory,
	extractor jwthelper.UserExtractor,
) *privateUserHttpDelivery {
	handler := &privateUserHttpDelivery{
		privateUserUsecase: privateUserUsecase,
		logger:             logger,
		extractor:          extractor,
	}

	_ = api.Group("v1/private-users", "api/v1/private-users")

	return handler
}
