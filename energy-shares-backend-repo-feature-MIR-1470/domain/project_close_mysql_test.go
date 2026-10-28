package domain

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// MIR-272: MySQL の REPEATABLE READ では、同じトランザクション内の通常の SELECT は最初に読んだ
// 時点のスナップショットを返す。運用終了・償還完了の前提チェックが、ファンドの取得のあとに
// 他がコミットした出資・予定を見落とさないこと(FOR UPDATE は最新の行を読む)を確かめる。
// MYSQL_TEST_DSN があるときだけ動く。専用の DB を作ってその中だけで動かし、終わったら消す。
// newMySQLCloseTestEngine は、専用の DB を作って表を用意し、ファンド1(運用中)を入れて返す。終了時に DB を消す。
func newMySQLCloseTestEngine(t *testing.T) (*xorm.Engine, string) {
	t.Helper()
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN が無いので MySQL のテストは飛ばす")
	}
	cfg, err := mysqldriver.ParseDSN(dsn)
	require.NoError(t, err)
	dbName := fmt.Sprintf("mir272_test_%d", time.Now().UnixNano())

	admin, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })
	_, err = admin.Exec("CREATE DATABASE " + dbName)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = admin.Exec("DROP DATABASE IF EXISTS " + dbName) })

	cfg.DBName = dbName
	engine, err := xorm.NewEngine("mysql", cfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { _ = engine.Close() })
	require.NoError(t, engine.Sync2(new(Project), new(Share), new(RepaymentDividendSchedule)))

	_, err = engine.Exec("INSERT INTO project (id, common_id, name, category_type, description, owner_id, project_status, min_share_qty) VALUES (1, 'c', 'n', 'cat', 'd', 1, ?, 1)", string(OPERATION_FUNDING))
	require.NoError(t, err)

	return engine, cfg.FormatDSN()
}

func TestProjectClose_MySQLSnapshot(t *testing.T) {
	engine, _ := newMySQLCloseTestEngine(t)
	var err error

	// 本人のリクエスト(トランザクション)。ファンドの取得でスナップショットが決まったあとに、別の
	// 接続が出資の成立と償還予定をコミットする。
	mine := engine.NewSession()
	defer mine.Close()
	require.NoError(t, mine.Begin())
	ctx := context.WithValue(context.Background(), factory.ConextDBType, mine)
	var warm Project
	_, err = mine.ID(1).Get(&warm)
	require.NoError(t, err)

	_, err = engine.Insert(&Share{Id: "s1", ProjectId: 1, OwnerId: 1, Amount: 1000, Qty: 1, InvestStatus: SUCCESS_INVEST, UpdateId: 1})
	require.NoError(t, err)
	_, err = engine.Exec("INSERT INTO repayment_dividend_schedule (project_id, execute_date, return_type, status, investment_balance) VALUES (1, '2026-10-13 00:00:00', 'REPAYMENT_RETURN', 'RETURN_SCHEDULED', 500)")
	require.NoError(t, err)

	t.Run("ファンドの行のロックは現在の状態を返し、遷移元が違えば拒否する", func(t *testing.T) {
		assert.NoError(t, lockProjectRowForTransition(ctx, 1, OPERATION_END_FUNDING))
		err := lockProjectRowForTransition(ctx, 1, CLOSE_FUNDING)
		require.Error(t, err)
		assert.Equal(t, int(factory.ErrorProjectStatusChanged.Code), apiErrorCode(t, err))
	})

	t.Run("運用終了の前提は、スナップショットのあとに成立した出資を見落とさない", func(t *testing.T) {
		count, err := countInvestmentsBlockingOperationEnd(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("償還完了の前提は、スナップショットのあとに作られた予定を見落とさない", func(t *testing.T) {
		got, err := Project{}.GetCloseCheck(ctx, 1, true)
		require.NoError(t, err)
		assert.Equal(t, CloseCheck{UnexecutedScheduleCount: 1, HasSchedule: true, RemainingPrincipal: 500, ConfirmationRequired: true}, got)
	})

	require.NoError(t, mine.Rollback())
}

// 償還完了の前提チェックが予定を FOR UPDATE で読むとき、project_id の索引が無いと走査した全行
// (別のファンドの予定も)がロックされ、分配償還ジョブの完了処理や予定の作成が待たされる。
// マイグレーション(add_repayment_dividend_schedule_project_index)の索引があれば、そのファンドの
// 行だけがロックされることを確かめる。
func TestProjectClose_MySQLScheduleLockScope(t *testing.T) {
	engine, dsn := newMySQLCloseTestEngine(t)

	_, err := engine.Exec("CREATE INDEX IDX_repayment_dividend_schedule_project_id_status ON repayment_dividend_schedule (project_id, status)")
	require.NoError(t, err)
	for _, pid := range []int{1, 2} {
		for i := 0; i < 20; i++ {
			_, err = engine.Exec("INSERT INTO repayment_dividend_schedule (project_id, execute_date, return_type, status, investment_balance) VALUES (?, '2026-10-13 00:00:00', 'REPAYMENT_RETURN', 'RETURN_COMPLETED', 0)", pid)
			require.NoError(t, err)
		}
	}

	mine := engine.NewSession()
	defer mine.Close()
	require.NoError(t, mine.Begin())
	defer mine.Rollback()
	ctx := context.WithValue(context.Background(), factory.ConextDBType, mine)
	require.NoError(t, lockProjectRowForTransition(ctx, 1, OPERATION_END_FUNDING))
	_, err = Project{}.GetCloseCheck(ctx, 1, true)
	require.NoError(t, err)

	// 別のファンド(2)の予定は、ファンド1の確認中でも更新・作成できる(待ち時間は1秒に縮める)。
	other, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer other.Close()
	other.SetMaxOpenConns(1)
	_, err = other.Exec("SET SESSION innodb_lock_wait_timeout = 1")
	require.NoError(t, err)
	_, err = other.Exec("UPDATE repayment_dividend_schedule SET status = 'RETURN_COMPLETED' WHERE project_id = 2 AND id = (SELECT id FROM (SELECT MIN(id) AS id FROM repayment_dividend_schedule WHERE project_id = 2) t)")
	assert.NoError(t, err, "別のファンドの予定の更新が待たされない")
	_, err = other.Exec("INSERT INTO repayment_dividend_schedule (project_id, execute_date, return_type, status, investment_balance) VALUES (2, '2026-10-14 00:00:00', 'DIVIDEND_RETURN', 'RETURN_SCHEDULED', 0)")
	assert.NoError(t, err, "別のファンドの予定の作成が待たされない")
}
