-- +goose Up
CREATE TABLE IF NOT EXISTS capital_raising_entity
(
    id                              bigint       auto_increment primary key,
    corporate_name                  varchar(255) not null,
    zip_code                        varchar(255) not null,
    province                        varchar(255) not null,
    city                            varchar(255) not null,
    address                         varchar(255) not null,
    address2                        varchar(255) not null,
    representative_family_name      varchar(255) not null,
    representative_first_name       varchar(255) not null,
    representative_family_name_kana varchar(255) not null,
    representative_first_name_kana  varchar(255) not null,
    image_url                       varchar(255) default null,
    created                         datetime     null,
    updated                         datetime     null,
    deleted                         datetime     null
);

-- +goose Down
DROP TABLE IF EXISTS capital_raising_entity;
