package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"energyshares-job-scheduler/pkg/timeutils"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type RepaymentDividendJobs struct{}

// RepaymentDividendProcessは、cron(server.goのscheduleCronJob経由)とMIR-1194で
// 追加した管理画面からの手動実行が同時に走るのを防ぐため、tryLockJobExecutionで
// DBレベルの排他ロックを取る(MIR-1249)。手動実行は本番では無効化されているが、
// ローリングデプロイ中は新旧タスクのcronが数十秒〜数分並走する窓が毎週開くため、
// プロセス内mutexではなくDBロックにする必要がある。
func (RepaymentDividendJobs) RepaymentDividendProcess(ctx context.Context) error {
	release, err := tryLockJobExecution(ctx, domain.JOB_EXECUTION_TYPE_DIVIDEND_PAYMENT)
	if err != nil {
		if !isJobExecutionLockedError(err) {
			return err
		}
		logrus.WithFields(logrus.Fields{
			"jobName": jobNameRepaymentDividend,
		}).Info("skip repayment dividend job because previous run is still in progress")
		if JobExecutionTriggerTypeFromContext(ctx) == domain.JOB_EXECUTION_TRIGGER_TYPE_MANUAL {
			return err
		}
		return nil
	}
	defer release()

	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_DIVIDEND_PAYMENT,
		JobName: jobNameRepaymentDividend,
	}, repaymentDividendProcess)
}

