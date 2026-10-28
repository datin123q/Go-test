-- +goose Up
-- MIR-1118のPR1(20260825000000_create_investment_waiters_winning.sql)と同じバックフィルを
-- 再実行する(MIR-1163、PR3)。PR1(テーブル作成、#534)とPR2(investment_waiters.share_idへの
-- 書き込みを止めてinvestment_waiters_winningへ切り替えるコード、#511)は別々のPRで、本番でも
-- 約1時間・developでは約8時間、テーブルはあるが旧コードがまだ旧列に直接書き込む窓が実在した。
-- ただし本番データを確認した結果、その窓による取りこぼしは無かった(旧列と中間テーブルの
-- 不一致0件、論理削除済みの当選だけを持つ待機者も0件)。PR3ではCOALESCEによる旧列
-- フォールバックを削除するため、万一見落としがあった場合の最終確認としてもう一度同じ
-- バックフィルを実行する。NOT EXISTSで既に取り込み済みの行は再挿入しない(PR1のバックフィルと
-- 同じ冪等性)。
--
-- バックフィル自体がここでの対策であり、これより後で追加の検証は行わない。INSERT ... SELECT
-- は部分成功しない(全行入るか、エラーで失敗してこのマイグレーション自体が失敗するかのどちらか)
-- ため、実行後に同じ条件で取りこぼしの有無を数えても、INSERTが成功していれば必ず0件になり
-- 検証にならない。取り返しのつかない操作(旧列のカラム自体を削除するPR4)の前に、取りこぼしが
-- 無いことを人が確認する手順を設けている(MIR-1163のPR4節を参照)。
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
-- バックフィルの取り消しは行わない(PR1のマイグレーションと同じ方針。取り込んだ行が
-- 元々旧コードによる正当な当選であり、他のバックフィル分と区別して安全に消せないため)。
