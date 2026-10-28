package domain

import (
	"errors"
)

var (
	ErrParameter = errors.New("parameter error")

	ErrParameterRepaymentAmountLogic   = errors.New("repayment amount logic error")
	ErrParameterDividendAmountLogic    = errors.New("dividend amount and broker-fee  logic error")
	ErrParameterRepaymentAmountExceeds = errors.New("repayment amount exceeds error")
	ErrDB                              = errors.New("database error")

	// ErrVersionConflict は、xormの楽観ロック(version列)を持つ構造体の
	// UPDATEが0行更新に終わった場合に返す(MIR-1247)。errors.Isで判別
	// できるようにするための共有sentinel。
	ErrVersionConflict = errors.New("incorrect account information. or not match version")

	// ErrWithdrawalStatusChanged は、Withdrawal.UpdateStatusIfCurrentが
	// 期待した現在のステータスと一致せず0行更新に終わった場合に返す
	// (MIR-1250)。同じ出金へのステータス変更(例: キャンセル)が並行して
	// 届いても、DB側の条件付きUPDATEで片方だけしか通らないようにする
	// ためのsentinel。errors.Isで判別できる。
	ErrWithdrawalStatusChanged = errors.New("withdrawal status has already been changed")

	// ErrInvestmentExecutionIncomplete は、分配・償還スケジュール作成時に
	// 出資実行(project_investment_execute)が完了していないと判定した場合に
	// 返す(MIR-1239)。出資実行の記録が無い・ExecuteStatusがCOMPLETEDでない・
	// OPERATION_INVEST+REFUND_INVESTの金額合計がexecute.Amountと一致しない、
	// のいずれかに該当する。errors.Isで判別できる。
	ErrInvestmentExecutionIncomplete = errors.New("investment execution is not completed for schedule creation")
)