func repaymentDividendProcess(ctx context.Context) error {
	curTime := timeutils.LocalZoneTime()
	executeDate := curTime.Format(time.DateOnly)

	var schedules []domain.RepaymentDividendSchedule
	q := factory.DB(ctx)
	// 実行日と完全一致ではなく「予定日を過ぎたもの全て」を対象にする(MIR-1249)。
	// 途中で失敗してRETURN_SCHEDULEDのまま残ったスケジュールを、翌日以降の実行が
	// 自動的に拾い直すようにするため。別途の検知・アラート機構を新設せずに済む
	// (MIR-1114のRePublishと同じ、取りこぼしを次回実行が拾う設計)。
	if err := q.Table("repayment_dividend_schedule").
		Where("DATE(CONVERT_TZ(execute_date, 'UTC', 'Asia/Tokyo')) <= ? AND status = ?", executeDate, domain.RETURN_SCHEDULED).Find(&schedules); err != nil {
		logrus.WithFields(logrus.Fields{
			"execute_date": executeDate,
			"time":         time.Now(),
		}).Error("Fail Settlement Search")
		return err
	}

	// totalNonRetriableFailuresは、支払い(transaction_status=TRANSACTION_OK)や
	// スケジュールのCompleted化が既に確定しており、次回実行では二度と拾われない
	// 失敗(支払調書イベント作成・投資ステータス更新など)の件数。再試行される
	// totalFailedSchedulesとは意味が異なるため分けているが、どちらも最終的な
	// 成否判定(195行)では同列に扱う。
	var totalUnpaidDividends, totalFailedSchedules, totalNonRetriableFailures int

	for _, sched := range schedules {

		// 결산일이 오늘인 분배금 데이터 조회(未払い=transaction_status != TRANSACTION_OKの行のみ)
		dividends, err := domain.UserDividend{}.GetByRepaymentDividendScheduleId(ctx, sched.Id)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"schedule_id":  sched.Id,
				"execute_date": executeDate,
				"err":          err,
				"time":         time.Now(),
			}).Error("Fail User Dividend inquiry")
			totalFailedSchedules++
			continue // 1件のスケジュール取得失敗で他のスケジュールを止めない(MIR-1249)
		}

		// dividends が0件のとき、「全員支払い済み」と「配当行が1件も作られていない」
		// を区別する(MIR-1249)。管理画面のスケジュール作成はスケジュール本体と
		// 配当行の作成が別トランザクションのため、配当行の作成に失敗した(または
		// 未着手の)スケジュールが残っていると、区別せずにCompletedにしてしまい、
		// 誰にも支払われないまま「成功」として処理が終わってしまう。
		if len(dividends) == 0 {
			hasAny, countErr := hasDividendRowsForSchedule(ctx, sched.Id)
			if countErr != nil {
				logrus.WithFields(logrus.Fields{
					"schedule_id":  sched.Id,
					"execute_date": executeDate,
					"err":          countErr,
					"time":         time.Now(),
				}).Error("Fail to check whether schedule has any dividend rows")
				totalFailedSchedules++
				continue
			}
			if !hasAny {
				logrus.WithFields(logrus.Fields{
					"schedule_id":  sched.Id,
					"execute_date": executeDate,
					"time":         time.Now(),
				}).Error("Repayment Dividend schedule has no dividend rows at all; not completing (MIR-1249)")
				totalFailedSchedules++
				continue // 配当行が1件も無いスケジュールを誤ってCompletedにしない
			}
		}

		var unpaidCount, eventFailedCount int
		for _, dividend := range dividends {
			unpaid, eventFailed := processOneDividend(ctx, dividend, sched, executeDate)
			if unpaid {
				unpaidCount++
			}
			if eventFailed {
				eventFailedCount++
			}
		}
		totalNonRetriableFailures += eventFailedCount

		if unpaidCount > 0 {
			totalUnpaidDividends += unpaidCount
			totalFailedSchedules++
			logrus.WithFields(logrus.Fields{
				"schedule_id":  sched.Id,
				"execute_date": executeDate,
				"unpaid_count": unpaidCount,
				"total_count":  len(dividends),
				"time":         time.Now(),
			}).Error("Repayment Dividend schedule has unpaid items; keeping RETURN_SCHEDULED for retry")
			continue // 全員分の支払いが済んでいないため、このスケジュールはCompletedにしない(MIR-1249)
		}

		logrus.WithFields(logrus.Fields{
			"schedule_id":  sched.Id,
			"execute_date": executeDate,
			"time":         time.Now(),
		}).Infof("Success execute %v of schedule id %v", sched.ReturnType, sched.Id)

		if err := sched.Completed(ctx); err != nil {
			logrus.WithFields(logrus.Fields{
				"schedule_id":  sched.Id,
				"execute_date": executeDate,
				"err":          err,
				"time":         time.Now(),
			}).Error("Fail schedule update status to completed")
			totalFailedSchedules++
			continue // 次のスケジュールの処理を続ける(MIR-1249)。このスケジュールはRETURN_SCHEDULEDのまま残り、次回以降に再試行される
		}

		// 最後の償還(残元本 0)を払ったあと、投資ステータスを「償還済み」に更新。この時点でスケジュールは
		// 既にRETURN_COMPLETEDであり、以降の失敗は「予定日以前」の再拾い上げ対象に
		// ならない(次回実行では二度と選ばれない)。retryできない以上、
		// totalNonRetriableFailuresでジョブ自体を失敗として記録し、ログだけで
		// 無音にならないようにする(MIR-1249)。
		if sched.ReturnType == domain.REPAYMENT_RETURN && sched.InvestmentBalance <= 0 {
			shareErr := domain.Share{}.RefundOperationInvestsByProject(ctx, sched.ProjectId, 0)
			if shareErr != nil {
				totalNonRetriableFailures++
				logrus.WithFields(logrus.Fields{
					"schedule_id":       sched.Id,
					"project_id":        sched.ProjectId,
					"execute_date":      executeDate,
					"investmentBalance": sched.InvestmentBalance,
					"err":               shareErr,
					"time":              time.Now(),
				}).Error("Fail to update share status to REFUND_INVEST after repayment (schedule already completed, will not be retried)")
			}
		}

		// 지불조서 이벤트는 배당금에 대해 원천징수를 한 정보를 담는 것이다。
		// 上と同じ理由で、失敗はtotalNonRetriableFailuresへ計上する。
		if sched.ReturnType == domain.DIVIDEND_RETURN {
			if eventErr := domain.CreateDividendExecutedEvent(ctx, sched); eventErr != nil {
				totalNonRetriableFailures++
				logrus.WithFields(logrus.Fields{
					"schedule_id":  sched.Id,
					"execute_date": executeDate,
					"err":          eventErr,
					"time":         time.Now(),
				}).Error("Fail create dividend executed event (schedule already completed, will not be retried)")
			}
		}
	}

	logFields := logrus.Fields{
		"execute_date":           executeDate,
		"schedule_count":         len(schedules),
		"failed_schedules":       totalFailedSchedules,
		"unpaid_dividends":       totalUnpaidDividends,
		"non_retriable_failures": totalNonRetriableFailures,
		"time":                   time.Now(),
	}
	if totalFailedSchedules > 0 || totalNonRetriableFailures > 0 {
		logrus.WithFields(logFields).Error("Repayment Dividend Jobs finished with failures")
		return errors.Errorf(
			"repayment dividend job: %d schedule(s) had failures (%d dividend(s) unpaid, %d non-retriable failure(s))",
			totalFailedSchedules, totalUnpaidDividends, totalNonRetriableFailures,
		)
	}
	logrus.WithFields(logFields).Info("Success Repayment Dividend Jobs")
	return nil
}

