package businessday

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const holidayCSVURL = "https://www8.cao.go.jp/chosei/shukujitsu/syukujitsu.csv"

// holidayHTTPClientはタイムアウトを明示する。http.DefaultClientはタイムアウトが
// 無く、相手が応答しないまま接続を保持し続けるとcronジョブのgoroutineが無期限に
// 張り付いてしまう(MIR-1194と同じ理由)。元のシェルスクリプト(curl --max-time 20)
// に合わせている。
var holidayHTTPClient = &http.Client{Timeout: 20 * time.Second}

var weekdayLabelsJP = [...]string{"日", "月", "火", "水", "木", "金", "土"}

// IsBusinessDay は土日・holidaysに含まれる日を除いた営業日かどうかを返す。
// holidaysのキーはdateKeyの形式("2006-01-02")で揃える必要がある。
func IsBusinessDay(date time.Time, holidays map[string]bool) bool {
	switch date.Weekday() {
	case time.Saturday, time.Sunday:
		return false
	}
	return !holidays[dateKey(date)]
}

// NextBusinessDay はfromの翌日からmaxDays日先まで走査し、最初に見つかった営業日を返す。
// 見つからなければokがfalseになる。
func NextBusinessDay(from time.Time, holidays map[string]bool, maxDays int) (next time.Time, ok bool) {
	for i := 1; i <= maxDays; i++ {
		candidate := from.AddDate(0, 0, i)
		if IsBusinessDay(candidate, holidays) {
			return candidate, true
		}
	}
	return time.Time{}, false
}

// ReminderLabel は"09月18日(金)"の形式で日付を表示する。
func ReminderLabel(date time.Time) string {
	return fmt.Sprintf("%02d月%02d日(%s)", date.Month(), date.Day(), weekdayLabelsJP[date.Weekday()])
}

func dateKey(date time.Time) string {
	return date.Format("2006-01-02")
}

// FetchHolidays は内閣府の祝日CSV(Shift-JIS、ヘッダー行あり)を取得し、
// 日付キー("2006-01-02")の集合として返す。取得・パースに失敗した場合は
// エラーを返すので、呼び出し側で「土日のみで判定する」フォールバックを行うこと。
func FetchHolidays(ctx context.Context) (map[string]bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, holidayCSVURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build holiday csv request")
	}

	resp, err := holidayHTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch holiday csv")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status fetching holiday csv: %d", resp.StatusCode)
	}

	return parseHolidayCSV(resp.Body)
}

func parseHolidayCSV(body io.Reader) (map[string]bool, error) {
	decoded := transform.NewReader(body, japanese.ShiftJIS.NewDecoder())
	scanner := bufio.NewScanner(decoded)

	holidays := map[string]bool{}
	first := true
	for scanner.Scan() {
		if first {
			first = false
			continue // ヘッダー行(「国民の祝日・休日月日,国民の祝日・休日名称」)を読み飛ばす
		}
		line := strings.TrimRight(scanner.Text(), "\r")
		if line == "" {
			continue
		}
		dateStr, _, found := strings.Cut(line, ",")
		if !found {
			continue
		}
		t, err := time.ParseInLocation("2006/1/2", strings.TrimSpace(dateStr), time.UTC)
		if err != nil {
			continue // 年表示などの余分な行はスキップする
		}
		holidays[dateKey(t)] = true
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read holiday csv")
	}

	// 内閣府が配信するCSVには常に十数件以上の祝日が載っている。0件は、
	// メンテナンス画面がステータス200で返ってきた・CSVの書式が変わって
	// 1行もパースできなかった、といった異常を示す。呼び出し側(jobs.
	// attendanceReminderJob)はこのエラーをログに残したうえで空集合へ
	// フォールバックする(=祝日を知らない状態で判定を続ける、fail-open)。
	// ここでエラーにする目的は誤動作を防ぐことではなく、異常が起きたことを
	// ログに残し、気付けるようにすること。
	if len(holidays) == 0 {
		return nil, errors.New("holiday csv parsed to zero entries")
	}

	return holidays, nil
}
