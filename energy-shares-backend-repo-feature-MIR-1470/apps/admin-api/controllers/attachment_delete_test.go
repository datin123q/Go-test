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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

// MIR-1403: 管理画面の添付・書類の削除も、対象(ファンド・資金調達相談・回答・会員)のものだけにする。

func newAdminAttachmentTestEngine(t *testing.T, beans ...interface{}) (*xorm.Engine, context.Context) {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync(beans...))
	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	ctx, _ = factory.WithAfterCommit(ctx)
	return engine, ctx
}

func insertTestAttachment(t *testing.T, engine *xorm.Engine, resourceId, name string) domain.Attachment {
	t.Helper()
	a := domain.Attachment{FileName: name, OriginalName: name + ".pdf", FileExtension: ".pdf", FilePath: "/attachment_file/", ResourceId: resourceId}
	_, err := engine.Insert(&a)
	require.NoError(t, err)
	return a
}

func isDeleted(t *testing.T, engine *xorm.Engine, bean interface{}, id int64) bool {
	t.Helper()
	n, err := engine.Unscoped().ID(id).Where("deleted IS NOT NULL AND deleted > '0001-01-02'").Count(bean)
	require.NoError(t, err)
	return n > 0
}

func TestDeleteOwnedAttachments(t *testing.T) {
	engine, ctx := newAdminAttachmentTestEngine(t, new(domain.Attachment))
	own := insertTestAttachment(t, engine, "project-a", "own")
	ownDeleted := insertTestAttachment(t, engine, "project-a", "own-deleted")
	_, err := engine.ID(ownDeleted.Id).Delete(&domain.Attachment{})
	require.NoError(t, err)
	policy := insertTestAttachment(t, engine, "policy-terms", "terms")

	values := map[string][]string{"attachments_delete[0]": {
		strconv.FormatInt(own.Id, 10), strconv.FormatInt(ownDeleted.Id, 10), strconv.FormatInt(policy.Id, 10), "999999",
	}}
	got, err := deleteOwnedAttachments(ctx, values, "project-a", logrus.Fields{"user": 1})

	require.NoError(t, err, "管理画面の操作は止めない(対象外はログを残して無視)")
	require.Len(t, got, 1)
	assert.Equal(t, own.Id, got[0].Id, "S3から消すのは対象の有効な添付だけ(二重送信の削除済みは含めない)")
	assert.True(t, isDeleted(t, engine, &domain.Attachment{}, own.Id))
	assert.False(t, isDeleted(t, engine, &domain.Attachment{}, policy.Id), "規約の添付は消えない")

	t.Run("数字でない値は400", func(t *testing.T) {
		_, err := deleteOwnedAttachments(ctx, map[string][]string{"attachments_delete[0]": {"abc"}}, "project-a", nil)
		var fe *factory.Error
		require.True(t, errors.As(err, &fe))
		assert.Equal(t, factory.ErrorParameter.Code, fe.Code)
	})
}

func TestClassifyUserFileDeletes(t *testing.T) {
	found := []domain.UserFile{{Id: 1, UserId: 10}, {Id: 2, UserId: 10}, {Id: 3, UserId: 20}}
	tests := []struct {
		name        string
		requested   []int64
		wantDelete  []int64
		wantIgnored []int64
	}{
		{name: "空", requested: nil},
		{name: "対象の会員の書類は消す", requested: []int64{1, 2}, wantDelete: []int64{1, 2}},
		{name: "他の会員の書類は無視", requested: []int64{3}, wantIgnored: []int64{3}},
		{name: "存在しない・論理削除済み(foundに無い)は無視", requested: []int64{99}, wantIgnored: []int64{99}},
		{name: "重複は1回だけ", requested: []int64{1, 1, 3, 3}, wantDelete: []int64{1}, wantIgnored: []int64{3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toDelete, ignored := classifyUserFileDeletes(tt.requested, found, 10)
			assert.ElementsMatch(t, tt.wantDelete, toDelete)
			assert.ElementsMatch(t, tt.wantIgnored, ignored)
		})
	}
}

