package jobs

import (
	"context"
	"time"

	"energyshares-job-scheduler/pkg/timeutils"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/sirupsen/logrus"
)

type DepositCheckJobs struct {
}

// 入金確認：入金履歴を照会して入金リクエスト処理
// 払い戻し確認：未使用金額に対する払い戻し
// 払い戻しモデル必要：正確な払い戻しのための払い戻しリクエストプロセス必要

func (DepositCheckJobs) VaDepositNotUseItemCheckJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_VA_DEPOSIT_NOT_USE_ITEM_CHECK,
		JobName: jobNameVaDepositNotUseItemCheck,
	}, vaDepositNotUseItemCheckJob)
}

func vaDepositNotUseItemCheckJob(ctx context.Context) error {
	logrus.WithFields(logrus.Fields{
		"time": time.Now(),
	}).Info("Va Transaction Check Start")

	_, items, err := domain.VaDeposit{}.GetNotUseDeposit(ctx)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"Va Deposit Items": items,
		"time":             time.Now(),
	}).Info("Va Transaction Checking")

	// MIR-1251: depositAutoProcessJobと同じ理由で、1件の失敗でループ全体を
	// 打ち切らない。失敗しても次のVA入金へ進む。この関数は手動実行
	// (POST /va-deposit-not-use-item-check-job、jobs-schedulerのミドルウェアが
	// リクエスト単位でtxを開いている)経由でも呼ばれるため、素のengineだけを
	// 持つctxへ差し替える。差し替えないと手動実行時は全件が1つのトランザクション
	// を共有し、最後の集計エラーで成功分も含めて全部ロールバックされてしまう。
	// DepositEquityProcess自体はトランザクションを持たない(既存の実装のまま)
	// が、内部で呼ぶDepositCheckProcessはfactory.InTransaction(ctx)を見て
	// 自前で1件ごとにトランザクションを開くため、ここで独自にWithTxを張って
	// しまうと、その中はもうトランザクション内と判定されてしまい、
	// DepositCheckProcess側の「トランザクションを開く前に読み取りを済ませる」
	// 対策が効かなくなる(実際にCIの並行実行テストで踏んだ)。
	engine, err := factory.Engine(ctx)
	if err != nil {
		return factory.ErrorDB.New(err, "Fail Get DB Engine")
	}
	engineCtx := context.WithValue(ctx, factory.ConextDBType, engine)

	var failedCount int
	for i := range items {
		va, err := domain.UserVirtualAccount{}.GetUserVirtualAccountByBranchAndAccountNumber(engineCtx, items[i].VaBranchCode, items[i].VaAccountNumber)
		if err != nil {
			failedCount++
			logrus.WithFields(logrus.Fields{
				"va_deposit_id": items[i].Id,
				"err":           err,
			}).Error("VaDepositNotUseItemCheckJob: fail Get Virtual Account by BranchCode and AccountNumber, skip this item")
			continue
		}
		if va == nil {
			continue
		}
		items[i].UserId = va.UserId
		if err := items[i].DepositEquityProcess(engineCtx); err != nil {
			failedCount++
			logrus.WithFields(logrus.Fields{
				"va_deposit_id": items[i].Id,
				"user_id":       items[i].UserId,
				"err":           err,
			}).Error("VaDepositNotUseItemCheckJob: fail DepositEquityProcess, skip this item")
			continue
		}
	}

	logrus.WithFields(logrus.Fields{
		"vaTransactions": items,
		"failedCount":    failedCount,
		"time":           time.Now(),
	}).Info("Va Transaction Check End")

	if failedCount > 0 {
		return errors.Errorf("VaDepositNotUseItemCheckJob: %d/%d items failed", failedCount, len(items))
	}
	return nil

}

// 00時を基準に処理
func (DepositCheckJobs) DepositRequestExpireJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_DEPOSIT_REQUEST_EXPIRE,
		JobName: jobNameDepositRequestExpire,
	}, depositRequestExpireJob)
}

