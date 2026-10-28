package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type LotteryApplicationRepository interface {
	// CreateIfNotExists は、抽選申込を作成します。既に同一ユーザーが同一プロジェクトに抽選申込をしている場合は、挿入せずに false を返します。
	CreateIfNotExists(ctx context.Context, app domain.LotteryApplication) (string, bool, error)
	GetAll(ctx context.Context, options ...LotteryApplicationQueryOption) (int64, []domain.LotteryApplication, error)
	GetWinningQtyTotalsByProjectID(ctx context.Context, projectIDs []int64) (map[int64]int64, error)
	GetById(ctx context.Context, id string) (domain.LotteryApplication, error)
	LockById(ctx context.Context, id string) (domain.LotteryApplication, error)
	// IsAlreadyApplied は、指定されたユーザーとプロジェクトに対してアクティブな抽選申込（キャンセル済みを除く）が存在するかを返します。
	IsAlreadyApplied(ctx context.Context, projectId int64, userId int64) (bool, error)
	// MarkAsWinning は、指定されたIDの抽選申込を当選状態にマークします。キャンセル済みの場合は更新されず false を返します。
	MarkAsWinning(ctx context.Context, id string) (bool, error)
	// MarkAsLost は、指定されたIDの抽選申込を落選状態にマークします。キャンセル済みの場合は更新されず false を返します。
	MarkAsLost(ctx context.Context, id string) (bool, error)
	// MarkAsCanceled は、指定されたIDの抽選申込をキャンセル状態にマークします。当選済み・落選済み・キャンセル済みの場合は更新されず false を返します。
	MarkAsCanceled(ctx context.Context, id string) (bool, error)
}

type LotteryService interface {
	ExportCSV(items []domain.LotteryApplication) CSVData
	ExportCSVForBulkLottery(items []domain.LotteryApplication) CSVData
}

type LotteryApplicationQueryOption = func(*LotteryApplicationQueryOptionFields)

func LotteryApplicationQueryOptionFunc() LotteryApplicationQueryOptionFields {
	return LotteryApplicationQueryOptionFields{}
}

type LotteryApplicationQueryOptionFields struct {
	projectId      *int64
	userId         *int64
	userName       *string
	projectName    *string
	amountMin      *int64
	amountMax      *int64
	lterequestDate *string
	gterequestDate *string
	status         *domain.LOTTERY_STATUS
	QueryOptions
}

func (qo LotteryApplicationQueryOptionFields) WithProjectId(projectId int64) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.projectId = &projectId }
}

func (qo LotteryApplicationQueryOptionFields) WithUserId(userId int64) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.userId = &userId }
}

func (qo LotteryApplicationQueryOptionFields) WithUserName(userName string) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.userName = &userName }
}

func (qo LotteryApplicationQueryOptionFields) WithProjectName(projectName string) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.projectName = &projectName }
}

func (qo LotteryApplicationQueryOptionFields) WithAmountMin(amountMin int64) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.amountMin = &amountMin }
}

func (qo LotteryApplicationQueryOptionFields) WithAmountMax(amountMax int64) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.amountMax = &amountMax }
}

func (qo LotteryApplicationQueryOptionFields) WithLTERequestDate(requestDate string) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.lterequestDate = &requestDate }
}

func (qo LotteryApplicationQueryOptionFields) WithGTERequestDate(requestDate string) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.gterequestDate = &requestDate }
}

func (qo LotteryApplicationQueryOptionFields) WithStatus(status domain.LOTTERY_STATUS) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.status = &status }
}

func (qo LotteryApplicationQueryOptionFields) WithSortBy(sortBy string) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.sortBy = &sortBy }
}

func (qo LotteryApplicationQueryOptionFields) WithOrder(order string) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.order = &order }
}

func (qo LotteryApplicationQueryOptionFields) WithOffset(offset int) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.offset = &offset }
}

func (qo LotteryApplicationQueryOptionFields) WithLimit(limit int) LotteryApplicationQueryOption {
	return func(qo *LotteryApplicationQueryOptionFields) { qo.limit = &limit }
}

func (qo LotteryApplicationQueryOptionFields) ProjectId() *int64 {
	if qo.projectId == nil || *qo.projectId == 0 {
		return nil
	}
	return qo.projectId
}

func (qo LotteryApplicationQueryOptionFields) UserId() *int64 {
	if qo.userId == nil || *qo.userId == 0 {
		return nil
	}
	return qo.userId
}

func (qo LotteryApplicationQueryOptionFields) UserName() *string {
	if qo.userName == nil || *qo.userName == "" {
		return nil
	}
	return qo.userName
}

func (qo LotteryApplicationQueryOptionFields) ProjectName() *string {
	if qo.projectName == nil || *qo.projectName == "" {
		return nil
	}
	return qo.projectName
}

func (qo LotteryApplicationQueryOptionFields) AmountMin() *int64 {
	if qo.amountMin == nil || *qo.amountMin == 0 {
		return nil
	}
	return qo.amountMin
}

func (qo LotteryApplicationQueryOptionFields) AmountMax() *int64 {
	if qo.amountMax == nil || *qo.amountMax == 0 {
		return nil
	}
	return qo.amountMax
}

func (qo LotteryApplicationQueryOptionFields) LTERequestDate() *string {
	if qo.lterequestDate == nil || *qo.lterequestDate == "" {
		return nil
	}
	return qo.lterequestDate
}

func (qo LotteryApplicationQueryOptionFields) GTERequestDate() *string {
	if qo.gterequestDate == nil || *qo.gterequestDate == "" {
		return nil
	}
	return qo.gterequestDate
}

func (qo LotteryApplicationQueryOptionFields) Status() *domain.LOTTERY_STATUS {
	if qo.status == nil || *qo.status == "" {
		return nil
	}
	return qo.status
}
