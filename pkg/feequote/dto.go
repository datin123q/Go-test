package feequote

// WithdrawalFeeDetail represents fee per withdrawal in the backend-job contract.
type WithdrawalFeeDetail struct {
	WithdrawalID int64 `json:"withdrawalId"`
	TransferFee  int64 `json:"transferFee"`
}

// WithdrawalFeeBulkResponse is the response schema for fee estimation between backend and job.
//
// IsFallbackは、GMOあおぞら銀行の手数料見積もりAPIが失敗し、固定値へ
// フォールバックした応答であることを示す(MIR-1172)。見積もりが成功した
// ときはfalse。
type WithdrawalFeeBulkResponse struct {
	TotalFee   int64                 `json:"totalFee"`
	Details    []WithdrawalFeeDetail `json:"details"`
	IsFallback bool                  `json:"is_fallback"`
}
