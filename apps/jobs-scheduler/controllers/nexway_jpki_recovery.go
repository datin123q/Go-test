package controllers

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

// 署名の結果の通知(Webhook)を取りこぼした記録を拾い直す定期ジョブと、署名が完了しなかった会員への
// 本人確認のやり直しの通知(MIR-1470)。
//
// 検証結果通知(Webhook)は、ネクスウェイが15分間隔で2回までしか再送せず(仕様書 p.17)、サーバーの停止や
// 障害で約30分以上受け取れないと、通知は届かなかったことになる。その記録は、署名用URLを発行した
// ままの状態(未確定・BPO未申請)で残り、会員は先に進めない。このジョブは、そうした記録を定期的に探し、
// 通知を受けたときと同じ処理(検証結果取得APIで確かめて、BPOへの申請または失敗の記録)を行う。

const (
	// jpkiRecoveryMinAge は、記録の作成からこの時間が経った行だけを対象にする。通知の再送(15分間隔で
	// 2回)が尽きるまでの約30分に、余裕を持たせた値。通知が届く見込みのある行を、先回りして処理しない。
	jpkiRecoveryMinAge = time.Hour
	// jpkiRecoveryMaxAge は、記録の作成からこの期間を過ぎた行は対象にしない。ネクスウェイの結果の
	// 保管期間は契約で1〜60日(既定15日。手引き p.6)で、過ぎると検証結果取得APIで引けない。
	jpkiRecoveryMaxAge = 14 * 24 * time.Hour
	// jpkiRecoveryBatchSize は、1回に処理する件数の上限。検証結果取得APIの処理の上限は
	// 1分に30件(仕様書 p.4)で、そのうち余裕を残す。残りは次の回で処理する。
	jpkiRecoveryBatchSize = 25
)

// RecoverStuckSignatureResults は、署名の結果が確定せず、BPOへも申請されないまま止まっている記録を
// 拾い直す定期ジョブ。実行間隔は設定(cronの JpkiSignatureRecoveryJob)で決める。1時間に1回を想定している。
func (c NexwayJpkiControllers) RecoverStuckSignatureResults(ctx context.Context) error {
	return c.recoverStuckSignatureResults(ctx, time.Now())
}

func (c NexwayJpkiControllers) recoverStuckSignatureResults(ctx context.Context, now time.Time) error {
	logs, err := domain.FindStuckJpkiRequestLogs(ctx,
		now.Add(-jpkiRecoveryMinAge), now.Add(-jpkiRecoveryMaxAge), jpkiRecoveryBatchSize)
	if err != nil {
		return errors.Wrap(err, "failed to find stuck jpki eKYC request logs")
	}
	if len(logs) == 0 {
		return nil
	}

	var pending, errored int
	for _, ekycLog := range logs {
		if ekycLog.RequestId == nil {
			continue
		}
		// 通知と同じ形の request を会員の記録から作り、通知を受けたときと同じ処理を呼ぶ。
		// Resultは、通知が言っている結果で、ジョブには無い。空にして、ネクスウェイの結果だけで判断する。
		request := NexwayJpkiWebhookRequest{ReferenceId: ekycLog.EkycId, RequestId: *ekycLog.RequestId}

		err := c.handleSignatureResult(ctx, request, signatureResultFromJob)
		switch {
		case err == nil:
			// 処理した(BPOへ申請した・失敗にした)、または対象外として飛ばした。
		case errors.Is(err, errJpkiSignatureNotFinal):
			// ネクスウェイ側でまだ確定していない。次の回で確認し直す。異常ではない。
			pending++
		default:
			errored++
			logrus.WithFields(nexwayJpkiLogFields(request)).WithError(err).
				Warn("nexway jpki signature recovery: failed to process, will retry in the next run")
		}
	}

	logrus.WithFields(logrus.Fields{
		"found":   len(logs),
		"pending": pending,
		"errored": errored,
		"time":    time.Now().String(),
	}).Info("nexway jpki signature recovery: finished")

	// 1件でも処理できていれば、ジョブとしては動いている。全件が失敗するのは、設定の不備・
	// ネクスウェイの障害など、全体に及ぶ問題の可能性が高いため、エラーにして気付けるようにする。
	if errored > 0 && errored == len(logs) {
		return errors.Errorf("all %d stuck jpki eKYC request logs failed to process", len(logs))
	}
	return nil
}

// jpkiRetryNotifier は、署名が完了しなかった会員へ、本人確認のやり直しをお願いするメールを送る。
// テストで差し替えられるようにしてある。
type jpkiRetryNotifier interface {
	NotifyRetryVerify(ctx context.Context, userId int64, reasons []string) error
}

type mailJpkiRetryNotifier struct {
	send   func(ctx context.Context, sender, email, name string, reasons []string) error
	sender func() string
}

func newMailJpkiRetryNotifier() jpkiRetryNotifier {
	return mailJpkiRetryNotifier{
		send:   domain.Mail{}.JpkiRetryVerifyEmail,
		sender: serviceutils.GetServiceInfoMail,
	}
}

// NotifyRetryVerify は、会員へメールを送る。送る必要が無い会員(本人確認が済んでいる・本登録が確定している)
// には、何もせず、エラーにもしない。
func (n mailJpkiRetryNotifier) NotifyRetryVerify(ctx context.Context, userId int64, reasons []string) error {
	privateUser, err := domain.PrivateUser{}.GetById(ctx, userId)
	if err != nil {
		return errors.Wrap(err, "failed to get private user")
	}
	if privateUser == nil {
		return nil
	}

	// 本人確認が既に済んでいる会員(ホ方式などで通過済み)や、本登録が確定した会員(承認完了・無効)へ
	// 「もう一度本人確認をお願いします」と送ると、誤った案内になる。
	if privateUser.IsEkyc || isFinalizedAuditStatus(privateUser.AuditStatus) {
		logrus.WithFields(logrus.Fields{
			"user_id":      userId,
			"audit_status": privateUser.AuditStatus,
			"is_ekyc":      privateUser.IsEkyc,
		}).Info("nexway jpki signature result: retry mail skipped (identity already verified or registration already decided)")
		return nil
	}
	if strings.TrimSpace(privateUser.Email) == "" {
		return nil
	}

	return n.send(ctx, n.sender(), privateUser.Email, privateUser.Name, reasons)
}
