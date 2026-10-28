-- +goose Up
alter table policy
    change update_id modifier_id bigint null;

alter table policy
    change updator_email modifier_email varchar(255) null;

-- +goose Down
alter table policy
    change modifier_id update_id bigint null;

alter table policy
    change modifier_email updator_email varchar(255) null;
