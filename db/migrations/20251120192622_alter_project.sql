-- +goose Up
ALTER TABLE project ADD COLUMN dividend_type ENUM('INCOME_AND_CAPITAL_GAIN', 'INCOME_GAIN', 'CAPITAL_GAIN') NULL COMMENT '配当方式' AFTER offering_type;

-- +goose Down
ALTER TABLE project DROP COLUMN dividend_type;
