package controllers

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

// MIR-1403: 会員のQnAの更新で、他人の添付(規約・方針の書類を含む)をIDで消せない・紐づけ先を書き換えられないこと。

type qnaUpdateFixture struct {
	ctx       context.Context
	engine    *xorm.Engine
	qna       domain.Qna
	own       []domain.Attachment
	policy    domain.Attachment
	otherQna  domain.Attachment
	s3Deleted []string
	uploaded  []string
}

func setupQnaUpdate(t *testing.T, ownCount int) *qnaUpdateFixture {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync(new(domain.Qna), new(domain.Answer), new(domain.Attachment)))

	f := &qnaUpdateFixture{engine: engine}
	f.qna = domain.Qna{CommonId: "qna-own", UserId: 7, Title: "old", Name: "n", Email: "e@example.invalid", Content: "old body", Read: true, Status: domain.QNA_RECEIVED}
	_, err = engine.Insert(&f.qna)
	require.NoError(t, err)
	other := domain.Qna{CommonId: "qna-other", UserId: 8, Title: "other", Name: "o", Email: "o@example.invalid", Content: "x", Status: domain.QNA_RECEIVED}
	_, err = engine.Insert(&other)
	require.NoError(t, err)

	insert := func(resourceId, name string) domain.Attachment {
		a := domain.Attachment{FileName: name, OriginalName: name + ".pdf", FileExtension: ".pdf", FilePath: "/attachment_file/", ResourceId: resourceId}
		_, err := engine.Insert(&a)
		require.NoError(t, err)
		return a
	}
	for i := 0; i < ownCount; i++ {
		f.own = append(f.own, insert("qna-own", "own"+strconv.Itoa(i)))
	}
	f.policy = insert("policy-terms", "terms")
	f.otherQna = insert("qna-other", "other")

	origDelete, origUpload := domain.DeleteAttachmentFileFunc, domain.UploadAttachmentFileFunc
	t.Cleanup(func() { domain.DeleteAttachmentFileFunc, domain.UploadAttachmentFileFunc = origDelete, origUpload })
	domain.DeleteAttachmentFileFunc = func(_ context.Context, a *domain.Attachment) error {
		f.s3Deleted = append(f.s3Deleted, a.GetKey())
		return nil
	}
	domain.UploadAttachmentFileFunc = func(_ context.Context, _ string, obj s3utils.S3Object, fh *multipart.FileHeader) (string, error) {
		f.uploaded = append(f.uploaded, fh.Filename)
		return "https://cdn.example.invalid/" + obj.GetKey(), nil
	}

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	f.ctx, _ = factory.WithAfterCommit(ctx)
	return f
}

// callUpdate は、画面と同じく multipart で PUT /api/qna/:id を呼ぶ。削除の指定は、画面が実際に送るキー
// attachments_delete[N] で送る。
func (f *qnaUpdateFixture) callUpdate(t *testing.T, userId int64, qnaId int64, fields map[string]string, deletes []string, files []string) (*httptest.ResponseRecorder, error) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}
	for i, d := range deletes {
		require.NoError(t, w.WriteField("attachments_delete["+strconv.Itoa(i)+"]", d))
	}
	for i, name := range files {
		fw, err := w.CreateFormFile("attachments["+strconv.Itoa(i)+"]", name)
		require.NoError(t, err)
		_, err = fw.Write([]byte("%PDF-1.4"))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())

	e := echo.New()
	e.Validator = validator.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPut, "/api/qna/"+strconv.FormatInt(qnaId, 10), &body)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	req = req.WithContext(f.ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.FormatInt(qnaId, 10))
	c.Set("user", jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": strconv.FormatInt(userId, 10), "name": "n", "email": "e@example.invalid", "role": []string{"ROLE_USER"},
	}))
	return rec, QnaApiController{}.Update(c)
}

func (f *qnaUpdateFixture) runAfterCommit(t *testing.T) {
	t.Helper()
	require.NoError(t, factory.RunAfterCommit(f.ctx, f.ctx))
}

func (f *qnaUpdateFixture) attachment(t *testing.T, id int64) domain.Attachment {
	t.Helper()
	var a domain.Attachment
	has, err := f.engine.Unscoped().ID(id).Get(&a)
	require.NoError(t, err)
	require.True(t, has)
	return a
}

func (f *qnaUpdateFixture) reloadQna(t *testing.T) domain.Qna {
	t.Helper()
	var q domain.Qna
	has, err := f.engine.ID(f.qna.Id).Get(&q)
	require.NoError(t, err)
	require.True(t, has)
	return q
}

