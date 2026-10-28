-- +goose Up
-- MIR-1249: RepaymentDividendJob(配当・償還処理)にDBレベルの排他ロックを追加する。
-- ロック対象のjob_typeは実行時にINSERT IGNOREで用意せず、ここで前もって1行だけ
-- 用意する(理由は20260914000000_seed_job_execution_lock.sqlと同じ)。
INSERT IGNORE INTO job_execution_lock (job_type) VALUES ('DIVIDEND_PAYMENT');

-- +goose Down
DELETE FROM job_execution_lock WHERE job_type = 'DIVIDEND_PAYMENT';
