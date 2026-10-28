-- +goose Up
alter table user
drop column satori_hashcode;

-- +goose Down
alter table user
    add satori_hashcode varchar(50) NULL;
