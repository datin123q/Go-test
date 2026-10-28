package controllers

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

// MIR-1413: 反社チェックの書類の削除(antisocial_check_file_delete[N])を、対象の会員の書類だけに絞る。以前は指定された
// IDをそのまま消し、別の会員の反社チェックの証跡も消せたうえ、S3のファイルをDBの確定前に消していた。

// deleteAntisocialCheckFileFunc は、反社チェックの書類のS3のファイルを消す。テストで差し替えるために変数にしている。
var deleteAntisocialCheckFileFunc = func(ctx context.Context, f *domain.AntisocialCheckFile) error {
	return s3utils.DeleteFile(ctx, s3utils.BUCKET_PRIVATE_NAME, f)
}

// parseAntisocialCheckFileDeleteIds は、フォームの antisocial_check_file_delete[N] の値をIDに直す。数字でない値が1つでも
// あればエラーを返す(書き換えやメールの送信より前に呼ぶ)。
func parseAntisocialCheckFileDeleteIds(values map[string][]string) ([]int64, error) {
	ids := make([]int64, 0)
	for key, vs := range values {
		if !strings.Contains(key, "antisocial_check_file_delete") {
			continue
		}
		for _, v := range vs {
			id, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid antisocial_check_file_delete value: %w", err)
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// classifyAntisocialCheckFileDeletes は、削除の指定のうち、対象の会員(checkUserId)の書類だけを消す対象に分ける。他の
// 会員の書類・存在しないIDは ignored に入れる。対象の会員の論理削除済みの書類は、画面からの送り直しで届くので、
// どちらにも入れずに黙って飛ばす(MIR-1403の添付と同じ扱い)。重複は1回だけ扱う。
func classifyAntisocialCheckFileDeletes(requested []int64, found []domain.AntisocialCheckFile, checkUserId int64) (toDelete []domain.AntisocialCheckFile, ignored []int64) {
	byId := make(map[int64]domain.AntisocialCheckFile, len(found))
	for _, f := range found {
		byId[f.Id] = f
	}
	seen := make(map[int64]bool, len(requested))
	for _, id := range requested {
		if seen[id] {
			continue
		}
		seen[id] = true
		if f, ok := byId[id]; ok && f.CheckId == checkUserId {
			if f.Deleted.IsZero() {
				toDelete = append(toDelete, f)
			}
			continue
		}
		ignored = append(ignored, id)
	}
	return toDelete, ignored
}

// deleteOwnedAntisocialCheckFiles は、対象の会員の書類だけを論理削除し、S3から消すべき書類を返す。対象の会員の論理
// 削除済みは黙って飛ばし、それ以外はIDだけをログに残して無視する(管理画面の操作は止めない。元のファイル名は氏名を
// 含みうるので出さない)。
func deleteOwnedAntisocialCheckFiles(ctx context.Context, ids []int64, checkUserId int64, fields logrus.Fields) ([]domain.AntisocialCheckFile, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	found, err := domain.AntisocialCheckFile{}.FindByIdsUnscoped(ctx, ids)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	toDelete, ignored := classifyAntisocialCheckFileDeletes(ids, found, checkUserId)
	if len(ignored) > 0 {
		factory.Logger(ctx).WithFields(fields).WithField("antisocial_check_file_ids", ignored).
			Warn("ignored antisocial_check_file_delete that do not belong to the user")
	}
	for _, f := range toDelete {
		if err := (domain.AntisocialCheckFile{}).DeleteByIdAndCheckId(ctx, f.Id, checkUserId); err != nil {
			return nil, factory.ErrorDB.New(err)
		}
	}
	return toDelete, nil
}

// deleteAntisocialCheckFilesAfterCommit は、論理削除した書類のS3のファイルを、DBの確定のあとに消す。DBの書き込みが
// 全部済んだ最後に呼ぶこと(確定後の仕組みが無い環境ではその場で実行されるため)。失敗やpanicはIDとキーだけをログに
// 残し、応答は成功のままにする(画面からは消えていて、ファイルが残るだけ)。
func deleteAntisocialCheckFilesAfterCommit(ctx context.Context, files []domain.AntisocialCheckFile) {
	if len(files) == 0 {
		return
	}
	targets := make([]domain.AntisocialCheckFile, len(files))
	copy(targets, files)
	sort.Slice(targets, func(i, j int) bool { return targets[i].Id < targets[j].Id })

	if err := factory.AfterCommit(ctx, func(afterCommitCtx context.Context) error {
		for i := range targets {
			deleteAntisocialCheckFileSafely(afterCommitCtx, &targets[i])
		}
		return nil
	}); err != nil {
		factory.Logger(ctx).WithError(err).Error("failed to register antisocial check file deletes")
	}
}

func deleteAntisocialCheckFileSafely(ctx context.Context, f *domain.AntisocialCheckFile) {
	fields := logrus.Fields{"antisocial_check_file_id": f.Id, "key": f.GetKey()}
	defer func() {
		if r := recover(); r != nil {
			factory.Logger(ctx).WithFields(fields).Errorf("panic while deleting antisocial check file: %v", r)
		}
	}()
	if err := deleteAntisocialCheckFileFunc(ctx, f); err != nil {
		factory.Logger(ctx).WithFields(fields).WithError(err).Error("failed to delete antisocial check file")
	}
}
