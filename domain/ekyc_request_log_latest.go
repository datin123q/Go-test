package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// HasNewerAppliedRequest は、同じ会員に、この行より新しく、実際に申し込まれた eKYC リクエスト記録が
// あるかを返す(MIR-1468)。あれば、この行は古い申請の記録で、結果が届いても会員の状態は動かさない
// (新しい申請で本人確認が通った会員を、古い申請の結果で保留へ落とさないため)。
//
// 新しい行が「ある」だけでは判断しない。ホ方式の画面は、開くだけで新しい行(NONE)を作る(承認待機の
// 会員が開き直した場合など。VerifyWithEkyc は SUCCESS の行を使い回さない)。その行は申し込まれて
// いないので結果は来ず、古い行の結果を捨てると、会員は承認待機から動かなくなる。JPKI方式の入口を
// 押してやめた行(署名前の NONE)も同じ。
//
// 「申し込まれた」の印は、次のどれかで、かつ FAIL でないこと。
//   - SUCCESS(会員の画面からの完了報告がある)
//   - 結果がある(BPOの結果が届いている)
//   - JPKI方式の行で申請IDがある(署名に成功し、BPOへ申請済み)。ホ方式の行の申請IDは結果の通知でしか
//     入らないため、ホ方式は結果の有無で見る
//
// 申し込まれたかを行から判断できない場合(ホ方式をPCで終え、完了報告が無く、結果も届く前)は、
// 申し込まれていないとみなし、古い行の結果を適用する。その後に新しい行の結果が届けば、それで
// 上書きされる(この判定を入れる前と同じ動き)。
//
// 新しさは id で判断する。created は秒単位で、同じ秒に作られた行を区別できない(id は作成順に増える)。
func (c EkycRequestLog) HasNewerAppliedRequest(ctx context.Context) (bool, error) {
	newer, err := factory.DB(ctx).
		Where("user_id = ? AND id > ?", c.UserId, c.Id).
		And("status <> ?", EKYC_REQUEST_STATUS_FAIL).
		And("(status = ? OR (result IS NOT NULL AND result <> '') OR (request_id IS NOT NULL AND request_id <> '' AND applicant_id <> 0))",
			EKYC_REQUEST_STATUS_SUCCESS).
		Count(&EkycRequestLog{})
	if err != nil {
		return false, err
	}
	return newer > 0, nil
}
