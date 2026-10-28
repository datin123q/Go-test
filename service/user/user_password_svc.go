package user

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func (u userService) ChangePassword(ctx context.Context, user domain.User, password string) error {

	// check if password contains the email
	if !u.IsValidPassword(ctx, user, password) {
		return errors.Wrap(cdomain.ErrItemInvalid, "invalid password")
	}

	// return error if same password
	inputPassword := u.pwhelper.EncodePassword(password, user.Salt)
	if u.pwhelper.IsSamePassword(user.Password, inputPassword) {
		return errors.Wrap(cdomain.ErrItemConflict, "same password")
	}

	// update password
	user.Salt = u.pwhelper.GenerateSalt()
	user.Password = u.pwhelper.EncodePassword(password, user.Salt)
	if err := u.userRepository.Update(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	return nil
}

func (u userService) IsValidPassword(ctx context.Context, user domain.User, password string) bool {

	// check if password contains the email
	if strings.Contains(password, user.Email) {
		return false
	}

	return true
}
