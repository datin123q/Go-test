-- +goose Up
ALTER TABLE lottery_csv_upload_job_detail
    ADD COLUMN expected_applied_at VARCHAR(32) NULL
    AFTER expected_lottery_status;

-- +goose Down
ALTER TABLE lottery_csv_upload_job_detail
    DROP COLUMN expected_applied_at;
