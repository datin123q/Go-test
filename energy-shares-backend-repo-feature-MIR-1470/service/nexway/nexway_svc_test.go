package nexway

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

// Test_New_SetsExpectedTimeout はMIR-1327の確認。共有クライアントを持たなく
// なった後も、本来の打ち切り時間(10秒、service.NexwayApiTimeoutと同じ値)が
// 明示的に設定されていることを確認する。
func Test_New_SetsExpectedTimeout(t *testing.T) {
	svc := New(&config.Config{}, factory.NewLoggerFactory()).(*nexwayService)

	assert.Equal(t, 10*time.Second, svc.restyClient.GetClient().Timeout)
}

// Test_apiTimeout_MatchesServiceNexwayApiTimeout はapiTimeout(このパッケージの
// ローカル定数)がservice.NexwayApiTimeout(MIR-1223)からずれていないことを保証
// する。値をそのまま参照できない事情(コメント参照)により複製しているため、
// 片方だけ変更されて分岐することをここで検出する。
func Test_apiTimeout_MatchesServiceNexwayApiTimeout(t *testing.T) {
	assert.Equal(t, service.NexwayApiTimeout, apiTimeout)
}
