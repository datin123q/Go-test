package domain

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// TestAllowedFromStatusesは、行き先ごとに許す遷移元が設計どおりであることを
// 確認する(MIR-1314)。「募集終了→募集中」(ファンド詳細の保存で終了日時を
// 延長したときの再開)は、状態変更APIの表には含めない(ReopenFundingが
// 別に扱う専用の遷移のため)。
func TestAllowedFromStatuses(t *testing.T) {
	tests := []struct {
		to   PROJECT_STATUS
		want []PROJECT_STATUS
	}{
		{OPEN_FUNDING, []PROJECT_STATUS{STANDBY_FUNDING}},
		{END_FUNDING, []PROJECT_STATUS{OPEN_FUNDING}},
		{SUCCESS_FUNDING, []PROJECT_STATUS{END_FUNDING}},
		{FAIL_FUNDING, []PROJECT_STATUS{END_FUNDING}},
		{OPERATION_FUNDING, []PROJECT_STATUS{SUCCESS_FUNDING}},
		// MIR-272: 運用中→運用終了→償還完了。運用中から償還完了へは直接進めない。
		{OPERATION_END_FUNDING, []PROJECT_STATUS{OPERATION_FUNDING}},
		{CLOSE_FUNDING, []PROJECT_STATUS{OPERATION_END_FUNDING}},
		{INVALID_FUNDING, []PROJECT_STATUS{STANDBY_FUNDING, OPEN_FUNDING, END_FUNDING}},
		{STANDBY_FUNDING, nil},
		{PROJECT_STATUS("GARBAGE"), nil},
	}
	for _, tt := range tests {
		t.Run(string(tt.to), func(t *testing.T) {
			assert.ElementsMatch(t, tt.want, allowedFromStatuses(tt.to))
		})
	}
}

// newProjectStatusTestEngineは、Projectの状態遷移まわり(UpdateStatus・
// ReopenFunding・Update・UpdateForAdminDetail)を検証するための最小限の
// sqliteエンジンを用意する。engine.Sync2でProject構造体からテーブルを作る
// ことで、Update系メソッドが参照する全カラムを漏れなく用意する
// (CLAUDE.mdの「SQLiteテストは本番に存在しないカラムを隠す」の逆向き、
// 「本番にあるがテストに無い」カラム不足を避けるため)。
func newProjectStatusTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})
	require.NoError(t, engine.Sync2(new(Project)))
	return engine
}

func insertTestProject(t *testing.T, engine *xorm.Engine, id int64, status PROJECT_STATUS) {
	t.Helper()
	_, err := engine.Exec(
		"INSERT INTO project (id, common_id, name, category_type, description, owner_id, project_status, min_share_qty) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		id, "common", "name", "category", "desc", 1, string(status), 1,
	)
	require.NoError(t, err)
}

func readProjectStatus(t *testing.T, engine *xorm.Engine, id int64) PROJECT_STATUS {
	t.Helper()
	var p Project
	has, err := engine.ID(id).Get(&p)
	require.NoError(t, err)
	require.True(t, has)
	return p.ProjectStatus
}

