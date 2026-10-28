-- +goose Up
alter table corporate_profile
drop
column manager_is_address_change;

-- +goose Down
alter table corporate_profile
    add manager_is_address_change tinyint(1) null;

