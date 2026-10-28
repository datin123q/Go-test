-- +goose Up
CREATE TABLE IF NOT EXISTS lottery_application (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    project_id bigint not null,
    user_id bigint not null,
    qty bigint not null,
    unit_price bigint not null,
    amount bigint not null,

    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    lost_at DATETIME,
    winning_at DATETIME,
    canceled_at DATETIME,

    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted DATETIME NULL,
    INDEX idx_lottery_application_project_id (project_id),
    INDEX idx_lottery_application_user_id (user_id, project_id)
);

CREATE TABLE IF NOT EXISTS lottery_winning (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    lottery_application_id VARCHAR(36) NOT NULL,
    share_id varchar(255) NOT NULL,

    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted DATETIME NULL,

    INDEX idx_lottery_winning_lottery_application_id_share_id (lottery_application_id, share_id)
);

CREATE TABLE IF NOT EXISTS lottery_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    application_id VARCHAR(36) NOT NULL,
    history_type VARCHAR(100) NOT NULL,
    prev_status VARCHAR(100) NOT NULL,
    new_status VARCHAR(100) NOT NULL,

    actor_id bigint not null,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE project ADD COLUMN selection_type VARCHAR(100) NOT NULL DEFAULT 'FIRST_COME_FIRST_SERVED' AFTER category_type;

-- +goose Down
ALTER TABLE project DROP COLUMN selection_type;
drop table if exists lottery_application;
drop table if exists lottery_winning;
drop table if exists lottery_history;
