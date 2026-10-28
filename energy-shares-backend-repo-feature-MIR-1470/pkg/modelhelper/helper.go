package modelhelper

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// sortColumnPattern は ORDER BY に使ってよい列名の形を定義する。裸の識別子
// (id, created_at)か、それを「.」で1回だけ区切ったテーブル修飾形(share.id)
// しか通さない。MIR-1113: sortby は画面から届く外部入力で、これまで検証なしに
// ORDER BY へそのまま渡していた。xorm(MySQL)の Quoter は列名をバッククォートで
// 囲むが文字列内部のバッククォートはエスケープしないため、
// 「id`+SLEEP(5)+`id」のような値を渡すと ORDER BY `id`+SLEEP(5)+`id` として
// 実行され、時間差によるブラインド抽出が成立する。列名としてあり得ない文字
// (バッククォート・演算子・空白・関数呼び出しなど)を機械的に弾く。
var sortColumnPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)?$`)

// ValidateSortColumns はカンマ区切りの列名がすべて識別子として妥当かを検証し、
// 前後の空白を落とした列名のスライスを返す。pkg/orm.SortOrder にも全く同じ
// 脆弱性(検証なしに ORDER BY へ渡す実装)があり、同じ検証を共有するため
// export する。空(前後の空白だけ含む場合を含む)の項目は、列を指定しない
// プレースホルダとして許可し、そのまま残す(呼び出し側で読み飛ばす)。
// "pin," のように固定の主キーへ画面指定の列を条件付きで連結する呼び出し元が
// 複数あり、画面側が指定しなかったときに末尾へ空項目が残るため。空でない
// 項目は識別子の形以外を1件でも含めばエラーにする(部分適用を避ける。MIR-472で
// 個別列ごとの許可リストが3周しても収束しなかった経緯があるため、識別子の形
// だけを見る一元検証に寄せている)。
func ValidateSortColumns(sortStr string) ([]string, error) {
	columns := strings.Split(sortStr, ",")
	for i, v := range columns {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" && !sortColumnPattern.MatchString(trimmed) {
			return nil, fmt.Errorf("invalid sort column: %q", v)
		}
		columns[i] = trimmed
	}
	return columns, nil
}

const (
	MaxFile     = 15
	MaxPinCount = 3
)

// 範囲値検索
func SetBetweenRange(q xorm.Interface, column string, start, end *int) error {
	var queryStr string
	if start != nil {
		queryStr = column + " >= ?"
		q.Where(queryStr, start)
	}
	if end != nil {
		queryStr = column + " <= ?"
		q.Where(queryStr, end)
	}
	return nil
}

// 期間検索機能
func SetBetweenDate(q xorm.Interface, column, startDate, endDate string) error {
	var queryStr string
	if len(startDate) > 0 {
		//queryStr = "DATE_FORMAT(" + column + ",'%Y-%m-%d %H:%i:%s') >= ?"
		queryStr = column + " >= ?"
		q.Where(queryStr, startDate)
	}
	if len(endDate) > 0 {
		queryStr = column + " <= ?"
		q.Where(queryStr, endDate)
	}

	return nil
}

// SetSortOrderForTable は、テーブル名の付いていない並び替え項目に既定の
// テーブル名を補ってから SetSortOrder を呼ぶ。
//
// JOIN した先にも同じ名前の列があると、修飾なしの ORDER BY は
// 「Column 'created' in order clause is ambiguous」で実行できない。
// 並び替えの指定は画面から来るため、DB のテーブル構成を画面に持ち込まずに
// 済むよう、ここで補う。
//
// 既に「テーブル名.列名」の形で渡されたものはそのまま使う。
func SetSortOrderForTable(q xorm.Interface, table, sortStr, orderStr string) error {
	return SetSortOrder(q, QualifySortColumns(table, sortStr), orderStr)
}

// QualifySortColumns は修飾のない列名へテーブル名を付ける。
// 項目数は変えない。SetSortOrder が並び順の指定と数を突き合わせるため。
func QualifySortColumns(table, sortStr string) string {
	if len(table) == 0 || len(sortStr) == 0 {
		return sortStr
	}

	columns := strings.Split(sortStr, ",")
	for i, column := range columns {
		trimmed := strings.TrimSpace(column)
		if len(trimmed) == 0 {
			continue
		}
		// 修飾済みでも空白は落とす。残すと列名の一部として扱われ、
		// 「 dividend.id」のような指定がそのままクォートされて壊れる。
		if strings.Contains(trimmed, ".") {
			columns[i] = trimmed
			continue
		}
		columns[i] = table + "." + trimmed
	}

	return strings.Join(columns, ",")
}

func SetSortOrder(q xorm.Interface, sortStr, orderStr string) error {
	if len(sortStr) <= 0 || len(orderStr) <= 0 {
		return nil
	}
	sortBy, err := ValidateSortColumns(sortStr)
	if err != nil {
		return err
	}
	order := strings.Split(orderStr, ",")

	if len(sortBy) != 0 {
		if len(sortBy) == len(order) {
			// 1) for each sort field, there is an associated order
			//
			// MIR-1113: 以前はorderが不正/空だとその場でエラーを返し、以降の
			// 組を一切処理しなかった("pin,name" + "desc," のように、既定列の
			// あとに画面指定の列だけorderが欠けるケースで、後続の正当な列が
			// 巻き添えで失われていた)。orderは常にasc/descの二択判定にしか
			// 使わずSQLへ混入しないため、その組だけ読み飛ばしても安全側に
			// 倒れる。列名側の空要素と同じ「読み飛ばす」方式に揃える。
			for i, v := range sortBy {
				if v == "" {
					continue
				}
				switch strings.ToLower(strings.TrimSpace(order[i])) {
				case "desc":
					q.Desc(v)
				case "asc":
					q.Asc(v)
				}
			}
		} else if len(sortBy) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			dir := strings.ToLower(strings.TrimSpace(order[0]))
			if dir != "asc" && dir != "desc" {
				return errors.New("Invalid order. Must be either [asc|desc]")
			}
			for _, v := range sortBy {
				if v == "" {
					continue
				}
				if dir == "desc" {
					q.Desc(v)
				} else {
					q.Asc(v)
				}
			}
		} else if len(sortBy) != len(order) && len(order) != 1 {
			return errors.New("'sortBy', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return errors.New("unused 'order' fields")
		}
	}
	return nil
}

// SearchQueryはallowedKeysに含まれる列名だけを検索条件として許可する。keyStrは
// controllers層のクエリパラメータ(searchKeys)から届く外部入力で、以前は検証なしに
// そのままbuilder.Likeの列名として連結していた。xorm/builderのLike.WriteToは列名を
// 無クォート・無エスケープで出力するため、ORDER BY(MIR-1113)と違いバッククォートを
// 閉じる細工すら不要で、WHERE句へ直接SQLインジェクションが成立していた(MIR-1232)。
// 呼び出し元が許可する列名を明示的に渡す形にし、allowedKeysに無い列名が1件でも
// あれば検索全体をエラーにする(部分適用を避ける。MIR-1113のValidateSortColumnsと
// 同じ判断)。呼び出し元(AntiSocialWaitUserList)は現時点でこのパラメータを画面から
// 一切使っていないため、allowedKeysは空で渡している(MIR-1232)。
func SearchQuery(q xorm.Interface, tabelName, keyStr, valueStr string, allowedKeys map[string]struct{}) error {
	if len(keyStr) <= 0 || len(valueStr) <= 0 {
		return nil
	}
	keys := strings.Split(keyStr, ",")
	values := strings.Split(valueStr, ",")
	// 敵対的検証の指摘(MIR-1232): 許可リストが空の現状では実害が無いが、将来列名を
	// 追加したときに備え、要素数の不一致(values[idx]の範囲外アクセス)と、検証した
	// 値(トリム済み)と実際にクエリへ渡す値がずれる問題を先に塞いでおく。
	if len(keys) != len(values) {
		return fmt.Errorf("search keys and values count mismatch: %d keys, %d values", len(keys), len(values))
	}
	for i, key := range keys {
		trimmed := strings.TrimSpace(key)
		if _, ok := allowedKeys[trimmed]; !ok {
			return fmt.Errorf("invalid search key: %q", key)
		}
		keys[i] = trimmed
	}
	queryCond := builder.NewCond()

	for idx, key := range keys {
		if tabelName != "" {
			queryCond = queryCond.And(builder.Like{tabelName + "." + key, values[idx]}) //Map調整
		} else {
			queryCond = queryCond.And(builder.Like{key, values[idx]}) //Map調整
		}
	}
	if ok := queryCond.IsValid(); ok {
		q.(*xorm.Session).And(queryCond)
	} else {
		return errors.New("fail query valid")
	}

	return nil
}

// keyStr はジョイン後のエイリアスではなく、実際のテーブルを指定する必要があります
func SearchLikeQuery(q xorm.Interface, tableName, keyStr string, valueStr *string) error {
	if keyStr == "" || valueStr == nil || len(*valueStr) == 0 {
		return nil
	}
	queryCond := builder.NewCond()

	if tableName != "" {
		queryCond = queryCond.And(builder.Like{tableName + "." + keyStr, *valueStr})
	} else {
		queryCond = queryCond.And(builder.Like{keyStr, *valueStr})
	}

	if ok := queryCond.IsValid(); ok {
		q.(*xorm.Session).And(queryCond)
	} else {
		return errors.New("fail query valid")
	}

	return nil
}

// KanaNameColumns は姓・名のフリガナ列の組(MIR-534)。
type KanaNameColumns struct {
	Family string
	First  string
}

// SearchKanaLikeQuery はフリガナの部分一致を、複数列の OR として1つのグループで足す(MIR-534)。
//   - 姓・名それぞれの列に加え、姓名をつなげた値(CONCAT)にも照合する。
//     入力の空白は姓名の照合では取り除くため、「ヤマダタロウ」「ヤマダ タロウ」でも当たる。
//   - 照合順序(utf8mb4_general_ci)はひらがなとカタカナ、半角と全角を区別するため、
//     入力そのままの値に加えて全角カタカナにそろえた値でも検索する(バリエーションの OR)。
//   - % と _ はワイルドカードとして効かないようエスケープし、純粋な部分一致にする。
func SearchKanaLikeQuery(q xorm.Interface, tableName string, names []KanaNameColumns, valueStr *string) error {
	if len(names) == 0 || valueStr == nil {
		return nil
	}
	value := strings.TrimSpace(*valueStr)
	if value == "" {
		return nil
	}

	qualify := func(keyStr string) string {
		if tableName == "" {
			return keyStr
		}
		return tableName + "." + keyStr
	}
	like := func(expr, v string) builder.Cond {
		return builder.Expr(expr+" LIKE ? ESCAPE '"+likeEscape+"'", "%"+likeEscaper.Replace(v)+"%")
	}

	queryCond := builder.NewCond()
	for _, v := range uniqueNonEmptyStrings(value, ToZenkakuKatakana(value)) {
		fullName := removeSpaces(v)
		for _, n := range names {
			family, first := qualify(n.Family), qualify(n.First)
			// MySQL の CONCAT は NULL を含むと NULL になる(SQLite は無視する)ため、COALESCE でそろえる。
			concat := "CONCAT(COALESCE(" + family + ", ''), COALESCE(" + first + ", ''))"
			queryCond = queryCond.Or(like(family, v), like(first, v), like(concat, fullName))
		}
	}

	if ok := queryCond.IsValid(); ok {
		q.(*xorm.Session).And(queryCond)
	} else {
		return errors.New("fail query valid")
	}

	return nil
}

// removeSpaces は半角・全角の空白をすべて取り除く。
func removeSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

// likeEscape は LIKE のエスケープ文字。MySQL と SQLite の両方で1文字として通る記号を使う。
const likeEscape = "!"

var likeEscaper = strings.NewReplacer(likeEscape, likeEscape+likeEscape, "%", likeEscape+"%", "_", likeEscape+"_")

// ToZenkakuKatakana はフリガナ検索用に、入力を全角カタカナへそろえる(MIR-534)。
//   - 半角カタカナ → 全角カタカナ。NFKC で濁点・半濁点も1文字に合成する(ﾀﾞ → ダ)。
//   - ひらがな → カタカナ(やまだ → ヤマダ)。
//
// NFKC は全角英数字も半角にするが、フリガナの照合には影響しない。
func ToZenkakuKatakana(s string) string {
	return strings.Map(func(r rune) rune {
		// ぁ(U+3041)〜ゖ(U+3096)、ゝゞ(U+309D, U+309E)は 0x60 先が対応するカタカナ。
		if (r >= 'ぁ' && r <= 'ゖ') || r == 'ゝ' || r == 'ゞ' {
			return r + 0x60
		}
		return r
	}, norm.NFKC.String(s))
}

// uniqueNonEmptyStrings は空文字と重複を除いた値を、元の順序のまま返す。
func uniqueNonEmptyStrings(values ...string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func SearchLikeQueryForOrOperation(q xorm.Interface, tabelName, keyStr string, valueStr *string) error {
	if keyStr == "" || valueStr == nil || len(*valueStr) == 0 {
		return nil
	}

	queryCond := builder.NewCond()

	if tabelName != "" {
		queryCond = queryCond.Or(builder.Like{tabelName + "." + keyStr, *valueStr})
	} else {
		queryCond = queryCond.Or(builder.Like{keyStr, *valueStr})
	}

	if ok := queryCond.IsValid(); ok {
		q.(*xorm.Session).Or(queryCond)
	} else {
		return errors.New("fail query valid")
	}

	return nil
}

// SearchEqQuery キー値はジョイン後のエイリアスではなく、実際のテーブルを指定する必要があります
func SearchEqQuery(q xorm.Interface, tableName, keyStr string, valueStr *string) error {
	if keyStr == "" || valueStr == nil || len(*valueStr) == 0 {
		return nil
	}

	queryCond := builder.NewCond()

	if tableName != "" {
		queryCond = queryCond.And(builder.Eq{tableName + "." + keyStr: *valueStr})
	} else {
		queryCond = queryCond.And(builder.Eq{keyStr: *valueStr})
	}

	if ok := queryCond.IsValid(); ok {
		q.(*xorm.Session).And(queryCond)
	} else {
		return fmt.Errorf("fail query valid table: %s key: %s value %s", tableName, keyStr, *valueStr)
	}

	return nil
}

// keyStr はジョイン後のエイリアスではなく、実際のテーブルを指定する必要があります
func SearchBoolQuery(q xorm.Interface, tabelName, keyStr string, valueStr *bool) error {
	if keyStr == "" || valueStr == nil {
		return nil
	}

	queryCond := builder.NewCond()

	if tabelName != "" {
		queryCond = queryCond.And(builder.Eq{tabelName + "." + keyStr: *valueStr})
	} else {
		queryCond = queryCond.And(builder.Eq{keyStr: *valueStr})
	}

	if ok := queryCond.IsValid(); ok {
		q.(*xorm.Session).And(queryCond)
	} else {
		return errors.New("fail query valid")
	}

	return nil
}

func Find(items []string, val string) bool {
	for _, v := range items {
		if v == val {
			return true
		}
	}

	return false
}

func YearDiff(a, b time.Time) (years int) {

	var months int

	m := a.Month()
	for a.Before(b) {
		a = a.Add(time.Hour * 24)
		m2 := a.Month()
		if m2 != m {
			months++
		}
		m = m2
	}
	years = months / 12
	return
}

// 価格として表現
func ConvertNumberToPrice(number string) string {

	var price string

	index := 0
	for i := len(number) - 1; i >= 0; i-- {

		if index != 0 && (index%3) == 0 {
			price = "," + price
		}

		price = string(number[i]) + price
		index++
	}

	return price
}

// GetPostalCode returns a random Postal Code.
func GetPostalCode() (string, error) {

	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Int31n(90000)+10000), nil
}
