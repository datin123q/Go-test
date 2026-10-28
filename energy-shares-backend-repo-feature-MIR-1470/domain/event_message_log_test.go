package domain

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// イベントのログにイベント本体や dto を丸ごと渡していないかを見張る。
//
// content には氏名・取引金額・残高が入る（TransactionDto など）。以前は
// これをそのままログへ出しており、出資のような一括処理では1件ごとに
// 566〜870 バイトが 1,000 件以上残っていた。
//
// 同じ機能の実装が domain と service/event_message の2つに並存しているため、
// 片方だけ直すと見落とす。ソースを走査して両方まとめて防ぐ。
// TestEventMessageLog_DoesNotLeakSensitiveContent はイベントのログに
// イベント本体や dto を丸ごと渡していないかを見張る。
func TestEventMessageLog_DoesNotLeakSensitiveContent(t *testing.T) {
	targets := []string{
		"event_message.go",
		"../service/event_message/event_message_svc.go",
	}

	// logrus.Fields のコロン区切り（"event": e）と、WithField のカンマ区切り
	// （WithField("event", eventMessage)）の両方がある。片方だけ見ていると
	// 素通りするため、区切り文字は : と , の両方で拾う。
	//
	// \s は改行にも一致するので、gofmt でキーと値が別行に分かれても拾える。
	forbidden := []*regexp.Regexp{
		// 中身を丸ごと載せるキー。値が何であれ載せてはいけないので、
		// 変数名は問わずキー名だけで弾く。変数を e から ev へ替えたり
		// eventMessage.Content のようにフィールドを取り出しても素通りしない。
		regexp.MustCompile(`"(event|data|dto|content|payload|body|request|response)"\s*[:,]\s*[^"\s]`),
		// キー名を変えても、値が content / dto なら同じことになる。
		// キー付きの形に限る。json.Marshal(e.Content) のような通常の処理まで
		// 拾うと、ログと関係のない変更で落ちてしまう。
		regexp.MustCompile(`"[^"]*"\s*[:,]\s*\w+\.(Content|Dto|DTO)\b`),
	}

	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			body, err := os.ReadFile(target)
			require.NoError(t, err)

			src := string(body)

			var found []string
			for _, re := range forbidden {
				for _, loc := range re.FindAllStringIndex(src, -1) {
					line := strings.Count(src[:loc[0]], "\n") + 1
					found = append(found,
						strings.TrimSpace(src[loc[0]:loc[1]])+" (L"+strconv.Itoa(line)+")")
				}
			}

			assert.Empty(t, found,
				"イベントや dto を丸ごとログへ渡している。識別子だけにすること: %v", found)
		})
	}
}

// TestEventMessageLog_IncludesBusinessIdentifier は、イベント発行のログに
// 対象を追跡できる業務上の識別子(investment_id・user_id・product_id など)が
// 残っていることをイベント種別ごとに確かめる。
//
// 前のテスト(丸ごと渡していないか)だけでは「削っただけで何も残っていない」
// 状態を見逃す。障害調査で対象を特定できなくなると本末転倒になるため、
// イベントごとに何を識別子として残すかを個別に固定する。
func TestEventMessageLog_IncludesBusinessIdentifier(t *testing.T) {
	body, err := os.ReadFile("event_message.go")
	require.NoError(t, err)
	src := string(body)

	tests := []struct {
		funcName       string
		wantIdentifier string
	}{
		{"CreateCustomerLedgerDepositEvent", "user_id"},
		{"CreateInvestmentCancelEvent", "investment_id"},
		{"CreateInvestmentVoidEvent", "investment_id"},
		{"SendCustomerLedgerWithdrawalEvent", "account_id"},
		{"CreateFundFailureEvent", "product_id"},
		{"CreateDividendExecutedEvent", "schedule_id"},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			fn := extractFuncBody(t, src, tt.funcName)
			assert.Contains(t, fn, `"`+tt.wantIdentifier+`"`,
				"%s のログに %q が含まれていない。障害調査で対象を特定できなくなる",
				tt.funcName, tt.wantIdentifier)
		})
	}

	// 以下は EventIdentifierFields(dto) に切り替えた関数。ログキーは実行時に
	// 決まるためリテラル走査では見えない(TestEventIdentifierFields が抽出
	// ロジック自体は担保する)。ここでは「dto に識別子フィールドを詰める行
	// 自体が消えていないか」を Go の構造体フィールド名(ログキーではない)で
	// 確認する。EventIdentifierFields はゼロ値を出力しないため、この代入が
	// 削られると識別子が黙って消える。
	sharedExtractorFuncs := []struct {
		funcName   string
		wantFields []string
	}{
		{"CreateRequestInvestmentEvent", []string{"InvestmentId", "UserId", "ProductId"}},
		{"CreateInvestmentWaitRequestEvent", []string{"InvestmentWaiterId", "UserId", "ProductId"}},
		{"CreateCompletedInvestmentEvent", []string{"InvestmentId", "UserId", "ProductId"}},
		{"CreateLotteryRequestEvent", []string{"UserId", "ProductId"}},
		{"CreateCustomerLedgerDividendEvent", []string{"InvestmentId", "UserId", "ProductId"}},
		{"CreateUserDividendExecutedEvent", []string{"UserId", "ProductId"}},
	}
	for _, tt := range sharedExtractorFuncs {
		t.Run(tt.funcName, func(t *testing.T) {
			fn := extractFuncBody(t, src, tt.funcName)
			assert.Contains(t, fn, "EventIdentifierFields(dto)",
				"%s が共通の識別子抽出(EventIdentifierFields)を使っていない", tt.funcName)
			for _, field := range tt.wantFields {
				assert.Contains(t, fn, field+":",
					"%s の dto 構築から %s: への代入が消えている。障害調査で対象を特定できなくなる",
					tt.funcName, field)
			}
		})
	}
}

