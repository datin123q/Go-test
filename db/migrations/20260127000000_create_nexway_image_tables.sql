-- +goose Up
-- nexway_applicant_image: eKYC申請時に提出された本人確認書類の画像
CREATE TABLE nexway_applicant_image (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    ekyc_request_log_id BIGINT NOT NULL COMMENT 'eKYC申請ログID',
    image_type VARCHAR(20) NOT NULL COMMENT '画像種別',
    s3_key VARCHAR(1000) NOT NULL COMMENT 'S3キー',
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日時',
    INDEX idx_ekyc_request_log_id (ekyc_request_log_id),
    INDEX idx_image_type (image_type),
    CONSTRAINT fk_nexway_applicant_image_ekyc_request_log
        FOREIGN KEY (ekyc_request_log_id) REFERENCES ekyc_request_log(id) ON DELETE RESTRICT
) COMMENT='Nexway eKYC申請画像（元画像）';

-- nexway_masked_image: 個人情報をマスキング処理した画像
CREATE TABLE nexway_masked_image (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    ekyc_request_log_id BIGINT NOT NULL COMMENT 'eKYC申請ログID',
    image_type VARCHAR(20) NOT NULL COMMENT '画像種別',
    s3_key VARCHAR(1000) NOT NULL COMMENT 'S3キー',
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日時',
    INDEX idx_ekyc_request_log_id (ekyc_request_log_id),
    INDEX idx_image_type (image_type),
    CONSTRAINT fk_nexway_masked_image_ekyc_request_log
        FOREIGN KEY (ekyc_request_log_id) REFERENCES ekyc_request_log(id) ON DELETE RESTRICT
) COMMENT='Nexway eKYC申請画像（マスキング済み）';

-- +goose Down
DROP TABLE IF EXISTS nexway_masked_image;
DROP TABLE IF EXISTS nexway_applicant_image;
