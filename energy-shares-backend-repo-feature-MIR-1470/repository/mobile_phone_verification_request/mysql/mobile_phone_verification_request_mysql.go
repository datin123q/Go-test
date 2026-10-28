package mysql

import (
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type mobilePhoneVerificationRequestMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.MobilePhoneVerificationRequestRepository {
	return &mobilePhoneVerificationRequestMysqlRepository{
		dbFactory: dbFactory,
	}
}
