-- +goose Up
ALTER TABLE lottery_history DROP COLUMN history_type;

-- +goose Down
ALTER TABLE lottery_history
    ADD COLUMN history_type VARCHAR(100) NOT NULL DEFAULT '' AFTER application_id;
