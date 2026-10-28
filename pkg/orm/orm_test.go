package orm

import (
	"fmt"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	xormlog "xorm.io/xorm/log"
)

type sortOrderTestRow struct {
	Id      int64  `xorm:"pk autoincr 'id'"`
	Created string `xorm:"created"`
}

func (sortOrderTestRow) TableName() string { return "sort_order_test_row" }

// sqlLogger は xorm が組み立てた SQL を控える。sqlite は JOIN 先と同名の
// 列を ORDER BY に書いても通してしまい、MySQL で出る「あいまい」を
// 再現できない。そのため SQL の中身そのものを確かめる。
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

func setupOrmTestEngine(t *testing.T) (*xorm.Engine, *sqlLogger) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = engine.Close() })

	require.NoError(t, engine.Sync2(new(sortOrderTestRow)))

	logger := &sqlLogger{}
	engine.SetLogger(logger)
	engine.ShowSQL(true)

	return engine, logger
}

// 並び替えの指定は画面から修飾なしで届く。JOIN 先にも同じ名前の列があると
// ORDER BY があいまいになり実行できないため、既定のテーブル名を補う。
func TestOrm_SortOrderForTable(t *testing.T) {
	tests := []struct {
		name    string
		table   string
		sortStr string
		orderBy string
		want    []string
		notWant []string
	}{
		{
			name:    "(成功) 修飾のない項目にテーブル名を補う",
			table:   "sort_order_test_row",
			sortStr: "created",
			orderBy: "desc",
			want:    []string{"`sort_order_test_row`.`created`"},
		},
		{
			name:    "(成功) 複数項目すべてに補う",
			table:   "sort_order_test_row",
			sortStr: "created,id",
			orderBy: "desc,desc",
			want:    []string{"`sort_order_test_row`.`created`", "`sort_order_test_row`.`id`"},
		},
		{
			// 画面が JOIN 先の列で並べたい場合に備え、指定済みは尊重する。
			name:    "(成功) 既に修飾されているものは変えない",
			table:   "sort_order_test_row",
			sortStr: "project.name",
			orderBy: "asc",
			want:    []string{"`project`.`name`"},
			notWant: []string{"`sort_order_test_row`.`project`"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, logger := setupOrmTestEngine(t)

			rows := make([]sortOrderTestRow, 0)
			_ = Table(engine, sortOrderTestRow{}).
				SortOrderForTable(tt.table, tt.sortStr, tt.orderBy).
				Find(&rows)

			// xorm は「テーブル名.列名」を分解して各部をバッククォートで
			// 囲むため、期待値もその形で突き合わせる。
			sql := logger.joined()
			for _, w := range tt.want {
				assert.Contains(t, sql, w)
			}
			for _, nw := range tt.notWant {
				assert.NotContains(t, sql, nw)
			}
		})
	}
}

// 抽選一覧は applied_at と id の2キーで並べる。lottery_application は
// project / user / user_account を JOIN しており、いずれも id を持つため、
// 修飾しないと「あいまい」で落ちる。
func TestOrm_SortOrderForTable_QualifiesAmbiguousId(t *testing.T) {
	engine, logger := setupOrmTestEngine(t)

	rows := make([]sortOrderTestRow, 0)
	_ = Table(engine, sortOrderTestRow{}).
		SortOrderForTable("lottery_application", "applied_at,id", "desc,desc").
		Find(&rows)

	sql := logger.joined()
	assert.Contains(t, sql, "`lottery_application`.`applied_at`")
	assert.Contains(t, sql, "`lottery_application`.`id`")
}

func TestOrm_SortOrder_ValidColumns(t *testing.T) {
	engine, logger := setupOrmTestEngine(t)

	rows := make([]sortOrderTestRow, 0)
	_ = Table(engine, sortOrderTestRow{}).
		SortOrder("created,id", "desc,asc").
		Find(&rows)

	sql := logger.joined()
	assert.Contains(t, sql, "`created` DESC")
	assert.Contains(t, sql, "`id` ASC")
}

// MIR-1113: modelhelper.SetSortOrder と同じ形の実装が pkg/orm にも独立に
// あり、列名の検証が無いまま ORDER BY へ渡していた。/api/v1/faqs 等、
// 未認証で到達できる一覧がここ(clean-arch のリポジトリ層)を通るため、
// 注入を試みる列名は1件でも弾かれ、ORDER BY 自体が積まれないことを確認する。
func TestOrm_SortOrder_RejectsInvalidColumns(t *testing.T) {
	invalidCases := []struct {
		name    string
		sortStr string
	}{
		{"バッククォートとSLEEPによる注入", "id`+SLEEP(5)+`id"},
		{"複数列のうち1つでも不正なら全体を却下する", "created,id`+SLEEP(5)+`id"},
		{"演算子を含む式", "id+1"},
		{"空白を含む式", "id created"},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			engine, logger := setupOrmTestEngine(t)

			rows := make([]sortOrderTestRow, 0)
			_ = Table(engine, sortOrderTestRow{}).
				SortOrder(tc.sortStr, "asc").
				Find(&rows)

			sql := logger.joined()
			assert.NotContains(t, sql, "SLEEP")
			assert.NotContains(t, sql, "ORDER BY")
		})
	}
}

// FAQ一覧は "主キー," + 画面指定のsortbyを連結して渡す。sortby/orderが
// 未指定だと末尾に空要素が残る。空要素は列を指定しないプレースホルダとして
// 読み飛ばすため、既定列(ここでは created)の並び替えはそのまま効くことを
// 確認する。
func TestOrm_SortOrder_SkipsTrailingEmptySegmentButAppliesDefaultColumn(t *testing.T) {
	engine, logger := setupOrmTestEngine(t)

	rows := make([]sortOrderTestRow, 0)
	_ = Table(engine, sortOrderTestRow{}).
		SortOrder("created,", "desc,").
		Find(&rows)

	sql := logger.joined()
	assert.Contains(t, sql, "`created` DESC")
	// 修正前は空要素にもAsc/Descを呼んでいたため、空の識別子(``)が
	// ORDER BYへ追加で積まれて構文エラーになっていた。それが無いことを
	// 確認する(単に"created DESC"を含むだけでは、この回帰を検出できない)。
	assert.NotContains(t, sql, "``")
}

// MIR-1113で見つかった回帰の再現: order だけ指定されて sortby 側が空だと、
// sortByとorderの項目数がずれる("created,"の2項目 vs "desc,asc"の2項目は
// 数としては一致するが、以前はここで既定列の並び替えまで丸ごと消えていた)。
// 既定列の並び替えは効いたままであることを確認する。
func TestOrm_SortOrder_OrderOnlyStillAppliesDefaultColumn(t *testing.T) {
	engine, logger := setupOrmTestEngine(t)

	rows := make([]sortOrderTestRow, 0)
	_ = Table(engine, sortOrderTestRow{}).
		SortOrder("created,", "desc,asc").
		Find(&rows)

	sql := logger.joined()
	assert.Contains(t, sql, "`created` DESC")
	// 修正前は空要素にもAsc/Descを呼んでいたため、空の識別子(``)が
	// ORDER BYへ追加で積まれて構文エラーになっていた。それが無いことを
	// 確認する(単に"created DESC"を含むだけでは、この回帰を検出できない)。
	assert.NotContains(t, sql, "``")
}
