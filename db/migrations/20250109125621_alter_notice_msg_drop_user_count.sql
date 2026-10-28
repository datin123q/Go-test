-- +goose Up
ALTER TABLE notice_msg
    DROP COLUMN notice_user_count;

-- +goose Down
ALTER TABLE notice_msg
    ADD COLUMN notice_user_count BIGINT NULL;
