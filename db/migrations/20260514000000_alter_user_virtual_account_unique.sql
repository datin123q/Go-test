-- +goose Up
ALTER TABLE user_virtual_account DROP KEY UQE_user_virtual_account_va_account_number;
ALTER TABLE user_virtual_account ADD UNIQUE KEY UQE_user_virtual_account_va_branch_account (va_branch_code, va_account_number);

-- +goose Down
ALTER TABLE user_virtual_account DROP KEY UQE_user_virtual_account_va_branch_account;
ALTER TABLE user_virtual_account ADD UNIQUE KEY UQE_user_virtual_account_va_account_number (va_account_number);
