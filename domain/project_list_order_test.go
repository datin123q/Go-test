package domain

import (
	"context"
	"fmt"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	xormlog "xorm.io/xorm/log"
)

// sqlLogger は xorm が組み立てた SQL を控える。ファンド一覧の SELECT は
// 相関サブクエリと now() を含み sqlite では実行できないため、並びの検証は
// 発行された SQL の中身で行う。
type sqlLogger struct {
	statements []string
}

func (l *sqlLogger) Debug(v ...interface{})                 { l.record(v...) }
func (l *sqlLogger) Debugf(format string, v ...interface{}) { l.recordf(format, v...) }
func (l *sqlLogger) Error(v ...interface{})                 {}
func (l *sqlLogger) Errorf(format string, v ...interface{}) {}
func (l *sqlLogger) Info(v ...interface{})                  { l.record(v...) }
func (l *sqlLogger) Infof(format string, v ...interface{})  { l.recordf(format, v...) }
func (l *sqlLogger) Warn(v ...interface{})                  {}
func (l *sqlLogger) Warnf(format string, v ...interface{})  {}
func (l *sqlLogger) Level() xormlog.LogLevel                { return xormlog.LOG_DEBUG }
func (l *sqlLogger) SetLevel(lv xormlog.LogLevel)           {}
func (l *sqlLogger) ShowSQL(show ...bool)                   {}
func (l *sqlLogger) IsShowSQL() bool                        { return true }

func (l *sqlLogger) record(v ...interface{}) { l.statements = append(l.statements, fmt.Sprint(v...)) }
func (l *sqlLogger) recordf(f string, v ...interface{}) {
	l.statements = append(l.statements, fmt.Sprintf(f, v...))
}

func (l *sqlLogger) joined() string { return strings.Join(l.statements, "\n") }

type projectOrderRow struct {
	Id int64 `xorm:"pk autoincr 'id'"`
}

func (projectOrderRow) TableName() string { return "project" }

func setupProjectListOrder(t *testing.T) (*xorm.Engine, *sqlLogger) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = engine.Close() })

	require.NoError(t, engine.Sync2(new(projectOrderRow)))

	logger := &sqlLogger{}
	engine.SetLogger(logger)
	engine.ShowSQL(true)

	return engine, logger
}

// 一覧は募集ステータスの決まった順で並べる。投資家が追う流れに沿わせ、
// 不成立を最後に置く。並びが崩れると、募集中のファンドが埋もれる。
func TestProjectStatusOrderExpr(t *testing.T) {
	expr := ProjectStatusOrderExpr()

	t.Run("(成功) 指定された順に並ぶ", func(t *testing.T) {
		want := []string{
			"'STANDBY_FUNDING'",
			"'OPEN_FUNDING'",
			"'END_FUNDING'",
			"'SUCCESS_FUNDING'",
			"'OPERATION_FUNDING'",
			"'OPERATION_END_FUNDING'",
			"'CLOSE_FUNDING'",
			"'FAIL_FUNDING'",
		}

		prev := -1
		for _, status := range want {
			at := strings.Index(expr, status)
			require.NotEqual(t, -1, at, "%s が式に無い: %s", status, expr)
			assert.Greater(t, at, prev, "%s の位置が前後している", status)
			prev = at
		}
	})

	// 一覧から除外しているが、万一混ざったときに先頭を占めないようにする。
	t.Run("(成功) 無効は並びに含めない", func(t *testing.T) {
		assert.NotContains(t, expr, string(INVALID_FUNDING))
	})

	// FIELD は並びに無い値へ 0 を返す。昇順では 0 が先頭に来てしまう。
	t.Run("(成功) 並びに無い値を末尾へ回す", func(t *testing.T) {
		assert.Contains(t, expr, "= 0 ASC")
	})

	t.Run("(成功) テーブル名を修飾する", func(t *testing.T) {
		assert.Contains(t, expr, "project.project_status")
	})
}

