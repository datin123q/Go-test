package controllers

type NexwayEkycError struct {
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

type NexwayEkycResult struct {
	ApplicantId            int64             `json:"applicant_id"`              // 申請ID(申請に対して一意に振られるID)
	ReferenceId            string            `json:"reference_id"`              // 連携ID
	DocumentType           string            `json:"document_type"`             // 本人確認書類
	AcceptedAt             string            `json:"accepted_at"`               // 受付日時（YYYY-MM-DDThh:mm:ss+09:00）
	Status                 string            `json:"status"`                    // completed
	RequireFaceMatch       bool              `json:"require_face_match"`        // 顔照合有無
	ResultWithoutFaceMatch string            `json:"result_without_face_match"` // 顔照合以外の確認結果 (match, unmatch)
	Result                 string            `json:"result"`                    // 顔照合を含む照合結果 (match, unmatch)
	CompletedAt            string            `json:"completed_at"`              // 確認完了日時（YYYY-MM-DDThh:mm:ss+09:00）
	VerificationErrors     []NexwayEkycError `json:"verification_errors"`       // MG項目
	OcrValidationErrors    []NexwayEkycError `json:"ocr_validation_errors"`     // OCR整合性チェックエラー
	RiskDetect             RiskDetect        `json:"risk_detect"`               // リスク情報検索情報
	OverallResult          string            `json:"overall_result"`            // 総合結果(本人確認+リスク情報検索の結果)(cleared(該当なし)detected(確認通過(リスクあり))unmatch(不一致))
}

type CallTestWebhookRequest struct {
	Email      string `json:"email" swagger:"required,desc(이메일)"`
	IsApproved bool   `json:"is_approved" swagger:"required,desc(eKYC 통과 여부, true: 승인, false: 미승인)"`
}

type RiskDetect struct {
	DetectedAt string `json:"detected_at"` // リスク情報検索日時（YYYY-MM-DDThh:mm:ss+09:00）
	Result     string `json:"result"`      // リスク情報検索結果（cleared(該当なし)detected(リスクあり)not_checked(実施なし)）
}

/*
------------
DocumentType
1:運転免許証 (운전 면허증)
2:マイナンバーカード (마이 넘버 카드)
3:在留カード (체류 카드)
4:運転経歴証明書 (운전 경력 증명서)
5:特別永住者証明書 (특별 영주자 증명서)
6:住民基本台帳カード (주민 기본 대장 카드)
7:保険証 (보험증)
------------
RequireFaceMatch
- true: eKYC나 BPO 중 하나로 얼굴 대조를 실시하고 있습니다.
- false: 얼굴 대조를 실시하지 않았습니다.
------------
ResultWithoutFaceMatch
- match, unmatch
- 육안으로 확인한 결과입니다.
------------
Result
- match, unmatch
- eKYC 를 가미한 얼굴 대조 결과를 포함하여 최종 확인 결과입니다.
- '7: 보험증' 또는 얼굴 대조 여부가 false인 경우 얼굴 대조 결과가 없으므로 '얼굴 대조 이외의 대조 결과'와 동일한 값입니다.
------------
*/
