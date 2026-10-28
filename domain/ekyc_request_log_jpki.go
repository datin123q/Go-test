package domain

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// MarkFailedIfNone は、eKYCリクエスト記録を「失敗」にする(MIR-1470)。まだ結果が確定していない
// (Status=NONE)ときだけ書き換え、書き換えた行数を返す。
//
// 条件をWHEREに含めるのは、Webhookの再送や画面からの完了報告(PUT)と同時に走っても、
// 先に確定した結果(SUCCESS・FAIL)を後から上書きしないため。読んでから書くと、その間に
// 変わった状態を踏み潰す。
func (c *EkycRequestLog) MarkFailedIfNone(ctx context.Context) (int64, error) {
	return factory.DB(ctx).
		Where("id = ? AND status = ?", c.Id, EKYC_REQUEST_STATUS_NONE).
		Cols("status").
		Update(&EkycRequestLog{Status: EKYC_REQUEST_STATUS_FAIL})
}

// SetApplicantIdIfUnset は、BPOから得た申請IDを記録する(MIR-1470)。まだ申請IDが無い
// (ApplicantId=0)ときだけ書き、書いた行数を返す。
//
// 申請IDはBPOの結果通知(Webhook)でも書かれる。通知のほうが先に届いて確定した値を、
// こちらが後から上書きしないよう、条件をWHEREに含める。
//
// 結果の状態(Status)は条件にしない。会員の画面からの完了報告(PUT)で行がSUCCESSになっていても、
// 申請IDが無ければ記録する(状態を条件にすると、完了報告が先に届いたときに、申請したのに申請IDが
// 残らない)。署名前の行を、新しい発行で失敗にした直後に申請が済んだ場合も、申請IDは残す。
func (c *EkycRequestLog) SetApplicantIdIfUnset(ctx context.Context, applicantId int64) (int64, error) {
	return factory.DB(ctx).
		Where("id = ? AND applicant_id = 0", c.Id).
		Cols("applicant_id").
		Update(&EkycRequestLog{ApplicantId: applicantId})
}

// HasNewerJpkiRequest は、同じ会員に、この行より新しい JPKI方式の行(署名用URLを発行した行)があるかを
// 返す(MIR-1470)。署名用URLの発行し直しは、会員が JPKI方式をやり直したことを表し、user-api は発行の
// 時点で古い行を失敗にする。JPKI方式の署名の結果では、申し込まれた行(HasNewerAppliedRequest)に加えて、
// これも「置き換えられた」とみなす(古い行の失敗の記録が漏れた場合でも、URLを2回発行して両方で署名した
// 会員を、BPOへ2回申請しないため)。ホ方式の行は数えない(画面を開いただけで作られるため)。
func (c EkycRequestLog) HasNewerJpkiRequest(ctx context.Context) (bool, error) {
	newer, err := factory.DB(ctx).
		Where("user_id = ? AND id > ?", c.UserId, c.Id).
		And("request_id IS NOT NULL AND request_id <> ''").
		Count(&EkycRequestLog{})
	if err != nil {
		return false, err
	}
	return newer > 0, nil
}

// MarkFailedIfNoneWithMessage は、MarkFailedIfNone と同じく、結果が確定していない(Status=NONE)行だけを
// 「失敗」にし、あわせてエラーメッセージ(失敗の理由。署名の検証処理のエラーコードなど)を記録する
// (MIR-1470)。書き換えた行数を返す。メッセージが空なら、MarkFailedIfNone と同じ(メッセージには触れない)。
//
// 状態とメッセージを同じ1回の条件付きUPDATEで書くのは、先に確定した結果(SUCCESS・FAIL)に、
// 後からメッセージだけが残る、またはメッセージの無い失敗になる、といった食い違いを作らないため。
func (c *EkycRequestLog) MarkFailedIfNoneWithMessage(ctx context.Context, message string) (int64, error) {
	if message == "" {
		return c.MarkFailedIfNone(ctx)
	}
	return factory.DB(ctx).
		Where("id = ? AND status = ?", c.Id, EKYC_REQUEST_STATUS_NONE).
		Cols("status", "error_message").
		Update(&EkycRequestLog{Status: EKYC_REQUEST_STATUS_FAIL, ErrorMessage: &message})
}

// FindStuckJpkiRequestLogs は、JPKI方式で、署名の結果が確定せず、BPOへも申請されないまま
// 止まっている記録を、新しいものから最大 limit 件返す(MIR-1470)。
//
// 対象は、次をすべて満たす行。
//   - 検証依頼IDがある(JPKI方式。ホ方式の行は検証依頼IDが無い)
//   - 結果が確定していない(Status=NONE)
//   - BPOへ申請していない(申請ID=0)
//   - createdが olderThan より前(Webhookの再送が尽きたと言える時間が経っている)
//   - createdが newerThan より後(ネクスウェイが結果を保管している期間の内。これより前は、
//     検証結果取得APIで引けないため、対象にしない)
//
// 検証結果通知(Webhook)は、受け取れなかった場合の再送が15分間隔で2回までのため、通知を取りこぼす
// (サーバー停止・障害)と、行はこの状態のまま残る。
//
// 新しい行から返すのは、会員が待っている行を先に処理するため。BPOが「受付済み」を返す(申請IDが
// 記録されない)行などは、この条件に残り続けるが、古い行から処理すると、それらが毎回の件数の枠を
// 使い切り、新しい行が処理されなくなる。createdは秒単位で、同じ秒の行を区別できないため、
// 並びはId(作成順に増える)で決める。
func FindStuckJpkiRequestLogs(ctx context.Context, olderThan, newerThan time.Time, limit int) ([]EkycRequestLog, error) {
	logs := make([]EkycRequestLog, 0)
	// DBのタイムゾーンへ変換される書き込みと違い、Whereに渡す時刻は変換されない。MySQLのドライバは
	// 渡した時刻をDSNのlocへ直して送るが、SQLiteはそのまま比べるため、UTCにそろえておく。
	err := factory.DB(ctx).
		Where("request_id IS NOT NULL AND request_id <> '' AND status = ? AND applicant_id = 0 AND created < ? AND created > ?",
			EKYC_REQUEST_STATUS_NONE, olderThan.UTC(), newerThan.UTC()).
		Desc("id").
		Limit(limit).
		Find(&logs)
	if err != nil {
		return nil, err
	}
	return logs, nil
}
