-- +goose Up
CREATE TABLE IF NOT EXISTS mail_retry
(
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    email_request JSON         NOT NULL,
    receivers     JSON         NOT NULL,
    template_name VARCHAR(255) NOT NULL,
    retry_status  VARCHAR(50)  NOT NULL, -- PENDING, IN_PROGRESS, FAILED
    created       DATETIME     NOT NULL
);

CREATE TABLE IF NOT EXISTS mail_failed
(
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    email_request JSON         NOT NULL,
    receivers     JSON         NOT NULL,
    template_name VARCHAR(255) NOT NULL,
    retry_status  VARCHAR(50)  NOT NULL, -- PENDING, IN_PROGRESS, FAILED
    created       DATETIME     NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS mail_retry;
DROP TABLE IF EXISTS mail_failed;

