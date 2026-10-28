-- +goose Up
UPDATE attachment a
    JOIN policy p
ON a.id = p.attachment_id
    SET a.resource_id = p.common_id
WHERE p.attachment_id IS NOT NULL;

-- +goose Down
UPDATE attachment a
    JOIN policy p
ON a.id = p.attachment_id
    SET a.resource_id = NULL
WHERE p.attachment_id IS NOT NULL;
