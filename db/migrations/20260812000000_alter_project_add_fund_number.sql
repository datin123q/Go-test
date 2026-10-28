-- +goose Up
ALTER TABLE project
    ADD COLUMN fund_number BIGINT NULL COMMENT '投資家へ提示するファンド番号';

-- 未入力は NULL とし、MySQL の仕様で NULL 同士は重複とみなされない。
CREATE UNIQUE INDEX ux_project_fund_number ON project (fund_number);

-- +goose Down
DROP INDEX ux_project_fund_number ON project;

ALTER TABLE project
    DROP COLUMN fund_number;
