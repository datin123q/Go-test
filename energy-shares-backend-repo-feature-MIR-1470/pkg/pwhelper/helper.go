package pwhelper

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"

	"golang.org/x/crypto/pbkdf2"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/tool"
)

const SALT_LENGTH = 10

type PasswordHelper interface {
	EncodePassword(password, salt string) string
	IsSamePassword(storedPassword, inputPassword string) bool
	GenerateSalt() string
}

type passwordHelper struct{}

func NewPasswordHelper() PasswordHelper {
	return &passwordHelper{}
}

func (p passwordHelper) EncodePassword(password, salt string) string {
	newPassword := pbkdf2.Key([]byte(password), []byte(salt), 10000, 50, sha256.New)
	return fmt.Sprintf("%x", newPassword)
}

func (p passwordHelper) IsSamePassword(storedPassword, inputPassword string) bool {
	return subtle.ConstantTimeCompare([]byte(storedPassword), []byte(inputPassword)) == 1
}

func (p passwordHelper) GenerateSalt() string {
	v, _ := tool.RandomString(SALT_LENGTH)
	return v
}
