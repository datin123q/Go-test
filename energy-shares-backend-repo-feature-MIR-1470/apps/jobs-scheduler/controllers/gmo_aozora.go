package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"energyshares-job-scheduler/pkg/gmo_aozora"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// ログタグ定数。各処理フェーズのログ識別に使用する。
const (
	WEB_HOOK_POST_TAG    = "GMO AOZORA WEBHOOK POST"
	AOZORA_REDIRECT_POST = "GMO AOZORA REDIRECT POST"
	VA_DEPOSIT_UPLOAD    = "VA DEPOSIT UPLOAD"
)

// AozoraApiControllers は GMO あおぞら銀行 API に関連するHTTPハンドラをまとめたコントローラ。
// 仮想口座への入金Webhookの受信、OAuth認証フローの処理を担う。
type AozoraApiControllers struct {
}

// Init は AozoraApiControllers のルーティングを echoswagger の ApiGroup に登録する。
// 登録されるエンドポイント:
//   - GET  /auth/login       : GMO あおぞら銀行の認証URLを取得しWebhookサブスクリプションを開始する
//   - POST /webhook          : 仮想口座への入金通知Webhookを受信する
//   - GET  /auth/access-token: OAuth認証後のリダイレクトURIを処理しアクセストークンを取得する
func (c AozoraApiControllers) Init(g echoswagger.ApiGroup) {
	g.GET("/auth/login", c.GetLogin).
		AddParamQuery(GetLoginInput{}, "query", "", true).
		AddResponse(http.StatusOK, "GMO Aozora認証成功", nil, nil).
		SetDescription("GMO Aozora認証。stateはadmin-apiが発行したワンタイム値").
		SetSummary("GMO Aozora認証")

	g.POST("/webhook", c.AozoraWebHook).
		AddParamBody(AozoraWebHookMessageInput{}, "body", "", true).
		AddResponse(http.StatusOK, "", nil, nil).
		SetDescription("GMO AozoraWebhook。一般的な入金に関するウェブフック受信").
		SetSummary("GMO AozoraWebhook")

	// 成功・失敗とも HTML で返す(ポップアップで開かれるため)。ここに
	// AddResponse で JSON の型を宣言しても実体と一致しないので付けない。
	g.GET("/auth/access-token", c.AozoraRedirectUri).
		AddParamQuery(RedirectUriInput{}, "body", "", true).
		SetDescription("GMO AozoraRedirect Uri受信。成功・失敗ともHTMLを返す").
		SetSummary("GMO AozoraRedirect Uri受信")
}

