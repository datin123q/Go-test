package clean_arch_domain

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type InvestmentWaitersRepository interface {
	GetAll(ctx context.Context, options ...InvestmentWaitersQueryOption) (int64, []domain.InvestmentWaiters, error)
	GetById(ctx context.Context, id int64) (domain.InvestmentWaiters, error)
	LockById(ctx context.Context, id int64) (domain.InvestmentWaiters, error)
	Update(ctx context.Context, investmentWaiters domain.InvestmentWaiters) (int64, error)
	// LockAndSumAdvancedQty は、待機者に紐づく当選(investment_waiters_winning経由)の
	// share.qty合計(繰上げ済み累計)を解決して返す(MIR-1118のCurrentShareIdをMIR-1068で
	// 複数当選対応に拡張したもの)。investment_waiters_winningをロック読み(FOR UPDATE)する
	// ため、LockByIdと同一トランザクション内で、LockByIdの直後に呼ぶこと。旧列
	// investment_waiters.share_idへのフォールバックはPR3(MIR-1163)で削除した。非ロック読み
	// (通常のSELECT)にしないのは、呼び出し元のトランザクションがLockByIdより前に他のSELECTを
	// 1回でも実行していると、MySQLのREPEATABLE READではその時点のスナップショットに固定され、
	// 直後にLockByIdでFOR UPDATEしても中間テーブル側の非ロック読みだけは古いデータを返し
	// 続けるため(敵対的検証でMySQL 8.0にて実測・再現。ZOLOME側のMIR-1118から移植)。
	LockAndSumAdvancedQty(ctx context.Context, id int64) (int64, error)
	// GetWinningsByWaiterId は、待機者に紐づく当選明細一覧を繰上げを実行した順の新しい順で返す
	// (MIR-1068、管理画面の繰上げ履歴表示)。1待機者ぶんの明細を複数行で見せることが
	// 目的のクエリなので、GetAll/GetByIdと違い実JOINで良い。
	GetWinningsByWaiterId(ctx context.Context, id int64) ([]domain.InvestmentWaitersWinningItem, error)
}
type InvestmentRequestRepository interface {
	Create(ctx context.Context, investmentRequest domain.InvestmentRequest) (int64, error)
	Update(ctx context.Context, investmentRequest domain.InvestmentRequest) error
	GetByInvestmentWaitersId(ctx context.Context, investmentWaitersId int64) (domain.InvestmentRequest, error)
}

type InvestmentWaitersService interface {
	ExportCSV(items []domain.InvestmentWaiters) CSVData
	ChangeDepositRequestStatus(ctx context.Context, investmentWaiters domain.InvestmentWaiters, userId int64) error
	ChangeWaitingEndStatus(ctx context.Context, investmentWaiters domain.InvestmentWaiters, userId int64) error
	ChangeProposalStatus(ctx context.Context, investmentWaiters domain.InvestmentWaiters, userId int64) error
	ChangeCancelledStatus(ctx context.Context, investmentWaiters domain.InvestmentWaiters, userId int64) error
}

type InvestmentWaitersQueryOption = func(*InvestmentWaitersQueryOptionFields)

// InvestmentWaitersQueryOptionFunc Query Option
func InvestmentWaitersQueryOptionFunc() InvestmentWaitersQueryOptionFields {
	return InvestmentWaitersQueryOptionFields{}
}

type InvestmentWaitersQueryOptionFields struct {
	projectId *int64
	QueryOptions
}

func (ivw InvestmentWaitersQueryOptionFields) WithProjectId(projectId int64) func(*InvestmentWaitersQueryOptionFields) {
	return func(option *InvestmentWaitersQueryOptionFields) {
		option.projectId = &projectId
	}
}

func (ivw InvestmentWaitersQueryOptionFields) ProjectId() *string {
	if ivw.projectId == nil || *ivw.projectId == 0 {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%v", *ivw.projectId))
}

func (ivw InvestmentWaitersQueryOptionFields) WithSortBy(sortBy string) func(*InvestmentWaitersQueryOptionFields) {
	return func(ivw *InvestmentWaitersQueryOptionFields) {
		ivw.sortBy = &sortBy
	}
}

func (ivw InvestmentWaitersQueryOptionFields) WithOrder(order string) func(*InvestmentWaitersQueryOptionFields) {
	return func(ivw *InvestmentWaitersQueryOptionFields) {
		ivw.order = &order
	}
}

func (ivw InvestmentWaitersQueryOptionFields) WithOffset(offset int) func(*InvestmentWaitersQueryOptionFields) {
	return func(ivw *InvestmentWaitersQueryOptionFields) {
		ivw.offset = &offset
	}
}

func (ivw InvestmentWaitersQueryOptionFields) WithLimit(limit int) func(*InvestmentWaitersQueryOptionFields) {
	return func(ivw *InvestmentWaitersQueryOptionFields) {
		ivw.limit = &limit
	}
}
