-- +goose Up
ALTER TABLE user
    ADD COLUMN signup_status VARCHAR(50) NULL;

-- +goose Down
ALTER TABLE user
    DROP COLUMN signup_status;