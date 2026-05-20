-- name: FindUserByID :one
SELECT id, email, name, age, created_at
FROM users
WHERE id = ?;

-- name: ListUsers :many
SELECT id, email, name, age, created_at
FROM users
LIMIT ? OFFSET ?;