// hasDividendRowsForScheduleは、指定したスケジュールに配当行が1件でも
// (支払い済み・未払いを問わず)存在するかを返す(MIR-1249)。管理画面の
// スケジュール作成(CreateRepaymentDividendSchedules)はスケジュール本体と
// 配当行の作成が別トランザクションのため、配当行が1件も作られていない
// スケジュールが残りうる。GetByRepaymentDividendScheduleIdが返す「未払い
// 行が0件」だけでは、全員支払い済みなのか配当行自体が無いのかを区別できない。
func hasDividendRowsForSchedule(ctx context.Context, scheduleId int64) (bool, error) {
	// bean(&domain.Dividend{})を渡さずにCount()を呼ぶと、xormがRefTableを
	// 解決できず論理削除(deleted)の除外条件が付かない。GetByRepaymentDividendScheduleId
	// (Findを使うため除外条件が付く)との間で論理削除済み行の扱いがずれ、
	// 全行が論理削除されたスケジュールを「配当行がある」と誤判定してしまう
	// (このガード自体が防ごうとしている事象)。beanを渡して同じ除外条件を効かせる。
	// GetByRepaymentDividendScheduleIdと同じINNER JOIN userも付ける。JOIN無しで
	// 数えると、dividend.user_idに対応するuser行が無い孤立行まで「ある」と
	// 数えてしまい、GetByRepaymentDividendScheduleId(JOIN経由でその行を除外する)
	// の0件という結果と食い違う。
	count, err := factory.DB(ctx).Table("dividend").
		Join("INNER", "user", "dividend.user_id = user.id").
		Where("dividend.settlement_id = ?", scheduleId).
		Count(&domain.Dividend{})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// processOneDividendは1件の配当・償還支払いを処理する(MIR-1249)。戻り値は
// (unpaid, eventFailed)。unpaid=trueは支払いそのものが未完了で次回再試行が
// 必要なことを、eventFailed=trueは支払いは成立済み(TRANSACTION_OK)だが
// 支払調書イベント(CreateUserDividendExecutedEvent)の作成に失敗したことを表す。
// 後者はGetByRepaymentDividendScheduleIdがTRANSACTION_OKの行を除外するため
// 再試行では拾えず、unpaidとは別枠で集計する。
//
// deferでpanicを回収しているのは、単純なcontinueでは拾えないpanicが
// repaymentDividendProcess全体を巻き戻し、ジョブ全体を止めてしまう
// (このチケットが解消しようとしている「1件の失敗が全員を止める」問題が
// panic経由で再発する)ため。DoTransactionが開くWithTxはpanic中でも
// defer経由でロールバックされるためDBの整合性は保たれる。
func processOneDividend(ctx context.Context, dividend domain.UserDividend, sched domain.RepaymentDividendSchedule, executeDate string) (unpaid bool, eventFailed bool) {
	// paidは、DoTransactionが成功したか(=支払いはコミット済みか)を表す。panicが
	// DoTransaction成功後(CreateUserDividendExecutedEvent側)で起きた場合、支払い
	// 済みの行をunpaidとして扱うと実態と食い違うため、deferのrecoverで参照する。
	var paid bool
	defer func() {
		if r := recover(); r != nil {
			if paid {
				eventFailed = true
			} else {
				unpaid = true
			}
			logrus.WithFields(logrus.Fields{
				"schedule_id":  sched.Id,
				"dividend_id":  dividend.Dividend.Id,
				"execute_date": executeDate,
				"paid":         paid,
				"panic":        fmt.Sprintf("%v", r),
				"time":         time.Now(),
			}).Error("Panic recovered while processing dividend")
		}
	}()

	ledgerEventFailed, err := dividend.DoTransaction(ctx, domain.TRANSACTION_OK)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"schedule_id":  sched.Id,
			"dividend_id":  dividend.Dividend.Id,
			"execute_date": executeDate,
			"err":          err,
			"time":         time.Now(),
		}).Error("Fail Do Transaction update")
		return true, false // 1件の支払い失敗で残りの会員への支払いを止めない(MIR-1249)
	}
	paid = true
	if ledgerEventFailed {
		// 支払い自体は成立済みだが、顧客元帳イベント(CreateCustomerLedgerDividendEvent)
		// の作成にコミット後失敗している。支払いをやり直す対象(unpaid)へは戻さず、
		// CreateUserDividendExecutedEventの失敗と同じeventFailedとして計上する。
		eventFailed = true
	}

	// 支払い自体は成立済み(transaction_status=TRANSACTION_OK)。ここでの
	// イベント作成失敗は、支払いをやり直す対象へは戻さない(戻すと二重払いになる)。
	if dividend.TransactionType == domain.DEPOSIT_DIVIDEND {
		if eventErr := domain.CreateUserDividendExecutedEvent(ctx, dividend, sched.ReturnRate); eventErr != nil {
			logrus.WithFields(logrus.Fields{
				"schedule_id":  sched.Id,
				"dividend_id":  dividend.Dividend.Id,
				"execute_date": executeDate,
				"err":          eventErr,
				"time":         time.Now(),
			}).Error("Fail create dividend executed event (payment already succeeded, not retried)")
			return false, true
		}
	}
	return false, eventFailed
}
