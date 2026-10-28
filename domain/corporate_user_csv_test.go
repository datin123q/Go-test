package domain

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
)

// MIR-1107: 法人会員CSVの固定列に「メールアドレス(ID)」を代表取締役氏名の直後へ出す。
// 基本情報側のメール列は外したので、基本情報をチェックしても重複しないこと。
func TestCorporateUser_PrepareAndReturnCsv_EmailColumn(t *testing.T) {
	require.NoError(t, i18nutils.InitI18nFiles(pathutils.GetPathOfAssets()))

	items := []CorporateUser{{
		User:             User{Id: 1, Name: "株式会社テスト", Email: "corp@example.com", ZipCode: "1000001"},
		CorporateProfile: CorporateProfile{RepresentativeFamilyName: "山田", RepresentativeFirstName: "太郎"},
	}}

	t.Run("基本情報なし", func(t *testing.T) {
		records := CorporateUser{}.PrepareAndReturnCsv(items, false, false, false, false, false, false)

		require.Len(t, records, 2)
		require.Equal(t, []string{"企業名", "代表取締役氏名", "メールアドレス(ID)", "郵便番号", "企業所在地"}, records[0])
		require.Equal(t, "corp@example.com", records[1][2])
	})

	t.Run("基本情報あり", func(t *testing.T) {
		records := CorporateUser{}.PrepareAndReturnCsv(items, false, true, false, false, false, false)

		require.Equal(t, "メールアドレス(ID)", records[0][2])
		require.Equal(t, 1, lo.Count(records[0], "メールアドレス(ID)"))
		require.NotContains(t, records[0], "メールアドレス")
		require.Len(t, records[1], len(records[0]))
	})
}
