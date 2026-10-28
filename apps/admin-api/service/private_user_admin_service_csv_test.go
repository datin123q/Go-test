package service

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/illuminarean/energy-shares-backend-repo/service/dto"
)

// MIR-1107: 個人会員CSVの固定列に「メールアドレス(ID)」を投資家名の直後へ出す。
// 基本情報側のメール列は外したので、基本情報をチェックしても重複しないこと。
func TestPrivateUserAdminService_PrepareAndReturnCsv_EmailColumn(t *testing.T) {
	require.NoError(t, i18nutils.InitI18nFiles(pathutils.GetPathOfAssets()))

	items := []*dto.PrivateUserCsv{{Id: 1, Name: "山田　太郎", Email: "taro@example.com"}}

	t.Run("基本情報なし", func(t *testing.T) {
		records := (&PrivateUserAdminService{}).PrepareAndReturnCsv(items, false, false, false, false, false, false)

		require.Len(t, records, 2)
		require.Equal(t, []string{"投資家名", "メールアドレス(ID)", "生年月日"}, records[0])
		require.Equal(t, "taro@example.com", records[1][1])
	})

	t.Run("基本情報あり", func(t *testing.T) {
		records := (&PrivateUserAdminService{}).PrepareAndReturnCsv(items, false, true, false, false, false, false)

		require.Equal(t, "メールアドレス(ID)", records[0][1])
		require.Equal(t, 1, lo.Count(records[0], "メールアドレス(ID)"))
		require.NotContains(t, records[0], "メールアドレス")
		require.Len(t, records[1], len(records[0]))
	})
}
