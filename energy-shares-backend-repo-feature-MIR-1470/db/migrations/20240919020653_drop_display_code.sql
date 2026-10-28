-- +goose Up
drop table if exists display_code;
-- +goose Down
create table display_code (
    id bigint auto_increment primary key,
    code varchar(255) null,
    active tinyint(1) null,
    created datetime null,
    updated datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;