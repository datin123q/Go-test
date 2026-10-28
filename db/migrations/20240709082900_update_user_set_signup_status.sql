-- +goose Up
UPDATE user
SET signup_status = 'SIGNUP_COMPLETE'
WHERE signup_status IS NULL
   OR signup_status = '';

-- +goose Down
UPDATE user
SET signup_status = NULL
WHERE created < '2024-01-01 00:00:00';