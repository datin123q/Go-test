package jobs

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// TestIsJobExecutionLockedError は、DepositAutoProcessJobが「他の実行中と
// ぶつかったので弾いた」ケースだけを見分けられることを確認する(MIR-1194)。
// ここを誤ると、DB接続断など本来失敗として扱うべきエラーまでスキップ扱いに
// なり、実行されていないのに問題無しとして握りつぶしてしまう。
//
// tryLockJobExecution自体(FOR UPDATE NOWAITを使うMySQL専用SQL)はsqliteでは
// 構文自体が通らないため、この単体テストの対象にはできない(repository/project/mysql
// の TryLockById と同じ制約)。ロック競合の実地確認はQA/staging環境での手動確認に
// 委ねる。
func TestIsJobExecutionLockedError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrJobExecutionLockedそのもの",
			err:  factory.ErrJobExecutionLocked.New(errors.New("locked")),
			want: true,
		},
		{
			name: "別のfactory.Error(DBエラー)",
			err:  factory.ErrorDB.New(errors.New("db down")),
			want: false,
		},
		{
			name: "factory.Errorではない一般的なエラー",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "nilエラー",
			err:  nil,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isJobExecutionLockedError(tt.err))
		})
	}
}
