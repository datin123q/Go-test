package gofpdf

import (
	"os"
	"strings"
	"testing"
)

// フォントの実体は energy-shares-backend-repo のリポジトリ側にある。
// thirdparty/gofpdf は独立したGoモジュールだが、ディレクトリの相対位置は
// 変わらないため、このパスでリポジトリ側のテストデータを参照する。
const mir1201TestFontPath = "../../assets/font/NotoSansCJK/NotoSansCJKtc-Regular.ttf"

func loadMIR1201TestFont(t *testing.T) *utf8FontFile {
	t.Helper()
	data, err := os.ReadFile(mir1201TestFontPath)
	if err != nil {
		t.Skipf("テスト用フォントが見つからないためスキップ: %v", err)
	}
	utf := newUTF8Font(&fileReader{array: data})
	if err := utf.parseFile(); err != nil {
		t.Fatalf("parseFile failed: %v", err)
	}
	return utf
}

// mir1201AllSupplementaryPlaneUsedRunesは、以降のテストで共通して使う
// usedRunes。補助面(グリフあり/なし)・私用領域(BMP/補助面)・異体字
// セレクタ・通常のBMP内文字を混在させ、実際に契約書面へ書き込まれうる
// 値のばらつきを再現している。
var mir1201AllSupplementaryPlaneUsedRunes = map[int]int{
	0x25F5C:  0x25F5C,  // 𥽜、グリフあり、ミラリタ1803・ZOLOME372の氏名
	0x20A64:  0x20A64,  // 𠩤、グリフあり、ZOLOME289の氏名
	0x20BB7:  0x20BB7,  // 𠮷、グリフあり
	0xE000:   0xE000,   // グリフ無し、BMP私用領域
	0x103FCB: 0x103FCB, // グリフ無し、補助面私用領域(最初に発見した不具合の文字)
	0xE0100:  0xE0100,  // 異体字セレクタ(IVS)、maxSupportedRune以上
	0x5C71:   0x5C71,   // 山、通常のBMP内文字
}

// TestGenerateCutFontNoPanicWithSupplementaryPlaneは、実際に使われる
// GenerateCutFont(サブセットフォント生成)を、補助面・私用領域・異体字
// セレクタを含むusedRunesで呼び出し、panicしないことを確認する。パッチ前は
// ここでindex out of rangeのpanicが発生していた(MIR-1201で最初に発見した
// 不具合そのもの)。
func TestGenerateCutFontNoPanicWithSupplementaryPlane(t *testing.T) {
	utf := loadMIR1201TestFont(t)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("GenerateCutFont panicked: %v", r)
		}
	}()

	out := utf.GenerateCutFont(mir1201AllSupplementaryPlaneUsedRunes)
	if len(out) == 0 {
		t.Fatal("GenerateCutFont returned empty output")
	}

	// LastRuneがmaxSupportedRuneを超えていないこと(cidToGidMap/CharWidths
	// の配列サイズを超えていないこと)を確認する。0xE0100(異体字セレクタ)は
	// maxSupportedRune以上なので、ここに反映されてはいけない。
	if utf.LastRune >= maxSupportedRune {
		t.Fatalf("LastRune = %d, maxSupportedRune(%d)以上になってはいけない", utf.LastRune, maxSupportedRune)
	}
}

// TestFormat12SupplementaryPlaneGlyphsは、MIR-1201のパッチ本体
// (cmap format 12読み取り)が、実フォントのUnicode補助面の文字を正しく
// グリフIDへ解決できることを確認する。期待するグリフID(gid)は、
// golang.org/x/image/font/sfntで独立に確認した値(GlyphIndex)と一致する。
//
// charSymbolDictionary(Unicodeコードポイント→グリフID)は、
// GenerateCutFont内部のgenerateCMAP()が構築する(parseFile単体では
// 構築されない、gofpdf自体の設計)ため、先にGenerateCutFontを呼んでおく。
func TestFormat12SupplementaryPlaneGlyphs(t *testing.T) {
	utf := loadMIR1201TestFont(t)
	utf.GenerateCutFont(mir1201AllSupplementaryPlaneUsedRunes)

	tests := []struct {
		name    string
		char    int
		wantGID int
	}{
		{"𥽜 U+25F5C ミラリタ1803・ZOLOME372の氏名に含まれる字", 0x25F5C, 60776},
		{"𠩤 U+20A64 ZOLOME289の氏名に含まれる字", 0x20A64, 59582},
		{"𠮷 U+20BB7", 0x20BB7, 59596},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gid, ok := utf.charSymbolDictionary[tt.char]
			if !ok {
				t.Fatalf("U+%X: charSymbolDictionaryに登録されていない(format12が読めていない)", tt.char)
			}
			if gid != tt.wantGID {
				t.Errorf("U+%X: gid = %d, want %d", tt.char, gid, tt.wantGID)
			}
		})
	}
}

