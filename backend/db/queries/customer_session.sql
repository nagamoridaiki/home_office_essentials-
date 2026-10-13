-- name: CreateCustomerSession :one
INSERT INTO customer_sessions (customer_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCustomerSessionByTokenHash :one
SELECT * FROM customer_sessions WHERE token_hash = $1;

-- name: RevokeCustomerSession :execrows
-- ログアウト済みの行は変えない。0 行なら「すでに有効でない」と判断する
UPDATE customer_sessions
SET revoked_at = $2
WHERE id = $1 AND revoked_at IS NULL;