func assertParameterError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var fe *factory.Error
	require.True(t, errors.As(err, &fe), "error type = %T", err)
	assert.Equal(t, factory.ErrorParameter.Code, fe.Code)
}

func TestQnaUpdate_RejectsAttachmentsNotOwnedByQna(t *testing.T) {
	body := map[string]string{"title": "new", "content": "new body", "common_id": "qna-own"}
	tests := []struct {
		name    string
		deletes func(f *qnaUpdateFixture) []string
	}{
		{name: "規約の添付のID", deletes: func(f *qnaUpdateFixture) []string { return []string{strconv.FormatInt(f.policy.Id, 10)} }},
		{name: "他人のQnAの添付のID", deletes: func(f *qnaUpdateFixture) []string { return []string{strconv.FormatInt(f.otherQna.Id, 10)} }},
		{name: "存在しないID", deletes: func(f *qnaUpdateFixture) []string { return []string{"999999"} }},
		{name: "数字でない値(1件目の自分の添付も消さない)", deletes: func(f *qnaUpdateFixture) []string {
			return []string{strconv.FormatInt(f.own[0].Id, 10), "abc"}
		}},
		{name: "自分の添付と規約の添付の混在(自分の添付も消さない)", deletes: func(f *qnaUpdateFixture) []string {
			return []string{strconv.FormatInt(f.own[0].Id, 10), strconv.FormatInt(f.policy.Id, 10)}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupQnaUpdate(t, 1)

			_, err := f.callUpdate(t, 7, f.qna.Id, body, tt.deletes(f), []string{"new.pdf"})

			assertParameterError(t, err)
			f.runAfterCommit(t)
			assert.Empty(t, f.s3Deleted, "S3のファイルを消さない")
			assert.Empty(t, f.uploaded, "新しい添付も上げない")
			assert.True(t, f.attachment(t, f.policy.Id).Deleted.IsZero(), "規約の添付は残る")
			assert.True(t, f.attachment(t, f.otherQna.Id).Deleted.IsZero(), "他人の添付は残る")
			assert.True(t, f.attachment(t, f.own[0].Id).Deleted.IsZero(), "自分の添付も消えない")
			q := f.reloadQna(t)
			assert.Equal(t, "old", q.Title, "QnAの本文も変わらない")
			assert.True(t, q.Read)
		})
	}
}

func TestQnaUpdate_DeletesOwnAttachmentAfterCommit(t *testing.T) {
	f := setupQnaUpdate(t, 2)

	rec, err := f.callUpdate(t, 7, f.qna.Id, map[string]string{"title": "new", "content": "new body", "common_id": "qna-own"},
		[]string{strconv.FormatInt(f.own[1].Id, 10)}, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, f.attachment(t, f.own[1].Id).Deleted.IsZero(), "自分の添付は論理削除される")
	assert.True(t, f.attachment(t, f.own[0].Id).Deleted.IsZero())
	assert.Empty(t, f.s3Deleted, "ハンドラが終わった時点では、S3はまだ消さない")

	f.runAfterCommit(t)
	assert.Equal(t, []string{"/attachment_file/own1.pdf"}, f.s3Deleted, "確定のあとで1回だけ消す")
}

// 紐づけ先(common_id)に規約のresource_idを入れて送っても、QnAの紐づけ先は変わらず、新しい添付は自分のQnAに付く。
func TestQnaUpdate_IgnoresCommonIdFromRequest(t *testing.T) {
	f := setupQnaUpdate(t, 0)

	_, err := f.callUpdate(t, 7, f.qna.Id, map[string]string{"title": "new", "content": "new body", "common_id": "policy-terms"}, nil, []string{"inject.pdf"})

	require.NoError(t, err)
	assert.Equal(t, "qna-own", f.reloadQna(t).CommonId)
	var added []domain.Attachment
	require.NoError(t, f.engine.Where("original_name = ?", "inject.pdf").Find(&added))
	require.Len(t, added, 1)
	assert.Equal(t, "qna-own", added[0].ResourceId, "新しい添付は自分のQnAに付く")
	var policyAttachments int64
	policyAttachments, err = f.engine.Where("resource_id = ?", "policy-terms").Count(&domain.Attachment{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), policyAttachments, "規約のページには何も足されない")
}

