package contract

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpumodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFillFields(t *testing.T) {
	InitContract()

	type args struct {
		templateData []byte
		price        string
		amount       string
		sharesNum    string
		name         string
		address      string
		managerName  string
		spvName      string
		projectName  string
		date1        jsonutils.JSONTime
	}

	templateData, err := os.ReadFile("../../assets/testdata/contract/契約締結時交付書面_20220401_FORM_NEW.pdf")
	require.NoError(t, err)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "私用領域の外字を含む長い氏名でもフォーム項目のあるテンプレートを埋められる",
			args: args{
				templateData: templateData,
				price:        "100",
				amount:       "1000",
				sharesNum:    "10",
				name:         "\U00103FCB\U00106B4A\U001038D6\U00109FA5",
				address:      "東京都千代田区千代田1-1",
				managerName:  "山田太郎",
				spvName:      "test",
				projectName:  "test",
				date1:        jsonutils.JSONTime(time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC)),
			},
			wantErr: false,
		},
		{
			// ミラリタ1803・ZOLOME372の氏名に含まれる字(会員IDのみ、氏名本体は
			// 使わない。MIR-1201)。グリフがあるので正しく描画される想定。
			name: "Unicode補助面のCJK拡張B(𥽜)を含む氏名を埋められる",
			args: args{
				templateData: templateData,
				price:        "100",
				amount:       "1000",
				sharesNum:    "10",
				name:         "山田\U00025F5C太郎",
				address:      "東京都千代田区千代田1-1",
				managerName:  "山田太郎",
				spvName:      "test",
				projectName:  "test",
				date1:        jsonutils.JSONTime(time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC)),
			},
			wantErr: false,
		},
		{
			// ZOLOME289の氏名に含まれる字(会員IDのみ。MIR-1201)。
			name: "Unicode補助面のCJK拡張B(𠩤)を含む氏名を埋められる",
			args: args{
				templateData: templateData,
				price:        "100",
				amount:       "1000",
				sharesNum:    "10",
				name:         "山田\U00020A64太郎",
				address:      "東京都千代田区千代田1-1",
				managerName:  "山田太郎",
				spvName:      "test",
				projectName:  "test",
				date1:        jsonutils.JSONTime(time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC)),
			},
			wantErr: false,
		},
		{
			name: "Unicode補助面のCJK拡張B(𠮷)を含む氏名を埋められる",
			args: args{
				templateData: templateData,
				price:        "100",
				amount:       "1000",
				sharesNum:    "10",
				name:         "山田\U00020BB7太郎",
				address:      "東京都千代田区千代田1-1",
				managerName:  "山田太郎",
				spvName:      "test",
				projectName:  "test",
				date1:        jsonutils.JSONTime(time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC)),
			},
			wantErr: false,
		},
		{
			// 異体字セレクタ(IVS)付きの漢字。セレクタは取り除かれ、直前の
			// 漢字("辻")だけが描画される想定(UniPDF現行と同じ挙動)。
			name: "異体字セレクタ付きの漢字を含む氏名を埋められる",
			args: args{
				templateData: templateData,
				price:        "100",
				amount:       "1000",
				sharesNum:    "10",
				name:         "辻\U000E0100田太郎",
				address:      "東京都千代田区千代田1-1",
				managerName:  "山田太郎",
				spvName:      "test",
				projectName:  "test",
				date1:        jsonutils.JSONTime(time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC)),
			},
			wantErr: false,
		},
		{
			// ミラリタ9310・9560の氏名に含まれる異体字セレクタ(出資実績なし)。
			name: "BMP内の私用領域の文字を含む氏名を埋められる",
			args: args{
				templateData: templateData,
				price:        "100",
				amount:       "1000",
				sharesNum:    "10",
				name:         "山田太郎",
				address:      "東京都千代田区千代田1-1",
				managerName:  "山田太郎",
				spvName:      "test",
				projectName:  "test",
				date1:        jsonutils.JSONTime(time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC)),
			},
			wantErr: false,
		},
		{
			name: "空の値でもエラーにならない",
			args: args{
				templateData: templateData,
				price:        "",
				amount:       "",
				sharesNum:    "",
				name:         "",
				address:      "",
				managerName:  "",
				spvName:      "",
				projectName:  "",
				date1:        jsonutils.JSONTime(time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// MIR-1201: どんな文字が来てもpanicしないことが最優先(invest-worker/
			// ContractCreateJobの両方が落ちる経路のため)。
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("FillFields panicked: %v", r)
				}
			}()
			got, err := FillFields(tt.args.templateData, tt.args.price, tt.args.amount, tt.args.sharesNum, tt.args.name, tt.args.spvName, tt.args.projectName, tt.args.address, tt.args.managerName, tt.args.date1)
			if (err != nil) != tt.wantErr {
				t.Errorf("FillFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.NotNil(t, got)
			assert.Positive(t, got.Len())

			// テンプレート(契約締結時交付書面_20220401_FORM_NEW.pdf)のフィールド名は
			// date1/date2/invest_name/project_name/shares_count/spv_name。
			// FillFieldsが値を持つのはこのうちdate1/invest_name/project_name/
			// spv_nameだけで、shares_count(share_countではない)とdate2は
			// 値マップに無い。UniPDFのFlattenFields(true, ...)は値の有無に
			// 関わらず全フィールドを平坦化するため、値の無いフィールドも
			// 含めて全部AcroFormから消える想定(MIR-1201、敵対的検証で
			// 「値の無い項目が入力可能な欄のまま残る」不具合を発見し修正した)。
			ctx, err := pdfcpuapi.ReadValidateAndOptimize(bytes.NewReader(got.Bytes()), pdfcpumodel.NewDefaultConfiguration())
			require.NoError(t, err)

			fields, err := readAcroFormFields(ctx)
			require.NoError(t, err)
			assert.Empty(t, fields)

			// MIR-1201: 契約締結時交付書面は法定の書面であり、記入した氏名や
			// 金額がオプショナルコンテンツ(レイヤー)の中にあると、閲覧ソフト
			// でレイヤーを非表示にされたり、対応しない変換ツールを通したり
			// した場合に、記入欄が空の書面に見えるおそれがある。pdfcpuの
			// watermark機能は既定でこのレイヤーを付けるため、後処理で
			// 取り除いている。UniPDFの出力はレイヤーではなかった。
			catalog, err := ctx.Catalog()
			require.NoError(t, err)
			assert.NotContains(t, catalog, "OCProperties")
		})
	}
}

