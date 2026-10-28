-- +goose Up
alter table user
drop column agree_pro_investor;

alter table user
drop column my_number;

-- +goose Down
alter table user
    add agree_pro_investor tinyint(1) null;

alter table user
    add my_number text null;