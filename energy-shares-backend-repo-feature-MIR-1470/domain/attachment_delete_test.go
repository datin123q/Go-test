package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func TestParseAttachmentDeleteIds(t *testing.T) {
	tests := []struct {
		name    string
		values  map[string][]string
		want    []int64
		wantErr bool
	}{
		{name: "指定なし", values: map[string][]string{"title": {"t"}}, want: []int64{}},
		{name: "画面が送るキー(attachments_delete[N])", values: map[string][]string{"attachments_delete[0]": {"12"}, "title": {"t"}}, want: []int64{12}},
		{name: "1つのキーに複数の値", values: map[string][]string{"attachments_delete": {"3", "4"}}, want: []int64{3, 4}},
		{name: "数字でない値が1つでもあればエラー", values: map[string][]string{"attachments_delete[0]": {"5"}, "attachments_delete[1]": {"abc"}}, wantErr: true},
		{name: "空文字もエラー", values: map[string][]string{"attachments_delete[0]": {""}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAttachmentDeleteIds(tt.values)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestClassifyAttachmentDeletes(t *testing.T) {
	deleted := jsonutils.JSONTime(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	own1 := Attachment{Id: 1, ResourceId: "qna-a"}
	own2 := Attachment{Id: 2, ResourceId: "qna-a"}
	ownDeleted := Attachment{Id: 3, ResourceId: "qna-a", Deleted: deleted}
	owned := []Attachment{own1, own2, ownDeleted}

	tests := []struct {
		name        string
		requested   []int64
		wantDelete  []int64
		wantUnknown []int64
	}{
		{name: "空", requested: nil},
		{name: "自分の有効な添付は消す", requested: []int64{1, 2}, wantDelete: []int64{1, 2}},
		{name: "自分の論理削除済みの添付は黙って飛ばす(二重送信・別タブ)", requested: []int64{3}},
		{name: "他人の添付・規約の添付・存在しないIDは unknown", requested: []int64{99, 100, 101}, wantUnknown: []int64{99, 100, 101}},
		{name: "同じIDの重複は1回だけ", requested: []int64{1, 1, 99, 99}, wantDelete: []int64{1}, wantUnknown: []int64{99}},
		{name: "自分の添付と他人のIDの混在は、両方に分かれる(呼び出し側で判断)", requested: []int64{1, 99}, wantDelete: []int64{1}, wantUnknown: []int64{99}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toDelete, unknown := ClassifyAttachmentDeletes(tt.requested, owned)
			gotIds := make([]int64, 0, len(toDelete))
			for _, a := range toDelete {
				gotIds = append(gotIds, a.Id)
			}
			assert.ElementsMatch(t, tt.wantDelete, gotIds)
			assert.ElementsMatch(t, tt.wantUnknown, unknown)
		})
	}

	t.Run("紐づけ先が空(対象の添付が無い)なら、どのIDも unknown", func(t *testing.T) {
		toDelete, unknown := ClassifyAttachmentDeletes([]int64{1}, nil)
		assert.Empty(t, toDelete)
		assert.Equal(t, []int64{1}, unknown)
	})
}

func newAttachmentTestCtx(t *testing.T) (context.Context, *xorm.Engine) {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync(new(Attachment), new(Qna)))
	return context.WithValue(context.Background(), factory.ConextDBType, engine), engine
}

func insertAttachment(t *testing.T, engine *xorm.Engine, resourceId string, name string) Attachment {
	t.Helper()
	a := Attachment{FileName: name, OriginalName: name + ".pdf", FileExtension: ".pdf", FilePath: "/attachment_file/", ResourceId: resourceId}
	_, err := engine.Insert(&a)
	require.NoError(t, err)
	return a
}

func TestAttachment_FindAllByResourceIdUnscoped(t *testing.T) {
	ctx, engine := newAttachmentTestCtx(t)
	var ids []int64
	for i := 0; i < 12; i++ { // 今のFindByResourceIdの上限(呼び出し側で10)を超える件数
		ids = append(ids, insertAttachment(t, engine, "qna-a", "f"+string(rune('a'+i))).Id)
	}
	insertAttachment(t, engine, "policy-x", "policy")
	_, err := engine.ID(ids[0]).Delete(&Attachment{}) // 論理削除
	require.NoError(t, err)

	got, err := Attachment{}.FindAllByResourceIdUnscoped(ctx, "qna-a")
	require.NoError(t, err)
	require.Len(t, got, 12, "論理削除済みを含み、件数の上限なしで取る")
	assert.False(t, got[0].Deleted.IsZero(), "論理削除済みが区別できる")
	for _, a := range got {
		assert.Equal(t, "qna-a", a.ResourceId)
	}

	empty, err := Attachment{}.FindAllByResourceIdUnscoped(ctx, "")
	require.NoError(t, err)
	assert.Empty(t, empty, "紐づけ先が空なら何も対象にしない")
}

func TestAttachment_DeleteByIdAndResourceId(t *testing.T) {
	ctx, engine := newAttachmentTestCtx(t)
	own := insertAttachment(t, engine, "qna-a", "own")
	policy := insertAttachment(t, engine, "policy-x", "policy")

	require.NoError(t, Attachment{}.DeleteByIdAndResourceId(ctx, policy.Id, "qna-a"))
	require.NoError(t, Attachment{}.DeleteByIdAndResourceId(ctx, own.Id, "qna-a"))

	var rows []Attachment
	require.NoError(t, engine.Unscoped().Asc("id").Find(&rows))
	require.Len(t, rows, 2)
	assert.False(t, rows[0].Deleted.IsZero(), "対象の添付は論理削除される")
	assert.True(t, rows[1].Deleted.IsZero(), "紐づけ先が違う添付は、IDを渡しても消えない")
}

func TestDeleteAttachmentFilesAfterCommit(t *testing.T) {
	orig := DeleteAttachmentFileFunc
	t.Cleanup(func() { DeleteAttachmentFileFunc = orig })

	t.Run("確定のあとにだけ消す", func(t *testing.T) {
		var keys []string
		DeleteAttachmentFileFunc = func(_ context.Context, a *Attachment) error {
			keys = append(keys, a.GetKey())
			return nil
		}
		ctx, _ := factory.WithAfterCommit(context.Background())
		files := []Attachment{{Id: 2, FilePath: "/p/", FileName: "b", FileExtension: ".pdf"}, {Id: 1, FilePath: "/p/", FileName: "a", FileExtension: ".pdf"}}

		DeleteAttachmentFilesAfterCommit(ctx, files)
		files[0].FileName = "changed" // 呼び出し側があとで書き換えても、登録時の内容で消す
		assert.Empty(t, keys, "確定前には消さない")

		require.NoError(t, factory.RunAfterCommit(ctx, context.Background()))
		assert.Equal(t, []string{"/p/a.pdf", "/p/b.pdf"}, keys)
	})

	t.Run("S3の失敗やpanicは確定後の処理を止めず、ほかのファイルも消す", func(t *testing.T) {
		var called []int64
		DeleteAttachmentFileFunc = func(_ context.Context, a *Attachment) error {
			called = append(called, a.Id)
			if a.Id == 1 {
				panic("boom")
			}
			return errors.New("s3 error")
		}
		ctx, _ := factory.WithAfterCommit(context.Background())
		DeleteAttachmentFilesAfterCommit(ctx, []Attachment{{Id: 1}, {Id: 2}})

		require.NotPanics(t, func() { require.NoError(t, factory.RunAfterCommit(ctx, context.Background())) })
		assert.Equal(t, []int64{1, 2}, called)
	})

	t.Run("対象が無ければ何も登録しない", func(t *testing.T) {
		called := 0
		DeleteAttachmentFileFunc = func(context.Context, *Attachment) error { called++; return nil }
		ctx, _ := factory.WithAfterCommit(context.Background())
		DeleteAttachmentFilesAfterCommit(ctx, nil)
		require.NoError(t, factory.RunAfterCommit(ctx, context.Background()))
		assert.Equal(t, 0, called)
	})
}

// 会員の編集では、件名・本文・行為者・未読だけが変わり、紐づけ先などは変わらない。
func TestQna_UpdateByUser(t *testing.T) {
	ctx, engine := newAttachmentTestCtx(t)
	before := Qna{CommonId: "qna-a", UserId: 7, Title: "old", Name: "n", Email: "e@example.invalid", Content: "old body", AnswerCount: 2, Read: true, Status: QNA_RESPONSE_COMPLETED}
	_, err := engine.Insert(&before)
	require.NoError(t, err)

	q := &Qna{Id: before.Id, CommonId: "policy-x", UserId: 99, Title: "new", Content: "new body", AnswerCount: 0, Status: QNA_RECEIVED, Actor: ChangelogActor{Id: 7, Name: "n", Email: "e@example.invalid"}}
	_, err = q.UpdateByUser(ctx)
	require.NoError(t, err)

	var after Qna
	has, err := engine.ID(before.Id).Get(&after)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, "new", after.Title)
	assert.Equal(t, "new body", after.Content)
	assert.False(t, after.Read, "編集すると未読に戻る")
	assert.Equal(t, int64(7), after.Actor.Id)
	assert.Equal(t, "qna-a", after.CommonId, "紐づけ先は書き換えない")
	assert.Equal(t, int64(7), after.UserId)
	assert.Equal(t, int64(2), after.AnswerCount)
	assert.Equal(t, QNA_RESPONSE_COMPLETED, after.Status)
}
