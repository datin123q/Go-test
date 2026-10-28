package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"energyshares-job-scheduler/pkg/gmo_aozora"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

func setupGmoAozoraOAuthControllerTestDB(t *testing.T) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(domain.GmoAozoraOAuthState)))

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}

func executeGetLoginForTest(ctx context.Context, state string) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodGet, "/auth/login?state="+state, nil).WithContext(ctx)
	recorder := httptest.NewRecorder()
	echoContext := echo.New().NewContext(req, recorder)

	err := (AozoraApiControllers{}).GetLogin(echoContext)
	return recorder, err
}

// TestGetLogin_RejectsUnissuedOrExpiredOrUsedState はMIR-1298の要求:
// GetLoginはadmin-apiが発行し保存済みのstateでなければ400で弾く。
func TestGetLogin_RejectsUnissuedOrExpiredOrUsedState(t *testing.T) {
	t.Run("state未指定", func(t *testing.T) {
		ctx := setupGmoAozoraOAuthControllerTestDB(t)
		_, err := executeGetLoginForTest(ctx, "")
		require.Error(t, err)
	})

	t.Run("未発行のstate(攻撃者による推測)", func(t *testing.T) {
		ctx := setupGmoAozoraOAuthControllerTestDB(t)
		_, err := executeGetLoginForTest(ctx, "never-issued")
		require.Error(t, err)
		var apiErr *factory.Error
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.Status())
	})

	t.Run("期限切れのstate", func(t *testing.T) {
		ctx := setupGmoAozoraOAuthControllerTestDB(t)
		state := "S-expired"
		require.NoError(t, (domain.GmoAozoraOAuthState{}).Create(ctx, domain.Sha256Hex(state), 1, time.Now().UTC().Add(-time.Minute)))

		_, err := executeGetLoginForTest(ctx, state)
		require.Error(t, err)
	})

	// ②レビューの指摘(MIR-1298): ClaimByStateHashは`WHERE state_hash = ? AND
	// used_at IS NULL`が条件のため、used_atが立つ行は必ずstate_hashも
	// 立っている(UsedAt単独では起こらない)。このテストは実質的に
	// 「state_hash発行済み(=GetLoginを既に1回通した)」を確認している。
	t.Run("state_hash発行済み(GetLoginを既に1回通した)", func(t *testing.T) {
		ctx := setupGmoAozoraOAuthControllerTestDB(t)
		state := "S-used"
		lookup := domain.Sha256Hex(state)
		require.NoError(t, (domain.GmoAozoraOAuthState{}).Create(ctx, lookup, 1, time.Now().UTC().Add(time.Hour)))
		row, err := (domain.GmoAozoraOAuthState{}).GetByLookup(ctx, lookup)
		require.NoError(t, err)
		_, err = (domain.GmoAozoraOAuthState{}).SetStateHashOnce(ctx, row.Id, "some-hash", "some-nonce")
		require.NoError(t, err)
		_, err = (domain.GmoAozoraOAuthState{}).ClaimByStateHash(ctx, "some-hash", time.Now().UTC())
		require.NoError(t, err)

		_, err = executeGetLoginForTest(ctx, state)
		require.Error(t, err)
	})
}

// TestGetLogin_ValidStateReturnsURLAndPersistsHash はMIR-1298の正常系。
func TestGetLogin_ValidStateReturnsURLAndPersistsHash(t *testing.T) {
	ctx := setupGmoAozoraOAuthControllerTestDB(t)
	state := "S-valid"
	lookup := domain.Sha256Hex(state)
	require.NoError(t, (domain.GmoAozoraOAuthState{}).Create(ctx, lookup, 9, time.Now().UTC().Add(time.Hour)))

	recorder, err := executeGetLoginForTest(ctx, state)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "state=")

	row, err := (domain.GmoAozoraOAuthState{}).GetByLookup(ctx, lookup)
	require.NoError(t, err)
	require.NotNil(t, row.StateHash)
	require.NotNil(t, row.Nonce)
	assert.NotEqual(t, state, *row.StateHash, "保存するstate_hashはSの生値ではなくハッシュであること")
}

// TestGetLogin_SecondCallForSameStateIsRejected は①敵対的検証の指摘
// (MIR-1298): GetLoginが同じstateで2回呼ばれた場合、2回目はSetStateHashOnceが
// 書き戻さない(1回目のnonceを保護するため)。2回目のURLをそのまま成功として
// 返すと、そのURLでログインしてもnonceが一致せず必ず失敗する。2回目は
// エラーとして返し、担当者に1回目のURLを使わせる。
func TestGetLogin_SecondCallForSameStateIsRejected(t *testing.T) {
	ctx := setupGmoAozoraOAuthControllerTestDB(t)
	state := "S-double-call"
	require.NoError(t, (domain.GmoAozoraOAuthState{}).Create(ctx, domain.Sha256Hex(state), 1, time.Now().UTC().Add(time.Hour)))

	first, err := executeGetLoginForTest(ctx, state)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, first.Code)

	_, err = executeGetLoginForTest(ctx, state)
	require.Error(t, err, "2回目は成功として返さないこと(1回目のnonceを保護するため)")
}

