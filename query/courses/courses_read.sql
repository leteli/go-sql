-- name: FindCourseByID :one
SELECT id, slug, title, price
FROM courses
WHERE id = ?;

-- name: ListCourses :many
SELECT id, slug, title, price
FROM courses
LIMIT ? OFFSET ?;

-- name: FindCoursesByIDs :many
SELECT id, slug, title, price
FROM courses
WHERE id IN (sqlc.slice('ids'));

-- name: ListCoursesByMaxPrice :many
SELECT id, slug, title, price
FROM courses
WHERE price <= ?;