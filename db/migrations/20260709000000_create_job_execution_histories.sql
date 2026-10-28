-- +goose Up
CREATE TABLE IF NOT EXISTS job_execution_histories (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    job_type VARCHAR(64) NOT NULL,
    job_name VARCHAR(128) NOT NULL,
    status VARCHAR(16) NOT NULL,
    started_at DATETIME NOT NULL,
    finished_at DATETIME NULL,
    error_type VARCHAR(64) NULL,
    error_log MEDIUMTEXT NULL,
    request_payload JSON NULL,
    result_payload JSON NULL,
    trigger_type VARCHAR(16) NOT NULL,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_job_execution_histories_status_started_at (status, started_at),
    KEY idx_job_execution_histories_job_type_started_at (job_type, started_at),
    KEY idx_job_execution_histories_started_at (started_at)
);

-- +goose Down
DROP TABLE IF EXISTS job_execution_histories;
