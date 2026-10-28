-- +goose Up
ALTER TABLE project ADD COLUMN is_closed_fund boolean NOT NULL DEFAULT false;
ALTER TABLE project ADD COLUMN closed_fund_slug varchar(255) NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE project DROP COLUMN is_closed_fund;
ALTER TABLE project DROP COLUMN closed_fund_slug;
