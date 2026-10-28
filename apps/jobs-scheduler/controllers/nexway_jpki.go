package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	userservice "github.com/illuminarean/energyshares-user-api/service"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

// NexwayJpkiWebhookRequest は、本人確認デジタル認証サービスの検証結果通知(Webhook)の本体
// (仕様書 p.17)。基本4情報は含まれない。このWebhookには認証も署名も無い。
type NexwayJpkiWebhookRequest struct {
	ReferenceId string `json:"reference_id"` // URL発行依頼APIで設定した連携ID
	RequestId   string `json:"request_id"`   // デジタル認証サービスが発行した検証依頼ID
	Result      string `json:"result"`       // succeeded / failed / error / canceled
	CompletedAt string `json:"completed_at"` // YYYY-MM-DDThh:mm:ss+09:00
}

// NexwayJpkiControllers は、JPKI方式(電子署名検証方式)の署名検証結果を受け取り、
// 署名に成功した会員を本人確認BPOサービスへ申請する(MIR-1470)。
//
// BPOからの結果通知(ekyc_type=jpki)は、ホ方式と同じ /nexway/bpo/webhook で受ける(MIR-1468)。
type NexwayJpkiControllers struct {
	privateUserService *userservice.PrivateUserService
	jpkiService        service.IJpkiNexwayService
	// notifier は、署名が完了しなかった会員へ、本人確認のやり直しをお願いするメールを送る。
	// nilなら送らない(テスト用)。
	notifier jpkiRetryNotifier
}

func NewNexwayJpkiControllers() *NexwayJpkiControllers {
	return &NexwayJpkiControllers{
		privateUserService: userservice.NewPrivateUserService(),
		jpkiService:        service.NewJpkiNexwayService(),
		notifier:           newMailJpkiRetryNotifier(),
	}
}

func (c NexwayJpkiControllers) Init(g echoswagger.ApiGroup) {
	g.POST("/webhook", c.CheckSignatureResult).
		AddParamBody(NexwayJpkiWebhookRequest{}, "body", "", true).
		AddResponse(http.StatusOK, "Webhook呼び出し true/false", "", nil).
		SetDescription("本人確認デジタル認証サービスの署名検証結果をwebhookで受け取り、署名に成功した場合は本人確認BPOサービスへ申請します。").
		SetSummary("JPKI方式 署名検証結果")
}

// CheckSignatureResult は、署名検証結果のWebhookを受け取る。
//
// 応答のHTTPステータスは、Nexwayの再送(15分間隔で2回)を使うかどうかを決める。
//   - 200: 処理した、または再送しても意味が無い(対象外・不正な通知・既に処理済み)
//   - 400: 本体が読めない・必須項目が無い
//   - 500: 今は処理できないが、再送すれば成功しうる(通信障害・DB障害など)
func (c NexwayJpkiControllers) CheckSignatureResult(ctx echo.Context) error {
	// 認証の無い公開口のため、BPOの結果通知と同じく、本文の大きさに上限を設ける。
	if _, status := readNexwayWebhookBody(ctx); status != 0 {
		return ctx.JSON(status, ApiResult{Success: false, Result: false})
	}
	var request NexwayJpkiWebhookRequest
	if err := ctx.Bind(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, ApiResult{Success: false, Result: false})
	}
	if request.ReferenceId == "" || request.RequestId == "" {
		return ctx.JSON(http.StatusBadRequest, ApiResult{Success: false, Result: false})
	}

	logrus.WithFields(nexwayJpkiLogFields(request)).Info("check nexway jpki signature result")

	if err := c.handleSignatureResult(ctx.Request().Context(), request, signatureResultFromWebhook); err != nil {
		logrus.WithFields(nexwayJpkiLogFields(request)).WithError(err).Error("nexway jpki signature result: failed to process")
		return ctx.JSON(http.StatusInternalServerError, false)
	}
	return renderSucc(ctx, http.StatusOK, true)
}

// nexwayJpkiLogFields は、ログへ出してよい項目だけを返す(構造体を丸ごと渡さない)。
func nexwayJpkiLogFields(request NexwayJpkiWebhookRequest) logrus.Fields {
	return logrus.Fields{
		"reference_id": request.ReferenceId,
		"request_id":   request.RequestId,
		"result":       request.Result,
		"time":         time.Now().String(),
	}
}

// signatureResultSource は、署名の結果を処理するきっかけ。
type signatureResultSource int

