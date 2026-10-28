package contract

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"
	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpumodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	pdfcputypes "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/pkg/errors"
)

// init: pdfcpuのNewDefaultConfigurationは、既定では設定ディレクトリ
// ($XDG_CONFIG_HOME または $HOME/.config 配下)へ書き込み、失敗すると
// os.Exit(1)でプロセスごと終了する(呼び出し元へのエラー伝播ではない)。
// invest-worker/jobs-schedulerの実行ユーザーでこのディレクトリが書き込み
// 可能かは未確認で、書き込めない環境ではAPI全体が落ちる(②レビューで
// 指摘、MIR-1201)。ここで使うAPI(AcroForm読み取り・RemoveFormFields・
// PDFWatermarkForReadSeekerによる画像スタンプ)はユーザーフォントを必要と
// しないため、DisableConfigDirで設定ディレクトリへの依存自体を無くす。
func init() {
	pdfcpuapi.DisableConfigDir()
}

// contractFontNameは、gofpdfへ登録するフォントの識別名(任意の文字列でよい)。
const contractFontName = "noto"

// contractFontFileは、繁体字版のCJKフォント。UniPDF時代からファイル自体は
// 変えず、日本語版への切り替えは別作業として先送りしている(MIR-1201)。
const contractFontFile = "NotoSansCJKtc-Regular.ttf"

// placementは、1つの記入項目を1箇所の矩形へ書き込むための情報。
type placement struct {
	text     string
	rect     *pdfcputypes.Rectangle
	quadding int
}

// fillContractTemplateは、テンプレート(templateData)のAcroFormフィールドの
// うち、valuesにキーがあるものだけ値を書き込み、フォーム定義を取り除いた
// PDFを返す。フォーム項目を持たないテンプレートは何も書き込まず、
// templateDataをそのまま返す(MIR-1201、既存のUniPDF実装と同じ挙動)。
//
// UniPDFのAcroForm.Fill + FlattenFields + SkipAcroFormと同じ結果になるよう、
// pdfcpuの低レベルAPIでフィールド定義(矩形)を読み取り、gofpdfで
// フィールドと同じ位置にテキストを描いたオーバーレイページを生成して
// スタンプとして重ね、最後にAcroForm定義を削除する。
func fillContractTemplate(fontDir string, templateData []byte, values map[string]string) (*bytes.Buffer, error) {
	// pdfcpu(v0.8.0)のReadValidateAndOptimizeは、ファイルサイズが512バイト
	// 未満だとその値を読み取り単位として使う実装になっており、0バイトだと
	// 単位が0になってSeekが無限に成功し続け、返らなくなる(ZOLOME側の
	// ①敵対的検証で発見、MIR-1201)。テンプレート取得元(S3/HTTP)が失敗
	// すると空のバイト列がここへ渡り得る。UniPDFの旧実装は空データを
	// 即座にエラーとして返していたため、ここで同じ挙動に揃える。
	if len(templateData) == 0 {
		return nil, errors.New("fail contract template data is empty")
	}

	// 異体字セレクタ(IVS/SVS)はここで取り除く。直前の漢字自体は残るため、
	// UniPDF現行と同じく「元の漢字がそのまま描かれる」形になる(MIR-1201)。
	strippedValues := make(map[string]string, len(values))
	for k, v := range values {
		strippedValues[k] = stripVariationSelectors(v)
	}
	values = strippedValues

	// 続けて、Unicode補助面の文字(拡張漢字等)をBMP私用領域の番号に
	// 置き換える(astralAliasTable、MIR-1201)。テーブルは1回の書面生成の
	// 中でだけ有効で、同じ文字には同じ番号を割り当てる。値の中に本物の
	// 私用領域の文字(外字)が既にあれば、その番号は避ける(レビューで
	// 発見した不具合。避けないと、補助面の文字にたまたま同じ番号を割り当てて
	// しまい、本物の外字が別の拡張漢字のグリフに置き換わって見えてしまう)。
	astralTable := newAstralAliasTable(collectPrivateUseRunes(values))
	sanitizedValues := make(map[string]string, len(values))
	for k, v := range values {
		sanitizedValues[k] = astralTable.remap(v)
	}
	values = sanitizedValues

	conf := pdfcpumodel.NewDefaultConfiguration()

	ctx, err := pdfcpuapi.ReadValidateAndOptimize(bytes.NewReader(templateData), conf)
	if err != nil {
		return nil, errors.Wrap(err, "fail contract read template")
	}

	fields, err := readAcroFormFields(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail contract read acroform fields")
	}

	byPage := map[int][]placement{}
	// allFieldNamesは、値の有無に関わらずテンプレートに存在する全フィールド
	// 名。RemoveFormFieldsに渡し、値を書き込まなかったフィールドも含めて
	// フォーム定義を消す。UniPDFのFlattenFields(true, ...)は値の有無に
	// 関わらず全フィールドを平坦化しており、値を渡さなかったフィールドが
	// 入力可能な欄のまま残ることはなかった。当初は値を書き込んだフィールド
	// だけをRemoveFormFieldsの対象にしており、値の無い項目(既存テストの
	// date2/shares_count等)が入力可能な欄のまま残ってしまっていた
	// (敵対的検証で発見した不具合、MIR-1201)。
	var allFieldNames []string
	for name, widgets := range fields {
		allFieldNames = append(allFieldNames, name)

		value, ok := values[name]
		if !ok {
			continue
		}
		for _, w := range widgets {
			if w.page == 0 {
				// ウィジェットが属するページを特定できなかった場合。想定外の
				// テンプレート構造なので、安全側に倒して書き込まない。
				continue
			}
			byPage[w.page] = append(byPage[w.page], placement{
				text:     value,
				rect:     w.rect,
				quadding: w.quadding,
			})
		}
	}

	if len(allFieldNames) == 0 {
		return bytes.NewBuffer(templateData), nil
	}

	dims, err := pdfcpuapi.PageDims(bytes.NewReader(templateData), conf)
	if err != nil {
		return nil, errors.Wrap(err, "fail contract page dims")
	}

	current := templateData
	for pageNr := 1; pageNr <= len(dims); pageNr++ {
		placements := byPage[pageNr]
		if len(placements) == 0 {
			continue
		}
		dim := dims[pageNr-1]

		overlay, err := renderOverlayPage(fontDir, dim.Width, dim.Height, placements, astralTable.aliases)
		if err != nil {
			return nil, errors.Wrapf(err, "fail contract render overlay page %d", pageNr)
		}

		// MIR-1201: rotation:0を明示しないと、pdfcpuの既定(ページ対角線に
		// 沿った斜め配置、DiagonalLLToUR)になり、書き込んだ文字が全部斜めに
		// 傾いた位置に描かれてしまう(敵対的検証で発見した重大な不具合)。
		wm, err := pdfcpuapi.PDFWatermarkForReadSeeker(bytes.NewReader(overlay), 1, "scale:1 abs, rotation:0, position:c", true, false, pdfcputypes.POINTS)
		if err != nil {
			return nil, errors.Wrapf(err, "fail contract pdf watermark page %d", pageNr)
		}

		out := &bytes.Buffer{}
		if err := pdfcpuapi.AddWatermarks(bytes.NewReader(current), out, []string{fmt.Sprint(pageNr)}, wm, conf); err != nil {
			return nil, errors.Wrapf(err, "fail contract add watermark page %d", pageNr)
		}
		current = out.Bytes()
	}

	out := &bytes.Buffer{}
	if err := pdfcpuapi.RemoveFormFields(bytes.NewReader(current), out, allFieldNames, conf); err != nil {
		return nil, errors.Wrap(err, "fail contract remove form fields")
	}

	final, err := removeOptionalContent(out.Bytes(), conf)
	if err != nil {
		return nil, errors.Wrap(err, "fail contract remove optional content")
	}

	return bytes.NewBuffer(final), nil
}

