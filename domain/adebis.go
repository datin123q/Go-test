package domain

import (
	"net/http"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/adebis"
)

type AdEbis struct {
}

func (AdEbis) SetITPMeasuresCookie(req *http.Request, res http.ResponseWriter) {
	adEbis := adebis.NewSetEbisCookie(req, res)
	adEbis.SetEbisCookie()
}
