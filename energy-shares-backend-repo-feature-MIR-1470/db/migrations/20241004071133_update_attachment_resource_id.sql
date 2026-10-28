-- +goose Up
UPDATE attachment a
    JOIN policy p ON a.id = p.attachment_id
    SET a.resource_id = p.common_id;

-- +goose Down
UPDATE attachment a
    JOIN policy p ON a.id = p.attachment_id
    SET a.resource_id = NULL;
