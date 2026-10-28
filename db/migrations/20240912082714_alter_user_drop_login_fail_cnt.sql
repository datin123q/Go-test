-- +goose Up
alter table user
drop column login_fail_cnt;

-- +goose Down
alter table user
add column login_fail_cnt int not null default 0;
