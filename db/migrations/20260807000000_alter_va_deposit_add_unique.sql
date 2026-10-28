-- +goose Up
-- 既存データに同じ(va_account_number, item_key)がある場合は、このDDLが失敗して適用を中止する。
ALTER TABLE va_deposit
    ADD UNIQUE KEY uq_va_deposit_account_item_key (va_account_number, item_key),
    ALGORITHM=INPLACE,
    LOCK=NONE;

-- +goose Down
ALTER TABLE va_deposit
    DROP KEY uq_va_deposit_account_item_key,
    ALGORITHM=INPLACE,
    LOCK=NONE;