// TestNoGlyphCharactersAreNotRegisteredは、グリフを持たない文字が
// charSymbolDictionaryに登録されない(=描画されず欠落する)ことを確認する。
// UniPDF現行の挙動(グリフが無い文字は欠落する)と合わせるための仕様。
func TestNoGlyphCharactersAreNotRegistered(t *testing.T) {
	utf := loadMIR1201TestFont(t)
	utf.GenerateCutFont(mir1201AllSupplementaryPlaneUsedRunes)

	tests := []struct {
		name string
		char int
	}{
		{"私用領域(BMP) U+E000", 0xE000},
		{"私用領域(補助面) U+103FCB", 0x103FCB},
		{"異体字セレクタ U+E0100", 0xE0100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gid, ok := utf.charSymbolDictionary[tt.char]; ok {
				t.Errorf("U+%X: グリフが無いはずが登録されている(gid=%d)", tt.char, gid)
			}
		})
	}
}

// TestGlyphAliasResolvesToSourceGIDは、GlyphAliases(私用領域の番号→補助面
// の文字)を登録した場合に、私用領域の番号がその補助面文字と同じグリフIDを
// 指すようになることを確認する(MIR-1201、河内さんの指示による「空いている
// 番号を借りて描く」方式の中核部分)。ここではpkg/contract側が呼ぶ
// SetGlyphAliasは経由せず、utf8FontFile.GlyphAliasesを直接設定して
// パッチ本体(applyGlyphAliases)だけを検証する。
func TestGlyphAliasResolvesToSourceGID(t *testing.T) {
	utf := loadMIR1201TestFont(t)
	utf.GlyphAliases = map[int]int{
		0xE000: 0x20BB7, // 𠮷
		0xE001: 0x25F5C, // 𥽜
	}

	usedRunes := map[int]int{
		0xE000: 0xE000,
		0xE001: 0xE001,
		0x5C71: 0x5C71, // 山、比較用の通常文字
	}
	utf.GenerateCutFont(usedRunes)

	tests := []struct {
		name       string
		alias      int
		wantSameAs int
	}{
		{"U+E000は𠮷(U+20BB7)と同じGIDを指す", 0xE000, 0x20BB7},
		{"U+E001は𥽜(U+25F5C)と同じGIDを指す", 0xE001, 0x25F5C},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aliasGID, ok := utf.charSymbolDictionary[tt.alias]
			if !ok {
				t.Fatalf("U+%X: charSymbolDictionaryに登録されていない", tt.alias)
			}
			// wantSameAsのGIDは、format12のパッチで別途確認済みの値
			// (TestFormat12SupplementaryPlaneGlyphs)と照合する。
			wantGID := map[int]int{0x20BB7: 59596, 0x25F5C: 60776}[tt.wantSameAs]
			if aliasGID != wantGID {
				t.Errorf("gid = %d, want %d(U+%Xと同じ)", aliasGID, wantGID, tt.wantSameAs)
			}
		})
	}
}

