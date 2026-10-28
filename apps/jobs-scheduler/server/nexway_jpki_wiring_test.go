package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/nexway"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

// JPKI方式の設定が起動時にserviceutilsへ渡ること。渡し忘れると、Nexwayへの通信が
// 本番でだけ「未設定」で止まり、ユニットテストでは気付けない(MIR-1470)。
func TestInitServiceUtils_SetsNexwayJpkiConfig(t *testing.T) {
	// 秘密鍵の取得(AWS)を飛ばすため、設定済みにしておく。
	originalKey := serviceutils.SERVICE_PRIVATE_KEY
	serviceutils.SERVICE_PRIVATE_KEY = "test-key"
	t.Cleanup(func() { serviceutils.SERVICE_PRIVATE_KEY = originalKey })
	original := nexway.Config{
		JpkiAuthBaseUrl:      serviceutils.GetServiceNexwayJpkiAuthBaseUrl(),
		JpkiAuthToken:        serviceutils.GetServiceNexwayJpkiAuthToken(),
		JpkiBpoBaseUrl:       serviceutils.GetServiceNexwayJpkiBpoBaseUrl(),
		JpkiBpoToken:         serviceutils.GetServiceNexwayJpkiBpoToken(),
		JpkiBpoSigningSecret: serviceutils.GetServiceNexwayJpkiBpoSigningSecret(),
	}
	t.Cleanup(func() {
		serviceutils.SetNexwayConfig(original)
		serviceutils.SetNexwayJpkiConfig(original)
	})

	var c config.Config
	c.Env = "local"
	c.Nexway = nexway.Config{
		JpkiAuthBaseUrl:      "https://auth.example.jp",
		JpkiAuthToken:        "auth-token",
		JpkiBpoBaseUrl:       "https://bpo.example.jp",
		JpkiBpoToken:         "bpo-token",
		JpkiBpoSigningSecret: "secret-secret-secret-secret-secret",
	}

	assert.NoError(t, initServiceUtils(c))

	assert.Equal(t, "https://auth.example.jp", serviceutils.GetServiceNexwayJpkiAuthBaseUrl())
	assert.Equal(t, "auth-token", serviceutils.GetServiceNexwayJpkiAuthToken())
	assert.Equal(t, "https://bpo.example.jp", serviceutils.GetServiceNexwayJpkiBpoBaseUrl())
	assert.Equal(t, "bpo-token", serviceutils.GetServiceNexwayJpkiBpoToken())
	assert.Equal(t, "secret-secret-secret-secret-secret", serviceutils.GetServiceNexwayJpkiBpoSigningSecret())
}

func TestNewRouter_RegistersNexwayWebhooks(t *testing.T) {
	routes := map[string]bool{}
	for _, r := range NewRouter(false).Routes() {
		routes[r.Method+" "+r.Path] = true
	}

	// カ方式の署名検証結果(MIR-1470)が増えても、ホ方式のBPO結果通知の受信口は変わらない。
	assert.True(t, routes[http.MethodPost+" /nexway/jpki/webhook"], "署名検証結果のWebhook")
	assert.True(t, routes[http.MethodPost+" /nexway/bpo/webhook"], "BPO結果のWebhook(ホ方式・JPKI方式共通)")
}

// BPOの結果通知(JPKI)の署名検証が、実際のルートに付いていること。ミドルウェア単体のテストが
// 通っていても、ルートへ付け忘れれば検証なしで受けてしまう(MIR-1470)。
func TestNewRouter_BpoWebhookRequiresSignatureForJpki(t *testing.T) {
	original := serviceutils.GetServiceNexwayJpkiBpoSigningSecret()
	originalCfg := nexway.Config{
		JpkiAuthBaseUrl: serviceutils.GetServiceNexwayJpkiAuthBaseUrl(),
		JpkiAuthToken:   serviceutils.GetServiceNexwayJpkiAuthToken(),
	}
	t.Cleanup(func() {
		originalCfg.JpkiBpoSigningSecret = original
		serviceutils.SetNexwayJpkiConfig(originalCfg)
	})
	serviceutils.SetNexwayJpkiConfig(nexway.Config{JpkiBpoSigningSecret: "9ab6bf804c3937fcf27ddaad49721c6b"})

	req := httptest.NewRequest(http.MethodPost, "/nexway/bpo/webhook",
		strings.NewReader(`{"applicant_id":13,"reference_id":"ref-1","ekyc_type":"jpki","status":"completed","result":"match"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	NewRouter(false).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code, "署名の無いJPKIの通知は後続の処理へ進まない")
}
