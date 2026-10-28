package clean_arch_domain

import (
	"errors"

	"github.com/samber/lo"
)

var (
	ErrItemNotFound         = errors.New("item not found")
	ErrItemConflict         = errors.New("item conflict")
	ErrBadRequest           = errors.New("bad request")
	ErrRequestLimitExceeded = errors.New("request limit exceeded")
	ErrAlreadyExpired       = errors.New("already expired")
	ErrAlreadyExists        = errors.New("already exists")
	ErrServiceUnavailable   = errors.New("service unavailable")
	ErrAlreadyDeleted       = errors.New("user deleted")
	ErrTokenInvalid         = errors.New("token invalid")
	ErrItemNotAffected      = errors.New("item not affected")
	ErrItemInvalid          = errors.New("invalid item")
	ErrAuditInvalid         = errors.New("invalid audit")
	ErrTooManyRequests      = errors.New("too many requests")
	ErrResourceLocked       = errors.New("resource is locked")
	ErrAlreadyOpenReport    = errors.New("already open report")
	ErrForbidden            = errors.New("forbidden")
	// ErrExecuteInvestmentDateMismatch と ErrExecuteInvestmentAmountMismatch は
	// MIR-1204の分割実行で、既存のproject_investment_executeと異なる
	// execute_date・amountが指定されたことを表す。identity判定用の
	// センチネルで、利用者に見せる文言はUnwrap可能なラッパー型(usecase側)
	// が別途持つ。
	ErrExecuteInvestmentDateMismatch   = errors.New("execute investment date mismatch")
	ErrExecuteInvestmentAmountMismatch = errors.New("execute investment amount mismatch")
	// ErrProfileIncompleteForEkyc はMIR-1121で追加。eKYC送信に必須の氏名・
	// 生年月日・住所のいずれかが未入力のまま本登録の利用規約同意へ進もうと
	// したことを表す。ErrServiceUnavailable(既に資産・年収チェック等で
	// 使われている)を使い回すと、delivery層のエラーコード変換で「資産・年収が
	// 無効」という無関係な文言になってしまうため、専用のセンチネルにする。
	ErrProfileIncompleteForEkyc = errors.New("profile incomplete for ekyc")
	// ErrConcurrentUpdate はMIR-1068で追加。手動の繰上げ(WaitInvest)が別々の待機者に
	// 対して同時実行され、MySQLがデッドロックを検出して片方を中断させたことを表す
	// (敵対的検証で実測・再現)。データは壊れておらず、もう一度実行すれば通常は成功する
	// ため、汎用の500ではなく再試行を促せる専用のセンチネルにする。
	ErrConcurrentUpdate = errors.New("concurrent update conflict")
	// ErrCannotManuallyChangeInvestStatus はMIR-1295で追加。管理画面から
	// 出資の状態を手で変えようとしたが、遷移元・遷移先・出資実行の状態の
	// 組み合わせが許可されていないことを表す。identity判定用のセンチネルで、
	// 利用者に見せる文言はUnwrap可能なラッパー型(usecase側)が別途持つ。
	ErrCannotManuallyChangeInvestStatus = errors.New("cannot manually change invest status")
)

type CSVData [][]string

type QueryOptions struct {
	// 共通オプションフィールド
	sortBy *string
	order  *string
	offset *int
	limit  *int
}

type SearchInput struct {
	Sortby         string `query:"sortby" description:"정렬 기준"`
	Order          string `query:"order" description:"정렬 순서" enum:"asc,desc"`
	SkipCount      int    `query:"skip_count" description:"건너뛰기 개수" format:"gte=0" default:"0" validate:"omitempty,gte=0"`
	MaxResultCount int    `query:"max_result_count" description:"최대 노출 개수" format:"gte=0,lte=100" validate:"omitempty,gte=0,lte=100"`
}

func (qo QueryOptions) SortBy() string {
	return lo.FromPtr(qo.sortBy)
}

func (qo QueryOptions) Order() string {
	return lo.FromPtr(qo.order)
}

func (qo QueryOptions) Offset() int {
	return lo.FromPtr(qo.offset)
}

func (qo QueryOptions) Limit() int {
	return lo.FromPtr(qo.limit)
}

type KsqlSearchInput struct {
	StartDate      *string `query:"start_date" description:"조회 시작일" format:"yyyy-mm-dd HH:mm:ss" example:"2024-11-01 00:00:00"`
	EndDate        *string `query:"end_date" description:"조회 종료일" format:"yyyy-mm-dd HH:mm:ss" example:"2024-11-01 23:59:59"`
	MaxResultCount int     `query:"max_result_count" description:"최대 노출 개수" format:"gte=0,lte=100" validate:"omitempty,gte=0,lte=100"`
}

type KsqlQueryOptions struct {
	startDate *string
	endDate   *string
	limit     int // MaxResultCount
}

func (qo KsqlQueryOptions) StartDate() *string {
	return qo.startDate
}

func (qo KsqlQueryOptions) EndDate() *string {
	return qo.endDate
}

func (qo KsqlQueryOptions) Limit() int {
	return qo.limit
}
