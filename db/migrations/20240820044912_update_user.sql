-- +goose Up
update user
set is_surplus_fund = null,
    has_experience  = null
where signup_status not in
      ('SIGNUP_INVESTMENT_SUITABILITY', 'SIGNUP_AGREE', 'SIGNUP_MYNUMBER', 'SIGNUP_KYC', 'SIGNUP_COMPLETE')
;

update private_profile
set is_pro_invest_finance_asset = null,
    is_pro_invest_asset         = null
where user_id in (select id
                  from user
                  where signup_status not in
                        ('SIGNUP_INVESTMENT_SUITABILITY', 'SIGNUP_AGREE', 'SIGNUP_MYNUMBER', 'SIGNUP_KYC',
                         'SIGNUP_COMPLETE'))
;

-- +goose Down
update user
set is_surplus_fund = 0,
    has_experience  = 0
where signup_status in
      ('SIGNUP_INVESTMENT_SUITABILITY', 'SIGNUP_AGREE', 'SIGNUP_MYNUMBER', 'SIGNUP_KYC', 'SIGNUP_COMPLETE')
;

update private_profile
set is_pro_invest_finance_asset = 0,
    is_pro_invest_asset         = 0
where user_id in (select id
                  from user
                  where signup_status in
                        ('SIGNUP_INVESTMENT_SUITABILITY', 'SIGNUP_AGREE', 'SIGNUP_MYNUMBER', 'SIGNUP_KYC',
                         'SIGNUP_COMPLETE'))
;
