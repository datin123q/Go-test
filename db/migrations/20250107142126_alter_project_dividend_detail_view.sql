-- +goose Up
DROP VIEW IF EXISTS project_dividend_detail;

CREATE VIEW project_dividend_detail AS
SELECT dividend.id                                                                           AS id,
       dividend.user_id                                                                      AS user_id,
       dividend.project_id                                                                   AS project_id,
       dividend.project_type                                                                 AS project_type,
       project.name                                                                          AS project_name,
       project.project_status                                                                AS project_status,
       (SELECT SUM(s.qty)
        FROM share s
        WHERE s.project_id = dividend.project_id
          AND s.owner_id = dividend.user_id
          AND s.invest_status IN ('REFUND_INVEST', 'OPERATION_INVEST')
          AND (s.deleted = '0001-01-01 00:00:00' OR s.deleted IS NULL))                      AS investment_qty,
       (SELECT SUM(s.amount)
        FROM share s
        WHERE s.project_id = dividend.project_id
          AND s.owner_id = dividend.user_id
          AND s.invest_status IN ('REFUND_INVEST', 'OPERATION_INVEST')
          AND (s.deleted = '0001-01-01 00:00:00' OR s.deleted IS NULL))                      AS investment_amount,
       SUM(dividend.amount)                                                                  AS total_amount,
       SUM(dividend.tax)                                                                     AS total_tax,
       SUM(dividend.deposit_amount)                                                          AS total_deposit_amount,
       SUM(IF((dividend.transaction_type = 'DEPOSIT_REPAYMENT'), dividend.amount, 0))        AS repayment_amount,
       SUM(IF((dividend.transaction_type = 'DEPOSIT_DIVIDEND'), dividend.amount, 0))         AS dividend_amount,
       SUM(IF((dividend.transaction_type = 'DEPOSIT_DIVIDEND'), dividend.deposit_amount, 0)) AS total_revenue_amount
FROM dividend
         JOIN project
              ON dividend.project_id = project.id
                  AND (project.deleted = '0001-01-01 00:00:00' OR project.deleted IS NULL)
WHERE dividend.transaction_status = 'TRANSACTION_OK'
  AND (dividend.deleted = '0001-01-01 00:00:00' OR dividend.deleted IS NULL)
GROUP BY dividend.project_id, dividend.user_id;
