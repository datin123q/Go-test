package actclient

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

//加盟店名：energyshares
//成果対象ID：100

// DONE 成果報酬型マーケティング企業とのセッション接続
// 会員登録完了後、マーケティング企業にセッションを転送して成果報酬型マーケティングを進行できるようにする。
const (
	COMPANY_NAME string = "energyshares"
	TARGET_ID    string = "100"
)

const BASE_URL = "https://is.accesstrade.net/cgi-bin/isatV2/"

type Client struct {
	CompanyName string // 会社名
	TargetId    string // 成果対象ID
	Client      *http.Client
}

func NewClient() *Client {
	return &Client{
		CompanyName: COMPANY_NAME,
		TargetId:    TARGET_ID,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func MakeSignUpRewardUrl(companyName, targetId, marketingCode string, signRequestId string) string {
	return fmt.Sprintf(BASE_URL+"%s/isatWeaselV2.cgi?result_id=%s&verify=%s&rk=%s", companyName, targetId, signRequestId, marketingCode)
}

func MakePurchaseRewardUrl(companyName, targetId, marketingCode, investmentId, value string) string {
	return fmt.Sprintf(BASE_URL+"%s/isatWeaselV2.cgi?result_id=%s&verify=%s&value=%s&rk=%s", companyName, targetId, investmentId, value, marketingCode)
}
func (c *Client) SignUpReward(marketingCode string, signRequestId string) error {
	v := CompanyRemover(marketingCode)

	logrus.WithFields(logrus.Fields{
		"companyName":   c.CompanyName,
		"targetId":      c.TargetId,
		"marketingCode": marketingCode,
		"rk":            v,
		"signRequestId": signRequestId,
	}).Info("sign up reward")
	if serviceutils.GetEnv() != "PROD" {
		return errors.New("sign up reward is being ignored because the environment is not PROD")
	}

	url := MakeSignUpRewardUrl(c.CompanyName, c.TargetId, v, signRequestId)
	resp, err := c.Client.Post(url, "text/plain", nil)
	if err != nil {
		return errors.Wrap(err, "failed to signup reward post request")
	}
	logrus.WithFields(
		logrus.Fields{
			"status": resp.Status,
			"header": resp.Header,
			"body":   resp.Body,
		}).Info("sign up reward response")

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	return nil
}

func CompanyRemover(marketingCode string) string {
	idx := strings.Index(marketingCode, "_")
	v := marketingCode[idx+1:]
	return v
}

func (c *Client) PurchaseReward(marketingCode, investmentId, value string) error {
	url := MakePurchaseRewardUrl(c.CompanyName, c.TargetId, investmentId, value, marketingCode)
	resp, err := c.Client.Post(url, "text/plain", nil)
	if err != nil {
		return errors.Wrap(err, "failed to signup reward post request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
	return nil
}
