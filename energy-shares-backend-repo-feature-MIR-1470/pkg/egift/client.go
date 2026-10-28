package egift

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
)

const (
	_egiftTargetId = "100"

	// apiTimeoutはMIR-1327の修正前、DIコンテナが全機能に共有するresty.New()を
	// そのまま(Cloneもせず)保持していたことで実質的に効いていた値(emailutils.New
	// がClone().SetTimeoutしていた10秒)と同じ。修正後も打ち切り時間を変えない
	// ため、同じ値を明示的に設定する(河内さん判断、2026-09-29)。
	apiTimeout = 10 * time.Second
)

type egiftClient struct {
	cfg         *config.Config
	restyClient *resty.Client
	logger      factory.LoggerFactory
}

type Egift interface {
	RewardSignUp(ctx context.Context, marketingCode string, signUpRequestId string) error
}

// MIR-1327: 以前はDIコンテナが全機能に共有するresty.New()をそのまま(Cloneも
// せず)保持していたため、他の機能(email等)がClone().SetTimeout()すると、
// Clone()が内部の*http.Clientをポインタのまま引き継ぐ仕様により、この
// クライアントの打ち切り時間まで意図せず変わっていた。resty.New()で独立した
// クライアントを持ち、修正前に実際に効いていた値(10秒)を明示的に設定する。
func New(cfg *config.Config, logger factory.LoggerFactory) Egift {
	// E-Giftサービスはサンドボックスを運営していないため、本番環境でのみ機能します。
	// スタブの使用にさらに柔軟性を持たせたい場合は、環境を確認するのではなく、別のフラグで設定することも検討できます。
	if cfg != nil && cfg.AccessTradeService.Stub == false {
		return &egiftClient{
			cfg:         cfg,
			restyClient: resty.New().SetTimeout(apiTimeout),
			logger:      logger,
		}
	}

	// Dev/Qa/Stage環境ではスタブを使用します。
	return &egiftClientStub{
		logger: logger,
	}
}

func (e egiftClient) RewardSignUp(ctx context.Context, marketingCode string, signUpRequestId string) error {
	apiUrl := fmt.Sprintf("%s/cgi-bin/isatV2/energyshares/isatWeaselV2.cgi", e.cfg.AccessTradeService.Url)

	_, err := e.restyClient.R().
		SetContext(ctx).
		SetQueryParam("result_id", _egiftTargetId).
		SetQueryParam("verify", signUpRequestId).
		SetQueryParam("rk", marketingCode).
		SetHeader("Content-Type", "text/plain").
		Post(apiUrl)

	if err != nil {
		return errors.Wrapf(err, "failed to signup reward post request : %v", apiUrl)
	}

	return nil
}
