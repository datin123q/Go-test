package jsonutils

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_JSONTime_EqualDate(t *testing.T) {
	t1 := time.Date(2024, 3, 16, 15, 0, 0, 0, jst)
	t2 := time.Date(2024, 3, 16, 23, 59, 59, 0, jst)
	t3 := time.Date(2024, 3, 17, 0, 0, 0, 0, jst)

	// 同じ日付で比較
	t.Run("2024/03/16 15:00:00 - 2024-03-16 23:59:59", func(t *testing.T) {
		jsonTime := JSONTime(t1)
		assert.Truef(t, jsonTime.EqualDate(t2), "Expected: %v, Actual: %v", t1.Format(time.DateOnly), t2.Format(time.DateOnly))
	})

	// 異なる日付で比較
	t.Run("2024/03/16 15:00:00 - 2024-03-17 00:00:00", func(t *testing.T) {
		jsonTime := JSONTime(t1)
		assert.Falsef(t, jsonTime.EqualDate(t3), "Expected: %v, Actual: %v", t1.Format(time.DateOnly), t3.Format(time.DateOnly))
	})

}

func TestJSONTime_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		value    JSONTime
		wantZero bool
	}{
		{
			name:     "ゼロ値の場合 true を返す",
			value:    JSONTime{},
			wantZero: true,
		},
		{
			name:     "非ゼロ値（UTC）の場合 false を返す",
			value:    JSONTime(time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)),
			wantZero: false,
		},
		{
			name:     "非ゼロ値（JST）の場合 false を返す",
			value:    JSONTime(time.Date(2025, 4, 1, 0, 0, 0, 0, jst)),
			wantZero: false,
		},
		{
			name:     "ParseJSONTimeJST 経由の値の場合 false を返す",
			value:    func() JSONTime { v, _ := ParseJSONTimeJST(time.DateTime, "2025-04-01 00:00:00"); return v }(),
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantZero, tt.value.IsZero())
		})
	}
}

// MIR-1384: 基準を正午にする理由はParseJSONDateJSTのgodocを参照。
// ここではUTC・JSTのどちらに変換しても年月日が変わらないことを確認する。
func Test_ParseJSONDateJST(t *testing.T) {
	t.Run("正午JSTとして解釈するため、UTCへ変換しても日付が変わらない", func(t *testing.T) {
		got, err := ParseJSONDateJST(time.DateOnly, "2026-07-25")
		require.NoError(t, err)

		assert.Equal(t, "2026-07-25", time.Time(got).In(time.UTC).Format(time.DateOnly))
		assert.Equal(t, "2026-07-25", time.Time(got).In(jst).Format(time.DateOnly))
	})

	t.Run("前後の空白を無視する", func(t *testing.T) {
		got, err := ParseJSONDateJST(time.DateOnly, "  2026-07-25  ")
		require.NoError(t, err)
		assert.Equal(t, "2026-07-25", time.Time(got).In(time.UTC).Format(time.DateOnly))
	})

	t.Run("空文字はエラーを返す", func(t *testing.T) {
		_, err := ParseJSONDateJST(time.DateOnly, "")
		assert.Error(t, err)
	})

	t.Run("不正な形式はエラーを返す", func(t *testing.T) {
		_, err := ParseJSONDateJST(time.DateOnly, "2026/07/25")
		assert.Error(t, err)
	})
}

func TestUnmarshalJSON(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	type testCase struct {
		name     string
		input    string
		expected time.Time
		hasError bool
	}

	testCases := []testCase{
		// time.DateTime形式テスト
		{
			name:     "time.DateTime形式でタイムゾーン情報がない場合、JSTとして認識する",
			input:    `"2024-10-16 12:10:19"`,
			expected: time.Date(2024, 10, 16, 12, 10, 19, 0, jst),
			hasError: false,
		},
		// RFC3339形式テスト
		{
			name:     "time.RFC3339形式でタイムゾーン情報がある場合、ローカル時間に変換する",
			input:    `"2024-10-16T03:10:19Z"`,
			expected: time.Date(2024, 10, 16, 12, 10, 19, 0, jst), // JST로 변환
			hasError: false,
		},
		// 不正な形式
		{
			name:     "不正な形式は処理しない。",
			input:    `"invalid time format"`,
			expected: time.Time{},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var jt JSONTime
			err := json.Unmarshal([]byte(tc.input), &jt)

			if tc.hasError {
				assert.Error(t, err, "expected error but got none for input %q", tc.input)
			} else {
				assert.NoError(t, err, "unexpected error for input %q: %v", tc.input, err)
				assert.True(t, jt.EqualDate(tc.expected), "for input %q expected %v but got %v", tc.input, tc.expected, jt)
			}
		})
	}
}
