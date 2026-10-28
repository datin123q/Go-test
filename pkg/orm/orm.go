package orm

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func Table(x xorm.Interface, source interface{}) Orm {
	return &orm{
		source:  source,
		session: x.Table(source),
	}
}

// NotDeletedCond は xorm の内部実装(internal/statements/statement.go の CondDeleted)と
// 同じ「未削除」判定を、生SQL(サブクエリ・JOIN)で使うための条件文字列を返す。ゼロ時刻
// (0001-01-01)はxormの世界では「未削除」を表す値であり、deleted IS NULL だけで判定すると、
// 将来ゼロ時刻が書き込まれる経路ができた場合に誤って「削除済み」扱いしてしまう
// (MIR-516で確認済み)。
func NotDeletedCond(table string) string {
	const zeroDeletedTime = "0001-01-01 00:00:00"
	return fmt.Sprintf("(%[1]s.deleted IS NULL OR %[1]s.deleted = '%[2]s')", table, zeroDeletedTime)
}

type orm struct {
	source  interface{}
	session *xorm.Session
}

type Orm interface {
	Select(args ...interface{}) Orm
	InnerJoin(target interface{}, conditions ...interface{}) Orm
	LeftJoin(target interface{}, conditions ...interface{}) Orm
	SortOrder(sortStr string, orderStr string) Orm
	SortOrderForTable(table string, sortStr string, orderStr string) Orm
	Asc(cols ...string) Orm
	Desc(cols ...string) Orm
	Like(key string, value *string) Orm
	LikeJson(key string, value *string) Orm
	Equal(key string, value interface{}) Orm
	EqualJson(key string, value *string) Orm
	NotEqual(key string, value interface{}) Orm
	IsNull(key string) Orm
	NotNull(key string) Orm
	GTE(key string, value interface{}) Orm
	GT(key string, value interface{}) Orm
	LTE(key string, value interface{}) Orm
	LT(key string, value interface{}) Orm
	In(key string, values ...interface{}) Orm
	NotIn(key string, values ...interface{}) Orm
	Limit(limit int, offset int) Orm
	GroupBy(key string) Orm
	Between(key string, start interface{}, end interface{}) Orm
	Before(key string, date time.Time) Orm
	After(key string, date time.Time) Orm
	BeforeOrEqual(key string, date time.Time) Orm
	AfterOrEqual(key string, date time.Time) Orm
	Unscoped() Orm
	Find(rowSlicePtr interface{}) error
	FindAndCount(rowSlicePtr interface{}) (int64, error)
	Id(id interface{}) Orm
	Get(beans interface{}) (bool, error)
	Count() (int64, error)
	Sum(columnName string) (float64, error)
	Sums(columnNames []string) ([]float64, error)
	SumInt(columnName string) (int64, error)
	SumsInt(columnNames []string) ([]int64, error)
}

func (o *orm) Select(args ...any) Orm {
	var parts []string

	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			parts = append(parts, v)
		case ormObject:
			parts = append(parts, v.SelectAll())
		default:
			// 該当する状態がない場合は無視して進みます。
			return o
		}
	}

	o.session = o.session.Select(strings.Join(parts, ", "))

	return o
}

func (o *orm) join(joinOp string, target any, conditions ...any) Orm {
	// othersは現在の基準で無視して記述します。

	var (
		joinCondition string
		sourceObject  ormObject
		targetObject  ormObject
	)

	sourceObject, ok := toOrmObject(o.source)
	if !ok {
		return o
	}

	targetObject, ok = toOrmObject(target)
	if !ok {
		return o
	}

	if conditions == nil {
		leftPK, err := getPrimaryKey(o.source)
		if err != nil {
			return o
		}

		rightTableFieldName := strings.ToLower(fmt.Sprintf("%s_id", sourceObject.TableName()))

		joinCondition = fmt.Sprintf("%s.%s = %s.%s", sourceObject.TableName(), leftPK, targetObject.TableName(), rightTableFieldName)

	} else {
		joinCondition = buildJoinConditions(conditions)
		if len(joinCondition) == 0 {
			return o
		}
	}

	o.session = o.session.Join(joinOp, targetObject.TableName(), joinCondition)

	return o
}

