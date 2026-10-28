package gmo_aozora

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"

	ganblib "github.com/gmoaozora/gmo-aozora-api-go"
	corporateclient "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// TestValidateToken はMIR-1298の妥当性確認そのもの。ライブラリ
// (gmo-aozora-api-go)のOpenIDGetToken/RefreshTokensは、応答のJSON解析に
// 失敗すると空のトークンとエラー無しを返すことがあり、RefreshTokensは
// HTTPの状態コードを見ないため{"error":"invalid_grant"}のような応答でも
// 解析自体は成功する。中身を見て初めて弾けることを確認する。
func TestValidateToken(t *testing.T) {
	tests := []struct {
		name    string
		token   ganblib.Token
		wantErr string
	}{
		{
			name:    "正常なトークン",
			token:   ganblib.Token{AccessToken: "at", RefreshToken: "rt"},
			wantErr: "",
		},
		{
			name:    "token.Errorが空でない(銀行がエラーを返した)",
			token:   ganblib.Token{Error: "invalid_grant", AccessToken: "at", RefreshToken: "rt"},
			wantErr: "gmo aozora token error",
		},
		{
			name:    "AccessTokenが空(ライブラリのJSON解析失敗パターン)",
			token:   ganblib.Token{RefreshToken: "rt"},
			wantErr: "access token is empty",
		},
		{
			name:    "RefreshTokenが空",
			token:   ganblib.Token{AccessToken: "at"},
			wantErr: "refresh token is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateToken(tt.token)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func setupGmoAozoraClientTestDB(t *testing.T, initial domain.GmoAozoraClient) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	require.NoError(t, engine.Sync2(new(domain.GmoAozoraClient)))

	initial.Id = 1
	_, err = initial.Create(context.WithValue(context.Background(), factory.ConextDBType, engine))
	require.NoError(t, err)

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}

// TestUpdateStoredToken_PersistsAndRoundTripsは、GmoAozoraClient.Updateが
// 渡した構造体のAccessToken/RefreshTokenを暗号文へ書き換える(MIR-1298の
// 設計コメント参照)ため、DBから読み直した値が平文のトークンと一致することを
// 確認する。
func TestUpdateStoredToken_PersistsAndRoundTrips(t *testing.T) {
	ctx := setupGmoAozoraClientTestDB(t, domain.GmoAozoraClient{
		ClientId: "cid", ClientSecret: "csecret", AccessToken: "old-at", RefreshToken: "old-rt",
	})

	err := updateStoredToken(ctx, ganblib.Token{AccessToken: "new-at", RefreshToken: "new-rt"})
	require.NoError(t, err)

	reloaded, err := domain.GmoAozoraClient{}.GetByID(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, reloaded)
	assert.Equal(t, "new-at", reloaded.AccessToken)
	assert.Equal(t, "new-rt", reloaded.RefreshToken)
}

// TestRefreshTokenWithRetryOnRejection_RereadsAndRetriesOnce はMIR-1298の
// 更新ジョブの要求: 銀行がrefresh_tokenを拒否した(token.Errorが空でない)場合、
// デプロイ中の他タスクが既に更新したrefresh_tokenをDBから読み直して1回だけ
// 再試行する。
func TestRefreshTokenWithRetryOnRejection_RereadsAndRetriesOnce(t *testing.T) {
	ctx := setupGmoAozoraClientTestDB(t, domain.GmoAozoraClient{
		AccessToken: "unused", RefreshToken: "new-rt-from-other-task",
	})

	originalCall := callRefreshTokens
	t.Cleanup(func() { callRefreshTokens = originalCall })

	var seenTokens []string
	callRefreshTokens = func(refreshToken string) (ganblib.Token, error) {
		seenTokens = append(seenTokens, refreshToken)
		if refreshToken == "stale-rt" {
			return ganblib.Token{Error: "invalid_grant"}, nil
		}
		return ganblib.Token{AccessToken: "at", RefreshToken: "rt2"}, nil
	}

	token, err := refreshTokenWithRetryOnRejection(ctx, "stale-rt")
	require.NoError(t, err)
	assert.Equal(t, "at", token.AccessToken)
	assert.Equal(t, []string{"stale-rt", "new-rt-from-other-task"}, seenTokens, "拒否されたら1回だけDBの最新値で再試行すること")
}

func TestRefreshTokenWithRetryOnRejection_NoRetryWhenDBValueIsSame(t *testing.T) {
	ctx := setupGmoAozoraClientTestDB(t, domain.GmoAozoraClient{
		AccessToken: "unused", RefreshToken: "same-rt",
	})

	originalCall := callRefreshTokens
	t.Cleanup(func() { callRefreshTokens = originalCall })

	callCount := 0
	callRefreshTokens = func(refreshToken string) (ganblib.Token, error) {
		callCount++
		return ganblib.Token{Error: "invalid_grant"}, nil
	}

	token, err := refreshTokenWithRetryOnRejection(ctx, "same-rt")
	require.NoError(t, err)
	assert.Equal(t, "invalid_grant", token.Error)
	assert.Equal(t, 1, callCount, "DBの値が最初に使った値と同じなら再試行しないこと")
}

// withClientInfoRefreshToken は、テストごとにclientInfo.RefreshTokenを
// 指定の値にして復元する。
func withClientInfoRefreshToken(t *testing.T, refreshToken string) {
	t.Helper()
	original := clientInfo
	t.Cleanup(func() { clientInfo = original })
	clientInfo = ClientInfo{RefreshToken: refreshToken}
}

// TestRefreshTokenLocked_RereadsLatestBeforeCallingBank はMIR-1298の
// 「銀行を呼ぶ前にDBから最新のトークンを読み直す」の確認。デプロイ中の古い
// タスクが、既に他タスクが更新した新しい系列を古い値で上書きしないため。
func TestRefreshTokenLocked_RereadsLatestBeforeCallingBank(t *testing.T) {
	ctx := setupGmoAozoraClientTestDB(t, domain.GmoAozoraClient{
		AccessToken: "old-at", RefreshToken: "latest-rt-in-db",
	})
	// ②レビュー2周目の指摘(MIR-1298): メモリをDBの値とわざと食い違わせる
	// ("stale-rt")。一致させていると、実装がメモリの値を使っても偶然テストが
	// 通ってしまう(「DBを読み直した」ことの証明にならない)。
	withClientInfoRefreshToken(t, "stale-rt")

	originalCall := callRefreshTokens
	t.Cleanup(func() { callRefreshTokens = originalCall })
	var usedRefreshToken string
	callRefreshTokens = func(refreshToken string) (ganblib.Token, error) {
		usedRefreshToken = refreshToken
		return ganblib.Token{AccessToken: "new-at", RefreshToken: "new-rt"}, nil
	}

	err := refreshTokenLocked(ctx)
	require.NoError(t, err)
	assert.Equal(t, "latest-rt-in-db", usedRefreshToken, "呼び出し元(このタスクが起動時に持っていた古い値)ではなく、DBの最新値を使うこと")

	reloaded, err := domain.GmoAozoraClient{}.GetByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "new-at", reloaded.AccessToken)
	assert.Equal(t, "new-rt", reloaded.RefreshToken)
}

// TestRefreshTokenLocked_MemoryKeepsNewTokenEvenWhenDBSaveFails は②レビューの
// 指摘(MIR-1298): ①の対応で一時追加した「DBの値がメモリと食い違えば他タスク
// 更新済みとみなしてスキップする」最適化は、削除した(client.goのコメント参照)。
// 理由をこのテストで固定する: 更新ジョブが銀行から新しいトークンを受け取った
// あとDB保存に失敗しても、メモリは新しいトークンのまま保たれ、次回の実行で
// 古いDBの値に巻き戻されないこと。
func TestRefreshTokenLocked_MemoryKeepsNewTokenEvenWhenDBSaveFails(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(domain.GmoAozoraClient)))
	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	seed := domain.GmoAozoraClient{Id: 1, AccessToken: "old-at", RefreshToken: "old-rt"}
	_, err = seed.Create(ctx)
	require.NoError(t, err)
	withClientInfoRefreshToken(t, "old-rt")

	callCount := 0
	originalCall := callRefreshTokens
	t.Cleanup(func() { callRefreshTokens = originalCall })
	callRefreshTokens = func(refreshToken string) (ganblib.Token, error) {
		callCount++
		// 呼ばれるたびに違う値を返す。②で見つかった退行(DBの値がメモリと
		// 食い違えば「他タスク更新済み」とみなしてメモリをDBの古い値へ
		// 巻き戻す)が再発していれば、2回目の呼び出し後にメモリがold-atへ
		// 戻ってしまうはず。
		return ganblib.Token{
			AccessToken:  fmt.Sprintf("new-at-%d", callCount),
			RefreshToken: fmt.Sprintf("new-rt-%d", callCount),
		}, nil
	}

	// UPDATE(保存)だけを確実に失敗させる。SELECT/INSERTは通るトリガーを
	// 張ることで、refreshTokenLocked冒頭の読み込みは成功させたまま、
	// updateStoredToken側だけを失敗させる。DBには古い値(old-at/old-rt)が
	// 残り続ける。
	_, err = engine.Exec(`CREATE TRIGGER block_update BEFORE UPDATE ON gmo_aozora_client BEGIN SELECT RAISE(ABORT, 'blocked for test'); END;`)
	require.NoError(t, err)

	err = refreshTokenLocked(ctx)
	require.Error(t, err, "DB保存の失敗はエラーとして呼び出し元(cron)へ伝わること")
	assert.Equal(t, "new-at-1", GetAccessToken(), "DB保存に失敗しても、メモリは銀行から受け取った新しいトークンのままであること(退行しないこと)")
	assert.Equal(t, "new-rt-1", GetRefreshTokenValue())

	// 翌日のcron実行を模す。DBは相変わらずold-rtのまま(保存が失敗し続けて
	// いるため)。この2回目の呼び出しでメモリがold-atへ巻き戻れば、②で
	// 見つかった退行の再発を意味する。
	err = refreshTokenLocked(ctx)
	require.Error(t, err)
	assert.Equal(t, "new-at-2", GetAccessToken(), "2回目の実行後もメモリが古いDBの値(old-at)へ巻き戻らないこと")
	assert.Equal(t, "new-rt-2", GetRefreshTokenValue())
}

