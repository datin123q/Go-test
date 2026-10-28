-- +goose Up
CREATE TABLE IF NOT EXISTS user_login_lock
(
    id             BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id        BIGINT   NOT NULL,
    login_fail_cnt INT      NOT NULL DEFAULT 0,
    lock_at        DATETIME NULL,
    created        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted        DATETIME NULL
);

-- +goose Down
DROP TABLE IF EXISTS user_login_lock;
