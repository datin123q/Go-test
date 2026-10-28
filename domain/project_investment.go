package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// ProjectInvestmentExecute 대출 실행 정보
type ProjectInvestmentExecute struct {
	Id            int64              `json:"id" xorm:"id pk autoincr" description:"desc(출자 실행 ID)"`
	ProjectId     int64              `json:"project_id" xorm:"project_id" description:"desc(상품 ID)"`
	ExecuteStatus SETTLEMENT_STATUS  `json:"execute_status" xorm:"execute_status varchar(255)" description:"desc(출자 상태)"`
	Amount        int64              `json:"amount" xorm:"amount" description:"desc(출자 금액)"`
	ExecuteDate   jsonutils.JSONTime `json:"execute_date" xorm:"execute_date datetime" description:"desc(출자 실행일자)"`
	Created       jsonutils.JSONTime `json:"created" xorm:"created datetime" description:"desc(생성일)"`
	// PublishFailedCount は法定帳簿へのイベント送信に失敗した累積件数(MIR-1202)。
	// 送信はコミット後(AfterCommit)に行われるため、admin-apiは失敗をHTTPで
	// 返せない。0件なら全件送信済み、1件以上ならis_publish=falseの行が残って
	// おり、再送(MIR-1114のRePublish)で解消されうる。
	PublishFailedCount int64 `json:"publish_failed_count" xorm:"publish_failed_count" description:"desc(법정장부 이벤트 송신 실패 누적 건수)"`
}

func (ProjectInvestmentExecute) TableName() string {
	return "project_investment_execute"
}

func (p ProjectInvestmentExecute) SelectAll() string {
	return p.TableName() + ".*"
}

func (ProjectInvestmentExecute) GetByProjectId(ctx context.Context, id int64) (*ProjectInvestmentExecute, error) {
	var investment ProjectInvestmentExecute
	has, err := factory.DB(ctx).Where("project_id = ?", id).Get(&investment)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &investment, nil
}

func (ProjectInvestmentExecute) GetById(ctx context.Context, id int64) (*ProjectInvestmentExecute, error) {
	var loan ProjectInvestmentExecute
	_, err := factory.DB(ctx).ID(id).Get(&loan)
	if err != nil {
		return nil, err
	}

	return &loan, nil
}

// InvestmentExecutionState は、あるプロジェクトの出資実行
// (project_investment_execute)がどこまで進んでいるかを表す(MIR-1295)。
type InvestmentExecutionState string

const (
	// InvestmentExecutionNone は、project_investment_executeの記録が
	// 1件も無い状態。出資実行(運用開始)を一度も行っていない。
	InvestmentExecutionNone InvestmentExecutionState = "NONE"
	// InvestmentExecutionInProgress は、記録が1件以上あるが、全行が
	// COMPLETEDではない状態(分割実行の途中、または未知の値)。
	InvestmentExecutionInProgress InvestmentExecutionState = "IN_PROGRESS"
	// InvestmentExecutionCompleted は、記録が1件以上あり、全行がCOMPLETEDの状態。
	InvestmentExecutionCompleted InvestmentExecutionState = "COMPLETED"
)

// DetermineInvestmentExecutionState は、あるプロジェクトの
// project_investment_execute全行から出資実行の進捗状態を判定する
// (MIR-1295)。project_idにはユニーク制約が無く複数行が入り得るため、
// 単一行のGetByProjectIdではなく全行を見て判定する。全行がCOMPLETEDの
// ときだけ「完了」とし、SCHEDULED・CONFIRM・空文字・未知の値が1件でも
// あれば安全側に倒して「途中」とする。1件も無ければ「記録なし」。
func DetermineInvestmentExecutionState(executes []ProjectInvestmentExecute) InvestmentExecutionState {
	if len(executes) == 0 {
		return InvestmentExecutionNone
	}
	for _, execute := range executes {
		if execute.ExecuteStatus != COMPLETED {
			return InvestmentExecutionInProgress
		}
	}
	return InvestmentExecutionCompleted
}

// CanManuallyChangeInvestStatus は、管理画面の出資明細画面から出資の状態を
// 手で変えてよいかを判定する(MIR-1295)。許可するのは次の2つだけ:
// 申込中(REQUEST_INVEST)→取消(CANCEL_INVEST)、成立(SUCCESS_INVEST)→
// 取消(CANCEL_INVEST、ただし出資実行が途中でない場合)。それ以外の遷移は
// すべて拒否する。自己募集ファンド(ZOLOME固有。ミラリタは常にfalse)は
// 全遷移を拒否する。引数だけで完結する純粋関数で、DBアクセスは呼び出し側が
// 行う(判定・読み込みはロックのあとに行う必要があるため)。
func CanManuallyChangeInvestStatus(
	from, to INVEST_STATUS,
	execution InvestmentExecutionState,
	isSelfFund bool,
) error {
	if isSelfFund {
		return newCannotChangeInvestStatusError(from, to)
	}
	switch {
	case from == REQUEST_INVEST && to == CANCEL_INVEST:
		return nil
	case from == SUCCESS_INVEST && to == CANCEL_INVEST:
		if execution == InvestmentExecutionInProgress {
			return newCannotChangeInvestStatusError(from, to)
		}
		return nil
	default:
		return newCannotChangeInvestStatusError(from, to)
	}
}
