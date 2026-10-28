-- +goose Up
UPDATE user_virtual_account
SET va_branch_name = 'アジサイ'
WHERE va_branch_name = 'ｱｼﾞｻｲ';

-- va_holder_name_kana 값이 'ｶ)ﾊｺﾌﾞﾈ'인 경우 'カ）ハコブネ'로 업데이트
UPDATE user_virtual_account
SET va_holder_name_kana = 'カ）ハコブネ'
WHERE va_holder_name_kana = 'ｶ)ﾊｺﾌﾞﾈ';

-- +goose Down
UPDATE user_virtual_account
SET va_branch_name = 'ｱｼﾞｻｲ'
WHERE va_branch_name = 'アジサイ';

UPDATE user_virtual_account
SET va_holder_name_kana = 'ｶ)ﾊｺﾌﾞﾈ'
WHERE va_holder_name_kana = 'カ）ハコブネ';