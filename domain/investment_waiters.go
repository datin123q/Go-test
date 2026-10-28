package domain

import (
	"context"
	"encoding/csv"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type INVESTMENT_WAITING_TYPE string

const (
	INVESTMENT_WAITING_CODE                         = "INVESTMENT_WAITING_CODE"
	WAITING                 INVESTMENT_WAITING_TYPE = "WAITING"         // 待機中
	DEPOSIT_REQUEST         INVESTMENT_WAITING_TYPE = "DEPOSIT_REQUEST" // 入金要求
	PROPOSAL                INVESTMENT_WAITING_TYPE = "PROPOSAL"        // 提案中
	WAITING_END             INVESTMENT_WAITING_TYPE = "WAITING_END"     // 待機完了
	CANCELLED               INVESTMENT_WAITING_TYPE = "CANCELLED"       // キャンセル待ちキャンセル
)

func (InvestmentWaiters) GetTypeCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(WAITING), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_TYPE.WAITING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(DEPOSIT_REQUEST), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_TYPE.DEPOSIT_REQUEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROPOSAL), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_TYPE.PROPOSAL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WAITING_END), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_TYPE.WAITING_END", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CANCELLED), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_TYPE.CANCELLED", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: INVESTMENT_WAITING_CODE,
		Codes:     codes,
	}
}

type INVESTMENT_WAITING_STATUS string

const (
	INVESTMENT_WAITING_STATUS_CODE                                             = "INVESTMENT_WAITING_STATUS_CODE"
	INVESTMENT_WAITING_STATUS_WAITING                INVESTMENT_WAITING_STATUS = "WAITING"                // キャンセル待ち
	INVESTMENT_WAITING_STATUS_END_FUNDING            INVESTMENT_WAITING_STATUS = "END_FUNDING"            // 募集終了
	INVESTMENT_WAITING_STATUS_CANCELLED              INVESTMENT_WAITING_STATUS = "CANCELLED"              // キャンセル
	INVESTMENT_WAITING_STATUS_WINNING                INVESTMENT_WAITING_STATUS = "WINNING"                // 当選（未入金）
	INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS  INVESTMENT_WAITING_STATUS = "WINNING_AFTER_SUCCESS"  // 当選（確定）
	INVESTMENT_WAITING_STATUS_WINNING_AFTER_CANCELED INVESTMENT_WAITING_STATUS = "WINNING_AFTER_CANCELED" // 当選（キャンセル）
	INVESTMENT_WAITING_STATUS_WINNING_AFTER_FAILED   INVESTMENT_WAITING_STATUS = "WINNING_AFTER_FAILED"   // 当選（不成立）
	INVESTMENT_WAITING_STATUS_PROPOSAL               INVESTMENT_WAITING_STATUS = "PROPOSAL"               // 投資提案中（廃止済み機能の互換用）
)

func (InvestmentWaiters) GetStatusCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_WAITING_STATUS_WAITING), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_STATUS.WAITING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_WAITING_STATUS_END_FUNDING), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_STATUS.END_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_WAITING_STATUS_CANCELLED), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_STATUS.CANCELLED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_WAITING_STATUS_WINNING), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_STATUS.WINNING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_STATUS.WINNING_AFTER_SUCCESS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_WAITING_STATUS_WINNING_AFTER_CANCELED), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_STATUS.WINNING_AFTER_CANCELED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_WAITING_STATUS_WINNING_AFTER_FAILED), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_STATUS.WINNING_AFTER_FAILED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_WAITING_STATUS_PROPOSAL), Name: i18nutils.Translate("commoncode.INVESTMENT_WAITING_STATUS.PROPOSAL", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: INVESTMENT_WAITING_STATUS_CODE,
		Codes:     codes,
	}
}

// NotDeletedCondition は orm.NotDeletedCond のエイリアス(MIR-1118)。判定条件の実体は
// pkg/orm 側の1箇所にまとめてあり、ここでは呼び出し元の互換のために委譲するだけにしている。
func NotDeletedCondition(table string) string {
	return orm.NotDeletedCond(table)
}

// InvestmentWaitersWinningExistsByWaiterIdCond は「この investment_waiters 行に対する
// 未削除の当選履歴が investment_waiters_winning に1件でもあるか」の判定式(MIR-1118)。
// GetCurrentShareAmountsForUserView(domain/project.go)が、当選済みの待機者をshare側の金額と
// 二重計上しないための除外条件として使う。当選取消(winning行の論理削除)後に待機状態へ
// 戻った待機者を「当選済み」のまま除外し続けないよう、CurrentShareIdと同じく未削除の行だけを
// 見る。
var InvestmentWaitersWinningExistsByWaiterIdCond = "" +
	"EXISTS (SELECT 1 FROM investment_waiters_winning" +
	" WHERE investment_waiters_winning.investment_waiters_id = investment_waiters.id" +
	" AND " + NotDeletedCondition("investment_waiters_winning") + ")"

