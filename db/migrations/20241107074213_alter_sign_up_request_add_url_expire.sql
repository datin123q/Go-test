-- +goose Up
alter table sign_up_request
    add url_expire datetime null after email_expire;

UPDATE sign_up_request
SET url_expire = DATE_ADD(email_expire, INTERVAL 20 MINUTE);


-- +goose Down
alter table sign_up_request
    drop column url_expire;
