package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/nexwaysign"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

// 署名検証に失敗した受信は、本体(個人情報を含みうる)を出さず、理由だけをログに残す。
const nexwaySignatureLogMessage = "nexway bpo webhook: signature verification"

// nexwayWebhookMaxBodyBytes は、結果通知の本文の上限。通知は数KB程度で、1MBあれば十分余裕がある。
const nexwayWebhookMaxBodyBytes = 1 << 20

// readNexwayWebhookBody は、ネクスウェイからの通知(認証の無い公開口)の本文を、上限まで読む。
// 上限+1バイトまで読み、超えていれば 413 を返す(全部読むと、巨大な本文でメモリを使い切られる)。
// 読めなければ 400。読めたときは、後続のBindが読めるよう本文を戻し、0を返す。
// BPOの結果通知(/nexway/bpo/webhook)と、検証結果通知(/nexway/jpki/webhook)で共通。
func readNexwayWebhookBody(c echo.Context) ([]byte, int) {
	body, err := io.ReadAll(io.LimitReader(c.Request().Body, nexwayWebhookMaxBodyBytes+1))
	if err != nil {
		return nil, http.StatusBadRequest
	}
	if int64(len(body)) > nexwayWebhookMaxBodyBytes {
		return nil, http.StatusRequestEntityTooLarge
	}
	c.Request().Body = io.NopCloser(bytes.NewReader(body))
	return body, 0
}

// nexwaySignedEkycType は、署名検証オプションの対象となる通知のekyc_type。
// MIR-1468のdomain.NexwayEkycTypeJpkiと同じ値だが、PRを独立に保つため参照しない。
const nexwaySignedEkycType = "jpki"

// VerifyNexwayBpoSignature は、本人確認BPOサービス JPKI版の結果通知(Webhook)の署名を検証する
// (MIR-1470。手引き p.47「署名検証オプション【受信時】」)。
//
// 対象はekyc_type=jpkiの通知だけ。ホ方式(ProTech ID Checker)の通知には署名オプションが無いため、
// そのまま通す。ekyc_typeは署名の検証前に本体から読むので、送信元は「ホ方式」と偽って検証を
// 避けられる。ただし、その場合は後続の処理がProTechへ照会するため、BPO JPKI版の申請は見つからず
// 弾かれる(MIR-1223・MIR-1468のフェイルクローズ)。署名の検証はこの二重目の守りに加える層。
//
// 結果:
//   - 署名が正しい: 通す(本体は後続のハンドラでもう一度読めるよう戻す)
//   - 署名が無い・不一致・古い(10分以上): 401
//   - 署名用シークレットが未設定: 500。署名検証オプションが有効な契約なのに設定が無いのは
//     こちらの不備で、黙って通すと検証なしで受けることになる。BPOの再送(10分×5、60分×5)で
//     設定後に受け直せる
func VerifyNexwayBpoSignature() echo.MiddlewareFunc {
	return verifyNexwayBpoSignature(serviceutils.GetServiceNexwayJpkiBpoSigningSecret, time.Now)
}

func verifyNexwayBpoSignature(secretFn func() string, now func() time.Time) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 認証の無い公開口のため、本文の大きさに上限を設ける。
			body, status := readNexwayWebhookBody(c)
			if status != 0 {
				return c.JSON(status, ApiResult{Success: false, Result: false})
			}

			if !isJpkiWebhookBody(body) {
				return next(c)
			}

			secret := secretFn()
			err := nexwaysign.Verify(
				secret,
				body,
				c.Request().Header.Get(nexwaysign.HeaderSignature),
				c.Request().Header.Get(nexwaysign.HeaderTime),
				now(),
			)
			if err == nil {
				return next(c)
			}

			fields := logrus.Fields{"reason": err.Error(), "time": time.Now().String()}
			if errors.Is(err, nexwaysign.ErrSecretNotSet) {
				logrus.WithFields(fields).Error(nexwaySignatureLogMessage + ": signing secret is not set")
				return c.JSON(http.StatusInternalServerError, false)
			}
			logrus.WithFields(fields).Warn(nexwaySignatureLogMessage + ": rejected")
			return c.JSON(http.StatusUnauthorized, ApiResult{Success: false, Result: false})
		}
	}
}

// isJpkiWebhookBody は、本体のekyc_typeがjpkiかを返す。読めない本体はfalse
// (ホ方式として後続に渡し、後続のBindが400にする)。
func isJpkiWebhookBody(body []byte) bool {
	var probe struct {
		EkycType string `json:"ekyc_type"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return false
	}
	return strings.EqualFold(probe.EkycType, nexwaySignedEkycType)
}
