package domain

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// MIR-1492: ファンド詳細の保存で detail_version が SQL 側で1増えること、
// 構造体が持っていた古い値で書き戻されないことを固定する。古い値で書き戻すと、
// 同時に保存した別の管理者の加算を打ち消し、競合を検知できなくなる。
func TestProject_UpdateForAdminDetail_IncrementsDetailVersion(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	require.NoError(t, engine.Sync(new(Project)))

	_, err = engine.Insert(&Project{Id: 1, Name: "before"})
	require.NoError(t, err)
	_, err = engine.Exec(`UPDATE project SET detail_version = 5 WHERE id = 1`)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	// 画面を開いた時点の値(3)を持ったまま保存しても、DB の値(5)から加算される。
	p := &Project{Id: 1, Name: "after", DetailVersion: 3}
	updated, err := p.UpdateForAdminDetail(ctx, nil)
	require.NoError(t, err)
	assert.True(t, updated)

	var got struct {
		Name          string
		DetailVersion int64
	}
	has, err := engine.SQL(`SELECT name, detail_version FROM project WHERE id = 1`).Get(&got)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, "after", got.Name)
	assert.Equal(t, int64(6), got.DetailVersion)
}

func newDetailVersionTestEngine(t *testing.T) (*xorm.Engine, context.Context) {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})
	require.NoError(t, engine.Sync(new(Project)))
	return engine, context.WithValue(context.Background(), factory.ConextDBType, engine)
}

func readDetailVersion(t *testing.T, engine *xorm.Engine, id int64) (string, int64) {
	t.Helper()
	var got struct {
		Name          string
		DetailVersion int64
	}
	has, err := engine.SQL(`SELECT name, detail_version FROM project WHERE id = ?`, id).Get(&got)
	require.NoError(t, err)
	require.True(t, has)
	return got.Name, got.DetailVersion
}

// MIR-1492: 新しいファンドは 0 から数え、最初の保存で 1 になること。列が NULL の
// ままだと加算しても NULL のままになり、以後の判定ができなくなる。
func TestProject_UpdateForAdminDetail_FirstSaveOfNewProject(t *testing.T) {
	engine, ctx := newDetailVersionTestEngine(t)
	_, err := engine.Insert(&Project{Id: 1, Name: "before"})
	require.NoError(t, err)

	_, version := readDetailVersion(t, engine, 1)
	require.Equal(t, int64(0), version)

	p := &Project{Id: 1, Name: "after"}
	updated, err := p.UpdateForAdminDetail(ctx, lo.ToPtr(int64(0)))
	require.NoError(t, err)
	assert.True(t, updated)

	name, version := readDetailVersion(t, engine, 1)
	assert.Equal(t, "after", name)
	assert.Equal(t, int64(1), version)
}

// MIR-1492: 事前チェックを通ったあとに別の保存が先に確定していた場合(ほぼ同時の
// 保存)、書き込みの時点で弾き、何も書かないこと。
func TestProject_UpdateForAdminDetail_RejectsWhenVersionChanged(t *testing.T) {
	engine, ctx := newDetailVersionTestEngine(t)
	_, err := engine.Insert(&Project{Id: 1, Name: "before"})
	require.NoError(t, err)
	// 開いた時点は 3 だったが、そのあと別の保存で 4 になった。
	_, err = engine.Exec(`UPDATE project SET detail_version = 4 WHERE id = 1`)
	require.NoError(t, err)

	p := &Project{Id: 1, Name: "stale"}
	updated, err := p.UpdateForAdminDetail(ctx, lo.ToPtr(int64(3)))
	require.NoError(t, err)
	assert.False(t, updated)

	name, version := readDetailVersion(t, engine, 1)
	assert.Equal(t, "before", name, "古い内容で書き戻さないこと")
	assert.Equal(t, int64(4), version, "保存回数も進めないこと")
}

// MIR-1492: 開いた時点の値と一致すれば、通常どおり保存して1増やすこと。
func TestProject_UpdateForAdminDetail_SavesWhenVersionMatches(t *testing.T) {
	engine, ctx := newDetailVersionTestEngine(t)
	_, err := engine.Insert(&Project{Id: 1, Name: "before"})
	require.NoError(t, err)
	_, err = engine.Exec(`UPDATE project SET detail_version = 3 WHERE id = 1`)
	require.NoError(t, err)

	p := &Project{Id: 1, Name: "after"}
	updated, err := p.UpdateForAdminDetail(ctx, lo.ToPtr(int64(3)))
	require.NoError(t, err)
	assert.True(t, updated)

	name, version := readDetailVersion(t, engine, 1)
	assert.Equal(t, "after", name)
	assert.Equal(t, int64(4), version)
}

// MIR-1492: 投資・キャンセル待ちなどの処理が使う Project.Update は、ファンドを丸ごと
// 書き戻すが、detail_version は読み込んだ時点の古い値で書き戻さないこと(読み取り専用
// のタグ)。書き戻すと、その間にあった詳細の保存の回数が戻り、競合を検知できなくなる。
func TestProject_Update_DoesNotWriteDetailVersion(t *testing.T) {
	engine, ctx := newDetailVersionTestEngine(t)
	_, err := engine.Insert(&Project{Id: 1, Name: "before"})
	require.NoError(t, err)
	_, err = engine.Exec(`UPDATE project SET detail_version = 5 WHERE id = 1`)
	require.NoError(t, err)

	// 回数が 3 のときに読み込んだファンドを、その後に書き戻す。
	p := &Project{Id: 1, Name: "after", DetailVersion: 3}
	require.NoError(t, p.Update(ctx))

	name, version := readDetailVersion(t, engine, 1)
	assert.Equal(t, "after", name, "他のフィールドは通常どおり書き込まれること")
	assert.Equal(t, int64(5), version)
}
