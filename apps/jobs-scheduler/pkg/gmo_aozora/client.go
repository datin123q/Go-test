package gmo_aozora

import (
	"context"
	"fmt"
	"net/http"
	neturl "net/url"
	"sync"
	"time"

	"github.com/antihax/optional"
	ganblib "github.com/gmoaozora/gmo-aozora-api-go"
	corporateclient "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	webhookclient "github.com/gmoaozora/gmo-aozora-api-go/webhookclient"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/unknwon/com"
	"golang.org/x/time/rate"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

const (
	ORDER_ASC  = "1"
	ORDER_DESC = "2"
)

type SUBSCRIPTION_STATUS string

const (
	STOP_SUBSCRIPTION  SUBSCRIPTION_STATUS = "0" // 구독 종료
	START_SUBSCRIPTION SUBSCRIPTION_STATUS = "1" // 구독 시작
)

var corporateClientCfg *corporateclient.Configuration
var webHookClientCfg *webhookclient.Configuration

var (
	// GMO API 呼び出しで共有するクライアントを初期化するための制御
	corporateClientOnce    sync.Once
	corporateClientInitErr error
	corporateClient        *corporateclient.APIClient

	// GetAccountInfoのテストが実APIを呼ばずに差し替えられるようにするための間接呼び出し。
	getAccountInfoClient = GetCorporateClient
	getAccounts          = func(ctx context.Context, client *corporateclient.APIClient) (corporateclient.AccountsResponse, *http.Response, error) {
		return client.AccountApi.AccountsUsingGET(ctx, GetAccessToken())
	}

	// VerifyAccountsIncludeConfiguredのテストが実APIを呼ばずに差し替えられるようにする
	// ための間接呼び出し。getAccountsと違い、呼び出し元が渡すaccessTokenを使う
	// (保存が確定する前のトークンを検証するため、グローバルなclientInfoは使えない)。
	getAccountsWithToken = func(ctx context.Context, client *corporateclient.APIClient, accessToken string) (corporateclient.AccountsResponse, *http.Response, error) {
		return client.AccountApi.AccountsUsingGET(ctx, accessToken)
	}

	// RequestAozoraNewTokenのテストが実際に銀行へ認可コードを送らずに差し替え
	// られるようにするための間接呼び出し(②レビュー2周目の指摘: 保存経路が
	// withTokenRefreshLockを通ることにテストが無かった)。
	openIDGetToken = func(ganb ganblib.Ganb, redirectUri, code, authMethod string) (ganblib.Token, error) {
		return ganb.OpenIDGetToken(redirectUri, code, authMethod)
	}

	// 3秒に1回リクエストに制限するためのレートリミッタ
	// NOTE: 1秒に1回のリクエスト制限では、エラーになるケースがあるため、余裕を持たせて3秒に1回に設定しています
	liveRequestLimiter   = rate.NewLimiter(rate.Every(3*time.Second), 1)
	rateLimitTimeout     = 30 * time.Second
	liveLimiterPrimeOnce sync.Once
)

type rateLimitedTransport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.limiter != nil {
		// リクエストを送る前にレートリミットを消費して待機する
		if err := t.limiter.Wait(req.Context()); err != nil {
			return nil, err
		}
	}

	transport := t.base
	if transport == nil {
		transport = http.DefaultTransport
	}
	return transport.RoundTrip(req)
}

func newRateLimitedHTTPClient() *http.Client {
	primeLiveRequestLimiter()
	// GMO API クライアント全体で同一の HTTP クライアントを共有し、トラフィックを制御する
	return &http.Client{
		Transport: &rateLimitedTransport{
			base:    http.DefaultTransport,
			limiter: liveRequestLimiter,
		},
		Timeout: rateLimitTimeout,
	}
}

func primeLiveRequestLimiter() {
	liveLimiterPrimeOnce.Do(func() {
		// 起動直後にバースト用トークンを消費し、初回リクエストにも1秒インターバルを適用する
		_ = liveRequestLimiter.Wait(context.Background())
	})
}

var Env string

