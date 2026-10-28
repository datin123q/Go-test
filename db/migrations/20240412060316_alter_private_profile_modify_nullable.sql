-- +goose Up
alter table private_profile
    modify birthday datetime null;

alter table private_profile
    modify first_name varchar(255) null;

alter table private_profile
    modify family_name varchar(255) null;

alter table private_profile
    modify first_name_kana varchar(255) null;

alter table private_profile
    modify family_name_kana varchar(255) null;

alter table private_profile
    modify asset varchar(255) null;

alter table private_profile
    modify annual_income varchar(255) null;

alter table private_profile
    modify is_pro_invest_asset tinyint(1) null;

alter table private_profile
    modify is_pro_invest_finance_asset tinyint(1) null;

alter table private_profile
    modify nationality varchar(255) null;

alter table private_profile
    modify job varchar(255) null;

alter table private_profile
    modify gender varchar(255) null;

alter table private_profile
    modify request_id varchar(255) null;



-- +goose Down
alter table private_profile
    modify birthday datetime not null;

alter table private_profile
    modify first_name varchar(255) not null;

alter table private_profile
    modify family_name varchar(255) not null;

alter table private_profile
    modify first_name_kana varchar(255) not null;

alter table private_profile
    modify family_name_kana varchar(255) not null;

alter table private_profile
    modify asset varchar(255) not null;

alter table private_profile
    modify annual_income varchar(255) not null;

alter table private_profile
    modify is_pro_invest_asset tinyint(1) not null;

alter table private_profile
    modify is_pro_invest_finance_asset tinyint(1) not null;

alter table private_profile
    modify nationality varchar(255) not null;

alter table private_profile
    modify job varchar(255) not null;

alter table private_profile
    modify gender varchar(255) not null;

alter table private_profile
    modify request_id varchar(255) not null;

