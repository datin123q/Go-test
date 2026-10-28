-- +goose Up
UPDATE user
SET investment_purpose = 'ETC'
WHERE investment_purpose = 'MANAGING_SURPLUS_FUNDS';

-- +goose Down
