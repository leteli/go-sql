-- name: CreateUser :one
INSERT INTO users
(email, name, age)
VALUES (?, ?, ?)
RETURNING id;

-- name: UpdateUser :one
UPDATE users
SET
  email = COALESCE(sqlc.narg(email), email),
  name = COALESCE(sqlc.narg(name), name),
  age = COALESCE(sqlc.narg(age), age)
WHERE id = @id
RETURNING id, email, name, age, created_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;