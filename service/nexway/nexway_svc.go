package nexway

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// apiTimeoutはservice.NexwayApiTimeout(MIR-1223)と同じ10秒。service パッケージは
// domain 等の重い依存を持つため、値をそのまま参照せずここで揃えている。
const apiTimeout = 10 * time.Second

type nexwayService struct {
	cfg         *config.Config
	restyClient *resty.Client
	logger      factory.LoggerFactory
}

// MIR-1327: 以前は引数のrestyClient(DIコンテナが全機能に共有するresty.New())を
// そのまま(Cloneもせず)保持していたため、他の機能(email等)がClone().SetTimeout()
// すると、Clone()が内部の*http.Clientをポインタのまま引き継ぐ仕様により、この
// クライアントの打ち切り時間まで意図せず変わっていた。apps/user-api/usecase/
// private_user/private_user_ucase.goがこの実装をフィールドとして保持している
// (現時点でメソッド呼び出しは無い。実際にeKYC照会に使われているのは別実装の
// service.NewNexwayService()、apps/jobs-scheduler・apps/admin-apiのprivate_user.go
// 経由)。呼び出しの有無によらず、New呼び出し自体でSetTimeoutが実行され共有
// クライアントに影響するため、独立クライアントへの修正は必要。resty.New()で
// 独立したクライアントを持ち、本来の打ち切り時間(10秒)を明示的に設定する
// (河内さん判断、2026-09-29)。
func New(
	cfg *config.Config,
	logger factory.LoggerFactory,
) cdomain.NexwayService {
	return &nexwayService{
		cfg:         cfg,
		restyClient: resty.New().SetTimeout(apiTimeout),
		logger:      logger,
	}
}

// GetEkycResultByApplicantId eKYC結果照会
func (n nexwayService) GetEkycResultByApplicantId(ctx context.Context, applicantId string) (cdomain.NexwayEkycResultOutput, error) {
	var ekycResultOutput cdomain.NexwayEkycResultOutput
	serviceUrl := n.cfg.Nexway.ApplicantsApi + "/" + applicantId + "/include_ekyc"

	_, err := n.restyClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Token "+n.cfg.Nexway.Token).
		SetResult(&ekycResultOutput).
		Get(serviceUrl)
	if err != nil {
		return cdomain.NexwayEkycResultOutput{}, errors.Wrap(err, "failed to call nexway api")
	}

	return ekycResultOutput, nil
}
