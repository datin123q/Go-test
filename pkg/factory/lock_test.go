package factory

import (
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func Test_IsLockNotAvailableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "MySQLエラー3572(ロック取得できず)",
			err:  &mysql.MySQLError{Number: 3572, Message: "Statement aborted because lock(s) could not be acquired immediately"},
			want: true,
		},
		{
			name: "ラップされたMySQLエラー3572",
			err:  fmt.Errorf("query failed: %w", &mysql.MySQLError{Number: 3572, Message: "..."}),
			want: true,
		},
		{
			name: "別のMySQLエラー番号",
			err:  &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"},
			want: false,
		},
		{
			name: "MySQLエラーでない",
			err:  fmt.Errorf("some other error"),
			want: false,
		},
		{
			name: "nil",
			err:  nil,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsLockNotAvailableError(tt.err))
		})
	}
}

func Test_IsDeadlockError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "MySQLエラー1213(デッドロック)",
			err:  &mysql.MySQLError{Number: 1213, Message: "Deadlock found when trying to get lock; try restarting transaction"},
			want: true,
		},
		{
			name: "ラップされたMySQLエラー1213",
			err:  fmt.Errorf("query failed: %w", &mysql.MySQLError{Number: 1213, Message: "..."}),
			want: true,
		},
		{
			name: "別のMySQLエラー番号",
			err:  &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"},
			want: false,
		},
		{
			name: "MySQLエラーでない",
			err:  fmt.Errorf("some other error"),
			want: false,
		},
		{
			name: "nil",
			err:  nil,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsDeadlockError(tt.err))
		})
	}
}