// TestUpdateStatus_ConditionalUpdateは、UpdateStatusがDBの現在の状態を条件に
// した更新になっていることを確認する(MIR-1314)。遷移元が合わなければ
// ErrorProjectStatusChangedを返し、実際にDBの値も変わらないこと。
func TestUpdateStatus_ConditionalUpdate(t *testing.T) {
	t.Run("許された遷移元なら更新できる", func(t *testing.T) {
		engine := newProjectStatusTestEngine(t)
		insertTestProject(t, engine, 1, OPEN_FUNDING)
		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

		p := &Project{Id: 1, ProjectStatus: END_FUNDING}
		err := p.UpdateStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, END_FUNDING, readProjectStatus(t, engine, 1))
	})

	t.Run("遷移元が合わなければErrorProjectStatusChangedを返しDBも変わらない", func(t *testing.T) {
		engine := newProjectStatusTestEngine(t)
		// 画面を開いた時点はOPEN_FUNDINGのつもりだったが、実際は既にFAIL_FUNDING
		// (別の管理者が不成立にした)。SUCCESS_FUNDINGへの遷移元として許すのは
		// END_FUNDINGだけなので、FAIL_FUNDINGからは拒否される。
		insertTestProject(t, engine, 1, FAIL_FUNDING)
		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

		p := &Project{Id: 1, ProjectStatus: SUCCESS_FUNDING}
		err := p.UpdateStatus(ctx)
		require.Error(t, err)

		var apiErr *factory.Error
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, factory.ErrorProjectStatusChanged.Code, apiErr.Code)
		// ErrorTemplate.Newの第2引数以降に文言を渡すと、ひな形に
		// プレースホルダーが無いため"%!(EXTRA string=...)"が末尾に付いて
		// 壊れる。ここでメッセージがひな形どおりであることを確認する。
		assert.Equal(t, factory.ErrorProjectStatusChanged.Message, apiErr.Message)
		assert.Equal(t, FAIL_FUNDING, readProjectStatus(t, engine, 1))
	})

	t.Run("行き先が表に無い値なら更新前にnot supported statusを返す", func(t *testing.T) {
		engine := newProjectStatusTestEngine(t)
		insertTestProject(t, engine, 1, OPEN_FUNDING)
		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

		p := &Project{Id: 1, ProjectStatus: STANDBY_FUNDING}
		err := p.UpdateStatus(ctx)
		require.Error(t, err)
		assert.Equal(t, OPEN_FUNDING, readProjectStatus(t, engine, 1), "更新を試みる前に弾かれ、DBは変わらないこと")
	})
}

// TestUpdateStatusForJobは、ジョブが渡す時刻条件がWHERE句に反映されることを
// 確認する(MIR-1314)。一覧取得後にstart_at/end_atが変わっていた場合や、
// 同じジョブが二重に実行された場合を弾くための条件。
func TestUpdateStatusForJob(t *testing.T) {
	t.Run("時刻条件を満たせば更新できる", func(t *testing.T) {
		engine := newProjectStatusTestEngine(t)
		insertTestProject(t, engine, 1, STANDBY_FUNDING)
		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

		p := &Project{Id: 1, ProjectStatus: OPEN_FUNDING}
		err := p.updateStatusForJob(ctx, "id <= ?", int64(1))
		require.NoError(t, err)
		assert.Equal(t, OPEN_FUNDING, readProjectStatus(t, engine, 1))
	})

	t.Run("時刻条件を満たさなければErrorProjectStatusChangedを返す", func(t *testing.T) {
		engine := newProjectStatusTestEngine(t)
		insertTestProject(t, engine, 1, STANDBY_FUNDING)
		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

		p := &Project{Id: 1, ProjectStatus: OPEN_FUNDING}
		err := p.updateStatusForJob(ctx, "id > ?", int64(1))
		require.Error(t, err)

		var apiErr *factory.Error
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, factory.ErrorProjectStatusChanged.Code, apiErr.Code)
		assert.Equal(t, STANDBY_FUNDING, readProjectStatus(t, engine, 1))
	})
}

// TestReopenFundingは、ReopenFundingがEND_FUNDINGのときだけOPEN_FUNDINGへ
// 戻すことを確認する(MIR-1314)。事前の状態チェックはWHERE句だけで行うため、
// DBが既にEND_FUNDING以外(別の管理者が成立にした等)なら影響行数0件で
// falseを返す。
func TestReopenFunding(t *testing.T) {
	t.Run("END_FUNDINGならOPEN_FUNDINGへ戻る", func(t *testing.T) {
		engine := newProjectStatusTestEngine(t)
		insertTestProject(t, engine, 1, END_FUNDING)
		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

		p := &Project{Id: 1}
		ok, err := p.ReopenFunding(ctx)
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, OPEN_FUNDING, readProjectStatus(t, engine, 1))
	})

	t.Run("END_FUNDING以外なら何もせずfalseを返す", func(t *testing.T) {
		engine := newProjectStatusTestEngine(t)
		insertTestProject(t, engine, 1, SUCCESS_FUNDING)
		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

		p := &Project{Id: 1}
		ok, err := p.ReopenFunding(ctx)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, SUCCESS_FUNDING, readProjectStatus(t, engine, 1), "対象外の状態は変えない")
	})
}

