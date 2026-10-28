-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.
DROP TABLE IF EXISTS mail;
DROP TABLE IF EXISTS mail_request;
DROP TABLE IF EXISTS map_mail_receiver;
-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back.
CREATE TABLE IF NOT EXISTS mail (
    id bigint auto_increment primary key,
    status varchar(255) null,
    template_type varchar(255) null,
    subject varchar(255) null,
    name varchar(255) null,
    sender varchar(255) null,
    content longtext null,
    mail_type varchar(255) null,
    count bigint null,
    send_date datetime null,
    updated datetime null,
    created datetime null
) engine = InnoDB charset = utf8mb4 collate = utf8mb4_general_ci;
CREATE INDEX IF NOT EXISTS mail_status_index ON mail (status desc);
CREATE TABLE IF NOT EXISTS mail_request (
    id bigint auto_increment primary key,
    mail_id bigint null,
    status varchar(255) null,
    receiver varchar(255) null,
    template_data longtext null,
    message_id varchar(255) null,
    message_status varchar(255) null,
    message_error varchar(255) null,
    send_date datetime null,
    updated datetime null,
    created datetime null
) engine = InnoDB charset = utf8mb4 collate = utf8mb4_general_ci;
CREATE INDEX IF NOT EXISTS IDX_mail_request_mail_id ON mail_request (mail_id);
CREATE TABLE IF NOT EXISTS map_mail_receiver (
    id bigint auto_increment primary key,
    mail_id int not null,
    receiver varchar(50) null,
    created datetime not null
) engine = InnoDB charset = utf8mb4 collate = utf8mb4_general_ci;