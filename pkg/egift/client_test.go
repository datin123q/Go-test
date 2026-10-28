package egift

import (
	"context"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// Test_New_UsesIndependentClientWithExpectedTimeout はMIR-1327の回帰テスト。
// New()を呼ぶたびに独立したhttp.Clientを使い、修正前に共有クライアント経由で
// 実質的に効いていた値(10秒)を明示的に保つことを確認する。
func Test_New_UsesIndependentClientWithExpectedTimeout(t *testing.T) {
	cfg := &config.Config{CommonConfig: config.CommonConfig{
		AccessTradeService: config.AccessTradeConfig{Url: "https://is.dev-accesstrade.net", Stub: false},
	}}

	a := New(cfg, factory.NewLoggerFactory()).(*egiftClient)
	b := New(cfg, factory.NewLoggerFactory()).(*egiftClient)

	assert.NotSame(t, a.restyClient.GetClient(), b.restyClient.GetClient(), "New()を呼ぶたびに独立したhttp.Clientを使うこと")
	assert.Equal(t, 10*time.Second, a.restyClient.GetClient().Timeout)
}

func Test_egiftClient_RewardSignUp(t *testing.T) {
	t.Run("SignUp Reward를 성공적으로 전달합니다.", func(t *testing.T) {
		restyClient := resty.New()

		marketingCode := "energyshares_0100plz8000005-b16fc880305abe1554e865a82c2d6dee"
		signUpRequestId := "01d1884e-5a90-4160-b160-68fd379172af"

		httpmock.Activate()
		httpmock.ActivateNonDefault(restyClient.GetClient())
		httpmock.RegisterResponder("POST", "https://is.dev-accesstrade.net/cgi-bin/isatV2/energyshares/isatWeaselV2.cgi"+
			"?result_id=100&rk=energyshares_0100plz8000005-b16fc880305abe1554e865a82c2d6dee&verify=01d1884e-5a90-4160-b160-68fd379172af",
			httpmock.NewStringResponder(200, ""))
		defer httpmock.DeactivateAndReset()

		client := egiftClient{
			cfg: &config.Config{
				CommonConfig: config.CommonConfig{
					AccessTradeService: config.AccessTradeConfig{
						Url: "https://is.dev-accesstrade.net",
					},
				},
			},
			restyClient: restyClient,
		}

		assert.NoErrorf(t, client.RewardSignUp(context.Background(), marketingCode, signUpRequestId), "RewardSignUp() error")
	})
}
