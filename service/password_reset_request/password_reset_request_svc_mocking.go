package password_reset_request

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockPasswordResetRequest1 = domain.PasswordResetRequest{
		Id:        1000,
		Email:     "test@email.com",
		SendCount: 0,
		Expire:    jsonutils.JSONTime(time.Now().Add(time.Minute * 10).In(time.Local)),
	}

	mockPasswordResetRequest2 = domain.PasswordResetRequest{
		Id:        1000,
		Email:     "test@email.com",
		SendCount: 1,
		Expire:    jsonutils.JSONTime(time.Now().Add(time.Minute * 10).In(time.Local)),
	}

	mockPasswordResetRequest3 = domain.PasswordResetRequest{
		Id:        1000,
		Email:     "test@email.com",
		SendCount: 0,
		Expire:    jsonutils.JSONTime(time.Now().Add(time.Minute * -10).In(time.Local)),
	}
)
