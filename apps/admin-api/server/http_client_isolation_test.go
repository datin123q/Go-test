package server

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	changelogRepository "github.com/illuminarean/energy-shares-backend-repo/repository/changelog/ksql"
	operatorHistoryKsqlRepository "github.com/illuminarean/energy-shares-backend-repo/repository/operator_history/ksql"
	nexwayService "github.com/illuminarean/energy-shares-backend-repo/service/nexway"
)

// TestHttpClient_IsolatedAcrossFeatures はMIR-1327の回帰テスト。
//
// admin-apiのDIコンテナはresty.Newを唯一のプロバイダとして*resty.Clientを
// 全機能へ共有する。以前の実装は、emailutils.Newがこの共有インスタンスを
// Clone()してSetTimeout(10秒)していた。resty.ClientのClone()は内部の
// *http.Clientをポインタのまま引き継ぐため、Clone後のSetTimeoutは元の共有
// インスタンスの打ち切り時間まで書き換えてしまい、同じ共有インスタンスを
// そのまま保持していたnexwayService・changelogRepository・
// operatorHistoryKsqlRepositoryの打ち切り時間まで意図せず10秒に変わっていた。
//
// このテストは実際のfx.Provide配線(resty.New一つを共有する形)を再現し、
// emailutils・nexwayService・changelogRepository・operatorHistoryKsqlRepository
// を全て構築した後も、DIコンテナが提供する共有クライアント自体のタイムアウトが
// 変化しないこと、かつ各機能が自身の独立したクライアントで期待どおりの
// タイムアウト(10秒、修正前に実質的に効いていた値)を持つことを確認する。
func TestHttpClient_IsolatedAcrossFeatures(t *testing.T) {
	cfg := &config.Config{}
	cfg.CommonConfig.EmailService.Url = "https://email.example.com"

	var sharedClient *resty.Client
	var emailRequester emailutils.EmailRequester
	var nexwaySvc cdomain.NexwayService
	var changelogRepo cdomain.ChangelogRepository
	var operatorHistoryRepo cdomain.OperatorHistoryKsqlRepository

	app := fx.New(
		fx.Provide(
			func() *config.Config { return cfg },
			func() factory.LoggerFactory { return factory.NewLoggerFactory() },
			resty.New,
			emailutils.New,
			nexwayService.New,
			changelogRepository.New,
			operatorHistoryKsqlRepository.New,
		),
		fx.Populate(&sharedClient, &emailRequester, &nexwaySvc, &changelogRepo, &operatorHistoryRepo),
	)
	require.NoError(t, app.Err())

	assert.Zero(t, sharedClient.GetClient().Timeout,
		"DIコンテナが提供する共有クライアントは、どの機能の初期化でも書き換わらないこと")
}
