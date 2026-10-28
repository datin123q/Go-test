package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type INexwayService interface {
	GetEkycResultByApplicantId(ctx context.Context, applicantId string) (*EkycResultOutput, error)
}

var _ INexwayService = (*NexwayService)(nil)

type NexwayService struct {
}

func NewNexwayService() INexwayService {
	return &NexwayService{}
}

// NexwayApiTimeoutは照会APIの呼び出しに設定するタイムアウト。本番ログの実測
// (2026-09-17時点、直近3日分)では応答は常にミリ秒単位で返っており、10秒は
// 十分な余裕を持たせた値(MIR-1223)。ALBのアイドルタイムアウト(240秒)まで
// ハンドラを滞留させないための上限でもある。呼び出し元(jobs-scheduler側)が
// 受信リクエストのctxから独立した照会用ctxを作る際にも、同じ値を使う。
const NexwayApiTimeout = 10 * time.Second

// ErrNexwayUnavailable はnexwayへの通信自体が失敗したことを表す(接続エラー・
// タイムアウト等)。応答そのものが得られていないため、呼び出し元はフェイル
// オープン(Webhookの値をそのまま使う)の判断に使える(MIR-1223)。
var ErrNexwayUnavailable = errors.New("nexway: service unavailable")

// ErrNexwayInvalidResponse はnexwayから応答はあったが、内容が信頼できない
// (HTTPステータス異常・空ボディ・パース失敗等)ことを表す。応答が来ている
// 以上「相手が今は答えられない」とは言えないため、呼び出し元はフェイル
// クローズ(弾く)の判断に使う(MIR-1223)。
var ErrNexwayInvalidResponse = errors.New("nexway: invalid response")

func (s *NexwayService) GetEkycResultByApplicantId(ctx context.Context, applicantId string) (*EkycResultOutput, error) {
	var ekycResultOutput EkycResultOutput

	serviceUrl := serviceutils.GetServiceNexwayApplicantsAPI() + "/" + applicantId + "/include_ekyc"
	body, err := s.callNexwayApi(ctx, echo.GET, serviceUrl, "", true)
	if err != nil {
		return nil, err
	}

	parsedResult, sanitizedBody, parseErr := parseEkycResponseBody(body)
	logrus.WithFields(logrus.Fields{
		"api":  serviceUrl,
		"body": sanitizedBody,
		"time": time.Now(),
	}).Info("GetEkycResultByApplicantId - call nexway api")

	if parseErr != nil {
		logrus.WithFields(logrus.Fields{
			"api":      serviceUrl,
			"err":      parseErr,
			"time":     time.Now(),
			"response": sanitizedBody,
		}).Error("GetEkycResultByApplicantId - Fail unmarshal nexway response")
		return nil, errors.Wrapf(ErrNexwayInvalidResponse, "failed to unmarshal nexway response: %v", parseErr)
	}

	ekycResultOutput = *parsedResult
	sanitizedResult := SanitizeEkycResultOutputForLog(&ekycResultOutput)

	logrus.WithFields(logrus.Fields{
		"api":              serviceUrl,
		"ekycResultOutput": sanitizedResult,
		"time":             time.Now(),
	}).Info("GetEkycResultByApplicantId - call nexway api and parse success")

	return parsedResult, nil
}

func SanitizeEkycResultOutputForLog(result *EkycResultOutput) *EkycResultOutput {
	if result == nil {
		return nil
	}
	clone := *result
	clone.ApplicantImages = sanitizeApplicantImagesForLog(result.ApplicantImages)
	clone.MaskedImages = sanitizeApplicantImagesForLog(result.MaskedImages)
	return &clone
}

func sanitizeApplicantImagesForLog(images []ApplicantImageDto) []ApplicantImageDto {
	if images == nil {
		return nil
	}
	if len(images) == 0 {
		return []ApplicantImageDto{}
	}
	sanitized := make([]ApplicantImageDto, len(images))
	copy(sanitized, images)
	for i := range sanitized {
		sanitized[i].ImageData = ""
	}
	return sanitized
}

func parseEkycResponseBody(body []byte) (*EkycResultOutput, string, error) {
	if len(body) == 0 {
		return nil, "", errors.New("empty response body")
	}
	var result EkycResultOutput
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, "[unable to parse eKYC response for logging]", err
	}
	sanitized := SanitizeEkycResultOutputForLog(&result)
	jsonBytes, err := json.Marshal(sanitized)
	if err != nil {
		return &result, "[unable to marshal sanitized eKYC response]", err
	}
	return &result, string(jsonBytes), nil
}

