-- name: GetUser :one
SELECT * from users
WHERE name = $1 LIMIT 1;
