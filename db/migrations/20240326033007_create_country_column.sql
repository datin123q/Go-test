-- +goose Up
alter table project
    add country varchar(50) not null default 'JP';

-- +goose Down
alter table project
drop column country;