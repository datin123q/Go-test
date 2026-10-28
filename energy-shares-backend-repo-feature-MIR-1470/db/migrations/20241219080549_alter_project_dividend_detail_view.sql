-- +goose Up
DROP VIEW IF EXISTS project_dividend_detail;

CREATE VIEW project_dividend_detail AS
SELECT dividend.id                 AS id,
       dividend.user_id            AS user_id,
       dividend.project_id         AS project_id,
       dividend.project_type       AS project_type,
       dividend.investment_amount  AS investment_amount,
       dividend.investment_qty     AS investment_qty,
       dividend.amount             AS amount,
       dividend.qty                AS qty,
       dividend.deposit_amount     AS deposit_amount,
       dividend.tax                AS tax,
       dividend.dividend_date      AS dividend_date,
       dividend.transaction_status AS transaction_status,
       dividend.transaction_type   AS transaction_type,
       dividend.transaction_date   AS transaction_date,
       dividend.transaction_id     AS transaction_id,
       dividend.created            AS created,
       dividend.updated            AS updated,
       dividend.deleted            AS deleted,
       project.name                AS project_name,
       project.project_status      AS project_status,
       SUM(IF((dividend.transaction_type = 'DEPOSIT_REPAYMENT'), dividend.amount,
              0))                  AS repayment_amount,
       SUM(IF((dividend.transaction_type = 'DEPOSIT_DIVIDEND'), dividend.amount,
              0))                  AS dividend_amount,
       SUM(dividend.amount)        AS total_amount,
       SUM(IF((dividend.transaction_type = 'DEPOSIT_DIVIDEND'), dividend.deposit_amount,
              0))                  AS total_revenue_amount,
       SUM(dividend.tax)           AS total_tax
FROM dividend
         JOIN project
              ON dividend.project_id = project.id
GROUP BY dividend.project_id, dividend.user_id;