// removeOptionalContentは、pdfcpuのwatermark機能がスタンプに付与する
// オプショナルコンテンツ(レイヤー)の参照を取り除く。契約締結時交付書面は
// 法定の書面であり、記入した氏名や金額がレイヤーの中にあると、閲覧ソフトで
// レイヤーが非表示にされたり、レイヤーに対応しない変換ツールを通したりした
// 場合に、記入欄が空の書面に見えてしまうおそれがある。UniPDFの出力は平坦化
// された通常のページコンテンツで、レイヤーではなかったため、それに合わせる
// (敵対的検証の指摘を受けてPMと相談し追加した。MIR-1201)。
//
// pdfcpuのAddWatermarksはOCGを付けずに重ねるオプションを提供していないため、
// 重ねた後にCatalogのOCPropertiesと、各ページのForm XObject辞書が持つ/OC
// エントリを取り除く後処理で対応する。
//
// テンプレート自身が元々レイヤーを持っている場合、この処理で巻き添えに
// 消してしまう懸念があったが、ミラリタの本番テンプレート44件・ZOLOMEの
// 4件を確認した限り、いずれもOCPropertiesを持たない(②レビューで指摘、
// MIR-1201)。
func removeOptionalContent(data []byte, conf *pdfcpumodel.Configuration) ([]byte, error) {
	ctx, err := pdfcpuapi.ReadValidateAndOptimize(bytes.NewReader(data), conf)
	if err != nil {
		return nil, errors.Wrap(err, "fail contract read for ocg removal")
	}

	ctx.RootDict.Delete("OCProperties")

	for pageNr := 1; pageNr <= ctx.PageCount; pageNr++ {
		pageDict, _, _, err := ctx.PageDict(pageNr, false)
		if err != nil {
			continue
		}
		resDict, err := ctx.DereferenceDict(pageDict["Resources"])
		if err != nil {
			continue
		}
		xobjDict, err := ctx.DereferenceDict(resDict["XObject"])
		if err != nil {
			continue
		}
		for _, v := range xobjDict {
			ref, ok := v.(pdfcputypes.IndirectRef)
			if !ok {
				continue
			}
			sd, err := ctx.DereferenceXObjectDict(ref)
			if err != nil {
				continue
			}
			sd.Delete("OC")
		}
	}

	out := &bytes.Buffer{}
	if err := pdfcpuapi.Write(ctx, out, conf); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// newContractPdf は、renderOverlayPage用のgofpdfインスタンスを1枚作る
// (実際の出力用・幅計測用の両方で同じ設定を使うための共通処理)。
// astralAliases(補助面の文字→割り当てた私用領域の番号)は、gofpdf側の
// グリフ解決に反映するため、フォント登録の直後にSetGlyphAliasで登録する
// (MIR-1201)。
func newContractPdf(fontDir string, widthPt, heightPt float64, astralAliases map[rune]rune) *gofpdf.Fpdf {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "pt",
		Size:           gofpdf.SizeType{Wd: widthPt, Ht: heightPt},
		FontDirStr:     fontDir,
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddUTF8Font(contractFontName, "", contractFontFile)
	// astralAliasesの実体はastralAliasTable.aliases(「補助面の文字→割り
	// 当てた私用領域の番号」)なので、mapのキーがsource(実際の文字)・値が
	// alias(私用領域の番号)になる。SetGlyphAliasの引数順(alias, source)に
	// 合わせて、ループ変数もその対応で受け取る(MIR-1201、最初はここを
	// 取り違えて逆に登録しており、グリフが解決されずCIDToGIDMapがGID 0を
	// 指す不具合になっていた)。
	for source, alias := range astralAliases {
		// AddUTF8Font自体が失敗した場合(フォントファイルが読めない等)は
		// ここもエラーになるが、その時点でpdf.Err()は既にtrueになっており、
		// renderOverlayPage側の`if pdf.Err()`チェックで検出できるため、
		// ここで個別にハンドリングする必要はない(②レビューで指摘、
		// コメントの「必ず成功する」という説明が不正確だったため修正。
		// MIR-1201)。
		_ = pdf.SetGlyphAlias(contractFontName, "", alias, source)
	}
	pdf.AddPage()
	// SetFontはAddPageの後で呼ぶこと。gofpdfのAddPageは、直前に設定されて
	// いたフォントを内部で再適用する仕様のため、AddPageより前に呼ぶと
	// フォント設定のTfオペレータが余分に(二重に)出力される(MIR-1201、
	// 実測して確認した)。
	pdf.SetFont(contractFontName, "", 12)
	return pdf
}

// renderOverlayPageは、1ページ分の透明な背景に、placementsで指定された
// 位置・サイズでテキストだけを描いたPDFを1枚生成する。
func renderOverlayPage(fontDir string, widthPt, heightPt float64, placements []placement, astralAliases map[rune]rune) ([]byte, error) {
	pdf := newContractPdf(fontDir, widthPt, heightPt, astralAliases)

	// 文字列幅の計測専用インスタンス。gofpdfのSetFontSizeは呼ぶたびに
	// content streamへTfオペレータを直接書き込む実装のため、出力用の
	// pdfでSetFontSizeを幅計測のためだけに呼ぶと、実際には使わない
	// フォントサイズのTfが混入してしまう(MIR-1201、実測して確認した。
	// 1フィールドあたり本来1回で済むはずのTfが、初期化分と合わせて
	// 数倍に増えていた)。出力に使わない別インスタンスで計測することで、
	// pdf側には実際に使うフォントサイズのTfだけが残るようにする。
	measure := newContractPdf(fontDir, widthPt, heightPt, astralAliases)

	// TextOriginの計算にはフォントのCapHeight(1000 unitsPerEmスケール)が
	// 要る。UniPDFのannotator実装と同じ値を使うため、gofpdfが内部で
	// TTFから読み取ったCapHeightをそのまま使う(MIR-1201)。
	capHeightRatio := float64(pdf.GetFontDesc(contractFontName, "").CapHeight) / 1000.0

	for _, p := range placements {
		widthAtSize := func(size float64) float64 {
			measure.SetFontSize(size)
			return measure.GetStringWidth(p.text)
		}

		fontSize := FitFontSize(p.text, p.rect.Width(), p.rect.Height(), widthAtSize)
		textWidthPt := widthAtSize(fontSize)

		dx, dy := TextOrigin(fontSize, p.rect.Width(), p.rect.Height(), textWidthPt, capHeightRatio, p.quadding)
		baseX := p.rect.LL.X + dx
		baseY := p.rect.LL.Y + dy

		pdf.SetFontSize(fontSize)
		pdf.Text(baseX, heightPt-baseY, p.text)
	}

	if pdf.Err() {
		return nil, errors.Wrap(pdf.Error(), "fail contract render overlay page gofpdf")
	}

	buf := &bytes.Buffer{}
	if err := pdf.Output(buf); err != nil {
		return nil, errors.Wrap(err, "fail contract render overlay page output")
	}
	return buf.Bytes(), nil
}