// AozoraWebHook は GMO あおぞら銀行から送信される仮想口座入金通知Webhookを処理するハンドラ。
//
// 処理フロー:
//  1. リクエストボディを [AozoraWebHookMessageInput] にバインドし、ヘッダー・口座・取引情報をログ出力する
//  2. Authorizationヘッダーを検証し、不正なリクエストを拒否する
//  3. Webhookの仮想口座取引情報を [domain.VaDeposit] モデルに変換する
//  4. 仮想口座番号からユーザーを特定し、vaDeposit.UserId を設定する（口座未登録の場合は 0）
//  5. VaDeposit レコードをDBに保存する（同じ銀行取引が登録済みの場合は内容一致を確認して成功終了する）
//  6. UserId が特定できた場合、[domain.VaDeposit.DepositEquityProcess] で入金後の口数割当処理を実行する
//  7. スーパー管理者名義で操作履歴（監査ログ）を記録する
//
// エラー:
//   - Authorizationヘッダー不正: 401
//   - リクエストボディ不正 / モデル変換失敗: 400
//   - DB操作失敗 / 同じ取引キーで内容が異なるデータ競合: 500
func (AozoraApiControllers) AozoraWebHook(c echo.Context) error {
	var m AozoraWebHookMessageInput
	if err := c.Bind(&m); err != nil {
		return factory.ErrorParameter.New(err)
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"UserAgent":   c.Request().UserAgent(),
		"x-eventType": c.Request().Header.Get("x-eventType"),
	}).Info(WEB_HOOK_POST_TAG + " Header Info")

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"messageId":    m.MessageId,
		"raId":         m.Account.RaId,
		"raBranchCode": m.Account.RaBranchCode,
		"baseDate":     m.Account.BaseDate,
		"baseTime":     m.Account.BaseTime,
	}).Info(WEB_HOOK_POST_TAG + " Account Info")

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"vaId":              m.VaTransaction.VaId,
		"transactionDate":   m.VaTransaction.TransactionDate,
		"valueDate":         m.VaTransaction.ValueDate,
		"vaBranchCode":      m.VaTransaction.VaBranchCode,
		"depositAmount":     m.VaTransaction.DepositAmount,
		"paymentBankName":   m.VaTransaction.PaymentBankName,
		"paymentBranchName": m.VaTransaction.PaymentBranchName,
		"partnerName":       m.VaTransaction.PartnerName,
		"itemKey":           m.VaTransaction.ItemKey,
	}).Info(WEB_HOOK_POST_TAG + " VaTransaction Info")

	// Check authorization
	if !gmo_aozora.CheckAuthorization(c.Request().Context(), c.Request().Header.Get("Authorization")) {
		return factory.ErrorFailAuthorized.New(errors.New("invalid authorization"))
	}

	vaDeposit, err := m.VaTransaction.ToModel()
	if err != nil {
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"error": err,
		}).Error(WEB_HOOK_POST_TAG + " VaTransaction ToModel Error")

		return factory.ErrorParameter.New(errors.Wrap(err, "invalid va transaction"))
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"vaId":                  vaDeposit.VaId,
		"transactionDate":       vaDeposit.TransactionDate,
		"valueDate":             vaDeposit.ValueDate,
		"vaBranchCode":          vaDeposit.VaBranchCode,
		"depositAmount":         vaDeposit.DepositAmount,
		"paymentBankName":       vaDeposit.PaymentBankName,
		"paymentBranchName":     vaDeposit.PaymentBranchName,
		"partnerName":           vaDeposit.PartnerName,
		"itemKey":               vaDeposit.ItemKey,
		"depositDate":           vaDeposit.DepositDate,
		"useAmount":             vaDeposit.UseAmount,
		"depositStatus":         vaDeposit.DepositStatus,
		"transactionDetailType": vaDeposit.TransactionDetailType,
	}).Info(WEB_HOOK_POST_TAG + " vaDeposit Info")

	account, err := domain.UserVirtualAccount{}.GetUserVirtualAccountByBranchAndAccountNumber(c.Request().Context(), vaDeposit.VaBranchCode, vaDeposit.VaAccountNumber)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if account == nil {
		vaDeposit.UserId = 0
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"itemKey":         vaDeposit.ItemKey,
			"vaBranchCode":    vaDeposit.VaBranchCode,
			"vaAccountNumber": vaDeposit.VaAccountNumber,
			"error":           "Virtual Account is nil",
		}).Error(WEB_HOOK_POST_TAG + " Virtual Account")
	} else {
		vaDeposit.UserId = account.UserId
	}

	created, err := vaDeposit.CreateOrConfirmExisting(c.Request().Context())
	if err != nil {
		if errors.Is(err, domain.ErrVaDepositItemKeyRequired) {
			return factory.ErrorParameter.New(err)
		}
		if errors.Is(err, domain.ErrVaDepositTransactionConflict) {
			factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
				"itemKey": vaDeposit.ItemKey,
				"error":   err,
			}).Error(WEB_HOOK_POST_TAG + " Va deposit transaction conflict")
			return factory.ErrorSystem.New(err)
		}

		return factory.ErrorDB.New(err)
	}
	if !created {
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"itemKey": vaDeposit.ItemKey,
		}).Info(WEB_HOOK_POST_TAG + " Va deposit already registered")
		// 定期照会で先に登録されている場合も、同一取引であれば受付済みとして成功を返す。
		return renderSucc(c, http.StatusOK, nil)
	}

	if vaDeposit.UserId != 0 {
		if err := vaDeposit.DepositEquityProcess(c.Request().Context()); err != nil {
			return factory.ErrorSystem.New(err)
		}
	}

	// audit log: GMO API 실행
	superAdmin, err := domain.AdminUser{}.GetSuperAdmin(c.Request().Context())
	if err != nil {
		return errors.Wrap(err, "fail super admin id")
	}
	history := domain.OperatorHistory{
		UserId: superAdmin.Id,
		Email:  superAdmin.Email,
		Menu:   "commoncode.LOG_CONTENT_TYPE.GMO_API_REQUEST",
		Ip:     "::1",
		State:  domain.OPERATOR_STATE("REGISTER"),
	}
	_, err = history.Create(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(errors.Wrap(err, "failed to create audit log"))
	}

	return renderSucc(c, http.StatusOK, nil)
}