func (s *NexwayService) getBuf(v any) (*bytes.Buffer, error) {
	encodedInput, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(encodedInput)

	return buf, nil
}

func (s *NexwayService) readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *NexwayService) callNexwayApi(ctx context.Context, method, url string, input any, needToken bool) ([]byte, error) {
	client := &http.Client{Timeout: NexwayApiTimeout}

	buf, err := s.getBuf(input)
	if err != nil {
		return nil, factory.ErrorJSONMarshal.New(err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		// URLが不正・未設定(ApplicantsApiが空など)。一時的な通信障害では
		// なく恒久的な設定不備なので、フェイルオープンにはしない。
		return nil, errors.Wrap(ErrNexwayInvalidResponse, err.Error())
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if needToken {
		req.Header.Set("Authorization", "Token "+serviceutils.GetServiceNexwayToken())
	}

	res, err := client.Do(req)
	if err != nil {
		// 接続失敗・タイムアウト・呼び出し元のctxキャンセル。応答そのものが
		// 得られていないため、フェイルオープンの判断材料になる。
		return nil, errors.Wrap(ErrNexwayUnavailable, err.Error())
	}
	body, err := s.readBody(res)
	if err != nil {
		// ヘッダ・ステータスコードは受け取れているので「応答はある」側。
		// フェイルオープンにはしない。
		return nil, errors.Wrap(ErrNexwayInvalidResponse, err.Error())
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		// 応答は得られているが内容を信頼できない(WAF/レート制限のエラー
		// ページ等の可能性がある)ため、フェイルクローズの判断材料になる。
		return nil, errors.Wrapf(ErrNexwayInvalidResponse, "unexpected status code %d", res.StatusCode)
	}

	return body, nil
}

type ProfileInput struct {
	ReferenceId string `json:"reference_id"`
	Name        string `json:"name"`
	Birthday    string `json:"birthday"`
	Address     string `json:"address"`
	Job         string `json:"job"`
	Purpose     string `json:"purpose"`
}

type (
	ApplicantImageDto = domain.NexwayApplicantImageDto
	EkycResult        = domain.NexwayEkycResult
	EkycError         = domain.NexwayEkycError
	EkycResultOutput  = domain.EkycResultOutput
)

/*
[EkycResultOutput 설명]
- /api/v1/applicants/:applicant_id/include_ekyc API 의 응답

------
DocumentType
1:運転免許証 (운전 면허증)
2:マイナンバーカード (마이 넘버 카드)
3:在留カード (체류 카드)
4:運転経歴証明書 (운전 경력 증명서)
5:特別永住者証明書 (특별 영주자 증명서)
6:住民基本台帳カード (주민 기본 대장 카드)
7:保険証 (보험증)
------
Status
- 신청 API(확인 서류 이미지 없음)를 이용하는 경우에만 표시됨
- applying: 신청중
- accepted: 확인 대기
- processing: 확인중
- completed: 확인 완료
------
Result
- match, unmatch
- eKYC 가미한 얼굴 대조 결과를 포함하여 최종 확인 결과
- '7: 보험증' 또는 얼굴 대조 여부가 false인 경우 얼굴 대조 결과가 없으므로 '얼굴 대조 이외의 대조 결과'와 동일한 값
------
RequireFaceMatch
- eKYC 의 결과
- false 일 때 セルフィー画像(셀피 화상)을 지정하면 에러
- true 일 때 본인 확인 서류가 '7.보험증' 인 경우 에러
------
ApplicantImages
- ImageType: front, back, diagonal, front_trim, back_trim, diagonal_trim, selfie
- ImageData: base64 인코딩된 이미지 데이터
- 본인 이미지는 신청 시 얼굴 대조 여부가 false인 경우는 존재하지 않음
- '2.마이 넘버 카드' 의 경우 뒷면과 뒷면 트리밍 이미지는 존재하지 않음
------
ResultWithoutFaceMatch
- match, unmatch
- 육안으로 확인한 결과
*/

/*
- 뒷면: '2.마이 넘버 카드' 일 경우 설정하면 에러
- 사선: 필수
- 본인 얼굴 이미지: '7.보험증' 일 경우 설정하면 에러
*/
