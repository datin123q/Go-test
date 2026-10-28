-- +goose Up
CREATE TABLE IF NOT EXISTS code_ownership
(
    id          bigint         auto_increment primary key,
    owner_type  varchar(255)   not null,
    owner_id    bigint         not null,
    code_type   varchar(255)   not null,
    code        varchar(255)   not null,
    created     datetime     null,
    updated     datetime     null,
    deleted     datetime     null,
    unique key UQE_code_ownership_code_type_code (code_type, code)
);

CREATE TABLE IF NOT EXISTS code_usage
(
    id                  bigint          auto_increment primary key,
    used_owner_type     varchar(255)    not null,
    used_owner_id       bigint          not null,
    code_ownership_id   bigint          not null,
    code_input_at       datetime        null,
    created             datetime        null,
    updated             datetime        null,
    deleted             datetime        null
);

-- +goose Down
DROP TABLE IF EXISTS code_usage;
DROP TABLE IF EXISTS code_ownership;
