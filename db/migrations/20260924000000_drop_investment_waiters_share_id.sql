-- +goose Up
-- investment_waiters.share_id の実カラムを撤去する(MIR-1118/MIR-1163 PR4)。
-- 当選紐づけは investment_waiters_winning への移行が完了しており(PR1-3、
-- MIR-1163)、DROP直前に本番DBで取りこぼし確認クエリ(投資待機者チケット参照)
-- を実行済み(ミラリタ・ZOLOMEとも0件)。domain.InvestmentWaiters.ShareId
-- フィールドは、DEFAULT_INVESTMENT_WAIT_LIST内のサブクエリ(investment_waiters_
-- winning経由、AS share_id)が返す値の受け皿として残す。このカラムを対象と
-- する全リポジトリのGet/Find呼び出しは.Select()で明示的にカラムを指定して
-- おり、暗黙の全フィールドSELECT(構造体からの自動列挙)は存在しないことを
-- 確認済み(Count/Sum等の集計関数は個別カラムに依存しないため対象外。確認
-- 結果はCLAUDE.mdのReview Checklist参照)。
ALTER TABLE investment_waiters DROP COLUMN share_id;

-- +goose Down
-- 撤去したデータ(旧繰上げ方式で書き込まれたshare_idの値)は戻せない。
-- カラムだけ元の定義(varchar(255) null)で復元する。
ALTER TABLE investment_waiters ADD COLUMN share_id VARCHAR(255) NULL;