const (
	// share_id は investment_waiters_winning へ移行済みで、investment_waiters の実カラムは
	// 参照しない(MIR-1118)。中間テーブルからサブクエリで引いた値をこれまでと同じ別名
	// share_id で返す。1申込につき当選は最大1件(繰上げは1回で終わる)ため通常は行が1件に
	// 絞られるが、異常データや将来(MIR-1068)の複数当選に備えて ORDER BY・LIMIT 1 で常に
	// 最新の1件に決定的に絞っている。旧 investment_waiters.share_id 列へのCOALESCE
	// フォールバックはPR3(MIR-1163)で削除した。PR1(マイグレーション)適用後・PR2(このコード)
	// 適用前の期間に旧コードが直接書き込んだ当選は、PR3のバックフィル再実行で
	// investment_waiters_winning へ取り込み済みのため、もう旧列を参照する必要が無い。実カラム
	// 自体は PR4 で削除する。未削除判定を NotDeletedCondition() 経由にせずリテラルで書いて
	// いるのは、この式を DEFAULT_INVESTMENT_WAIT_LIST と合わせて const のまま(パッケージ外
	// から書き換え不能な状態)に保つため。const は関数呼び出しの結果を値にできない。判定条件を
	// 変える場合は NotDeletedCondition() 側とここの両方を直すこと。
	InvestmentWaitersShareIdExpr = "" +
		"(SELECT investment_waiters_winning.share_id FROM investment_waiters_winning" +
		" WHERE investment_waiters_winning.investment_waiters_id = investment_waiters.id" +
		" AND (investment_waiters_winning.deleted IS NULL OR investment_waiters_winning.deleted = '0001-01-01 00:00:00')" +
		" ORDER BY investment_waiters_winning.id DESC LIMIT 1) AS share_id"

	DEFAULT_INVESTMENT_WAIT_LIST = "" +
		//investment-waiters
		"investment_waiters.id," +
		"investment_waiters.project_id," +
		"investment_waiters.user_id," +
		"investment_waiters.waiting_turn," +
		"investment_waiters.qty,investment_waiters.amount, " +
		"investment_waiters.status," +
		"investment_waiters.waiting_date," +
		InvestmentWaitersShareIdExpr + "," +
		// user
		"user.investor_type AS user_investor_type," +
		"user.name AS user_name, user.email AS user_email," +
		"user.mobile_phone AS user_mobile_phone"
)

// InvestmentWaitersInvestStatusExpr は、待機者に紐づく最新の当選(share)のinvest_statusを
// 解決するスカラーサブクエリ(MIR-1068)。GetAll/GetByIdはこれまでinvestment_waiters_winning
// とshareへ実JOINしてinvest_statusを引いていたが、1申込に複数当選がある場合に行が当選件数
// ぶん水増しされる(totalCount・ページングがずれる)ため、InvestmentWaitersShareIdExprと
// 同じ理由でJOINをやめてスカラーサブクエリに変える。常に最新1件
// (ORDER BY investment_waiters_winning.id DESC LIMIT 1)のinvest_statusを返す。
var InvestmentWaitersInvestStatusExpr = "" +
	"(SELECT share.invest_status FROM investment_waiters_winning" +
	" INNER JOIN share ON share.id = investment_waiters_winning.share_id" +
	" WHERE investment_waiters_winning.investment_waiters_id = investment_waiters.id" +
	" AND " + NotDeletedCondition("investment_waiters_winning") +
	" AND " + NotDeletedCondition("share") +
	" ORDER BY investment_waiters_winning.id DESC LIMIT 1) AS invest_status"

// AdvancedQtyExpr は、待機者に紐づく繰上げ済み累計(未削除のinvestment_waiters_winning
// 経由のshare.qty合計)を解決するスカラーサブクエリ(MIR-1068)。investment_waiters_winning
// と share を実JOINすると、1申込に複数当選がある場合に一覧クエリの行が当選件数ぶん
// 水増しされる(GetAllのtotalCount・waiting_investment_jobs.goの対象件数がずれる)ため、
// InvestmentWaitersWinningShareIdByShareExprと同じ理由でJOINにしない。
var AdvancedQtyExpr = "" +
	"(SELECT COALESCE(SUM(share.qty), 0) FROM investment_waiters_winning" +
	" INNER JOIN share ON share.id = investment_waiters_winning.share_id" +
	" WHERE investment_waiters_winning.investment_waiters_id = investment_waiters.id" +
	" AND " + NotDeletedCondition("investment_waiters_winning") +
	" AND " + NotDeletedCondition("share") +
	") AS advanced_qty"

