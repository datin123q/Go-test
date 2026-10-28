// Package nexwaysign は、本人確認BPOサービス JPKI版の署名検証オプション(HMAC-SHA256)の
// 署名の作成と検証を行う(MIR-1470。手引き p.45-47)。
//
// 署名の対象は「リクエスト本体 + 署名作成時のタイムスタンプ」で、鍵はAPIトークンとは別に
// 発行される署名用のシークレットキー。結果は16進(小文字)で、ヘッダー x-kyc-signature に、
// タイムスタンプは x-kyc-time に付ける。本体が無いリクエスト(結果取得API)は
// タイムスタンプだけを署名する。
//
// このオプションは本人確認BPOサービス側のAPIだけが対象で、本人確認デジタル認証サービスの
// API(URL発行・検証結果取得)には適用されない。
package nexwaysign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	// HeaderSignature・HeaderTime は、署名とタイムスタンプを載せるHTTPヘッダー名。
	HeaderSignature = "x-kyc-signature"
	HeaderTime      = "x-kyc-time"

	// TimestampLayout は x-kyc-time の形式(YYYY-MM-DDThh:mm:ss+09:00)。
	TimestampLayout = "2006-01-02T15:04:05-07:00"

	// MaxAge は、受信した署名のタイムスタンプを有効とみなす時間。手引き p.47 は
	// 「10分程度経過していればエラー」を勧めている(再生攻撃対策)。
	MaxAge = 10 * time.Minute
)

var (
	// ErrSecretNotSet は、署名用のシークレットキーが未設定であることを表す。
	// 未設定のまま署名を作る・検証すると、空の鍵で計算した推測可能な署名が通ってしまうため、
	// 呼び出し側へ必ずエラーで返す。
	ErrSecretNotSet = errors.New("nexwaysign: secret is not set")
	// ErrSignatureMissing は、検証対象のリクエストに署名またはタイムスタンプが無いことを表す。
	ErrSignatureMissing = errors.New("nexwaysign: signature or timestamp is missing")
	// ErrTimestampInvalid は、タイムスタンプがRFC3339(手引きの形式YYYY-MM-DDThh:mm:ss+09:00を含む)で解析できないことを表す。
	ErrTimestampInvalid = errors.New("nexwaysign: timestamp is invalid")
	// ErrTimestampExpired は、タイムスタンプがMaxAgeより古い(または未来すぎる)ことを表す。
	ErrTimestampExpired = errors.New("nexwaysign: timestamp is expired")
	// ErrSignatureMismatch は、再計算した署名がヘッダーの署名と一致しないことを表す。
	ErrSignatureMismatch = errors.New("nexwaysign: signature does not match")
)

// jst は、署名のタイムスタンプに使う固定オフセット(+09:00)。
// Asia/Tokyoのタイムゾーンデータが無い環境でも同じ表記になるよう、固定オフセットを使う。
var jst = time.FixedZone("JST", 9*60*60)

// FormatTime は、署名作成時のタイムスタンプを x-kyc-time の形式(+09:00)に整形する。
func FormatTime(t time.Time) string {
	return t.In(jst).Format(TimestampLayout)
}

// Sign は、body+timestamp のHMAC-SHA256を16進(小文字)で返す。
// 結果取得API(GET)のように本体が無い場合は、bodyにnilを渡す。
func Sign(secret string, body []byte, timestamp string) (string, error) {
	if secret == "" {
		return "", ErrSecretNotSet
	}
	return sign(secret, body, timestamp), nil
}

func sign(secret string, body []byte, timestamp string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	mac.Write([]byte(timestamp))
	return hex.EncodeToString(mac.Sum(nil))
}

// SignRequest は、いま(now)の時刻でbodyに署名し、ヘッダーへ付ける値(署名, タイムスタンプ)を返す。
func SignRequest(secret string, body []byte, now time.Time) (signature, timestamp string, err error) {
	timestamp = FormatTime(now)
	signature, err = Sign(secret, body, timestamp)
	if err != nil {
		return "", "", err
	}
	return signature, timestamp, nil
}

// Verify は、受信したリクエストの署名を検証する(手引き p.47)。
//
// 署名の一致を先に、タイムスタンプの鮮度を後に確認する。比較は一定時間で行い、
// タイミング攻撃で署名を1文字ずつ推測されないようにする。
// 署名が一致していても、タイムスタンプがnowからMaxAge以上離れていれば弾く。
// 過去の正規リクエストの再送(再生攻撃)を防ぐため。未来側も同じ幅で弾く
// (ネクスウェイ側と時計がずれていても、10分以内のずれは許容する)。
//
// 受信側は、送信側の形式の細かい違いを許容する。手引きは16進の小文字、時刻を
// YYYY-MM-DDThh:mm:ss+09:00と定めるが、送信側の実装が大文字の16進や、Z・小数秒付きの
// RFC3339で送ってきても、通知がすべて弾かれて結果が失われるのを避けるため、署名は大文字小文字を
// 区別せず比較し、時刻はRFC3339で解析する。署名の計算はヘッダーの時刻の文字列をそのまま使うため、
// これで安全性は落ちない(16進の大文字小文字は同じ値、RFC3339の解析は鮮度の確認にだけ使う)。
// 送信側(Sign・SignRequest)は、手引きの形式で作る。
func Verify(secret string, body []byte, signature, timestamp string, now time.Time) error {
	if secret == "" {
		return ErrSecretNotSet
	}
	if signature == "" || timestamp == "" {
		return ErrSignatureMissing
	}

	expected := sign(secret, body, timestamp)
	// 16進文字列同士を比較する。hmac.Equalは長さが違えば即falseを返すが、
	// 長さは公開情報(SHA-256なら64文字固定)なので漏れて困る情報ではない。
	if !hmac.Equal([]byte(expected), []byte(strings.ToLower(signature))) {
		return ErrSignatureMismatch
	}

	signedAt, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return errors.Wrap(ErrTimestampInvalid, err.Error())
	}
	age := now.Sub(signedAt)
	if age < 0 {
		age = -age
	}
	if age >= MaxAge {
		return ErrTimestampExpired
	}
	return nil
}
