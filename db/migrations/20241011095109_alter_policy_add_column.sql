-- +goose Up
alter table policy
    add is_posted tinyint default 0 not null after attachment_id;

alter table policy
    add label date null after is_posted;


-- +goose Down
alter table policy
drop column is_posted;

alter table policy
drop column label;
