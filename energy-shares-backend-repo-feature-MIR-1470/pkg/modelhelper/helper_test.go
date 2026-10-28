package modelhelper

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// 並び替えの指定は画面から修飾なしで届く。JOIN 先にも同じ名前の列があると
// ORDER BY があいまいになり実行できないため、既定のテーブル名を補う。
func TestQualifySortColumns(t *testing.T) {
	tests := []struct {
		name    string
		table   string
		sortStr string
		want    string
	}{
		{
			name:    "(成功) 修飾のない1項目にテーブル名を補う",
			table:   "share",
			sortStr: "created",
			want:    "share.created",
		},
		{
			name:    "(成功) 複数項目すべてに補う",
			table:   "withdrawal",
			sortStr: "created,id",
			want:    "withdrawal.created,withdrawal.id",
		},
		{
			// 画面が JOIN 先の列で並べたい場合に備え、指定済みは尊重する。
			name:    "(成功) 既に修飾されているものは変えない",
			table:   "share",
			sortStr: "project.name",
			want:    "project.name",
		},
		{
			// 空白を残すと列名の一部として扱われ、クォートされて壊れる。
			name:    "(成功) 修飾済みでも前後の空白を落とす",
			table:   "dividend",
			sortStr: "transaction_date, dividend.id",
			want:    "dividend.transaction_date,dividend.id",
		},
		{
			name:    "(成功) 修飾ありと無しが混ざっていても項目ごとに判断する",
			table:   "share",
			sortStr: "project.name,created",
			want:    "project.name,share.created",
		},
		{
			name:    "(成功) 前後の空白を落として補う",
			table:   "dividend",
			sortStr: "transaction_date, id",
			want:    "dividend.transaction_date,dividend.id",
		},
		{
			name:    "(成功) テーブル名が空なら何もしない",
			table:   "",
			sortStr: "created,id",
			want:    "created,id",
		},
		{
			name:    "(成功) 並び替えの指定が空なら何もしない",
			table:   "share",
			sortStr: "",
			want:    "",
		},
		{
			// SetSortOrder が並び順の指定と項目数を突き合わせるため、
			// 空の項目があっても数を変えない。
			name:    "(成功) 空の項目があっても項目数を保つ",
			table:   "share",
			sortStr: "created,,id",
			want:    "share.created,,share.id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QualifySortColumns(tt.table, tt.sortStr)

			assert.Equal(t, tt.want, got)
		})
	}
}

// newSortOrderTestEngine は SetSortOrder が実際に ORDER BY を組み立てられる
// ことを確認するための sqlite エンジンを用意する。amount の昇順と id の昇順が
// 一致しないデータを入れておくことで、列を取り違えると検出できるようにする。
func newSortOrderTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	_, err = engine.Exec(`
CREATE TABLE item (
    id INTEGER PRIMARY KEY,
    amount INTEGER
);
INSERT INTO item (id, amount) VALUES (1, 30), (2, 10), (3, 20);
`)
	require.NoError(t, err)

	return engine
}

type item struct {
	Id     int64
	Amount int64
}

