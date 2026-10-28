-- +goose Up
CREATE TABLE IF NOT EXISTS lottery_csv_upload_job (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    job_id VARCHAR(64) NOT NULL,
    project_id BIGINT NOT NULL,
    requested_by VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL,
    total_row_count INT NOT NULL DEFAULT 0,
    success_row_count INT NOT NULL DEFAULT 0,
    failed_row_count INT NOT NULL DEFAULT 0,
    started_at DATETIME NULL,
    finished_at DATETIME NULL,
    error_message TEXT NULL,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted DATETIME NULL,
    UNIQUE KEY uq_lottery_csv_upload_job_job_id (job_id),
    KEY idx_lottery_csv_upload_job_project_id_created (project_id, created),
    KEY idx_lottery_csv_upload_job_status_created (status, created)
);

CREATE TABLE IF NOT EXISTS lottery_csv_upload_job_detail (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    job_id BIGINT NOT NULL,
    `row_number` INT NOT NULL,
    lottery_application_id VARCHAR(255) NOT NULL,
    input_applied_qty BIGINT NULL,
    input_winning_qty BIGINT NOT NULL,
    process_type VARCHAR(16) NOT NULL,
    status VARCHAR(32) NOT NULL,
    error_code VARCHAR(64) NULL,
    error_message TEXT NULL,
    processed_at DATETIME NULL,
    created_share_id VARCHAR(255) NULL,
    created_deposit_request_id VARCHAR(255) NULL,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted DATETIME NULL,
    UNIQUE KEY uq_lottery_csv_upload_job_detail_job_id_row_number (job_id, `row_number`),
    KEY idx_lottery_csv_upload_job_detail_job_id_status (job_id, status),
    KEY idx_lottery_csv_upload_job_detail_application_id (lottery_application_id),
    CONSTRAINT fk_lottery_csv_upload_job_detail_job_id
        FOREIGN KEY (job_id) REFERENCES lottery_csv_upload_job (id)
);

-- +goose Down
DROP TABLE IF EXISTS lottery_csv_upload_job_detail;
DROP TABLE IF EXISTS lottery_csv_upload_job;
