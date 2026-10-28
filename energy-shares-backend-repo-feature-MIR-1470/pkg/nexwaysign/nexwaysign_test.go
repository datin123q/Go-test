package nexwaysign

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testSecret = "9ab6bf804c3937fcf27ddaad49721c6b"

func TestSign_Rfc4231Vector(t *testing.T) {
	// RFC 4231 Test Case 2 (key="Jefe", data="what do ya want for nothing?")。
	// 署名対象は「本体 + タイムスタンプ」の単純な連結なので、データを2つに割って渡しても
	// 同じ値になることで、連結順(本体が先)とハッシュ関数・16進(小文字)を同時に固定する。
	got, err := Sign("Jefe", []byte("what do ya want "), "for nothing?")
	assert.NoError(t, err)
	assert.Equal(t, "5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843", got)
}

func TestSign_MultibyteBodyMatchesOpenssl(t *testing.T) {
	// 日本語を含む本体はUTF-8のバイト列のまま署名する。期待値は別実装(openssl)で計算した値:
	//   printf '%s' '<body><timestamp>' | openssl dgst -sha256 -hmac <secret>
	got, err := Sign(testSecret, []byte(`{"reference_id":"abc","name":"山田太郎"}`), "2026-10-15T12:00:00+09:00")
	assert.NoError(t, err)
	assert.Equal(t, "02648b1ad0e0c4a7f1eaaf3312116323c91f2d15a7ee8bb3ecae8934fc1a6357", got)
}

func TestSign_BodylessRequestSignsTimestampOnly(t *testing.T) {
	// 結果取得API(GET)は本体が無いのでタイムスタンプだけを署名する。nilと空は同じ結果。
	withNil, err := Sign("Jefe", nil, "what do ya want for nothing?")
	assert.NoError(t, err)
	withEmpty, err := Sign("Jefe", []byte{}, "what do ya want for nothing?")
	assert.NoError(t, err)
	assert.Equal(t, "5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843", withNil)
	assert.Equal(t, withNil, withEmpty)
}

func TestSign_EmptySecretIsRejected(t *testing.T) {
	// 空の鍵で計算した署名は誰でも作れる。黙って通さず、エラーにする。
	_, err := Sign("", []byte(`{"a":1}`), "2026-10-15T09:00:00+09:00")
	assert.ErrorIs(t, err, ErrSecretNotSet)

	_, _, err = SignRequest("", []byte(`{"a":1}`), time.Now())
	assert.ErrorIs(t, err, ErrSecretNotSet)
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want string
	}{
		{"UTCはJST(+09:00)へ変換される", time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), "2026-10-15T09:00:00+09:00"},
		{"日付をまたぐ変換", time.Date(2026, 10, 14, 15, 30, 5, 0, time.UTC), "2026-10-15T00:30:05+09:00"},
		{"秒未満は切り捨てられる", time.Date(2026, 10, 15, 9, 0, 0, 999_000_000, jst), "2026-10-15T09:00:00+09:00"},
		{"他のオフセットでも+09:00で出る", time.Date(2026, 10, 15, 0, 0, 0, 0, time.FixedZone("X", -5*3600)), "2026-10-15T14:00:00+09:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatTime(tt.in))
		})
	}
}

func TestSignRequest_RoundTripsThroughVerify(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, jst)
	body := []byte(`{"reference_id":"abc","name":"山田太郎"}`)

	signature, timestamp, err := SignRequest(testSecret, body, now)
	assert.NoError(t, err)
	assert.Equal(t, "2026-10-15T12:00:00+09:00", timestamp)
	assert.Len(t, signature, 64)
	assert.Equal(t, signature, strings.ToLower(signature), "16進は小文字")

	assert.NoError(t, Verify(testSecret, body, signature, timestamp, now))
}

