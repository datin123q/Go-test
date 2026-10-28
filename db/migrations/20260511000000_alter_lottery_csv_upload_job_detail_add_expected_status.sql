-- +goose Up
ALTER TABLE lottery_csv_upload_job_detail
    ADD COLUMN expected_lottery_status VARCHAR(32) NULL
    AFTER lottery_application_id;

-- +goose Down
ALTER TABLE lottery_csv_upload_job_detail
    DROP COLUMN expected_lottery_status;
