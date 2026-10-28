package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type JobApiController struct{}

type REQUEST_TIME_TYPE string

const (
	DEPOSIT_CHECK         REQUEST_TIME_TYPE = "DEPOSIT_CHECK"
	ISSUE_VIRTUAL_ACCOUNT REQUEST_TIME_TYPE = "ISSUE_VIRTUAL_ACCOUNT"
	TRANSFER_REQUEST      REQUEST_TIME_TYPE = "TRANSFER_REQUEST"
	TRANSFER_RESULT       REQUEST_TIME_TYPE = "TRANSFER_RESULT"
	DEPOSIT_PROCESS       REQUEST_TIME_TYPE = "DEPOSIT_PROCESS"
)

func (c JobApiController) Init(g echoswagger.ApiGroup) {

	g.GET("/execute-time", c.JobExecuteTime,
		echoutils.AllowRoles(domain.ROLE_PAYMENT)).
		AddParamQueryNested(JobExecuteTimeSearchInput{}).SetDescription("外部API最終実行時間").
		AddResponse(http.StatusOK, "API実行時間", nil, nil).
		SetSummary("API実行時間照会").
		SetSecurity("Authorization")

	// MIR-1243: ジョブを実行する副作用のある操作のため、GETではなくPOSTにする(GETは再送や先読みの対象になりうる)。
	g.POST("/execute-deposit-check-api", c.DepositCheckJobExecute,
		echoutils.AllowRoles(domain.ROLE_PAYMENT)).
		AddResponse(http.StatusOK, "入金確認API連動ジョブ実行", nil, nil).
		SetSummary("入金確認API実行").
		SetDescription("入金確認API実行時間を照会します.").
		SetSecurity("Authorization")

	g.POST("/execute-deposit-process-api", c.DepositAutoProcessJobExecute,
		echoutils.AllowRoles(domain.ROLE_PAYMENT)).
		AddResponse(http.StatusOK, "保有預託金自動入金処理API連動ジョブ実行", nil, nil).
		SetSummary("保有預託金自動入金実行API実行").
		SetDescription("保有預託金自動入金実行APIを実行します.").
		SetSecurity("Authorization")

	g.GET("/gmo-connect-url", c.GmoConnectUrl,
		echoutils.AllowRoles(domain.ROLE_BASE)).
		AddResponse(http.StatusOK, "GMO青空銀行ログインURL要求", nil, nil).
		SetSummary("GMO青空銀行ログインURL要求").
		SetDescription("GMO青空銀行ログインURL要求").
		SetSecurity("Authorization")
}

func (JobApiController) JobExecuteTime(c echo.Context) error {
	var v JobExecuteTimeSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	lastExecuteTime, err := domain.JobExecuteTime{}.GetJobLastExecuteTime(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if lastExecuteTime == nil {
		return apirender.RenderSucc(c, http.StatusOK, "0000-00-00 00:00:00")
	}

	switch v.RequestTime {
	case DEPOSIT_CHECK:
		return apirender.RenderSucc(c, http.StatusOK, lastExecuteTime.DepositCheckTime)
	case ISSUE_VIRTUAL_ACCOUNT:
		return apirender.RenderSucc(c, http.StatusOK, lastExecuteTime.VaAccountIssueTime)
	case TRANSFER_REQUEST:
		return apirender.RenderSucc(c, http.StatusOK, lastExecuteTime.TransferRequestTime)
	case TRANSFER_RESULT:
		return apirender.RenderSucc(c, http.StatusOK, lastExecuteTime.TransferResultTime)
	case DEPOSIT_PROCESS:
		return apirender.RenderSucc(c, http.StatusOK, lastExecuteTime.DepositProcessTime)
	default:
		return factory.ErrorParameter.New(errors.New("invalid 'RequestTime' type"))
	}
}

// isJobSchedulerCallFailed はjobs-schedulerからのレスポンスが失敗かどうかを判定する。
func isJobSchedulerCallFailed(statusCode int) bool {
	return statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices
}

// callJobScheduler はjobs-schedulerへリクエストし、レスポンス本文とステータス
// コードを返す。リクエストの構築自体ができない場合・接続自体ができない場合
// (タイムアウト・DNS失敗等)やレスポンス本文が読めない場合、Goのerrには
// jobs-schedulerのホスト名等の内部情報が含まれることがあるため、その内容は
// ログにのみ残し、呼び出し元(=画面)へは汎用的なfactory.ErrorServiceUnavailable
// だけを返す。
func callJobScheduler(ctx context.Context, method, url, name string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		factory.Logger(ctx).WithError(redactURLFromErr(err, url)).Error("fail to build request for " + name)
		return nil, 0, factory.ErrorServiceUnavailable.New(errors.New("jobs-scheduler is unreachable"))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		factory.Logger(ctx).WithError(redactURLFromErr(err, url)).Error("fail to connect jobs-scheduler for " + name)

		return nil, 0, factory.ErrorServiceUnavailable.New(errors.New("jobs-scheduler is unreachable"))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		factory.Logger(ctx).WithError(err).Error("fail to read jobs-scheduler response for " + name)

		return nil, 0, factory.ErrorServiceUnavailable.New(errors.New("jobs-scheduler is unreachable"))
	}

	return body, resp.StatusCode, nil
}

