package user

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	password        = "1234"
	encodedPassword = "encoded-1234"

	loginMockUser1 = domain.User{
		Id:          1,
		Email:       "test@example.com",
		Password:    encodedPassword,
		Salt:        "salt",
		Name:        "test-name",
		AuditStatus: domain.COMPLETED_AUDIT,
	}

	loginMockUser2 = domain.User{
		Id:          1,
		Email:       "test@example.com",
		Password:    encodedPassword,
		Salt:        "salt",
		Name:        "test-name",
		AuditStatus: domain.INVALIDITY_AUDIT,
	}

	loginMockUserLoginLock1 = domain.UserLoginLock{
		UserId:       1,
		LoginFailCnt: 1,
	}

	loginMockUserLoginLock4 = domain.UserLoginLock{
		UserId:       1,
		LoginFailCnt: 4,
	}

	loginMockUserLoginLock6 = domain.UserLoginLock{
		UserId:       1,
		LoginFailCnt: 6,
	}
)
