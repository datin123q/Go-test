package serviceutils

import "github.com/illuminarean/energy-shares-backend-repo/pkg/nexway"

// 本人確認デジタル認証サービス(JPKI方式の署名)と、本人確認BPOサービス JPKI版の署名検証
// オプション用の設定(MIR-1470)。ProTech ID Checker(ホ方式)とは別契約・別トークン。
//
// BPO JPKI版のベースURL・トークン(SERVICE_NEXWAY_JPKI_BPO_*)は serviceutils.go にある。
var (
	SERVICE_NEXWAY_JPKI_AUTH_BASE_URL string
	SERVICE_NEXWAY_JPKI_AUTH_TOKEN    string
	SERVICE_NEXWAY_JPKI_BPO_SECRET    string
)

// SetNexwayJpkiConfig は、JPKI方式で使う設定のうち SetNexwayConfig が受け持たないものを設定する。
func SetNexwayJpkiConfig(c nexway.Config) {
	SERVICE_NEXWAY_JPKI_AUTH_BASE_URL = c.JpkiAuthBaseUrl
	SERVICE_NEXWAY_JPKI_AUTH_TOKEN = c.JpkiAuthToken
	SERVICE_NEXWAY_JPKI_BPO_SECRET = c.JpkiBpoSigningSecret
}

// GetServiceNexwayJpkiAuthBaseUrl は、本人確認デジタル認証サービスのベースURL(ホスト部まで)。
func GetServiceNexwayJpkiAuthBaseUrl() string {
	return SERVICE_NEXWAY_JPKI_AUTH_BASE_URL
}

// GetServiceNexwayJpkiAuthToken は、本人確認デジタル認証サービスのAPIトークン。
func GetServiceNexwayJpkiAuthToken() string {
	return SERVICE_NEXWAY_JPKI_AUTH_TOKEN
}

// GetServiceNexwayJpkiBpoSigningSecret は、本人確認BPOサービスの署名用シークレットキー。
func GetServiceNexwayJpkiBpoSigningSecret() string {
	return SERVICE_NEXWAY_JPKI_BPO_SECRET
}
