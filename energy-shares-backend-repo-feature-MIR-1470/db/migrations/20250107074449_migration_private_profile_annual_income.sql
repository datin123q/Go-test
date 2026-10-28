-- +goose Up
UPDATE private_profile
SET annual_income = CASE
                        WHEN annual_income = 'ASSET1' THEN 'ANNUALINCOME1'
                        WHEN annual_income = 'ASSET2' THEN 'ANNUALINCOME2'
                        WHEN annual_income = 'ASSET3' THEN 'ANNUALINCOME3'
                        WHEN annual_income = 'ASSET4' THEN 'ANNUALINCOME4'
                        WHEN annual_income = 'ASSET5' THEN 'ANNUALINCOME5'
                        WHEN annual_income = 'ASSET6' THEN 'ANNUALINCOME6'
    END
WHERE annual_income IN ('ASSET1', 'ASSET2', 'ASSET3', 'ASSET4', 'ASSET5', 'ASSET6');

