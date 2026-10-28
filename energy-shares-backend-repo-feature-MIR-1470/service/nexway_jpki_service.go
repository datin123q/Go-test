package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/nexwaysign"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

// 本人確認デジタル認証サービス・本人確認BPOサービス JPKI版のパス。
const (
	jpkiAuthResultPathFormat = "/api/verification/%s/result"
	jpkiBpoRegisterPath      = "/api/v2/applicants/profiles"
)

// 署名検証結果(本人確認デジタル認証サービス)の値。仕様書 p.17-19。
const (
	JpkiResultSucceeded = "succeeded"
	JpkiResultFailed    = "failed"
	JpkiResultError     = "error"
	JpkiResultCanceled  = "canceled"

	JpkiStatusCompleted = "completed"
)

// jpkiErrorCodeRequestNotFound は、検証結果取得APIの404で、request_idが存在しないことを表すエラーコード。
// 仕様書の訂正(ネクスウェイ、2026-10-23)で、404はこのコードだけと確認した。試験環境の実際の応答も
// {"error_code":"4040002","error_detail":"not found"} だった(2026-10-26)。
const jpkiErrorCodeRequestNotFound = "4040002"

var (
	// ErrNexwayJpkiNotConfigured は、JPKI方式のURL・トークン・署名用キーのいずれかが未設定で、
	// 通信前に止めたことを表す。ホ方式の設定へ流用はしない(別契約・別トークンのため)。
	// 一時的な通信障害ではなく設定不備なので、再送しても直らない。
	ErrNexwayJpkiNotConfigured = errors.New("nexway jpki: not configured")

	// ErrNexwayJpkiRequestNotFound は、検証依頼IDがデジタル認証サービスに存在しない(404・4040002)ことを表す。
	// 通知が正規のものなら起きない。
	ErrNexwayJpkiRequestNotFound = errors.New("nexway jpki: verification request not found")

	// ErrNexwayJpkiAlreadyRegistered は、BPOがその申請を既に受け付けている(4000010・4000011)
	// ことを表す。重複して送ったので新しい申請IDは得られないが、申請自体は有効。
	ErrNexwayJpkiAlreadyRegistered = errors.New("nexway jpki: applicant already registered")

	// ErrNexwayJpkiApplicantRejected は、BPOが申請内容の不備(4000008・4000009)で受け付けなかった
	// ことを表す。同じ内容で再送しても通らない。
	ErrNexwayJpkiApplicantRejected = errors.New("nexway jpki: applicant rejected as a bad request")
)

// IJpkiNexwayService は、JPKI方式でミラリタが呼ぶNexwayのAPIのうち、jobs-schedulerが使うもの(MIR-1470)。
//
// ホ方式のINexwayServiceとは契約・URL・トークンが別のため、別のインターフェースにしている。
// 署名用URLの発行(MIR-1469)はuser-api側(service/nexway/jpki_auth_svc.go)にある。
type IJpkiNexwayService interface {
	// GetSignatureResult は、検証結果取得API(GET /api/verification/:request_id/result)で
	// 署名検証の結果を取得する。Webhookの内容をそのまま信用しないための裏取りに使う。
	GetSignatureResult(ctx context.Context, requestId string) (*JpkiSignatureResult, error)
	// RegisterBpoApplicant は、確認依頼(申請)API JPKI版(eKYC履歴なし)で、BPOへ申請する。
	RegisterBpoApplicant(ctx context.Context, input JpkiBpoApplicantInput) (*JpkiBpoApplicantOutput, error)
}

var _ IJpkiNexwayService = (*JpkiNexwayService)(nil)

// JpkiSignatureResult は、検証結果取得APIのレスポンスのうち、ミラリタが使う項目だけ。
//
// このAPIのレスポンスには基本4情報(jpki_result: 氏名・生年月日・住所・性別)が含まれるが、
// ここでは受け取らない。構造体に項目が無いので、デコード時に捨てられる。ミラリタが
// 保存したり、ログへ出したりする情報を、最初から持たないための設計(MIR-612の方針)。
type JpkiSignatureResult struct {
	ReferenceId string `json:"reference_id"`
	RequestId   string `json:"request_id"`
	Status      string `json:"status"`
	Result      string `json:"result"`
	CompletedAt string `json:"completed_at"`
	// Errors は、署名の検証処理でのエラーコード(C01〜C09・E01・E05。仕様書 p.14)。
	// エラーが無い場合は空。形が想定と違っても、ほかの項目の読み取りは失敗させない。
	Errors JpkiSignatureErrors `json:"errors"`
}