func (o *orm) InnerJoin(target any, others ...any) Orm {
	return o.join("INNER", target, others...)
}

func (o *orm) LeftJoin(target any, others ...any) Orm {
	return o.join("LEFT", target, others...)
}

// SortOrder applies sorting based on the specified sort and order strings.
// sortStr specifies the fields to sort by, and orderStr specifies the corresponding sort orders.
//
// Usage Examples:
//
//   - Single field sort: SortOrder("name", "asc")
//     This sorts by the "name" field in ascending order.
//
//   - Multiple fields sort with individual orders: SortOrder("name,age", "asc,desc")
//     This sorts by the "name" field in ascending order and then by the "age" field in descending order.
//
//   - Multiple fields sort with a common order: SortOrder("name,age", "asc")
//     This sorts both "name" and "age" fields in ascending order.
func (o *orm) SortOrder(sortStr string, orderStr string) Orm {
	// MIR-1113: 以前はここに modelhelper.SetSortOrder と行単位でほぼ同一の
	// (検証なしにORDER BYへ渡す)実装を独立に持っており、片方だけ直すと
	// 検証内容がずれる状態だった。o.session は *xorm.Session で
	// xorm.Interface を満たすため、本体を委譲する。エラーは元々このメソッドの
	// シグネチャ(Ormを返すだけ)では表現できず、呼び出し元も個別に見ていな
	// かったため、そのまま握りつぶす(不正な入力なら並び替えを行わないだけ)。
	_ = modelhelper.SetSortOrder(o.session, sortStr, orderStr)
	return o
}

// SortOrderForTable は、テーブル名の付いていない並び替え項目に既定の
// テーブル名を補ってから SortOrder を呼ぶ。
//
// JOIN した先にも同じ名前の列があると、修飾なしの ORDER BY は
// 「Column 'id' in order clause is ambiguous」で実行できない。
// 並び替えの指定は画面から来るため、DB のテーブル構成を画面に持ち込まずに
// 済むよう、ここで補う。
//
// 既に「テーブル名.列名」の形で渡されたものはそのまま使う。
func (o *orm) SortOrderForTable(table string, sortStr string, orderStr string) Orm {
	// MIR-1113: SortOrder と同じ理由でmodelhelper.SetSortOrderForTableへ委譲する。
	// ここで QualifySortColumns を自前で呼んで SortOrder に渡す形のままだと、
	// SetSortOrderForTable(helper.go)と同じ組み立てをもう1箇所で行うことに
	// なり、将来どちらか一方だけ直される余地が残る。
	_ = modelhelper.SetSortOrderForTable(o.session, table, sortStr, orderStr)
	return o
}

func (o *orm) Asc(cols ...string) Orm {
	o.session = o.session.Asc(cols...)
	return o
}

func (o *orm) Desc(cols ...string) Orm {
	o.session = o.session.Desc(cols...)
	return o
}

func (o *orm) Like(key string, value *string) Orm {
	if key == "" || value == nil || len(*value) == 0 {
		return o
	}
	queryCond := builder.NewCond().And(builder.Like{key, *value})

	if ok := queryCond.IsValid(); ok {
		o.session = o.session.And(queryCond)
	} else {
		fmt.Errorf("fail query valid")
		return o
	}

	return o
}

func (o *orm) EqualJson(key string, value *string) Orm {
	if key == "" || value == nil || len(*value) == 0 {
		return o
	}

	// MySQLのJSON_CONTAINS関数を使用して、JSONフィールド内の特定の値を検索します。
	queryCond := builder.NewCond().And(builder.Expr(fmt.Sprintf("JSON_CONTAINS(%s, ?)", key), fmt.Sprintf(`"%s"`, *value)))

	if ok := queryCond.IsValid(); ok {
		o.session = o.session.And(queryCond)
	} else {
		fmt.Errorf("fail query valid")
		return o
	}

	return o
}

