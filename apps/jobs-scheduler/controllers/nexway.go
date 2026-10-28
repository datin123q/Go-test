package controllers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	userservice "github.com/illuminarean/energyshares-user-api/service"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

// ネクスウェイからWebhookでユーザー更新があったことを記録するための定数
const (
	nexwayProcessId   = 1
	nexwayProcessName = "Nexway eKYC Webhook"
)

type NexwayControllers struct {
	privateUserService     *userservice.PrivateUserService
	nexwayService          service.INexwayService
	userInfoHistoryService service.IUserInfoHistoryService
}

func NewNexwayControllers() *NexwayControllers {
	return &NexwayControllers{
		privateUserService:     userservice.NewPrivateUserService(),
		nexwayService:          service.NewNexwayService(),
		userInfoHistoryService: service.NewUserInfoHistoryService(),
	}
}

func (c NexwayControllers) Init(g echoswagger.ApiGroup) {
	g.POST("/bpo/webhook", c.CheckNexwayResult, VerifyNexwayBpoSignature()).
		AddParamBody(NexwayEkycResult{}, "body", "", true).
		AddResponse(http.StatusOK, "Webhook呼び出し true/false", "", nil).
		SetDescription("nexwayを通じた本人確認結果をwebhookで受け取ります。").
		SetSummary("nexway本人確認結果")

	g.POST("/bpo/webhook-test", c.CallTestWebhook).
		AddParamBody(CallTestWebhookRequest{}, "body", "", true).
		AddResponse(http.StatusOK, "eKYC最終結果承認/未承認処理成功", "", nil).
		SetDescription("nexwayウェブフックを手動で呼び出します。(STG環境専用)").
		SetSummary("STGテスト用nexwayウェブフック手動呼び出し")
}

func (a NexwayControllers) CheckNexwayResult(c echo.Context) error {
	var result NexwayEkycResult
	if err := c.Bind(&result); err != nil {
		return c.JSON(http.StatusBadRequest, ApiResult{
			Success: false,
			Result:  false,
		})
	}

	logrus.WithFields(logrus.Fields{
		"body": result,
		"time": time.Now().String(),
	}).Info("check nexway result")

	// 本人確認後処理。Webhookの内容をnexwayの照会APIで裏取りする検証
	// (MIR-1223)は、自社DBに存在するeKYCリクエストに対してのみ意味があるため、
	// ekycAfterProcess内でGetEkycRequestLogの直後に行う。
	if a.ekycAfterProcess(c.Request().Context(), result, true) != nil {
		return c.JSON(http.StatusInternalServerError, false)
	}

	return renderSucc(c, http.StatusOK, true)
}

func (a NexwayControllers) CallTestWebhook(c echo.Context) error {
	var request CallTestWebhookRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, ApiResult{
			Success: false,
			Result:  false,
		})
	}
	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, ApiResult{
			Success: false,
			Result:  false,
		})
	}

	// 環境チェック
	if serviceutils.EnvIsProd() {
		return c.JSON(http.StatusForbidden, ApiResult{
			Success: false,
			Result:  false,
		})
	}

	// メールで最新のeKYCリクエスト記録を照会
	ekycReqLog := domain.EkycRequestLog{}
	latestReq, err := ekycReqLog.GetLatestByEmail(c.Request().Context(), request.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ApiResult{
			Success: false,
			Result:  false,
		})
	}

	// nexwayからウェブフック呼び出し時に入力するデータを作成
	var result NexwayEkycResult
	if request.IsApproved { // 最終承認時
		result = NexwayEkycResult{
			ApplicantId: latestReq.ApplicantId,
			ReferenceId: latestReq.EkycId,
			Status:      "completed",
			Result:      "match",
		}
	} else { // 最終却下時
		result = NexwayEkycResult{
			ApplicantId: latestReq.ApplicantId,
			ReferenceId: latestReq.EkycId,
			Status:      "completed",
			Result:      "unmatch",
			VerificationErrors: []NexwayEkycError{
				{ErrorCode: "Q1", ErrorMessage: "顔照合結果がアンマッチ（同一人物でない）"},
			},
			OcrValidationErrors: []NexwayEkycError{
				{ErrorCode: "E1", ErrorMessage: "免許証番号の不整合を検出しました。"},
			},
		}
	}

	// 本人確認後処理。ここはSTG限定の手動テスト用エンドポイントで、自分自身が
	// 組み立てたダミーのApplicantId/ReferenceIdを使うため、nexwayでの裏取り
	// 検証(MIR-1223)は行わない。
	if a.ekycAfterProcess(c.Request().Context(), result, false) != nil {
		return c.JSON(http.StatusInternalServerError, false)
	}

	return renderSucc(c, http.StatusOK, true)
}

// isFinalizedAuditStatusは、審査ステータスが既に確定していて、Webhookの再送で
// 巻き戻してはいけない状態かどうかを返す(MIR-1222)。COMPLETED_AUDIT(本登録承認
// 完了)とINVALIDITY_AUDIT(無効)が対象。
func isFinalizedAuditStatus(status domain.AUDIT_STATUS) bool {
	return status == domain.COMPLETED_AUDIT || status == domain.INVALIDITY_AUDIT
}

// shouldAdvanceOnClearedは、CLEAREDの結果を受けてAuditStatusを
// WAIT_FINAL_APPROVAL_AUDITへ進めてよいかを返す(MIR-1276、①)。進めないのは
// 次の2つ。
//
//   - isFinalizedAuditStatus(確定済み。MIR-1222で既にガード済み)
//   - CHECK_ANTISOCIAL_AUDIT・WAIT_ANTISOCIAL_AUDIT(反社確認中・反社確認待ち)。
//     verifyAndReconcileNexwayResultの照会による裏取りは、本人の過去の正当な
//     CLEARED結果が別経路(反社チェック等)で反社確認中へ落ちた後に再送された
//     ケースを区別できない(同コメント参照)。会員が自分でeKYCのWebhookへ
//     overall_result:clearedを送るだけで、反社チェックの結果を待たずに
//     最終承認待機まで進めてしまう。
//
// HOLD_AUDIT(保留)はここでは対象にしない。判定済みのeKYCリクエストが届き
// 直したか(②)で判断が変わり、保留中の会員がeKYCをやり直して新しい結果の
// CLEAREDを受け取るケースは今までどおり進める必要があるため。②はMIR-1358
// (判定済みの行を使い回さなくなる)が前提のため、別途追加する(MIR-1276の
// 本チケットでは未実装)。
//
// 引数だけで完結する純粋関数。
func shouldAdvanceOnCleared(status domain.AUDIT_STATUS) bool {
	if isFinalizedAuditStatus(status) {
		return false
	}
	return status != domain.CHECK_ANTISOCIAL_AUDIT && status != domain.WAIT_ANTISOCIAL_AUDIT
}