func TestDeleteOwnedUserFiles(t *testing.T) {
	engine, ctx := newAdminAttachmentTestEngine(t, new(domain.UserFile))
	insert := func(userId int64, name string) domain.UserFile {
		f := domain.UserFile{UserId: userId, FileName: name, OriginalName: name, FileType: domain.IDENTIFICATION1_FRONT, UserType: domain.PRIVATE}
		_, err := engine.Insert(&f)
		require.NoError(t, err)
		return f
	}
	own := insert(10, "own")
	other := insert(20, "other")
	deleted := insert(10, "deleted")
	_, err := engine.ID(deleted.Id).Delete(&domain.UserFile{})
	require.NoError(t, err)

	values := map[string][]string{
		"attachments_delete[0]": {strconv.FormatInt(own.Id, 10)},
		"attachments_delete[1]": {strconv.FormatInt(other.Id, 10)},
		"attachments_delete[2]": {strconv.FormatInt(deleted.Id, 10)},
	}
	require.NotPanics(t, func() {
		require.NoError(t, deleteOwnedUserFiles(ctx, values, 10, logrus.Fields{"user_id": 10}))
	})
	assert.True(t, isDeleted(t, engine, &domain.UserFile{}, own.Id))
	assert.False(t, isDeleted(t, engine, &domain.UserFile{}, other.Id), "他の会員の本人確認書類などは消えない")
}

