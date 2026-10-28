package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

const projectAmountsTestSchema = `
CREATE TABLE share (
    id TEXT PRIMARY KEY,
    project_id INTEGER,
    amount INTEGER,
    invest_status TEXT,
    deleted DATETIME
);
CREATE TABLE investment_waiters (
    id INTEGER PRIMARY KEY,
    project_id INTEGER,
    amount INTEGER,
    status TEXT,
    share_id TEXT NULL,
    deleted DATETIME
);
CREATE TABLE investment_waiters_winning (
    id INTEGER PRIMARY KEY,
    investment_waiters_id INTEGER,
    share_id TEXT,
    deleted DATETIME
);
`

// TestProject_GetCurrentShareAmountsForUserView_ExcludesWaitersWonViaWinningTable は、
// 中間テーブル(investment_waiters_winning)経由で当選済みと判定される待機者が、待機金額として
// share側の金額と二重に計上されないことを確認する(MIR-1118)。旧列 investment_waiters.share_id
// へのフォールバックはPR3(MIR-1163)で削除したため、この判定は中間テーブルの有無だけで行う。
func TestProject_GetCurrentShareAmountsForUserView_ExcludesWaitersWonViaWinningTable(t *testing.T) {
	tests := []struct {
		name       string
		projectId  int64
		setupSQL   string
		wantAmount int64
	}{
		{
			name:       "未当選の待機者は待機金額に計上されること",
			projectId:  10,
			setupSQL:   `INSERT INTO investment_waiters (id, project_id, amount, status) VALUES (200, 10, 50000, 'WAITING');`,
			wantAmount: 50000,
		},
		{
			// 新コードで繰上げ当選済み(中間テーブルにのみ記録され、investment_waiters.share_id
			// は書き込まれない)。待機金額としては計上されてはならない(share側の金額と
			// 二重計上になるため)。
			name:      "中間テーブル経由の当選はshare側のみ計上され二重計上されないこと",
			projectId: 11,
			setupSQL: `
INSERT INTO share (id, project_id, amount, invest_status) VALUES ('share-2', 11, 70000, 'SUCCESS_INVEST');
INSERT INTO investment_waiters (id, project_id, amount, status) VALUES (201, 11, 70000, 'DEPOSIT_REQUEST');
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id) VALUES (201, 'share-2');
`,
			wantAmount: 70000,
		},
		{
			// 旧列にのみ値が残り(中間テーブルには記録が無い)状態。旧列へのフォールバックは
			// 削除済みのため、中間テーブルに記録が無い以上は待機金額として計上され、対応する
			// shareの金額とも二重に計上される。本番ではバックフィル(db/migrations/20260907000000)
			// 自体がこの状態を解消するため発生しない前提。ここでは関数単体の挙動として、この
			// データに遭遇した場合の結果を確認する。
			name:      "旧列にのみ値が残る待機者はshare側と二重に計上されること",
			projectId: 12,
			setupSQL: `
INSERT INTO share (id, project_id, amount, invest_status) VALUES ('share-3', 12, 30000, 'REQUEST_INVEST');
INSERT INTO investment_waiters (id, project_id, amount, status, share_id) VALUES (202, 12, 30000, 'PROPOSAL', 'share-3');
`,
			wantAmount: 60000,
		},
		{
			// projectAmountにDeletedフィールドが無くxormのsoft-deleteフィルタが自動付与
			// されないため、明示的にNotDeletedConditionを付けて除外する必要がある
			// (敵対的検証で発見)。
			name:       "論理削除済みの待機者は計上されないこと",
			projectId:  13,
			setupSQL:   `INSERT INTO investment_waiters (id, project_id, amount, status, deleted) VALUES (203, 13, 40000, 'PROPOSAL', '2026-08-25 00:00:00');`,
			wantAmount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := xorm.NewEngine("sqlite3", ":memory:")
			require.NoError(t, err)
			engine.SetMaxOpenConns(1)
			t.Cleanup(func() {
				require.NoError(t, engine.Close())
			})

			_, err = engine.Exec(projectAmountsTestSchema)
			require.NoError(t, err)
			_, err = engine.Exec(tt.setupSQL)
			require.NoError(t, err)

			ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
			result, err := Project{}.GetCurrentShareAmountsForUserView(ctx, []int64{tt.projectId})
			require.NoError(t, err)

			assert.Equal(t, tt.wantAmount, result[tt.projectId])
		})
	}
}
