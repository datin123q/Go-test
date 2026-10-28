-- +goose Up
ALTER TABLE project ADD COLUMN final_funding_amount BIGINT NOT NULL DEFAULT 0 COMMENT '募集終了時の申込金額' AFTER country;

-- +goose Down
ALTER TABLE project DROP COLUMN final_funding_amount;