const (
	// signatureResultFromWebhook は、ネクスウェイからの検証結果通知(Webhook)。送信元を確かめられない。
	signatureResultFromWebhook signatureResultSource = iota
	// signatureResultFromJob は、通知を取りこぼした記録を拾い直す定期ジョブ。対象は自社の記録から
	// 選んでおり、通知のように外から送り込まれたものではない。
	signatureResultFromJob
)

// errJpkiSignatureNotFinal は、署名の結果がネクスウェイ側でまだ確定していないことを表す。
// Webhookでは再送で確認し直し、ジョブでは次の回で確認し直す。異常ではない。
var errJpkiSignatureNotFinal = errors.New("signature result is not final on nexway")

// handleSignatureResult は署名の結果を処理する。nilを返すと200、エラーを返すと500(Nexwayが再送)になる。
// 定期ジョブも、会員の記録から作った通知と同じ形の request で、同じ処理を呼ぶ。
//
// Webhookの内容は信用しない。このWebhookには認証も署名も無く、reference_idとrequest_idを
// 知っていれば誰でも「succeeded」を送れる。送られると本人確認BPOサービスへの申請(有料)が
// 走り、「failed」を送られると会員の進行中の本人確認が失敗扱いになる。そのため、通知が
// 何と言っていても、デジタル認証サービスの検証結果取得APIで確認し、その結果で判断する。
func (c NexwayJpkiControllers) handleSignatureResult(ctx context.Context, request NexwayJpkiWebhookRequest, source signatureResultSource) error {
	log := logrus.WithFields(nexwayJpkiLogFields(request))

	ekycRequest, err := c.privateUserService.GetEkycRequestLog(ctx, request.ReferenceId)
	if err != nil {
		return errors.Wrap(err, "failed to get eKYC request log")
	}
	if ekycRequest.Id == 0 {
		// 自社に無い連携ID。存在しないIDでNexwayへの通信を起こさせないため、ここで止める。
		log.Warn("nexway jpki signature result: unknown reference_id, ignored")
		return nil
	}
	// この通知が、記録(JPKI方式の署名用URLの発行時に検証依頼IDを記録した行)のものかを確かめる。
	// 検証依頼IDを持たない行(ホ方式の行)や、別の検証依頼IDの行へは、申請も失敗の記録もしない。
	// Nexwayへの照会でも検証依頼IDは突き合わせるが、こちらは自社の記録との突合で、守りを2重にする。
	if ekycRequest.RequestId == nil || *ekycRequest.RequestId != request.RequestId {
		log.WithField("ekyc_request_log_id", ekycRequest.Id).
			Error("nexway jpki signature result: request_id does not match the eKYC request log, ignored")
		return nil
	}

	// 処理済みかどうかは、「BPOへ申請済み(申請IDあり)」か「失敗と記録済み」だけで判断する。
	// 状態(Status)が成功(SUCCESS)でも、申請IDが無ければ、まだBPOへ申請していない。会員の画面からの
	// 完了報告(PUT /signup/ekyc)が、署名結果の通知より先に届くと、行はSUCCESSになる。それを
	// 「処理済み」として何もしないと、BPOへ一度も申請されず、結果も来ない。
	// (画面は signature_verified が true になってから報告する設計で、user-apiも、申請IDの無い
	// JPKIの行への完了報告を断る。ここは、その守りが外れた場合に備える)
	if ekycRequest.ApplicantId != 0 || ekycRequest.Status == domain.EKYC_REQUEST_STATUS_FAIL {
		// 既にBPOへ申請済み、または失敗と記録済み。Nexwayの再送や重複した通知。
		log.WithFields(logrus.Fields{
			"status":       ekycRequest.Status,
			"applicant_id": ekycRequest.ApplicantId,
		}).Info("nexway jpki signature result: already processed, ignored")
		return nil
	}

	// Nexwayへの通信は、受信リクエストのctxから独立させる。この受信口は無認証で、送信元は
	// TCP接続を握っている。受信ctxをそのまま使うと、POST直後に接続を切られるだけで、
	// BPOへの申請(受け付けられたかが不明になる)を途中で打ち切らせられてしまう。
	apiCtx, cancel := context.WithTimeout(context.Background(), 2*service.NexwayApiTimeout)
	defer cancel()

	verified, err := c.jpkiService.GetSignatureResult(apiCtx, request.RequestId)
	if err != nil {
		if errors.Is(err, service.ErrNexwayJpkiRequestNotFound) {
			if source == signatureResultFromJob {
				// ジョブの対象は、自社が署名用URLを発行して記録した行。ネクスウェイが引けないのは、
				// 保管期間が過ぎた等で、これ以上待っても結果は取れない。失敗として閉じ、会員へ知らせる。
				log.WithError(err).Warn("nexway jpki signature result: request_id was not found on nexway for a recorded request, marked as failed")
				return c.failSignature(ctx, ekycRequest, "request not found on nexway", nil)
			}
			log.WithError(err).Error("nexway jpki signature result: request_id does not exist on nexway, ignored")
			return nil
		}
		return errors.Wrap(err, "failed to verify signature result with nexway")
	}
	if verified.ReferenceId != request.ReferenceId || verified.RequestId != request.RequestId {
		// 他人のrequest_idを騙った通知の可能性。確認結果は別の申請のものなので使えない。
		log.WithFields(logrus.Fields{
			"verified_reference_id": verified.ReferenceId,
			"verified_request_id":   verified.RequestId,
		}).Error("nexway jpki signature result: webhook does not match nexway's record, ignored")
		return nil
	}
	if verified.Status != service.JpkiStatusCompleted {
		// 通知は完了後に送られるはずなのに、確認すると完了していない。再送で確認し直す。
		return errors.Wrapf(errJpkiSignatureNotFinal, "status=%s", verified.Status)
	}
	if request.Result != "" && verified.Result != request.Result {
		log.WithField("verified_result", verified.Result).
			Warn("nexway jpki signature result: webhook result differs from nexway's record, using nexway's")
	}

	switch verified.Result {
	case service.JpkiResultSucceeded:
		// この行が新しい申請で置き換えられていれば、この申請の結果は会員の状態を動かさない。有料のBPOへ
		// 申請しても無駄になるため、申請せず、この行を失敗にする(isSupersededJpkiRequest 参照)。
		// 画面を開いただけの行(署名のあとにホ方式の画面を開いた、など)では置き換えられたとみなさない。
		// みなすと、署名に成功した会員がBPOへ申請されず、本人確認が進まなくなる。
		superseded, latestErr := isSupersededJpkiRequest(ctx, ekycRequest)
		if latestErr != nil {
			return errors.Wrap(latestErr, "failed to check whether the eKYC request has been superseded")
		}
		if superseded {
			log.WithField("ekyc_request_log_id", ekycRequest.Id).
				Warn("nexway jpki signature result: the eKYC request has been superseded by a newer one, bpo application skipped")
			return c.markFailed(ctx, ekycRequest, "superseded by a newer eKYC request")
		}
		return c.registerBpoApplicant(ctx, apiCtx, ekycRequest, request)
	case service.JpkiResultFailed, service.JpkiResultError, service.JpkiResultCanceled:
		return c.failSignature(ctx, ekycRequest, "signature "+verified.Result, verified)
	default:
		// unverified(未検証)や未知の値。確定した結果ではないので、再送で確認し直す。
		return errors.Wrapf(errJpkiSignatureNotFinal, "result=%s", verified.Result)
	}
}

