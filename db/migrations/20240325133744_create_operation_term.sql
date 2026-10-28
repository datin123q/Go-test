-- +goose Up
alter table project
    add operation_term int DEFAULT 0;

UPDATE project
SET operation_term = CAST(loan_term AS SIGNED INTEGER) WHERE IFNULL(loan_term,0);

alter table project
drop column loan_term;

-- +goose Down
alter table energysharesdb.project
drop column operation_term;