// RequestAozoraNewTokenの失敗理由を呼び出し元(コールバックのハンドラ)が
// errors.Isで判別できるようにするためのセンチネル(MIR-1298)。画面に出す
// 文言をどのカテゴリにするかの判定に使う。errors.Wrap(sentinel, err.Error())
// の形で包む(sentinelをcauseとして残し、実際のエラー文言をメッセージ側に
// 乗せる。通常のWrapの使い方(err側が原因、messageが説明)とは引数の
// 意味が逆になる点に注意)。
var (
	ErrTokenExchangeFailed = errors.New("gmo aozora: token exchange failed")
	ErrTokenInvalid        = errors.New("gmo aozora: token invalid")
	ErrAccountVerifyFailed = errors.New("gmo aozora: account verify failed")
	ErrTokenSaveFailed     = errors.New("gmo aozora: token save failed")
)

// HTTPStatusCode は *http.Response からステータスコードだけを取り出す。
// *http.Response をそのままログに渡すと Request.GetBody が原因で JSON marshal に失敗し、
// ログ行全体が破棄されるため、ステータスコードのみを抜き出して使う。
func HTTPStatusCode(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

type ClientInfo struct {
	ClientId     string
	ClientSecret string
	Scope        string
	RedirectUri  string
	AccessToken  string
	RefreshToken string
}

// clientInfoMuはclientInfoを保護する(MIR-1298)。HTTPハンドラ(GetLogin等)と
// cron(RefreshToken)が同時にclientInfoを読み書きしうるため。読み出しは
// 必ずこのファイルのgetter(GetAccessToken等)経由にそろえ、直接clientInfo.Xと
// 書かない——このファイル自身の中でも同様(Authorization・RequestAozoraNewToken等)。
// 唯一の例外はLoadClient(下記)で、サーバー起動時にHTTPハンドラ・cronの
// どちらもまだ動いていない一度きりの初期化なので、ロックを取らずに直接
// 代入する。
var (
	clientInfoMu sync.RWMutex
	clientInfo   ClientInfo
)

func GetAccessToken() string {
	clientInfoMu.RLock()
	defer clientInfoMu.RUnlock()
	return clientInfo.AccessToken
}

func GetRefreshTokenValue() string {
	clientInfoMu.RLock()
	defer clientInfoMu.RUnlock()
	return clientInfo.RefreshToken
}

func ClientId() string {
	clientInfoMu.RLock()
	defer clientInfoMu.RUnlock()
	return clientInfo.ClientId
}

func ClientSecret() string {
	clientInfoMu.RLock()
	defer clientInfoMu.RUnlock()
	return clientInfo.ClientSecret
}

func GetScope() string {
	clientInfoMu.RLock()
	defer clientInfoMu.RUnlock()
	return clientInfo.Scope
}

func GetRedirectUri() string {
	clientInfoMu.RLock()
	defer clientInfoMu.RUnlock()
	return clientInfo.RedirectUri
}

// setClientInfo はトークン(access/refresh)だけを書き換える。ClientId・
// ClientSecret・Scope・RedirectUriはLoadClient以降変わらないため触らない。
func setClientInfo(accessToken, refreshToken string) {
	clientInfoMu.Lock()
	defer clientInfoMu.Unlock()
	clientInfo.AccessToken = accessToken
	clientInfo.RefreshToken = refreshToken
}

func LoadClient(db xorm.Interface, env string, cfg config.Config) error {
	Env = env
	Client, err := domain.GmoAozoraClient{}.InitClient(db, cfg)

	if err != nil {
		return errors.Wrap(err, "fail init client")
	}
	if Client == nil {
		return errors.New("not found gmo client")
	}

	clientInfo = ClientInfo{
		ClientId:     Client.ClientId,
		ClientSecret: Client.ClientSecret,
		Scope:        Client.Scope,
		RedirectUri:  Client.RedirectUri,
		AccessToken:  Client.AccessToken,
		RefreshToken: Client.RefreshToken,
	}

	corporateClientCfg = NewClientConfigure(env)
	corporateClientCfg.HTTPClient = newRateLimitedHTTPClient()

	webHookClientCfg = NewWebHookConfigure(env)

	logrus.WithFields(logrus.Fields{
		"env":               env,
		"CorporateBasePath": corporateClientCfg.BasePath,
		"WebHookBasePath":   webHookClientCfg.BasePath,
	}).Info("LoadClient")

	err = GetAccountInfo(context.Background(), BankAccounts)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("GetAccountInfo")
	}

	return nil
}
// nonceCheckAndDelete は ganblib.New の checkAndDelete 引数として渡す既定値。
// 実際のnonce照合はRequestAozoraNewTokenがgmo_aozora_oauth_stateの保存値に
// 対して行う専用クロージャ(checkNonce)で行うため、この関数が呼ばれる経路は
// 無い(Authorization/RefreshTokenはOpenIDGetTokenを呼ばず、isTokenValidの
// nonce照合を経由しないため)。ganblib.Newが引数を必須とするための placeholder。
func nonceCheckAndDelete(nonce string) error {
	return nil
}

