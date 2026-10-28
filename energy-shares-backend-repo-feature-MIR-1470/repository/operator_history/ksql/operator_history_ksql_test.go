package ksql

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// Test_New_UsesIndependentClientWithExpectedTimeout はMIR-1327の回帰テスト。
// New()を呼ぶたびに独立したhttp.Clientを使い、修正前に共有クライアント経由で
// 実質的に効いていた値(10秒)を明示的に保つことを確認する。
func Test_New_UsesIndependentClientWithExpectedTimeout(t *testing.T) {
	cfg := &config.Config{}
	logger := factory.NewLoggerFactory()

	a := New(cfg, logger).(*operatorHistoryKsqlRepository)
	b := New(cfg, logger).(*operatorHistoryKsqlRepository)

	assert.NotSame(t, a.restyClient.GetClient(), b.restyClient.GetClient(), "New()を呼ぶたびに独立したhttp.Clientを使うこと")
	assert.Equal(t, 10*time.Second, a.restyClient.GetClient().Timeout)
}

// Test_RunKsqlQuery_DebugStaysOffRegardlessOfConfigはMIR-1365の回帰テスト。以前は
// runKsqlQueryがSetDebug(r.cfg.Debug)で設定値を見ており、cfg.Debugがtrueに
// なるとksqlの問い合わせ内容(個人情報を含みうる)がログに出ていた。New()は
// クライアントのDebugを設定しないため、実際に送るリクエストのDebugを左右するのは
// 呼び出し後のSetDebugの有無だけであり、それをOnBeforeRequestフックで確認する。
func Test_RunKsqlQuery_DebugStaysOffRegardlessOfConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"header":{"queryId":"q","schema":"s"}},{"finalMessage":"Limit Reached"}]`))
	}))
	defer server.Close()

	cfg := &config.Config{}
	cfg.Debug = true
	logger := factory.NewLoggerFactory()

	r := New(cfg, logger).(*operatorHistoryKsqlRepository)

	var gotDebug bool
	r.restyClient.OnBeforeRequest(func(_ *resty.Client, req *resty.Request) error {
		gotDebug = req.Debug
		return nil
	})

	_, err := r.runKsqlQuery(context.Background(), server.URL, "SELECT * FROM x;")

	assert.NoError(t, err)
	assert.False(t, gotDebug, "cfg.Debug=trueでも実際に送るリクエストのDebugはfalseのままであるべき")
}
