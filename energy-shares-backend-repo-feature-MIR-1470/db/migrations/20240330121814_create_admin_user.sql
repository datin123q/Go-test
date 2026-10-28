-- +goose Up
DROP TABLE IF EXISTS admin_user;

create table admin_user
(
    id                  bigint auto_increment,
    common_id           varchar(255) null,
    email               varchar(255) not null,
    name                varchar(255) null,
    password            varchar(255) null,
    salt                varchar(255) null,
    mobile_phone        varchar(255) null,
    last_login_date     datetime null,
    login_fail_cnt      int null,
    last_login_ip       varchar(50) null,
    last_login_agent    varchar(255) null,
    role_updated        datetime null,
    validity_start_date datetime null,
    validity_end_date   datetime null,
    created             datetime null,
    updated             datetime null,
    deleted             datetime null,
    constraint admin_user_pk
        primary key (id)
) comment 'admin_user'
    charset = utf8mb4 collate = utf8mb4_general_ci;

INSERT INTO admin_user (id, common_id, email,name, password, salt, mobile_phone, role_updated, validity_start_date,
                        validity_end_date, created, updated,
                        deleted)
SELECT u.id,
       u.common_id,
       u.email,
       u.name,
       u.password,
       u.salt,
       u.mobile_phone,
       ap.role_updated,
       ap.validity_start_date,
       ap.validity_end_date,
       ap.created,
       ap.updated,
       ap.deleted
FROM admin_profile ap
         JOIN user u ON ap.user_id = u.id
WHERE u.user_type = 'ADMIN';

-- drop table admin_profile;
-- DELETE FROM user WHERE user_type = 'ADMIN';

-- +goose Down
drop table admin_user;
