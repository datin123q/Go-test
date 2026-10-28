-- +goose Up
alter table user
drop
column is_address_change;

-- +goose Down
alter table user
    add is_address_change tinyint(1) null;