func executeAozoraRedirectForTest(ctx context.Context, query string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/auth/access-token?"+query, nil).WithContext(ctx)
	recorder := httptest.NewRecorder()
	echoContext := echo.New().NewContext(req, recorder)

	// AozoraRedirectUriは常にnilを返しHTMLへ直接書く設計のため、errは無視してよい。
	_ = (AozoraApiControllers{}).AozoraRedirectUri(echoContext)
	return recorder
}

// TestAozoraRedirectUri_MissingErrorDescriptionDoesNotPanic はMIR-1298で
// 修正した不具合(旧実装はerror_descriptionのnil参照でpanicしていた)の回帰確認。
func TestAozoraRedirectUri_MissingErrorDescriptionDoesNotPanic(t *testing.T) {
	ctx := setupGmoAozoraOAuthControllerTestDB(t)

	assert.NotPanics(t, func() {
		recorder := executeAozoraRedirectForTest(ctx, "error=access_denied")
		assert.GreaterOrEqual(t, recorder.Code, 400)
	})
}

func TestAozoraRedirectUri_MissingCodeOrState(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"codeが無い", "state=abc"},
		{"stateが無い", "code=abc"},
		{"どちらも無い", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupGmoAozoraOAuthControllerTestDB(t)
			recorder := executeAozoraRedirectForTest(ctx, tt.query)
			assert.Equal(t, http.StatusBadRequest, recorder.Code)
		})
	}
}

// TestAozoraRedirectUri_SetsSecurityHeaders はMIR-1298の「エラーと画面」節:
// Cache-Control: no-store、Referrer-Policy: no-referrer、
// Content-Security-Policy: default-src 'none' を成功・失敗どちらでも付ける。
func TestAozoraRedirectUri_SetsSecurityHeaders(t *testing.T) {
	ctx := setupGmoAozoraOAuthControllerTestDB(t)
	recorder := executeAozoraRedirectForTest(ctx, "code=abc&state=never-issued")

	assert.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
	assert.Equal(t, "no-referrer", recorder.Header().Get("Referrer-Policy"))
	assert.Equal(t, "default-src 'none'", recorder.Header().Get("Content-Security-Policy"))
	assert.Contains(t, recorder.Header().Get("Content-Type"), "text/html")
}

// TestClassifyAozoraTokenError は②レビューの指摘(MIR-1298):
// 失敗理由ごとに画面文言・ステータスコードを振り分けるロジック自体にテストが
// 無かった。
func TestClassifyAozoraTokenError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "口座確認失敗",
			err:        gmo_aozora.ErrAccountVerifyFailed,
			wantStatus: http.StatusBadGateway,
			wantMsg:    "口座を確認できませんでした。管理画面からやり直してください。",
		},
		{
			name:       "DB保存失敗",
			err:        gmo_aozora.ErrTokenSaveFailed,
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "システムエラーが発生しました。管理画面からやり直してください。",
		},
		{
			name:       "トークン不正(銀行側のエラー)",
			err:        gmo_aozora.ErrTokenInvalid,
			wantStatus: http.StatusBadGateway,
			wantMsg:    "銀行側でエラーが発生しました。管理画面からやり直してください。",
		},
		{
			name:       "交換失敗(通信エラー)",
			err:        gmo_aozora.ErrTokenExchangeFailed,
			wantStatus: http.StatusBadGateway,
			wantMsg:    "銀行との通信に失敗しました。管理画面からやり直してください。",
		},
		{
			name:       "未分類のエラーは銀行側のエラーとして扱う",
			err:        errors.New("something else"),
			wantStatus: http.StatusBadGateway,
			wantMsg:    "銀行側でエラーが発生しました。管理画面からやり直してください。",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantStatus, classifyAozoraTokenErrorStatus(tt.err))
			assert.Equal(t, tt.wantMsg, classifyAozoraTokenErrorMessage(tt.err))
		})
	}
}

func TestAozoraRedirectUri_UnknownStateFails(t *testing.T) {
	ctx := setupGmoAozoraOAuthControllerTestDB(t)

	recorder := executeAozoraRedirectForTest(ctx, "code=abc&state=never-issued")
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "never-issued", "受け取ったstateの値を画面へ出さないこと")
}

