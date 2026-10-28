package auth

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	mockUser1 = domain.User{
		Id:          int64(1),
		Email:       "mockUser1@test.com",
		Name:        "mockUser1",
		UserType:    domain.PRIVATE,
		AuditStatus: domain.COMPLETED_AUDIT,
	}
)
