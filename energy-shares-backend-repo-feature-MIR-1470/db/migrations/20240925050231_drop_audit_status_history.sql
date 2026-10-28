-- +goose Up
drop table if exists audit_status_history;
-- +goose Down
create table if not exists audit_status_history (
    id bigint auto_increment primary key,
    user_id bigint null,
    owner_id bigint null,
    update_id bigint null,
    processor varchar(255) null,
    state varchar(255) null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_audit_status_history_owner_id on audit_status_history (owner_id);
create index IDX_audit_status_history_update_id on audit_status_history (update_id);