package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/feequote"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

const TRANSFER_LIMIT = 98

type TRANSFER_STATUS_TYPE string

const (
	TRANSFER_STATUS_TYPE_CODE                      = "TRANSFER_STATUS_TYPE"
	TRANSFER_STATUS_WAIT      TRANSFER_STATUS_TYPE = "TRANSFER_STATUS_WAIT"      // 미완료 送金待ち
	TRANSFER_STATUS_SUCCESS   TRANSFER_STATUS_TYPE = "TRANSFER_STATUS_SUCCESS"   // 완료 送金成功
	TRANSFER_STATUS_FAIL      TRANSFER_STATUS_TYPE = "TRANSFER_STATUS_FAIL"      // 실패 送金失敗
	TRANSFER_STATUS_REPAYMENT TRANSFER_STATUS_TYPE = "TRANSFER_STATUS_REPAYMENT" // 환불 払戻し
	TRANSFER_STATUS_CANCEL    TRANSFER_STATUS_TYPE = "TRANSFER_STATUS_CANCEL"    // 취소 送金キャンセル
)

func (Transfer) GetTransferStatusTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(TRANSFER_STATUS_WAIT), Name: i18nutils.Translate("commoncode.TRANSFER_STATUS_TYPE.TRANSFER_STATUS_WAIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(TRANSFER_STATUS_SUCCESS), Name: i18nutils.Translate("commoncode.TRANSFER_STATUS_TYPE.TRANSFER_STATUS_SUCCESS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(TRANSFER_STATUS_FAIL), Name: i18nutils.Translate("commoncode.TRANSFER_STATUS_TYPE.TRANSFER_STATUS_FAIL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(TRANSFER_STATUS_REPAYMENT), Name: i18nutils.Translate("commoncode.TRANSFER_STATUS_TYPE.TRANSFER_STATUS_REPAYMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(TRANSFER_STATUS_CANCEL), Name: i18nutils.Translate("commoncode.TRANSFER_STATUS_TYPE.TRANSFER_STATUS_CANCEL", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: TRANSFER_STATUS_TYPE_CODE,
		Codes:     codes,
	}
}

// 송금 정보
type Transfer struct {
	Id                    int64                        `json:"id" xorm:"pk autoincr not null"`
	ItemId                int                          `json:"item_id" xorm:"not null"`
	RequestId             int64                        `json:"request_id" xorm:"not null"`
	TransactionId         string                       `json:"transaction_id" xorm:"UUID NOT NULL INDEX"`
	UserId                int64                        `json:"user_id" xorm:"user_id"`
	ProjectId             int64                        `json:"project_id" xorm:"project_id"`
	TransferDate          jsonutils.JSONTime           `json:"transfer_date" xorm:"transfer_date"`
	TransferAmount        int64                        `json:"transfer_amount" xorm:"transfer_amount"`
	AccountNumber         string                       `json:"account_number" xorm:"account_number"`
	BankCode              string                       `json:"bank_code" xorm:"bank_code"`
	BankName              string                       `json:"bank_name" xorm:"bank_name"`
	BranchCode            string                       `json:"branch_code" xorm:"branch_code"`
	BranchName            string                       `json:"branch_name" xorm:"branch_name"`
	Depositor             string                       `json:"depositor" xorm:"depositor"`
	AccountTypeCode       string                       `json:"account_type_code" xorm:"account_type_code"`
	TransferFee           int64                        `json:"transfer_fee" xorm:"transfer_fee"`
	ErrorCode             string                       `json:"error_code" xorm:"error_code"`
	ErrorMsg              string                       `json:"error_msg" xorm:"error_msg"`
	TransactionDetailType USER_TRANSACTION_DETAIL_TYPE `json:"transaction_type" swagger:"desc(거래종류)"`
	TransferStatus        TRANSFER_STATUS_TYPE         `json:"transfer_status" xorm:"transfer_status"`
	RepaymentDate         jsonutils.JSONTime           `json:"repayment_date" xorm:"repayment_date"`
	Created               jsonutils.JSONTime           `json:"created" xorm:"created"`
	Updated               jsonutils.JSONTime           `json:"updated" xorm:"updated"`
}

// TransferInput is the payload for fee estimation / transfer requests
type TransferInput struct {
	ItemId          string                  `json:"item_id,omitempty"`       // 内部マッピング用（ジョブからのレスポンスと突合）
	WithdrawalId    int64                   `json:"withdrawal_id,omitempty"` // 依頼元IDを明示的に送る
	UserId          int64                   `json:"user_id,omitempty"`
	ProjectId       int64                   `json:"project_id,omitempty"`
	TransactionId   string                  `json:"transaction_id,omitempty"`
	RequestDate     string                  `json:"request_date,omitempty"`
	RequestAmount   int64                   `json:"request_amount,omitempty"`
	AccountNumber   string                  `json:"account_number,omitempty"`
	BankCode        string                  `json:"bank_code,omitempty"`
	BankName        string                  `json:"bank_name,omitempty"`
	BranchCode      string                  `json:"branch_code,omitempty"`
	BranchName      string                  `json:"branch_name,omitempty"`
	Depositor       string                  `json:"depositor,omitempty"`
	AccountTypeCode string                  `json:"account_type_code,omitempty"`
	RequestStatus   TRANSFER_REQUEST_STATUS `json:"request_status,omitempty"`
}

func (t Transfer) TableName() string {
	return "transfer"
}

func (t Transfer) SelectAll() string {
	return t.TableName() + ".*"
}

// transferWithWithdrawalId はGetLatestByWithdrawalIds専用の結果保持型。withdrawal_idは
// transfer自身のカラムではなくJOIN結果としてのみ得られる値のため、xorm:"<-"で読み取り専用に
// し、Insert/Updateへ渡しても存在しないカラムへ書き込もうとしないようにしている(MIR-1180)。
type transferWithWithdrawalId struct {
	Transfer     `xorm:"extends"`
	WithdrawalId int64 `xorm:"withdrawal_id <-"`
}

// GetLatestByWithdrawalIds returns the latest transfer per withdrawal id (descending by transfer.id)
// Key: withdrawal_id, Value: latest Transfer
func (Transfer) GetLatestByWithdrawalIds(ctx context.Context, withdrawalIds []int64) (map[int64]Transfer, error) {
	if len(withdrawalIds) == 0 {
		return map[int64]Transfer{}, nil
	}

	// withdrawal_idをマップのキーに使うためJOIN結果へ含める。transfer.*ではなくtransferの
	// カラムを列挙しているのは、将来transferにwithdrawal_idという実カラムが増えたときに、
	// xormがエイリアスより実カラムを優先して黙って取り違え、このバグが無言で再発するのを
	// 防ぐため(CLAUDE.mdのReview Checklist参照)。
	items := make([]transferWithWithdrawalId, 0, len(withdrawalIds))
	err := factory.DB(ctx).Table("transfer").
		Select("transfer.id, transfer.item_id, transfer.request_id, transfer.transaction_id, "+
			"transfer.user_id, transfer.project_id, transfer.transfer_date, transfer.transfer_amount, "+
			"transfer.account_number, transfer.bank_code, transfer.bank_name, transfer.branch_code, "+
			"transfer.branch_name, transfer.depositor, transfer.account_type_code, transfer.transfer_fee, "+
			"transfer.transaction_detail_type, transfer.created, transfer.updated, transfer.transfer_status, "+
			"transfer.repayment_date, transfer.error_code, transfer.error_msg, "+
			"withdrawal_transfer_mapping.withdrawal_id").
		Join("INNER", "withdrawal_transfer_mapping", "withdrawal_transfer_mapping.transaction_id = transfer.transaction_id").
		In("withdrawal_transfer_mapping.withdrawal_id", withdrawalIds).
		Desc("transfer.id").
		Find(&items)
	if err != nil {
		return nil, err
	}

	latestMap := make(map[int64]Transfer, len(withdrawalIds))
	for _, tr := range items {
		// Keep the first (because Desc order). 1つの出金依頼に複数の送金(再送金)が
		// 紐づきうるため、withdrawal_idごとに最初に現れたもの(=transfer.id最大)を残す。
		if _, exists := latestMap[tr.WithdrawalId]; !exists {
			latestMap[tr.WithdrawalId] = tr.Transfer
		}
	}
	return latestMap, nil
}

// TransferDesignatedDateは、GMOあおぞら銀行への振込指定日(YYYY-MM-DD)を返す。
// MIR-1172: 手数料見積もりAPI(BulkTransferFeeUsingPOST)は、以前は
// jsonutils.JSONTime(time.Now()).ToString()(当日・時刻付き)を渡しており、
// 銀行側のバリデーションにほぼ毎回落ちていた(詳細はdesk、JIRAコメント13897)。
// 本振込のジョブ(apps/jobs-scheduler/jobs/transfer_jobs.go)は「JSTの翌日・
// 日付のみ」を渡しており、これは通っている。見積もり側もこの形式に
// そろえ、同じ日付になるようにする。
func TransferDesignatedDate(now time.Time) string {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		// Asia/Tokyoは標準のtzdataに常に含まれるため実運用では起こらない。
		// フォールバックとしてUTC+9の固定オフセットを使う。
		loc = time.FixedZone("Asia/Tokyo", 9*60*60)
	}
	return now.In(loc).Add(24 * time.Hour).Format(time.DateOnly)
}

