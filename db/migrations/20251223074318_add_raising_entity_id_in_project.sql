-- +goose Up
ALTER TABLE project ADD COLUMN capital_raising_entity_id bigint DEFAULT NULL;

-- +goose Down
ALTER TABLE project DROP COLUMN capital_raising_entity_id;
