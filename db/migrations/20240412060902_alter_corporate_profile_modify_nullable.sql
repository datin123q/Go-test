-- +goose Up
alter table corporate_profile
    modify corp_type varchar(255) null;

alter table corporate_profile
    modify employee_count varchar(255) null;

alter table corporate_profile
    modify capital_amount varchar(255) null;

alter table corporate_profile
    modify sales varchar(255) null;

alter table corporate_profile
    modify representative_family_name varchar(255) null;

alter table corporate_profile
    modify representative_first_name varchar(255) null;

alter table corporate_profile
    modify representative_family_name_kana varchar(255) null;

alter table corporate_profile
    modify representative_first_name_kana varchar(255) null;

alter table corporate_profile
    modify representative_mobile_phone varchar(255) null;

alter table corporate_profile
    modify representative_email varchar(255) null;

alter table corporate_profile
    modify manager_family_name varchar(255) null;

alter table corporate_profile
    modify manager_first_name varchar(255) null;

alter table corporate_profile
    modify manager_family_name_kana varchar(255) null;

alter table corporate_profile
    modify manager_first_name_kana varchar(255) null;

alter table corporate_profile
    modify manager_email varchar(255) null;

alter table corporate_profile
    modify manager_zip_code varchar(255) null;

alter table corporate_profile
    modify manager_province varchar(255) null;

alter table corporate_profile
    modify manager_city varchar(255) null;

alter table corporate_profile
    modify manager_address varchar(255) null;

alter table corporate_profile
    modify manager_mobile_phone varchar(255) null;

alter table corporate_profile
    modify manager_birthday datetime null;

alter table corporate_profile
    modify department varchar(255) null;

alter table corporate_profile
    modify position varchar(255) null;



-- +goose Down
alter table corporate_profile
    modify corp_type varchar(255) not null;

alter table corporate_profile
    modify employee_count varchar(255) not null;

alter table corporate_profile
    modify capital_amount varchar(255) not null;

alter table corporate_profile
    modify sales varchar(255) not null;

alter table corporate_profile
    modify representative_family_name varchar(255) not null;

alter table corporate_profile
    modify representative_first_name varchar(255) not null;

alter table corporate_profile
    modify representative_family_name_kana varchar(255) not null;

alter table corporate_profile
    modify representative_first_name_kana varchar(255) not null;

alter table corporate_profile
    modify representative_mobile_phone varchar(255) not null;

alter table corporate_profile
    modify representative_email varchar(255) not null;

alter table corporate_profile
    modify manager_family_name varchar(255) not null;

alter table corporate_profile
    modify manager_first_name varchar(255) not null;

alter table corporate_profile
    modify manager_family_name_kana varchar(255) not null;

alter table corporate_profile
    modify manager_first_name_kana varchar(255) not null;

alter table corporate_profile
    modify manager_email varchar(255) not null;

alter table corporate_profile
    modify manager_zip_code varchar(255) not null;

alter table corporate_profile
    modify manager_province varchar(255) not null;

alter table corporate_profile
    modify manager_city varchar(255) not null;

alter table corporate_profile
    modify manager_address varchar(255) not null;

alter table corporate_profile
    modify manager_mobile_phone varchar(255) not null;

alter table corporate_profile
    modify manager_birthday datetime not null;

alter table corporate_profile
    modify department varchar(255) not null;

alter table corporate_profile
    modify position varchar(255) not null;