func (Transfer) RequestFee(ctx context.Context, user User, amount int64) (int64, error) {
	// MIR-1256: bank_code/branch_codeはDBでNULLを許容している(Go側は*string)。
	// 以前はnilチェック無しで参照してパニックしていた。user-apiは
	// echoutils.CustomRecover()でリクエスト単位にパニックを回復するため
	// プロセス全体は落ちないが、この関数はuser-apiの出金リクエスト
	// (apps/user-api/controllers/withdrawal.go)から呼ばれるため、会員本人の
	// 口座情報が未登録なだけで意図しない500・パニックのスタックトレースに
	// なっていた。エラーを返すだけにする。
	if err := WithdrawalBankInfoError(user.BankCode, user.BranchCode, user.BankAccount, user.DepositType); err != nil {
		return 0, errors.Wrap(err, "fail request transfer fee")
	}

	var fee int64

	v := TransferInput{
		UserId:          user.Id,
		RequestDate:     TransferDesignatedDate(time.Now()),
		RequestAmount:   amount,
		AccountNumber:   user.BankAccount,
		BankCode:        *user.BankCode,
		BankName:        user.BankName,
		BranchCode:      *user.BranchCode,
		BranchName:      user.BranchName,
		Depositor:       user.Depositor,
		AccountTypeCode: DepositAccountTypeCheck(user.DepositType),
	}

	pbytes, _ := json.Marshal(v)
	buff := bytes.NewBuffer(pbytes)
	resp, err := http.Post(serviceutils.GetJobSchedulerUrl()+"/bulk-transfers/fee", "application/json", buff)
	if err != nil {
		return 0, errors.Wrap(err, "fail request transfer fee")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		str := string(respBody)
		factory.Logger(ctx).WithFields(logrus.Fields{
			"user_id":   user.Id,
			"user_name": user.Name,
			"amount":    amount,
			"resp":      str,
		}).Error("fail http fee request")
		return 0, err
	}

	var res struct {
		Result any `json:"result"`
	}
	if err := json.Unmarshal(respBody, &res); err != nil {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"user_id":   user.Id,
			"user_name": user.Name,
			"response":  string(respBody),
			"err":       err,
		}).Error("fail http fee Unmarshal")
		return 0, errors.Wrap(err, "fail http fee Unmarshal")
	}

	fee, err = parseFeeResult(res.Result)
	if err != nil {
		return 0, err
	}
	return fee, nil
}

