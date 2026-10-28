package jobs

import (
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// isProjectStatusChangedは、ChangeStatusのエラーを警告(想定内)にするか
// エラー(想定外)にするかの分岐に使う(MIR-1314)。判定を誤ると、DBエラー等の
// 想定外の失敗までWarnで握りつぶしてしまう。
func TestIsProjectStatusChanged(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrorProjectStatusChangedそのものならtrue",
			err:  factory.ErrorProjectStatusChanged.New(errors.New("project status already changed")),
			want: true,
		},
		{
			name: "pkg/errors.Wrapで包まれていてもたどれる",
			err:  errors.Wrap(factory.ErrorProjectStatusChanged.New(errors.New("project status already changed")), "wrapped"),
			want: true,
		},
		{
			name: "別のfactory.Errorならfalse",
			err:  factory.ErrorDB.New(errors.New("db error")),
			want: false,
		},
		{
			name: "factory.Error以外の通常のerrorならfalse",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "nilならfalse",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isProjectStatusChanged(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
