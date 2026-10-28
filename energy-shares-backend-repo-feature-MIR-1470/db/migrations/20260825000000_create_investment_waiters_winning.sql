-- +goose Up
-- investment_waiters.share_id(単一カラム)は、抽選の lottery_winning と同じ中間テーブルへ
-- 移行する。既存の繰上げ実績(share_id が入っている行)を先にバックフィルしてから、
-- 参照箇所を移し終えた後の別マイグレーションで share_id 自体を落とす(MIR-1118)。
-- ミラリタ本番調査時点(2026-08-25)で investment_waiters は1,643件、うち share_id が
-- 空文字(未紐づけ)1,142件・値あり501件・NULL 0件。バックフィル対象は501件
-- (このうち論理削除済みの行は0件で、下記の deleted 条件による除外対象はなし)。
CREATE TABLE IF NOT EXISTS investment_waiters_winning (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    investment_waiters_id BIGINT NOT NULL,
    share_id VARCHAR(255) NOT NULL,

    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted DATETIME NULL,

    INDEX idx_iww_investment_waiters_id_share_id (investment_waiters_id, share_id),
    INDEX idx_iww_share_id (share_id)
);

-- CREATE TABLE は MySQL の暗黙コミット対象で、この INSERT が失敗して同じマイグレーションを
-- 再実行した場合、テーブルは既に存在するため CREATE 側はスキップされ INSERT だけが再実行
-- されうる。NOT EXISTS で二重挿入を防いでいる。
-- investment_waiters は論理削除対象(deleted)。生SQLはxormのソフトデリートフィルタを
-- 経由しないため、ここで明示的に未削除行だけに絞る(xormの規約に合わせ、NULLと
-- ゼロ時刻の両方を未削除として扱う)。
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id)
SELECT iw.id, iw.share_id
FROM investment_waiters iw
WHERE iw.share_id IS NOT NULL AND iw.share_id != ''
  AND (iw.deleted IS NULL OR iw.deleted = '0001-01-01 00:00:00')
  AND NOT EXISTS (
      SELECT 1 FROM investment_waiters_winning iww
      WHERE iww.investment_waiters_id = iw.id
  );

-- +goose Down
DROP TABLE IF EXISTS investment_waiters_winning;
