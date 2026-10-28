package mysql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"xorm.io/xorm"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

// 取得クエリでは、loadWinningsAndStatusForAppsで当選Share情報をセットしステータスを更新すること。
// ステータスはDBレベルではwinning_at/lost_at/canceled_atの有無で判定し、厳密なLOTTERY_STATUSの区別はGo側のComputeStatus()に任せている。

type lotteryApplicationRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.LotteryApplicationRepository {
	return &lotteryApplicationRepository{dbFactory: dbFactory}
}

func (r *lotteryApplicationRepository) CreateIfNotExists(ctx context.Context, app domain.LotteryApplication) (string, bool, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return "", false, errors.Wrap(err, "err get db")
	}
	app.Id = uuid.New().String()
	if app.AppliedAt.IsZero() {
		app.AppliedAt = jsonutils.JSONTime(time.Now())
	}
	sql := `
      INSERT INTO lottery_application                                                                                                          
          (id, project_id, user_id, qty, unit_price, amount, applied_at)
      SELECT ?, ?, ?, ?, ?, ?, ?                                                                                                               
      WHERE NOT EXISTS (
          SELECT 1 FROM lottery_application                                                                                                    
          WHERE project_id = ? AND user_id = ? AND deleted IS NULL AND canceled_at IS NULL                                                                             
      )
  `
	result, err := db.Exec(sql,
		app.Id, app.ProjectId, app.UserId, app.Qty, app.UnitPrice, app.Amount, time.Time(app.AppliedAt),
		app.ProjectId, app.UserId,
	)
	if err != nil {
		return "", false, errors.Wrap(err, "err insert lottery application")
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return "", false, nil // 重複により未挿入
	}
	return app.Id, true, nil
}