// respondJobSchedulerResult はjobs-schedulerのレスポンスを検査し、失敗なら
// factory.ErrorServiceUnavailable を返す。生レスポンスは画面へ転記しない。
// ALBのゲートウェイエラー等でHTMLが返ることがあり、また将来jobs-scheduler側の
// エラー文言に内部情報が混じる変更が入っても、画面側の安全性を個別に見直さずに
// 済むため。画面へはステータスコードのみ返し、本文はログにのみ残す。
func respondJobSchedulerResult(ctx context.Context, statusCode int, body []byte, name string) error {
	if isJobSchedulerCallFailed(statusCode) {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"body":        string(body),
			"status_code": statusCode,
		}).Error("fail " + name)

		return factory.ErrorServiceUnavailable.New(fmt.Errorf("jobs-scheduler returned status %d", statusCode))
	}

	// 成功時は本文をログに残さない(code-review指摘)。GmoConnectUrlの本文は
	// GMOあおぞら銀行のOAuth認可URL(client_id・state等を含む)そのものなので、
	// 常時ログへ出すとCLAUDE.mdの「No secrets in logs」に触れる。診断に要るのは
	// 失敗時の本文であり、成功時に残す必要はない。
	factory.Logger(ctx).Info("success " + name)

	return nil
}

func (JobApiController) DepositCheckJobExecute(c echo.Context) error {
	ctx := c.Request().Context()
	url := serviceutils.GetJobSchedulerUrl() + "/va-accounts/job/execute"

	body, statusCode, err := callJobScheduler(ctx, http.MethodPost, url, "deposit check job execute")
	if err != nil {
		return err
	}

	if err := respondJobSchedulerResult(ctx, statusCode, body, "deposit check job execute"); err != nil {
		return err
	}

	return apirender.RenderSucc(c, http.StatusOK, nil)
}

func (JobApiController) DepositAutoProcessJobExecute(c echo.Context) error {
	// jobs-scheduler側の /depositprocess グループは2024-09-04(e60922f5)で削除されて
	// おり、このURLは存在しない(常に404)。同じジョブ(DEPOSIT_AUTO_PROCESS)を実行する
	// execute グループのルートはQA専用ガード付きで、本番から呼ぶと空ボディ200を返すため
	// (MIR-1186で確認)、こちらへ付け替えるとかえって失敗を検知できなくなる。本番で
	// 使える手動実行ルートの新設はMIR-1194へ切り出した。それまでは404のまま、
	// 「常に失敗」を正直に返す。
	ctx := c.Request().Context()
	url := serviceutils.GetJobSchedulerUrl() + "/depositprocess/job/execute"

	body, statusCode, err := callJobScheduler(ctx, http.MethodPost, url, "deposit auto process job execute")
	if err != nil {
		return err
	}

	if err := respondJobSchedulerResult(ctx, statusCode, body, "deposit auto process job execute"); err != nil {
		return err
	}

	return apirender.RenderSucc(c, http.StatusOK, nil)
}