// nonceSave は ganblib.New の save 引数。OpenIDAuthorization(Authorization/
// GetLogin経由)がここへnonceを渡すが、実際に照合へ使うnonceはAuthorizationが
// URLをパースして取り出し、呼び出し元がgmo_aozora_oauth_stateへ保存する
// (このファイルのAuthorizationResult参照)。nonceの値自体はログに出さない
// (MIR-1298、code・state・URLと同じ方針)。
func nonceSave(nonce string) error {
	return nil
}

// GetAccountInfoはMIR-1271で、口座一覧取得(AccountsUsingGET)の1回だけにした。
// 以前は口座数ぶん残高取得(BalancesUsingGET)・入金明細取得
// (AccountsDepositTransactionsUsingGET)を繰り返しており、GMO側の3秒間隔の
// レート制限と相まって1+2N回のリクエストが必要だった。残高・入金明細は
// ログ出力のみで業務処理には使っておらず(入金明細の保存・消込はバッチ・
// Webhook側の別経路が担う)、削除しても挙動は変わらない。
//
// あわせて、Configに設定した口座がGMOの応答に含まれない場合にエラーを
// 返す検証を追加した。以前はInitAccountInfoが対象外の口座を黙って無視して
// いたため、口座設定のズレ(例: 口座番号の入力間違い)があっても気付けなかった。
func GetAccountInfo(ctx context.Context, accounts Accounts) error {
	configuredAccountCount := 0
	for _, account := range accounts {
		if account != nil && account.AccountNumber != "" {
			configuredAccountCount++
		}
	}
	if configuredAccountCount == 0 {
		return errors.New("empty configured bank accounts")
	}

	client, err := getAccountInfoClient()
	if err != nil {
		return errors.Wrap(err, "fail load corporate client")
	}

	res, httpResp, err := CallWith429Retry(ctx, "AccountApi.AccountsUsingGET", func(ctx context.Context) (corporateclient.AccountsResponse, *http.Response, error) {
		return getAccounts(ctx, client)
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"httpStatusCode": HTTPStatusCode(httpResp),
			"error":          err,
			"logTime":        time.Now(),
		}).Error("Fail Account Get")
		return errors.Wrap(err, "fail account get")
	}

	initializedAccountNumbers := make(map[string]struct{}, len(res.Accounts))
	for _, apiAccount := range res.Accounts {
		accountNumber := apiAccount.AccountNumber
		if accountNumber == "" {
			continue
		}
		if account := accounts.GetByAccountNumber(accountNumber); account != nil && apiAccount.AccountId != "" {
			accounts.InitAccountInfo(accountNumber, apiAccount.AccountId, apiAccount.AccountTypeCode, apiAccount.TransferLimitAmount)
			initializedAccountNumbers[accountNumber] = struct{}{}

			// 口座番号(accountNumber)は出さない。診断にはaccountUseType
			// (Configのどの枠か)で足り、番号自体はParameter Storeで
			// 確認できる(CLAUDE.mdのReview Checklist「口座番号もログに
			// 出さない」)。
			logrus.WithFields(logrus.Fields{
				"accountUseType": account.AccountUseType,
				"accountId":      apiAccount.AccountId,
				"accountType":    apiAccount.AccountTypeCode,
				"accountLimit":   apiAccount.TransferLimitAmount,
				"logTime":        time.Now(),
			}).Info("success init configured account info")
		}
	}

	missingAccounts := make([]string, 0)
	for _, account := range accounts {
		if account == nil || account.AccountNumber == "" {
			continue
		}
		if _, ok := initializedAccountNumbers[account.AccountNumber]; ok {
			continue
		}
		missingAccounts = append(missingAccounts, string(account.AccountUseType))
	}
	if len(missingAccounts) > 0 {
		return errors.Errorf("configured bank accounts not found in GMO response: %v", missingAccounts)
	}

	return nil
}