// isSupersededJpkiRequest は、JPKI方式の行が、同じ会員の新しい申請で置き換えられているかを返す。
//   - より新しく、実際に申し込まれた行がある(HasNewerAppliedRequest。MIR-1468 の BPO の結果と同じ判定)
//   - より新しい JPKI方式の行がある(HasNewerJpkiRequest。署名用URLの発行し直し)
//
// 画面を開いただけのホ方式の行(NONE・結果なし)は、どちらにも当たらない。
func isSupersededJpkiRequest(ctx context.Context, ekycRequest *domain.EkycRequestLog) (bool, error) {
	applied, err := ekycRequest.HasNewerAppliedRequest(ctx)
	if err != nil || applied {
		return applied, err
	}
	return ekycRequest.HasNewerJpkiRequest(ctx)
}

// markFailed は、eKYCリクエスト記録を失敗にする。会員の画面は、この状態を見て失敗を表示する。
// 会員へのメールは送らない(後から新しい発行で置き換えられた行など、会員へ知らせる必要の無い失敗用)。
func (c NexwayJpkiControllers) markFailed(ctx context.Context, ekycRequest *domain.EkycRequestLog, reason string) error {
	_, err := c.markFailedWithMessage(ctx, ekycRequest, reason, "")
	return err
}

// markFailedWithMessage は、eKYCリクエスト記録を失敗にし、失敗の理由(message)を記録する。
// 書き換えた行数を返す。0は、既に確定した行(SUCCESS・FAIL)で、何も変えていない。
func (c NexwayJpkiControllers) markFailedWithMessage(ctx context.Context, ekycRequest *domain.EkycRequestLog, reason, message string) (int64, error) {
	updated, err := ekycRequest.MarkFailedIfNoneWithMessage(ctx, message)
	if err != nil {
		return 0, errors.Wrap(err, "failed to mark eKYC request log as failed")
	}
	fields := logrus.Fields{
		"reference_id": ekycRequest.EkycId,
		"user_id":      ekycRequest.UserId,
		"reason":       reason,
		"updated":      updated,
		"time":         time.Now().String(),
	}
	if updated == 0 {
		// 結果が確定していない(NONE)行にだけ書く条件付き更新のため、0件は、既に確定した行(SUCCESS等)。
		logrus.WithFields(fields).Warn("nexway jpki signature result: eKYC request was not marked as failed (already decided)")
		return 0, nil
	}
	logrus.WithFields(fields).Info("nexway jpki signature result: eKYC request marked as failed")
	return updated, nil
}

