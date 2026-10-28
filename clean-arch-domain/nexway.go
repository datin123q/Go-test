package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type NexwayService interface {
	GetEkycResultByApplicantId(context.Context, string) (NexwayEkycResultOutput, error)
}

type NexwayProfileInput struct {
	ReferenceId string `json:"reference_id"`
	Name        string `json:"name"`
	Birthday    string `json:"birthday"`
	Address     string `json:"address"`
	Job         string `json:"job"`
	Purpose     string `json:"purpose"`
}

type NexwayProfileOutput struct {
	ApplicantId int64    `json:"applicant_id"`
	ReferenceId string   `json:"reference_id"`
	AcceptedAt  string   `json:"accepted_at"`
	ErrorCode   string   `json:"error_code"`
	Message     string   `json:"message"`
	Details     []string `json:"details"`
}

// NexwayEkycOutput /api/v1/applicants/:applicant_id/include_ekyc API 의 응답
type NexwayEkycOutput struct {
	Result  string `json:"result"`
	EkycUrl string `json:"ekycUrl"`

	ErrorCode string `json:"errorCode"`
}

type NexwayApplicantImageDto struct {
	ImageType domain.NexwayImageType `json:"image_type"`
	ImageData string                 `json:"image_data"`
}

type NexwayEkycResult struct {
	FrontMatch     string `json:"front_match"`
	ThicknessMatch string `json:"thickness_match"`
	LiveMatch      string `json:"live_match"`
	EkycEndAt      string `json:"ekyc_end_at"`
}

type NexwayEkycError struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"error_message"`
}

type NexwayEkycResultOutput struct {
	// success
	ApplicantId            int64                  `json:"applicant_id"`
	ReferenceId            string                 `json:"reference_id"`
	DocumentType           string                 `json:"document_type"` // 7개
	Status                 string                 `json:"status"`        // applying, accepted, processing, completed
	Result                 string                 `json:"result"`        // match, unmatch
	EkycResult             NexwayEkycResult       `json:"ekyc_result"`
	RequireFaceMatch       bool                   `json:"require_face_match"`
	AcceptedAt             string                 `json:"accepted_at"`               // YYYY-MM-DDThh:mm:ss+09:00
	ResultWithoutFaceMatch string                 `json:"result_without_face_match"` // match, unmatch
	CompletedAt            string                 `json:"completed_at"`              // YYYY-MM-DDThh:mm:ss+09:00
	VerificationErrors     []NexwayEkycError         `json:"verification_errors"`
	OcrValidationErrors    []NexwayEkycError         `json:"ocr_validation_errors"`
	MaskedImages           []NexwayApplicantImageDto `json:"masked_images"`    // front, back, diagonal
	ApplicantImages        []NexwayApplicantImageDto `json:"applicant_images"` // front, back, diagonal, front_trim, back_trim, diagonal_trim, selfie

	// fail
	ErrorCode string   `json:"error_code"`
	Message   string   `json:"message"`
	Details   []string `json:"details"`
}

/*
------
DocumentType
1:運転免許証 (運転免許証)
2:マイナンバーカード (マイナンバーカード)
3:在留カード (在留カード)
4:運転経歴証明書 (運転経歴証明書)
5:特別永住者証明書 (特別永住者証明書)
6:住民基本台帳カード (住民基本台帳カード)
7:保険証 (保険証)
------
Status
- 申請API（確認書類画像なし）を使用する場合のみ表示されます
- applying: 申請中
- accepted: 確認待ち
- processing: 確認中
- completed: 確認完了
------
Result
- match, unmatch
- eKYC付き顔照合結果を含む最終確認結果
- '7: 保険証' または顔照合有無がfalseの場合、顔照合結果がないため '顔照合以外の照合結果' と同じ値
------
RequireFaceMatch
- eKYCの結果
- falseの場合、セルフィー画像(セルフィー画像)を指定するとエラー
- trueの場合、本人確認書類が '7.保険証' の場合はエラー
------
ApplicantImages
- ImageType: front, back, diagonal, front_trim, back_trim, diagonal_trim, selfie
- ImageData: base64エンコードされた画像データ
- 本人画像は申請時に顔照合有無がfalseの場合は存在しない
- '2.マイナンバーカード' の場合、裏面と裏面トリミング画像は存在しない
------
ResultWithoutFaceMatch
- match, unmatch
- 目視で確認した結果
*/

/*
- 裏面: '2.マイナンバーカード' の場合に設定するとエラー
- 斜線: 必須
- 本人顔画像: '7.保険証' の場合に設定するとエラー
*/