// MIR-1113: sortby は画面から届く外部入力で、SetSortOrder は検証なしに
// ORDER BY へそのまま渡していた。xorm(MySQL)の Quoter は列名の前後だけを
// バッククォートで囲み、文字列内部のバッククォートはエスケープしないため、
// 「id`+SLEEP(5)+`id」のような値を渡すと ORDER BY `id`+SLEEP(5)+`id` として
// 実行され、列 id と SLEEP(5) の加算式になって時間差によるブラインド抽出が
// 成立する(xorm.io/xorm@v1.3.9 schemas/quote.go の quoteWordTo で確認済み)。
// 識別子として妥当な形以外は積む前に弾くことを確認する。
func TestSetSortOrder(t *testing.T) {
	t.Run("(成功) 単一列・descで並ぶ", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SetSortOrder(q, "amount", "desc"))

		var items []item
		require.NoError(t, q.Find(&items))
		require.Len(t, items, 3)
		assert.Equal(t, []int64{30, 20, 10}, []int64{items[0].Amount, items[1].Amount, items[2].Amount})
	})

	t.Run("(成功) テーブル修飾つきの列でも並ぶ", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SetSortOrder(q, "item.amount", "asc"))

		var items []item
		require.NoError(t, q.Find(&items))
		require.Len(t, items, 3)
		assert.Equal(t, []int64{10, 20, 30}, []int64{items[0].Amount, items[1].Amount, items[2].Amount})
	})

	t.Run("(成功) 複数列・複数orderの組み合わせ", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SetSortOrder(q, "amount,id", "desc,asc"))

		var items []item
		require.NoError(t, q.Find(&items))
		require.Len(t, items, 3)
		assert.Equal(t, []int64{30, 20, 10}, []int64{items[0].Amount, items[1].Amount, items[2].Amount})
	})

	invalidCases := []struct {
		name    string
		sortStr string
	}{
		{"バッククォートとSLEEPによる注入", "id`+SLEEP(5)+`id"},
		{"コメントマーカー", "id -- "},
		{"セミコロンでの文分割", "id;DROP TABLE item"},
		{"UNION SELECT", "id) UNION SELECT 1--"},
		{"空白を含む式", "id amount"},
		{"括弧を含む式", "id)"},
		{"演算子を含む式", "id+1"},
		{"シングルクォート", "id' OR '1'='1"},
		{"ドットが2つ(スキーマ修飾は想定外)", "db.item.id"},
		{"空要素と不正な列が混在していても不正側で全体を却下する", ",id`+SLEEP(5)+`id"},
	}

	for _, tc := range invalidCases {
		t.Run("(失敗) "+tc.name, func(t *testing.T) {
			engine := newSortOrderTestEngine(t)
			q := engine.NewSession()
			defer q.Close()

			err := SetSortOrder(q, tc.sortStr, "asc")
			require.Error(t, err)

			// 拒否した場合、ORDER BY は一切積まれていないことを確認する
			// (一部の列だけ有効でも部分的に適用しない)。
			var items []item
			require.NoError(t, q.Find(&items))
			require.Len(t, items, 3)
		})
	}

	t.Run("(成功→失敗) 複数列のうち1つでも不正なら全体を却下する", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		err := SetSortOrder(q, "amount,id`+SLEEP(5)+`id", "asc,asc")
		require.Error(t, err)
	})

	// MIR-1113: "主キー,"+画面指定 のように固定列へ条件付きで連結する呼び出し元
	// (FAQ一覧など)は、画面側が指定しなかったとき末尾に空要素を残す。空要素を
	// 不正として全体を拒否すると、既定の並び替えまで失われる回帰が起きるため、
	// 列を指定しないプレースホルダとして読み飛ばす。
	t.Run("(成功) 末尾の空要素は列指定なしとして読み飛ばす", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SetSortOrder(q, "amount,", "desc,"))

		var items []item
		require.NoError(t, q.Find(&items))
		require.Len(t, items, 3)
		assert.Equal(t, []int64{30, 20, 10}, []int64{items[0].Amount, items[1].Amount, items[2].Amount})
	})

	// order だけ指定されて sortby 側が空(画面がsortbyを送らずorderだけ送った
	// 場合の再現)でも、既定列の並び替えが消えないことを確認する。
	t.Run("(成功) sortby側が空でもorderの項目数がずれても既定列は効く", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SetSortOrder(q, "amount,", "desc,asc"))

		var items []item
		require.NoError(t, q.Find(&items))
		require.Len(t, items, 3)
		assert.Equal(t, []int64{30, 20, 10}, []int64{items[0].Amount, items[1].Amount, items[2].Amount})
	})

	t.Run("(成功) 前後の空白を落として検証する", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SetSortOrder(q, " amount , id ", "desc,asc"))

		var items []item
		require.NoError(t, q.Find(&items))
		require.Len(t, items, 3)
		assert.Equal(t, []int64{30, 20, 10}, []int64{items[0].Amount, items[1].Amount, items[2].Amount})
	})

	// order側の前後の空白を落とさないと、"desc"に対して" desc"が一致せず
	// 並び替えが無言で効かなくなる(列名側だけ空白を落とすのは非対称)。
	t.Run("(成功) orderの前後の空白も落として判定する", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SetSortOrder(q, "amount", " desc "))

		var items []item
		require.NoError(t, q.Find(&items))
		require.Len(t, items, 3)
		assert.Equal(t, []int64{30, 20, 10}, []int64{items[0].Amount, items[1].Amount, items[2].Amount})
	})

	t.Run("(既存の挙動) sortStrかorderStrが空なら何もしない", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SetSortOrder(q, "", "asc"))
		require.NoError(t, SetSortOrder(q, "amount", ""))
	})

	t.Run("(既存の挙動) sortByとorderの項目数が合わずorderが複数ならエラー", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		err := SetSortOrder(q, "amount", "asc,desc")
		assert.EqualError(t, err, "'sortBy', 'order' sizes mismatch or 'order' size is not 1")
	})

	// MIR-1113: 以前は列に対応するorderが不正だとその場でエラーを返し、以降の
	// 組を一切処理しなかった。orderは常にasc/descの二択判定にしか使わずSQLへ
	// 混入しないため、その組だけ読み飛ばしても安全側に倒れる。列名側の空要素と
	// 同じ「読み飛ばす」方式に揃え、後続の正当な列が巻き添えで失われないように
	// する。
	t.Run("(成功) 列に対応するorderが不正なら、その列だけ読み飛ばす", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SetSortOrder(q, "amount", "invalid"))

		var items []item
		require.NoError(t, q.Find(&items))
		require.Len(t, items, 3)
	})

	t.Run("(成功) 複数列のうち1つだけorderが不正でも後続の列は適用される", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		// idの昇順は挿入順(1,2,3)と偶然一致してしまい、"適用されなかった"場合と
		// 区別できないため、区別できるdescを使う。
		require.NoError(t, SetSortOrder(q, "amount,id", "invalid,desc"))

		var items []item
		require.NoError(t, q.Find(&items))
		require.Len(t, items, 3)
		assert.Equal(t, []int64{3, 2, 1}, []int64{items[0].Id, items[1].Id, items[2].Id})
	})

	t.Run("(既存の挙動) 複数列に対して単一のorderが不正ならエラー", func(t *testing.T) {
		engine := newSortOrderTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		err := SetSortOrder(q, "amount,id", "invalid")
		require.Error(t, err)
	})
}

