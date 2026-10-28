-- +goose Up
drop table if exists project_catalog;
-- +goose Down
create table if not exists project_catalog (
    id bigint auto_increment primary key,
    project_id bigint not null,
    file_name varchar(255) null,
    original_name varchar(255) null,
    file_type varchar(255) null,
    file_size bigint not null,
    file_extension varchar(255) null,
    file_path varchar(255) null,
    file_url varchar(255) null,
    owner_id bigint not null,
    update_id bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_project_catalog_file_name unique (file_name)
) charset = utf8mb4 collate = utf8mb4_general_ci;