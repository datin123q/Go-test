package contract

import (
	pdfcpumodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	pdfcputypes "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/pkg/errors"
)

// fieldWidgetは、AcroFormの1つのフィールド名に対応する1つの配置箇所を表す。
// ZOLOMEのテンプレートは同じフィールド名(invest_name、date1)が複数ページに
// 配置されているため、フィールド名ごとに複数のfieldWidgetを持つ(MIR-1201)。
//
// DA(フィールドごとのデフォルト外観、フォントサイズを含む)は読み取らない。
// UniPDFの実際の出力を本番テンプレート全件と突き合わせた結果、DAで固定
// サイズが指定されていても無視され、常に矩形の高さから計算される自動
// サイズ(FitFontSize)が使われることを確認したため(MIR-1201)。
type fieldWidget struct {
	page int
	rect *pdfcputypes.Rectangle
	// quadding は/Qの値(0=左揃え、1=中央揃え、2=右揃え)。省略時は0。
	// 敵対的検証で、揃えを無視すると本番テンプレートの一部(右揃えの
	// フィールドを含むもの)で見た目がずれることが分かり追加した(MIR-1201)。
	quadding int
}

// readAcroFormFieldsは、テンプレートのAcroFormから、フィールド名ごとの
// 配置箇所(ページ番号・矩形)を読み取る。UniPDFを使わず、pdfcpuの低レベル
// API(XRefTableの生のPDFオブジェクト)だけで読む(MIR-1201、UniPDFの
// ライセンス依存を断つため)。フォーム項目を持たないテンプレートの場合は
// 空のmapを返す(エラーにしない。今と同じく何も書き込まずに出力するため)。
func readAcroFormFields(ctx *pdfcpumodel.Context) (map[string][]fieldWidget, error) {
	result := map[string][]fieldWidget{}

	catalog, err := ctx.Catalog()
	if err != nil {
		return nil, errors.Wrap(err, "fail contract read catalog")
	}
	acroFormObj, ok := catalog["AcroForm"]
	if !ok {
		return result, nil
	}
	acroForm, err := ctx.DereferenceDict(acroFormObj)
	if err != nil {
		return nil, errors.Wrap(err, "fail contract dereference acroform")
	}
	fieldsObj, ok := acroForm["Fields"]
	if !ok {
		return result, nil
	}
	fieldsArr, err := ctx.DereferenceArray(fieldsObj)
	if err != nil {
		return nil, errors.Wrap(err, "fail contract dereference acroform fields")
	}

	// 各ウィジェット注釈がどのページに置かれているかを、ページのAnnots配列
	// から事前に引けるようにしておく(フィールド辞書自体にはページ番号を
	// 直接持たないため)。
	pageOfAnnotObjNr := map[int]int{}
	for pageNr := 1; pageNr <= ctx.PageCount; pageNr++ {
		pageDict, _, _, err := ctx.PageDict(pageNr, false)
		if err != nil {
			continue
		}
		annotsObj, ok := pageDict["Annots"]
		if !ok {
			continue
		}
		annotsArr, err := ctx.DereferenceArray(annotsObj)
		if err != nil {
			continue
		}
		for _, a := range annotsArr {
			if ir, ok := a.(pdfcputypes.IndirectRef); ok {
				pageOfAnnotObjNr[ir.ObjectNumber.Value()] = pageNr
			}
		}
	}

	var walk func(fieldObj pdfcputypes.Object, fieldDict pdfcputypes.Dict, parentName string, inheritedQ int) error
	walk = func(fieldObj pdfcputypes.Object, fieldDict pdfcputypes.Dict, parentName string, inheritedQ int) error {
		name := parentName
		if t, ok := fieldDict["T"]; ok {
			if s, ok := t.(pdfcputypes.StringLiteral); ok {
				if parentName != "" {
					name = parentName + "." + string(s)
				} else {
					name = string(s)
				}
			}
		}

		quadding := inheritedQ
		if q, ok := fieldDict["Q"]; ok {
			if iv, ok := q.(pdfcputypes.Integer); ok {
				quadding = int(iv)
			}
		}

		_, hasRect := fieldDict["Rect"]
		if kidsObj, ok := fieldDict["Kids"]; ok && !hasRect {
			// Kidsは子フィールドまたはウィジェット注釈の配列。このフィールド
			// 自体はRectを持たない(=末端のウィジェットではない)ので辿る。
			kidsArr, err := ctx.DereferenceArray(kidsObj)
			if err != nil {
				return errors.Wrap(err, "fail contract dereference kids")
			}
			for _, k := range kidsArr {
				kidDict, err := ctx.DereferenceDict(k)
				if err != nil {
					continue
				}
				if err := walk(k, kidDict, name, quadding); err != nil {
					return err
				}
			}
			return nil
		}

		if !hasRect {
			return nil
		}
		rectArr, err := ctx.DereferenceArray(fieldDict["Rect"])
		if err != nil {
			return errors.Wrap(err, "fail contract dereference rect")
		}
		rect, err := ctx.RectForArray(rectArr)
		if err != nil {
			return errors.Wrap(err, "fail contract rect for array")
		}

		page := 0
		if ir, ok := fieldObj.(pdfcputypes.IndirectRef); ok {
			page = pageOfAnnotObjNr[ir.ObjectNumber.Value()]
		}

		result[name] = append(result[name], fieldWidget{
			page:     page,
			rect:     rect,
			quadding: quadding,
		})
		return nil
	}

	for _, f := range fieldsArr {
		fieldDict, err := ctx.DereferenceDict(f)
		if err != nil {
			continue
		}
		if err := walk(f, fieldDict, "", 0); err != nil {
			return nil, err
		}
	}

	return result, nil
}
