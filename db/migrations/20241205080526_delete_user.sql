-- +goose Up
DELETE user
FROM user
INNER JOIN admin_user ON user.id = admin_user.id
WHERE user_type = 'ADMIN';

-- +goose Down
-- 별도 작업 하지 않음