// failSignature は、署名が完了しなかった(失敗・中断・有効期限切れ・取れなくなった)行を失敗にし、
// 失敗の理由(エラーコード)を記録して、会員へ本人確認のやり直しをお願いするメールを送る。
//
// verified は、検証結果取得APIの結果。取れていない(nil)ときは、理由を記録せず、具体的な理由を
// 添えずにメールを送る。メールは、この呼び出しで行を失敗にできたとき(書き換えた行数が1)だけ送る。
// Webhookの再送や、ジョブとWebhookの同時の処理で、同じ失敗を2回処理しても、メールは1通になる。
func (c NexwayJpkiControllers) failSignature(ctx context.Context, ekycRequest *domain.EkycRequestLog, reason string, verified *service.JpkiSignatureResult) error {
	var message string
	var memberReasons []string
	var errorCodes []string
	if verified != nil {
		message = verified.ErrorMessage()
		memberReasons = verified.MemberReasons()
		errorCodes = verified.ErrorCodes()
	} else {
		memberReasons = service.JpkiSignatureResult{}.MemberReasons()
	}

	updated, err := c.markFailedWithMessage(ctx, ekycRequest, reason, message)
	if err != nil {
		return err
	}
	if updated == 0 {
		return nil
	}
	if len(errorCodes) > 0 {
		logrus.WithFields(logrus.Fields{
			"reference_id": ekycRequest.EkycId,
			"user_id":      ekycRequest.UserId,
			"error_codes":  errorCodes,
		}).Info("nexway jpki signature result: signature error codes recorded")
	}
	// 行は既に失敗として確定している。メールは、受信リクエストのctxから独立させる(値は引き継ぐ)。
	// Webhookの送信元が接続を切るだけで、ここでメールの送信が打ち切られ、会員へ知らせられないまま
	// 行だけが失敗になるのを防ぐ(ネクスウェイへの問い合わせを独立させているのと同じ理由)。
	c.notifyRetryVerify(context.WithoutCancel(ctx), ekycRequest, memberReasons)
	return nil
}

// notifyRetryVerify は、本人確認のやり直しをお願いするメールを送る。失敗しても処理は失敗させない。
// 行は既に失敗として確定しており、再送やジョブの次の回では、この行は「処理済み」で飛ばされ、
// メールだけをもう一度送ることはできない。送れなかったことは、エラーのログに残す。
func (c NexwayJpkiControllers) notifyRetryVerify(ctx context.Context, ekycRequest *domain.EkycRequestLog, reasons []string) {
	if c.notifier == nil {
		return
	}
	fields := logrus.Fields{
		"reference_id": ekycRequest.EkycId,
		"user_id":      ekycRequest.UserId,
	}

	// 会員がその後に新しい申請を申し込んでいる、または新しい署名用URLを発行している場合は、会員は既に
	// やり直している。古い行の失敗を知らせると、会員を混乱させる(古いURLの有効期限切れの通知など)。
	// 判定は、BPOへの申請を飛ばすときと同じ(isSupersededJpkiRequest)。
	superseded, err := isSupersededJpkiRequest(ctx, ekycRequest)
	if err != nil {
		logrus.WithFields(fields).WithError(err).Error("nexway jpki signature result: failed to check whether the eKYC request has been superseded, retry mail skipped")
		return
	}
	if superseded {
		logrus.WithFields(fields).Info("nexway jpki signature result: the eKYC request has been superseded by a newer one, retry mail skipped")
		return
	}

	if err := c.notifier.NotifyRetryVerify(ctx, ekycRequest.UserId, reasons); err != nil {
		logrus.WithFields(fields).WithError(err).Error("nexway jpki signature result: failed to send the retry verification mail")
		return
	}
	logrus.WithFields(fields).Info("nexway jpki signature result: retry verification mail sent")
}

