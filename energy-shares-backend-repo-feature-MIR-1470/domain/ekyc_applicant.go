package domain

import "strings"

// BuildEkycAddress は都道府県・市区町村・住所・建物名(任意)を1つの住所文字列に
// 組み立てる(MIR-1121)。フロントエンドが以前ローカル状態とuser情報を
// 全部空かの二択で切り替えていた組み立てロジックと同じ並び順・区切り方だが、
// 単一のDB上の値だけから組み立てるため、一部だけ欠けた住所が混在することはない。
// setOtherProfileがProvince/Cityをtrimしないため、ここで各項目をtrimしてから
// 判定・連結する(前後の空白がそのまま住所へ混入し、本人確認書類との照合で
// 弾かれることを防ぐ)。
//
// ホ方式(user-apiが画面へ返す住所)とカ方式(jobs-schedulerがBPOへ申請する住所、MIR-1470)で
// 同じ住所を組み立てるため、ここに1つだけ置く。
func BuildEkycAddress(province, city, address string, address2 *string) string {
	parts := make([]string, 0, 4)
	for _, part := range []string{province, city, address} {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	if address2 != nil {
		if trimmed := strings.TrimSpace(*address2); trimmed != "" {
			parts = append(parts, " "+trimmed)
		}
	}
	return strings.Join(parts, "")
}

// BuildEkycName は、姓と名をスペースなしで連結した氏名を返す(本人確認書類・JPKIの
// 氏名との照合用)。フロントエンド(energy-share-next の buildEkycAccount)がホ方式で
// Nexwayへ渡している氏名と同じ組み立て方で、空の項目は飛ばす。
func BuildEkycName(familyName, firstName string) string {
	parts := make([]string, 0, 2)
	for _, part := range []string{familyName, firstName} {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, "")
}
