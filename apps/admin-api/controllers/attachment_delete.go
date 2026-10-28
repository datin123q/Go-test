package controllers

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// deleteOwnedAttachments は、フォームの attachments_delete[N] のうち、紐づけ先(resourceId。更新前の対象の
// common_idなどDBの値)の添付だけを論理削除し、S3から消すべき添付を返す(MIR-1403)。管理画面の操作は止めない
// よう、対象の添付でないID・対象の論理削除済みの添付(二重送信)は、IDだけをログに残して無視する
// (project_file_delete の前例に合わせた。元のファイル名は氏名を含みうるので出さない)。S3の削除は、呼び出し元が
// DBの書き込みを全部済ませた最後に domain.DeleteAttachmentFilesAfterCommit へ渡す。
func deleteOwnedAttachments(ctx context.Context, values map[string][]string, resourceId string, fields logrus.Fields) ([]domain.Attachment, error) {
	ids, err := domain.ParseAttachmentDeleteIds(values)
	if err != nil {
		return nil, factory.ErrorParameter.New(err)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	owned, err := domain.Attachment{}.FindAllByResourceIdUnscoped(ctx, resourceId)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	toDelete, unknown := domain.ClassifyAttachmentDeletes(ids, owned)
	if len(unknown) > 0 {
		factory.Logger(ctx).WithFields(fields).WithField("attachment_ids", unknown).
			Warn("ignored attachments_delete that do not belong to the target")
	}
	for _, a := range toDelete {
		if err := (domain.Attachment{}).DeleteByIdAndResourceId(ctx, a.Id, resourceId); err != nil {
			return nil, factory.ErrorDB.New(err)
		}
	}
	return toDelete, nil
}

// classifyUserFileDeletes は、削除の指定のうち、対象の会員(userId)の書類だけを消す対象に分ける。他の会員の
// 書類・存在しないID・論理削除済み(foundに含まれない)は ignored に入れる。重複は1回だけ扱う。
func classifyUserFileDeletes(requested []int64, found []domain.UserFile, userId int64) (toDelete []int64, ignored []int64) {
	ownerById := make(map[int64]int64, len(found))
	for _, f := range found {
		ownerById[f.Id] = f.UserId
	}
	seen := make(map[int64]bool, len(requested))
	for _, id := range requested {
		if seen[id] {
			continue
		}
		seen[id] = true
		if owner, ok := ownerById[id]; ok && owner == userId {
			toDelete = append(toDelete, id)
			continue
		}
		ignored = append(ignored, id)
	}
	return toDelete, ignored
}

// deleteOwnedUserFiles は、フォームの attachments_delete[N] のうち、対象の会員(userId。DBから取った会員のID)の
// 書類(本人確認書類など)だけを論理削除する(MIR-1403)。それ以外はIDだけをログに残して無視する。
func deleteOwnedUserFiles(ctx context.Context, values map[string][]string, userId int64, fields logrus.Fields) error {
	ids, err := domain.ParseAttachmentDeleteIds(values)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if len(ids) == 0 {
		return nil
	}
	found, err := domain.UserFile{}.FindByIds(ctx, ids)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	toDelete, ignored := classifyUserFileDeletes(ids, found, userId)
	if len(ignored) > 0 {
		factory.Logger(ctx).WithFields(fields).WithField("user_file_ids", ignored).
			Warn("ignored attachments_delete that do not belong to the user")
	}
	for _, id := range toDelete {
		if err := (domain.UserFile{}).Delete(ctx, id); err != nil {
			return factory.ErrorDB.New(err)
		}
	}
	return nil
}
