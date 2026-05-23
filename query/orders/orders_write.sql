-- name: CreateOrder :exec
INSERT INTO orders
(user_id, course_id, amount_cents)
VALUES (?, ?, ?);