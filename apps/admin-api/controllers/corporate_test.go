package controllers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func TestEvaluateDepositorUpdate(t *testing.T) {
	tests := []struct {
		name     string
		updated  string
		existing string
		want     depositorUpdateVerdict
	}{
		{
			name:     "変更なし・既存値も正常",
			updated:  "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			existing: "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			want:     depositorUpdateAllow,
		},
		{
			name:     "変更なし・既存値が漢字を含む不正値(MIR-1171の既存70件相当)",
			updated:  "株式会社テスト",
			existing: "株式会社テスト",
			want:     depositorUpdateWarnUnchanged,
		},
		{
			name:     "前後の空白のみの違いは変更なし扱い",
			updated:  " ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ ",
			existing: "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			want:     depositorUpdateAllow,
		},
		{
			name:     "変更あり・変更後が正常な文字種",
			updated:  "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ2",
			existing: "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			want:     depositorUpdateAllow,
		},
		{
			name:     "変更あり・変更後が漢字を含む不正値",
			updated:  "株式会社テスト",
			existing: "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			want:     depositorUpdateReject,
		},
		{
			name:     "変更あり・既存値も不正だったが正常な値に修正",
			updated:  "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			existing: "株式会社テスト",
			want:     depositorUpdateAllow,
		},
		// MIR-1171: 口座名義人欄はonBlurで半角カナ→全角カナへ変換されるため、内容を
		// 変えていなくても幅だけが変わって届くことがある。表記ゆれだけなら変更なし扱い
		// にする(ZOLOME側の敵対的検証で発見)。
		{
			name:     "全角/半角の表記ゆれのみは変更なし扱い(半角→全角)",
			updated:  "株式会社ミラリタ",
			existing: "株式会社ﾐﾗﾘﾀ",
			want:     depositorUpdateWarnUnchanged,
		},
		{
			name:     "全角/半角の表記ゆれのみは変更なし扱い(全角→半角)",
			updated:  "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			existing: "カブシキガイシャテスト",
			want:     depositorUpdateAllow,
		},
		// MIR-1171: depositorには`*`量指定子のためIsValidDepositorCharset("")がtrueに
		// なる。既存値が非空のまま空文字へ変更しようとする更新は、意図しない消去を
		// 防ぐため明示的に拒否する(code-reviewで発見)。
		{
			name:     "既存値が非空のまま空文字へ変更しようとした場合は拒否",
			updated:  "",
			existing: "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			want:     depositorUpdateReject,
		},
		{
			name:     "既存値も空文字のまま変更なしなら許可",
			updated:  "",
			existing: "",
			want:     depositorUpdateAllow,
		},
		// MIR-1171: 既存値が全角スペースのみの場合、canonical化すると空文字になるため
		// 「変更なし」分岐へ落ち、existing自体の検証(全角スペースは正規表現上valid)を
		// 素通りして空文字への更新が許可されてしまっていた(code-reviewで発見)。
		{
			name:     "既存値が全角スペースのみの場合に空文字へ変更しようとしても拒否",
			updated:  "",
			existing: "　",
			want:     depositorUpdateReject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, evaluateDepositorUpdate(tt.updated, tt.existing))
		})
	}
}

// MIR-1171: TrimSpaceだけの比較では、既存値に無い改行・NBSP等を末尾へ足しても
// 「変更なし」と誤判定され、検証をすり抜けたまま生の値が保存されてしまっていた
// (code-reviewで発見)。ハンドラは sanitizeDepositorForUpdate で保存前にtrimしてから
// evaluateDepositorUpdate へ渡すため、その一連の流れを検証する。
func TestSanitizeDepositorForUpdate_ThenEvaluate(t *testing.T) {
	tests := []struct {
		name         string
		updated      string
		existing     string
		wantSanitize string
		wantVerdict  depositorUpdateVerdict
	}{
		{
			name:         "末尾に改行が付いていても、trim後は変更なしとして許可され、保存値からは改行が消える",
			updated:      "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ\n",
			existing:     "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			wantSanitize: "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			wantVerdict:  depositorUpdateAllow,
		},
		{
			name:         "末尾にNBSPが付いていても、trim後は変更なしとして許可され、保存値からはNBSPが消える",
			updated:      "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ ",
			existing:     "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			wantSanitize: "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			wantVerdict:  depositorUpdateAllow,
		},
		{
			name:         "内部に改行を含む場合はtrimでは消えず、変更ありとして拒否される",
			updated:      "ｶﾌﾞｼｷ\nｶﾞｲｼｬﾃｽﾄ",
			existing:     "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
			wantSanitize: "ｶﾌﾞｼｷ\nｶﾞｲｼｬﾃｽﾄ",
			wantVerdict:  depositorUpdateReject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := sanitizeDepositorForUpdate(tt.updated)
			assert.Equal(t, tt.wantSanitize, sanitized)
			assert.Equal(t, tt.wantVerdict, evaluateDepositorUpdate(sanitized, tt.existing))
		})
	}
}

// MIR-1252: CorporateUserInput.ToModel()が組み立てる新規構造体には郵便コード確認の
// 承認日が含まれずゼロ値のままになる。この画面で法人会員情報を編集・保存するたびに
// 元の値へ戻さないと、片方だけ確認済みの会員を編集しただけで承認日が消えてしまう
// (②検証で発見。CorporatePostalCodeApprovalDate側は本チケット着手前から同じ理由で
// 消えていた既存不具合)。
func TestCarryOverPostalCodeApprovalDates(t *testing.T) {
	corpApproved := jsonutils.JSONTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	managerApproved := jsonutils.JSONTime(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name   string
		origin domain.CorporateUser
	}{
		{
			name:   "両方未確認",
			origin: domain.CorporateUser{},
		},
		{
			name: "法人側のみ確認済み",
			origin: domain.CorporateUser{
				User: domain.User{CodeApprovalDate: corpApproved},
			},
		},
		{
			name: "担当者側のみ確認済み",
			origin: domain.CorporateUser{
				CorporateProfile: domain.CorporateProfile{CorporatePostalCodeApprovalDate: managerApproved},
			},
		},
		{
			name: "両方確認済み",
			origin: domain.CorporateUser{
				User:             domain.User{CodeApprovalDate: corpApproved},
				CorporateProfile: domain.CorporateProfile{CorporatePostalCodeApprovalDate: managerApproved},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ToModel()が返す新規構造体は承認日フィールドを埋めないため、ゼロ値で再現する。
			// ここへ意図的に別の非ゼロ値を仕込んでおき、originの値で必ず上書きされる
			// (updatedが既に何か値を持っていても無条件で引き継ぐ)ことも確かめる。
			updated := &domain.CorporateUser{
				User:             domain.User{CodeApprovalDate: jsonutils.JSONTime(time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC))},
				CorporateProfile: domain.CorporateProfile{CorporatePostalCodeApprovalDate: jsonutils.JSONTime(time.Date(1999, 1, 2, 0, 0, 0, 0, time.UTC))},
			}

			carryOverPostalCodeApprovalDates(updated, &tt.origin)

			assert.Equal(t, tt.origin.CodeApprovalDate, updated.CodeApprovalDate)
			assert.Equal(t, tt.origin.CorporatePostalCodeApprovalDate, updated.CorporatePostalCodeApprovalDate)
		})
	}
}