// LatestWaiterDepositCheckStatusExpr は、待機者に紐づく当選のうち最新1件の
// deposit_request.deposit_check_status を解決するスカラーサブクエリ(MIR-1068)。
// waiting_investment_jobs.goのfindWaitingInvestmentItemsが以前投げていた
// investment_waiters_winning・deposit_requestへの実JOINを置き換える(複数当選時の
// 行水増し防止、AdvancedQtyExprと同じ理由)。
var LatestWaiterDepositCheckStatusExpr = "" +
	"(SELECT deposit_request.deposit_check_status FROM investment_waiters_winning" +
	" INNER JOIN deposit_request ON deposit_request.share_id = investment_waiters_winning.share_id" +
	" WHERE investment_waiters_winning.investment_waiters_id = investment_waiters.id" +
	" AND " + NotDeletedCondition("investment_waiters_winning") +
	" AND " + NotDeletedCondition("deposit_request") +
	" ORDER BY investment_waiters_winning.id DESC LIMIT 1) AS deposit_check_status"

// InvestmentWaitersWinningShareIdByShareExpr は、ある share が待機者由来(繰上げ)の
// 投資かどうかを、share.id 起点で判定するためのスカラーサブクエリ(MIR-1118)。
// investment_waiters_share_id という別名で返し、呼び出し側は "" かどうかで判定する。
// apps/jobs-scheduler/jobs/deposit_check_jobs.go の入金期限切れ判定で使う。
// DEFAULT_INVESTMENT_WAIT_LIST 等の const には連結されないため、InvestmentWaitersShareIdExpr と
// 違い const にする理由が無い。var にして NotDeletedCondition() を直接呼んでいる(未削除判定を
// ここでリテラル複製しない)。旧列へのCOALESCEフォールバックはPR3(MIR-1163)で削除した。
// InvestmentWaitersShareIdExpr と同じ理由(PR3のバックフィル再実行で旧コード分も
// investment_waiters_winningへ取り込み済み)。
var InvestmentWaitersWinningShareIdByShareExpr = "" +
	"(SELECT investment_waiters_winning.share_id FROM investment_waiters_winning" +
	" WHERE investment_waiters_winning.share_id = share.id" +
	" AND " + NotDeletedCondition("investment_waiters_winning") +
	" ORDER BY investment_waiters_winning.id DESC LIMIT 1) AS investment_waiters_share_id"

type InvestmentWaitersComparer struct {
	Before InvestmentWaiters
	After  InvestmentWaiters
}

func (i InvestmentWaitersComparer) IsStatusEqual() bool {
	if i.Before.Status == i.After.Status {
		return true
	}

	return false
}

func (i InvestmentWaitersComparer) BeforeId() int64 {
	return i.Before.Id
}

func (i InvestmentWaitersComparer) AfterId() int64 {
	return i.After.Id
}

func (i InvestmentWaitersComparer) BeforeStatus() INVESTMENT_WAITING_TYPE {
	return i.Before.Status
}

func (i InvestmentWaitersComparer) AfterStatus() INVESTMENT_WAITING_TYPE {
	return i.After.Status
}