// TestGlyphAliasSharesWidthWithSourceは、エイリアス(私用領域の番号)の
// 文字幅(CharWidths)が、元の補助面文字と同じ値になることを確認する。
// widthAtSize(FitFontSize)の計算がエイリアス経由でも正しく行えるために
// 必要な性質(MIR-1201)。
func TestGlyphAliasSharesWidthWithSource(t *testing.T) {
	utf := loadMIR1201TestFont(t)
	utf.GlyphAliases = map[int]int{0xE000: 0x20BB7}

	usedRunes := map[int]int{0xE000: 0xE000, 0x20BB7: 0x20BB7}
	utf.GenerateCutFont(usedRunes)

	aliasWidth := utf.CharWidths[0xE000]
	sourceWidth := utf.CharWidths[0x20BB7]
	if aliasWidth == 0 {
		t.Fatal("CharWidths[0xE000] is 0, alias width was not applied")
	}
	if aliasWidth != sourceWidth {
		t.Errorf("CharWidths[0xE000] = %d, want same as CharWidths[0x20BB7] = %d", aliasWidth, sourceWidth)
	}
}

// TestBuildToUnicodeWithAliasesは、ToUnicode CMapに、エイリアス(私用領域の
// 番号)から元の補助面文字(UTF-16サロゲートペア)への対応が、bfcharとして
// 追加されることを確認する(MIR-1201)。コピー&ペーストやテキスト検索で
// 元の文字に戻るようにするための仕組み。
func TestBuildToUnicodeWithAliases(t *testing.T) {
	t.Run("エイリアスが無ければ従来通りの固定文字列", func(t *testing.T) {
		got := buildToUnicode(nil)
		if got != toUnicodeHeader+toUnicodeFooter {
			t.Errorf("buildToUnicode(nil) should equal the fixed identity CMap")
		}
	})

	t.Run("エイリアスは補助面文字のサロゲートペアに対応づけられる", func(t *testing.T) {
		got := buildToUnicode(map[int]int{0xE000: 0x20BB7})
		// 𠮷(U+20BB7)のUTF-16サロゲートペアはD842 DFB7。
		for _, want := range []string{"1 beginbfchar", "<E000> <D842DFB7>", "endbfchar"} {
			if !strings.Contains(got, want) {
				t.Errorf("buildToUnicode output missing %q:\n%s", want, got)
			}
		}
	})
}

// TestSetGlyphAliasAppliesWidthToCurrentFontは、Fpdf.SetGlyphAlias(公開API)
// が、実際にGetStringWidth(FitFontSizeの幅計算が使う経路)へも反映される
// ことを確認する(MIR-1201、②レビューで発見した不具合)。
//
// TestGlyphAliasSharesWidthWithSourceはutf8FontFile.CharWidthsを直接見て
// いるが、これはGenerateCutFont(PDF出力時にのみ呼ばれる)が作り直す配列で
// あり、fontDefType.Cw(GetStringWidthが参照する配列)とは別物になり得る。
// SetGlyphAlias単体でCwへ反映されているかどうかは、このテストでしか検出
// できない。
func TestSetGlyphAliasAppliesWidthToCurrentFont(t *testing.T) {
	pdf := New("P", "pt", "A4", "")
	pdf.AddUTF8Font("noto", "", mir1201TestFontPath)
	pdf.SetFont("noto", "", 10)
	if pdf.Err() {
		t.Fatalf("AddUTF8Font/SetFont failed: %v", pdf.Error())
	}

	if err := pdf.SetGlyphAlias("noto", "", 0xE000, 0x20BB7); err != nil {
		t.Fatalf("SetGlyphAlias failed: %v", err)
	}

	aliasWidth := pdf.GetStringWidth(string(rune(0xE000)))
	sourceWidth := pdf.GetStringWidth(string(rune(0x20BB7)))
	if aliasWidth != sourceWidth {
		t.Errorf("GetStringWidth(alias) = %v, want same as GetStringWidth(source) = %v (MissingWidthにフォールバックしている疑い)", aliasWidth, sourceWidth)
	}
}

// TestSetGlyphAliasErrorsOnUnknownFontは、Fpdf.SetGlyphAlias(公開API)が、
// 未登録のフォントを指定した場合にエラーを返すことを確認する(MIR-1201)。
func TestSetGlyphAliasErrorsOnUnknownFont(t *testing.T) {
	pdf := NewCustom(&InitType{
		OrientationStr: "P",
		UnitStr:        "pt",
		Size:           SizeType{Wd: 595, Ht: 842},
	})
	if err := pdf.SetGlyphAlias("not-registered", "", 0xE000, 0x20BB7); err == nil {
		t.Error("SetGlyphAlias should error for a font that was never added via AddUTF8Font")
	}
}
