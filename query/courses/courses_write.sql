-- name: CreateCourse :one
INSERT INTO courses
(slug, title, price)
VALUES (?, ?, ?)
RETURNING id;

-- name: UpdateCoursePrice :one
UPDATE courses
SET price = ?
WHERE id = ?
RETURNING id, slug, title, price;