type InvestmentWaiters struct {
	Id          int64                   `json:"id" xorm:"pk autoincr"`
	ProjectId   int64                   `json:"project_id" swagger:"desc(プロジェクトID)"`
	ProjectName string                  `json:"project_name" xorm:"<-" swagger:"desc(プロジェクト名)"`
	UserId      int64                   `json:"user_id" swagger:"desc(ユーザーID)"`
	WaitingTurn int64                   `json:"waiting_turn" swagger:"desc(待機順番)"`
	Qty         int64                   `json:"qty" swagger:"desc(投資待機申請数量)"`
	Amount      int64                   `json:"amount" swagger:"desc(投資待機申請金額)"`
	Status      INVESTMENT_WAITING_TYPE `json:"status" swagger:"desc(待機状態)"`
	WaitingDate jsonutils.JSONTime      `json:"waiting_date" swagger:"desc(投資待機申請時間)"`
	// ShareId は investment_waiters_winning 経由のサブクエリ(DEFAULT_INVESTMENT_WAIT_LIST内の
	// InvestmentWaitersShareIdExpr)でのみ埋まる読み取り専用フィールド(MIR-1118)。実カラムは
	// PR4(MIR-1163)で撤去済み(xorm:"<-"により、撤去前もUpdate()での書き込みは無視されていた)。
	// .Select()を指定せずGet/Findを呼ぶと、xormが構造体の全フィールドから列リストを組み立て
	// investment_waiters.share_idをSELECT文に含めようとする。JOINが無ければUnknown column
	// で例外になり気付けるが、JOIN先のテーブルにも同名のカラムが無い場合は該当フィールドが
	// 単に空文字列のまま返るだけでエラーにならず、実測で確認済み(気付きにくいぶん危険)。
	// JOIN先に同名のカラム(別のshare_id列)がある場合はそちらの値が黙って入る可能性がある
	// (未実測の推測)。この型に対するクエリは必ずSelect(DEFAULT_INVESTMENT_WAIT_LIST等)か
	// ワイルドカード(table.*、生SQLまたはxormのSelect("table.*"))を明示すること。当選の記録は
	// InvestmentWaitersWinning.Create() を使うこと。
	ShareId string `json:"share_id" xorm:"<-" swagger:"desc(購入ID)"`
	// AdvancedQty は AdvancedQtyExpr でのみ埋まる読み取り専用フィールド(MIR-1068)。
	// ShareIdと同じ注意点(xorm:"<-"、Select省略時のUnknown columnまたは無言のゼロ値化)が
	// 当てはまる。この型に対するクエリは必ずSelectでAdvancedQtyExprを含めること。
	AdvancedQty int64              `json:"advanced_qty" xorm:"<-" swagger:"desc(繰上げ済み累計数量)"`
	Created     jsonutils.JSONTime `json:"created" xorm:"created" swagger:"desc(作成日)"`
	Updated     jsonutils.JSONTime `json:"updated" xorm:"updated" swagger:"desc(更新日)"`
	Deleted     jsonutils.JSONTime `json:"deleted" xorm:"deleted" swagger:"desc(削除日)"`

	UserInvestorType INVESTOR_TYPE `json:"user_investor_type" xorm:"<-" swagger:"desc(ユーザー投資家タイプ)"`
	UserEmail        string        `json:"user_email" xorm:"<-" swagger:"desc(ユーザーメールアドレス)"`
	UserName         string        `json:"user_name" xorm:"<-" swagger:"desc(ユーザー名)"`
	UserMobilePhone  string        `json:"user_mobile_phone" xorm:"<-" swagger:"desc(ユーザー電話番号)"`

	DepositCheckStatus DEPOSIT_CHECK_STATUS `json:"deposit_check_status" xorm:"<-" swagger:"desc(入金確認状態)"`
	InvestStatus       *INVEST_STATUS       `json:"-" xorm:"<-"`
	ProjectStatus      *PROJECT_STATUS      `json:"-" xorm:"<-"`
}

func (InvestmentWaiters) TableName() string {
	return "investment_waiters"
}

func (i InvestmentWaiters) SelectAll() string {
	return i.TableName() + ".*"
}

// InvestmentWaiterIsStillPending は、待機者がまだ当選もキャンセルもされていない
// (=Cancel APIの対象にしてよい)状態かを判定する(MIR-1118)。advancedQty は
// LockAndSumAdvancedQty の戻り値(繰上げ済み累計)を渡すこと。investment_waiters.share_id の
// 実カラムをそのまま渡してはいけない。LockAndSumAdvancedQty は中間テーブル
// (investment_waiters_winning)の行ロック読みだけで判定する(旧列へのフォールバックは
// PR3(MIR-1163)で削除した)。MIR-1068で繰上げが複数回になっても、1件でも当選実績が
// あればキャンセル不可という判定自体は変えない(一部繰上げ済みの残りだけをキャンセル
// する機能は無い)。
func InvestmentWaiterIsStillPending(status INVESTMENT_WAITING_TYPE, advancedQty int64) bool {
	return status == WAITING && advancedQty == 0
}

