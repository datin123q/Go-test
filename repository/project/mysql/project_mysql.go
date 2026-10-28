package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type projectMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.ProjectRepository {
	return &projectMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (p projectMysqlRepository) GetById(ctx context.Context, id int64) (domain.Project, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Project{}, errors.Wrap(err, "failed to get db")
	}

	var project domain.Project

	has, err := orm.Table(db, domain.Project{}).Id(id).Select(domain.PROJECT_DETAIL_COLUMN_LIST).Get(&project)
	if err != nil {
		return domain.Project{}, errors.Wrap(err, "failed to get project by id")
	}

	if !has {
		return domain.Project{}, errors.Wrap(cdomain.ErrItemNotFound, "project not found")
	}

	return project, nil
}

func (p projectMysqlRepository) LockById(ctx context.Context, id int64) (domain.Project, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Project{}, errors.Wrap(err, "failed to get db")
	}

	var project domain.Project

	// project.* だとRaisedAmount等(xorm:"<-"、shareへの相関サブクエリで埋まる列)が
	// 埋まらずゼロ値のまま返ってしまうため、GetByIdと同じPROJECT_DETAIL_COLUMN_LISTを使う。
	has, err := db.SQL(
		"SELECT "+domain.PROJECT_DETAIL_COLUMN_LIST+" FROM project WHERE project.id = ? AND project.deleted IS NULL FOR UPDATE",
		id,
	).Get(&project)
	if err != nil {
		return domain.Project{}, errors.Wrap(err, "failed to lock project by id")
	}

	if !has {
		return domain.Project{}, errors.Wrap(cdomain.ErrItemNotFound, "project not found")
	}

	return project, nil
}

// TryLockById はLockByIdと違い、他のトランザクションがロック中なら待たずに
// ErrItemConflictを返す(FOR UPDATE NOWAIT)。ECSのタスク入れ替えなどでロックを
// 持ったまま復旧できなくなる状態を避けるため、実行中フラグは持たず、ロック自体を
// 排他制御の唯一の手段にしている(MIR-1204、2026-09-08河内さん判断)。
func (p projectMysqlRepository) TryLockById(ctx context.Context, id int64) (domain.Project, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Project{}, errors.Wrap(err, "failed to get db")
	}

	var project domain.Project

	// project.* だとRaisedAmount等(xorm:"<-"、shareへの相関サブクエリで埋まる列)が
	// 埋まらずゼロ値のまま返ってしまうため、GetByIdと同じPROJECT_DETAIL_COLUMN_LISTを使う。
	// これが無いと、ExecuteInvestmentの「RaisedAmount < request.Amount」判定が常に
	// trueになり、出資実行が本番で必ず失敗する(MIR-1204、code-reviewで検出)。
	has, err := db.SQL(
		"SELECT "+domain.PROJECT_DETAIL_COLUMN_LIST+" FROM project WHERE project.id = ? AND project.deleted IS NULL FOR UPDATE NOWAIT",
		id,
	).Get(&project)
	if err != nil {
		if factory.IsLockNotAvailableError(err) {
			return domain.Project{}, errors.Wrap(cdomain.ErrItemConflict, "failed to lock project: locked by another request")
		}
		return domain.Project{}, errors.Wrap(err, "failed to try-lock project by id")
	}

	if !has {
		return domain.Project{}, errors.Wrap(cdomain.ErrItemNotFound, "project not found")
	}

	return project, nil
}

func (p projectMysqlRepository) ReduceFinalFundingAmount(ctx context.Context, projectID int64, amount int64) error {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Exec(
		"UPDATE project SET final_funding_amount = final_funding_amount - ? WHERE id = ?",
		amount,
		projectID,
	); err != nil {
		return errors.Wrap(err, "err reduce final funding amount")
	}

	return nil
}

func (p projectMysqlRepository) Update(ctx context.Context, project domain.Project) error {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	// project_status はここでは書かない(MIR-1314)。呼び出し元(メンテナンス・
	// share_svc)はどちらもGetByIdで読んだ値をそのまま渡しているだけで状態を
	// 変える意図は無いが、書いたままにすると並行して動くジョブや状態変更API
	// による変更を、この保存で古い値に戻してしまう。
	if _, err = db.ID(project.Id).
		Cols("offering_type", "is_wait_invest", "end_at"). // 今後更新が必要な列を追加します。
		Update(&project); err != nil {
		return errors.Wrap(err, "err update project")
	}

	return nil
}

func (p projectMysqlRepository) AddFinalFundingAmount(ctx context.Context, projectId int64, amount int64) error {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Exec(
		"UPDATE project SET final_funding_amount = COALESCE(final_funding_amount, 0) + ? WHERE id = ?",
		amount,
		projectId,
	); err != nil {
		return errors.Wrap(err, "err add final funding amount")
	}

	return nil
}