// gmoConnectUrlStateTTL はGmoConnectUrlが発行するワンタイムstateの有効期限
// (MIR-1298)。法人ログインのワンタイムパスワード入力までを見込んだ想定値。
const gmoConnectUrlStateTTL = 30 * time.Minute

func (JobApiController) GmoConnectUrl(c echo.Context) error {
	var result struct {
		Result  string `json:"result"`
		Success bool   `json:"success"`
	}

	ctx := c.Request().Context()

	// ②レビューの指摘(MIR-1298): TokenToUserは既にfactory.ErrorTokenInvalid
	// (401)でラップして返すため、ここでErrorPermissionDenied(403)へ包み
	// 直さない。このリポジトリの他の52箇所と同じくそのまま返す。
	adminUserId, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	state, err := newGmoConnectUrlState()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}
	lookup := domain.Sha256Hex(state)
	if err := (domain.GmoAozoraOAuthState{}).Create(ctx, lookup, adminUserId, time.Now().UTC().Add(gmoConnectUrlStateTTL)); err != nil {
		return factory.ErrorDB.New(err)
	}

	// POSTにしない(GETのまま)。POSTにすると、この行のINSERTがコミットされる
	// 前にjobs-schedulerを呼ぶことになり、行が見えずに400になる(このハンドラの
	// DBセッションはPOST/PUT/DELETE/PATCHのときだけ明示的にコミットされる。
	// apps/admin-api/server/server.goのcontextDBWithNameとの整合)。
	url := serviceutils.GetJobSchedulerUrl() + "/gmo-aozora/auth/login?state=" + neturl.QueryEscape(state)

	body, statusCode, err := callJobScheduler(ctx, http.MethodGet, url, "gmo connect url")
	if err != nil {
		return err
	}

	if err := respondJobSchedulerResult(ctx, statusCode, body, "gmo connect url"); err != nil {
		return err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return factory.ErrorSystem.New(err)
	}
	if !result.Success {
		return factory.ErrorSystem.New(errors.New("GMO Connect Url Fail"))
	}

	return apirender.RenderSucc(c, http.StatusOK, result.Result)
}

// newGmoConnectUrlState は、GMOあおぞら銀行の認可コールバックの戻り先を
// 検証するためのワンタイム値Sを作る(MIR-1298)。crypto/randの32バイトを
// hexにする(256bit、推測不可能な長さ)。
func newGmoConnectUrlState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// redactURLFromErr は、rawURLのクエリ文字列(?以降)がerrのメッセージへ
// そのまま含まれている場合に取り除く(MIR-1298)。GmoConnectUrlが呼ぶ
// jobs-schedulerのURLは?state=<S>を含み、net/httpが返すエラー(*url.Error)は
// メッセージにリクエストURLをそのまま埋め込むため、接続失敗時のログへ
// stateが漏れる。他の呼び出し元にクエリ文字列を含むURLは無く、
// 影響しない。
func redactURLFromErr(err error, rawURL string) error {
	if err == nil {
		return nil
	}
	idx := strings.Index(rawURL, "?")
	if idx == -1 {
		return err
	}

	// ①敵対的検証の指摘(MIR-1298): net/httpが返す*url.Error(http.NewRequest*・
	// Client.Doの両方が返しうる)は、埋め込んだURLをStripPassword等で
	// 書き換えていることがある(userinfoを含む場合等)ため、rawURLの生の文字列と
	// 完全一致するとは限らない。err.Error()に対する文字列置換に頼らず、
	// *url.Errorの構造からURLフィールドを直接クエリ無しへ書き換える方が確実。
	var urlErr *neturl.Error
	if errors.As(err, &urlErr) {
		redacted := *urlErr
		redacted.URL = redactQueryString(urlErr.URL)
		return &redacted
	}

	// *url.Error以外(想定外の経路)は、フォールバックとして文字列置換する。
	msg := err.Error()
	if !strings.Contains(msg, rawURL) {
		return err
	}
	return errors.New(strings.ReplaceAll(msg, rawURL, rawURL[:idx]))
}

func redactQueryString(rawURL string) string {
	if idx := strings.Index(rawURL, "?"); idx != -1 {
		return rawURL[:idx]
	}
	return rawURL
}
