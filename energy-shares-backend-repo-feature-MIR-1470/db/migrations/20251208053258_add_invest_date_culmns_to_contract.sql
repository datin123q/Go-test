-- +goose Up
ALTER TABLE contract ADD COLUMN invest_date datetime NULL COMMENT '投資日' AFTER user_name;



-- +goose Down
ALTER TABLE contract DROP COLUMN invest_date;

