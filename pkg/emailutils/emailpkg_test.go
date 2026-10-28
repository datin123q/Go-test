package emailutils

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

var emailRequester EmailRequester

const (
	// DEV、QA用の統合メールサービス
	// emailServiceUrl = "https://ilm-svc-dev-api-service.illuminarean.com:5000"
	emailServiceUrl = "https://ilm-svc-prod-api-service.illuminarean.com:5000"

	// 既に知られているmessage id
	messageId = "108428"
)

func TestMain(m *testing.M) {
	var cfg config.Config
	cfg.CommonConfig.EmailService = config.EmailServiceConfig{
		Url: emailServiceUrl,
	}

	emailRequester = New(&cfg, factory.NewLoggerFactory())
	m.Run()
}

// TestNew_UsesIndependentClientWithExpectedTimeout はMIR-1327の回帰テスト。
// 以前の実装(restyClient.Clone().SetTimeout(10秒))は、DIコンテナが全機能に
// 共有するresty.New()をCloneしてSetTimeoutしていた。Clone()は内部の
// *http.Clientをポインタのまま引き継ぐため、元の共有インスタンスのタイムアウト
// まで書き換わってしまい、eKYC・egift等、他の機能の打ち切り時間が意図せず
// 変わっていた(実際にpkg/emailutils・pkg/egift・service/nexwayを横断して
// 確認する統合テストはapps/user-api/server/http_client_isolation_test.go
// にある)。ここではNewが受け取る引数を持たず、毎回独立したresty.New()を
// 使うことを確認する。
func TestNew_UsesIndependentClientWithExpectedTimeout(t *testing.T) {
	var cfg config.Config
	cfg.CommonConfig.EmailService = config.EmailServiceConfig{Url: emailServiceUrl}

	a := New(&cfg, factory.NewLoggerFactory()).(*emailpkg)
	b := New(&cfg, factory.NewLoggerFactory()).(*emailpkg)

	assert.NotSame(t, a.restyClient.GetClient(), b.restyClient.GetClient(), "New()を呼ぶたびに独立したhttp.Clientを使うこと")
	assert.Equal(t, 10*time.Second, a.restyClient.GetClient().Timeout)
}

// TestNew_DebugIsDisabled はMIR-1345の回帰テスト。New()が返すクライアントの
// SetDebug(true)を外した。デバッグ出力(~~~ REQUEST ~~~)には宛先・氏名・
// 振込先がそのまま含まれ、本番のログ(CloudWatch、保持30日)に出ていた。
func TestNew_DebugIsDisabled(t *testing.T) {
	var cfg config.Config
	cfg.CommonConfig.EmailService = config.EmailServiceConfig{Url: emailServiceUrl}

	e := New(&cfg, factory.NewLoggerFactory()).(*emailpkg)

	assert.False(t, e.restyClient.Debug, "New()が返すクライアントはSetDebug(true)になっていないこと")
}

// このテストはDEV、QA用の統合メールサービスに直接アクセスしてテストを実行する。
func TestEmailpkg_GetAllRequest(t *testing.T) {
	t.Skip("実際のサービスにアクセスしてテストするため、CI/CDではテストをスキップします。")

	query := GetAllQuery{
		ServiceName: "ESJP",
		ServiceEnv:  "DEV",
		Sender:      lo.ToPtr("info@energyshares.jp"),
	}
	response, err := emailRequester.GetAllRequest(context.Background(), query)
	if err != nil {
		t.Errorf("error sending email request: %v", err)
	}
	log.Printf("response: %+v", response)

}

// このテストはDEV、QA用の統合メールサービスに直接アクセスしてテストを実行する。
func TestEmailpkg_GetByIdRequest(t *testing.T) {
	t.Skip("実際のサービスにアクセスしてテストするため、CI/CDではテストをスキップします。")

	response, err := emailRequester.GetByIdRequest(context.Background(), messageId)
	if err != nil {
		t.Errorf("error sending email request: %v", err)
	}
	log.Println("time.Time: ", time.Time(response.SendDateTime))
	log.Printf("response: %+v", response)
}
