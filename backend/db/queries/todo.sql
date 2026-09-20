-- name: ListTodos :many
-- * は sqlc が生成時に列名へ展開する。全クエリで同じ Todo 型になり、変換を 1 か所にまとめられる
SELECT * FROM todos ORDER BY id;

-- name: CreateTodo :one
-- id は GENERATED ALWAYS AS IDENTITY なので指定できない。採番結果は RETURNING で受け取る
INSERT INTO todos (title) VALUES ($1) RETURNING *;

-- name: DeleteTodo :exec
DELETE FROM todos WHERE id = $1;
