package jobs

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/businessday"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/slackutils"
)

// attendanceReminderMaxLookaheadDays は翌営業日を探索する日数。祝日が連続しても
// 現実的にはここまで続かないが、想定外の事態に気付けるよう上限を設けている。
const attendanceReminderMaxLookaheadDays = 14

var jstLocationForAttendanceReminder = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.FixedZone("Asia/Tokyo", 9*60*60)
	}
	return loc
}()

type AttendanceReminderJobs struct{}

// AttendanceReminderJobは、営業日に翌営業日の勤怠連絡を促すメッセージをSlackへ
// 投稿する。GitHub Actionsのスケジュール実行がこの組織では機能しないため、
// energy-shares-next-adminの.github/workflows/attendance-reminder.ymlから
// バックエンドのcronジョブへ移した。
func (AttendanceReminderJobs) AttendanceReminderJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_ATTENDANCE_REMINDER,
		JobName: jobNameAttendanceReminder,
	}, attendanceReminderJob)
}

func attendanceReminderJob(ctx context.Context) error {
	// Webhook URL未設定の環境(cronだけ設定してParameter Storeへの追加待ちの
	// 期間、ローカル等)で毎回失敗としてSentryへ上げ続けないよう、ここで
	// 自分自身を無効化する。cronの設定漏れはscheduleCronJobがログを出すのみで
	// ジョブ自体は登録されるため、この2つは別々に設定miss しうる。
	webhookURL := serviceutils.GetAttendanceReminderSlackWebhookUrl()
	if webhookURL == "" {
		// Warnにしているのは、この経路がジョブ実行履歴上はSUCCESSとして記録され
		// (実行自体は正常終了のため)、履歴の一覧だけでは「投稿した日」と
		// 「設定漏れで何もしなかった日」を区別できないため。ログでだけは
		// 気付けるようにしておく。
		logrus.WithFields(logrus.Fields{
			"desc": "Slack Webhook URLが未設定のため実行しない",
		}).Warn("AttendanceReminderJob")
		return nil
	}

	today := time.Now().In(jstLocationForAttendanceReminder)

	holidays, err := businessday.FetchHolidays(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"desc": "祝日一覧を取得できなかったため、土日のみで営業日を判定する",
			"err":  err,
		}).Warn("AttendanceReminderJob")
		holidays = map[string]bool{}
	}

	text, post, err := buildReminderMessage(today, holidays, attendanceReminderMaxLookaheadDays)
	if err != nil {
		return err
	}
	if !post {
		logrus.WithFields(logrus.Fields{
			"desc":  "本日は営業日ではないため投稿しない",
			"today": today.Format(time.DateOnly),
		}).Info("AttendanceReminderJob")
		return nil
	}

	if err := slackutils.PostMessage(ctx, webhookURL, text); err != nil {
		return errors.Wrap(err, "failed to post attendance reminder to slack")
	}

	logrus.WithFields(logrus.Fields{
		"desc": "勤怠連絡リマインドを投稿した",
		"text": text,
	}).Info("AttendanceReminderJob")
	return nil
}

// buildReminderMessageは、今日の日付と祝日集合から、投稿すべきか・何を投稿する
// かを決める判断部分だけを切り出したもの。Slack投稿やログ出力を含む
// attendanceReminderJob本体からロジックだけを分離し、table-drivenテストで
// 固定できるようにしている。
func buildReminderMessage(today time.Time, holidays map[string]bool, maxLookaheadDays int) (text string, post bool, err error) {
	if !businessday.IsBusinessDay(today, holidays) {
		return "", false, nil
	}

	next, ok := businessday.NextBusinessDay(today, holidays, maxLookaheadDays)
	if !ok {
		return "", false, errors.Errorf("%d日以内に翌営業日が見つからない", maxLookaheadDays)
	}

	return businessday.ReminderLabel(next) + "の勤怠連絡をお願いします。", true, nil
}

// AttendanceMorningReminderJobは、AttendanceReminderJob(前日の投稿で翌営業日を
// 案内する)とは別に、出勤日当日の朝にも同じSlackチャンネルへ投稿する。既存の
// ジョブへ文言を足すのではなく別ジョブにしているのは、cronの時刻設定を
// Parameter Storeにジョブ名単位で追加できるようにするため(JobSchedulerConfig.Cron
// はジョブ名で引く作りになっている)。
func (AttendanceReminderJobs) AttendanceMorningReminderJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_ATTENDANCE_MORNING_REMINDER,
		JobName: jobNameAttendanceMorningReminder,
	}, attendanceMorningReminderJob)
}

func attendanceMorningReminderJob(ctx context.Context) error {
	// attendanceReminderJobと同じく、Webhook URL未設定の環境で毎回失敗として
	// Sentryへ上げ続けないよう、ここで自分自身を無効化する。
	webhookURL := serviceutils.GetAttendanceReminderSlackWebhookUrl()
	if webhookURL == "" {
		logrus.WithFields(logrus.Fields{
			"desc": "Slack Webhook URLが未設定のため実行しない",
		}).Warn("AttendanceMorningReminderJob")
		return nil
	}

	today := time.Now().In(jstLocationForAttendanceReminder)

	holidays, err := businessday.FetchHolidays(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"desc": "祝日一覧を取得できなかったため、土日のみで営業日を判定する",
			"err":  err,
		}).Warn("AttendanceMorningReminderJob")
		holidays = map[string]bool{}
	}

	text, post := buildMorningReminderMessage(today, holidays)
	if !post {
		logrus.WithFields(logrus.Fields{
			"desc":  "本日は営業日ではないため投稿しない",
			"today": today.Format(time.DateOnly),
		}).Info("AttendanceMorningReminderJob")
		return nil
	}

	if err := slackutils.PostMessage(ctx, webhookURL, text); err != nil {
		return errors.Wrap(err, "failed to post attendance morning reminder to slack")
	}

	logrus.WithFields(logrus.Fields{
		"desc": "勤怠連絡リマインド(当日朝)を投稿した",
		"text": text,
	}).Info("AttendanceMorningReminderJob")
	return nil
}

// buildMorningReminderMessageは、buildReminderMessageの当日朝版。翌営業日を
// 探す必要が無く、今日が営業日かどうかだけを見るため、翌営業日探索に伴う
// エラー(maxLookaheadDays超過)は発生しない。
func buildMorningReminderMessage(today time.Time, holidays map[string]bool) (text string, post bool) {
	if !businessday.IsBusinessDay(today, holidays) {
		return "", false
	}
	return "本日の勤怠連絡をお願いします。", true
}
