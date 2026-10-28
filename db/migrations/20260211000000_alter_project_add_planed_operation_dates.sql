-- +goose Up
ALTER TABLE project ADD COLUMN planed_operation_start_date datetime DEFAULT NULL;
ALTER TABLE project ADD COLUMN planed_operation_end_date datetime DEFAULT NULL;

-- +goose Down
ALTER TABLE project DROP COLUMN planed_operation_start_date;
ALTER TABLE project DROP COLUMN planed_operation_end_date;
