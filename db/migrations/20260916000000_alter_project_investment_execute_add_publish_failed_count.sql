-- +goose Up
-- 出資実行(project_investment_execute)は、法定帳簿へのイベント送信を
-- コミット後(AfterCommit)へ移した(MIR-1202)。これにより、コミットが
-- 完了していてもKafka送信自体が失敗するケースが起き得るが、admin-apiの
-- ミドルウェアは既にレスポンスを返した後のため、失敗をHTTPで伝えられない。
-- 送信失敗があったことを画面から見えるようにするため、このプロジェクトの
-- 出資実行を通じた累積の送信失敗件数を持たせる。
ALTER TABLE project_investment_execute
    ADD COLUMN publish_failed_count BIGINT NOT NULL DEFAULT 0 COMMENT '法定帳簿へのイベント送信に失敗した累積件数(MIR-1202)';

-- +goose Down
ALTER TABLE project_investment_execute
    DROP COLUMN publish_failed_count;
