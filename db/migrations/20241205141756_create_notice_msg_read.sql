-- +goose Up
CREATE TABLE IF NOT EXISTS notice_msg_read
(
    id     BIGINT PRIMARY KEY AUTO_INCREMENT,
    msg_id  BIGINT    NOT NULL,
    user_id BIGINT    NOT NULL,
    is_read BOOLEAN   NOT NULL DEFAULT FALSE,
    read_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS notice_msg_read;