// GetAll は、options で指定された条件に基づいて抽選申込を取得します。
func (r *lotteryApplicationRepository) GetAll(ctx context.Context, options ...cdomain.LotteryApplicationQueryOption) (int64, []domain.LotteryApplication, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.LotteryApplicationQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	totalCount, err := r.buildQuery(db, queryOpts).Count()
	if err != nil {
		return 0, nil, errors.Wrap(err, "err count lottery applications")
	}

	type appRow struct {
		domain.LotteryApplication `xorm:"extends"`
		ProjectName               string `xorm:"project_name"`
		ProjectStatus             string `xorm:"project_status"`
		TargetShareQty            int64  `xorm:"target_share_qty"`
		MinShareQty               int64  `xorm:"min_share_qty"`
		UserName                  string `xorm:"user_name"`
		UserEmail                 string `xorm:"user_email"`
		UserAccountBalance        *int64 `xorm:"user_account_balance"`
		UserType                  domain.USER_TYPE `xorm:"user_type"`
	}
	appRows := make([]appRow, 0)
	if err = r.buildQuery(db, queryOpts).
		LeftJoin(domain.UserAccount{}, "user_account.user_id = lottery_application.user_id").
		Select(
			"lottery_application.*",
			"project.name AS project_name",
			"project.project_status AS project_status",
			"project.target_share_qty AS target_share_qty",
			"project.min_share_qty AS min_share_qty",
			"user.name AS user_name",
			"user.email AS user_email",
			"user_account.balance AS user_account_balance",
			"user.user_type AS user_type",
		).
		// project / user / user_account を JOIN しており、いずれも id を持つ。
		// 修飾しないと sortby=applied_at,id が「あいまい」で落ちる。
		SortOrderForTable(domain.LotteryApplication{}.TableName(), queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		Find(&appRows); err != nil {
		return 0, nil, errors.Wrap(err, "err find lottery applications")
	}

	apps := make([]domain.LotteryApplication, len(appRows))
	for i, row := range appRows {
		apps[i] = row.LotteryApplication
		apps[i].ProjectName = row.ProjectName
		apps[i].ProjectStatus = domain.PROJECT_STATUS(row.ProjectStatus)
		apps[i].TargetShareQty = row.TargetShareQty
		apps[i].MinShareQty = row.MinShareQty
		apps[i].UserName = row.UserName
		apps[i].UserEmail = row.UserEmail
		apps[i].UserAccountBalance = row.UserAccountBalance
		apps[i].UserType = row.UserType
	}

	if err = r.loadWinningsAndStatusForApps(db, apps); err != nil {
		return 0, nil, err
	}

	return totalCount, apps, nil
}

// buildQuery は共通フィルター条件（JOIN・Equal・Like・範囲・ステータス）を適用した orm.Orm を返す。
// Count と Find の両クエリで再利用することで、総件数と取得結果のフィルター条件を一致させる。
func (r *lotteryApplicationRepository) buildQuery(db xorm.Interface, queryOpts cdomain.LotteryApplicationQueryOptionFields) orm.Orm {
	return r.applyStatusFilter(
		orm.Table(db, domain.LotteryApplication{}).
			LeftJoin(domain.Project{}, "project.id = lottery_application.project_id AND project.deleted IS NULL").
			LeftJoin(domain.User{}, "user.id = lottery_application.user_id AND user.deleted IS NULL").
			Equal("lottery_application.project_id", queryOpts.ProjectId()).
			Equal("lottery_application.user_id", queryOpts.UserId()).
			Like("user.name", queryOpts.UserName()).
			Like("project.name", queryOpts.ProjectName()).
			GTE("lottery_application.amount", queryOpts.AmountMin()).
			LTE("lottery_application.amount", queryOpts.AmountMax()).
			GTE("lottery_application.applied_at", queryOpts.GTERequestDate()).
			LTE("lottery_application.applied_at", queryOpts.LTERequestDate()),
		queryOpts.Status(),
	)
}

// applyStatusFilter はステータス条件をDBレベルのタイムスタンプ比較に変換して適用する。
// WINNING と PARTIAL_WINNING はDBレベルでは区別できないため同じ条件（winning_at IS NOT NULL）を使用し、
// 厳密な区別は Go 側の ComputeStatus() が担う。
func (r *lotteryApplicationRepository) applyStatusFilter(q orm.Orm, status *domain.LOTTERY_STATUS) orm.Orm {
	if status == nil {
		return q
	}
	switch *status {
	case domain.LOTTERY_STATUS_APPLIED:
		return q.
			IsNull("lottery_application.winning_at").
			IsNull("lottery_application.lost_at").
			IsNull("lottery_application.canceled_at")
	case domain.LOTTERY_STATUS_WINNING, domain.LOTTERY_STATUS_PARTIAL_WINNING:
		return q.
			NotNull("lottery_application.winning_at").
			IsNull("lottery_application.canceled_at")
	case domain.LOTTERY_STATUS_LOST:
		return q.
			NotNull("lottery_application.lost_at").
			IsNull("lottery_application.canceled_at")
	case domain.LOTTERY_STATUS_CANCELED:
		return q.NotNull("lottery_application.canceled_at")
	}
	return q
}

func (r *lotteryApplicationRepository) GetById(ctx context.Context, id string) (domain.LotteryApplication, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.LotteryApplication{}, errors.Wrap(err, "err get db")
	}

	type appRow struct {
		domain.LotteryApplication `xorm:"extends"`
		ProjectName               string `xorm:"project_name"`
		ProjectStatus             string `xorm:"project_status"`
		TargetShareQty            int64  `xorm:"target_share_qty"`
		MinShareQty               int64  `xorm:"min_share_qty"`
		UserName                  string `xorm:"user_name"`
		UserType                  domain.USER_TYPE `xorm:"user_type"`
	}
	row := appRow{}
	has, err := orm.Table(db, domain.LotteryApplication{}).
		Select(
			"lottery_application.*",
			"project.name AS project_name",
			"project.project_status AS project_status",
			"project.target_share_qty AS target_share_qty",
			"project.min_share_qty AS min_share_qty",
			"user.name AS user_name",
			"user.user_type AS user_type",
		).
		LeftJoin(domain.Project{}, "project.id = lottery_application.project_id AND project.deleted IS NULL").
		LeftJoin(domain.User{}, "user.id = lottery_application.user_id AND user.deleted IS NULL").
		Equal("lottery_application.id", &id).
		Get(&row)
	if err != nil {
		return domain.LotteryApplication{}, errors.Wrap(err, "err get lottery application")
	}
	if !has {
		return domain.LotteryApplication{}, cdomain.ErrItemNotFound
	}

	app := row.LotteryApplication
	app.ProjectName = row.ProjectName
	app.ProjectStatus = domain.PROJECT_STATUS(row.ProjectStatus)
	app.TargetShareQty = row.TargetShareQty
	app.MinShareQty = row.MinShareQty
	app.UserName = row.UserName
	app.UserType = row.UserType

	apps := []domain.LotteryApplication{app}
	if err = r.loadWinningsAndStatusForApps(db, apps); err != nil {
		return domain.LotteryApplication{}, err
	}

	return apps[0], nil
}

func (r *lotteryApplicationRepository) LockById(ctx context.Context, id string) (domain.LotteryApplication, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.LotteryApplication{}, errors.Wrap(err, "err get db")
	}

	application := domain.LotteryApplication{}
	has, err := db.
		SQL("SELECT lottery_application.* FROM lottery_application WHERE lottery_application.id = ? FOR UPDATE", id).
		Get(&application)
	if err != nil {
		return domain.LotteryApplication{}, errors.Wrap(err, "err lock lottery application")
	}
	if !has {
		return domain.LotteryApplication{}, cdomain.ErrItemNotFound
	}

	return application, nil
}

func (r *lotteryApplicationRepository) IsAlreadyApplied(ctx context.Context, projectId int64, userId int64) (bool, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return false, errors.Wrap(err, "err get db")
	}

	count, err := orm.Table(db, domain.LotteryApplication{}).
		Equal("project_id", projectId).
		Equal("user_id", userId).
		IsNull("deleted").
		IsNull("canceled_at").
		Count()
	if err != nil {
		return false, errors.Wrap(err, "err check already applied")
	}

	return count > 0, nil
}

