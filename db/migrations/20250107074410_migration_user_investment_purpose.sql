-- +goose Up

UPDATE user
SET investment_purpose = 'ETC'
WHERE investment_purpose = 'SURPLUS_MONEY'
;
