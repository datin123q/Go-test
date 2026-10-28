-- +goose Up
-- 테이블이 존재하면 제거합니다.
DROP TABLE IF EXISTS deposit_withdrawal_execute_history;
-- +goose Down
-- 테이블이 존재하지 않으면 생성합니다.
CREATE TABLE IF NOT EXISTS deposit_withdrawal_execute_history (
    id bigint auto_increment primary key,
    investor_id bigint null,
    investor_name varchar(255) null,
    account_number varchar(255) null,
    amount bigint null,
    operator_name varchar(255) null,
    operator_email varchar(255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;