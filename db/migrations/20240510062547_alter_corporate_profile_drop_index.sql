-- +goose Up
ALTER TABLE corporate_profile
DROP KEY UQE_corporate_profile_corp_no;

-- +goose Down
ALTER TABLE corporate_profile
ADD UNIQUE KEY UQE_corporate_profile_corp_no (corp_no);

