-- +goose Up
alter table sign_up_request add referral_code varchar(255);

-- +goose Down
alter table sign_up_request drop column referral_code;


