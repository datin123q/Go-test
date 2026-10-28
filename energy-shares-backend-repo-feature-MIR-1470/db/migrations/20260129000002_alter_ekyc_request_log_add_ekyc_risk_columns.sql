-- +goose Up
-- eKYC照合結果カラム追加
ALTER TABLE ekyc_request_log
    ADD COLUMN ekyc_front_match VARCHAR(50) NULL COMMENT 'eKYC表面照合結果',
    ADD COLUMN ekyc_thickness_match VARCHAR(50) NULL COMMENT 'eKYC厚み照合結果',
    ADD COLUMN ekyc_live_match VARCHAR(50) NULL COMMENT 'eKYC生体検知結果',
    ADD COLUMN ekyc_end_at DATETIME NULL COMMENT 'eKYC終了日時';

-- リスク情報検索結果カラム追加
ALTER TABLE ekyc_request_log
    ADD COLUMN risk_detect_detected_at DATETIME NULL COMMENT 'リスク情報検索日時',
    ADD COLUMN risk_detect_result VARCHAR(50) NULL COMMENT 'リスク情報検索結果 (cleared, detected, not_checked)';

-- +goose Down
ALTER TABLE ekyc_request_log
    DROP COLUMN risk_detect_result,
    DROP COLUMN risk_detect_detected_at,
    DROP COLUMN ekyc_end_at,
    DROP COLUMN ekyc_live_match,
    DROP COLUMN ekyc_thickness_match,
    DROP COLUMN ekyc_front_match;