// InvestmentWaiterCanAdvance は、待機者が繰上げ可能かどうかを判定する(MIR-1068)。
// remaining(繰上げ可能残 = Qty - advancedQty)が0以下なら、いずれの状態でも不可。
// WAITING_ENDは「繰上げ完了」と「募集失敗・提案期限切れによる打ち切り」の両方を
// 意味するため、繰上げ実績(hasWinning、advancedQty > 0)の有無で見分ける。打ち切り
// 済み(実績なし)は再繰上げできない。
func InvestmentWaiterCanAdvance(status INVESTMENT_WAITING_TYPE, remaining int64, hasWinning bool) bool {
	if remaining <= 0 {
		return false
	}
	switch status {
	case WAITING:
		return true
	case WAITING_END:
		return hasWinning
	default:
		return false
	}
}

// InvestmentWaitersWinning は investment_waiters と share の中間テーブル。
// repository 層でのJOINにのみ使用し、ドメイン層では意識しない(lottery_winning に倣う、MIR-1118)。
//
// 一覧・詳細系のクエリ(GetAll/GetById/GetByUserId、waiting_investment_jobs.go)は、第1段階
// (1申込につき当選は最大1件)の間はこのテーブルへ直接 LEFT JOIN していたが、MIR-1068 で
// 1申込に複数当選を許すようになったため、当選件数ぶんの行水増し(totalCount・対象件数の
// ずれ)を避けてスカラーサブクエリ(InvestmentWaitersInvestStatusExpr・AdvancedQtyExpr・
// LatestWaiterDepositCheckStatusExpr)に置き換え済み。当選ごとの明細一覧(管理画面の
// 繰上げ履歴、複数行を意図的に返したい)だけは GetWinningsByWaiterId が実JOINで返す。
type InvestmentWaitersWinning struct {
	Id                  int64              `xorm:"pk autoincr"`
	InvestmentWaitersId int64              `xorm:"not null"`
	ShareId             string             `xorm:"not null"`
	Created             jsonutils.JSONTime `xorm:"created"`
	Updated             jsonutils.JSONTime `xorm:"updated"`
	Deleted             jsonutils.JSONTime `xorm:"deleted"`
}

func (InvestmentWaitersWinning) TableName() string { return "investment_waiters_winning" }

// SelectAll は pkg/orm の LeftJoin/InnerJoin が要求する ormObject インターフェースを満たすために
// 必要(TableName だけでは足りない)。無いと該当の Join 呼び出しがエラー無く静かに無視される。
func (i InvestmentWaitersWinning) SelectAll() string { return i.TableName() + ".*" }

// InvestmentWaitersWinningItem は、待機者1件に紐づく当選明細(管理画面の繰上げ履歴表示、
// MIR-1068)。GetWinningsByWaiterIdの戻り値。1待機者ぶんの明細を複数行で返すのが目的の
// クエリなので、一覧・詳細系のクエリと違って実JOINでよい(水増しではなく、複数当選の
// 各行を見せることが要件)。
type InvestmentWaitersWinningItem struct {
	ShareId            string               `xorm:"share_id" json:"share_id" swagger:"desc(当選share ID)"`
	Qty                int64                `xorm:"qty" json:"qty" swagger:"desc(当選数量)"`
	Price              int64                `xorm:"price" json:"price" swagger:"desc(1口あたりの価格)"`
	Amount             int64                `xorm:"amount" json:"amount" swagger:"desc(当選金額)"`
	InvestDate         jsonutils.JSONTime   `xorm:"invest_date" json:"invest_date" swagger:"desc(当選日時)"`
	InvestStatus       INVEST_STATUS        `xorm:"invest_status" json:"invest_status" swagger:"desc(投資状態)"`
	DepositCheckStatus DEPOSIT_CHECK_STATUS `xorm:"deposit_check_status" json:"deposit_check_status" swagger:"desc(入金確認状態)"`
}

func (w *InvestmentWaitersWinning) Create(ctx context.Context) (int64, error) {
	// 繰上げで作成した share と待機者を紐付ける。
	return factory.DB(ctx).Insert(w)
}

type InvestmentRequest struct { // 메일로 전송된 URL 정보
	Id                  int64              `json:"id" gorm:"pk autoincr"`
	InvestmentWaitersId int64              `json:"investment_waiters_id"`
	ProjectId           int64              `json:"project_id"`
	RequestId           string             `json:"request_id" xorm:"not null UUID"`
	Email               string             `json:"email" swagger:"desc(수신 메일)"`
	Url                 string             `json:"url" swagger:"desc(구매 가입 주소)"`
	Qty                 int64              `json:"qty" swagger:"desc(구매 가능 구수 )"`
	IsExpire            bool               `json:"is_expire" swagger:"desc(만료 일자)"`
	Created             jsonutils.JSONTime `json:"created" xorm:"created"`
}

func (InvestmentRequest) TableName() string {
	return "investment_request"
}

