-- +goose Up
ALTER TABLE code_ownership ADD COLUMN notes TEXT NULL COMMENT 'notes' AFTER code;

-- +goose Down
ALTER TABLE code_ownership DROP COLUMN notes;
