-- +goose Up
alter table sign_up_request
drop column url_expire;

-- +goose Down
alter table sign_up_request
    add url_expire datetime null;