// JpkiBpoApplicantInput は、確認依頼(申請)API JPKI版(eKYC履歴なし)のリクエスト(手引き p.13)。
// 職業(job)・目的(purpose)は任意項目のため送らない。
type JpkiBpoApplicantInput struct {
	ReferenceId string `json:"reference_id"` // URL発行依頼APIで設定した連携ID
	RequestId   string `json:"request_id"`   // デジタル認証サービスが発行した検証依頼ID
	Name        string `json:"name"`
	Birthday    string `json:"birthday"` // YYYY-MM-DD
	Address     string `json:"address"`  // 郵便番号なし
}

// JpkiBpoApplicantOutput は、確認依頼(申請)APIのレスポンス(手引き p.13)。
type JpkiBpoApplicantOutput struct {
	ApplicantId int64  `json:"applicant_id"`
	ReferenceId string `json:"reference_id"`
	AcceptedAt  string `json:"accepted_at"`
}

// JpkiNexwayService は IJpkiNexwayService の実装。
type JpkiNexwayService struct {
	httpClient *http.Client
	now        func() time.Time
}

func NewJpkiNexwayService() IJpkiNexwayService {
	return &JpkiNexwayService{
		httpClient: &http.Client{Timeout: NexwayApiTimeout},
		now:        time.Now,
	}
}

// GetSignatureResult は、検証結果取得APIを呼ぶ。
//
// 通信できない(接続・タイムアウト)ときは ErrNexwayUnavailable、応答はあるが信頼できない
// (HTTPステータス異常・パース失敗)ときは ErrNexwayInvalidResponse を返す。呼び出し元は
// どちらも「いまは確認できない」として扱う。404でerror_codeが4040002(request_idが存在しない)の
// ときだけ ErrNexwayJpkiRequestNotFound。
// レスポンス本文には基本4情報が入っているので、エラーにもログにも本文を含めない。
func (s *JpkiNexwayService) GetSignatureResult(ctx context.Context, requestId string) (*JpkiSignatureResult, error) {
	baseUrl := strings.TrimRight(serviceutils.GetServiceNexwayJpkiAuthBaseUrl(), "/")
	token := serviceutils.GetServiceNexwayJpkiAuthToken()
	if baseUrl == "" || token == "" {
		return nil, errors.Wrap(ErrNexwayJpkiNotConfigured, "digital auth base url or token is empty")
	}
	if requestId == "" {
		return nil, errors.Wrap(ErrNexwayInvalidResponse, "request_id is empty")
	}

	endpoint := baseUrl + fmt.Sprintf(jpkiAuthResultPathFormat, url.PathEscape(requestId))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(ErrNexwayInvalidResponse, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+token)

	body, status, err := s.do(req)
	if err != nil {
		return nil, err
	}
	switch {
	case status == http.StatusNotFound && apiErrorCode(body) == jpkiErrorCodeRequestNotFound:
		return nil, errors.Wrapf(ErrNexwayJpkiRequestNotFound, "error_code=%s", jpkiErrorCodeRequestNotFound)
	case status < 200 || status >= 300:
		// 4040002以外の404(ベースURLの設定誤りで、別のサーバーが返した404など)は、request_idが
		// 存在しない証拠にならない。存在しないとして扱うと、定期ジョブが記録を失敗にして会員へメールを
		// 送ってしまうため、確認できない応答として扱い、あとで確認し直す。
		return nil, errors.Wrapf(ErrNexwayInvalidResponse, "unexpected status code %d (error_code=%s)", status, apiErrorCode(body))
	}

	var result JpkiSignatureResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.Wrap(ErrNexwayInvalidResponse, "failed to parse digital auth result response")
	}
	return &result, nil
}