// parseFeeResultは、/bulk-transfers/feeのresultフィールド(型は不明、stringまたはfloat64
// のときだけ手数料として受け入れる)を解釈する。MIR-1172: この単件の応答の型は変えない。
// is_fallbackを運ぶために構造体化すると、不正な応答(resultがnull等)がfee=0・エラー無し
// として静かに通ってしまう。string/float64以外を弾くdefault分岐が安全弁であり、判断を
// 切り出してテストできるようにした。
func parseFeeResult(result any) (int64, error) {
	switch v := result.(type) {
	case string:
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "fail parse result (string)")
		}
		return val, nil
	case float64:
		return int64(v), nil
	default:
		return 0, errors.New("unexpected result type for fee")
	}
}

// RequestFeeBulk 複数件まとめて手数料見積もりを取得する
// 戻り値: 各明細の手数料スライス, 合計手数料, 固定値へのフォールバックが
// 使われたか(MIR-1172), エラー
func (Transfer) RequestFeeBulk(ctx context.Context, inputs []TransferInput) ([]int64, int64, bool, error) {
	if len(inputs) == 0 {
		return nil, 0, false, errors.New("inputs is empty")
	}

	type TransferBatchInput struct {
		Transfers []TransferInput `json:"transfers"`
	}

	body := TransferBatchInput{Transfers: inputs}
	pbytes, _ := json.Marshal(body)
	buff := bytes.NewBuffer(pbytes)

	resp, err := http.Post(serviceutils.GetJobSchedulerUrl()+"/bulk-transfers/fee-bulk", "application/json", buff)
	if err != nil {
		return nil, 0, false, errors.Wrap(err, "fail request transfer fee bulk")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, false, errors.Wrap(err, "fail read fee bulk response")
	}

	var wrapper struct {
		Result feequote.WithdrawalFeeBulkResponse `json:"result"`
	}
	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return nil, 0, false, errors.Wrap(err, "fail unmarshal fee bulk response")
	}
	res := wrapper.Result

	feeMap := make(map[int64]int64, len(res.Details))
	for _, d := range res.Details {
		feeMap[d.WithdrawalID] = d.TransferFee
	}

	fees := make([]int64, 0, len(inputs))
	for _, in := range inputs {
		id := in.WithdrawalId
		if id == 0 {
			// fallback: old item_id を数値化
			if in.ItemId != "" {
				if parsed, err := strconv.ParseInt(in.ItemId, 10, 64); err == nil {
					id = parsed
				}
			}
		}
		fee, ok := feeMap[id]
		if !ok {
			return nil, 0, false, errors.Errorf("fee not found for withdrawal_id=%d", id)
		}
		fees = append(fees, fee)
	}

	if len(res.Details) == 0 {
		if res.TotalFee != 0 {
			return []int64{}, res.TotalFee, res.IsFallback, nil
		}
		return nil, 0, false, errors.New("transfer_fee_details not found")
	}

	totalFee := res.TotalFee

	return fees, totalFee, res.IsFallback, nil
}

