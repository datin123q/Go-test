package mysql

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type userInfoHistoryDBFactoryMock struct {
	engine *xorm.Engine
}

func (d userInfoHistoryDBFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	return d.engine, nil
}

func (d userInfoHistoryDBFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}

// 本番スキーマ（db/sql/version/energyshares.base.schema.sql）の列に合わせたテーブル定義。
const (
	createUserInfoHistoryTable = `
		CREATE TABLE user_info_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NULL,
			process_user_id INTEGER NULL,
			substance_type TEXT NULL,
			process_method TEXT NULL,
			name TEXT NULL,
			email TEXT NULL,
			created DATETIME NULL
		);`
	createUserChangedContentTable = `
		CREATE TABLE user_changed_content (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			info_id INTEGER NULL,
			content_type TEXT NULL,
			prev_content TEXT NULL,
			curr_content TEXT NULL,
			prev_content_html TEXT NULL,
			curr_content_html TEXT NULL,
			prev_url TEXT NULL,
			curr_url TEXT NULL
		);`
)

// setupUserInfoHistoryRepo は ddls で指定したテーブルだけを作成した sqlite でリポジトリを組み立てます。
// テーブルを作らないことで INSERT の失敗を再現します。
func setupUserInfoHistoryRepo(t *testing.T, ddls ...string) (userInfoHistoryMysqlRepository, *xorm.Engine) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = engine.Close() })

	for _, ddl := range ddls {
		_, err = engine.Exec(ddl)
		require.NoError(t, err)
	}

	return userInfoHistoryMysqlRepository{dbFactory: userInfoHistoryDBFactoryMock{engine: engine}}, engine
}

// tableRowCount はテーブルの件数を返します。テーブルが無い場合は 0 を返します。
func tableRowCount(t *testing.T, engine *xorm.Engine, table string) int64 {
	t.Helper()

	exist, err := engine.IsTableExist(table)
	require.NoError(t, err)
	if !exist {
		return 0
	}
	count, err := engine.Table(table).Count()
	require.NoError(t, err)
	return count
}

func Test_userInfoHistoryMysqlRepository_RecordContentChanges(t *testing.T) {
	twoChanges := cdomain.ContentChangeEntry{
		UserId:        1001,
		ActorId:       1001,
		ActorName:     "山田太郎",
		ActorEmail:    "new@example.com",
		ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
		Changes: []cdomain.ContentChange{
			{ContentType: domain.USER_CHANGED_CONTENT_EMAIL, Prev: "old@example.com", Curr: "new@example.com"},
			{ContentType: domain.USER_CHANGED_CONTENT_PASSWORD, Prev: domain.USER_HISTORY_MASKED_VALUE, Curr: domain.USER_HISTORY_MASKED_VALUE},
		},
	}

	tests := []struct {
		name             string
		ddls             []string
		entry            cdomain.ContentChangeEntry
		wantErr          string
		wantHistoryCount int64
		wantContentCount int64
	}{
		{
			name:             "変更項目が2件の場合は見出し1件と明細2件を登録する",
			ddls:             []string{createUserInfoHistoryTable, createUserChangedContentTable},
			entry:            twoChanges,
			wantHistoryCount: 1,
			wantContentCount: 2,
		},
		{
			name:             "変更項目が空の場合は何も登録しない",
			ddls:             []string{createUserInfoHistoryTable, createUserChangedContentTable},
			entry:            cdomain.ContentChangeEntry{UserId: 1001},
			wantHistoryCount: 0,
			wantContentCount: 0,
		},
		{
			name:             "見出しの INSERT に失敗した場合はエラーを返し、明細は登録しない",
			ddls:             []string{createUserChangedContentTable},
			entry:            twoChanges,
			wantErr:          "user_info_history の INSERT に失敗しました",
			wantHistoryCount: 0,
			wantContentCount: 0,
		},
		{
			// 既存の RecordAuditStatusChange と同じくトランザクションを張らないため、見出しは残る
			name:             "明細の INSERT に失敗した場合はエラーを返し、見出しは残る",
			ddls:             []string{createUserInfoHistoryTable},
			entry:            twoChanges,
			wantErr:          "user_changed_content の INSERT に失敗しました",
			wantHistoryCount: 1,
			wantContentCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, engine := setupUserInfoHistoryRepo(t, tt.ddls...)

			err := repo.RecordContentChanges(context.Background(), tt.entry)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantHistoryCount, tableRowCount(t, engine, "user_info_history"))
			assert.Equal(t, tt.wantContentCount, tableRowCount(t, engine, "user_changed_content"))
		})
	}
}

func Test_userInfoHistoryMysqlRepository_RecordContentChanges_Rows(t *testing.T) {
	// 登録した見出しと明細の各列、および明細が見出しの id に紐づくことを確認する
	repo, engine := setupUserInfoHistoryRepo(t, createUserInfoHistoryTable, createUserChangedContentTable)

	require.NoError(t, repo.RecordContentChanges(context.Background(), cdomain.ContentChangeEntry{
		UserId:        1001,
		ActorId:       9,
		ActorName:     "管理者",
		ActorEmail:    "admin@example.com",
		ProcessMethod: domain.USER_HISTORY_PROCESS_OPERATOR,
		Changes: []cdomain.ContentChange{
			{ContentType: domain.USER_CHANGED_CONTENT_EMAIL, Prev: "old@example.com", Curr: "new@example.com"},
			{ContentType: domain.USER_CHANGED_CONTENT_PASSWORD, Prev: domain.USER_HISTORY_MASKED_VALUE, Curr: domain.USER_HISTORY_MASKED_VALUE},
		},
	}))

	var histories []domain.UserInfoHistory
	require.NoError(t, engine.Table("user_info_history").Find(&histories))
	require.Len(t, histories, 1)
	assert.Equal(t, int64(1001), histories[0].UserId)
	assert.Equal(t, int64(9), histories[0].ProcessUserId)
	assert.Equal(t, string(domain.USER_HISTORY_CONTENTS), histories[0].SubstanceType)
	assert.Equal(t, domain.USER_HISTORY_PROCESS_OPERATOR, histories[0].ProcessMethod)
	assert.Equal(t, "管理者", histories[0].Name)
	assert.Equal(t, "admin@example.com", histories[0].Email)

	var contents []domain.UserChangedContent
	require.NoError(t, engine.Table("user_changed_content").Asc("id").Find(&contents))
	require.Len(t, contents, 2)
	assert.Equal(t, histories[0].Id, contents[0].InfoId)
	assert.Equal(t, string(domain.USER_CHANGED_CONTENT_EMAIL), contents[0].ContentType)
	assert.Equal(t, "old@example.com", contents[0].PrevContent)
	assert.Equal(t, "new@example.com", contents[0].CurrContent)
	assert.Equal(t, histories[0].Id, contents[1].InfoId)
	assert.Equal(t, string(domain.USER_CHANGED_CONTENT_PASSWORD), contents[1].ContentType)
	assert.Equal(t, "********", contents[1].PrevContent)
	assert.Equal(t, "********", contents[1].CurrContent)
}
