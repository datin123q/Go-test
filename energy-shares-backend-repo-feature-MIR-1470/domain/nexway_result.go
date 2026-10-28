package domain

// NexwayImageType represents the type of identification image from Nexway API
type NexwayImageType string

const (
	ImageTypeFront        NexwayImageType = "front"
	ImageTypeBack         NexwayImageType = "back"
	ImageTypeDiagonal     NexwayImageType = "diagonal"
	ImageTypeFrontTrim    NexwayImageType = "front_trim"
	ImageTypeBackTrim     NexwayImageType = "back_trim"
	ImageTypeDiagonalTrim NexwayImageType = "diagonal_trim"
	ImageTypeSelfie       NexwayImageType = "selfie"
)

// Valid checks if the ImageType is valid according to Nexway API specification
func (it NexwayImageType) Valid() bool {
	switch it {
	case ImageTypeFront, ImageTypeBack, ImageTypeDiagonal,
		ImageTypeFrontTrim, ImageTypeBackTrim, ImageTypeDiagonalTrim,
		ImageTypeSelfie:
		return true
	}
	return false
}

// String returns the string representation of ImageType
func (it NexwayImageType) String() string {
	return string(it)
}

// IsTrimmed checks if the image is a trimmed version
func (it NexwayImageType) IsTrimmed() bool {
	return it == ImageTypeFrontTrim ||
		it == ImageTypeBackTrim ||
		it == ImageTypeDiagonalTrim
}

// IsDocument checks if the image is a document (not selfie)
func (it NexwayImageType) IsDocument() bool {
	return it != ImageTypeSelfie
}

// RequiresFaceMatch checks if this image type is used for face matching
func (it NexwayImageType) RequiresFaceMatch() bool {
	return it == ImageTypeSelfie
}

// NexwayRiskDetect models the risk detection payload.
type NexwayRiskDetect struct {
	DetectedAt string `json:"detected_at"`
	Result     string `json:"result"`
}

// NexwayEkycError models error details returned by Nexway.
type NexwayEkycError struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"error_message"`
}

// NexwayApplicantImageDto represents an identification image provided to Nexway.
type NexwayApplicantImageDto struct {
	ImageType NexwayImageType `json:"image_type"`
	ImageData string          `json:"image_data"`
}

// NexwayEkycResult contains biometric matching results.
type NexwayEkycResult struct {
	FrontMatch     string `json:"front_match"`
	ThicknessMatch string `json:"thickness_match"`
	LiveMatch      string `json:"live_match"`
	EkycEndAt      string `json:"ekyc_end_at"`
}

// NexwayOperator 確認者情報
type NexwayOperator struct {
	CompanyId      int64  `json:"company_id"`       // 確認会社ID
	CompanyName    string `json:"company_name"`     // 確認会社名
	UserId         int64  `json:"user_id"`          // 確認者ID
	UserName       string `json:"user_name"`        // 確認者名
	LoginId        string `json:"login_id"`         // 確認者ログインID
	OperatorType   string `json:"operator_type"`    // 確認者権限(first, second, sv)
	OperationEndAt string `json:"operation_end_at"` // BPO終了日時
}

// EkycResultOutput expresses the full Nexway eKYC response payload.
type EkycResultOutput struct {
	ApplicantId            int64                  `json:"applicant_id"`               // 確認依頼（申請）ID
	ReferenceId            string                 `json:"reference_id"`               // 連携ID
	DocumentType           string                 `json:"document_type"`              // 本人確認書類タイプ
	RequireFaceMatch       bool                   `json:"require_face_match"`         // 顔照合有無
	AcceptedAt             string                 `json:"accepted_at"`                // 受付日時
	Status                 string                 `json:"status"`                     // ステータス
	EkycType               string                 `json:"ekyc_type"`                  // eKYC方式
	ResultWithoutFaceMatch string                 `json:"result_without_face_match"`  // 顔照合以外の確認結果
	Result                 string                 `json:"result"`                     // 顔照合を含む確認結果
	CompletedAt            string                 `json:"completed_at"`               // 確認完了日時
	VerificationErrors     []NexwayEkycError         `json:"verification_errors"`        // 確認エラー項目
	MaskedImages           []NexwayApplicantImageDto `json:"masked_images,omitempty"`    // マスキング画像
	Operators              []NexwayOperator          `json:"operators,omitempty"`        // オペレーターID一覧
	RiskDetect             *NexwayRiskDetect         `json:"risk_detect,omitempty"`      // リスク情報検索情報
	OcrValidationErrors    []NexwayEkycError         `json:"ocr_validation_errors"`      // OCR機械判定エラー項目
	EkycResult             *NexwayEkycResult         `json:"ekyc_result,omitempty"`      // eKYC照合結果
	ApplicantImages        []NexwayApplicantImageDto `json:"applicant_images,omitempty"` // 確認書類画像

	// fail時のみ
	ErrorCode string   `json:"error_code,omitempty"`
	Message   string   `json:"message,omitempty"`
	Details   []string `json:"details,omitempty"`
}
