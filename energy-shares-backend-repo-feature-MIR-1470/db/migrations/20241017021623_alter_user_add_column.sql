-- +goose Up
ALTER TABLE user
ADD satori_hashcode VARCHAR(50) NULL;

-- +goose Down
ALTER TABLE user
DROP COLUMN satori_hashcode;