// TestFillFieldsResolvesAstralGlyphsは、Unicode補助面の文字(拡張漢字)を
// 含む氏名で、実際にその文字が正しく描画され(化けたり欠落したりせず)、
// かつテキスト抽出でも元の文字のまま取り出せることを確認する
// (MIR-1201、河内さんの指示による「空いている私用領域の番号を借りて描く」
// 方式)。TestFillFieldsは「panicしないこと」しか見ていないため、別立てで
// 用意した。
//
// テキスト抽出はpoppler(pdftotext)で行う(MIR-1341、desk JIRAコメント13884)。
// 以前はUniPDFを使っていたが、ライセンス契約が2026-10末で切れるため、本番の
// コード(MIR-1201で既に置き換え済み)に続きテストからも外した。
func TestFillFieldsResolvesAstralGlyphs(t *testing.T) {
	InitContract()

	templateData, err := os.ReadFile("../../assets/testdata/contract/契約締結時交付書面_20220401_FORM_NEW.pdf")
	require.NoError(t, err)

	date1 := jsonutils.JSONTime(time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name string
		char string
	}{
		{"𥽜 U+25F5C ミラリタ1803・ZOLOME372の氏名に含まれる字", "\U00025F5C"},
		{"𠩤 U+20A64 ZOLOME289の氏名に含まれる字", "\U00020A64"},
		{"𠮷 U+20BB7", "\U00020BB7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			investName := "山田" + tt.char + "太郎"
			spvName := "中央太陽光合同会社" // 同じ書面の他の項目が巻き添えで
			// 化けないことも確認する対象(MIR-1201、敵対的検証で発見した
			// 「4バイトUTF-8の処理漏れで後続の文字列全体がずれる」不具合)。

			got, err := FillFields(templateData, "100", "1000", "10", investName, spvName, "test", "東京都千代田区千代田1-1", "山田太郎", date1)
			require.NoError(t, err)

			text := extractPDFText(t, got.Bytes())

			assert.True(t, strings.Contains(text, investName), "extracted text does not contain the astral name: %q", text)
			assert.True(t, strings.Contains(text, spvName), "extracted text does not contain the SPV name (corrupted by the astral char?): %q", text)
		})
	}
}