func (Transfer) GetAll(ctx context.Context, sortby, order string, offset, limit int, startDate, endDate string) (int64, []Transfer, error) {
	q := factory.DB(ctx).Table("transfer").Select("transfer.*")
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "request_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]Transfer, 0, 0)
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

// GetByRequestId 클린아키텍처: GetAll 에 포함
func (Transfer) GetByRequestId(ctx context.Context, id int64) (int64, []Transfer, error) {
	q := factory.DB(ctx).Table("transfer").Select("transfer.*")
	items := make([]Transfer, 0, 0)

	q.Where("request_id = ?", id).OrderBy("id asc")

	var totalCount int64
	totalCount, err := q.FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (Transfer) GetByID(ctx context.Context, id int64) (*Transfer, error) {
	var d Transfer
	has, err := factory.DB(ctx).ID(id).Get(&d)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &d, nil
}

// GetByWithdrawalId 클린아키텍처: GetAll 에 포함
func (Transfer) GetByWithdrawalId(ctx context.Context, id int64) (int64, []Transfer, error) {
	q := factory.DB(ctx).Table("transfer").Select("transfer.*").
		Join("INNER", "withdrawal_transfer_mapping", "withdrawal_transfer_mapping.transaction_id = transfer.transaction_id").
		Where("withdrawal_transfer_mapping.withdrawal_id = ?", id).
		OrderBy("transfer.id DESC")

	items := make([]Transfer, 0, 0)
	var totalCount int64

	totalCount, err := q.FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (d *Transfer) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(d)
}

func (d *Transfer) Update(ctx context.Context) error {
	q := factory.DB(ctx)
	_, err := q.ID(d.Id).Update(d)
	return err
}

// InsertMany は渡されたitemsをまとめてINSERTする。以前はここで独自に
// session.Begin/Commit/Closeを呼んでいたため、呼び出し元がfactory.TxManager等で
// 張った外側のトランザクションを、この呼び出し1回で早期コミット・クローズして
// しまっていた(MIR-1181)。トランザクション制御は呼び出し側に委ね、ここでは
// ctxが差すセッション/エンジンへそのままINSERTするだけにする。ctxがSession以外
// (Engine)のときも黙って0件・エラー無しを返す分岐があったが、Engineでも
// Insertは実行できるため、その場合分け自体を無くした。
func (Transfer) InsertMany(ctx context.Context, items []Transfer) (int64, error) {
	return factory.DB(ctx).Insert(&items)
}

// UpdateMany は渡されたitemsを1件ずつUPDATEする。以前はここで独自に
// session.Begin/Commit/defer Closeを呼んでいたため、cronジョブ(bulkTransferJob等)の
// ループの中で複数回呼ばれると、1回目の呼び出しでジョブ全体のセッションを閉じてしまい、
// 2回目以降が失敗していた(MIR-1183)。InsertMany(MIR-1181)と同じくトランザクション
// 制御を呼び出し側に委ね、ctxが差すセッション/エンジンへそのままUPDATEするだけにする。
func (Transfer) UpdateMany(ctx context.Context, items []Transfer) (int64, error) {
	db := factory.DB(ctx)
	var count int64
	for _, transfer := range items {
		n, err := db.ID(transfer.Id).Update(&transfer)
		if err != nil {
			return count, err
		}
		count += n
	}
	return count, nil
}

// 송금 요청 결과
type TransferRequestResult struct {
	Id int64 `json:"id"`
	// 口座ID 半角英数字 口座を識別するID
	AccountId string `json:"account_id" xorm:"account_id"`
	// 振込依頼人名 半角文字 指定しない場合は口座名義がデフォルト値となります 振込許容文字を指定可 ただし、一部の非許容文字は、許容文字に変換を行います
	RemitterName string `json:"remitter_name,omitempty" xorm:"remitter_name"`
	// 振込指定日 半角文字 YYYY-MM-DD形式 振込対象の指定日 ただし、振込指定日が非営業日で、非営業日に実施できない振込（他行宛振込み）が振込情報に1件以上存在する場合、以下のとおりとなります ・「振込指定日休日コード」が1または省略の場合、振込指定日の翌営業日を振込対象の指定日となります ・「振込指定日休日コード」が2の場合、振込指定日の前営業日を振込対象の指定日となります ・「振込指定日休日コード」が3の場合、エラーとなり「400 Bad Request」を返却します
	TransferDesignatedDate string `json:"transfer_designated_date" xorm:"transfer_designated_date"`
	// 振込指定日休日コード 半角数字 1：翌営業日、2：前営業日、3：エラー返却 省略可（省略した場合は1とみなします）
	TransferDateHolidayCode string `json:"transfer_date_holiday_code,omitempty" xorm:"transfer_date_holiday_code"`
	// 合計件数 半角数字 1以上99件まで指定可能（0はエラー） 1件のみの場合は省略可（項目自体の設定が不要です）
	TotalCount string `json:"total_count,omitempty" xorm:"total_count"`
	// 合計金額 半角数字 1以上999,999,999,999円以下　数値のみで0、カンマ、マイナス不可 1件のみの場合は省略可（項目自体の設定が不要です）
	TotalAmount string `json:"total_amount,omitempty" xorm:"total_amount"`
	// 振込申請メモ（申請コメント） 全半角文字 ビジネスID管理申し込みかつ、申請になるユーザーの場合のみ有効 それ以外は値を設定しても銀行で読み捨て 該当する情報が無い場合は省略可（項目自体の設定が不要です）
	ApplyComment string           `json:"apply_comment,omitempty" xorm:"apply_comment"`
	Transfers    []TransferResult `json:"transfers" xorm:"-"`
}

// 송금 결과
type TransferResult struct {
	Id int64 `json:"id"`
	// 明細番号 半角数字 重複/0はエラー　1～99の範囲で１からの連番とします 1件のみの場合は省略可（項目自体の設定が不要です）
	ItemId string `json:"item_Id,omitempty"`
	// 振込金額 半角数字 1以上、整数のみ
	TransferAmount string `json:"transfer_amount"`
	// EDI情報 半角文字 振込許容文字を指定可
	EdiInfo string `json:"edi_info,omitempty"`
	// 被仕向金融機関番号 半角数字
	BeneficiaryBankCode string `json:"beneficiary_bank_code"`
	// 被仕向金融機関名カナ 半角文字 参考値、処理に利用しません
	BeneficiaryBankName string `json:"beneficiary_bank_name,omitempty"`
	// 被仕向支店番号 半角数字
	BeneficiaryBranchCode string `json:"beneficiary_branch_code"`
	// 被仕向支店名カナ 半角文字 参考値、処理に利用しません
	BeneficiaryBranchName string `json:"beneficiary_branch_name,omitempty"`
	// 科目コード（預金種別コード） 半角数字 1：普通、2：当座、4：貯蓄、9：その他
	AccountTypeCode string `json:"account_type_code"`
	// 口座番号 半角数字 7桁未満の番号は右詰で、前ゼロで埋めること
	AccountNumber string `json:"account_number"`
	// 受取人名 半角文字 振込許容文字を指定可 ただし、一部の非許容文字は、許容文字に変換を行います
	BeneficiaryName string `json:"beneficiary_name"`
}