// TestEventIdentifierFields は EventIdentifierFields が dto の型ごとに
// 期待どおりの識別子を拾うことを確かめる。
//
// イベント種別の判定はここでは行わず、フィールド名だけを見て機械的に拾う
// 実装なので、テストも型ごとにどのフィールドが立っているかで確認する。
func TestEventIdentifierFields(t *testing.T) {
	tests := []struct {
		name    string
		content any
		want    logrus.Fields
	}{
		{
			name:    "InvestmentRequestDto",
			content: InvestmentRequestDto{UserId: 1, InvestmentId: "inv-1", ProductId: 2},
			want:    logrus.Fields{"user_id": int64(1), "investment_id": "inv-1", "product_id": int64(2)},
		},
		{
			name:    "InvestmentWaitRequestDto",
			content: InvestmentWaitRequestDto{UserId: 1, InvestmentWaiterId: 3, ProductId: 2},
			want:    logrus.Fields{"user_id": int64(1), "investment_waiter_id": int64(3), "product_id": int64(2)},
		},
		{
			name:    "LotteryRequestDto",
			content: LotteryRequestDto{UserId: 1, ProductId: 2},
			want:    logrus.Fields{"user_id": int64(1), "product_id": int64(2)},
		},
		{
			name:    "InvestmentCompletedDto",
			content: InvestmentCompletedDto{UserId: 1, InvestmentId: "inv-1", ProductId: 2},
			want:    logrus.Fields{"user_id": int64(1), "investment_id": "inv-1", "product_id": int64(2)},
		},
		{
			name:    "InvestmentVoidDto",
			content: InvestmentVoidDto{InvestmentId: "inv-1"},
			want:    logrus.Fields{"investment_id": "inv-1"},
		},
		{
			name:    "FundFailureDto",
			content: FundFailureDto{ProductId: 2},
			want:    logrus.Fields{"product_id": int64(2)},
		},
		{
			name:    "InvestmentCancelDto",
			content: InvestmentCancelDto{InvestmentId: "inv-1"},
			want:    logrus.Fields{"investment_id": "inv-1"},
		},
		{
			// 障害の実件数はこの型が2種類のイベントで発火し続けたことに由来する。
			// SpvId のような識別子以外のフィールドは拾わない。
			name:    "TransactionDto",
			content: TransactionDto{UserId: 1, InvestmentId: "inv-1", ProductId: 2, SpvId: 9},
			want:    logrus.Fields{"user_id": int64(1), "investment_id": "inv-1", "product_id": int64(2)},
		},
		{
			name:    "DividendDecisionDto",
			content: DividendDecisionDto{ScheduleId: 4, ProductId: 2},
			want:    logrus.Fields{"schedule_id": int64(4), "product_id": int64(2)},
		},
		{
			name:    "UserDividendDecisionDto",
			content: UserDividendDecisionDto{ScheduleId: 4, ProductId: 2, UserId: 1},
			want:    logrus.Fields{"schedule_id": int64(4), "product_id": int64(2), "user_id": int64(1)},
		},
		{
			// InvestmentId を持たない dto。product_id と user_id だけで
			// 一括処理内の1件を特定できないのは既知の限界(報告済み)。
			name:    "TransferFundDto",
			content: TransferFundDto{ProductId: 2, UserId: 1},
			want:    logrus.Fields{"product_id": int64(2), "user_id": int64(1)},
		},
		{
			name:    "ポインタでも同じ結果になる",
			content: &TransactionDto{UserId: 1, InvestmentId: "inv-1", ProductId: 2},
			want:    logrus.Fields{"user_id": int64(1), "investment_id": "inv-1", "product_id": int64(2)},
		},
		{
			name:    "nilポインタは何も返さない",
			content: (*TransactionDto)(nil),
			want:    logrus.Fields{},
		},
		{
			name:    "contentがnilなら何も返さない",
			content: nil,
			want:    logrus.Fields{},
		},
		{
			// ゼロ値のフィールドは「識別子が無い」のか「本当に0/空」なのか
			// 区別できないため出さない。identifier系フィールドはこの
			// コードベースでは自動採番でゼロを取らない前提。
			name:    "ゼロ値のフィールドは出さない",
			content: TransactionDto{},
			want:    logrus.Fields{},
		},
		{
			name:    "struct以外は何も返さない",
			content: "not a struct",
			want:    logrus.Fields{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EventIdentifierFields(tt.content)
			assert.Equal(t, tt.want, got)
		})
	}
}

// extractFuncBody は src からトップレベル関数 name の本文(次の関数の
// 直前まで)を切り出す。gofmt 済みのソースであること(関数は列0の
// "func " で始まり、次の関数との間に空行がある)を前提にした簡易版で、
// 括弧の対応は数えない。
func extractFuncBody(t *testing.T, src, name string) string {
	t.Helper()

	marker := "\nfunc " + name + "("
	start := strings.Index(src, marker)
	require.NotEqual(t, -1, start, "関数 %s が見つからない", name)

	rest := src[start+1:]
	if next := strings.Index(rest, "\nfunc "); next != -1 {
		return rest[:next]
	}
	return rest
}
