-- +goose Up
-- 償還完了の前提チェック(MIR-272)は、ファンドの予定を FOR UPDATE で読む。この表には project_id の
-- 索引が無く、MySQL の REPEATABLE READ では走査した全行と隙間がロックされるため、別のファンドの
-- 予定の更新(分配償還ジョブの完了処理)・作成まで、償還完了のトランザクションが終わるまで待たされる。
-- 索引で範囲をそのファンドの行に絞る。すでに同名の索引がある環境では何もしない。
SET @idx_exists = (
    SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'repayment_dividend_schedule'
      AND index_name = 'IDX_repayment_dividend_schedule_project_id_status'
);
SET @ddl = IF(@idx_exists = 0,
    'CREATE INDEX IDX_repayment_dividend_schedule_project_id_status ON repayment_dividend_schedule (project_id, status)',
    'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- +goose Down
-- 索引がある場合だけ消す(Up と対称)。索引が無い環境でも失敗しない。
SET @idx_exists = (
    SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'repayment_dividend_schedule'
      AND index_name = 'IDX_repayment_dividend_schedule_project_id_status'
);
SET @ddl = IF(@idx_exists > 0,
    'DROP INDEX IDX_repayment_dividend_schedule_project_id_status ON repayment_dividend_schedule',
    'SELECT 1');
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
