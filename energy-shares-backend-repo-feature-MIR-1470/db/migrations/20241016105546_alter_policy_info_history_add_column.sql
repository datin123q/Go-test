-- +goose Up
alter table policy_info_history
    add is_posted tinyint default 0 not null after email;

alter table policy_info_history
    add label date null after is_posted;


-- +goose Down
alter table policy_info_history
drop column is_posted;

alter table policy_info_history
drop column label;
