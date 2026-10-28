package jobs

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-share-core-api/service"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// TODO: 1. 投資待機処理プロジェクトの投資可能口数が存在する場合、自動投資を実行する
// TODO: 2. 投資待機処理プロセス完了後は入金要求確認ステータスに変更する
// TODO: 3. プロジェクト締切時間の4日前まで自動投資を実行する

type WaitingInvestmentJobs struct {
	investmentService service.InvestmentService
}

// findWaitingInvestmentItems は対象プロジェクトの待機者一覧を申請日時順で取得する
// （WAITING, DEPOSIT_REQUEST, PROPOSAL）。deposit_check_status・繰上げ済み累計
// (AdvancedQty)はいずれもスカラーサブクエリで解決する(MIR-1068)。investment_waiters_winning
// への実JOINにすると、1申込に複数当選がある場合に行が当選件数ぶん水増しされ、
// AutoInvestmentJobの対象件数がずれるため(deposit_check_jobs.goのfindExpiredDepositRequests
// と同じ理由)。
// AutoInvestmentJob 内の投資処理前・処理後の2箇所で同じクエリを使うため、ここに切り出している。
func findWaitingInvestmentItems(ctx context.Context, projectId int64) ([]domain.InvestmentWaiters, error) {
	var items []domain.InvestmentWaiters
	err := factory.DB(ctx).
		Table(domain.InvestmentWaiters{}.TableName()).
		Select(domain.DEFAULT_INVESTMENT_WAIT_LIST+", "+domain.LatestWaiterDepositCheckStatusExpr+", "+domain.AdvancedQtyExpr).
		Join("INNER", domain.User{}.TableName(), "investment_waiters.user_id = user.id").
		Where(
			"investment_waiters.project_id = ? and investment_waiters.status in (?, ?, ?)",
			projectId,
			domain.WAITING,
			domain.DEPOSIT_REQUEST,
			domain.PROPOSAL,
		).
		Asc("investment_waiters.project_id", "investment_waiters.waiting_date").
		Find(&items)
	return items, err
}

