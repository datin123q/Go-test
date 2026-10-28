-- +goose Up
-- nexway_operators: eKYC確認を行ったオペレーター（確認者）情報
CREATE TABLE nexway_operators (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    ekyc_request_log_id BIGINT NOT NULL COMMENT 'eKYC申請ログID',
    company_id BIGINT NOT NULL COMMENT '確認会社ID',
    company_name VARCHAR(255) NOT NULL COMMENT '確認会社名',
    user_id BIGINT NOT NULL COMMENT '確認者ID',
    user_name VARCHAR(255) NOT NULL COMMENT '確認者名',
    login_id VARCHAR(255) NOT NULL COMMENT '確認者ログインID',
    operator_type VARCHAR(50) NOT NULL COMMENT '確認者権限 (first, second, sv)',
    operation_end_at DATETIME NOT NULL COMMENT 'BPO終了日時',
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日時',
    INDEX idx_ekyc_request_log_id (ekyc_request_log_id),
    INDEX idx_operator_type (operator_type),
    CONSTRAINT fk_nexway_operators_ekyc_request_log
        FOREIGN KEY (ekyc_request_log_id) REFERENCES ekyc_request_log(id) ON DELETE RESTRICT
) COMMENT='Nexway eKYC確認者（オペレーター）情報';

-- +goose Down
DROP TABLE IF EXISTS nexway_operators;
