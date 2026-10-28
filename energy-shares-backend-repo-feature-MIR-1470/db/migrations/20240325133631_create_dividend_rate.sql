-- +goose Up
alter table project
    add dividend_rate double DEFAULT 0;

UPDATE project SET dividend_rate = CAST(loan_interest_rate AS double) WHERE IFNULL(loan_interest_rate,0.0);

alter table project
drop column fund_internal_rate_return;

alter table project
drop column loan_interest_rate;

-- +goose Down
alter table energysharesdb.project
drop column dividend_rate;
