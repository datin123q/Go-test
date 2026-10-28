-- +goose Up
-- MIR-1492: 管理画面のファンド詳細を複数の管理者が同時に編集すると、後から保存した側が
-- 画面を開いた時点の値で全項目を書き戻すため、先に保存した側の変更が巻き戻る。
-- ファンド詳細の保存(UpdateForAdminDetail)ごとに1ずつ増やし、画面を開いた時点の
-- 値と比べて、その間に別の保存があったことを検知する。updated はジョブの状態変更でも
-- 更新され、上書きの起きない変更でも警告が出てしまうため使わない。
ALTER TABLE project
    ADD COLUMN detail_version INT NOT NULL DEFAULT 0
    COMMENT 'ファンド詳細の保存回数。管理画面の同時編集の検知に使う';

-- +goose Down
ALTER TABLE project
    DROP COLUMN detail_version;