// AuthorizationResult は認可URLに加え、ライブラリが実際にURLへ埋め込んだ
// state(H)とnonceを保持する。sessionIdそのものではなくこれらをDBへ書き戻す
// 必要がある(MIR-1298)。ライブラリ(gmo-aozora-api-go)はH = sha256(sessionId + salt)
// を計算してURLのstateパラメータへ載せるため、コールバックで受け取るstateは
// sessionIdではなくHになる。ライブラリのnonceSaveコールバック(このファイルの
// nonceSave、現状ログを出すだけ)には頼らず、返ってきたURLをパースして直接
// 取り出す。
type AuthorizationResult struct {
	URL   string
	State string
	Nonce string
}

func Authorization(sessionId string) (*AuthorizationResult, error) {
	ganb, err := ganblib.New(ClientId(), ClientSecret(), nonceSave, nonceCheckAndDelete)
	if err != nil {
		return nil, errors.Wrap(err, "fail create new ganb ")
	}

	rawURL, err := ganb.OpenIDAuthorization(sessionId, GetScope(), GetRedirectUri())
	if err != nil {
		// sessionId・scope・redirectUriはstate計算の入力や機微情報を含みうるため
		// ログに出さない(MIR-1298)。
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"logTime": time.Now(),
		}).Error("Fail OpenId Authorization")
		return nil, err
	}

	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return nil, errors.Wrap(err, "fail parse authorization url")
	}
	state := parsed.Query().Get("state")
	nonce := parsed.Query().Get("nonce")
	if state == "" || nonce == "" {
		return nil, errors.New("authorization url missing state or nonce")
	}

	// URL自体(state・nonce・client_idを含む)はログに出さない(MIR-1298)。
	logrus.WithFields(logrus.Fields{
		"logTime": time.Now(),
	}).Info("Success OpenId Authorization")

	return &AuthorizationResult{URL: rawURL, State: state, Nonce: nonce}, nil
}

