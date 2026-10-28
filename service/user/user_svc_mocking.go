package user

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	mockUserChangePasswordBefore = domain.User{
		Id:       1000,
		Email:    "ilmuser1@email.com",
		Password: "5ec02b91a4b59c6f59dd5fbe4ca649ece4fa8568cdb8ba36cf41426e8805522ba4e2aeac19a4821501cf609126ab01df2566", // password
		Salt:     "salt",
	}

	mockUserChangePasswordAfter = domain.User{
		Id:       1000,
		Email:    "ilmuser1@email.com",
		Password: "336855f6e57c3e408b96ed8045c7110702a4ebf8abc7979fbf8bf8a29e129e08a74c4bf2c39efa1bbe127f470d78d941b26d", // new_password
		Salt:     "salt",
	}
)
