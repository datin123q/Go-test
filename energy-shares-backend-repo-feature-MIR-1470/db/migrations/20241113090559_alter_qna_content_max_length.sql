-- +goose Up
ALTER TABLE qna MODIFY content VARCHAR(1000) NULL;

-- +goose Down
ALTER TABLE qna MODIFY content VARCHAR(255) NULL;
