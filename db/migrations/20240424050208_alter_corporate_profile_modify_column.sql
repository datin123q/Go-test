-- +goose Up
alter table corporate_profile
    modify corp_no varchar(255) null;

alter table corporate_profile
    modify user_id varchar(255) not null;

-- +goose Down
alter table corporate_profile
    modify corp_no varchar(255) not null;

alter table corporate_profile
    modify user_id varchar(255) null;