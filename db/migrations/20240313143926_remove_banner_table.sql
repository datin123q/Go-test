-- +goose Up
-- +goose StatementBegin
drop table project_banner;
-- +goose StatementEnd
-- +goose StatementBegin
drop table popular_info;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
create table project_banner
(
    id               bigint auto_increment
        primary key,
    project_id       bigint       null,
    posted           tinyint(1)   null,
    `index`          int          null,
    project_name     varchar(255) null,
    description      varchar(255) null,
    banner_image_url varchar(255) null,
    post_date        datetime     null,
    created          datetime     null
)
    charset = utf8mb4 collate = utf8mb4_general_ci;
-- +goose StatementEnd
-- +goose StatementBegin
create table popular_info
(
    id           bigint auto_increment
        primary key,
    `index`      int          null,
    project_id   bigint       null,
    posted       tinyint(1)   null,
    project_name varchar(255) null,
    updated      datetime     null,
    created      datetime     null,
    constraint UQE_popular_info_index
        unique (`index`)
)
    charset = utf8mb4 collate = utf8mb4_general_ci;
-- +goose StatementEnd