// TestFillFieldsRejectsEmptyTemplateは、空のテンプレートを渡した場合に
// 即座にエラーを返すことを確認する(ZOLOME側の①敵対的検証で発見、
// MIR-1201)。pdfcpu(v0.8.0)は512バイト未満のファイルを読み取り単位として
// 扱う実装になっており、0バイトだと単位が0になってSeekが無限に成功し
// 続け、返らなくなる。テンプレート取得元(S3/HTTP)が失敗すると空の
// バイト列がFillFieldsへ渡り得る。UniPDFの旧実装は即座にエラーを
// 返していたため、その挙動に揃える。
func TestFillFieldsRejectsEmptyTemplate(t *testing.T) {
	InitContract()

	done := make(chan struct{})
	var err error
	go func() {
		defer close(done)
		_, err = FillFields(nil, "100", "1000", "10", "山田太郎", "test", "test", "東京都千代田区千代田1-1", "山田太郎", jsonutils.JSONTime(time.Now()))
	}()
	select {
	case <-done:
		assert.Error(t, err, "FillFields with empty template data should return an error")
	case <-time.After(5 * time.Second):
		t.Fatal("FillFields did not return within 5s for empty template data (hang regression)")
	}
}

// extractPDFTextは、pdfデータの1ページ目からテキストを取り出す(poppler-utils
// のpdftotextコマンドを使う、MIR-1341)。以前のUniPDF版も1ページ目しか見て
// いなかったため、-f 1 -l 1で範囲をそろえ、テストの意味を変えない(②レビュー
// の指摘)。pdftotextが手元に無い場合、CIでは黙って飛ばさずt.Fatalで落とす。
// CI判定はgo.ymlが設定するENV=GITHUBに加え、GitHub Actionsが常に設定する
// GITHUB_ACTIONS=trueも見る(将来go.ymlからENV=GITHUBが無くなっても、CIで
// 黙ってt.Skipしないため)。手元の開発環境ではインストールを強制しないため
// t.Skipにする。
func extractPDFText(t *testing.T, pdfData []byte) string {
	t.Helper()

	isCI := os.Getenv("ENV") == "GITHUB" || os.Getenv("GITHUB_ACTIONS") == "true"

	if _, err := exec.LookPath("pdftotext"); err != nil {
		if isCI {
			t.Fatalf("pdftotext not found in PATH (CI must install poppler-utils): %v", err)
		}
		t.Skip("pdftotext not found in PATH (install poppler-utils to run this test locally)")
	}

	tmpFile, err := os.CreateTemp(t.TempDir(), "contract-*.pdf")
	require.NoError(t, err)
	defer tmpFile.Close()
	_, err = tmpFile.Write(pdfData)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	// pdftotextが固まった場合でもテストが10分(CIのジョブ全体のタイムアウトは
	// 20分)張り付かないよう、プロセス単位でも上限を設ける(②レビューの指摘)。
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "pdftotext", "-f", "1", "-l", "1", "-enc", "UTF-8", tmpFile.Name(), "-").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			t.Fatalf("pdftotext failed: %v\n%s", err, exitErr.Stderr)
		}
		t.Fatalf("pdftotext failed: %v", err)
	}

	return string(out)
}

func InitContract() {
	// SetDefaultFontはpathutils.GetPathOfAssets()相当の絶対パス
	// (<repo root>/assets)を期待する(MIR-1201)。
	err := SetDefaultFont("../../assets")
	if err != nil {
		panic(err)
	}
}

func TestAddRightBlankToAmount(t *testing.T) {
	type args struct {
		amount string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success add right blank to 5",
			args: args{
				amount: "5",
			},
			want: "5 ",
		},
		{
			name: "success add right blank to 100",
			args: args{
				amount: "100",
			},
			want: "100 ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, addRightBlankToNumber(tt.args.amount), "addRightBlankToNumber(%v)", tt.args.amount)
		})
	}
}
