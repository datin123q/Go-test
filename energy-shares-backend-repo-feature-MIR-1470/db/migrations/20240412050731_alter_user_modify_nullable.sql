-- +goose Up
alter table user
    modify city varchar(255) null;

alter table user
    modify investment_purpose varchar(255) null;

alter table user
    modify is_surplus_fund tinyint(1) null;

alter table user
    modify is_newsletter tinyint(1) null;

alter table user
    modify bank_name varchar(255) null;

alter table user
    modify deposit_type varchar(255) null;

alter table user
    modify postal_code varchar(255) null;

alter table user
    modify postal_code_status tinyint(1) null;

alter table user
    modify agree_pro_investor tinyint(1) null;



-- +goose Down
alter table user
    modify city varchar(255) not null;

alter table user
    modify investment_purpose varchar(255) not null;

alter table user
    modify is_surplus_fund tinyint(1) not null;

alter table user
    modify is_newsletter tinyint(1) not null;

alter table user
    modify bank_name varchar(255) not null;

alter table user
    modify deposit_type varchar(255) not null;

alter table user
    modify postal_code varchar(255) not null;

alter table user
    modify postal_code_status tinyint(1) not null;

alter table user
    modify agree_pro_investor tinyint(1) not null;