// RequestAozoraNewToken は認可コードをトークンへ交換し、検証のうえ保存する
// (MIR-1298)。expectedNonceは、GetLoginで発行したURLに埋め込まれ
// gmo_aozora_oauth_stateへ書き戻したnonce。ライブラリのisTokenValidが
// IDTokenのnonceクレームと照合する。
//
// 交換に成功した時点で銀行の認可コードは使用済みになり、取り消せない。
// 以降(妥当性確認・口座確認・DB保存)はDetachedContextで動かし、担当者が
// ポップアップを閉じる・リクエストがタイムアウトする等で親ctxがキャンセル
// されても、取得済みのトークンを保存する処理を最後まで走らせる。
func RequestAozoraNewToken(ctx context.Context, code, expectedNonce string) (*ganblib.Token, error) {
	if code == "" {
		return nil, errors.New("gmo aozora auth code is empty")
	}

	checkNonce := func(gotNonce string) error {
		if expectedNonce == "" || gotNonce != expectedNonce {
			return errors.New("nonce mismatch")
		}
		return nil
	}

	ganb, err := ganblib.New(ClientId(), ClientSecret(), nonceSave, checkNonce)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"logTime": time.Now(),
		}).Error("Fail Create ganb")
		return nil, err
	}

	token, err := openIDGetToken(ganb, GetRedirectUri(), code, "basic")
	if err != nil {
		// codeはログに出さない(MIR-1298)。
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"logTime": time.Now(),
		}).Error("Fail Request Aozora Token")
		return nil, errors.Wrap(ErrTokenExchangeFailed, err.Error())
	}

	workCtx := DetachedContext(ctx)

	if err := validateToken(token); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"logTime": time.Now(),
		}).Error("Fail Request Aozora Token Validate")
		return nil, errors.Wrap(ErrTokenInvalid, err.Error())
	}

	if err := VerifyAccountsIncludeConfigured(workCtx, token.AccessToken, BankAccounts); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"logTime": time.Now(),
		}).Error("Fail Request Aozora Token Account Verify")
		return nil, errors.Wrap(ErrAccountVerifyFailed, err.Error())
	}

	// ログイン経路はDB保存が確定するまでメモリを変えない。失敗したら、この
	// ログイン試行は無かったことにして(既存のトークンのまま)、担当者に
	// 管理画面からやり直してもらう。
	//
	// 確定設計(MIR-1298)の指示どおり、保存は更新ジョブ(RefreshToken)と同じ
	// gmo_aozora_token_refreshロックで囲む——ロックが無いと、cronが銀行を
	// 呼んでいる最中にログインの保存が先に終わった場合、cronが古い系列で
	// 上書きしてしまう(②レビューで見つかった抜け)。
	// ②レビュー2周目の指摘(MIR-1298): setClientInfoもロックの中で行う。
	// ロックの外だと、RELEASE直後にcronがこの行より先に書き込んだ場合、
	// 使用済みになったトークンがメモリに残るおそれがある(cronは銀行を
	// 呼んでから書き込むため起こりにくいが、ロックの中に置けば確実に防げる)。
	saveErr := withTokenRefreshLock(workCtx, func(ctx context.Context) error {
		if err := updateStoredToken(ctx, token); err != nil {
			return err
		}
		setClientInfo(token.AccessToken, token.RefreshToken)
		return nil
	})
	if saveErr != nil {
		logrus.WithFields(logrus.Fields{
			"error":   saveErr,
			"logTime": time.Now(),
		}).Error("Fail Client Model Update ")
		return nil, errors.Wrap(ErrTokenSaveFailed, saveErr.Error())
	}

	logrus.WithFields(logrus.Fields{
		"expiresIn": token.ExpiresIn,
		"tokenType": token.TokenType,
		"logTime":   time.Now(),
	}).Info("Success Request Aozora Token")

	// 購読開始は保存が確定した直後。有効なstateを持つ要求しか通らないため、
	// 外から起こせる形にはならない。
	if err := WebhookSubscription(START_SUBSCRIPTION); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Fail Aozora WebhookSubscription")
	}

	if err := GetAccountInfo(workCtx, BankAccounts); err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("fail refresh account info after new token")
	}

	return &token, nil
}

// RefreshToken は、保存済みのrefresh_tokenで新しいトークンを取得し保存する
// (毎日のcron・手動実行の両方から呼ばれる、MIR-1298)。デプロイ中は新旧2つの
// ECSタスクが並行するため、gmo_aozora_token_refreshの名前付きロックで
// 「DB読み直し→銀行呼び出し→DB書き込み」の区間全体を排他する。
//
// ②レビューの指摘(MIR-1298): 手動実行(POST /refresh-token-job、本番では
// 無効)はcontextDBWithNameがトランザクションを張るため、RELEASE_LOCKの
// あとハンドラの応答時にコミットされる。ロック解放とコミットの間にごく
// わずかな隙間ができ、その間に他タスクがロックを取ると未コミットの古い
// 値を読む恐れがあるため、DetachedContextでSessionをEngineへ差し替え、
// このジョブの中のDB書き込みを即時コミットにそろえる(cronは元々
// Begin()しないため、この差し替えは実質ログイン経路との対称性のためだけ)。
func RefreshToken(ctx context.Context) error {
	return withTokenRefreshLock(DetachedContext(ctx), refreshTokenLocked)
}