const gmoAozoraSuccessMessage = "GMOあおぞら銀行との連携が完了しました。このウィンドウを閉じてください。"

// aozoraRedirectHTML はAozoraRedirectUriの応答。statusはHTTPステータス、
// messageは画面表示用の文言(MIR-1298)。
type aozoraRedirectHTML struct {
	status  int
	message string
}

// AozoraRedirectUri は GMO あおぞら銀行のOAuth認証後にリダイレクトされる
// エンドポイントのハンドラ(MIR-1298)。ポップアップで開かれるため、成功・
// 失敗とも HTML で返す(c.HTMLで返してnilを返すことでCustomErrorHandlerの
// JSON化を通さない)。失敗は4xx/5xxを返す(アクセスログで成功と区別できる
// ようにするため)。トークン・code・stateはレスポンス・ログのどちらにも
// 出さない。
func (AozoraApiControllers) AozoraRedirectUri(c echo.Context) error {
	result := processAozoraRedirect(c)
	return renderAozoraRedirectHTML(c, result)
}

func renderAozoraRedirectHTML(c echo.Context, result aozoraRedirectHTML) error {
	h := c.Response().Header()
	h.Set("Cache-Control", "no-store")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Content-Security-Policy", "default-src 'none'")
	return c.HTML(result.status, "<!doctype html><html><body>"+result.message+"</body></html>")
}

func processAozoraRedirect(c echo.Context) aozoraRedirectHTML {
	ctx := c.Request().Context()

	var d RedirectUriInput
	if err := c.Bind(&d); err != nil {
		return aozoraRedirectHTML{http.StatusBadRequest, "無効な要求です。管理画面からやり直してください。"}
	}

	if d.Error != nil {
		desc := ""
		if d.ErrorDescription != nil {
			desc = *d.ErrorDescription
		}
		// error・error_descriptionは銀行が返すエラー区分であり秘密情報ではないが、
		// 万一の混入に備えrequest_idとあわせてログにのみ残す(画面には出さない)。
		logrus.WithFields(logrus.Fields{
			"error":      *d.Error,
			"errorDesc":  desc,
			"request_id": ctx.Value("request_id"),
		}).Error(AOZORA_REDIRECT_POST + " bank returned error")
		return aozoraRedirectHTML{http.StatusBadGateway, "銀行側でエラーが発生しました。管理画面からやり直してください。"}
	}

	if d.Code == nil || *d.Code == "" || d.State == nil || *d.State == "" {
		return aozoraRedirectHTML{http.StatusBadRequest, "無効な要求です。管理画面からやり直してください。"}
	}

	stateHash := domain.Sha256Hex(*d.State)
	now := time.Now().UTC()

	row, err := domain.GmoAozoraOAuthState{}.ClaimByStateHash(ctx, stateHash, now)
	if err != nil {
		logrus.WithError(err).WithField("request_id", ctx.Value("request_id")).Error(AOZORA_REDIRECT_POST + " claim state")
		return aozoraRedirectHTML{http.StatusInternalServerError, "システムエラーが発生しました。管理画面からやり直してください。"}
	}

	if row == nil {
		return handleAlreadyClaimedOrInvalidState(ctx, stateHash)
	}

	var nonce string
	if row.Nonce != nil {
		nonce = *row.Nonce
	}

	// ①敵対的検証の指摘(MIR-1298): 交換済みのコードを捨てないよう、これ以降は
	// 親ctxのキャンセル(ポップアップを閉じる等)から切り離す。
	// RequestAozoraNewToken自身も内部でDetachedContextするが、その戻り後に
	// このハンドラが行うRecordResultにも同じ保護が要る。
	workCtx := gmo_aozora.DetachedContext(ctx)

	if _, err := gmo_aozora.RequestAozoraNewToken(workCtx, *d.Code, nonce); err != nil {
		logrus.WithError(err).WithField("request_id", ctx.Value("request_id")).Error(AOZORA_REDIRECT_POST + " request new token")
		recordAozoraRedirectResult(workCtx, row.Id, domain.GmoAozoraOAuthResultFailure)
		return aozoraRedirectHTML{classifyAozoraTokenErrorStatus(err), classifyAozoraTokenErrorMessage(err)}
	}

	// ベストエフォート。結果の記録に失敗しても認可フロー自体は成功として扱う
	// (リロード時の「既に処理済みです」表示の精度が落ちるだけのため)。
	recordAozoraRedirectResult(workCtx, row.Id, domain.GmoAozoraOAuthResultSuccess)

	return aozoraRedirectHTML{http.StatusOK, gmoAozoraSuccessMessage}
}