// TestRefreshTokenLocked_InvalidTokenDoesNotOverwrite はMIR-1298:
// 空のトークンで上書きしないことの確認。
func TestRefreshTokenLocked_InvalidTokenDoesNotOverwrite(t *testing.T) {
	ctx := setupGmoAozoraClientTestDB(t, domain.GmoAozoraClient{
		AccessToken: "old-at", RefreshToken: "old-rt",
	})
	withClientInfoRefreshToken(t, "old-rt")

	originalCall := callRefreshTokens
	t.Cleanup(func() { callRefreshTokens = originalCall })
	callRefreshTokens = func(refreshToken string) (ganblib.Token, error) {
		return ganblib.Token{}, nil // ライブラリのJSON解析失敗パターン(空のトークン・エラー無し)
	}

	err := refreshTokenLocked(ctx)
	require.Error(t, err)

	reloaded, err := domain.GmoAozoraClient{}.GetByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "old-at", reloaded.AccessToken, "空のトークンでDBを上書きしないこと")
	assert.Equal(t, "old-rt", reloaded.RefreshToken)
}

// TestRefreshTokenLocked_DoesNotVerifyAccounts は①敵対的検証の指摘
// (MIR-1298、重大): 更新経路は既に自社のものと確定しているrefresh_tokenを
// 使い回すだけで、口座が別物にすり替わる余地が無い。口座確認をここで挟むと、
// GMOの一時的な通信障害だけで、銀行側では既に使用済みになった新しいトークン
// (refresh_tokenは使い切りのため元の値には戻れない)を保存できずに捨てる
// ことになり、守れるものが無いまま障害点を増やすだけだった(初回実装の不具合)。
// getAccountInfoClient・getAccountsWithTokenを差し替えず(呼ばれたら
// 即座に失敗する状態のまま)保存が成功することで、口座確認を行っていない
// ことを確認する。
func TestRefreshTokenLocked_DoesNotVerifyAccounts(t *testing.T) {
	ctx := setupGmoAozoraClientTestDB(t, domain.GmoAozoraClient{
		AccessToken: "old-at", RefreshToken: "old-rt",
	})
	withClientInfoRefreshToken(t, "old-rt")

	originalClient := getAccountInfoClient
	t.Cleanup(func() { getAccountInfoClient = originalClient })
	getAccountInfoClient = func() (*corporateclient.APIClient, error) {
		t.Fatal("refreshTokenLocked must not call the account API")
		return nil, nil
	}

	originalCall := callRefreshTokens
	t.Cleanup(func() { callRefreshTokens = originalCall })
	callRefreshTokens = func(refreshToken string) (ganblib.Token, error) {
		return ganblib.Token{AccessToken: "new-at", RefreshToken: "new-rt"}, nil
	}

	err := refreshTokenLocked(ctx)
	require.NoError(t, err)

	reloaded, err := domain.GmoAozoraClient{}.GetByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "new-at", reloaded.AccessToken)
}

