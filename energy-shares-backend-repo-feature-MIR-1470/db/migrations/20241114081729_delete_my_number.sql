-- +goose Up
-- 중복된 마이넘버 soft delete 처리 했던 것 hard delete 처리
DELETE
FROM my_number
WHERE deleted IS NOT NULL
  AND deleted != '0001-01-01 00:00:00';

-- 유저 아이디별로 가장 최근에 생성된 마이넘버만 남기고 나머지는 soft delete 처리
UPDATE my_number
SET deleted = NOW()
WHERE id NOT IN (SELECT id
                 FROM (SELECT MAX(id) AS id
                       FROM my_number
                       GROUP BY user_id) AS latest_records);


-- +goose Down
UPDATE my_number
SET deleted = NULL;

