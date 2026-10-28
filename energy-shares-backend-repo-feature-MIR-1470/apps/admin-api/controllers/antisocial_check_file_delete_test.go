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
	"time"

	"github.com/hillfolk/goutils/echomiddleware"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

// MIR-1413: 反社チェックの書類の削除は、対象の会員の書類だけ。S3のファイルはDBの確定のあとに消す。

func TestParseAntisocialCheckFileDeleteIds(t *testing.T) {
	tests := []struct {
		name    string
		values  map[string][]string
		want    []int64
		wantErr bool
	}{
		{name: "指定なし", values: map[string][]string{"checker": {"a"}}, want: []int64{}},
		{name: "画面が送るキー", values: map[string][]string{"antisocial_check_file_delete[0]": {"3"}, "antisocial_check_file_delete[1]": {"4"}}, want: []int64{3, 4}},
		{name: "数字でない値はエラー", values: map[string][]string{"antisocial_check_file_delete[0]": {"3"}, "antisocial_check_file_delete[1]": {"x"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAntisocialCheckFileDeleteIds(tt.values)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestClassifyAntisocialCheckFileDeletes(t *testing.T) {
	deletedAt := jsonutils.JSONTime(time.Now())
	found := []domain.AntisocialCheckFile{{Id: 1, CheckId: 10}, {Id: 2, CheckId: 10}, {Id: 3, CheckId: 20},
		{Id: 4, CheckId: 10, Deleted: deletedAt}, {Id: 5, CheckId: 20, Deleted: deletedAt}}
	tests := []struct {
		name        string
		requested   []int64
		wantDelete  []int64
		wantIgnored []int64
	}{
		{name: "空", requested: nil},
		{name: "対象の会員の書類は消す", requested: []int64{1, 2}, wantDelete: []int64{1, 2}},
		{name: "他の会員の書類は無視", requested: []int64{3}, wantIgnored: []int64{3}},
		{name: "存在しないIDは無視", requested: []int64{99}, wantIgnored: []int64{99}},
		{name: "対象の会員の論理削除済みは黙って飛ばす(送り直し)", requested: []int64{4}},
		{name: "他の会員の論理削除済みは無視", requested: []int64{5}, wantIgnored: []int64{5}},
		{name: "重複は1回だけ", requested: []int64{1, 1, 3, 3}, wantDelete: []int64{1}, wantIgnored: []int64{3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toDelete, ignored := classifyAntisocialCheckFileDeletes(tt.requested, found, 10)
			ids := make([]int64, 0, len(toDelete))
			for _, f := range toDelete {
				ids = append(ids, f.Id)
			}
			assert.ElementsMatch(t, tt.wantDelete, ids)
			assert.ElementsMatch(t, tt.wantIgnored, ignored)
		})
	}
}

func newAntisocialTestEngine(t *testing.T) (*xorm.Engine, context.Context) {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync(new(domain.AntisocialCheckFile)))
	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	ctx, _ = factory.WithAfterCommit(ctx)
	return engine, ctx
}

func insertAntisocialCheckFile(t *testing.T, engine *xorm.Engine, checkId int64, name string) domain.AntisocialCheckFile {
	t.Helper()
	f := domain.AntisocialCheckFile{CheckId: checkId, FileName: name, OriginalName: name + ".pdf", FileExtension: ".pdf", FilePath: "/antisocial/"}
	_, err := engine.Insert(&f)
	require.NoError(t, err)
	return f
}

func antisocialFileDeleted(t *testing.T, engine *xorm.Engine, id int64) bool {
	t.Helper()
	var f domain.AntisocialCheckFile
	has, err := engine.Unscoped().ID(id).Get(&f)
	require.NoError(t, err)
	require.True(t, has)
	return !f.Deleted.IsZero()
}

func TestDeleteOwnedAntisocialCheckFiles(t *testing.T) {
	engine, ctx := newAntisocialTestEngine(t)
	own := insertAntisocialCheckFile(t, engine, 10, "own")
	other := insertAntisocialCheckFile(t, engine, 20, "other")
	deleted := insertAntisocialCheckFile(t, engine, 10, "deleted")
	_, err := engine.ID(deleted.Id).Delete(&domain.AntisocialCheckFile{})
	require.NoError(t, err)
	var before domain.AntisocialCheckFile
	_, err = engine.Unscoped().ID(deleted.Id).Get(&before)
	require.NoError(t, err)
	logger, hook := logrustest.NewNullLogger()
	ctx = context.WithValue(ctx, echomiddleware.ContextLoggerName, logrus.NewEntry(logger))

	var got []domain.AntisocialCheckFile
	require.NotPanics(t, func() {
		got, err = deleteOwnedAntisocialCheckFiles(ctx, []int64{own.Id, other.Id, deleted.Id, 999999}, 10, logrus.Fields{"check_user_id": 10})
	})
	require.NoError(t, err, "管理画面の操作は止めない(対象外はログを残して無視)")
	require.Len(t, got, 1)
	assert.Equal(t, own.Id, got[0].Id)
	assert.True(t, antisocialFileDeleted(t, engine, own.Id))
	assert.False(t, antisocialFileDeleted(t, engine, other.Id), "他の会員の反社チェックの書類は消えない")

	// 自分の論理削除済み(画面からの送り直し)は警告に含めず、削除の時刻も変えない。
	var warned []int64
	for _, e := range hook.AllEntries() {
		if e.Level == logrus.WarnLevel {
			warned = append(warned, e.Data["antisocial_check_file_ids"].([]int64)...)
		}
	}
	assert.Equal(t, []int64{other.Id, 999999}, warned)
	var after domain.AntisocialCheckFile
	_, err = engine.Unscoped().ID(deleted.Id).Get(&after)
	require.NoError(t, err)
	assert.Equal(t, before.Deleted.ToString(), after.Deleted.ToString())
}

func TestAntisocialCheckFile_DeleteByIdAndCheckId(t *testing.T) {
	engine, ctx := newAntisocialTestEngine(t)
	other := insertAntisocialCheckFile(t, engine, 20, "other")
	require.NoError(t, domain.AntisocialCheckFile{}.DeleteByIdAndCheckId(ctx, other.Id, 10))
	assert.False(t, antisocialFileDeleted(t, engine, other.Id), "check_idが違えば、IDが合っても消さない")
	require.NoError(t, domain.AntisocialCheckFile{}.DeleteByIdAndCheckId(ctx, other.Id, 20))
	assert.True(t, antisocialFileDeleted(t, engine, other.Id))
}

func TestDeleteAntisocialCheckFilesAfterCommit(t *testing.T) {
	orig := deleteAntisocialCheckFileFunc
	t.Cleanup(func() { deleteAntisocialCheckFileFunc = orig })

	t.Run("確定のあとにだけ消す", func(t *testing.T) {
		var keys []string
		deleteAntisocialCheckFileFunc = func(_ context.Context, f *domain.AntisocialCheckFile) error {
			keys = append(keys, f.GetKey())
			return nil
		}
		ctx, _ := factory.WithAfterCommit(context.Background())
		deleteAntisocialCheckFilesAfterCommit(ctx, []domain.AntisocialCheckFile{{Id: 1, FilePath: "/antisocial/", FileName: "a", FileExtension: ".pdf"}})
		assert.Empty(t, keys, "確定前には消さない")
		require.NoError(t, factory.RunAfterCommit(ctx, context.Background()))
		assert.Equal(t, []string{"/antisocial/a.pdf"}, keys)
	})

	t.Run("S3の失敗やpanicでも、ほかのファイルを消して応答は壊さない", func(t *testing.T) {
		var called []int64
		deleteAntisocialCheckFileFunc = func(_ context.Context, f *domain.AntisocialCheckFile) error {
			called = append(called, f.Id)
			if f.Id == 1 {
				panic("boom")
			}
			return errors.New("s3 error")
		}
		ctx, _ := factory.WithAfterCommit(context.Background())
		deleteAntisocialCheckFilesAfterCommit(ctx, []domain.AntisocialCheckFile{{Id: 2}, {Id: 1}})
		require.NotPanics(t, func() { require.NoError(t, factory.RunAfterCommit(ctx, context.Background())) })
		assert.Equal(t, []int64{1, 2}, called)
	})
}

// 数字でない削除指定は、会員の状態の更新や不承認のメールより前に400で弾く。
func TestAntisocialCheckStatusUpdate_InvalidDeleteIdIsRejectedFirst(t *testing.T) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	require.NoError(t, w.WriteField("check_status", "2"))
	require.NoError(t, w.WriteField("checker", "admin"))
	require.NoError(t, w.WriteField("antisocial_check_file_delete[0]", "abc"))
	require.NoError(t, w.Close())

	e := echo.New()
	e.Validator = validator.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPut, "/", &body)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	c := e.NewContext(req, httptest.NewRecorder())
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(10))

	// DBもトークンも用意しない: 入力の検証だけで返ること(それより先に進むと、ここで落ちる)。
	err := UserApiController{}.AntisocialCheckStatusUpdate(c)
	var fe *factory.Error
	require.True(t, errors.As(err, &fe), "err = %v", err)
	assert.Equal(t, factory.ErrorParameter.Code, fe.Code)
}