// findExpiredDepositRequests は入金有効期限が切れた投資申請を、待機者由来かどうかの判定情報
// (InvestmentWaitersShareId)付きで取得する。investment_waiters.share_id は
// investment_waiters_winning へ移行済みのため、待機者由来かどうかは中間テーブル経由で判定する
// (MIR-1118)。この判定結果で入金期限切れメールの種別(WAITER_REQUEST)を決めているため、
// depositRequestExpireJob 本体(メール送信を含む)とは分けてクエリ単体でテストできるようにしている。
// JOINではなくスカラーサブクエリにしているのは、万一同じshare_idを指す当選行が複数存在しても
// 行が水増しされないようにするため(水増しされると UpdateInvestStatus が同じshareに対して
// 複数回呼ばれ、二重キャンセル・二重メールになる)。
func findExpiredDepositRequests(ctx context.Context) ([]domain.DepositRequestDetail, int64, error) {
	q := factory.DB(ctx).Table("deposit_request").
		Join("INNER", "share", "share.id = deposit_request.share_id").
		Select(""+
			"deposit_request.*,"+
			"share.*,"+
			domain.InvestmentWaitersWinningShareIdByShareExpr).
		Where("(deposit_request.expire_deposit_date < ?) AND (share.invest_status = ?) ", time.Now().Format(time.DateTime), domain.REQUEST_INVEST)

	var items []domain.DepositRequestDetail
	totalCount, err := q.FindAndCount(&items)
	return items, totalCount, err
}

func depositRequestExpireJob(ctx context.Context) error {
	// バッチの場合、長時間処理される可能性があるため、SQLセッションを修正で処理する。
	logrus.WithFields(logrus.Fields{
		"time": time.Now(),
	}).Info("DepositRequestExpireJob Start")

	// 入金ステータス変更
	// 投資完了処理
	// 入金有効期限超過 -> 投資キャンセル
	// 投資がキャンセルされた場合払い戻し
	// 払い戻しバッチモデル受取

	items, totalCount, err := findExpiredDepositRequests(ctx)
	if err != nil {
		return err
	}

	for i, _ := range items {
		// 投資キャンセル
		var investMailType domain.INVEST_MAIL_TYPE

		// 投資待機者が期限内に入金しないため入金有効期限が切れた場合
		if items[i].InvestmentWaitersShareId != "" {
			investMailType = domain.WAITER_REQUEST
		}

		if err := items[i].UpdateInvestStatus(ctx, domain.CANCEL_INVEST, 0, investMailType); err != nil {
			logrus.WithFields(logrus.Fields{
				"share_id":            items[i].ShareId,
				"project_id":          items[i].DepositRequest.ProjectId,
				"status":              items[i].InvestStatus,
				"deposit_status":      items[i].DepositCheckStatus,
				"invest_date":         items[i].Share.InvestDate,
				"expire_deposit_date": items[i].DepositRequest.ExpireDepositDate,
				"time":                time.Now(),
				"error":               err,
			}).Error("Fail Expire Shares")
		}
	}

	logrus.WithFields(logrus.Fields{
		"category": "Deposit Check",
		"count":    totalCount,
		"time":     time.Now(),
	}).Info("DepositRequestExpireJob End")
	return nil
}

func (DepositCheckJobs) DepositAutoProcessJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_DEPOSIT_AUTO_PROCESS,
		JobName: jobNameDepositAutoProcess,
	}, depositAutoProcessJob)
}

