-- +goose Up
ALTER TABLE withdrawal
    ADD COLUMN is_manual tinyint(1) NOT NULL DEFAULT 0 AFTER status;

-- +goose Down
ALTER TABLE withdrawal
    DROP COLUMN is_manual;