// 募集ステータスを最優先にし、同じステータスの中は登録日時の新しい順にする。
// 画面から届く sortby は使わない。ここが崩れると、募集中のファンドが埋もれる。
func TestSetProjectListOrder(t *testing.T) {
	engine, logger := setupProjectListOrder(t)

	session := engine.Table("project").Where("1 = 1")
	SetProjectListOrder(session)

	rows := make([]projectOrderRow, 0)
	_ = session.Find(&rows)

	sql := logger.joined()

	// 見つからないまま切り出すと範囲外で落ち、何が起きたのか分からなくなる。
	at := strings.Index(sql, "ORDER BY")
	require.NotEqual(t, -1, at, "ORDER BY が発行されていない: %s", sql)

	orderBy := sql[at:]

	t.Run("(成功) ステータスを最優先に積む", func(t *testing.T) {
		status := strings.Index(orderBy, "FIELD(")
		created := strings.Index(orderBy, "project.created DESC")

		require.NotEqual(t, -1, status, "ステータスの並びが無い: %s", orderBy)
		require.NotEqual(t, -1, created, "登録日時の並びが無い: %s", orderBy)
		assert.Less(t, status, created, "ステータスは登録日時より前に来る")
	})

	t.Run("(成功) 登録日時を二重に積まない", func(t *testing.T) {
		assert.Equal(t, 1, strings.Count(orderBy, "project.created DESC"))
	})

	// 進捗率や公開日を後ろに積んでも、登録日時でほぼ一意に決まるため到達しない。
	// 残すと効いているように見えるだけなので、積まないことを固定する。
	t.Run("(成功) 画面から届く並び替えは積まない", func(t *testing.T) {
		assert.NotContains(t, orderBy, "CASE WHEN", "進捗率の式が混ざっている")
		assert.NotContains(t, orderBy, "start_at", "公開日が混ざっている")
		assert.NotContains(t, orderBy, "end_at", "募集終了日が混ざっている")
	})

	// ORDER BY はステータスと登録日時の3項目だけで終わる。
	t.Run("(成功) 並びはステータスと登録日時だけ", func(t *testing.T) {
		assert.Equal(t, 3, strings.Count(orderBy, " ASC")+strings.Count(orderBy, " DESC"),
			"想定外のキーが積まれている: %s", orderBy)
	})
}

// ApplyProjectListSortOrder は project_status のときだけ固定順にし、他は
// papu-nika さんの指摘(MIR-472 PR #503)を受けて元の個別対応へ戻した経路。
// ここが崩れると、募集開始日順や進捗率順で見せていた画面が意図せずステータス順に
// 変わってしまう。
func TestApplyProjectListSortOrder(t *testing.T) {
	tests := []struct {
		name    string
		sortby  string
		order   string
		wantIn  []string
		wantOut []string
	}{
		{
			name:   "project_status なら募集ステータス→登録日時の固定順",
			sortby: "project_status",
			order:  "",
			wantIn: []string{"FIELD(", "project.created DESC"},
		},
		{
			name:   "public_view_share_amount なら調達金額順",
			sortby: "public_view_share_amount",
			order:  "",
			wantIn: []string{"CASE WHEN", "DESC"},
			wantOut: []string{
				"* 100)", // 調達金額そのもの。進捗率(% 換算)の式とは別物。
			},
		},
		{
			name:   "public_progress_invest_percent なら進捗率順",
			sortby: "public_progress_invest_percent",
			order:  "",
			wantIn: []string{"CASE WHEN", "* 100)", "ASC"},
		},
		{
			name:   "end_at なら募集終了日順",
			sortby: "end_at",
			order:  "",
			wantIn: []string{"project.end_at"},
		},
		{
			name:   "created なら登録日時の新しい順",
			sortby: "created",
			order:  "",
			wantIn: []string{"project.created DESC"},
		},
		{
			// 元の挙動のまま無修飾で汎用フォールバックへ渡すことを固定する。
			// project_view_count を LEFT JOIN しているため、id・updated のように
			// 両テーブルに存在する列名だと理論上 ambiguous column になりうるが、
			// これは MIR-472 より前から存在する挙動であり、このチケットの
			// スコープ外(詳細はApplyProjectListSortOrderのコメント参照)。
			name:   "想定外の値は汎用フォールバックへ無修飾のまま渡す(元の挙動)",
			sortby: "id",
			order:  "asc",
			wantIn: []string{"`id` ASC"},
		},
		{
			// modelhelper.SetSortOrder は空文字列だと何も積まずに返るため、
			// ORDER BY が一切付かない。これも元の挙動のまま(詳細は
			// ApplyProjectListSortOrder のコメント参照)。
			name:    "sortby が空なら元の挙動のまま何も積まない",
			sortby:  "",
			order:   "",
			wantOut: []string{"ORDER BY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, logger := setupProjectListOrder(t)

			session := engine.Table("project").Where("1 = 1")
			ApplyProjectListSortOrder(context.Background(), session, tt.sortby, tt.order)

			rows := make([]projectOrderRow, 0)
			_ = session.Find(&rows)

			sql := logger.joined()

			for _, want := range tt.wantIn {
				assert.Contains(t, sql, want, "発行されたSQL: %s", sql)
			}
			for _, notWant := range tt.wantOut {
				assert.NotContains(t, sql, notWant, "発行されたSQL: %s", sql)
			}
		})
	}
}