func (i InvestmentRequest) SelectAll() string {
	return i.TableName() + ".*"
}

func (i InvestmentRequest) WithIsExpire(isExpire bool) InvestmentRequest {
	i.IsExpire = isExpire
	return i
}

func (r InvestmentRequest) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(r)
}

func (InvestmentRequest) GetByRequestId(ctx context.Context, id string) (*InvestmentRequest, error) {
	var res InvestmentRequest

	has, err := factory.DB(ctx).Table("investment_request").Where("request_id = ?", id).Get(&res)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	} else if !has {
		return nil, factory.ErrorNotFound.New(nil)
	}
	if res.IsExpire {
		return nil, factory.ErrorExpireRequest.New(nil)
	}
	return &res, nil
}

func (i *InvestmentRequest) IsExpireTrue(ctx context.Context) error {
	i.IsExpire = true
	q := factory.DB(ctx).UseBool("is_expire")
	_, err := q.ID(i.Id).Update(i)

	return err

}

func (InvestmentRequest) GetByInvestmentWaitersId(ctx context.Context, investmentWaitersId int64) (*InvestmentRequest, error) {
	var res InvestmentRequest
	if has, err := factory.DB(ctx).Where("investment_waiters_id = ?", investmentWaitersId).Get(&res); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &res, nil
}

func (i *InvestmentWaiters) Create(ctx context.Context) (int64, error) {
	totalCount, err := factory.DB(ctx).Where("project_id=?", i.ProjectId).Count(InvestmentWaiters{})
	if err != nil {
		return 0, err
	}
	i.WaitingTurn = totalCount + 1
	return factory.DB(ctx).Insert(i)
}

func (InvestmentWaiters) GetById(ctx context.Context, id int64) (*InvestmentWaiters, error) {

	var res InvestmentWaiters
	if has, err := factory.DB(ctx).Table("investment_waiters").Select(DEFAULT_INVESTMENT_WAIT_LIST).
		Join("INNER", "user", "investment_waiters.user_id = user.id").Where("investment_waiters.id = ?", id).Get(&res); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &res, nil
}

