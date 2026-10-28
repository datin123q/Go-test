-- +goose Up
alter table refresh_token
    add user_type varchar(10) after user_id;


-- +goose Down
alter table refresh_token
drop column user_type;