func refreshTokenLocked(ctx context.Context) error {
	// デプロイ中の古いタスクが、他タスクが既に更新した新しい系列を古い値で
	// 上書きしないよう、銀行を呼ぶ直前にDBから最新のrefresh_tokenを読み直す。
	client, err := domain.GmoAozoraClient{}.GetByID(ctx, 1)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":     err,
			"logTime": time.Now(),
		}).Error("Fail Client Model Search ")
		return err
	}
	if client == nil {
		return errors.New("gmo aozora client row not found")
	}

	// ②レビューの指摘(MIR-1298): 「DBの値がメモリと食い違えば他タスク更新済み
	// とみなしてスキップする」という最適化を①の対応で一時追加したが、これは
	// 「自分自身のDB保存が失敗した」場合と区別できず、かえって重大な退行を
	// 生んだ(保存に失敗した翌日、メモリに残る新しいトークンをDBの古い値で
	// 上書きして捨ててしまう)。削除し、確定設計どおり常にDBを読み直して
	// 銀行を呼ぶ単純な実装に戻す。デプロイ中に2タスクが順にロックを取ると
	// 銀行を2回呼ぶことになるが、これは確定設計が「銀行が旧トークンを
	// 無効にするかは未確認」として許容している範囲であり、実装側で
	// 対処しようとして今回のような新しい不具合を生むより、設計どおりの
	// 単純な実装を保つ方を優先する。
	newToken, err := refreshTokenWithRetryOnRejection(ctx, client.RefreshToken)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"logTime": time.Now(),
		}).Error("Fail Request Aozora Refresh Token")
		return err
	}

	if err := validateToken(newToken); err != nil {
		// 空のトークンで上書きしない。ライブラリのRefreshTokensはHTTPの状態
		// コードを見ないため、銀行がエラーJSONを返しても解析自体は成功しうる。
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"logTime": time.Now(),
		}).Error("Fail Request Aozora Refresh Token Validate")
		return err
	}

	// ①敵対的検証の指摘(MIR-1298): 口座確認はここでは行わない。更新経路は
	// 既に自社のものと確定しているrefresh_tokenを使い回すだけで、口座が
	// 別物にすり替わる余地が無い(ログイン経路と違い、第三者のcodeが
	// 混入する入口が無い)。ここで確認を挟むと、GMO側の一時的な通信障害だけで
	// 既に銀行側では使用済みになった新しいトークンを捨てることになり
	// (refresh_tokenは使い切りのため、捨てても元の値には戻れない)、
	// 守れるものが無いまま失敗の入口を増やすだけだった(初回実装の不具合)。

	// 更新ジョブはDB更新の成否に関わらずメモリを新しい値にする。銀行側では
	// 既に新しいトークンへ入れ替わっており(refresh_tokenは使い切り)、古い値には
	// 戻れないため、アプリを動かし続けるにはメモリを新しい値にするしかない。
	// DBへの反映はここで数回試し、失敗しても次回のcron実行で再試行される。
	setClientInfo(newToken.AccessToken, newToken.RefreshToken)

	if err := updateStoredToken(ctx, newToken); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"logTime": time.Now(),
		}).Error("Fail Client Model Update ")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"logTime": time.Now(),
	}).Info("Success Update Aozora Token")

	return nil
}