func (InvestmentWaiters) GetAll(ctx context.Context, projectId *string, sortby, order string, offset, limit int) (int64, []InvestmentWaiters, error) {
	var err error
	q := factory.DB(ctx).Table("investment_waiters").Select(DEFAULT_INVESTMENT_WAIT_LIST).Join("INNER", "user", "investment_waiters.user_id = user.id")

	if err := modelhelper.SearchEqQuery(q, "investment_waiters", "project_id", projectId); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetSortOrderForTable(q, "investment_waiters", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []InvestmentWaiters
	var totalCount int64

	totalCount, err = q.Limit(limit, offset).FindAndCount(&items)

	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (InvestmentWaiters) DownloadCsv(ctx context.Context, w http.ResponseWriter, sortby, order string, offset, limit int, projectId *string) error {
	q := factory.DB(ctx).Table("investment_waiters").Select(DEFAULT_INVESTMENT_WAIT_LIST).Join("INNER", "user", "investment_waiters.user_id = user.id").Where("")

	if err := modelhelper.SearchEqQuery(q, "investment_waiters", "project_id", projectId); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetSortOrderForTable(q, "investment_waiters", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []InvestmentWaiters

	err := q.Limit(limit, offset).Find(&items)
	if err != nil {
		return err
	}

	records := InvestmentWaiters{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeader(w, "energyshares_investment_waiters_list.csv")
	return csv.NewWriter(w).WriteAll(records)
}

func (InvestmentWaiters) PrepareAndReturnCsv(items []InvestmentWaiters) [][]string {

	type InvestmentWaiters struct {
		Turn        int64              `json:"turn" i18n:"model.investment_waiters.turn"`
		UserType    string             `json:"user_type" i18n:"model.investment_waiters.user_type"`
		Email       string             `json:"email" i18n:"model.investment_waiters.email"`
		Name        string             `json:"name" i18n:"model.investment_waiters.name"`
		Phone       string             `json:"phone" i18n:"model.investment_waiters.phone"`
		Qty         int64              `json:"qty" i18n:"model.investment_waiters.qty"`
		Amount      int64              `json:"amount" i18n:"model.investment_waiters.amount"`
		Status      string             `json:"status" i18n:"model.investment_waiters.status"`
		WaitingDate jsonutils.JSONTime `json:"waiting_date" i18n:"model.investment_waiters.waiting_date"`
	}

	var csv []InvestmentWaiters

	for _, item := range items {
		c := InvestmentWaiters{
			Turn:        item.WaitingTurn,
			UserType:    i18nutils.DefaultTranslate("commoncode.INVESTOR_TYPE." + string(item.UserInvestorType)),
			Email:       item.UserEmail,
			Name:        item.UserName,
			Phone:       item.UserMobilePhone,
			Qty:         item.Qty,
			Amount:      item.Amount,
			Status:      i18nutils.DefaultTranslate("commoncode.INVESTMENT_WAITING_TYPE." + string(item.Status)),
			WaitingDate: item.WaitingDate,
		}
		csv = append(csv, c)
	}
	cols := factory.ReflectJsonFields(reflect.ValueOf(&InvestmentWaiters{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for index := range csv {
		var datas []string
		datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&csv[index]).Elem())...)
		records = append(records, datas)
	}

	return records

}

func (i *InvestmentWaiters) Update(ctx context.Context) (int64, error) {
	return factory.DB(ctx).ID(i.Id).Update(i)
}

func (iw *InvestmentWaiters) ChangeStatus(ctx context.Context, status INVESTMENT_WAITING_TYPE, uid int64) error {
	var mail Mail
	beforeInvestmentWaiters := *iw
	iw.Status = status
	user, err := User{}.GetById(ctx, iw.UserId)

	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if _, err := iw.Update(ctx); err != nil {
		return factory.ErrorDB.New(err)
	}
	switch iw.Status {
	case DEPOSIT_REQUEST:
		_, err = InvestmentWaitHistory{}.CreateStatusHistory(ctx, &beforeInvestmentWaiters, iw, USER_HISTORY_PROCESS_OPERATOR, uid)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
	case WAITING_END:
		investmentRequest, err := InvestmentRequest{}.GetByInvestmentWaitersId(ctx, iw.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		if investmentRequest != nil {
			err = investmentRequest.IsExpireTrue(ctx)
			if err != nil {
				return factory.ErrorDB.New(err)
			}
		}

		_, err = InvestmentWaitHistory{}.CreateStatusHistory(ctx, &beforeInvestmentWaiters, iw, USER_HISTORY_PROCESS_OPERATOR, uid)
		if err != nil {
			return factory.ErrorDB.New(err)
		}

		// (참고) INVEST_STANDBY_CANCELED 메일 발송은 CheckInvestmentProposeJob 에서 처리 중

	case PROPOSAL:
		rId, err := uuid.NewV4()
		if err != nil {
			return err
		}

		project, err := Project{}.GetShareRemainingQty(ctx, iw.ProjectId)
		if err != nil {
			return factory.ErrorDB.New(err)
		}

		requestId := rId.String()
		url := serviceutils.GetUserServiceDomain() + "/product/view" + "/" + strconv.Itoa(int(iw.ProjectId)) + "?request_id=" + requestId
		investmentRequest := InvestmentRequest{
			InvestmentWaitersId: iw.Id,
			ProjectId:           iw.ProjectId,
			RequestId:           requestId,
			Url:                 url,
			IsExpire:            false,
			Email:               user.Email,
			Qty:                 project.RemainingQty,
		}

		if _, err := investmentRequest.Create(ctx); err != nil {
			return factory.ErrorDB.New(err)
		}

		if err := mail.InvestOfferMail(ctx, serviceutils.GetServiceInfoMail(), user.Email, user.Name, strconv.Itoa(int(iw.Qty)), strconv.Itoa(int(project.RemainingQty)), url); err != nil {
			return factory.ErrorAWSSendMail.New(err)
		}

		_, err = InvestmentWaitHistory{}.CreateStatusHistory(ctx, &beforeInvestmentWaiters, iw, USER_HISTORY_PROCESS_OPERATOR, uid)
		if err != nil {
			return factory.ErrorDB.New(err)
		}

		_, err = InvestmentWaitHistory{}.CreateMailSendHistory(ctx, jsonutils.JSONTime(time.Now()), iw, uid)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
	default:
		return factory.ErrorParameter.New(errors.New(""), "Not Support Status")
	}

	return nil
}

func (i InvestmentWaiters) GetByProjectIdAndStatuses(ctx context.Context, projectId int64, status []INVESTMENT_WAITING_TYPE) ([]InvestmentWaiters, error) {
	var err error
	var items []InvestmentWaiters
	err = factory.DB(ctx).Table("investment_waiters").Select(DEFAULT_INVESTMENT_WAIT_LIST).
		Join("INNER", "user", "investment_waiters.user_id = user.id").
		Where("project_id = ?", projectId).
		In("status", status).
		GroupBy("investment_waiters.user_id").Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (i InvestmentWaiters) GetByProjectIdAndStatusesWithoutGrouping(ctx context.Context, projectId int64, status []INVESTMENT_WAITING_TYPE) ([]InvestmentWaiters, error) {
	var items []InvestmentWaiters
	err := factory.DB(ctx).Table("investment_waiters").Select(DEFAULT_INVESTMENT_WAIT_LIST).
		Join("INNER", "user", "investment_waiters.user_id = user.id").
		Where("project_id = ?", projectId).
		In("status", status).
		Desc("investment_waiters.waiting_date", "investment_waiters.created", "investment_waiters.id").
		Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (i InvestmentWaiters) GetByUserId(ctx context.Context, userId int64, sortBy, order string, offset, limit int) (int64, []InvestmentWaiters, error) {
	query := factory.DB(ctx).Table("investment_waiters")
	// investment_waiters_winning・shareへの実JOINは、1申込に複数当選がある場合にマイページの
	// 行を水増しするため使わない(MIR-1068)。GetAll/GetById(admin-api)と同じ理由。
	query.Select(DEFAULT_INVESTMENT_WAIT_LIST+
		",project.name AS project_name"+
		",project.project_status AS project_status"+
		","+InvestmentWaitersInvestStatusExpr+
		","+AdvancedQtyExpr).
		Join("INNER", "user", "investment_waiters.user_id = user.id").
		Join("INNER", "project", "investment_waiters.project_id = project.id").
		Where("user.id = ?", userId)

	if err := modelhelper.SetSortOrderForTable(query, "investment_waiters", sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []InvestmentWaiters
	var totalCount int64
	totalCount, err := query.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (i InvestmentWaiters) GetByUserIdAndProjectID(ctx context.Context, projectId int64, userId int64) ([]InvestmentWaiters, error) {
	var items []InvestmentWaiters

	if err := factory.DB(ctx).
		Table("investment_waiters").
		Select(DEFAULT_INVESTMENT_WAIT_LIST).
		Join("INNER", "user", "investment_waiters.user_id = user.id").
		Where("project_id = ? AND user.id = ?", projectId, userId).
		Find(&items); err != nil {
		return nil, err
	}

	return items, nil
}

func (i InvestmentWaiters) GetTotalAmountForWaitingStatus(ctx context.Context, projectId, userId int64) (int64, error) {

	q := factory.DB(ctx)
	amount, err := q.Where("project_id = ? AND user_id = ? AND status = 'WAITING'", projectId, userId).Sum(&InvestmentWaiters{}, "amount")
	if err != nil {
		return 0, errors.New("fail get total amount for waiting user")
	}
	return int64(amount), nil
}

func (i InvestmentWaiters) GetDisplayStatus() INVESTMENT_WAITING_STATUS {
	switch i.Status {
	case WAITING:
		if i.ProjectStatus == nil {
			return INVESTMENT_WAITING_STATUS_WAITING
		}

		switch *i.ProjectStatus {
		case STANDBY_FUNDING, OPEN_FUNDING, END_FUNDING:
			return INVESTMENT_WAITING_STATUS_WAITING
		case SUCCESS_FUNDING, FAIL_FUNDING, OPERATION_FUNDING, OPERATION_END_FUNDING, CLOSE_FUNDING, INVALID_FUNDING:
			return INVESTMENT_WAITING_STATUS_END_FUNDING
		default:
			return INVESTMENT_WAITING_STATUS_WAITING
		}
	case DEPOSIT_REQUEST:
		return INVESTMENT_WAITING_STATUS_WINNING
	case PROPOSAL:
		return INVESTMENT_WAITING_STATUS_PROPOSAL
	case CANCELLED:
		return INVESTMENT_WAITING_STATUS_CANCELLED
	case WAITING_END:
		if i.InvestStatus == nil {
			return INVESTMENT_WAITING_STATUS_END_FUNDING
		}

		switch *i.InvestStatus {
		case REQUEST_INVEST:
			return INVESTMENT_WAITING_STATUS_WINNING
		case SUCCESS_INVEST, OPERATION_INVEST, REFUND_INVEST:
			return INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS
		case CANCEL_INVEST:
			return INVESTMENT_WAITING_STATUS_WINNING_AFTER_CANCELED
		case FAIL_INVEST:
			return INVESTMENT_WAITING_STATUS_WINNING_AFTER_FAILED
		default:
			return INVESTMENT_WAITING_STATUS_END_FUNDING
		}
	}

	return ""
}
