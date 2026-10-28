package factory

import (
	stderrors "errors"

	"github.com/go-sql-driver/mysql"
)

// IsLockNotAvailableError は、FOR UPDATE NOWAITが「ロック取得できず」で中断された
// ことを示すMySQLエラー(3572)だけを判定する。エンティティに依存しない共通判定
// (MIR-1204のTryLockByIdとMIR-1175のWithdrawalClose.TryLockByIdが共有する)。
func IsLockNotAvailableError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return stderrors.As(err, &mysqlErr) && mysqlErr.Number == 3572
}

// IsDeadlockError は、MySQLがデッドロックを検出してこのトランザクションを中断させた
// ことを示すエラー(1213)だけを判定する。MIR-1068で、繰上げ済み累計の解決
// (LockAndSumAdvancedQty)が別々の待機者に対して同時実行されるとギャップロックが
// 競合し得ることが分かった(敵対的検証で実測・再現。データ破損は無く片方が失敗する
// だけ)。データは壊れておらず、呼び出し元がもう一度実行すれば通常は成功する。
func IsDeadlockError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return stderrors.As(err, &mysqlErr) && mysqlErr.Number == 1213
}