func (o *orm) LikeJson(key string, value *string) Orm {
	if key == "" || value == nil || len(*value) == 0 {
		return o
	}

	// MySQLのJSON_UNQUOTEおよびLIKE関数を使用して、JSONフィールド内の特定パターンを検索します。
	queryCond := builder.NewCond().And(builder.Expr(fmt.Sprintf("JSON_UNQUOTE(JSON_EXTRACT(%s, '$')) LIKE ?", key), fmt.Sprintf(`%%%s%%`, *value)))

	if ok := queryCond.IsValid(); ok {
		o.session = o.session.And(queryCond)
	} else {
		fmt.Errorf("fail query valid")
		return o
	}

	return o
}

func (o *orm) Equal(key string, value any) Orm {
	if value == nil {
		fmt.Errorf("value is nil")
		return o
	}

	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return o
	}
	if v.Kind() == reflect.String && v.Len() == 0 {
		return o
	}

	var queryCond builder.Cond
	switch t.Kind() {
	case reflect.Ptr:
		queryCond = builder.NewCond().And(builder.Eq{key: v.Elem().Interface()})
	case reflect.String, reflect.Int, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		queryCond = builder.NewCond().And(builder.Eq{key: v.Interface()})
	default:
		fmt.Errorf("unhandled type: %v", t)
		return o
	}

	if ok := queryCond.IsValid(); ok {
		o.session = o.session.And(queryCond)
	} else {
		fmt.Errorf("fail query valid key: %s value %s", key, formatValue(v))
		return o
	}

	return o
}

func (o *orm) NotEqual(key string, value any) Orm {
	if value == nil {
		fmt.Errorf("value is nil")
		return o
	}

	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return o
	}
	if v.Kind() == reflect.String && v.Len() == 0 {
		return o
	}

	var queryCond builder.Cond
	switch t.Kind() {
	case reflect.Ptr:
		queryCond = builder.NewCond().And(builder.Neq{key: v.Elem().Interface()})
	case reflect.String, reflect.Int, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		queryCond = builder.NewCond().And(builder.Neq{key: v.Interface()})
	default:
		fmt.Errorf("unhandled type: %v", t)
		return o
	}

	if ok := queryCond.IsValid(); ok {
		o.session = o.session.And(queryCond)
	} else {
		fmt.Errorf("fail query valid key: %s value %s", key, formatValue(v))
		return o
	}

	return o
}

func (o *orm) IsNull(key string) Orm {
	o.session = o.session.And(builder.IsNull{key})
	return o
}

func (o *orm) NotNull(key string) Orm {
	o.session = o.session.And(builder.NotNull{key})
	return o
}

func (o *orm) GTE(key string, value any) Orm {
	if value == nil {
		fmt.Errorf("value is nil")
		return o
	}

	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return o
	}
	if v.Kind() == reflect.String && v.Len() == 0 {
		return o
	}

	var queryCond builder.Cond
	switch t.Kind() {
	case reflect.Ptr:
		queryCond = builder.NewCond().And(builder.Gte{key: v.Elem().Interface()})
	case reflect.String, reflect.Int, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		queryCond = builder.NewCond().And(builder.Gte{key: v.Interface()})
	default:
		fmt.Errorf("unhandled type: %v", t)
		return o
	}

	if ok := queryCond.IsValid(); ok {
		o.session = o.session.And(queryCond)
	} else {
		fmt.Errorf("fail query valid key: %s value %s", key, formatValue(v))
		return o
	}

	return o
}

