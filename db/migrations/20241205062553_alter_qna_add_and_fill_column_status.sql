-- +goose Up
ALTER TABLE qna
    ADD COLUMN status VARCHAR(255) NULL;

UPDATE qna
SET status = CASE
                 WHEN `read` = 1 THEN 'RESPONSE_COMPLETED'
                 WHEN `read` = 0 THEN 'RECEIVED'
    END
WHERE status IS NULL;

ALTER TABLE qna
    MODIFY COLUMN status VARCHAR(255) NOT NULL;

-- +goose Down
ALTER TABLE qna
    DROP COLUMN status;
