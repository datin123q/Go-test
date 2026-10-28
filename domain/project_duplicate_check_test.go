package domain

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// 重複の判定は INSERT より前に呼ぶ。そのとき id はまだ 0 で、
// id != 0 の条件により全件が比較対象になることを固定する。
// ここが崩れると、新規登録の重複を拾えなくなる。
func setupProjectDuplicateCheck(t *testing.T) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = engine.Exec(`
		CREATE TABLE project (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fund_number INTEGER NULL,
			closed_fund_slug TEXT NULL,
			deleted DATETIME NULL
		);
	`)
	require.NoError(t, err)

	_, err = engine.Exec(`
		INSERT INTO project (id, fund_number, closed_fund_slug, deleted) VALUES
			(1, 202600001, 'alpha', NULL),
			(2, 202600002, 'beta',  NULL),
			(3, 202600003, 'gamma', '2026-08-01 00:00:00');
	`)
	require.NoError(t, err)

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}

func TestProject_CheckExistFundNumber(t *testing.T) {
	ctx := setupProjectDuplicateCheck(t)

	tests := []struct {
		name       string
		id         int64
		fundNumber *int64
		want       bool
	}{
		{
			name:       "新規登録で重複していれば拾う",
			id:         0,
			fundNumber: lo.ToPtr(int64(202600001)),
			want:       true,
		},
		{
			name:       "新規登録で重複していなければ通す",
			id:         0,
			fundNumber: lo.ToPtr(int64(202600999)),
			want:       false,
		},
		{
			name:       "更新時は自分自身を重複とみなさない",
			id:         1,
			fundNumber: lo.ToPtr(int64(202600001)),
			want:       false,
		},
		{
			name:       "更新時に他のファンドと重複すれば拾う",
			id:         1,
			fundNumber: lo.ToPtr(int64(202600002)),
			want:       true,
		},
		// DB のユニーク制約は deleted を見ない。ここで除外すると
		// 画面では通るのに保存で落ちる、という食い違いが起きる。
		{
			name:       "論理削除済みも重複とみなす",
			id:         0,
			fundNumber: lo.ToPtr(int64(202600003)),
			want:       true,
		},
		{
			name:       "未入力は重複とみなさない",
			id:         0,
			fundNumber: nil,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Project{FundNumber: tt.fundNumber}
			p.Id = tt.id

			got, err := p.CheckExistFundNumber(ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProject_CheckExistClosedSlug(t *testing.T) {
	ctx := setupProjectDuplicateCheck(t)

	tests := []struct {
		name string
		id   int64
		slug string
		want bool
	}{
		{name: "新規登録で重複していれば拾う", id: 0, slug: "alpha", want: true},
		{name: "新規登録で重複していなければ通す", id: 0, slug: "delta", want: false},
		{name: "更新時は自分自身を重複とみなさない", id: 1, slug: "alpha", want: false},
		{name: "未入力は重複とみなさない", id: 0, slug: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Project{ClosedFundSlug: tt.slug}
			p.Id = tt.id

			got, err := p.CheckExistClosedSlug(ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
