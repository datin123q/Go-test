package domain

import (
	"context"
	"strings"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

// projectStatusDisplayOrder は一覧に出すときの募集ステータスの並び。
// 投資家が追う流れ（募集前 → 募集 → 結果 → 運用 → 終了）に沿わせ、
// 不成立を最後に置く。INVALID_FUNDING は一覧から除外されるため含めない。
var projectStatusDisplayOrder = []PROJECT_STATUS{
	STANDBY_FUNDING,
	OPEN_FUNDING,
	END_FUNDING,
	SUCCESS_FUNDING,
	OPERATION_FUNDING,
	OPERATION_END_FUNDING,
	CLOSE_FUNDING,
	FAIL_FUNDING,
}

// ProjectStatusOrderExpr は募集ステータスを決まった順に並べる ORDER BY 式を返す。
//
// FIELD は並びに無い値へ 0 を返し、昇順では先頭に来てしまう。一覧から除外して
// いる INVALID_FUNDING が万一混ざったときに先頭を占めないよう、0 を末尾へ
// 回してから並びを見る。
func ProjectStatusOrderExpr() string {
	column := Project{}.TableName() + ".project_status"

	quoted := make([]string, 0, len(projectStatusDisplayOrder))
	for _, status := range projectStatusDisplayOrder {
		quoted = append(quoted, "'"+string(status)+"'")
	}

	field := "FIELD(" + column + "," + strings.Join(quoted, ",") + ")"

	return field + " = 0 ASC, " + field + " ASC"
}

// SetProjectListOrder は、ユーザー向けファンド一覧の並びを組み立てる。
//
// 募集ステータスを最優先にし、同じステータスの中は登録日時の新しい順にする。
// sortby="project_status" のときだけ ApplyProjectListSortOrder から呼ばれる
// (MIR-472)。他の並び(進捗率順・調達金額順など)は従来どおり sortby ごとに
// 個別の ORDER BY を積むため、ここでは決め切らない。
func SetProjectListOrder(q xorm.Interface) {
	q.OrderBy(ProjectStatusOrderExpr())
	q.OrderBy(Project{}.TableName() + ".created DESC")
}

// ApplyProjectListSortOrder は sortby の値に応じて一覧の並びを積む。
//
// project_status のときだけ SetProjectListOrder(募集ステータス→登録日時の
// 固定順)を使う。他の値(空文字列を含む)は元の分岐をそのまま復元しただけで、
// project_status の分岐だけが MIR-472 での追加(papu-nika さんの指摘、PR #503)。
//
// sortby が空、またはいずれにも一致しない値のときは modelhelper.SetSortOrder に
// そのまま渡す(元の挙動のまま。空文字列なら何も積まずに ORDER BY 無しで返る)。
//
// 一度は次の2点を「ついでに直す」形で足したが、いずれも差し戻した。
//   - 空文字列のときも SetProjectListOrder を使うようにする案:
//     sitemap-dynamic.xml.tsx がクエリパラメータ無し(sortby も未指定)でこの
//     一覧を呼んでおり、ページングもせず先頭30件をそのままサイトマップに使う。
//     並びを変えるとサイトマップに載る30件の構成が変わるが、これは SEO 側の
//     判断が要ることで、このチケットのスコープ(削除したソートオプションの
//     復元)を超える。
//   - id・updated のような project と project_view_count(LEFT JOIN 先)の
//     両方に存在する列名だけ SetSortOrderForTable でテーブル修飾する案:
//     sortby はカンマ区切りで複数列を渡せる仕様(modelhelper.QualifySortColumns
//     参照)のため、単純な完全一致の許可リストでは想定外の組み合わせ(例:
//     "view_count,id")を正しく扱えず、直すたびに新しい抜け道が見つかる状態に
//     なった。同じ場所を3周しても収束しなかったため、個別列ごとの解決(project
//     由来か project_view_count 由来か)をここで作り込むのはやめた。
//
// どちらも MIR-472 より前から存在する挙動(sortby 無指定/未知の値のときの
// 経路)であり、新規に持ち込んだ不具合ではない。
func ApplyProjectListSortOrder(ctx context.Context, q xorm.Interface, sortby, order string) {
	switch sortby {
	case "project_status":
		SetProjectListOrder(q)
	case "public_view_share_amount":
		q.OrderBy(PUBLIC_COLUMN_AMOUNT_ORDER)
	case "public_progress_invest_percent":
		q.OrderBy(PUBLIC_PROGRESS_INVEST_PERCENT + " ASC")
	case "end_at":
		q.OrderBy(PUBLIC_COLUMN_END_AT_ORDER)
	case "created":
		q.OrderBy(Project{}.TableName() + ".created DESC")
	default:
		if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
			factory.Logger(ctx).Error(err)
		}
	}
}
