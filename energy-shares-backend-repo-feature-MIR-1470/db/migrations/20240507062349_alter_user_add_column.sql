-- +goose Up
alter table user
    add has_experience tinyint(1) null;

update user
set has_experience = true
where has_experience is null;

-- +goose Down
alter table user
drop column has_experience;