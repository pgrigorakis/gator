-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
SELECT * from users WHERE name = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * from users WHERE id = $1 LIMIT 1;

-- name: GetAllUsers :many
SELECT * from users;

-- name: DeleteUsers :exec
DELETE FROM users;