func getVaDepositRequest(ctx context.Context, client *corporateclient.APIClient, raId, itemId, authKey, dataFrom, dataTo *string) (*corporateclient.VaDepositTransactionsResponse, error) {
	var nextItemKey optional.String

	if *itemId != "" {
		nextItemKey = optional.NewString(*itemId)
	} else {
		nextItemKey = optional.EmptyString()
	}

	apiAccount := client.VirtualAccountApi
	var opts = corporateclient.DepositTransactionsUsingGETOpts{
		RaId:          optional.NewString(*raId),
		DateFrom:      optional.NewString(*dataFrom),
		DateTo:        optional.NewString(*dataTo),
		SortOrderCode: optional.NewString(ORDER_DESC),
		NextItemKey:   nextItemKey,
	}

	res, httpResp, err := CallWith429Retry(ctx, "VirtualAccountApi.DepositTransactionsUsingGET", func(ctx context.Context) (corporateclient.VaDepositTransactionsResponse, *http.Response, error) {
		return apiAccount.DepositTransactionsUsingGET(ctx, GetAccessToken(), &opts)
	})
	if err != nil {
		var strItemId string

		if itemId != nil {
			strItemId = *itemId
		}

		logrus.WithFields(logrus.Fields{
			"raId":           *raId,
			"dateFrom":       *dataFrom,
			"dateTo":         *dataTo,
			"itemId":         strItemId,
			"httpStatusCode": HTTPStatusCode(httpResp),
			"err":            err,
			"errDetail":      GetSwaggerErrorDetailString(err),
			"logTime":        time.Now(),
		}).Error("fail deposit transactions using get")

		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"raId":           raId,
		"dateFrom":       dataFrom,
		"dateTo":         dataTo,
		"itemId":         itemId,
		"httpStatusCode": HTTPStatusCode(httpResp),
		"count":          len(res.VaTransactions),
		"logTime":        time.Now(),
	}).Info("Get VaDeposit Request")

	for _, t := range res.VaTransactions {

		logrus.WithFields(logrus.Fields{
			"DepositAmount":   t.DepositAmount,
			"TransactionDate": t.TransactionDate,
			"ItemKey":         t.ItemKey,
			"ValueDate":       t.ValueDate,
			"VaId":            t.VaId,
			"logTime":         time.Now(),
		}).Info("Get VaDeposit Request")
	}

	return &res, nil
}

func GetVaTransactions(ctx context.Context, raId, dataFrom, dataTo string) ([]corporateclient.VaTransaction, error) {
	var transactions []corporateclient.VaTransaction
	transactions = make([]corporateclient.VaTransaction, 0)

	client, err := GetCorporateClient()
	if err != nil {
		return nil, errors.Wrap(err, "fail load corporate client")
	}
	itemId := ""
	next := true

	for next {

		result, err := getVaDepositRequest(ctx, client, &raId, &itemId, nil, &dataFrom, &dataTo)
		if err != nil {

			logrus.WithFields(logrus.Fields{
				"raId":     raId,
				"dateFrom": dataFrom,
				"dateTo":   dataTo,
				"itemId":   itemId,
				"res":      result,
				"err":      err,
				"next":     next,
				"logTime":  time.Now(),
			}).Error("fail get Va Deposit Request ")

			next = false

			return nil, err

		}
		logrus.WithFields(logrus.Fields{
			"raId":     raId,
			"dateFrom": dataFrom,
			"dateTo":   dataTo,
			"itemId":   result.NextItemKey,
			"count":    len(result.VaTransactions),
			"logTime":  time.Now(),
		}).Info("GetEquityVaTransactionJob")

		next = result.HasNext
		if next {
			itemId = result.NextItemKey
		}
		transactions = append(transactions, result.VaTransactions...)
	}

	return transactions, nil
}

func WebhookSubscription(status SUBSCRIPTION_STATUS) error {
	logrus.WithFields(logrus.Fields{
		"status":  status,
		"logTime": time.Now(),
	}).Info("Webhook Subscription")

	apiClient := webhookclient.NewAPIClient(webHookClientCfg)
	apiAccount := apiClient.WebhooksApi

	ctx := context.Background()
	authorization := generateAuthorization()
	eventType := webhookclient.EventType{EventType: "va-deposit-transaction"}
	eventTypes := []webhookclient.EventType{eventType}

	body := webhookclient.SubscribeRequestBody{SubscribeStatus: string(status), EventTypes: eventTypes}
	logrus.WithFields(logrus.Fields{
		"status":  status,
		"body":    body,
		"logTime": time.Now(),
	}).Info("Webhook Subscription")

	resp, err := apiAccount.AccountsUsing(ctx, authorization, body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"status":         status,
			"httpStatusCode": HTTPStatusCode(resp),
			"error":          err,
		}).Error("Fail Webhook Subscription")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"status":         status,
		"httpStatusCode": HTTPStatusCode(resp),
		"logTime":        time.Now(),
	}).Info("Success Webhook Subscription")

	return nil
}

