-- +goose Up
alter table private_profile
    modify user_id varchar(255) not null;

alter table private_profile
    modify request_id varchar(255) not null;

-- +goose Down
alter table private_profile
    modify user_id varchar(255) null;

alter table private_profile
    modify request_id varchar(255) null;