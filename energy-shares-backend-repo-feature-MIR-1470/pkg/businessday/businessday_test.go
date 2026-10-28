package businessday

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestIsBusinessDay(t *testing.T) {
	holidays := map[string]bool{
		"2027-01-01": true,
	}

	tests := []struct {
		name string
		date time.Time
		want bool
	}{
		{"平日は営業日", date(2026, 9, 17), true},
		{"土曜は営業日ではない", date(2026, 9, 19), false},
		{"日曜は営業日ではない", date(2026, 9, 20), false},
		{"祝日(平日)は営業日ではない", date(2027, 1, 1), false},
		{"祝日でない平日は営業日", date(2027, 1, 4), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsBusinessDay(tt.date, holidays))
		})
	}
}

// PMから共有された実測値(2026年のカレンダー)。土日・敬老の日/国民の休日/秋分の日
// (9/21-23)・元日(1/1)・憲法記念日からの振替休日を含むゴールデンウィーク(5/3-6)を
// またぐケースを含む。
func TestNextBusinessDay(t *testing.T) {
	septemberSilverWeek := map[string]bool{
		"2026-09-21": true, // 敬老の日
		"2026-09-22": true, // 国民の休日
		"2026-09-23": true, // 秋分の日
	}
	newYear := map[string]bool{
		"2027-01-01": true, // 元日
	}
	goldenWeek := map[string]bool{
		"2026-05-03": true, // 憲法記念日
		"2026-05-04": true, // みどりの日
		"2026-05-05": true, // こどもの日
		"2026-05-06": true, // 振替休日
	}

	tests := []struct {
		name     string
		today    time.Time
		holidays map[string]bool
		wantNext time.Time
		wantOk   bool
	}{
		{
			name:     "翌日が営業日ならそのまま翌日",
			today:    date(2026, 9, 17),
			holidays: septemberSilverWeek,
			wantNext: date(2026, 9, 18),
			wantOk:   true,
		},
		{
			name:     "土日+敬老の日+国民の休日+秋分の日をまとめて飛ばす",
			today:    date(2026, 9, 18),
			holidays: septemberSilverWeek,
			wantNext: date(2026, 9, 24),
			wantOk:   true,
		},
		{
			name:     "年末年始(元日+土日)を飛ばす",
			today:    date(2026, 12, 31),
			holidays: newYear,
			wantNext: date(2027, 1, 4),
			wantOk:   true,
		},
		{
			name:     "ゴールデンウィーク(土日+3連休+振替休日)を飛ばす",
			today:    date(2026, 5, 1),
			holidays: goldenWeek,
			wantNext: date(2026, 5, 7),
			wantOk:   true,
		},
		{
			name:     "holidaysがすべてmaxDaysを超えて続く場合は見つからない",
			today:    date(2026, 1, 1),
			holidays: allDaysHoliday(date(2026, 1, 2), 20),
			wantNext: time.Time{},
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next, ok := NextBusinessDay(tt.today, tt.holidays, 14)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.True(t, tt.wantNext.Equal(next), "want %v, got %v", tt.wantNext, next)
			}
		})
	}
}

func allDaysHoliday(from time.Time, days int) map[string]bool {
	h := map[string]bool{}
	for i := 0; i < days; i++ {
		h[dateKey(from.AddDate(0, 0, i))] = true
	}
	return h
}

func TestReminderLabel(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{"金曜", date(2026, 9, 18), "09月18日(金)"},
		{"木曜", date(2026, 9, 24), "09月24日(木)"},
		{"月曜", date(2027, 1, 4), "01月04日(月)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ReminderLabel(tt.date))
		})
	}
}

// TestParseHolidayCSVはFetchHolidaysが実際に受け取るのと同じ形の入力(Shift-JIS、
// ヘッダー行あり、末尾に西暦の年だけの行が混ざる年もある)を、UTF-8から変換して検証する。
func TestParseHolidayCSV(t *testing.T) {
	utf8CSV := "国民の祝日・休日月日,国民の祝日・休日名称\r\n" +
		"1955/1/1,元日\r\n" +
		"2026/5/3,憲法記念日\r\n" +
		"2026/5/4,みどりの日\r\n" +
		"2026/5/5,こどもの日\r\n" +
		"2026/5/6,休日\r\n"

	sjis, err := io2SJIS(utf8CSV)
	require.NoError(t, err)

	holidays, err := parseHolidayCSV(strings.NewReader(sjis))
	require.NoError(t, err)

	assert.True(t, holidays["1955-01-01"])
	assert.True(t, holidays["2026-05-03"])
	assert.True(t, holidays["2026-05-04"])
	assert.True(t, holidays["2026-05-05"])
	assert.True(t, holidays["2026-05-06"])
	assert.False(t, holidays["2026-05-02"], "CSVに無い日はholidayではない")
	assert.Len(t, holidays, 5, "ヘッダー行はholidayとしてカウントされない")
}

func TestParseHolidayCSV_SkipsUnparsableLines(t *testing.T) {
	utf8CSV := "国民の祝日・休日月日,国民の祝日・休日名称\r\n" +
		"\r\n" + // 空行
		"2026/1/1,元日\r\n"

	sjis, err := io2SJIS(utf8CSV)
	require.NoError(t, err)

	holidays, err := parseHolidayCSV(strings.NewReader(sjis))
	require.NoError(t, err)
	assert.Len(t, holidays, 1)
	assert.True(t, holidays["2026-01-01"])
}

// TestParseHolidayCSV_ZeroEntriesIsAnErrorは①敵対的検証で見つかった穴。
// メンテナンス画面がステータス200で返る・CSVの書式が変わって1行もパースできない
// といった異常時、0件の空集合をそのまま返すとエラーにならず、呼び出し側が
// 「今年は祝日が無い」と誤認して祝日にも投稿してしまっていた。
func TestParseHolidayCSV_ZeroEntriesIsAnError(t *testing.T) {
	tests := []struct {
		name    string
		utf8CSV string
	}{
		{
			name:    "ヘッダー行だけで中身が無い",
			utf8CSV: "国民の祝日・休日月日,国民の祝日・休日名称\r\n",
		},
		{
			name: "書式が変わってどの行も日付として読めない",
			utf8CSV: "国民の祝日・休日月日,国民の祝日・休日名称\r\n" +
				"invalid,元日\r\n" +
				"also-invalid,休日\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sjis, err := io2SJIS(tt.utf8CSV)
			require.NoError(t, err)

			_, err = parseHolidayCSV(strings.NewReader(sjis))
			require.Error(t, err)
		})
	}
}

func io2SJIS(s string) (string, error) {
	var sb strings.Builder
	w := transform.NewWriter(&sb, japanese.ShiftJIS.NewEncoder())
	if _, err := w.Write([]byte(s)); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return sb.String(), nil
}