func (o *orm) GT(key string, value any) Orm {
	if value == nil {
		fmt.Errorf("value is nil")
		return o
	}

	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return o
	}
	if v.Kind() == reflect.String && v.Len() == 0 {
		return o
	}

	var queryCond builder.Cond
	switch t.Kind() {
	case reflect.Ptr:
		queryCond = builder.NewCond().And(builder.Gt{key: v.Elem().Interface()})
	case reflect.String, reflect.Int, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		queryCond = builder.NewCond().And(builder.Gt{key: v.Interface()})
	default:
		fmt.Errorf("unhandled type: %v", t)
		return o
	}

	if ok := queryCond.IsValid(); ok {
		o.session = o.session.And(queryCond)
	} else {
		fmt.Errorf("fail query valid key: %s value %s", key, formatValue(v))
		return o
	}

	return o
}

func (o *orm) LTE(key string, value any) Orm {
	if value == nil {
		fmt.Errorf("value is nil")
		return o
	}

	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return o
	}
	if v.Kind() == reflect.String && v.Len() == 0 {
		return o
	}

	var queryCond builder.Cond
	switch t.Kind() {
	case reflect.Ptr:
		queryCond = builder.NewCond().And(builder.Lte{key: v.Elem().Interface()})
	case reflect.String, reflect.Int, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		queryCond = builder.NewCond().And(builder.Lte{key: v.Interface()})
	default:
		fmt.Errorf("unhandled type: %v", t)
		return o
	}

	if ok := queryCond.IsValid(); ok {
		o.session = o.session.And(queryCond)
	} else {
		fmt.Errorf("fail query valid key: %s value %s", key, formatValue(v))
		return o
	}

	return o
}

func (o *orm) LT(key string, value any) Orm {
	if value == nil {
		fmt.Errorf("value is nil")
		return o
	}

	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return o
	}
	if v.Kind() == reflect.String && v.Len() == 0 {
		return o
	}

	var queryCond builder.Cond
	switch t.Kind() {
	case reflect.Ptr:
		queryCond = builder.NewCond().And(builder.Lt{key: v.Elem().Interface()})
	case reflect.String, reflect.Int, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		queryCond = builder.NewCond().And(builder.Lt{key: v.Interface()})
	default:
		fmt.Errorf("unhandled type: %v", t)
		return o
	}

	if ok := queryCond.IsValid(); ok {
		o.session = o.session.And(queryCond)
	} else {
		fmt.Errorf("fail query valid key: %s value %s", key, formatValue(v))
		return o
	}

	return o
}

func (o *orm) Limit(limit int, offset int) Orm {
	if limit <= 0 {
		return o
	}

	o.session = o.session.Limit(limit, offset)
	return o
}

func (o *orm) Find(rowSlicePtr any) error {
	return o.session.Find(rowSlicePtr)
}

func (o *orm) GroupBy(key string) Orm {
	if key == "" {
		return o
	}
	o.session = o.session.GroupBy(key)
	return o
}

func (o *orm) Between(key string, start any, end any) Orm {
	var startIsValid, endIsValid bool

	if start == nil {
		startIsValid = false
	} else {
		startType := reflect.TypeOf(start)
		startValue := reflect.ValueOf(start)

		if startType.Kind() == reflect.Ptr && startValue.IsNil() {
			startIsValid = false
		} else if startType.Kind() == reflect.Ptr && startType.Elem().Kind() == reflect.String && startValue.Elem().Len() == 0 {
			startIsValid = false
		} else if startType.Kind() == reflect.String && startValue.Len() == 0 {
			startIsValid = false
		} else if startType == reflect.TypeOf(time.Time{}) && startValue.Interface().(time.Time).IsZero() {
			endIsValid = false
		} else {
			startIsValid = true
		}
	}

	if end == nil {
		endIsValid = false
	} else {
		endType := reflect.TypeOf(end)
		endValue := reflect.ValueOf(end)

		if endType.Kind() == reflect.Ptr && endValue.IsNil() {
			endIsValid = false
		} else if endType.Kind() == reflect.Ptr && endType.Elem().Kind() == reflect.String && endValue.Elem().Len() == 0 {
			endIsValid = false
		} else if endType.Kind() == reflect.String && endValue.Len() == 0 {
			endIsValid = false
		} else if endType == reflect.TypeOf(time.Time{}) && endValue.Interface().(time.Time).IsZero() {
			endIsValid = false
		} else {
			endIsValid = true
		}
	}

	if !startIsValid && !endIsValid {
		return o
	}

	var queryStr string
	if startIsValid {
		queryStr = key + " >= ?"
		o.session = o.session.Where(queryStr, start)
	}
	if endIsValid {
		queryStr = key + " <= ?"
		o.session = o.session.Where(queryStr, end)
	}

	return o
}