// いちばん多い操作(本文だけの編集)。未読に戻る。
func TestQnaUpdate_BodyOnly(t *testing.T) {
	f := setupQnaUpdate(t, 1)

	rec, err := f.callUpdate(t, 7, f.qna.Id, map[string]string{"title": "new", "content": "new body", "common_id": "qna-own"}, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	q := f.reloadQna(t)
	assert.Equal(t, "new", q.Title)
	assert.Equal(t, "new body", q.Content)
	assert.False(t, q.Read, "編集すると未読に戻る(今までの動き)")
	assert.Equal(t, int64(7), q.UserId)
	f.runAfterCommit(t)
	assert.Empty(t, f.s3Deleted)
}

// 二重送信・別タブ: すでに論理削除した自分の添付のIDがもう一度届いても、400にせず飛ばす。
func TestQnaUpdate_SkipsOwnAlreadyDeletedAttachment(t *testing.T) {
	f := setupQnaUpdate(t, 1)
	_, err := f.engine.ID(f.own[0].Id).Delete(&domain.Attachment{})
	require.NoError(t, err)

	_, err = f.callUpdate(t, 7, f.qna.Id, map[string]string{"title": "new", "content": "new body"}, []string{strconv.FormatInt(f.own[0].Id, 10)}, nil)

	require.NoError(t, err)
	f.runAfterCommit(t)
	assert.Empty(t, f.s3Deleted, "二度消さない")
}

// 自分の添付が11件以上あっても、上限なしで取れていて消せる。
func TestQnaUpdate_ManyOwnAttachments(t *testing.T) {
	f := setupQnaUpdate(t, 12)

	_, err := f.callUpdate(t, 7, f.qna.Id, map[string]string{"title": "new", "content": "new body"}, []string{strconv.FormatInt(f.own[11].Id, 10)}, nil)

	require.NoError(t, err)
	assert.False(t, f.attachment(t, f.own[11].Id).Deleted.IsZero())
}

// 判定のあとで新しい添付の保存に失敗したら、S3の削除を登録しない(DBは巻き戻るので、ファイルだけ消えないように)。
func TestQnaUpdate_UploadFailureDoesNotDeleteFiles(t *testing.T) {
	f := setupQnaUpdate(t, 1)
	domain.UploadAttachmentFileFunc = func(context.Context, string, s3utils.S3Object, *multipart.FileHeader) (string, error) {
		return "", errors.New("upload failed")
	}

	_, err := f.callUpdate(t, 7, f.qna.Id, map[string]string{"title": "new", "content": "new body"}, []string{strconv.FormatInt(f.own[0].Id, 10)}, []string{"new.pdf"})

	require.Error(t, err)
	f.runAfterCommit(t)
	assert.Empty(t, f.s3Deleted)
}

// 確定後のS3の削除がpanicしても、応答は成功のまま壊れない。
func TestQnaUpdate_AfterCommitPanicDoesNotBreakResponse(t *testing.T) {
	f := setupQnaUpdate(t, 1)
	domain.DeleteAttachmentFileFunc = func(context.Context, *domain.Attachment) error { panic("boom") }

	rec, err := f.callUpdate(t, 7, f.qna.Id, map[string]string{"title": "new", "content": "new body"}, []string{strconv.FormatInt(f.own[0].Id, 10)}, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotPanics(t, func() { f.runAfterCommit(t) })
}

// 他人のQnAは更新できない(今までどおり)。
func TestQnaUpdate_OtherUsersQna(t *testing.T) {
	f := setupQnaUpdate(t, 1)

	_, err := f.callUpdate(t, 8, f.qna.Id, map[string]string{"title": "new", "content": "new body"}, []string{strconv.FormatInt(f.own[0].Id, 10)}, nil)

	require.Error(t, err)
	assert.True(t, f.attachment(t, f.own[0].Id).Deleted.IsZero())
	assert.Equal(t, "old", f.reloadQna(t).Title)
}

// 紐づけ先(common_id)が空のQnAに、削除やファイルが来たら400(紐づけ先 "" の行方不明の添付を作らない)。
func TestQnaUpdate_EmptyCommonId(t *testing.T) {
	f := setupQnaUpdate(t, 0)
	_, err := f.engine.Exec("UPDATE qna SET common_id = '' WHERE id = ?", f.qna.Id)
	require.NoError(t, err)

	_, err = f.callUpdate(t, 7, f.qna.Id, map[string]string{"title": "new", "content": "new body"}, nil, []string{"new.pdf"})
	assertParameterError(t, err)
	assert.Empty(t, f.uploaded)

	_, err = f.callUpdate(t, 7, f.qna.Id, map[string]string{"title": "new", "content": "new body"}, []string{strconv.FormatInt(f.policy.Id, 10)}, nil)
	assertParameterError(t, err)
	assert.True(t, f.attachment(t, f.policy.Id).Deleted.IsZero())

	_, err = f.callUpdate(t, 7, f.qna.Id, map[string]string{"title": "new", "content": "new body"}, nil, nil)
	require.NoError(t, err, "本文だけの編集は通る")
}
