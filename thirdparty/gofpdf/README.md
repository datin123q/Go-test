# gofpdf (patched, MIR-1201)

[jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf) v1.16.2 を元に、
Unicode補助面(基本多言語面を超える文字。CJK拡張B等の拡張漢字、一部の異体字
を含む)を扱えるようにパッチを当てて取り込んだもの。

jung-kurt/gofpdf はアーカイブされ開発が終了しており、後継として案内される
[go-pdf/fpdf](https://github.com/go-pdf/fpdf)(v0.9.0時点)も同じ制約を
引き継いでいたため、フォークして直接パッチを当てる方針にした。詳細な経緯は
JIRA MIR-1201を参照。

importパスは元のまま `github.com/jung-kurt/gofpdf` を使う。各モジュールの
`go.mod` の `replace` でこのディレクトリに差し替えている
(`replace github.com/jung-kurt/gofpdf => .../thirdparty/gofpdf`)。

## 何が壊れていたか

元の実装は、TrueTypeフォントのcmapテーブルのうちformat 4(基本多言語面、
U+0000-U+FFFFのみ)しか読んでおらず、それを前提に内部配列(文字幅のテーブル
`CharWidths`、CIDからグリフIDへの変換テーブル`cidToGidMap`)を
`256*256`(=0x10000)固定で確保していた。

一方で、実際に使われた文字を集計する処理(`parseSymbols`)は、cmapで
見つかったかどうかに関係なく、渡された文字のUnicodeコードポイントの最大値を
そのまま`LastRune`として記録していた。補助面の文字(コードポイントが
0x10000以上)を1文字でも含む文字列を描画しようとすると、`LastRune`が
上記の固定サイズを超え、内部配列への範囲外アクセスでpanicしていた。

フォントに実際にグリフがある補助面の文字(拡張漢字等)であっても、cmap
format 4では表現できないため、この問題は避けられなかった。

## 当てたパッチ

対象ファイルは `utf8fontfile.go` と `fpdf.go`。変更箇所には
`MIR-1201`を含むコメントを付けている(`grep -rn MIR-1201`で一覧できる)。

1. **cmap format 12の読み取りを追加**(`utf8fontfile.go`:
   `parseCMAPTable`, `generateCMAP`, `generateSCCSDictionariesFormat12`)
   format 4のサブテーブルを見つけても走査を打ち切らず、format 12
   (補助面に対応するグループ形式のcmap)のサブテーブルも探す。見つかった
   場合、そのグループ(開始文字コード・終了文字コード・開始グリフID)を
   読み、format 4で見つからなかった文字だけ追加登録する。

2. **内部配列のサイズをmaxSupportedRune(0x30000)に統一**
   (`utf8fontfile.go`の`CharWidths`、`fpdf.go`の`cidToGidMap`)
   0x30000は元のコードに既にあった`char < 196608`という条件
   (`parseHMTXTable`等)に合わせた値で、基本多言語面・補助多言語面・
   補助表意文字面(CJK拡張Bを含む)をカバーする。

3. **`LastRune`の更新にmaxSupportedRuneのガードを追加**
   (`utf8fontfile.go`: `parseSymbols`)
   0x30000以上の文字(異体字セレクタ等)がusedRunesに含まれていても
   `LastRune`を更新しないようにし、1・2のパッチを当てても内部配列を
   超えるインデックスにならないようにしている。この結果、
   0x30000以上の文字は(呼び出し側で事前に取り除かれていなくても)
   描画されず、欠落する形になる。パニックしないことを最優先にした
   フェイルセーフ。

4. **グリフの番号を借りる仕組み(GlyphAliases)を追加**
   (`utf8fontfile.go`: `GlyphAliases`フィールド, `applyGlyphAliases`,
   `buildToUnicode`/`utf16HexHighLow`。`fpdf.go`:
   `Fpdf.SetGlyphAlias`、ToUnicode生成呼び出し箇所)
   1〜3のパッチでグリフの解決(文字→グリフID)はできても、gofpdfは
   CID(PDFの文字コード)にUnicodeコードポイントをそのまま使う設計で、
   実際にPDFへ書き込むIdentity-Hエンコーディングは2バイト
   (0-0xFFFF)のCIDしか表現できない。補助面の文字(0x10000以上)は
   この時点でまだ正しく描画できなかった。
   `Fpdf.SetGlyphAlias(familyStr, styleStr, aliasRune, sourceRune)`で、
   BMP内の空き番号(aliasRune)を、実際に描画したい文字(sourceRune、
   補助面の文字)に対応づけて登録できるようにした。登録すると、
   `applyGlyphAliases`が`charSymbolDictionary`(文字→グリフ)・
   `CharWidths`(文字幅)の両方に、aliasRuneがsourceRuneと同じ値を
   持つよう反映する。`buildToUnicode`は、ToUnicode CMapに
   「aliasRune→sourceRuneのUTF-16(サロゲートペア)」をbfcharとして
   追加し、テキスト抽出やコピー&ペーストでも元の文字に戻るようにする。
   呼び出し側(`pkg/contract`のastralAliasTable)が、実際に描画したい
   補助面の文字をBMP私用領域の空き番号に一時的に置き換えてから
   `Text()`に渡し、対応するエイリアスをこの仕組みで登録する、という
   使い方を想定している。

## 意図した見た目

- グリフがある補助面の文字(拡張漢字等): 正しく描画される
  (パッチ適用前のUniPDFでは誤って別の文字に化けていたケースがあった。
  新方式ではその化けを踏襲せず、正しく描く。パッチ4の「番号を借りる」
  仕組みが無いと、CIDが2バイトに収まらずグリフが解決できないため、
  1〜3のパッチだけでは実現できなかった)
- グリフが無い文字(私用領域等): 描画されない(欠落する)。UniPDFの現行
  実装と同じ挙動
- 異体字セレクタ(IVS/SVS): 呼び出し側(`pkg/contract`)で事前に取り除く
  実装になっているため、直前の漢字だけがそのまま描画される。仮に
  取り除かれずに渡ってきた場合も、上記2のガードにより欠落するだけで
  パニックはしない

## 更新しないファイル

このディレクトリには、メインパッケージ(`gofpdf`)が実際に使うファイルだけを
取り込んでいる。`contrib/`, `doc/`, `font/`, `image/`, `internal/`,
`list/`, `makefont/`, `pdf/`などのサブディレクトリや、テストファイル
(`fpdf_test.go`, `ttfparser_test.go`)は取り込んでいない。

アップストリームの更新は追従しない(アーカイブ済みのため今後の更新は
想定していない)。将来的にパッチ以外の部分を更新したくなった場合は、
元のリポジトリの該当バージョンと`diff`を取り、このディレクトリの
`MIR-1201`コメント部分を残したまま反映すること。

## ライセンス

MIT License。`LICENSE`ファイルに原文をそのまま含めている。

```
MIT License

Copyright (c) 2017 Kurt Jung and contributors acknowledged in the documentation
```