func (o *orm) Before(key string, date time.Time) Orm {
	var queryStr string
	if !date.IsZero() {
		queryStr = key + " < ?"
		o.session = o.session.Where(queryStr, date.Format(time.DateTime))
	}

	return o
}

func (o *orm) After(key string, date time.Time) Orm {
	var queryStr string
	if !date.IsZero() {
		queryStr = key + " > ?"
		o.session = o.session.Where(queryStr, date.Format(time.DateTime))
	}

	return o
}

func (o *orm) BeforeOrEqual(key string, date time.Time) Orm {
	var queryStr string
	if !date.IsZero() {
		queryStr = key + " >= ?"
		o.session = o.session.Where(queryStr, date.Format(time.DateTime))
	}

	return o
}

func (o *orm) AfterOrEqual(key string, date time.Time) Orm {
	var queryStr string
	if !date.IsZero() {
		queryStr = key + " <= ?"
		o.session = o.session.Where(queryStr, date.Format(time.DateTime))
	}

	return o
}

func (o *orm) In(key string, values ...interface{}) Orm {
	if values == nil || len(values) == 0 {
		return o
	}

	o.session = o.session.In(key, values...)
	return o
}

func (o *orm) NotIn(key string, values ...interface{}) Orm {
	if values == nil || len(values) == 0 {
		return o
	}

	o.session = o.session.NotIn(key, values...)
	return o

}

func (o *orm) Unscoped() Orm {
	o.session = o.session.Unscoped()
	return o
}

func (o *orm) FindAndCount(rowSlicePtr any) (int64, error) {
	return o.session.FindAndCount(rowSlicePtr)
}

func (o *orm) Id(id interface{}) Orm {
	o.session = o.session.ID(id)
	return o
}

func (o *orm) Get(beans interface{}) (bool, error) {
	return o.session.Get(beans)
}

func (o *orm) Count() (int64, error) {
	return o.session.Count(o.source)
}

func (o *orm) Sum(columnName string) (float64, error) {
	return o.session.Sum(o.source, columnName)
}

func (o *orm) Sums(columnNames []string) ([]float64, error) {
	return o.session.Sums(o.source, columnNames...)
}

func (o *orm) SumInt(columnName string) (int64, error) {
	return o.session.SumInt(o.source, columnName)
}

func (o *orm) SumsInt(columnNames []string) ([]int64, error) {
	return o.session.SumsInt(o.source, columnNames...)
}

type ormObject interface {
	TableName() string
	SelectAll() string
}

func toOrmObject(v interface{}) (ormObject, bool) {
	if i, ok := v.(ormObject); ok {
		return i, true
	}
	return nil, false
}

func buildJoinConditions(conditions []interface{}) string {
	var result string

	for _, condition := range conditions {
		conditionStr, ok := condition.(string)
		if !ok {
			return ""
		}

		if len(result) > 0 {
			result = fmt.Sprintf("%s AND %s", result, conditionStr)
		} else {
			result = conditionStr
		}
	}

	return result
}

func getPrimaryKey(object any) (string, error) {
	t := reflect.TypeOf(object)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		tag := field.Tag.Get("xorm")
		if strings.Contains(tag, "pk") {
			return strings.ToLower(field.Name), nil
		}
	}
	return "", errors.New("pk not found")
}

func formatValue(v reflect.Value) string {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return "nil"
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return fmt.Sprintf("%v", v.Interface())
}
