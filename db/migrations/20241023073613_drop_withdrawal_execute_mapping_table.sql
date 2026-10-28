-- +goose Up
-- SQL in section 'Up' is executed when migrating up.
DROP TABLE IF EXISTS withdrawal_execute_mapping;
-- +goose Down
-- SQL in section 'Down' is executed when migrating down.
CREATE TABLE withdrawal_execute_mapping (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    history_id BIGINT NULL,
    transfer_id BIGINT NULL
) ENGINE = InnoDB COLLATE = utf8mb4_general_ci;