// logAuditStatusGuardEventは、確定済み(MIR-1222)または反社確認中・反社確認待ち
// (MIR-1276)の会員に対する審査ステータスの扱い(スキップした/しなかった)を
// 記録する。overallResultとekycCreatedを含めておくことで、正当な再送(同一の
// eKYCリクエストが繰り返し届く)なのか、それとも新しいeKYCリクエストの結果を
// 握りつぶしてしまったのかを、あとからreference_idで本番のeKYCリクエスト記録と
// 突き合わせて判別できるようにする。どちらの理由でスキップしたかはmessage引数で
// 呼び出し元が示す。
func logAuditStatusGuardEvent(privateUser *domain.PrivateUser, result NexwayEkycResult, ekycCreated jsonutils.JSONTime, message string) {
	logrus.WithFields(logrus.Fields{
		"message":        message,
		"user_id":        privateUser.User.Id,
		"audit_status":   privateUser.AuditStatus,
		"reference_id":   result.ReferenceId,
		"applicant_id":   result.ApplicantId,
		"overall_result": result.OverallResult,
		"ekyc_created":   ekycCreated,
		"time":           time.Now().String(),
	}).Warn("eKYCAfterProcess")
}

// verifyAndReconcileNexwayResult はWebhookで受け取ったresultをそのまま信用せず、
// nexwayの照会API(GetEkycResultByApplicantId)で裏取りする(MIR-1223)。この
// Webhookには署名検証も送信元IP制限も無く、ApplicantId/ReferenceIdさえ知って
// いれば誰でも任意の本人確認結果を送り込める状態だった。
//
// ReferenceIdの一致検証(なりすまし対策)と、照会結果によるResultの補正、
// そして主張(OverallResult)ごとの矛盾チェックは、主張の種類に関わらず常に
// 行う。DETECTED(反社ヒット)を例外にするのは「照会自体が失敗したときの
// フェイルオープンの可否」だけで、照会が成功している場合はCLEARED/UNMATCH/
// DETECTEDすべて矛盾チェックの対象になる。「DETECTEDは会員の状態に関わらず
// 必ず拾うべき情報(MIR-1222)」という制約が正当化するのは通信失敗時に
// フェイルオープンすることだけで、照会が成功しているのに中身を見ない理由
// にはならない。照会成功時はRiskDetect.Resultという正解が手元にあるので、
// 本物の反社ヒット(RiskDetect=detected)はこのチェックを必ず通過し見逃され
// ない。弾かれるのはnexwayが裏付けない偽のDETECTED主張だけになる。
//
// DETECTED主張の矛盾チェックはriskResultしか見ておらず、本人確認結果
// (output.Result)自体は見ていない(「反社ヒットを見逃さない」という
// MIR-1222の制約を優先し、本人確認がunmatchでもDETECTED主張自体は通す)。
// このため、実際には本人確認に失敗した(unmatch)人物が自分自身の正当な
// ApplicantId/ReferenceIdで偽のDETECTEDを送ると、反社チェックの実体さえ
// detectedでありさえすれば通過してしまう。これによりIsEkyc=trueのまま
// WAIT_ANTISOCIAL_AUDITへ進められることを防ぐガードは、呼び出し元
// (ekycAfterProcess)のDETECTEDケース内、result.Resultの値を見る箇所で
// 行っている(詳細はそちらのコメント参照)。マイナンバー承認画面
// (apps/admin-api/controllers/my_number.go)がIsEkycだけを見て本登録完了
// (COMPLETED_AUDIT)へ進めてしまうため、このガードが無いと反社チェック
// 待ちのはずの人物が本登録完了へ到達できてしまう。
//
// CLEAREDは「なりすましによる不正な承認(自己昇格)」、UNMATCHは「なりすまし
// による不当な降格(HOLD_AUDITへの降格・SIGNUP_NONEへのリセット・再本登録
// メール送信という、嫌がらせ・DoSに使える実害)」、DETECTEDは上記のとおり
// 「なりすましによる反社チェック未完了のままの本登録完了到達」に、それぞれ
// 使われうる。有効なApplicantId/ReferenceIdの組を知っていればどれも偽装
// できるため、3つとも照会結果との矛盾チェックの対象にする。
//
//	CLEARED主張   照会結果がResult=matchかつRiskDetect.Result=clearedである
//	              ことを要求する
//	UNMATCH主張   照会結果がResult=unmatchであることを要求する
//	DETECTED主張  照会結果がRiskDetect.Result=detectedであることのみを要求
//	              する(本人確認結果自体がmatch/unmatchのどちらか、あるいは
//	              空かは矛盾チェックの対象にしない。空でも弾くと、本物の
//	              反社ヒットがResult欠損という別要因で握りつぶされ「反社
//	              ヒットを見逃さない」というMIR-1222の意図に反するため。
//	              IsEkycを立ててよいかは呼び出し元が別途判定する。詳細は
//	              ekycAfterProcessのDETECTEDケース内のコメント参照)
//
// 照会APIのレスポンスにOverallResult相当のフィールドは無いため直接比較は
// できないが、上記の組み合わせから外れるのは明らかな矛盾なので、これだけで
// 十分に弾ける(nexway側の実際のOverallResult算出ルールが完全一致するかは
// 未確認)。CLEARED/UNMATCH/DETECTED以外の値(未知の値・空文字)は矛盾
// チェックの対象にせず素通りさせる。nexwayが仕様追加した新しい値を弾いて
// 本人確認を止めてしまわないためだが、気付けるようログには残す。
//
// 照会結果のStatusがcompleted以外(または空)の場合も弾く(フェイルクローズ、
// 理由は下のチェック箇所のインラインコメント参照)。ReferenceIdがWebhookの
// ReferenceIdと一致しない場合も、主張の種類によらず弾く。「他人の
// ApplicantIdを騙って自分のReferenceIdを送ってきた」なりすましの可能性が
// ある。
//
// ここで防げるのは「nexwayが返した結果が本物か」までで、「過去の正当な
// CLEARED結果を、会員が別経路(反社チェック等)でCHECK_ANTISOCIAL_AUDIT/
// WAIT_ANTISOCIAL_AUDIT/HOLD_AUDITへ落ちた後に再送された場合」は防げない。
// nexway側の記録は変わらずCLEAREDのままなので、照会による裏取りでは原理的に
// 区別できない。この層の防御(今の会員状態に対して適用してよいかのガード)は
// MIR-1222側の対象範囲。CHECK_ANTISOCIAL_AUDIT/WAIT_ANTISOCIAL_AUDITは
// MIR-1276でshouldAdvanceOnClearedへ追加した。HOLD_AUDITは、判定済みの
// eKYCリクエストが届き直したか(MIR-1358が前提)で判断が変わるため、
// MIR-1276の本対応には含めず別途追加する予定(MIR-1223のスコープ外)。
//
// 照会自体が失敗した場合、失敗の種類とresultの主張で扱いを分ける。
//
//	DETECTED主張 かつ 通信自体の失敗(service.ErrNexwayUnavailable)
//	  フェイルオープン。resultをそのまま返す。nexway側の一時的な不調で
//	  本人確認処理を止めない(反社ヒットの見逃し防止)
//	上記以外(CLEARED/UNMATCH主張、または応答はあるが内容が信頼できない場合)
//	  フェイルクローズ。弾く
//
// CLEARED/UNMATCH主張のときは通信失敗でもフェイルオープンにしない。
// フェイルオープンにすると、nexwayが一時的に落ちている間、認証の無いこの
// エンドポイントに誰でも偽の結果を送るだけで、裏取り無しに会員を進めたり
// 落としたりできてしまう(疑わしきは弾く、をフェイルオープンの経路で
// 回避されてはならない)。見送られる副作用は「nexwayが落ちている間は
// CLEARED/UNMATCHの反映が遅れる」ことに留まり、DETECTEDへの遅延(反社ヒット
// の見逃し)とは非対称に許容できる。
//
// DETECTEDのフェイルオープンで通した場合、戻り値のverifiedにfalseを返す。
// 呼び出し元(ekycAfterProcess)はこれを見てIsEkyc等の扱いを変える(詳細は
// そちらのDETECTEDケース内のコメント参照)。「反社ヒットを見逃さない」
// という制約が要求するのはAuditStatusを進めることだけで、IsEkycまで立てる
// 理由にはならない。
//
// フェイルオープンは本物のnexwayの一時的な不調でも起こり得る前提で設計
// している。この関数に渡すctxは、呼び出し元(ekycAfterProcess)が受信
// リクエストのctxから独立させたものである必要がある(理由は呼び出し箇所の
// コメント参照)。nexway自体が遅い・落ちているという本物の不調による
// フェイルオープンは、ctxの独立化とは無関係に引き続き起こり得る。
func (a NexwayControllers) verifyAndReconcileNexwayResult(ctx context.Context, result NexwayEkycResult) (NexwayEkycResult, bool, error) {
	applicantId := strconv.FormatInt(result.ApplicantId, 10)
	output, err := a.nexwayService.GetEkycResultByApplicantId(ctx, applicantId)
	if err != nil {
		if errors.Is(err, service.ErrNexwayUnavailable) &&
			strings.EqualFold(result.OverallResult, string(domain.EKYC_OVERALL_RESULT_STATUS_DETECTED)) {
			logrus.WithFields(logrus.Fields{
				"applicant_id": result.ApplicantId,
				"reference_id": result.ReferenceId,
				"error":        err,
				"event":        "nexway_detected_fail_open",
			}).Warn("MIR-1223: nexwayへの通信に失敗。Webhookの値でフェイルオープンする(IsEkycは更新しない)")
			return result, false, nil
		}
		logrus.WithFields(logrus.Fields{
			"applicant_id": result.ApplicantId,
			"reference_id": result.ReferenceId,
			"error":        err,
		}).Error("MIR-1223: nexwayの照会に失敗。フェイルクローズで弾く")
		return result, false, errors.New("nexway webhook result rejected: lookup failed")
	}

	if output == nil {
		logrus.WithFields(logrus.Fields{
			"applicant_id": result.ApplicantId,
			"reference_id": result.ReferenceId,
		}).Error("MIR-1223: nexway照会結果がnil。フェイルクローズで弾く")
		return result, false, errors.New("nexway webhook result rejected: nil lookup response")
	}

	if output.ErrorCode != "" || output.ReferenceId == "" {
		logrus.WithFields(logrus.Fields{
			"applicant_id": result.ApplicantId,
			"error_code":   output.ErrorCode,
		}).Error("MIR-1223: nexway照会結果が不正(ErrorCodeあり、またはReferenceId空)。弾く")
		return result, false, errors.New("nexway webhook result rejected: invalid lookup response")
	}

	if output.ReferenceId != result.ReferenceId {
		logrus.WithFields(logrus.Fields{
			"applicant_id":          result.ApplicantId,
			"webhook_reference_id":  result.ReferenceId,
			"verified_reference_id": output.ReferenceId,
		}).Error("MIR-1223: nexway照会結果のReferenceIdがWebhookのReferenceIdと不一致。なりすましの可能性")
		return result, false, errors.New("nexway webhook result rejected: reference id mismatch")
	}

	riskResult := ""
	if output.RiskDetect != nil {
		riskResult = output.RiskDetect.Result
	}

	// output.Statusがcompleted以外(applying/accepted/processing、または
	// 表示されず空)の場合、Result/RiskDetectがまだ確定していない可能性が
	// あるため弾く(フェイルクローズ)。Webhookは本来nexway側の処理が完了した
	// ことの通知のはずで、それにもかかわらず照会結果が未完了を返すこと自体が
	// 矛盾。素通りにすると、攻撃者が「値」を操作できなくても「未完了の時間帯
	// (本番実測で夜間申請は数時間に及ぶ)を狙って送信する」だけで矛盾チェック
	// 全体を回避でき、実質的な迂回路になってしまう。弾けば500を返しnexway側の
	// 再送に載るため、タイムラグが実在する場合も自然に解消する。
	if !strings.EqualFold(output.Status, "completed") {
		logrus.WithFields(logrus.Fields{
			"applicant_id":    result.ApplicantId,
			"reference_id":    result.ReferenceId,
			"overall_result":  result.OverallResult,
			"verified_status": output.Status,
		}).Error("MIR-1223: nexway照会結果のStatusが完了前(または空)。フェイルクローズで弾く")
		return result, false, errors.New("nexway webhook result rejected: lookup status not completed")
	}

	switch strings.ToUpper(result.OverallResult) {
	case string(domain.EKYC_OVERALL_RESULT_STATUS_CLEARED):
		matched := strings.EqualFold(output.Result, "match")
		cleared := strings.EqualFold(riskResult, "cleared")
		if !matched || !cleared {
			logrus.WithFields(logrus.Fields{
				"applicant_id":         result.ApplicantId,
				"reference_id":         result.ReferenceId,
				"verified_result":      output.Result,
				"verified_risk_result": riskResult,
			}).Error("MIR-1223: WebhookはCLEAREDを主張しているが、nexway照会結果(本人確認/反社チェック)と矛盾する。弾く")
			return result, false, errors.New("nexway webhook result rejected: cleared claim contradicts lookup")
		}
	case string(domain.EKYC_OVERALL_RESULT_STATUS_UNMATCH):
		if !strings.EqualFold(output.Result, "unmatch") {
			logrus.WithFields(logrus.Fields{
				"applicant_id":    result.ApplicantId,
				"reference_id":    result.ReferenceId,
				"verified_result": output.Result,
			}).Error("MIR-1223: WebhookはUNMATCHを主張しているが、nexway照会結果(本人確認)と矛盾する。弾く")
			return result, false, errors.New("nexway webhook result rejected: unmatch claim contradicts lookup")
		}
	case string(domain.EKYC_OVERALL_RESULT_STATUS_DETECTED):
		if !strings.EqualFold(riskResult, "detected") {
			logrus.WithFields(logrus.Fields{
				"applicant_id":         result.ApplicantId,
				"reference_id":         result.ReferenceId,
				"verified_risk_result": riskResult,
			}).Error("MIR-1223: WebhookはDETECTEDを主張しているが、nexway照会結果(反社チェック)と矛盾する。弾く")
			return result, false, errors.New("nexway webhook result rejected: detected claim contradicts lookup")
		}
		// riskResultだけを見るこのチェックは本人確認結果(output.Result)を
		// 一切見ていない。ここでoutput.Resultが空(または不明)だからといって
		// 弾くと、本物の反社ヒット(riskResult=detected)がResult欠損という
		// 別要因だけで握りつぶされ、「反社ヒットを見逃さない」というMIR-1222
		// の設計意図に反する。空でも矛盾チェック自体は通し、Webhookの詐称値を
		// 消す(空にする)ことでIsEkycを立てさせない対応を下のResult補正で行う
		// (詳細はそちらのコメント参照)。
	default:
		// CLEARED/UNMATCH/DETECTED以外の値。nexway側の仕様追加等で今後
		// 現れる可能性があるが、未知だからといって弾くと本人確認処理が
		// 止まってしまうため、矛盾チェックは行わず素通りさせる。気付ける
		// ようログにだけ残す。
		logrus.WithFields(logrus.Fields{
			"applicant_id":   result.ApplicantId,
			"reference_id":   result.ReferenceId,
			"overall_result": result.OverallResult,
		}).Warn("MIR-1223: 未知のOverallResult。矛盾チェックをスキップして素通りする")
	}

	// DETECTED主張は、output.Resultが空(または不明)でも矛盾チェックを通す
	// (上のcase参照)。その代わりここで無条件にresult.Resultをoutput.Result
	// (空なら空)で上書きする。output.Resultが空になりうる原因(照会対象の
	// Statusが完了前など)があるかは未確認だが、上書きしておけばWebhookの
	// 詐称値(例えばmatch)がresult.Resultに残ることは無くなり、呼び出し元
	// (ekycAfterProcess)のresult.Result==matchというIsEkyc判定は自動的に
	// falseになる(2026-09-22のレビューで指摘された抜け穴の対応)。AuditStatus
	// を進める処理(反社ヒットの記録)自体はこの上書きと無関係に実行される。
	//
	// CLEARED/UNMATCH主張は、output.Resultの値そのもの(match/unmatch)を
	// 矛盾チェックで直接見ているため、ここに到達した時点で必ず非空。未知の
	// OverallResult値(default分岐、矛盾チェックをスキップして素通りさせる)
	// だけは空でありうる。この場合、DBへの反映はekycAfterProcess側で
	// ekycRequest.Result==""ならMustColsで"result"を対象にするため、空の
	// まま上書きしても前回の申請試行のResultが残ったままになることは無い
	// (DETECTEDケースと同じ理屈)。それでも空でないときだけ上書きするのは、
	// 未知の値は矛盾チェック自体をスキップしており(=Webhookの内容を検証
	// していない)、検証していない照会結果で下手に上書きするより、Webhookの
	// 値をそのまま残す判断に揃えるため。
	if strings.EqualFold(result.OverallResult, string(domain.EKYC_OVERALL_RESULT_STATUS_DETECTED)) {
		result.Result = output.Result
	} else if output.Result != "" {
		result.Result = output.Result
	}
	return result, true, nil
}

