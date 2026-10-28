package jobs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func attendanceReminderTestDate(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// PMから共有された実測値(2026年のカレンダー)をそのままケースにしている。
func TestBuildReminderMessage(t *testing.T) {
	septemberSilverWeek := map[string]bool{
		"2026-09-21": true, // 敬老の日
		"2026-09-22": true, // 国民の休日
		"2026-09-23": true, // 秋分の日
	}

	tests := []struct {
		name     string
		today    time.Time
		holidays map[string]bool
		wantText string
		wantPost bool
		wantErr  bool
	}{
		{
			name:     "木曜、翌日が営業日ならそのまま案内する",
			today:    attendanceReminderTestDate(2026, 9, 17),
			holidays: septemberSilverWeek,
			wantText: "09月18日(金)の勤怠連絡をお願いします。",
			wantPost: true,
		},
		{
			name:     "金曜、シルバーウィークを飛ばして翌営業日を案内する",
			today:    attendanceReminderTestDate(2026, 9, 18),
			holidays: septemberSilverWeek,
			wantText: "09月24日(木)の勤怠連絡をお願いします。",
			wantPost: true,
		},
		{
			name:     "土曜は営業日ではないので投稿しない",
			today:    attendanceReminderTestDate(2026, 9, 19),
			holidays: septemberSilverWeek,
			wantPost: false,
		},
		{
			name:     "祝日が休みなくmaxLookaheadDaysを超えて続く場合はエラー",
			today:    attendanceReminderTestDate(2026, 1, 1),
			holidays: allDaysHolidayForTest(attendanceReminderTestDate(2026, 1, 2), 20),
			wantPost: false,
			wantErr:  true,
		},
		{
			// ②レビューで見つかった穴: 祝日一覧の取得に失敗するとattendanceReminderJob
			// はholidaysを空集合にフォールバックする(fail-open)。この場合、本来
			// 祝日である日にも投稿してしまうことを、意図した挙動として固定する。
			// 2026-09-23は実際には秋分の日だが、holidaysが空なのでcase「平日」として
			// 扱われる。
			name:     "祝日取得失敗でholidaysが空の場合、本来祝日の日にも投稿してしまう(fail-open)",
			today:    attendanceReminderTestDate(2026, 9, 22),
			holidays: map[string]bool{},
			wantText: "09月23日(水)の勤怠連絡をお願いします。",
			wantPost: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, post, err := buildReminderMessage(tt.today, tt.holidays, attendanceReminderMaxLookaheadDays)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPost, post)
			if tt.wantPost {
				assert.Equal(t, tt.wantText, text)
			}
		})
	}
}

// buildReminderMessageと同じ2026年のカレンダーで、当日朝版(翌営業日を探さず
// 今日だけを見る)の4ケースを確認する。
func TestBuildMorningReminderMessage(t *testing.T) {
	septemberSilverWeek := map[string]bool{
		"2026-09-21": true, // 敬老の日
		"2026-09-22": true, // 国民の休日
		"2026-09-23": true, // 秋分の日
	}

	tests := []struct {
		name     string
		today    time.Time
		holidays map[string]bool
		wantText string
		wantPost bool
	}{
		{
			name:     "平日なら本日の勤怠連絡を案内する",
			today:    attendanceReminderTestDate(2026, 9, 17),
			holidays: septemberSilverWeek,
			wantText: "本日の勤怠連絡をお願いします。",
			wantPost: true,
		},
		{
			name:     "土曜は営業日ではないので投稿しない",
			today:    attendanceReminderTestDate(2026, 9, 19),
			holidays: septemberSilverWeek,
			wantPost: false,
		},
		{
			name:     "祝日は営業日ではないので投稿しない",
			today:    attendanceReminderTestDate(2026, 9, 23),
			holidays: septemberSilverWeek,
			wantPost: false,
		},
		{
			// buildReminderMessageと同じfail-open: 祝日一覧の取得に失敗すると
			// holidaysが空集合になり、本来祝日の日にも投稿してしまう。
			name:     "祝日取得失敗でholidaysが空の場合、本来祝日の日にも投稿してしまう(fail-open)",
			today:    attendanceReminderTestDate(2026, 9, 23),
			holidays: map[string]bool{},
			wantText: "本日の勤怠連絡をお願いします。",
			wantPost: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, post := buildMorningReminderMessage(tt.today, tt.holidays)
			assert.Equal(t, tt.wantPost, post)
			if tt.wantPost {
				assert.Equal(t, tt.wantText, text)
			}
		})
	}
}

func allDaysHolidayForTest(from time.Time, days int) map[string]bool {
	h := map[string]bool{}
	for i := 0; i < days; i++ {
		h[from.AddDate(0, 0, i).Format("2006-01-02")] = true
	}
	return h
}
