package domain

import (
	"strconv"
	"testing"

	"github.com/hillfolk/goutils/test"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

func TestUtils_GetPostalCode(t *testing.T) {

	code, err := modelhelper.GetPostalCode()
	test.Ok(t, err)

	postalcode, err := strconv.ParseInt(code, 10, 64)
	test.Ok(t, err)
	test.Assert(t, postalcode >= 10000 && postalcode <= 99999, "postal code is not a number")
}
