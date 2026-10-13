-- name: GetCustomerByID :one
SELECT * FROM customers WHERE id = $1;

-- name: CreateCustomer :one
-- アカウント作成時はメールアドレスだけを入れる。氏名などは NULL のまま
INSERT INTO customers (email) VALUES ($1) RETURNING *;

-- name: CreateCustomerCredential :exec
INSERT INTO customer_credentials (customer_id, password_hash) VALUES ($1, $2);

-- name: GetPasswordHashByEmail :one
-- ログイン時の照合に使う。メールアドレスは小文字にそろえてから渡す
SELECT c.id AS customer_id, cc.password_hash
FROM customers c
JOIN customer_credentials cc ON cc.customer_id = c.id
WHERE c.email = $1;