func WebHooKGetList(c echo.Context) error {

	apiClient := webhookclient.NewAPIClient(webHookClientCfg)
	apiAccount := apiClient.WebhooksApi

	ctx := c.Request().Context()

	authorization := generateAuthorization()

	res, httpRes, err := apiAccount.AccountsUsingGET(ctx, authorization)
	if err != nil {
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"error":          err,
			"httpStatusCode": HTTPStatusCode(httpRes),
		}).Error(" Fail AccountsUsingGET ")
		return err
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"count": len(res.Messages),
	}).Info("Success AccountsUsingGET")

	return nil
}

func newCorporateConfig() (*corporateclient.Configuration, error) {
	if corporateClientCfg == nil {
		return nil, errors.New("corporateClientCfg is nil")
	}
	return corporateClientCfg, nil
}

func newWebHookConfig() (*webhookclient.Configuration, error) {
	if webHookClientCfg == nil {
		return nil, errors.New("webHookClientCfg is nil")
	}
	return webHookClientCfg, nil
}

func GetCorporateClient() (*corporateclient.APIClient, error) {
	// GMO API 呼び出しで共有するクライアントを初期化
	corporateClientOnce.Do(func() {
		cfg, err := newCorporateConfig()
		if err != nil {
			corporateClientInitErr = errors.Wrap(err, "fail load corporate config")
			return
		}

		cfg.HTTPClient = newRateLimitedHTTPClient()
		corporateClient = corporateclient.NewAPIClient(cfg)
	})

	if corporateClientInitErr != nil {
		return nil, corporateClientInitErr
	}

	if corporateClient == nil {
		return nil, errors.New("corporate client is nil")
	}

	return corporateClient, nil
}

func GetWebHookClient() (*webhookclient.APIClient, error) {
	cfg, err := newWebHookConfig()
	if err != nil {
		return nil, errors.Wrap(err, "fail load webhook config")
	}
	return webhookclient.NewAPIClient(cfg), nil
}

func generateAuthorization() string {
	return "Basic " + com.Base64Encode(fmt.Sprintf("%s:%s", ClientId(), ClientSecret()))
}

// CheckAuthorizationはMIR-1126の調査対象。本来は不一致時にfalseを返すべきだが、
// 現状は一致・不一致どちらでもtrueを返しており、Webhookの認証チェックが実質
// 無効化されている(2024-03-07の初回実装から)。不一致でfalseへ直すと、GMOが
// 送ってくるAuthorizationヘッダーの実際の生成方法が確証できていない状態では、
// 本物の入金通知まで401で弾いてしまう恐れがある。まずは値を一切ログに出さず
// 一致・不一致だけを本番ログで実測し、100%一致することを確認してからfalseへ
// 直す。認証情報(auth・生成した期待値)はログへ絶対に出さない(MIR-1122/
// MIR-1124で平文出力を直した経緯があるため)。
//
// ctxからrequest_idを取り出してログへ載せる。同じリクエストの他のログと
// 突き合わせて、GMOからの本物の通知か無関係な通信かを切り分けるため。
// 突き合わせ先はAozoraWebHook内の各ログ(Header Info等、factory.Logger経由
// でこちらもrequest_idが乗らない)ではなく、echoutils.CustomLogger()が出す
// アクセスログ(remote_ip・user_agent・uri・status等を持つ)の方になる。
// CustomLoggerはX-Request-IDヘッダーを見ており、CustomContextがレスポンス
// ヘッダーに同じridをセットしているため一致する。
// factory.Loggerは使わない。それが読むechomiddleware.ContextLoggerNameは
// このアプリのミドルウェアに配線されておらず常にフォールバックしてしまい、
// request_idが載らないことを敵対的検証で確認した。実際にrequest_idが乗るのは
// pkg/echoutils/echo.goのCustomContextがcontext.WithValueで直接設定している
// 生の文字列キー"request_id"の方(pkg/factory/factory.goのloggerFactory.Entry
// と同じ読み方)。
func CheckAuthorization(ctx context.Context, auth string) bool {
	matched := auth == generateAuthorization()
	logrus.WithFields(logrus.Fields{
		"matched":    matched,
		"request_id": ctx.Value("request_id"),
		"logTime":    time.Now(),
	}).Info("CheckAuthorization")

	if matched {
		return true
	}

	return true
}
