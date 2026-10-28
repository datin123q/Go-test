package nexway

type Config struct {
	Id                      string
	Password                string
	FormRegistrationApi     string
	FormFileRegistrationApi string
	FormOrderApi            string
	ApplicantsApi           string
	Token                   string

	// JpkiAuthBaseUrl・JpkiAuthTokenは、本人確認デジタル認証サービス(JPKI方式の署名用URLの
	// 発行と署名結果の通知)用。ProTech ID Checker(ホ方式、上のApplicantsApi・Token)とは
	// 別契約で、URLもトークンも別に発行される。JpkiAuthBaseUrlはホスト部まで
	// (例: https://sandbox.jpki.nexlink2.jp)で、パスは呼び出し側で付ける。
	JpkiAuthBaseUrl string
	JpkiAuthToken   string

	// JpkiBpoBaseUrl・JpkiBpoTokenは、本人確認BPOサービス JPKI版(ekyc_type=jpki)用。
	// ProTech ID Checker(ホ方式、上のApplicantsApi・Token)とは別契約で、URLもトークンも
	// 別に発行される。JpkiBpoBaseUrlはホスト部まで(例: https://example.jp)で、パスは
	// 呼び出し側で付ける(照会はv1、申請はv2とバージョンが異なるため)。
	JpkiBpoBaseUrl string
	JpkiBpoToken   string

	// JpkiBpoSigningSecretは、本人確認BPOサービスの署名検証オプション(HMAC-SHA256)で使う
	// 署名用のシークレットキー(英数字30〜50文字程度)。APIトークンとは別に発行される。
	// 未設定のときは、BPOへの申請を通信前にエラーにする(署名なしでは送らない)。
	JpkiBpoSigningSecret string
}
