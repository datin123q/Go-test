package factory

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const WrapErrorMessage = "error"

type stackTracer interface {
	StackTrace() errors.StackTrace
	Error() string
}

type Error struct {
	Code       int    `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	Details    string `json:"details,omitempty"`
	err        error
	status     int
	stacktrace errors.StackTrace
}

func (t ErrorTemplate) New(err error, v ...interface{}) *Error {
	e := Error{
		Code:    t.Code,
		Message: fmt.Sprintf(t.Message, v...),
		err:     errors.Wrap(err, t.Message),
		status:  t.status,
	}
	if err != nil {
		e.Details = fmt.Sprintf("%s: %s", WrapErrorMessage, err.Error())
	}

	ie, ok := errors.Cause(err).(stackTracer)
	if ok {
		e.stacktrace = ie.StackTrace()
	}

	return &e
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Details == "" {
		return e.Message
	}
	return e.Details
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func (e *Error) Status() int {
	if e == nil || e.status == 0 {
		return http.StatusInternalServerError
	}
	return e.status
}

func (e *Error) StackTrace() errors.StackTrace {
	return e.stacktrace
}

type ErrorTemplate Error

func (e *ErrorTemplate) Status() int {
	return e.status
}

// statusが指定されていないコードは500 StatusInternalServerErrorに指定される。
var (
	// System Error
	ErrorSystem                = ErrorTemplate{Code: 10001, Message: "System Error"}
	ErrorServiceUnavailable    = ErrorTemplate{Code: 10002, Message: "Service unavailable"}
	ErrorRemoteService         = ErrorTemplate{Code: 10003, Message: "Remote service error"}
	ErrorIPLimit               = ErrorTemplate{Code: 10004, Message: "ClientIp limit"}
	ErrorPermissionDenied      = ErrorTemplate{Code: 10005, Message: "Permission denied", status: http.StatusForbidden}
	ErrorIllegalRequest        = ErrorTemplate{Code: 10006, Message: "Illegal request", status: http.StatusBadRequest}
	ErrorHTTPMethod            = ErrorTemplate{Code: 10007, Message: "HTTP method is not supported for this request", status: http.StatusMethodNotAllowed}
	ErrorParameter             = ErrorTemplate{Code: 10008, Message: "Parameter error", status: http.StatusBadRequest}
	ErrorMissParameter         = ErrorTemplate{Code: 10009, Message: "Miss required parameter", status: http.StatusBadRequest}
	ErrorDB                    = ErrorTemplate{Code: 10010, Message: "DB error, please contact the administrator", status: http.StatusInternalServerError}
	ErrorTokenInvalid          = ErrorTemplate{Code: 10011, Message: "Token invalid", status: http.StatusUnauthorized}
	ErrorMissToken             = ErrorTemplate{Code: 10012, Message: "Miss token", status: http.StatusUnauthorized}
	ErrorVersion               = ErrorTemplate{Code: 10013, Message: "API version %s invalid"}
	ErrorNotFound              = ErrorTemplate{Code: 10014, Message: "Resource not found", status: http.StatusNotFound}
	ErrorInvalidForm           = ErrorTemplate{Code: 10015, Message: "Invalid Form", status: http.StatusBadRequest}
	ErrorOverMaxPin            = ErrorTemplate{Code: 10016, Message: "Static pin more than max", status: http.StatusBadRequest}
	ErrorNotFoundRequest       = ErrorTemplate{Code: 10017, Message: "Not Found Have Request"}
	ErrorCanNotDeleteSuperUser = ErrorTemplate{Code: 10018, Message: "Can't Delete Supper User"}
	ErrorFailDownloadCSV       = ErrorTemplate{Code: 10019, Message: "Fail to Download CSV File ", status: http.StatusInternalServerError}
	ErrorFailDownloadPDF       = ErrorTemplate{Code: 10020, Message: "Fail to Download PDF File", status: http.StatusInternalServerError}
	ErrorTokenExpired          = ErrorTemplate{Code: 10021, Message: "Token expired", status: http.StatusUnauthorized}
	ErrorFailContactMailSend   = ErrorTemplate{Code: 10022, Message: "Fail Contact Mail Send", status: http.StatusInternalServerError}
	ErrorJSONMarshal           = ErrorTemplate{Code: 10023, Message: "Fail JSON Marshal", status: http.StatusInternalServerError}
	ErrorAlreadyExists         = ErrorTemplate{Code: 10024, Message: "Already Exists", status: http.StatusBadRequest}
	ErrorCsrfInvalid           = ErrorTemplate{Code: 10025, Message: "csrf token invalid", status: http.StatusUnauthorized}
	ErrorTooManyRequests       = ErrorTemplate{Code: 10026, Message: "Too Many Requests", status: http.StatusTooManyRequests}
	ErrorStreamProcessing      = ErrorTemplate{Code: 10027, Message: "Stream processing error, please contact the administrator", status: http.StatusInternalServerError}
	// project.TryLockById(MIR-1204)と同じ「実行中の別プロセスとぶつかったので
	// 弾いた」ことを表す。呼び出し元は成功として扱わず、はっきり失敗として
	// 利用者へ伝える(MIR-1194)。
	ErrJobExecutionLocked = ErrorTemplate{Code: 10028, Message: "Job Execution Locked", status: http.StatusConflict}
	// MIR-1068: 別々の待機者への同時繰上げでMySQLがデッドロックを検出したことを表す
	// (clean_arch_domain.ErrConcurrentUpdate参照)。データは壊れていないので、
	// もう一度実行すれば通常は成功することを利用者へ伝える。
	ErrorConcurrentUpdate = ErrorTemplate{Code: 10029, Message: "Update conflicted with another operation, please try again", status: http.StatusConflict}

	// System Error - Repayment Dividend
	ErrParameterRepaymentAmountLogic   = ErrorTemplate{Code: 10101, Message: "Parameter Repayment Amount Logic Error", status: http.StatusBadRequest}
	ErrParameterDividendAmountLogic    = ErrorTemplate{Code: 10102, Message: "Parameter Dividend Amount Logic Error", status: http.StatusBadRequest}
	ErrParameterRepaymentAmountExceeds = ErrorTemplate{Code: 10103, Message: "Parameter Repayment Amount Exceeds Error", status: http.StatusBadRequest}
	ErrExecuteInvestmentLocked         = ErrorTemplate{Code: 10104, Message: "Execute Investment Locked", status: http.StatusConflict}
	// Messageは"%s"にし、呼び出し側がerr.Error()(usecaseで組み立てた日本語の
	// 利用者向け文言)をそのまま渡す。画面はこのmessageフィールドを表示する。
	ErrExecuteInvestmentDateMismatch   = ErrorTemplate{Code: 10105, Message: "%s", status: http.StatusBadRequest}
	ErrExecuteInvestmentAmountMismatch = ErrorTemplate{Code: 10106, Message: "%s", status: http.StatusBadRequest}
	ErrInvestmentExecutionIncomplete   = ErrorTemplate{Code: 10107, Message: "%s", status: http.StatusBadRequest}
	// ErrCannotManuallyChangeInvestStatus は出資の状態を手で変えてよい遷移元でないときに返す(MIR-1295)。10107はMIR-1239が
	// 使っているため10108を使う。
	ErrCannotManuallyChangeInvestStatus = ErrorTemplate{Code: 10108, Message: "%s", status: http.StatusBadRequest}

	// User Error
	ErrorUserNotExists    = ErrorTemplate{Code: 20001, Message: "User does not exists", status: http.StatusUnauthorized}     // 存在しないユーザー
	ErrorFailAuthorized   = ErrorTemplate{Code: 20002, Message: "Fail Authentication", status: http.StatusUnauthorized}      // 認証失敗
	ErrorOverLimitInvest  = ErrorTemplate{Code: 20003, Message: "Investment limit Over User", status: http.StatusBadRequest} // 投資限度超過
	ErrorBlockInvest      = ErrorTemplate{Code: 20004, Message: "Investment Block User"}                                     // 投資制限ユーザー
	ErrorSameEmailExists  = ErrorTemplate{Code: 20005, Message: "Same Email Exists", status: http.StatusBadRequest}          // 同じメールアドレスが存在
	ErrorFailSmsCert      = ErrorTemplate{Code: 20006, Message: "Fail Sms Cert", status: http.StatusUnauthorized}            // SMS認証に失敗した
	ErrorNotMatchPassword = ErrorTemplate{Code: 20007, Message: "Not Match Password", status: http.StatusUnauthorized}       // パスワード不一致
	ErrorSameCorpExists   = ErrorTemplate{Code: 20008, Message: "The same Corp No exists", status: http.StatusBadRequest}
	ErrorNotMatchAudit    = ErrorTemplate{Code: 20009, Message: "Not Match Audit Status", status: http.StatusBadRequest} // 会員登録未承認ユーザー
	ErrorCorpUser         = ErrorTemplate{Code: 20010, Message: "Corp User", status: http.StatusBadRequest}              // 法人会員として登録されたユーザー(ソーシャル会員登録、ログイン時)
	ErrorDeletedUser      = ErrorTemplate{Code: 20011, Message: "Deleted User", status: http.StatusBadRequest}           // 削除されたユーザー(30日以内に削除されたユーザー)
	ErrorExpireRequest    = ErrorTemplate{Code: 20012, Message: "Expired Request", status: http.StatusBadRequest}        // 有効期限切れのパスワード初期化リクエスト
	ErrorPrivateUser      = ErrorTemplate{Code: 20014, Message: "Registered Private User", status: http.StatusBadRequest}
	ErrorCodeMismatch     = ErrorTemplate{Code: 20015, Message: "Code Mismatch", status: http.StatusBadRequest} // コード不一致
	ErrorNoWithdrawal     = ErrorTemplate{Code: 20017, Message: "No Withdrawal", status: http.StatusBadRequest} // 退会不可状態
	ErrorSocialUser       = ErrorTemplate{Code: 20019, Message: "SocialUser", status: http.StatusBadRequest}
	ErrorNotHaveRole      = ErrorTemplate{Code: 20020, Message: "You don’t have required role or privilege", status: http.StatusForbidden}
	ErrorNotMatchOwner    = ErrorTemplate{Code: 20021, Message: "Not Match Owner", status: http.StatusForbidden}        // リソースオーナーが異なります
	ErrorInvalidityAudit  = ErrorTemplate{Code: 20022, Message: "Invalidity Audit User", status: http.StatusBadRequest} // 会員登録状態無効ユーザー

	ErrorNotSupportedFiles        = ErrorTemplate{Code: 20023, Message: "Not Support Files", status: http.StatusInternalServerError}            // アップロード時にサポートされていないファイル形式
	ErrorCorporateCodeMismatch    = ErrorTemplate{Code: 20024, Message: "Corporate User Code Mismatch", status: http.StatusBadRequest}          // コード不一致
	ErrorMyNumberNotRegistered    = ErrorTemplate{Code: 20025, Message: "MyNumber Not Registered", status: http.StatusInternalServerError}      // マイナンバー未登録
	ErrorSamePassword             = ErrorTemplate{Code: 20026, Message: "Cannot replace with the same password", status: http.StatusBadRequest} // 同じパスワード
	ErrorSentEmail                = ErrorTemplate{Code: 20027, Message: "email already sent", status: http.StatusBadRequest}                    // 既に送信されたメール
	ErrorMaxTransmissionsReached  = ErrorTemplate{Code: 20028, Message: "Max transmissions reached", status: http.StatusBadRequest}             // 最大送信回数超過
	ErrorRequestTimeNotExpired    = ErrorTemplate{Code: 20029, Message: "Request time not expired", status: http.StatusBadRequest}              // リクエスト時刻がまだ有効期限切れになっていない
	ErrorInvalidAuditStatus       = ErrorTemplate{Code: 20030, Message: "Invalid Audit Status", status: http.StatusBadRequest}                  // 無効な会員状態
	ErrorMyNumberAlreadyExist     = ErrorTemplate{Code: 20031, Message: "MyNumber Already Exist", status: http.StatusBadRequest}                // 既に登録されたマイナンバー
	ErrorInvalidAssetAndIncome    = ErrorTemplate{Code: 20032, Message: "Invalid Assets And Income", status: http.StatusBadRequest}             // 無効な資産および所得
	ErrorSignUpEmailMismatch      = ErrorTemplate{Code: 20033, Message: "SignUp Email Mismatch", status: http.StatusBadRequest}                 // 会員登録メール不一致
	ErrorInvalidPassword          = ErrorTemplate{Code: 20034, Message: "Invalid Password", status: http.StatusBadRequest}                      // 無効なパスワード
	ErrorInvalidUserType          = ErrorTemplate{Code: 20035, Message: "Invalid User Type", status: http.StatusBadRequest}                     // 無効なユーザータイプ
	ErrorUserLocked               = ErrorTemplate{Code: 20036, Message: "User locked", status: http.StatusLocked}                               // アカウントロック
	ErrorProfileIncompleteForEkyc = ErrorTemplate{Code: 20037, Message: "Profile Incomplete For Ekyc", status: http.StatusBadRequest}           // eKYC送信に必要な氏名・生年月日・住所が未入力

	// Project
	ErrorInvalidProject  = ErrorTemplate{Code: 30001, Message: "Invalid Project", status: http.StatusBadRequest}
	ErrorNotMatchProject = ErrorTemplate{Code: 30002, Message: "Not Match Project", status: http.StatusBadRequest}
	ErrorNotOpenFunding  = ErrorTemplate{Code: 30003, Message: "Not Open Funding Project", status: http.StatusBadRequest}
	ErrorInsufficientQty = ErrorTemplate{Code: 30004, Message: "Insufficient quantity", status: http.StatusBadRequest}

	ErrorNotFoundRepaymentSchedule  = ErrorTemplate{Code: 30005, Message: "Not Have SettlementSchedule ", status: http.StatusBadRequest}
	ErrorNotPostedProject           = ErrorTemplate{Code: 30006, Message: "Not Posted Project", status: http.StatusBadRequest}
	ErrorOverlapSchedule            = ErrorTemplate{Code: 30007, Message: "Invalid Settlement Schedule", status: http.StatusBadRequest}
	ErrorInvalidDividendDate        = ErrorTemplate{Code: 30008, Message: "Invalid Dividend Date", status: http.StatusBadRequest}
	ErrorIsSettledRepaymentSchedule = ErrorTemplate{Code: 30009, Message: "Repayment Schedule Is Settled ", status: http.StatusBadRequest}
	ErrorInvalidRepaymentSchedule   = ErrorTemplate{Code: 30010, Message: "Invalid Repayment Schedule  ", status: http.StatusBadRequest}
	// MIR-1314: 画面を開いている間にジョブや別の管理者が状態を変えていたことを表す。
	// 状態変更API・ファンド詳細の保存(募集再開)の両方で、意図した状態変更が
	// 実現できなかった(条件付き更新の影響行数0件、または保存前の事前チェック)
	// ときに返す。
	ErrorProjectStatusChanged = ErrorTemplate{Code: 30011, Message: "ファンドの状態が変わったため、変更できませんでした。画面を再読み込みしてください。", status: http.StatusConflict}
	// MIR-272: 運用終了にできない(出資実行が済んでいない出資が残っている)。
	ErrorProjectOperationEndBlocked = ErrorTemplate{Code: 30012, Message: "成立・申込中の出資が残っているため、運用終了にできません。出資実行を完了するか、申込中の出資を取り消してください。", status: http.StatusBadRequest}
	// MIR-272: 償還完了にする前に確認が要る(元本が残っている、または分配・償還の予定が1件も無い)。画面の確認に
	// 同意すると進められる。
	ErrorProjectCloseNeedsConfirmation = ErrorTemplate{Code: 30013, Message: "元本が残っている、または分配・償還の予定が無いため、確認が必要です。", status: http.StatusConflict}
	// MIR-272: 償還完了にできない(未実施の分配・償還がある)。
	ErrorProjectCloseBlockedBySchedule = ErrorTemplate{Code: 30014, Message: "未実施の分配・償還があるため、償還完了にできません。分配・償還が完了してから進めてください。", status: http.StatusBadRequest}
	// MIR-1492: ファンド詳細を開いた後に、別の保存(別の管理者・同じ管理者の別タブなど)が
	// あったことを表す。画面は確認ダイアログを出し、上書きを選んだ場合は
	// force_overwrite を付けて送り直す。コードは ZOLOME・IKKAQ でも空いている
	// 30015 にそろえる。
	ErrorProjectDetailUpdatedAfterLoad = ErrorTemplate{Code: 30015, Message: "ファンド詳細を開いた後に、ファンド情報が更新されています。", status: http.StatusConflict}

	// Aws Error
	ErrorAWSUpload   = ErrorTemplate{Code: 40001, Message: "S3 Upload Fail", status: http.StatusInternalServerError}
	ErrorAWSDelete   = ErrorTemplate{Code: 40002, Message: "S3 Delete Fail", status: http.StatusInternalServerError}
	ErrorAWSSendMail = ErrorTemplate{Code: 40003, Message: "Fail Send Email", status: http.StatusInternalServerError}
	ErrorAWSSendSMS  = ErrorTemplate{Code: 40004, Message: "Fail Send SMS", status: http.StatusInternalServerError}
	ErrorAWSGet      = ErrorTemplate{Code: 40005, Message: "S3 Get Fail", status: http.StatusInternalServerError}

	// Invest
	ErrorFailInvest          = ErrorTemplate{Code: 50001, Message: "Fail Invest", status: http.StatusInternalServerError}                // 投資に失敗しました
	ErrorFailCancelInvest    = ErrorTemplate{Code: 50002, Message: "Fail Cancel Invest Project", status: http.StatusInternalServerError} // 投資キャンセルの場合
	ErrorFailRefund          = ErrorTemplate{Code: 50003, Message: "Fail Refund Payment", status: http.StatusInternalServerError}        // 返金に失敗しました
	ErrorOverAgeInvestor     = ErrorTemplate{Code: 50004, Message: "Over Age Investor", status: http.StatusBadRequest}                   // 年齢による投資制限
	ErrorNotMatchInvest      = ErrorTemplate{Code: 50005, Message: "Not Match Order ", status: http.StatusBadRequest}                    // 注文情報が不正確です
	ErrorLessThanMinQuantity = ErrorTemplate{Code: 50006, Message: "Less Than Min Quantity", status: http.StatusBadRequest}              // 最小注文数量不足
	ErrorInsufficientAmount  = ErrorTemplate{Code: 50007, Message: "Insufficient Amount", status: http.StatusBadRequest}
	ErrorOverInvestorCount   = ErrorTemplate{Code: 50008, Message: "Over Investor Count", status: http.StatusBadRequest} // 投資可能人数超過
	ErrorOverLimitedAmount   = ErrorTemplate{Code: 50009, Message: "Over Limited Amount", status: http.StatusBadRequest} // 投資制限金額超過
	ErrorAlreadyWaiting      = ErrorTemplate{Code: 50010, Message: "Already Waiting", status: http.StatusBadRequest}

	// Bank
	ErrorLinkedVirtualAccount      = ErrorTemplate{Code: 60001, Message: "Linked Virtual Account", status: http.StatusInternalServerError}     // 既にリンクされたアカウントです
	ErrorHasVirtualAccount         = ErrorTemplate{Code: 60002, Message: "Has Virtual Account", status: http.StatusInternalServerError}        // 既に仮想口座を所持しています
	ErrorInsufficientBalance       = ErrorTemplate{Code: 60003, Message: "Balance is Insufficient", status: http.StatusInternalServerError}    // 残高が不足しています
	ErrorNotFoundTransaction       = ErrorTemplate{Code: 60004, Message: "Not Found Transaction", status: http.StatusInternalServerError}      // 取引履歴が見つかりません
	ErrorNotFoundVirtualAccount    = ErrorTemplate{Code: 60005, Message: "Not Found Virtual Account", status: http.StatusInternalServerError}  // 仮想口座が見つかりません
	ErrorNoVirtualAccountToLink    = ErrorTemplate{Code: 60006, Message: "No virtual account to link", status: http.StatusInternalServerError} // リンク可能な仮想口座がありません
	ErrorWithdrawalLessThanFee     = ErrorTemplate{Code: 60006, Message: "Withdrawal less than fee", status: http.StatusInternalServerError}
	ErrorWithdrawalRequest         = ErrorTemplate{Code: 60007, Message: "Failed to Withdrawal Request", status: http.StatusInternalServerError} // 出金依頼リクエスト失敗
	ErrorWithdrawalCloseLocked     = ErrorTemplate{Code: 60008, Message: "Withdrawal Close Locked", status: http.StatusConflict}                 // 他のリクエストが同じ締めを処理中(MIR-1175)
	ErrorWithdrawalCloseNotWaiting = ErrorTemplate{Code: 60009, Message: "Withdrawal Close Not Waiting", status: http.StatusConflict}            // 既に一括送金が実行済み・進行中(MIR-1175)
	ErrorWithdrawalCloseNotFound   = ErrorTemplate{Code: 60010, Message: "Withdrawal Close Not Found", status: http.StatusNotFound}              // 締めIDが存在しない(MIR-1175)
	ErrorWithdrawalStatusChanged   = ErrorTemplate{Code: 60011, Message: "Withdrawal Status Already Changed", status: http.StatusConflict}       // 並行リクエストで既にステータスが変わっている(MIR-1250)

	// Csv Error
	ErrorOutOfCsvForm = ErrorTemplate{Code: 70001, Message: "Out of csv form", status: http.StatusBadRequest} // CSV形式に合いません
	ErrorNoRecipients = ErrorTemplate{Code: 70002, Message: "No Recipients", status: http.StatusBadRequest}   // 受信者がいません

	// Contract Error
	ErrorNotCreatedContract = ErrorTemplate{Code: 80001, Message: "Contract Not Created", status: http.StatusInternalServerError} // 契約成立時に書類が作成されません

	// Report Error
	ErrorNotFoundReport    = ErrorTemplate{Code: 90001, Message: "Not Found Report", status: http.StatusNotFound}      // レポートが見つかりません
	ErrorAlreadyOpenReport = ErrorTemplate{Code: 90002, Message: "Already Open Report", status: http.StatusBadRequest} // レポートが既に公開されています
)
