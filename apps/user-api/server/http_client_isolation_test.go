package server

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/egift"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	nexwayService "github.com/illuminarean/energy-shares-backend-repo/service/nexway"
)

// TestHttpClient_IsolatedAcrossFeatures はMIR-1327の回帰テスト。
//
// admin-api/user-apiのDIコンテナはresty.Newを唯一のプロバイダとして*resty.Client
// を全機能へ共有する。以前の実装は、emailutils.Newがこの共有インスタンスを
// Clone()してSetTimeout(10秒)していた。resty.ClientのClone()は内部の
// *http.Clientをポインタのまま引き継ぐため、Clone後のSetTimeoutは元の共有
// インスタンスの打ち切り時間まで書き換えてしまい、同じ共有インスタンスをそのまま
// 保持していたegift・nexwayServiceの打ち切り時間まで意図せず10秒に変わっていた。
//
// このテストが確認するのはDIコンテナが提供する共有クライアント自体のタイムアウトが
// どの機能の初期化でも変化しないことだけ。各機能(emailutils・egift・
// nexwayService)が独立クライアントで期待どおりのタイムアウトを持つことは、
// それぞれのパッケージのユニットテスト(emailpkg_test.go・client_test.go・
// nexway_svc_test.go)で確認している。egift・nexwayServiceの型は非公開で
// このパッケージの外からは具体型にキャストできないため、ここでは検証できない。
func TestHttpClient_IsolatedAcrossFeatures(t *testing.T) {
	cfg := &config.Config{}
	cfg.CommonConfig.EmailService.Url = "https://email.example.com"
	cfg.CommonConfig.AccessTradeService.Url = "https://egift.example.com"
	cfg.CommonConfig.AccessTradeService.Stub = false

	var sharedClient *resty.Client
	var emailRequester emailutils.EmailRequester
	var egiftClient egift.Egift
	var nexwaySvc cdomain.NexwayService

	app := fx.New(
		fx.Provide(
			func() *config.Config { return cfg },
			func() factory.LoggerFactory { return factory.NewLoggerFactory() },
			resty.New,
			emailutils.New,
			egift.New,
			nexwayService.New,
		),
		fx.Populate(&sharedClient, &emailRequester, &egiftClient, &nexwaySvc),
	)
	require.NoError(t, app.Err())

	assert.Zero(t, sharedClient.GetClient().Timeout,
		"DIコンテナが提供する共有クライアントは、どの機能の初期化でも書き換わらないこと")
}