func depositAutoProcessJob(ctx context.Context) error {
	logrus.WithFields(logrus.Fields{
		"title": "DepositAutoProcessJob",
		"desc":  "=======Start Jobs=======",
		"time":  time.Now(),
	}).Info(" DepositCheckJobs")

	jobExecuteTime, err := domain.JobExecuteTime{}.GetJobLastExecuteTime(ctx)
	if err != nil {
		return err
	}
	lastExecuteTime := timeutils.LocalZoneTime()

	if err := jobExecuteTime.UpdateDepositProcessTime(ctx, jsonutils.JSONTime(lastExecuteTime)); err != nil {
		return errors.Wrap(err, "fail update deposit process time")
	}

	var depositRequestItems []domain.DepositRequest
	q := factory.DB(ctx).Table("deposit_request").Where("deposit_check_status = ?", domain.NO_DEPOSIT).
		And("project_type = ?", domain.EQUITY).And("DATEDIFF(expire_deposit_date,now()) > 0").
		Asc("expire_deposit_date").Asc("deposit_request.id").
		Join("LEFT", "user_account", "deposit_request.user_id = user_account.user_id").
		Where("user_account.balance > 0")

	err = q.Find(&depositRequestItems)
	if err != nil {
		return err
	}

	if len(depositRequestItems) <= 0 {
		logrus.WithFields(logrus.Fields{
			"title": "DepositAutoProcessJob items 0",
			"desc":  "=======End Jobs=======",
			"time":  time.Now(),
		}).Info(" DepositCheckJobs")
		return nil
	}
	failedCount := processDepositRequestItems(ctx, depositRequestItems)

	logrus.WithFields(logrus.Fields{
		"title":       "DepositAutoProcessJob ",
		"desc":        "=======End Jobs=======",
		"total":       len(depositRequestItems),
		"failedCount": failedCount,
		"time":        time.Now(),
	}).Info(" DepositCheckJobs")

	if failedCount > 0 {
		return errors.Errorf("DepositAutoProcessJob: %d/%d items failed", failedCount, len(depositRequestItems))
	}
	return nil
}

// processDepositRequestItemsはMIR-1251で切り出した。1件の競合(残高の
// バージョン不一致等)でループ全体を打ち切ると、その人より後ろに並ぶ全員が
// 次回実行まで待たされるため、1件ずつ独立して処理し、失敗しても次の
// 入金要求へ進む。失敗件数を返す(呼び出し元がジョブ全体の成否・ログを
// 決める)。SQL(depositAutoProcessJob側)がMySQL固有の関数を使っており
// sqliteでテストできないため、この関数はクエリ結果を受け取るだけにして
// 単体テストできるようにしている。
func processDepositRequestItems(ctx context.Context, items []domain.DepositRequest) (failedCount int) {
	// MIR-1251: この関数は手動実行(POST /deposit-auto-process-job/execute、
	// jobs-schedulerのミドルウェアがリクエスト単位でtxを開いている)経由でも
	// 呼ばれる。DepositCheckProcess自身はfactory.InTransaction(ctx)がtrueなら
	// 既存のtxにそのまま乗ってしまうため、素のengineだけを持つctxへ差し替え、
	// 呼び出し元(ここ)がどんなtxの中にいてもDepositCheckProcess自身が
	// 1件ごとに新しいトランザクションを開くようにする。差し替えないと手動
	// 実行時は全件が1つのトランザクションを共有し、最後の集計エラーで成功分
	// も含めて全部ロールバックされてしまう。DepositCheckProcess側の実装
	// (トランザクションを開く前に必要な読み取りを済ませる)がそのまま効くのは
	// この形だけで、ここで独自にWithTxを張ってしまうと、その中はもう
	// トランザクション内と判定されるため、内側の「開く前に読む」が効かなく
	// なる(実際にCIの並行実行テストで踏んだ)。
	engine, err := factory.Engine(ctx)
	if err != nil {
		logrus.WithError(err).Error("DepositAutoProcessJob: fail to get db engine")
		return len(items)
	}
	engineCtx := context.WithValue(ctx, factory.ConextDBType, engine)

	for i := range items {
		account, err := domain.UserAccount{}.GetByUserId(engineCtx, items[i].UserId)
		if err != nil {
			failedCount++
			logrus.WithFields(logrus.Fields{
				"deposit_request_id": items[i].Id,
				"user_id":            items[i].UserId,
				"err":                err,
			}).Error("DepositAutoProcessJob: fail to get user account, skip this item")
			continue
		}
		if err := account.DepositCheckProcess(engineCtx, &items[i], domain.NOT_IMMEDIATE); err != nil {
			failedCount++
			logrus.WithFields(logrus.Fields{
				"deposit_request_id": items[i].Id,
				"user_id":            items[i].UserId,
				"err":                err,
			}).Error("DepositAutoProcessJob: fail to process deposit request, skip this item")
			continue
		}
	}
	return failedCount
}