// verifyWithNexwayは、Webhookの内容をnexwayの照会APIで裏取りするかどうかを
// 制御する(MIR-1223)。CheckNexwayResult(本番のWebhook受信口)からはtrueで
// 呼び、CallTestWebhook(STG環境限定の手動テスト用エンドポイント)からは
// falseで呼ぶ。後者は既にserviceutils.EnvIsProd()で本番実行をガードして
// おり、自分自身が組み立てたダミーのApplicantId/ReferenceIdを使うため、
// なりすましの懸念が無いうえ、実在しない組み合わせだと検証で弾かれて
// テストが成立しなくなる。
func (a NexwayControllers) ekycAfterProcess(ctx context.Context, result NexwayEkycResult, verifyWithNexway bool) error {
	var err error
	var mail domain.Mail

	// eKYCリクエスト記録照会
	ekycRequest, err := a.privateUserService.GetEkycRequestLog(ctx, result.ReferenceId)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"message": "Failed to get latest eKYC request log",
			"error":   err,
			"time":    time.Now().String(),
		}).Error("eKYCAfterProcess")

		return err
	}

	// MIR-1223: 自社DBに存在しないReferenceId(ekycRequest.Id==0)では照会を
	// 発生させない。存在しないIDを送りつけてnexwayへの発信を誘発できない
	// ようにするため。nexwayServiceが未設定(一部のテスト)の場合もスキップする。
	//
	// verifiedはnexwayの裏付けが取れたかどうかを表す。検証自体を行わなかった
	// 場合(verifyWithNexway=false、対象外)はtrueのまま扱う。CallTestWebhook
	// (STG限定)は自分自身が組み立てたダミー値でなりすましの懸念が無いため。
	// verifyAndReconcileNexwayResultがfalseを返すのは、DETECTED主張が通信
	// 失敗でフェイルオープンした場合のみで、この場合はIsEkycを更新しない
	// (DETECTEDケース内で参照する)。
	verified := true
	if verifyWithNexway && a.nexwayService != nil && ekycRequest.Id != 0 {
		var verifiedResult NexwayEkycResult
		var verifyErr error
		// nexwayへの照会はctx(Webhookの受信リクエストctx)から独立させる。
		// このエンドポイントは無認証で、送信元(攻撃者になりうる)がTCP接続を
		// 握っている。受信ctxをそのまま照会に使うと、攻撃者がPOST直後に
		// 接続を切ることでnet/httpがctxをキャンセルし、DETECTED主張の
		// フェイルオープン(照会が通信失敗ErrNexwayUnavailableで終わった
		// ときの経路)を意図的に誘発できてしまう。context.Background()から
		// 派生した別ctxにすれば、攻撃者が自分の接続を切ってもサーバー側の
		// 照会処理は影響を受けず、この経路自体が構造的に成立しなくなる。
		lookupCtx, cancel := context.WithTimeout(context.Background(), service.NexwayApiTimeout)
		defer cancel()
		verifiedResult, verified, verifyErr = a.verifyAndReconcileNexwayResult(lookupCtx, result)
		if verifyErr != nil {
			logrus.WithFields(logrus.Fields{
				"message": "Rejected nexway webhook result",
				"error":   verifyErr,
				"time":    time.Now().String(),
			}).Error("eKYCAfterProcess")

			return verifyErr
		}
		result = verifiedResult
	}

	// MIR-1223: ApplicantId/Resultへの反映はverified==trueのときだけ行い、
	// verified=false(DETECTEDのフェイルオープン)のときは検証未了である
	// ことを表すため明示的にクリアする(ekycRequestUpdateColsで対象カラムを
	// 指定し、後段のekycRequest.Update(ctx, ...)でDBへ反映する)。
	//
	// クリアするのは、eKYCリクエスト行がStatus(NONE/SUCCESS/FAIL)がNONEの
	// ままなら再申請のたびに使い回される設計になっており(private_user_ucase.
	// go)、Webhook処理(このファイル)はこのStatusを一切更新しないため。1つの
	// 行が複数回の申請試行(それぞれ別のApplicantId)を跨いで使い回され得る
	// ので、検証未了時に「DBの既存値を維持する」実装にすると、それは「今回の
	// Webhookと一致する正しい値」ではなく「前回の申請試行の、今回とは無関係
	// かもしれない古い値」になってしまう。例えば1回目の申請がCLEARED
	// (Result=match)で正常に記録された後、2回目の申請で本物のDETECTED
	// (反社ヒット)が来たがnexwayの一時的な不調でフェイルオープンした場合、
	// AuditStatusは正しくWAIT_ANTISOCIAL_AUDITへ進む一方、維持していると
	// Resultは1回目の"match"のまま残り、管理画面で「反社チェック待ちなのに
	// 本人確認は済んでいる」という矛盾した表示になって、審査担当者の誤承認を
	// 招きかねない。
	//
	// クリア対象はApplicantId/Resultに加え、「eKYC詳細結果取得」ブロックが
	// 通常設定するRiskDetectResult等の詳細フィールド(このブロックはverified
	// =false時にスキップされるため、クリアしなければ前回試行の値が残った
	// ままになる)と、ErrorMessage(UNMATCHケースでのみ設定される本人確認
	// エラーの詳細文言。Resultだけクリアしてこちらを残すと、「結果は空なのに
	// エラー理由だけ表示される」という同種の矛盾が生まれる)を含む、計9項目。
	// EkycRequestLog構造体の全フィールドを1つずつ棚卸しして洗い出した対象
	// で、それ以外(Id/Email/UserId/EkycId/Status/Created/ResultResponse)は
	// 検証結果と無関係、または既存の設計上ここでは変更しない値のため対象外
	// (棚卸しの詳細はPRの説明を参照)。
	var ekycRequestUpdateCols []string
	var ekycRequestMustCols []string
	if verified {
		ekycRequest.ApplicantId = result.ApplicantId
		ekycRequest.Result = domain.EKYC_RESULT_STATUS(strings.ToUpper(result.Result))
		if ekycRequest.Result == "" {
			// DETECTED主張でnexwayの照会結果(output.Result)が空(本人確認結果
			// 不明)の場合、ここでResultはゼロ値になる。ゼロ値カラムはxormの
			// デフォルトどおりUPDATE文から除外されるため(EkycRequestLog.Update
			// のコメント参照)、指定しないとDBには前回の申請試行のResult
			// (例えばMATCH)が残ったままになり、上のコメントが警告する「反社
			// チェック待ちなのに本人確認済みと表示される」矛盾表示がこの経路
			// でも起きる(2026-09-24のレビューで指摘)。Colsで絞り込むと
			// ApplicantId等ほかの非ゼロ値カラムまでUPDATE対象から外れてしまう
			// ため(実測で確認済み、Colsは「指定したカラム以外を除外する」、
			// MustColsは「指定したカラムだけゼロ値でも強制的に含める」という
			// 別の仕組み)、MustColsで"result"だけを対象にする。
			ekycRequestMustCols = []string{"result"}
		}
	} else {
		ekycRequest.ApplicantId = 0
		ekycRequest.Result = ""
		ekycRequest.ErrorMessage = nil
		ekycRequest.RiskDetectResult = nil
		ekycRequest.RiskDetectDetectedAt = nil
		ekycRequest.EkycFrontMatch = nil
		ekycRequest.EkycThicknessMatch = nil
		ekycRequest.EkycLiveMatch = nil
		ekycRequest.EkycEndAt = nil
		ekycRequestUpdateCols = []string{
			"applicant_id", "result", "error_message",
			"risk_detect_result", "risk_detect_detected_at",
			"ekyc_front_match", "ekyc_thickness_match", "ekyc_live_match", "ekyc_end_at",
		}
	}

	// ユーザー情報照会
	// MIR-1374: 以前はekycRequest.Emailで照会しており、eKYC申込み後にメール
	// アドレスを変えた会員が見つからずUser not foundで失敗していた
	// (ネクスウェイ側は失敗とみなし再送を繰り返すが、何度送っても同じ理由で
	// 失敗し続ける)。ekyc_request_logにはuser_idが保存されているため、
	// そちらで照会する。
	privateUser, err := domain.PrivateUser{}.GetById(ctx, ekycRequest.UserId)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"message":       "Failed to get user by id",
			"error":         err,
			"userId":        ekycRequest.UserId,
			"ekycRequestId": ekycRequest.Id,
			"time":          time.Now().String(),
		}).Error("eKYCAfterProcess")

		return err
	}
	if privateUser == nil {
		// 会員IDでも見つからない(退会・削除等)場合は、エラーにしてネクスウェイに
		// 再送させても状況が変わらない。200を返して送り直しを止める(MIR-1373/1374、
		// 設計v1の決定事項)。識別子(ekyc_request_logのID・user_id)だけをログへ残し、
		// 氏名・メールは出さない。
		//
		// ekyc_request_log自体が自社に存在しないReferenceId(ekycRequest.Id==0・
		// UserId==0)が届いた場合もこの分岐に入る。この場合userId・ekycRequestIdは
		// 0のままで追えないため、ネクスウェイから届いたReferenceId・ApplicantId
		// (個人情報ではない)もログへ残す(MIR-1223でWebhookの認証・署名検証が
		// 入っているため、外部からの偽の通知で200が返ること自体の懸念は小さい)。
		logrus.WithFields(logrus.Fields{
			"message":       "User not found",
			"userId":        ekycRequest.UserId,
			"ekycRequestId": ekycRequest.Id,
			"referenceId":   result.ReferenceId,
			"applicantId":   result.ApplicantId,
			"time":          time.Now().String(),
		}).Error("eKYCAfterProcess")

		return nil
	}

	// ユーザー情報記録用 - before
	var before *service.UserInfoHistoryData
	before, err = a.userInfoHistoryService.GetDataForPrivateUser(ctx, privateUser.User.Id)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"message": "Failed to get user info history data",
			"error":   err,
			"time":    time.Now().String(),
		}).Error("eKYCAfterProcess")

		return err
	}

	// eKYC通過確認
	switch domain.EKYC_OVERALL_RESULT_STATUS(strings.ToUpper(result.OverallResult)) {
	case domain.EKYC_OVERALL_RESULT_STATUS_CLEARED:
		// 本人認証成功
		isApprovedMyNumber, err := domain.MyNumber{}.CheckMyNumberStatusByUserId(ctx, privateUser.User.Id)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"message": "Failed to get my number status by user id",
				"error":   err,
				"time":    time.Now().String(),
			}).Error("eKYCAfterProcess")

			return err
		}

		// ステータス変更。IsEkyc・SignupStatusは本人確認そのものの新しい結果なので
		// 常に更新するが、AuditStatusは次の2つの会員には進めない(MIR-1276で
		// shouldAdvanceOnClearedへ統合、理由は同関数のコメント参照)。
		//   - 既に本登録承認完了(COMPLETED_AUDIT)または無効(INVALIDITY_AUDIT)に
		//     なっている会員(MIR-1222。Webhookの再送により承認済みの会員が
		//     WAIT_FINAL_APPROVAL_AUDITへ巻き戻る事象が本番で35件発生していた)
		//   - 反社確認中・反社確認待ち(CHECK_ANTISOCIAL_AUDIT・
		//     WAIT_ANTISOCIAL_AUDIT)の会員(MIR-1276。会員が自分でeKYCの
		//     Webhookへclearedを送るだけで、反社チェックの結果を待たずに
		//     最終承認待機まで進められてしまっていた)
		// cleared自体は「進める」方向の結果なので、進めない会員に対して黙って
		// 据え置いても実害は無い。IsEkyc・SignupStatusは本人確認結果の記録として
		// 更新する(本人確認そのものは実際に通っているため)。
		privateUser.IsEkyc = true
		privateUser.SignupStatus = domain.SIGNUP_COMPLETE
		if isApprovedMyNumber {
			if shouldAdvanceOnCleared(privateUser.AuditStatus) {
				privateUser.AuditStatus = domain.WAIT_FINAL_APPROVAL_AUDIT
			} else {
				logAuditStatusGuardEvent(privateUser, result, ekycRequest.Created, "Skip audit status advance: finalized or antisocial check in progress")
			}
		}
		// TODO: マイナンバー登録復活時にもとに戻す elseで分岐させる
		// マイナンバー登録復活させるまではisApprovedMyNumberはtrue固定
		// privateUser.AuditStatus = domain.CHECK_MY_NUMBER_AUDIT

	case domain.EKYC_OVERALL_RESULT_STATUS_DETECTED:
		// 反社チェックNG。本人確認結果(match/unmatch)は問わずAuditStatusを
		// 進める(IsEkycを立てるかどうかは本人確認結果で分ける、詳細は下記
		// コメント参照)。COMPLETED_AUDITの会員には、CLEARED・
		// UNMATCHとは異なりガードを入れない(MIR-1222)。反社ヒットは会員の状態に
		// 関わらず必ず拾うべき情報で、「確定済みだから」という理由で握りつぶすと、
		// 管理者が別経路(反社チェック承認)で先にCOMPLETED_AUDITへ進めた直後に
		// 届いた本物の初回検知まで一緒に捨ててしまう。CLEARED・UNMATCHの再送で
		// 確定済み会員を巻き戻す実害(本番で継続発生)より、反社ヒットを見逃す
		// リスクを避けることを優先する。判断のためログだけは残す(動作は変えない)。
		//
		// 一方、INVALIDITY_AUDIT(既に反社で無効化済み)の会員はここでガードする。
		// user-api側の複数箇所(controllers/user.go・usecase/user・
		// usecase/user_verify)がAuditStatus==INVALIDITY_AUDITそのものをログイン・
		// 本人確認のブロック条件にしているため、ここでWAIT_ANTISOCIAL_AUDITへ
		// 進めてしまうとその専用ブロックが外れてしまう。COMPLETED_AUDITとは逆に、
		// 無効化状態から動かさないことが安全側になる。
		if privateUser.AuditStatus == domain.INVALIDITY_AUDIT {
			logAuditStatusGuardEvent(privateUser, result, ekycRequest.Created, "Skip audit status downgrade: already finalized")
		} else {
			if privateUser.AuditStatus == domain.COMPLETED_AUDIT {
				logAuditStatusGuardEvent(privateUser, result, ekycRequest.Created, "Audit status downgraded despite already finalized (DETECTED is not guarded for COMPLETED_AUDIT)")
			}
			privateUser.AuditStatus = domain.WAIT_ANTISOCIAL_AUDIT
		}
		privateUser.SignupStatus = domain.SIGNUP_COMPLETE
		// MIR-1223: IsEkyc(本人確認済みの証明。マイナンバー承認画面
		// apps/admin-api/controllers/my_number.goが会員の現在の審査状態を
		// 見ずにこれだけを見てCOMPLETED_AUDITへ進める)は、verified(nexwayの
		// 裏付けが取れたか)に加えて、result.Result(照会成功時はoutput.Result
		// に補正済み)がmatch(本人確認OK)であることも要求する。
		// verifyAndReconcileNexwayResultのDETECTED矛盾チェックはRiskDetect.
		// Result==detectedしか見ておらず本人確認結果自体は見ていないため、
		// 本人確認に失敗した(実際はunmatch)人物が自分の正当なApplicantId/
		// ReferenceIdで偽のDETECTEDを送っても、反社チェックの実体がdetected
		// でありさえすればverified=trueで通過してしまう(レビュー指摘、
		// 2026-09-22)。「反社ヒットを見逃さない」という制約(MIR-1222)が
		// 要求するのはAuditStatusを進めることだけなので、IsEkycの要件を
		// 絞っても反社ヒットの記録(AuditStatus/SignupStatus)は本人確認結果
		// に関わらず残る。
		//
		// フェイルオープン(verified=false)時にIsEkycを更新しない理由は、
		// verifyAndReconcileNexwayResultの「DETECTEDのフェイルオープンで
		// 通した場合」のコメント(「反社ヒットを見逃さない」制約が要求
		// するのはAuditStatusを進めることだけで、IsEkycまで立てる理由には
		// ならない、という判断)を参照。
		//
		// この防御(result.Result==matchの要求)が効くのは「今回のWebhook
		// 呼び出しでIsEkycを新たにtrueにするかどうか」の判断までで、過去に
		// 別の正当な経路(CLEARED等)で既にIsEkyc=trueになっている会員の値を、
		// 今回のDETECTED主張に応じてfalseへ戻すものではない(このcaseに
		// IsEkycをfalseにする分岐は無い)。
		if verified && strings.EqualFold(result.Result, "match") {
			privateUser.IsEkyc = true
		} else if verified {
			logrus.WithFields(logrus.Fields{
				"user_id":         privateUser.User.Id,
				"verified_result": result.Result,
				"event":           "nexway_detected_result_mismatch_isekyc_withheld",
			}).Warn("eKYCAfterProcess: DETECTEDだが照会結果のResultがmatchでないためIsEkycは更新しない")
		} else {
			logrus.WithFields(logrus.Fields{
				"user_id": privateUser.User.Id,
				"event":   "nexway_detected_fail_open_isekyc_withheld",
			}).Warn("eKYCAfterProcess: DETECTEDがフェイルオープンで通ったためIsEkycは更新しない")
		}
	case domain.EKYC_OVERALL_RESULT_STATUS_UNMATCH:
		// 本人認証失敗
		errorMessages := []string{}
		for _, v := range result.VerificationErrors {
			errorMessages = append(errorMessages, v.ErrorMessage)
		}
		for _, v := range result.OcrValidationErrors {
			errorMessages = append(errorMessages, v.ErrorMessage)
		}
		var errorMsg string
		if len(errorMessages) > 0 {
			errorMsg = strings.Join(errorMessages, "\n") + "\n"
			ekycRequest.ErrorMessage = &errorMsg
		}

		logrus.WithFields(logrus.Fields{
			"message":    "Failed to eKYC",
			"ekycResult": result,
			"time":       time.Now().String(),
		}).Info("eKYCAfterProcess")

		// 既に確定済み(COMPLETED_AUDIT・INVALIDITY_AUDIT)の会員は、Webhookの再送に
		// よるunmatchで本登録保留へ落とさない。落とすと登録状態がSIGNUP_NONEへ
		// リセットされ、「本人確認に失敗しました」という再本登録メールまで
		// 送られてしまう(MIR-1222)。ekycRequestへのエラーメッセージ記録は
		// eKYCリクエスト自体の結果なので、確定済みかどうかに関わらず行う。
		if isFinalizedAuditStatus(privateUser.AuditStatus) {
			logAuditStatusGuardEvent(privateUser, result, ekycRequest.Created, "Skip audit status downgrade: already finalized")
		} else {
			// HOLD_AUDITステータスに変更
			id, err := uuid.NewV4()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"message": "Failed to create hold user check code",
					"error":   err,
					"time":    time.Now().String(),
				}).Error("eKYCAfterProcess")

				return err
			}
			holdUserCheckCode := id.String()

			privateUser.User.AuditStatus = domain.HOLD_AUDIT
			privateUser.User.HoldUserCheckCode = &holdUserCheckCode
			privateUser.User.SignupStatus = domain.SIGNUP_NONE

			// 再本登録メール送信
			if err := mail.RetrySignUpEmail(ctx, serviceutils.GetServiceInfoMail(), privateUser.Email, privateUser.Name, errorMessages); err != nil {
				logrus.WithFields(logrus.Fields{
					"message": "Failed to send mail - RetrySignUpEmail",
					"error":   err,
					"time":    time.Now().String(),
				}).Error("eKYCAfterProcess")
				return err
			}
		}
	}

	// ユーザー情報更新。is_ekyc・signup_status・audit_status・hold_user_check_code
	// の4項目だけを書く(UpdateEkycResult)。以前はUpdate(ctx)でuser・private_profileの
	// 行を丸ごと書き戻しており、この処理と同じ秒に会員が別の操作(ログイン・本人情報の
	// 変更等)を行うと、互いの結果を消し合っていた(MIR-1373/1374)。
	if err := privateUser.UpdateEkycResult(ctx); err != nil {
		logrus.WithFields(logrus.Fields{
			"message": "Failed to update user",
			"error":   err,
			"time":    time.Now().String(),
		}).Error("eKYCAfterProcess")

		return err
	}

	// eKYC詳細結果取得。verified=false(DETECTEDのフェイルオープン)のときは
	// このブロックを丸ごとスキップする(MIR-1223)。verified=false時、
	// ekycRequest.ApplicantIdは上記で既に0へクリア済みで、この呼び出しを
	// しても意味が無い。裏付けが取れていない以上、この照会結果をekycRequest
	// へ保存してよい理由が無いためで、仮にスキップせず呼び出した場合、同じ
	// 通信障害で失敗すると、AuditStatus/SignupStatusは既にprivateUser.
	// Update(ctx)で確定済みなのに処理全体がエラー終了し、ekycRequestが更新
	// されないまま宙に浮いた状態が残る(この状態からの回復手段が無い点は
	// MIR-1223のスコープ外として別チケットで追跡する)。スキップの結果、
	// この場合のみApplicantImages・EkycResult・RiskDetectResult等の詳細
	// フィールドは保存されないが、nexwayの裏付けが取れていない以上、保存
	// すべき情報自体が無い。
	if verified && a.nexwayService != nil && ekycRequest.ApplicantId != 0 {
		// この呼び出しもctxから独立させる。理由は上記(verifyAndReconcileNexway
		// Result呼び出し前)のコメントと同じ。
		detailLookupCtx, cancel := context.WithTimeout(context.Background(), service.NexwayApiTimeout)
		defer cancel()
		ekycResult, svcErr := a.nexwayService.GetEkycResultByApplicantId(detailLookupCtx, strconv.FormatInt(ekycRequest.ApplicantId, 10))
		if svcErr != nil {
			logrus.WithFields(logrus.Fields{
				"message": "Failed call nexway service",
				"error":   svcErr,
				"time":    time.Now().String(),
			}).Error("eKYCAfterProcess")
			return svcErr
		}
		if ekycResult != nil && ekycResult.ErrorCode != "" {
			logrus.WithFields(logrus.Fields{
				"message":       "Nexway API returned error code",
				"error_code":    ekycResult.ErrorCode,
				"error_message": ekycResult.Message,
				"time":          time.Now().String(),
			}).Error("eKYCAfterProcess")
			return errors.New("Nexway API returned error code: " + ekycResult.ErrorCode)
		}

		// ApplicantImagesをS3に保存
		if ekycResult != nil && len(ekycResult.ApplicantImages) > 0 {
			if err := service.SaveApplicantImages(ctx, ekycRequest.Id, ekycRequest.ApplicantId, ekycResult.ApplicantImages); err != nil {
				logrus.WithFields(logrus.Fields{
					"message": "Failed to save applicant images",
					"error":   err,
					"time":    time.Now().String(),
				}).Error("eKYCAfterProcess")
				return err
			}
		}

		// eKYC照合結果をセット
		if ekycResult != nil && ekycResult.EkycResult != nil {
			ekycRequest.EkycFrontMatch = &ekycResult.EkycResult.FrontMatch
			ekycRequest.EkycThicknessMatch = &ekycResult.EkycResult.ThicknessMatch
			ekycRequest.EkycLiveMatch = &ekycResult.EkycResult.LiveMatch
			if ekycResult.EkycResult.EkycEndAt != "" {
				if ekycEndAt, parseErr := time.Parse(time.RFC3339, ekycResult.EkycResult.EkycEndAt); parseErr == nil {
					ekycEndAtJSON := jsonutils.JSONTime(ekycEndAt)
					ekycRequest.EkycEndAt = &ekycEndAtJSON
				} else {
					logrus.WithFields(logrus.Fields{
						"message": "Failed to parse ekyc_end_at",
						"error":   parseErr,
						"value":   ekycResult.EkycResult.EkycEndAt,
					}).Warn("eKYCAfterProcess")
				}
			}
		}

		// リスク情報検索結果をセット
		if ekycResult != nil && ekycResult.RiskDetect != nil {
			ekycRequest.RiskDetectResult = &ekycResult.RiskDetect.Result
			if ekycResult.RiskDetect.DetectedAt != "" {
				if detectedAt, parseErr := time.Parse(time.RFC3339, ekycResult.RiskDetect.DetectedAt); parseErr == nil {
					detectedAtJSON := jsonutils.JSONTime(detectedAt)
					ekycRequest.RiskDetectDetectedAt = &detectedAtJSON
				} else {
					logrus.WithFields(logrus.Fields{
						"message": "Failed to parse risk_detect_detected_at",
						"error":   parseErr,
						"value":   ekycResult.RiskDetect.DetectedAt,
					}).Warn("eKYCAfterProcess")
				}
			}
		}

		// オペレーター情報をDBに保存
		if ekycResult != nil && len(ekycResult.Operators) > 0 {
			if err := service.SaveOperators(ctx, ekycRequest.Id, ekycResult.Operators); err != nil {
				logrus.WithFields(logrus.Fields{
					"message": "Failed to save operators",
					"error":   err,
					"time":    time.Now().String(),
				}).Error("eKYCAfterProcess")
				return err
			}
		}
	} else if !verified {
		// MIR-1223: フェイルオープン後にこのブロックがスキップされたことを、
		// 識別子付きで残す(値そのものは出さない)。この処理は正常終了する
		// ため(エラー終了してnexwayの再送に載ることが無い)、この会員が
		// AuditStatus=WAIT_ANTISOCIAL_AUDITのまま詳細結果が永久に埋まらない
		// 状態で確定したことを、後からこのログでしか洗い出せない。
		// applicant_idはekycRequest.ApplicantId(このログより前に0へクリア
		// 済み)ではなくresult.ApplicantId(Webhookの申告値)を使う。前者を
		// 使うとログが常に0になり、後から突き合わせる手がかりにならない。
		logrus.WithFields(logrus.Fields{
			"ekyc_request_id": ekycRequest.Id,
			"applicant_id":    result.ApplicantId,
			"reference_id":    ekycRequest.EkycId,
			"user_id":         privateUser.User.Id,
			"event":           "nexway_detected_fail_open_detail_skipped",
		}).Warn("eKYCAfterProcess: フェイルオープンのため詳細結果取得をスキップした")
	}

	// ekycRequestMustColsとekycRequestUpdateColsは同時に使わない(前者は
	// verified=trueかつResult空、後者はverified=falseのときだけセットされる)。
	var updateErr error
	if len(ekycRequestMustCols) > 0 {
		_, updateErr = ekycRequest.UpdateWithMustCols(ctx, ekycRequestMustCols...)
	} else {
		_, updateErr = ekycRequest.Update(ctx, ekycRequestUpdateCols...)
	}
	if updateErr != nil {
		logrus.WithFields(logrus.Fields{
			"message": "Failed to update ekyc request log",
			"error":   updateErr,
			"time":    time.Now().String(),
		}).Error("eKYCAfterProcess")

		return updateErr
	}

	// ユーザー情報記録用 - after
	var after *service.UserInfoHistoryData
	after, err = a.userInfoHistoryService.GetDataForPrivateUser(ctx, privateUser.User.Id)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"message": "Failed to get user info history data",
			"error":   err,
			"time":    time.Now().String(),
		}).Error("eKYCAfterProcess")

		return err
	}

	// ユーザー情報の変更内容記録
	a.userInfoHistoryService.CreateUserInfoHistory(ctx, *before, *after, nexwayProcessId, nexwayProcessName, "")

	return nil
}
