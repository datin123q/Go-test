package jobs

import (
	"context"
	"time"

	"energyshares-job-scheduler/pkg/gmo_aozora"

	client "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

const VIRTUAL_ACCOUNT_ISSUE_COUNT = "10"
const VIRTUAL_ACCOUNT_TYPE = "2"

type AozoraJobs struct{}

// 仮想口座発行コード

func ModelToErrorResponse(v interface{}) (*client.ErrorResponse, error) {
	e := &client.ErrorResponse{}
	err := mapstructure.Decode(v, e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func ModelToTransferErrorResponse(v interface{}) (*client.TransferError, error) {
	e := &client.TransferError{}
	err := mapstructure.Decode(v, e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (a *AozoraJobs) RefreshToken(ctx context.Context) error {
	if err := gmo_aozora.RefreshToken(ctx); err != nil {
		logrus.WithFields(logrus.Fields{
			"desc": "fail refreshToken",
			"err":  err,
			"time": time.Now(),
		}).Error("Fail Refresh Token")
		return err
	}

	return nil
}

func (a *AozoraJobs) GetAccountInfo(ctx context.Context) error {
	if err := gmo_aozora.GetAccountInfo(ctx, gmo_aozora.BankAccounts); err != nil {
		logrus.WithFields(logrus.Fields{
			"desc":      "fail Get Account Info",
			"accountNo": serviceutils.GetServiceAccountNo(),
			"err":       err,
			"time":      time.Now(),
		}).Error("fail Get Account Info")
		return err
	}
	return nil
}
