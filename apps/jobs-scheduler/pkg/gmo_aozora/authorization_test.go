package gmo_aozora

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthorization_ExtractsHashedStateNotSessionId はMIR-1298の最重要の検証。
// ライブラリ(gmo-aozora-api-go)はOpenIDAuthorizationの中でstate =
// sha256(sessionID + salt)を計算してURLへ載せる(sessionIDそのものではない)。
// もし受け取ったsessionID自身をstateとして照合する実装だったら、正規の
// ログインが全件落ちる(設計コメント参照)。ここでは実際にライブラリを呼び、
// 返ってきたURLから取り出したstateが、sessionIdの生値ではなくハッシュ値に
// なっていることを確認する。
//
// テスト環境にconf.jsonが無いため、ライブラリのsalt(config.Salt)は空文字列
// のままになる。ganb.go の getState はこれに依存しないロジック
// (sha256(sessionID + salt)、salt=""でも正しく計算できる)なので、salt値の
// 有無に関わらずこのテストは有効。
func TestAuthorization_ExtractsHashedStateNotSessionId(t *testing.T) {
	originalClientInfo := clientInfo
	t.Cleanup(func() { clientInfo = originalClientInfo })
	clientInfo = ClientInfo{
		ClientId: "test-client-id", ClientSecret: "test-client-secret",
		Scope: "openid", RedirectUri: "https://example.test/callback",
	}

	sessionId := "S-test-session-value"

	result, err := Authorization(sessionId)
	require.NoError(t, err)
	require.NotNil(t, result)

	// ライブラリの内部実装(getState = sha256(sessionID + salt))を、salt=""
	// (このテスト環境の実測)として独立に再計算する。
	sum := sha256.Sum256([]byte(sessionId + ""))
	expectedState := hex.EncodeToString(sum[:])

	assert.Equal(t, expectedState, result.State, "URLのstateはsessionIdのsha256であり、生値ではないこと")
	assert.NotEqual(t, sessionId, result.State, "sessionIdをそのままstateとして使っていないこと")

	_, err = uuid.Parse(result.Nonce)
	assert.NoError(t, err, "nonceはUUID形式であること")

	assert.Contains(t, result.URL, "state="+result.State)
	assert.Contains(t, result.URL, "nonce="+result.Nonce)
}

func TestAuthorization_MissingClientCredentials(t *testing.T) {
	originalClientInfo := clientInfo
	t.Cleanup(func() { clientInfo = originalClientInfo })
	clientInfo = ClientInfo{}

	// clientId/clientSecretが空でも ganblib.New 自体はエラーにならない
	// (このライブラリの実装)。呼び出し側の責務は、Authorizationを呼ぶ前に
	// LoadClientでclientInfoが正しく初期化されていることを保証すること。
	// ここでは少なくともpanicせず、有効なstate/nonceを返すことだけ確認する。
	result, err := Authorization("session")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.State)
	assert.NotEmpty(t, result.Nonce)
}
