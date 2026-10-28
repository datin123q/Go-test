-- +goose Up
DROP TABLE IF EXISTS casbin_rule;

-- +goose Down
CREATE TABLE IF NOT EXISTS casbin_rule
(
    p_type VARCHAR(100) NULL,
    v0     VARCHAR(100) NULL,
    v1     VARCHAR(100) NULL,
    v2     VARCHAR(100) NULL,
    v3     VARCHAR(100) NULL,
    v4     VARCHAR(100) NULL,
    v5     VARCHAR(100) NULL
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4;

CREATE INDEX IF NOT EXISTS IDX_casbin_rule_p_type ON casbin_rule (p_type);
CREATE INDEX IF NOT EXISTS IDX_casbin_rule_v0 ON casbin_rule (v0);
CREATE INDEX IF NOT EXISTS IDX_casbin_rule_v1 ON casbin_rule (v1);
CREATE INDEX IF NOT EXISTS IDX_casbin_rule_v2 ON casbin_rule (v2);
CREATE INDEX IF NOT EXISTS IDX_casbin_rule_v3 ON casbin_rule (v3);
CREATE INDEX IF NOT EXISTS IDX_casbin_rule_v4 ON casbin_rule (v4);
CREATE INDEX IF NOT EXISTS IDX_casbin_rule_v5 ON casbin_rule (v5);