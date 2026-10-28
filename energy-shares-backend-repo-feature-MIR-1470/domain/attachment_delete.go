package domain

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

// MIR-1403: 添付(attachment)の削除指定(フォームの attachments_delete[N])を、その対象(QnA・回答・ファンド・
// 資金調達相談)の添付だけに絞るための部品。添付の表は規約・方針などと共用で、IDは連番なので、受け取った
// IDをそのまま消すと、他人の添付やサイトの規約の書類まで消せてしまう。

// DeleteAttachmentFileFunc は、添付のS3のファイルを消す。テストで差し替えるために変数にしている。
var DeleteAttachmentFileFunc = func(ctx context.Context, a *Attachment) error {
	return s3utils.DeleteFile(ctx, s3utils.BUCKET_PUBLIC_NAME, a)
}

// ParseAttachmentDeleteIds は、フォームの attachments_delete[N] の値をIDに直す。数字でない値が1つでもあれば
// エラーを返す(途中まで消してから失敗しないよう、書き換えの前に呼ぶ)。
func ParseAttachmentDeleteIds(values map[string][]string) ([]int64, error) {
	ids := make([]int64, 0)
	for key, vs := range values {
		if !strings.Contains(key, "attachments_delete") {
			continue
		}
		for _, v := range vs {
			id, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid attachments_delete value: %w", err)
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// ClassifyAttachmentDeletes は、削除の指定を、対象の添付(owned。論理削除済みを含む)と照らして分ける。
// toDelete は対象の有効な添付で、重複は1回だけ。対象の論理削除済みの添付は、二重送信や別タブでの編集で
// 届くため黙って飛ばす。対象の添付でないID(他人の添付・規約などの添付・存在しないID)は unknown に入れる。
func ClassifyAttachmentDeletes(requested []int64, owned []Attachment) (toDelete []Attachment, unknown []int64) {
	ownedById := make(map[int64]Attachment, len(owned))
	for _, a := range owned {
		ownedById[a.Id] = a
	}
	seen := make(map[int64]bool, len(requested))
	for _, id := range requested {
		if seen[id] {
			continue
		}
		seen[id] = true
		a, ok := ownedById[id]
		if !ok {
			unknown = append(unknown, id)
			continue
		}
		if !a.Deleted.IsZero() {
			continue
		}
		toDelete = append(toDelete, a)
	}
	return toDelete, unknown
}

// FindAllByResourceIdUnscoped は、紐づけ先の添付を、論理削除済みも含めて件数の上限なしで返す。紐づけ先が
// 空なら、どの添付も対象とみなさない(紐づけ先が "" の行と取り違えないため)。
func (Attachment) FindAllByResourceIdUnscoped(ctx context.Context, resourceId string) ([]Attachment, error) {
	items := make([]Attachment, 0)
	if resourceId == "" {
		return items, nil
	}
	if err := factory.DB(ctx).Unscoped().Where("resource_id = ?", resourceId).Asc("id").Find(&items); err != nil {
		return nil, err
	}
	return items, nil
}

// DeleteByIdAndResourceId は、添付を論理削除する。紐づけ先も条件に入れ、判定を通った添付だけを消す。
func (Attachment) DeleteByIdAndResourceId(ctx context.Context, id int64, resourceId string) error {
	_, err := factory.DB(ctx).Where("id = ? AND resource_id = ?", id, resourceId).Delete(&Attachment{})
	return err
}

// DeleteAttachmentFilesAfterCommit は、論理削除した添付のS3のファイルを、DBの確定のあとに消す。DBの書き込みが
// 全部済んだ最後に呼ぶこと(確定後の仕組みが無い環境ではその場で実行されるため)。失敗やpanicはログに残すだけで、
// 応答は成功のままにする(表示からは消えていて、ファイルが残るだけ)。手で消せるよう、ログにはキーを出す。
// 元のファイル名は氏名を含みうるので出さない。
func DeleteAttachmentFilesAfterCommit(ctx context.Context, files []Attachment) {
	if len(files) == 0 {
		return
	}
	targets := make([]Attachment, len(files))
	copy(targets, files)
	sort.Slice(targets, func(i, j int) bool { return targets[i].Id < targets[j].Id })

	if err := factory.AfterCommit(ctx, func(afterCommitCtx context.Context) error {
		for i := range targets {
			deleteAttachmentFileSafely(afterCommitCtx, &targets[i])
		}
		return nil
	}); err != nil {
		factory.Logger(ctx).WithError(err).Error("failed to register attachment file deletes")
	}
}

func deleteAttachmentFileSafely(ctx context.Context, a *Attachment) {
	fields := logrus.Fields{"attachment_id": a.Id, "key": a.GetKey()}
	defer func() {
		if r := recover(); r != nil {
			factory.Logger(ctx).WithFields(fields).Errorf("panic while deleting attachment file: %v", r)
		}
	}()
	if err := DeleteAttachmentFileFunc(ctx, a); err != nil {
		factory.Logger(ctx).WithFields(fields).WithError(err).Error("failed to delete attachment file")
	}
}
