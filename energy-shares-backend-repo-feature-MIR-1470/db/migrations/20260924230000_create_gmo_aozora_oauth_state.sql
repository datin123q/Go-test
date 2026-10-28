-- +goose Up
-- MIR-1298: GMOあおぞら銀行のOAuth認可コールバック(/gmo-aozora/auth/*)が
-- stateを照合していなかったため、外から呼んで本番のトークンを差し替えられた。
-- admin-apiが発行したワンタイムのstateをここに保存し、jobs-schedulerのコールバックで
-- 照合する。lookupはstateの生値のハッシュ(admin-apiが検索に使う)、state_hashは
-- ライブラリが実際にURLへ載せるsha256(state+salt)のハッシュ(jobs-schedulerが
-- コールバックの照合に使う)で、値そのものが違う別カラム。
CREATE TABLE IF NOT EXISTS gmo_aozora_oauth_state (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    lookup CHAR(64) NOT NULL,
    state_hash CHAR(64) NULL,
    nonce VARCHAR(64) NULL,
    admin_user_id BIGINT NOT NULL,
    expires_at DATETIME NOT NULL,
    used_at DATETIME NULL,
    result VARCHAR(32) NULL,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_gmo_aozora_oauth_state_lookup (lookup),
    UNIQUE KEY uk_gmo_aozora_oauth_state_state_hash (state_hash)
);

-- +goose Down
DROP TABLE IF EXISTS gmo_aozora_oauth_state;
