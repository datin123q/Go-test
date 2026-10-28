-- +goose Up
ALTER TABLE contract ADD COLUMN address VARCHAR(255) NULL COMMENT '住所' AFTER user_name;
ALTER TABLE contract ADD COLUMN manager_name VARCHAR(255) NULL COMMENT '担当者名' AFTER address;

-- +goose Down
ALTER TABLE contract DROP COLUMN address;
ALTER TABLE contract DROP COLUMN manager_name;