// registerBpoApplicant は、署名に成功した会員を本人確認BPOサービスへ申請する。
// 氏名・生年月日・住所は会員のプロフィールから組み立てる(基本4情報はNexwayが署名結果から
// BPOへ自動連携するため、ミラリタが渡す必要はなく、検証結果取得APIの個人情報も使わない)。
func (c NexwayJpkiControllers) registerBpoApplicant(ctx, apiCtx context.Context, ekycRequest *domain.EkycRequestLog, request NexwayJpkiWebhookRequest) error {
	log := logrus.WithFields(nexwayJpkiLogFields(request))

	privateUser, err := domain.PrivateUser{}.GetById(ctx, ekycRequest.UserId)
	if err != nil {
		return errors.Wrap(err, "failed to get private user")
	}
	if privateUser == nil {
		log.WithField("user_id", ekycRequest.UserId).Error("nexway jpki signature result: user not found, ignored")
		return nil
	}

	input, missing := buildJpkiBpoApplicantInput(*privateUser, request)
	if len(missing) > 0 {
		// プロフィールが欠けたまま送ってもBPOに弾かれる。送らずに失敗として記録する。
		log.WithField("missing", missing).Error("nexway jpki signature result: profile is incomplete, bpo application skipped")
		return c.markFailed(ctx, ekycRequest, "profile is incomplete")
	}

	output, err := c.jpkiService.RegisterBpoApplicant(apiCtx, input)
	switch {
	case err == nil:
		// 申請IDを記録する(画面が「署名成功・BPOへ申請済み」を知る signature_verified の元になり、
		// BPOへ状態を問い合わせる手掛かりにもなる)。記録に失敗してもBPOは申請を受け付け済みだが、
		// 500で返して再送を受ける。再送時は「受付済み」が返るので、重複して申請されることはない。
		updated, err := ekycRequest.SetApplicantIdIfUnset(ctx, output.ApplicantId)
		if err != nil {
			return errors.Wrap(err, "failed to save applicant_id")
		}
		fields := logrus.Fields{"applicant_id": output.ApplicantId, "updated": updated}
		if updated == 0 {
			// 申請IDが既にある(BPOの結果通知が先に届いた等)。上書きしていない。
			log.WithFields(fields).Warn("nexway jpki signature result: bpo applicant registered, applicant_id was already set")
			return nil
		}
		log.WithFields(fields).Info("nexway jpki signature result: bpo applicant registered")
		return nil
	case errors.Is(err, service.ErrNexwayJpkiAlreadyRegistered):
		// BPOが受付済み(4000010・4000011)と返す応答には、申請IDが含まれない(手引き p.28)。
		// 申請の直後に通信が切れた、または申請IDの保存に失敗して再送された場合は、この経路になり、
		// 申請IDは記録されない。この経路での申請IDは、BPOの結果通知(MIR-1468)で記録される。
		log.WithError(err).Info("nexway jpki signature result: bpo applicant was already registered (applicant_id is recorded by the bpo result webhook)")
		return nil
	case errors.Is(err, service.ErrNexwayJpkiApplicantRejected):
		log.WithError(err).Error("nexway jpki signature result: bpo rejected the application")
		return c.markFailed(ctx, ekycRequest, "bpo rejected the application")
	default:
		return errors.Wrap(err, "failed to register bpo applicant")
	}
}

// buildJpkiBpoApplicantInput は、会員のプロフィールからBPOへの申請内容を組み立てる。
// 欠けている項目があれば、その名前を返す(値は返さない)。
//
// 氏名・住所の組み立てはホ方式で画面がNexwayへ渡しているものと同じ(domain.BuildEkycName・
// BuildEkycAddress)。生年月日は日本時間のYYYY-MM-DD。
func buildJpkiBpoApplicantInput(privateUser domain.PrivateUser, request NexwayJpkiWebhookRequest) (service.JpkiBpoApplicantInput, []string) {
	input := service.JpkiBpoApplicantInput{
		ReferenceId: request.ReferenceId,
		RequestId:   request.RequestId,
		Name:        domain.BuildEkycName(privateUser.FamilyName, privateUser.FirstName),
		Birthday:    privateUser.Birthday.ToFormat(time.DateOnly),
		Address:     domain.BuildEkycAddress(privateUser.Province, privateUser.City, privateUser.Address, privateUser.Address2),
	}

	var missing []string
	if strings.TrimSpace(input.Name) == "" {
		missing = append(missing, "name")
	}
	if input.Birthday == "" {
		missing = append(missing, "birthday")
	}
	if strings.TrimSpace(input.Address) == "" {
		missing = append(missing, "address")
	}
	return input, missing
}