// AutoInvestmentJob は投資待機者の自動繰り上げ投資バッチジョブ。
// **is_wait_invest=true**のプロジェクトに対して、待機者を申請日時順に処理し、残口数が足りる限り WaitInvest で投資を実行する。
// 全待機者の処理後、残り待機者が0かつ残口数>=最小投資口数であればキャンセル待ち状態を解除する。
func (w *WaitingInvestmentJobs) AutoInvestmentJob(ctx context.Context) error {
	logrus.WithFields(logrus.Fields{
		"title": "AutoInvestmentJob",
		"desc":  "=======Start Jobs=======",
		"time":  time.Now(),
	}).Info(" AutoInvestmentJob")

	var projects []domain.Project

	// 募集中 or 募集終了かつキャンセル待ち有効なプロジェクトを取得
	q := factory.DB(ctx)
	q.Table("project").
		Select(domain.PROJECT_DETAIL_COLUMN_LIST).
		Where(
			"project.project_status in (?, ?) AND project.is_wait_invest = ?",
			domain.OPEN_FUNDING, domain.END_FUNDING, true)
	err := q.Find(&projects)
	if err != nil {
		return err
	}
	for _, project := range projects {
		// 対象プロジェクトの待機者一覧を申請日時順で取得（WAITING, DEPOSIT_REQUEST, PROPOSAL）。
		items, err := findWaitingInvestmentItems(ctx, project.Id)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"project_id": project.Id,
				"category":   "auto investment ",
				"items":      items,
				"time":       time.Now(),
				"err":        err,
			}).Info("InvestmentWaiter")
			return err
		}

		// 待機者を順番に処理し、残口数が足りる限り投資を実行する
		for _, item := range items {
			// 繰上げ可能残(申請口数-繰上げ済み累計)。バッチは既存仕様のまま、この待機者の
			// 繰上げ可能残の全額を投資できるときだけ実行し、部分的な繰上げはしない
			// (MIR-1068、2026-09-28 河内さんの判断)。
			remaining := item.Qty - item.AdvancedQty
			if remaining <= 0 {
				// 累計が既に申請口数に達している(通常はStatusがWAITING_ENDに変わって
				// この一覧に含まれないはずの異常値)。この待機者だけスキップして次へ。
				logrus.WithFields(logrus.Fields{
					"project_id":   item.ProjectId,
					"user_id":      item.UserId,
					"waiter_id":    item.Id,
					"qty":          item.Qty,
					"advanced_qty": item.AdvancedQty,
					"time":         time.Now(),
				}).Warn(" skip wait invest: no remaining qty")
				continue
			}

			// 最新の残口数を取得（前の待機者の投資で減少している可能性がある）
			p, err := domain.Project{}.GetShareRemainingQty(ctx, item.ProjectId)
			if err != nil {
				return err
			}
			if p == nil {
				return errors.New("not found project")
			}

			// 残口数不足またはキャンセル待ち無効の場合、以降の待機者もスキップ
			if !(p.RemainingQty >= remaining && p.IsWaitInvest) {
				logrus.WithFields(logrus.Fields{
					"project_id":             item.ProjectId,
					"user_id":                item.UserId,
					"amount":                 item.Amount,
					"waiter_id":              item.Id,
					"qty":                    item.Qty,
					"remaining_qty":          remaining,
					"project_remaining_qty":  p.RemainingQty,
					"project_is_wait_invest": p.IsWaitInvest,
					"time":                   time.Now(),
				}).Info(" break wait invest")
				break
			}
			user, err := domain.User{}.GetById(ctx, item.UserId)
			if err != nil {
				return err
			}

			share, err := service.InvestmentService{}.WaitInvest(ctx, user, &item, remaining, false)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"waiter_id":  item.Id,
					"project_id": item.ProjectId,
					"user_id":    item.UserId,
					"amount":     item.Amount,
					"qty":        remaining,
					"err":        err,
					"time":       time.Now(),
				}).Error(" failed wait invest")
				return err
			}
			logrus.WithFields(logrus.Fields{
				"waiter_id":      item.Id,
				"project_id":     item.ProjectId,
				"user_id":        item.UserId,
				"amount":         item.Amount,
				"qty":            remaining,
				"share_id":       share.Id,
				"share_status":   share.InvestStatus,
				"share_amount":   share.Amount,
				"share_qty":      share.Qty,
				"share_owner_id": share.OwnerId,
				"time":           time.Now(),
			}).Info(" success wait invest")

		}

		// 投資処理後、残りの待機者を再取得して待機解除判定を行う
		items, err = findWaitingInvestmentItems(ctx, project.Id)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"project_id": project.Id,
				"category":   "auto investment ",
				"items":      items,
				"time":       time.Now(),
				"err":        err,
			}).Info("InvestmentWaiter")
			return err
		}
		if len(items) == 0 {
			p, err := domain.Project{}.GetById(ctx, project.Id)
			if err != nil {
				return err
			}
			if p.RemainingQty >= p.MinShareQty && p.ProjectStatus != domain.END_FUNDING {
				// NOTE: キャンセル待機者が存在せず、かつ残口数が最小投資口数以上の場合はキャンセル待機状態を解除する
				// ただし、募集終了ステータスの場合はキャンセル待ち状態を維持する
				p.IsWaitInvest = false
				err = p.Update(ctx)
				if err != nil {
					return err
				}
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"title": "AutoInvestmentJob",
		"desc":  "=======End Jobs=======",
		"time":  time.Now(),
	}).Info(" AutoInvestmentJob")

	return nil
}

// CheckInvestmentProposeJob は投資提案リンクの有効期限切れを処理するバッチジョブ。
// 毎日0:00に実行され、期限切れの投資提案リンクを持つ待機者のステータスをWAITING_ENDに変更する。
func (w *WaitingInvestmentJobs) CheckInvestmentProposeJob(ctx context.Context) error {
	logrus.WithFields(logrus.Fields{
		"title": "CheckInvestmentProposeJob",
		"desc":  "=======Start Jobs=======",
		"time":  time.Now(),
	}).Info(" CheckInvestmentProposeJob")

	var investmentWaitersList []domain.InvestmentWaiters

	// 毎日0時に以下の条件でバッチ処理を実行する:
	// 1. 現在時刻より前に作成された(created)投資提案(investment_request)リンクが存在し
	// 2. 未失効(is_expire=0)の場合（投資提案リンクは作成後当日0時に失効）
	// 3. 投資提案リンクのステータスを失効に変更する
	factory.DB(ctx).
		Table("investment_waiters").
		Select("investment_waiters.*").
		Join("INNER", "investment_request", "investment_waiters.id = investment_request.investment_waiters_id").
		Where("investment_waiters.status = 'PROPOSAL'").
		And("investment_request.created < now()").
		And("investment_request.is_expire = 0").
		Find(&investmentWaitersList)

	for _, investmentWaiter := range investmentWaitersList {
		if err := investmentWaiter.ChangeStatus(ctx, domain.WAITING_END, 1); err != nil {
			logrus.WithFields(logrus.Fields{
				"title": "CheckInvestmentProposeJob",
				"desc":  "=======FAIL invest waiter status change=======",
				"time":  time.Now(),
				"err":   err,
			}).Error("Investment Waiter Fail Send Mail")
			return err
		}
	}

	logrus.WithFields(logrus.Fields{
		"title": "CheckInvestmentProposeJob",
		"desc":  "=======End Jobs=======",
		"time":  time.Now(),
	}).Info(" CheckInvestmentProposeJob")

	return nil
}
