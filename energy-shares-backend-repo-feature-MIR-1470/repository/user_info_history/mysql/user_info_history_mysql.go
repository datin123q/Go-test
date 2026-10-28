package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type userInfoHistoryMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.UserInfoHistoryRepository {
	return &userInfoHistoryMysqlRepository{dbFactory: dbFactory}
}

func (r userInfoHistoryMysqlRepository) RecordAuditStatusChange(ctx context.Context, entry cdomain.AuditStatusChangeEntry) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return err
	}

	history := domain.UserInfoHistory{
		UserId:        entry.UserId,
		ProcessUserId: entry.ActorId,
		SubstanceType: string(domain.USER_HISTORY_AUDIT),
		ProcessMethod: entry.ProcessMethod,
		Name:          entry.ActorName,
		Email:         entry.ActorEmail,
	}
	if _, err := db.Table("user_info_history").Insert(&history); err != nil {
		return errors.Wrap(err, "user_info_history の INSERT に失敗しました")
	}

	content := domain.UserChangedContent{
		InfoId:      history.Id,
		ContentType: string(domain.USER_CHANGED_CONTENT_AUDIT_STATUS),
		PrevContent: string(entry.PrevStatus),
		CurrContent: string(entry.CurrStatus),
	}
	if _, err := db.Table("user_changed_content").Insert(&content); err != nil {
		return errors.Wrap(err, "user_changed_content の INSERT に失敗しました")
	}

	return nil
}

func (r userInfoHistoryMysqlRepository) RecordContentChanges(ctx context.Context, entry cdomain.ContentChangeEntry) error {
	// 変更項目がない場合は履歴を作成しない
	if len(entry.Changes) == 0 {
		return nil
	}

	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return err
	}

	history := domain.UserInfoHistory{
		UserId:        entry.UserId,
		ProcessUserId: entry.ActorId,
		SubstanceType: string(domain.USER_HISTORY_CONTENTS),
		ProcessMethod: entry.ProcessMethod,
		Name:          entry.ActorName,
		Email:         entry.ActorEmail,
	}
	if _, err := db.Table("user_info_history").Insert(&history); err != nil {
		return errors.Wrap(err, "user_info_history の INSERT に失敗しました")
	}

	// 変更項目は1回の INSERT 文でまとめて登録する
	contents := make([]domain.UserChangedContent, 0, len(entry.Changes))
	for _, change := range entry.Changes {
		contents = append(contents, domain.UserChangedContent{
			InfoId:      history.Id,
			ContentType: string(change.ContentType),
			PrevContent: change.Prev,
			CurrContent: change.Curr,
		})
	}
	if _, err := db.Table("user_changed_content").Insert(&contents); err != nil {
		return errors.Wrap(err, "user_changed_content の INSERT に失敗しました")
	}

	return nil
}
