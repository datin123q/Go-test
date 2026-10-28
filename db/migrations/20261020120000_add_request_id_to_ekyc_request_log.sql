-- +goose Up
-- MIR-1469: JPKI方式(本人確認デジタル認証サービス)の署名用URLを発行したときに、Nexwayが
-- 発行する検証依頼ID(request_id)を保存する。後続の署名結果Webhook・BPO申請で突き合わせに使う。
--
-- request_idの列は migration 20240508043536 で一度削除している(当時のホ方式では使わなかった)。
-- 今回は別の用途(JPKI方式のみ)で追加し直す。ホ方式の行はNULLのまま。
-- 既存の行を書き換えず、NULL許可で追加するため、前のバージョンのアプリが読んでも影響しない。
-- (逆に、この列を持たないDBに新しいアプリが書き込むと失敗する。マイグレーションを先に適用すること)
--
-- 検証依頼IDはNexwayが依頼ごとに一意に発行する(仕様書 p.15)ため、索引は一意にする。
-- MySQLの一意索引はNULLを複数持てるので、ホ方式の行(NULL)は影響しない。
ALTER TABLE ekyc_request_log
    ADD COLUMN request_id VARCHAR(255) NULL AFTER ekyc_id,
    ADD UNIQUE INDEX uk_ekyc_request_log_request_id (request_id);

-- +goose Down
ALTER TABLE ekyc_request_log
    DROP INDEX uk_ekyc_request_log_request_id,
    DROP COLUMN request_id;
