-- name: CreateUser :one 
INSERT INTO users (id, created_at, updated_at, name)
VALUES ($1, now(), now(), $2)
RETURNING *;

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1;

-- name: ResetUsers :exec
DELETE FROM users;