// TestAozoraRedirectUri_ExpiredStateFails は②レビューの指摘(MIR-1298):
// 従来のUnknownOrExpiredという名前のテストは、実際には期限切れではなく
// 未発行のケースしか確認していなかった。期限切れを独立して確認する。
func TestAozoraRedirectUri_ExpiredStateFails(t *testing.T) {
	ctx := setupGmoAozoraOAuthControllerTestDB(t)
	receivedState := "H-already-expired"
	stateHash := domain.Sha256Hex(receivedState)

	require.NoError(t, (domain.GmoAozoraOAuthState{}).Create(ctx, domain.Sha256Hex("S-already-expired"), 1, time.Now().UTC().Add(-time.Minute)))
	row, err := (domain.GmoAozoraOAuthState{}).GetByLookup(ctx, domain.Sha256Hex("S-already-expired"))
	require.NoError(t, err)
	_, err = (domain.GmoAozoraOAuthState{}).SetStateHashOnce(ctx, row.Id, stateHash, "nonce")
	require.NoError(t, err)

	recorder := executeAozoraRedirectForTest(ctx, "code=abc&state="+receivedState)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestAozoraRedirectUri_ReplayOfAlreadyProcessedState はMIR-1298の
// 「同じstateで2回同時に来ても片方だけが通る」「リロードは既に出した結果を
// 案内する」の確認。processAozoraRedirectはコールバックで受け取ったstate
// (=H、ライブラリが実際にURLへ載せた値)をそのままsha256してstate_hashと
// 照合するため、ここではHの値そのもの("H-replay")をクエリのstateに使い、
// 保存するstate_hashはそのsha256にする。
func TestAozoraRedirectUri_ReplayOfAlreadyProcessedState(t *testing.T) {
	ctx := setupGmoAozoraOAuthControllerTestDB(t)
	receivedState := "H-replay"
	stateHash := domain.Sha256Hex(receivedState)

	require.NoError(t, (domain.GmoAozoraOAuthState{}).Create(ctx, domain.Sha256Hex("S-replay"), 1, time.Now().UTC().Add(time.Hour)))
	row, err := (domain.GmoAozoraOAuthState{}).GetByLookup(ctx, domain.Sha256Hex("S-replay"))
	require.NoError(t, err)
	_, err = (domain.GmoAozoraOAuthState{}).SetStateHashOnce(ctx, row.Id, stateHash, "nonce")
	require.NoError(t, err)

	claimed, err := (domain.GmoAozoraOAuthState{}).ClaimByStateHash(ctx, stateHash, time.Now().UTC())
	require.NoError(t, err)
	require.NotNil(t, claimed)
	require.NoError(t, (domain.GmoAozoraOAuthState{}).RecordResult(ctx, claimed.Id, domain.GmoAozoraOAuthResultSuccess))

	recorder := executeAozoraRedirectForTest(ctx, "code=abc&state="+receivedState)
	assert.Equal(t, http.StatusOK, recorder.Code, "既に成功として処理済みのstateは、リロードとして案内すること(コード交換をやり直さない)")
	assert.Contains(t, recorder.Body.String(), "既に処理済み")
}

// TestAozoraRedirectUri_ReplayWhileStillProcessing は①敵対的検証の指摘
// (MIR-1298): 口座確認のレートリミッタで1回目の処理に数秒かかる。その間に
// F5等で来た2回目は、使用済み(UsedAt有り)だがまだ結果(Result)が記録されて
// いない状態を見る。「失敗」と断定せず、処理中である旨を案内することを確認する。
func TestAozoraRedirectUri_ReplayWhileStillProcessing(t *testing.T) {
	ctx := setupGmoAozoraOAuthControllerTestDB(t)
	receivedState := "H-in-progress"
	stateHash := domain.Sha256Hex(receivedState)

	require.NoError(t, (domain.GmoAozoraOAuthState{}).Create(ctx, domain.Sha256Hex("S-in-progress"), 1, time.Now().UTC().Add(time.Hour)))
	row, err := (domain.GmoAozoraOAuthState{}).GetByLookup(ctx, domain.Sha256Hex("S-in-progress"))
	require.NoError(t, err)
	_, err = (domain.GmoAozoraOAuthState{}).SetStateHashOnce(ctx, row.Id, stateHash, "nonce")
	require.NoError(t, err)

	// 1回目が処理中(ClaimByStateHash済み、まだRecordResult前)を再現する。
	claimed, err := (domain.GmoAozoraOAuthState{}).ClaimByStateHash(ctx, stateHash, time.Now().UTC())
	require.NoError(t, err)
	require.NotNil(t, claimed)

	recorder := executeAozoraRedirectForTest(ctx, "code=abc&state="+receivedState)
	assert.NotContains(t, recorder.Body.String(), "失敗", "処理中を失敗と誤って案内しないこと")
	assert.Contains(t, recorder.Body.String(), "処理中")
}