// 回答の更新: 回答の添付の削除・追加は、このQnAの回答(DBのcommon_id)に付く。他のQnAの回答は書き換えない。
func TestQnaUpdateAnswer_UsesAnswerCommonIdFromDB(t *testing.T) {
	engine, ctx := newAdminAttachmentTestEngine(t, new(domain.Qna), new(domain.Answer), new(domain.Attachment),
		new(domain.OperatorHistory), new(domain.QnaInfoHistory), new(domain.AdminUser), new(domain.Role))
	_, err := engine.Table("admin_user").Insert(&domain.AdminUser{Id: 1, Email: "admin@example.invalid", Name: "admin"})
	require.NoError(t, err)
	qna := domain.Qna{CommonId: "qna-a", UserId: 7, Title: "t", Name: "n", Email: "e@example.invalid", Content: "c", Status: domain.QNA_RESPONSE_COMPLETED}
	_, err = engine.Insert(&qna)
	require.NoError(t, err)
	otherQna := domain.Qna{CommonId: "qna-b", UserId: 8, Title: "t", Name: "n", Email: "o@example.invalid", Content: "c", Status: domain.QNA_RESPONSE_COMPLETED}
	_, err = engine.Insert(&otherQna)
	require.NoError(t, err)
	answer := domain.Answer{QnaId: qna.Id, CommonId: "answer-a", Answer: "old"}
	_, err = engine.Insert(&answer)
	require.NoError(t, err)
	otherAnswer := domain.Answer{QnaId: otherQna.Id, CommonId: "answer-b", Answer: "other"}
	_, err = engine.Insert(&otherAnswer)
	require.NoError(t, err)
	ownAttachment := insertTestAttachment(t, engine, "answer-a", "own")
	policy := insertTestAttachment(t, engine, "policy-terms", "terms")

	var s3Deleted []string
	origDelete, origUpload := domain.DeleteAttachmentFileFunc, domain.UploadAttachmentFileFunc
	t.Cleanup(func() { domain.DeleteAttachmentFileFunc, domain.UploadAttachmentFileFunc = origDelete, origUpload })
	domain.DeleteAttachmentFileFunc = func(_ context.Context, a *domain.Attachment) error {
		s3Deleted = append(s3Deleted, a.GetKey())
		return nil
	}
	domain.UploadAttachmentFileFunc = func(_ context.Context, _ string, obj s3utils.S3Object, _ *multipart.FileHeader) (string, error) {
		return "https://cdn.example.invalid/" + obj.GetKey(), nil
	}

	call := func(qnaId, answerId int64, deletes []string, file string) error {
		var body bytes.Buffer
		w := multipart.NewWriter(&body)
		require.NoError(t, w.WriteField("answer", "new answer"))
		for i, d := range deletes {
			require.NoError(t, w.WriteField("attachments_delete["+strconv.Itoa(i)+"]", d))
		}
		if file != "" {
			fw, err := w.CreateFormFile("files", file)
			require.NoError(t, err)
			_, _ = fw.Write([]byte("%PDF-1.4"))
		}
		require.NoError(t, w.Close())
		e := echo.New()
		e.Validator = validator.NewCustomValidator()
		req := httptest.NewRequest(http.MethodPut, "/", &body)
		req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
		c := e.NewContext(req.WithContext(ctx), httptest.NewRecorder())
		c.SetParamNames("id", "answerId")
		c.SetParamValues(strconv.FormatInt(qnaId, 10), strconv.FormatInt(answerId, 10))
		c.Set("user", jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user": "1", "name": "admin", "email": "admin@example.invalid"}))
		return QnaApiController{}.UpdateAnswer(c)
	}

	t.Run("他のQnAの回答のIDは404で、書き換えない", func(t *testing.T) {
		err := call(qna.Id, otherAnswer.Id, nil, "")
		var fe *factory.Error
		require.True(t, errors.As(err, &fe))
		assert.Equal(t, factory.ErrorNotFound.Code, fe.Code)
		var got domain.Answer
		_, err = engine.ID(otherAnswer.Id).Get(&got)
		require.NoError(t, err)
		assert.Equal(t, "other", got.Answer)
		assert.Equal(t, otherQna.Id, got.QnaId)
	})

	t.Run("紐づけ先が空の回答にファイルが来たら400", func(t *testing.T) {
		emptyAnswer := domain.Answer{QnaId: qna.Id, Answer: "empty"}
		_, err := engine.Insert(&emptyAnswer)
		require.NoError(t, err)
		_, err = engine.Exec("UPDATE answer SET common_id = '' WHERE id = ?", emptyAnswer.Id)
		require.NoError(t, err)

		err = call(qna.Id, emptyAnswer.Id, nil, "new.pdf")
		var fe *factory.Error
		require.True(t, errors.As(err, &fe))
		assert.Equal(t, factory.ErrorParameter.Code, fe.Code)
		n, err := engine.Where("resource_id = ''").Count(&domain.Attachment{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), n, "紐づけ先が空の添付を作らない")
	})

	t.Run("回答の添付の削除と追加は、この回答に付く(resource_idが空にならない)", func(t *testing.T) {
		err := call(qna.Id, answer.Id, []string{strconv.FormatInt(ownAttachment.Id, 10), strconv.FormatInt(policy.Id, 10)}, "new.pdf")
		require.NoError(t, err)

		assert.True(t, isDeleted(t, engine, &domain.Attachment{}, ownAttachment.Id))
		assert.False(t, isDeleted(t, engine, &domain.Attachment{}, policy.Id), "規約の添付は消えない")
		var added []domain.Attachment
		require.NoError(t, engine.Where("original_name = ?", "new.pdf").Find(&added))
		require.Len(t, added, 1)
		assert.Equal(t, "answer-a", added[0].ResourceId)
		var got domain.Answer
		_, err = engine.ID(answer.Id).Get(&got)
		require.NoError(t, err)
		assert.Equal(t, "answer-a", got.CommonId)
		assert.Equal(t, "new answer", got.Answer)

		assert.Empty(t, s3Deleted, "S3は確定のあとで消す")
		require.NoError(t, factory.RunAfterCommit(ctx, ctx))
		assert.Equal(t, []string{"/attachment_file/own.pdf"}, s3Deleted)
	})
}

func TestFindAnswerById(t *testing.T) {
	answers := []*domain.Answer{{Id: 1}, nil, {Id: 3}}
	assert.Equal(t, int64(3), findAnswerById(answers, 3).Id)
	assert.Nil(t, findAnswerById(answers, 2))
	assert.Nil(t, findAnswerById(nil, 1))
}

// 資金調達相談の更新で、更新前の相談が無いときは panic せず404にする。
func TestFundingRequestUpdate_NotFound(t *testing.T) {
	_, ctx := newAdminAttachmentTestEngine(t, new(domain.FundingRequest), new(domain.Attachment))
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for k, v := range map[string]string{"name": "会社", "address": "住所", "email": "a@example.invalid", "phone": "0312345678"} {
		require.NoError(t, w.WriteField(k, v))
	}
	require.NoError(t, w.Close())
	e := echo.New()
	e.Validator = validator.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPut, "/", &body)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	c := e.NewContext(req.WithContext(ctx), httptest.NewRecorder())
	c.SetParamNames("id")
	c.SetParamValues("999")
	c.Set("user", jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user": "1", "email": "admin@example.invalid"}))

	var err error
	require.NotPanics(t, func() { err = FundingRequestApiController{}.Update(c) })
	var fe *factory.Error
	require.True(t, errors.As(err, &fe), "err = %v", err)
	assert.Equal(t, factory.ErrorNotFound.Code, fe.Code)
}