// TestProjectUpdate_DoesNotWriteStatusは、丸ごと保存(Update)がproject_status
// を書かないことを確認する(MIR-1314)。並行して動くジョブや状態変更ボタンが
// 変えた状態を、丸ごと保存で古い値に戻してしまう不具合の対策。
func TestProjectUpdate_DoesNotWriteStatus(t *testing.T) {
	engine := newProjectStatusTestEngine(t)
	insertTestProject(t, engine, 1, OPEN_FUNDING)
	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	// 画面は開いた時点の古い状態(STANDBY_FUNDING)を送ってくる想定。
	p := &Project{Id: 1, ProjectStatus: STANDBY_FUNDING, Name: "renamed"}
	err := p.Update(ctx)
	require.NoError(t, err)

	assert.Equal(t, OPEN_FUNDING, readProjectStatus(t, engine, 1), "project_statusは書き換わらないこと")

	var reloaded Project
	has, err := engine.ID(int64(1)).Get(&reloaded)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, "renamed", reloaded.Name, "他のフィールドは通常どおり書き込まれること")
}

// TestProjectUpdateForAdminDetail_DoesNotWriteStatusOrIsWaitInvestは、
// 管理画面のファンド詳細保存専用メソッドが、project_statusに加えて
// is_wait_investも書かないことを確認する(MIR-1314)。画面にこの2項目を
// 編集する欄は無く、開いた時点の値をそのまま送り返しているだけのため。
func TestProjectUpdateForAdminDetail_DoesNotWriteStatusOrIsWaitInvest(t *testing.T) {
	engine := newProjectStatusTestEngine(t)
	_, err := engine.Exec(
		"INSERT INTO project (id, common_id, name, category_type, description, owner_id, project_status, is_wait_invest, min_share_qty) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, "common", "name", "category", "desc", 1, string(OPEN_FUNDING), true, 1,
	)
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	p := &Project{Id: 1, ProjectStatus: STANDBY_FUNDING, IsWaitInvest: false, Name: "renamed"}
	_, err = p.UpdateForAdminDetail(ctx, nil)
	require.NoError(t, err)

	var reloaded Project
	has, err := engine.ID(int64(1)).Get(&reloaded)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, OPEN_FUNDING, reloaded.ProjectStatus, "project_statusは書き換わらないこと")
	assert.True(t, reloaded.IsWaitInvest, "is_wait_investも書き換わらないこと")
	assert.Equal(t, "renamed", reloaded.Name, "他のフィールドは通常どおり書き込まれること")
}

// TestChangeStatus_RejectsWhenCurrentStatusDoesNotMatchは、状態変更ボタンが
// 「成立」を押しても、DBが既に不成立(別の管理者が変えていた)なら拒否され、
// 出資を変えないことを確認する(MIR-1314、設計のテスト要件7章)。
func TestChangeStatus_RejectsWhenCurrentStatusDoesNotMatch(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	require.NoError(t, engine.Sync2(new(Project)))
	_, err = engine.Exec(`
CREATE TABLE share (
    id TEXT PRIMARY KEY,
    project_id INTEGER,
    owner_id INTEGER,
    amount INTEGER,
    invest_status TEXT,
    deleted DATETIME
);
`)
	require.NoError(t, err)

	_, err = engine.Exec(
		"INSERT INTO project (id, common_id, name, category_type, description, owner_id, project_status, min_share_qty, selection_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, "common", "name", "category", "desc", 1, string(FAIL_FUNDING), 1, string(PROJECT_SELECTION_LOTTERY),
	)
	require.NoError(t, err)
	_, err = engine.Exec(
		"INSERT INTO share (id, project_id, owner_id, amount, invest_status) VALUES (?, ?, ?, ?, ?)",
		"share-1", 1, 1, 1000, string(FAIL_INVEST),
	)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	p := &Project{Id: 1, ProjectStatus: FAIL_FUNDING, SelectionType: PROJECT_SELECTION_LOTTERY}
	err = p.ChangeStatus(ctx, SUCCESS_FUNDING, 1)
	require.Error(t, err)

	var apiErr *factory.Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, factory.ErrorProjectStatusChanged.Code, apiErr.Code)

	// 出資(share)のinvest_statusはFAIL_INVESTのまま変わっていないこと。
	var status string
	has, err := engine.Table("share").Select("invest_status").Where("id = ?", "share-1").Get(&status)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, string(FAIL_INVEST), status)
}