func TestUpdateStoredToken_MissingRowFails(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(domain.GmoAozoraClient)))
	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	err = updateStoredToken(ctx, ganblib.Token{AccessToken: "new-at", RefreshToken: "new-rt"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "not found")
}

// TestRequestAozoraNewToken_SavesThroughTokenRefreshLock は②レビュー2周目の
// 指摘(MIR-1298): 確定設計が要求する「認可の戻り先の保存も更新ジョブと同じ
// ロックで囲む」を、RequestAozoraNewTokenが実際にwithTokenRefreshLock経由で
// 保存していることを確認する。GET_LOCKはMySQL固有でsqliteでは検証できない
// ため、withTokenRefreshLock自体を差し替え、「fnがロックの中から呼ばれた」
// ことだけを記録する(実際のGET_LOCK/RELEASE_LOCKの往復はlock.goを対象に
// ローカルMySQLで別途実機確認済み)。
func TestRequestAozoraNewToken_SavesThroughTokenRefreshLock(t *testing.T) {
	ctx := setupGmoAozoraClientTestDB(t, domain.GmoAozoraClient{
		AccessToken: "old-at", RefreshToken: "old-rt",
	})

	originalClientInfo := clientInfo
	t.Cleanup(func() { clientInfo = originalClientInfo })
	clientInfo = ClientInfo{ClientId: "cid", ClientSecret: "csecret", Scope: "openid", RedirectUri: "https://example.test/callback"}

	originalBankAccounts := BankAccounts
	t.Cleanup(func() { BankAccounts = originalBankAccounts })
	BankAccounts = Accounts{Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT)}

	// RequestAozoraNewTokenは保存成功後にWebhookSubscriptionを呼ぶ。
	// webHookClientCfgはLoadClient(サーバー起動時)でのみ設定されるため、
	// このテストでも用意しておかないとnilな*Configurationでpanicする。
	// BasePathは、実際の外部ホストへ向かわないよう、閉じたローカルアドレスに
	// 差し替える(応答が無く即座に接続拒否で失敗する。CIでのネットワーク
	// 依存・待ち時間を避けるため)。エラー自体はRequestAozoraNewToken内で
	// ログに残るだけで処理は継続する。
	closedListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	closedAddr := closedListener.Addr().String()
	require.NoError(t, closedListener.Close())

	originalWebHookCfg := webHookClientCfg
	t.Cleanup(func() { webHookClientCfg = originalWebHookCfg })
	webHookClientCfg = NewWebHookConfigure(STG_ENV)
	webHookClientCfg.BasePath = "http://" + closedAddr

	originalGetToken := openIDGetToken
	t.Cleanup(func() { openIDGetToken = originalGetToken })
	openIDGetToken = func(ganb ganblib.Ganb, redirectUri, code, authMethod string) (ganblib.Token, error) {
		return ganblib.Token{AccessToken: "exchanged-at", RefreshToken: "exchanged-rt"}, nil
	}

	withVerifyAccountsStubs(t,
		func() (*corporateclient.APIClient, error) { return &corporateclient.APIClient{}, nil },
		func(context.Context, *corporateclient.APIClient, string) (corporateclient.AccountsResponse, *http.Response, error) {
			return corporateclient.AccountsResponse{Accounts: []corporateclient.Account{{AccountNumber: "1234567"}}}, nil, nil
		},
	)
	// RequestAozoraNewTokenは保存成功後にGetAccountInfoも呼ぶ。これは
	// getAccountsWithTokenではなくgetAccounts(グローバルなclientInfoの
	// トークンを使う経路)を使うため、そちらも別に差し替える。
	originalGetAccounts := getAccounts
	t.Cleanup(func() { getAccounts = originalGetAccounts })
	getAccounts = func(context.Context, *corporateclient.APIClient) (corporateclient.AccountsResponse, *http.Response, error) {
		return corporateclient.AccountsResponse{Accounts: []corporateclient.Account{{AccountNumber: "1234567", AccountId: "account-id-1"}}}, nil, nil
	}

	// WebhookSubscriptionは実APIへ接続しに行くため差し替える対象が無い
	// (indirectionが無い旧実装)。エラーはRequestAozoraNewToken内でログに
	// 残るだけで処理は継続するので、失敗しても本テストの結果には影響しない。
	// ただしCIでの実行時間・ネットワーク依存を避けるため、到達不能な
	// アドレスを指すよう設定を差し替えて即座に失敗させる。
	originalLock := withTokenRefreshLock
	t.Cleanup(func() { withTokenRefreshLock = originalLock })
	lockCalled := false
	withTokenRefreshLock = func(ctx context.Context, fn func(ctx context.Context) error) error {
		lockCalled = true
		return fn(ctx)
	}

	token, err := RequestAozoraNewToken(ctx, "auth-code", "")
	require.NoError(t, err)
	assert.Equal(t, "exchanged-at", token.AccessToken)
	assert.True(t, lockCalled, "保存はwithTokenRefreshLock経由で行うこと")

	reloaded, err := domain.GmoAozoraClient{}.GetByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "exchanged-at", reloaded.AccessToken)
	assert.Equal(t, "exchanged-at", GetAccessToken(), "setClientInfoもロックの中(=保存成功後)で行われること")
}
