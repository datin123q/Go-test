-- +goose Up
alter table sign_up_request
    add user_type varchar(50) null;

-- +goose Down
alter table sign_up_request
    drop column user_type;