// recordAozoraRedirectResult はベストエフォート(結果の記録に失敗しても認可
// フロー自体は失敗させない、リロード時の案内の精度が落ちるだけのため)だが、
// ②レビューの指摘(MIR-1298)で失敗を完全に握りつぶしていた点を直した:
// 「処理中です」が出続けて原因が追えなくなるのを避けるため、失敗はWarnで残す。
func recordAozoraRedirectResult(ctx context.Context, id int64, result string) {
	if err := (domain.GmoAozoraOAuthState{}).RecordResult(ctx, id, result); err != nil {
		logrus.WithError(err).WithField("state_id", id).Warn(AOZORA_REDIRECT_POST + " failed to record result")
	}
}

// handleAlreadyClaimedOrInvalidState は、ClaimByStateHashが更新できなかった
// (=行が無い/期限切れ/既に使用済み)場合の分岐。同じstateでのリロード
// (F5・二重送信)なら、既に出した結果をそのまま案内する。
func handleAlreadyClaimedOrInvalidState(ctx context.Context, stateHash string) aozoraRedirectHTML {
	existing, err := domain.GmoAozoraOAuthState{}.GetByStateHash(ctx, stateHash)
	if err != nil {
		logrus.WithError(err).WithField("request_id", ctx.Value("request_id")).Error(AOZORA_REDIRECT_POST + " reload state")
		return aozoraRedirectHTML{http.StatusInternalServerError, "システムエラーが発生しました。管理画面からやり直してください。"}
	}

	if existing != nil && existing.UsedAt != nil {
		if existing.Result == nil {
			// ①敵対的検証の指摘(MIR-1298): 口座確認のレートリミッタ(3秒に1回)で
			// 1回目の処理に数秒かかる。その間にF5等で2回目が来ると、まだresultが
			// 記録されていないだけで1回目は失敗していない。「失敗」と断定せず、
			// 処理中である旨を案内する。
			return aozoraRedirectHTML{http.StatusConflict, "処理中です。しばらくしてから管理画面で連携状況を確認してください。"}
		}
		if *existing.Result == domain.GmoAozoraOAuthResultSuccess {
			return aozoraRedirectHTML{http.StatusOK, "既に処理済みです(連携済み)。このウィンドウを閉じてください。"}
		}
		return aozoraRedirectHTML{http.StatusConflict, "既に処理済みです(失敗)。管理画面からやり直してください。"}
	}

	// 行が存在しない(=stateが未発行/偽造)か、期限切れのどちらか。どちらも
	// 攻撃者へ区別して教える必要は無いため同じ文言にする。
	return aozoraRedirectHTML{http.StatusBadRequest, "リンクの有効期限が切れているか、無効な要求です。管理画面からやり直してください。"}
}