func TestVerify(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, jst)
	body := []byte(`{"applicant_id":13,"result":"match"}`)
	timestamp := "2026-10-15T11:55:00+09:00" // 5分前
	signature, err := Sign(testSecret, body, timestamp)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		secret    string
		body      []byte
		signature string
		timestamp string
		now       time.Time
		wantErr   error
	}{
		{"正しい署名と新しいタイムスタンプ", testSecret, body, signature, timestamp, now, nil},
		{"本体が1文字違う", testSecret, []byte(`{"applicant_id":14,"result":"match"}`), signature, timestamp, now, ErrSignatureMismatch},
		{"鍵が違う", "another-secret-another-secret-0123", body, signature, timestamp, now, ErrSignatureMismatch},
		{"タイムスタンプを差し替えた(署名は元のまま)", testSecret, body, signature, "2026-10-15T11:59:00+09:00", now, ErrSignatureMismatch},
		{"署名が空", testSecret, body, "", timestamp, now, ErrSignatureMissing},
		{"タイムスタンプが空", testSecret, body, signature, "", now, ErrSignatureMissing},
		{"署名が短い", testSecret, body, signature[:32], timestamp, now, ErrSignatureMismatch},
		{"署名が大文字の16進: 同じ値なので受け付ける(受信側は大文字小文字を区別しない)", testSecret, body, strings.ToUpper(signature), timestamp, now, nil},
		{"署名の途中だけ大文字: 同じ値なので受け付ける", testSecret, body, signature[:20] + strings.ToUpper(signature[20:40]) + signature[40:], timestamp, now, nil},
		{"署名の1文字を別の16進へ変える", testSecret, body, flipFirstHex(signature), timestamp, now, ErrSignatureMismatch},
		{"鍵が未設定", "", body, signature, timestamp, now, ErrSecretNotSet},
		{"9分59秒前はOK", testSecret, body, mustSign(t, body, "2026-10-15T11:50:01+09:00"), "2026-10-15T11:50:01+09:00", now, nil},
		{"ちょうど10分前は期限切れ", testSecret, body, mustSign(t, body, "2026-10-15T11:50:00+09:00"), "2026-10-15T11:50:00+09:00", now, ErrTimestampExpired},
		{"1時間前の正規リクエストの再送", testSecret, body, mustSign(t, body, "2026-10-15T11:00:00+09:00"), "2026-10-15T11:00:00+09:00", now, ErrTimestampExpired},
		{"未来側も10分以上離れていれば弾く", testSecret, body, mustSign(t, body, "2026-10-15T12:10:00+09:00"), "2026-10-15T12:10:00+09:00", now, ErrTimestampExpired},
		{"未来側9分59秒先はOK(時計のずれを許容)", testSecret, body, mustSign(t, body, "2026-10-15T12:09:59+09:00"), "2026-10-15T12:09:59+09:00", now, nil},
		{"形式が違うタイムスタンプ(署名は一致)", testSecret, body, mustSign(t, body, "2026/10/15 11:59:00"), "2026/10/15 11:59:00", now, ErrTimestampInvalid},
		{"Z表記(UTC)のRFC3339: 時刻として比較する", testSecret, body, mustSign(t, body, "2026-10-15T02:55:00Z"), "2026-10-15T02:55:00Z", now, nil},
		{"小数秒付きのRFC3339: 受け付ける", testSecret, body, mustSign(t, body, "2026-10-15T11:55:00.123+09:00"), "2026-10-15T11:55:00.123+09:00", now, nil},
		{"Z表記でも10分以上前なら期限切れ", testSecret, body, mustSign(t, body, "2026-10-15T02:40:00Z"), "2026-10-15T02:40:00Z", now, ErrTimestampExpired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Verify(tt.secret, tt.body, tt.signature, tt.timestamp, tt.now)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestVerify_OffsetTimestampIsComparedAsAnInstant(t *testing.T) {
	// +09:00以外のオフセットで来ても、同じ時刻なら時刻として比較される(文字列比較ではない)。
	now := time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC) // JST 12:00
	timestamp := "2026-10-15T02:55:00+00:00"             // JST 11:55 と同じ時刻
	signature := mustSign(t, nil, timestamp)
	assert.NoError(t, Verify(testSecret, nil, signature, timestamp, now))
}

func mustSign(t *testing.T, body []byte, timestamp string) string {
	t.Helper()
	signature, err := Sign(testSecret, body, timestamp)
	assert.NoError(t, err)
	return signature
}

// flipFirstHex は、16進文字列の先頭1文字を別の16進へ変える。
func flipFirstHex(signature string) string {
	if signature[0] == '0' {
		return "1" + signature[1:]
	}
	return "0" + signature[1:]
}
