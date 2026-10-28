-- +goose Up
ALTER TABLE qna ADD COLUMN actor JSON NULL;
ALTER TABLE answer ADD COLUMN actor JSON NULL;
ALTER TABLE attachment ADD COLUMN actor JSON NULL;

-- +goose Down
ALTER TABLE qna DROP COLUMN actor;
ALTER TABLE answer DROP COLUMN actor;
ALTER TABLE attachment DROP COLUMN actor;