// newSearchQueryTestEngine は SearchQuery の検証用に、name 列を持つ search_item
// テーブルを用意する。newSortOrderTestEngine と分けているのは、Like検索(部分一致)を
// 確認する文字列カラムが必要なため。
func newSearchQueryTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	_, err = engine.Exec(`
CREATE TABLE search_item (
    id INTEGER PRIMARY KEY,
    name TEXT
);
INSERT INTO search_item (id, name) VALUES (1, 'apple'), (2, 'banana'), (3, 'grape');
`)
	require.NoError(t, err)

	return engine
}

type searchItem struct {
	Id   int64
	Name string
}

// MIR-1232: searchKeys は画面から届く外部入力で、SearchQuery は検証なしにそのまま
// builder.Like の列名として連結していた。xorm/builderのLike.WriteToは列名を
// 無クォート・無エスケープで出力するため(ORDER BYと違いバッククォートで囲まれない)、
// 細工不要でWHERE句へ任意の式を注入できた。allowedKeysに無い列名を1件でも含めば
// 検索全体を拒否することを確認する。
func TestSearchQuery(t *testing.T) {
	t.Run("(成功) keyStrが空なら何もせず全件返る", func(t *testing.T) {
		engine := newSearchQueryTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SearchQuery(q, "", "", "", map[string]struct{}{}))

		var items []searchItem
		require.NoError(t, q.Find(&items))
		assert.Len(t, items, 3)
	})

	t.Run("(成功) allowedKeysに含まれる列名なら絞り込まれる(前後の空白はトリムして検証・クエリの両方に使う)", func(t *testing.T) {
		engine := newSearchQueryTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SearchQuery(q, "", " name ", "apple", map[string]struct{}{"name": {}}))

		var items []searchItem
		require.NoError(t, q.Find(&items))
		require.Len(t, items, 1)
		assert.Equal(t, "apple", items[0].Name)
	})

	t.Run("(拒否) allowedKeysが空だとどんな列名も拒否される", func(t *testing.T) {
		engine := newSearchQueryTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		err := SearchQuery(q, "", "name", "apple", map[string]struct{}{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("(拒否) allowedKeysに無い列名を1件でも含めば検索全体を拒否する", func(t *testing.T) {
		engine := newSearchQueryTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		err := SearchQuery(q, "", "name,id", "apple,1", map[string]struct{}{"name": {}})
		require.Error(t, err, "idはallowedKeysに無いため、nameが許可されていても全体を拒否する")
	})

	t.Run("(拒否) SQLインジェクションを試みる列名も識別子と同じ扱いで拒否される", func(t *testing.T) {
		engine := newSearchQueryTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		err := SearchQuery(q, "", "id`=1 OR SLEEP(5)-- ", "1", map[string]struct{}{})
		require.Error(t, err)
	})

	t.Run("(拒否) keyStrとvalueStrの要素数が合わなければpanicせずエラーを返す", func(t *testing.T) {
		engine := newSearchQueryTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NotPanics(t, func() {
			err := SearchQuery(q, "", "name,id", "apple", map[string]struct{}{"name": {}, "id": {}})
			require.Error(t, err)
		})
	})
}

func TestToZenkakuKatakana(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "全角カタカナはそのまま", in: "ヤマダ", want: "ヤマダ"},
		{name: "ひらがなはカタカナへ", in: "やまだ", want: "ヤマダ"},
		{name: "半角カタカナは全角へ(濁点を合成)", in: "ﾔﾏﾀﾞ", want: "ヤマダ"},
		{name: "半濁点・長音", in: "ﾎﾟｰﾙ", want: "ポール"},
		{name: "小書き・踊り字", in: "ちょっとゞ", want: "チョットヾ"},
		{name: "ひらがなと半角の混在", in: "やまﾀﾞ", want: "ヤマダ"},
		{name: "空文字", in: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ToZenkakuKatakana(tt.in))
		})
	}
}

func newKanaSearchTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	_, err = engine.Exec(`
CREATE TABLE kana_item (
    id INTEGER PRIMARY KEY,
    family_kana TEXT,
    first_kana TEXT,
    status TEXT
);
INSERT INTO kana_item (id, family_kana, first_kana, status) VALUES
    (1, 'ヤマダ', 'タロウ', 'active'),
    (2, 'サトウ', 'ヤマト', 'active'),
    (3, 'タナカ', 'ハナコ', 'inactive'),
    (4, 'タ_ナ', 'テスト', 'active');
`)
	require.NoError(t, err)

	return engine
}

type kanaItem struct {
	Id         int64
	FamilyKana string
	FirstKana  string
	Status     string
}

// MIR-534: フリガナ検索。複数列の OR を1つのグループで足し、入力は全角カタカナにそろえた値でも
// 照合し、% と _ はワイルドカードとして扱わない。
func TestSearchKanaLikeQuery(t *testing.T) {
	columns := []KanaNameColumns{{Family: "family_kana", First: "first_kana"}}

	findIds := func(t *testing.T, q *xorm.Session) []int64 {
		t.Helper()
		var items []kanaItem
		require.NoError(t, q.Table("kana_item").Find(&items))
		ids := make([]int64, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.Id)
		}
		return ids
	}
	strPtr := func(s string) *string { return &s }

	t.Run("(成功) 値が nil・空文字・空白だけ、または列が無ければ何もせず全件返る", func(t *testing.T) {
		engine := newKanaSearchTestEngine(t)
		for _, tc := range []struct {
			name    string
			columns []KanaNameColumns
			value   *string
		}{
			{"nil", columns, nil},
			{"空文字", columns, strPtr("")},
			{"空白だけ(全角含む)", columns, strPtr(" 　")},
			{"列なし", nil, strPtr("ヤマ")},
		} {
			q := engine.NewSession()
			require.NoError(t, SearchKanaLikeQuery(q, "", tc.columns, tc.value), tc.name)
			assert.ElementsMatch(t, []int64{1, 2, 3, 4}, findIds(t, q), tc.name)
			q.Close()
		}
	})

	t.Run("(成功) どれか1列に部分一致すればヒットする", func(t *testing.T) {
		engine := newKanaSearchTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SearchKanaLikeQuery(q, "", columns, strPtr("ヤマ")))
		assert.ElementsMatch(t, []int64{1, 2}, findIds(t, q))
	})

	t.Run("(成功) 姓名をつなげた入力(空白の有無を問わない)でも当たる", func(t *testing.T) {
		engine := newKanaSearchTestEngine(t)
		for _, tc := range []struct {
			input string
			want  []int64
		}{
			{"ヤマダタロウ", []int64{1}},
			{"ヤマダ タロウ", []int64{1}},
			{"ヤマダ　タロウ", []int64{1}},
			{"ダタロ", []int64{1}},     // 姓の末尾から名の先頭にまたがる部分一致
			{"やまだ たろう", []int64{1}}, // ひらがな+空白
			{"ﾔﾏﾀﾞﾀﾛｳ", []int64{1}},
			{"タロウヤマダ", []int64{}}, // 名姓の逆順は当たらない
		} {
			q := engine.NewSession()
			require.NoError(t, SearchKanaLikeQuery(q, "", columns, strPtr(tc.input)), tc.input)
			assert.ElementsMatch(t, tc.want, findIds(t, q), tc.input)
			q.Close()
		}
	})

	t.Run("(成功) 名が NULL でも姓だけで当たり、姓名の照合でエラーにならない", func(t *testing.T) {
		engine := newKanaSearchTestEngine(t)
		_, err := engine.Exec(`INSERT INTO kana_item (id, family_kana, first_kana, status) VALUES (5, 'キムラ', NULL, 'active')`)
		require.NoError(t, err)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SearchKanaLikeQuery(q, "", columns, strPtr("キムラ")))
		assert.ElementsMatch(t, []int64{5}, findIds(t, q))
	})

	t.Run("(成功) 既存の条件とは AND で結び、列の OR は1つのグループにまとまる", func(t *testing.T) {
		engine := newKanaSearchTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		// グループ化されていないと「status = 'inactive' AND family LIKE … OR first LIKE …」となり、
		// active の 2(名=ヤマト)が混ざる。
		q.Where("status = ?", "inactive")
		require.NoError(t, SearchKanaLikeQuery(q, "", columns, strPtr("ヤマ")))
		assert.Empty(t, findIds(t, q))
	})

	t.Run("(成功) テーブル名を指定すると修飾した列名で検索する", func(t *testing.T) {
		engine := newKanaSearchTestEngine(t)
		q := engine.NewSession()
		defer q.Close()

		require.NoError(t, SearchKanaLikeQuery(q, "kana_item", columns, strPtr("ハナ")))
		assert.ElementsMatch(t, []int64{3}, findIds(t, q))
	})

	t.Run("(成功) ひらがな・半角カタカナの入力でも全角カタカナに当たる", func(t *testing.T) {
		engine := newKanaSearchTestEngine(t)
		for _, tc := range []struct {
			input string
			want  []int64
		}{
			{"やまだ", []int64{1}},
			{"ﾔﾏ", []int64{1, 2}},
			{"ﾀﾛｳ", []int64{1}},
		} {
			q := engine.NewSession()
			require.NoError(t, SearchKanaLikeQuery(q, "", columns, strPtr(tc.input)), tc.input)
			assert.ElementsMatch(t, tc.want, findIds(t, q), tc.input)
			q.Close()
		}
	})

	t.Run("(成功) % と _ はワイルドカードにならず、文字そのものとして照合する", func(t *testing.T) {
		engine := newKanaSearchTestEngine(t)
		for _, tc := range []struct {
			input string
			want  []int64
		}{
			// エスケープしないと「タ」+任意の1文字として 1(タロウ)・3(タナカ)にも当たる
			{"タ_", []int64{4}},
			{"%", []int64{}},
			{"!", []int64{}},
		} {
			q := engine.NewSession()
			require.NoError(t, SearchKanaLikeQuery(q, "", columns, strPtr(tc.input)), tc.input)
			assert.ElementsMatch(t, tc.want, findIds(t, q), tc.input)
			q.Close()
		}
	})
}