// GetLogin は GMO あおぞら銀行の認証URLを生成して返すハンドラ(MIR-1298)。
// stateはadmin-apiがGmoConnectUrlで発行し、gmo_aozora_oauth_stateへ保存済みの
// ワンタイム値でなければならない——未発行・使用済み・期限切れは400で弾く。
// WebhookSubscriptionはここでは呼ばない(認証前に外から呼べてしまうため、
// RequestAozoraNewToken側で認可の完了後にだけ呼ぶ)。
//
// エラー:
//   - stateが無効: 400
//   - 認証URL生成失敗: 400
func (AozoraApiControllers) GetLogin(c echo.Context) error {
	var in GetLoginInput
	if err := c.Bind(&in); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if in.State == "" {
		return factory.ErrorParameter.New(errors.New("state is required"))
	}

	ctx := c.Request().Context()
	lookup := domain.Sha256Hex(in.State)
	row, err := domain.GmoAozoraOAuthState{}.GetByLookup(ctx, lookup)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if row == nil {
		return factory.ErrorParameter.New(errors.New("invalid or expired state"))
	}
	// ①敵対的検証の指摘(MIR-1298): 同じstateで2回目が来た場合、ここで弾く。
	// 2回目もURLの生成自体は成功しうる(ライブラリはH=sha256(S+salt)を
	// 毎回同じ値で計算するが、nonceは呼ぶたびに新しく作るため)が、
	// state_hashは1回目の書き戻し分から変わらない。2回目を「成功」として
	// 返すと、そのURLで実際にログインしてもnonceが一致せず必ず失敗する
	// (担当者からは原因が分かりにくい無駄な失敗になる)。
	if row.UsedAt != nil || !row.ExpiresAt.After(time.Now().UTC()) || row.StateHash != nil {
		return factory.ErrorParameter.New(errors.New("invalid or expired state"))
	}

	result, err := gmo_aozora.Authorization(in.State)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Aozora Open Id Login Fail")
		return factory.ErrorParameter.New(err)
	}

	stateHash := domain.Sha256Hex(result.State)
	written, err := domain.GmoAozoraOAuthState{}.SetStateHashOnce(ctx, row.Id, stateHash, result.Nonce)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if !written {
		// GetByLookupで見た時点ではstate_hash未設定だったが、その後(並行した
		// 別のGetLogin呼び出しで)書き込まれた。生成したURLのnonceは保存されず
		// 使えないため、成功として返さない。
		return factory.ErrorParameter.New(errors.New("state already claimed by a concurrent request"))
	}

	return renderSucc(c, http.StatusOK, result.URL)
}

// classifyAozoraTokenErrorStatus/Message は、RequestAozoraNewTokenの失敗を
// 画面へ出す文言とHTTPステータスへ振り分ける(MIR-1298の「エラーと画面」節)。
// error_description・err.Error()そのものは画面に出さない(ログにのみ残す)。
func classifyAozoraTokenErrorStatus(err error) int {
	switch {
	case errors.Is(err, gmo_aozora.ErrAccountVerifyFailed):
		return http.StatusBadGateway
	case errors.Is(err, gmo_aozora.ErrTokenSaveFailed):
		return http.StatusInternalServerError
	default:
		return http.StatusBadGateway
	}
}

func classifyAozoraTokenErrorMessage(err error) string {
	switch {
	case errors.Is(err, gmo_aozora.ErrAccountVerifyFailed):
		return "口座を確認できませんでした。管理画面からやり直してください。"
	case errors.Is(err, gmo_aozora.ErrTokenSaveFailed):
		return "システムエラーが発生しました。管理画面からやり直してください。"
	case errors.Is(err, gmo_aozora.ErrTokenInvalid):
		return "銀行側でエラーが発生しました。管理画面からやり直してください。"
	case errors.Is(err, gmo_aozora.ErrTokenExchangeFailed):
		return "銀行との通信に失敗しました。管理画面からやり直してください。"
	default:
		return "銀行側でエラーが発生しました。管理画面からやり直してください。"
	}
}