// RegisterBpoApplicant は、BPOへ申請する。署名検証オプション(HMAC-SHA256)が有効な契約のため、
// 送信する本体そのもの(バイト列)に署名を付ける。署名用キーが未設定なら、署名なしで送らず
// 通信前に ErrNexwayJpkiNotConfigured で止める。
//
// 返すエラーは次の4つに分かれ、呼び出し元は扱いを変える。
//   - ErrNexwayJpkiAlreadyRegistered: 既に受付済み。成功と同じ扱いでよい
//   - ErrNexwayJpkiApplicantRejected: 申請内容の不備。再送しても通らない
//   - ErrNexwayUnavailable: 通信できなかった。再送してよい
//   - ErrNexwayInvalidResponse: 応答はあったが受け付けられていない(認証・署名・IP制限・
//     メンテナンス・サーバーエラー・想定外の応答)。申請が受け付けられたかは不明なので
//     再送してよい(重複はAlreadyRegisteredで検知できる)
//
// 申請内容(氏名・住所など)は個人情報なので、エラーにもログにも本文を含めない。
func (s *JpkiNexwayService) RegisterBpoApplicant(ctx context.Context, input JpkiBpoApplicantInput) (*JpkiBpoApplicantOutput, error) {
	baseUrl := strings.TrimRight(serviceutils.GetServiceNexwayJpkiBpoBaseUrl(), "/")
	token := serviceutils.GetServiceNexwayJpkiBpoToken()
	secret := serviceutils.GetServiceNexwayJpkiBpoSigningSecret()
	if baseUrl == "" || token == "" || secret == "" {
		return nil, errors.Wrap(ErrNexwayJpkiNotConfigured, "bpo base url, token or signing secret is empty")
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, errors.Wrap(ErrNexwayInvalidResponse, "failed to marshal bpo applicant request")
	}
	signature, timestamp, err := nexwaysign.SignRequest(secret, body, s.now())
	if err != nil {
		return nil, errors.Wrap(ErrNexwayJpkiNotConfigured, err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseUrl+jpkiBpoRegisterPath, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(ErrNexwayInvalidResponse, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set(nexwaysign.HeaderSignature, signature)
	req.Header.Set(nexwaysign.HeaderTime, timestamp)

	respBody, status, err := s.do(req)
	if err != nil {
		return nil, err
	}

	if status < 200 || status >= 300 {
		code := apiErrorCode(respBody)
		switch code {
		case "4000010", "4000011":
			return nil, errors.Wrapf(ErrNexwayJpkiAlreadyRegistered, "error_code=%s", code)
		case "4000008", "4000009":
			return nil, errors.Wrapf(ErrNexwayJpkiApplicantRejected, "error_code=%s", code)
		}
		return nil, errors.Wrapf(ErrNexwayInvalidResponse, "unexpected status code %d (error_code=%s)", status, code)
	}

	var output JpkiBpoApplicantOutput
	if err := json.Unmarshal(respBody, &output); err != nil {
		return nil, errors.Wrap(ErrNexwayInvalidResponse, "failed to parse bpo applicant response")
	}
	if output.ApplicantId == 0 {
		// 201なのに申請IDが無い。受け付けられたか判断できないので、成功とは扱わない。
		return nil, errors.Wrap(ErrNexwayInvalidResponse, "bpo applicant response has no applicant_id")
	}
	return &output, nil
}

// do は、リクエストを送って本文とHTTPステータスを返す。2xx以外の判断は呼び出し元が行う。
func (s *JpkiNexwayService) do(req *http.Request) ([]byte, int, error) {
	res, err := s.httpClient.Do(req)
	if err != nil {
		// 接続失敗・タイムアウト・ctxキャンセル。応答そのものが得られていない。
		// url.Errorはリクエストのクエリ文字列等を含みうるが、ここのURLには個人情報を載せない。
		return nil, 0, errors.Wrap(ErrNexwayUnavailable, err.Error())
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, errors.Wrap(ErrNexwayInvalidResponse, err.Error())
	}
	return body, res.StatusCode, nil
}

// apiErrorCode は、エラーレスポンスの error_code だけを取り出す。本文のその他の項目
// (details等)は出さない。パースできなければ空文字。
func apiErrorCode(body []byte) string {
	var apiError struct {
		ErrorCode string `json:"error_code"`
	}
	_ = json.Unmarshal(body, &apiError)
	return apiError.ErrorCode
}