// MarkAsWinning は、指定されたIDの抽選申込を当選状態にマークします。キャンセル済みの場合は更新されず false を返します。
func (r *lotteryApplicationRepository) MarkAsWinning(ctx context.Context, id string) (bool, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return false, errors.Wrap(err, "err get db")
	}

	affected, err := db.Table(domain.LotteryApplication{}).
		ID(id).
		Where("canceled_at IS NULL").
		Update(&domain.LotteryApplication{
			WinningAt: jsonutils.JSONTime(time.Now()),
		})
	if err != nil {
		return false, errors.Wrap(err, "err mark lottery application as winning")
	}
	return affected > 0, nil
}

// MarkAsLost は、指定されたIDの抽選申込を落選状態にマークします。キャンセル済みの場合は更新されず false を返します。
func (r *lotteryApplicationRepository) MarkAsLost(ctx context.Context, id string) (bool, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return false, errors.Wrap(err, "err get db")
	}

	affected, err := db.Table(domain.LotteryApplication{}).
		ID(id).
		Where("canceled_at IS NULL").
		Update(&domain.LotteryApplication{
			LostAt: jsonutils.JSONTime(time.Now()),
		})
	if err != nil {
		return false, errors.Wrap(err, "err mark lottery application as lost")
	}
	return affected > 0, nil
}

// MarkAsCanceled は、指定されたIDの抽選申込をキャンセル状態にマークします。当選済み・落選済み・キャンセル済みの場合は更新されず false を返します。
func (r *lotteryApplicationRepository) MarkAsCanceled(ctx context.Context, id string) (bool, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return false, errors.Wrap(err, "err get db")
	}

	affected, err := db.Table(domain.LotteryApplication{}).
		ID(id).
		Where("winning_at IS NULL").
		Where("lost_at IS NULL").
		Where("canceled_at IS NULL").
		Update(&domain.LotteryApplication{
			CanceledAt: jsonutils.JSONTime(time.Now()),
		})
	if err != nil {
		return false, errors.Wrap(err, "err mark lottery application as canceled")
	}
	return affected > 0, nil
}

// loadWinningsAndStatusForApps は複数の申込に対して当選Share情報をバッチ取得しセットする（N+1回避）。
// lottery_winning → share をJOINして申込IDでグループ化し、SetWinnings を呼ぶ。
func (r *lotteryApplicationRepository) loadWinningsAndStatusForApps(db xorm.Interface, apps []domain.LotteryApplication) error {
	if len(apps) == 0 {
		return nil
	}

	appIds := make([]string, len(apps))
	for i, a := range apps {
		appIds[i] = a.Id
	}

	type winningRow struct {
		LotteryApplicationId string `xorm:"lottery_application_id"`
		domain.Share         `xorm:"extends"`
	}
	rows := make([]winningRow, 0)
	session := db.Table(domain.LotteryWinning{}).
		Select("lottery_winning.lottery_application_id, share.*").
		Join("INNER", domain.Share{}, "share.id = lottery_winning.share_id AND share.deleted IS NULL").
		In("lottery_winning.lottery_application_id", appIds)

	if err := session.Find(&rows); err != nil {
		return errors.Wrap(err, "err load winnings for apps")
	}
	sharesByAppId := make(map[string][]domain.Share, len(apps))
	for _, row := range rows {
		sharesByAppId[row.LotteryApplicationId] = append(sharesByAppId[row.LotteryApplicationId], row.Share)
	}

	for i := range apps {
		apps[i].SetWinnings(sharesByAppId[apps[i].Id])
	}

	return nil
}

func (r *lotteryApplicationRepository) GetWinningQtyTotalsByProjectID(ctx context.Context, projectIDs []int64) (map[int64]int64, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "err get db")
	}

	result := make(map[int64]int64, len(projectIDs))
	if len(projectIDs) == 0 {
		return result, nil
	}

	type winningQtyRow struct {
		ProjectId int64 `xorm:"project_id"`
		Qty       int64 `xorm:"qty"`
	}

	rows := make([]winningQtyRow, 0)
	if err = db.Table("share").
		Select("share.project_id, IFNULL(SUM(share.qty), 0) AS qty").
		Join("INNER", "lottery_winning", "lottery_winning.share_id = share.id AND lottery_winning.deleted IS NULL").
		In("share.project_id", projectIDs).
		Where("share.deleted IS NULL").
		In("share.invest_status", domain.REQUEST_INVEST, domain.SUCCESS_INVEST).
		GroupBy("share.project_id").
		Find(&rows); err != nil {
		return nil, errors.Wrap(err, "err get winning qty totals by project id")
	}

	for _, row := range rows {
		result[row.ProjectId] = row.Qty
	}

	return result, nil
}
