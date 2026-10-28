-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- 기존 뷰가 존재하면 제거합니다.
DROP VIEW IF EXISTS deposit_match_data;

-- 필요한 열과 계산을 포함하여 새로운 뷰를 정의합니다.
CREATE VIEW deposit_match_data AS
SELECT u.id                                                        AS user_id,
       u.name                                                      AS investor_name,
       dr.investment_amount,
       dr.deposit_total_amount,
       CASE
           WHEN dr.investment_amount > dr.deposit_total_amount THEN 'LACK_DEPOSIT'
           ELSE 'SAME_DEPOSIT'
           END                                                     AS deposit_match_status,
       GREATEST(dr.investment_amount - dr.deposit_total_amount, 0) AS deposit_match_amount,
       GREATEST(dr.deposit_total_amount - dr.investment_amount, 0) AS lack_amount,
       dr.update_date
FROM (SELECT user.id,
             user.name,
             user.user_type
      FROM user
      WHERE user.user_type != 'ADMIN') AS u
-- 취소된 입금을 제외한 사용자별 입금 요청 데이터를 집계합니다.
         LEFT JOIN
     (SELECT deposit_request.user_id,
             SUM(deposit_request.request_amount) AS investment_amount,
             SUM(deposit_request.deposit_amount) AS deposit_total_amount,
             MAX(deposit_request.update_date)    AS update_date
      FROM deposit_request
      WHERE deposit_request.deposit_check_status != 'CANCEL_DEPOSIT'
      GROUP BY deposit_request.user_id) AS dr
-- 사용자 ID로 사용자 데이터와 집계된 입금 요청 데이터를 조인합니다.
     ON u.id = dr.user_id;

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

-- 뷰를 제거합니다.
DROP VIEW IF EXISTS deposit_match_data;