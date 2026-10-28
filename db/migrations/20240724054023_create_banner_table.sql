-- +goose Up
CREATE TABLE IF NOT EXISTS banner
(
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NULL,
    is_posted   TINYINT(1)   NOT NULL DEFAULT 1,
    sort_order  INT          NULL,
    url         VARCHAR(255) NOT NULL,
    start_date  DATETIME     NULL,
    end_date    DATETIME     NULL,
    new_tab     TINYINT(1)   NOT NULL DEFAULT 0,
    created     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated     DATETIME     NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted     DATETIME     NULL
);

CREATE TABLE IF NOT EXISTS banner_setting
(
    id            BIGINT AUTO_INCREMENT
        PRIMARY KEY,
    setting_key   VARCHAR(255) NOT NULL,
    setting_value VARCHAR(255) NOT NULL,
    created       DATETIME     NULL,
    updated       DATETIME     NULL,
    deleted       DATETIME     NULL
);

INSERT INTO banner_setting (setting_key, setting_value, created, updated)
VALUES ('ROLLING_SECOND', '5', NOW(), NOW());

-- +goose Down
DROP TABLE IF EXISTS banner;
DROP TABLE IF EXISTS banner_setting;
