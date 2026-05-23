-- name: GetUserByOrderID :one
SELECT
  o.id AS order_id,
  u.id AS user_id,
  u.email AS user_email,
  u.name AS user_name,
  u.age AS user_age
FROM orders AS o
LEFT JOIN users AS u ON o.user_id = u.id
WHERE o.id = ?;

-- name: ListOrders :many
SELECT
  o.id AS order_id,
  o.amount_cents AS order_amount_cents,
  o.created_at AS order_created_at,
  u.id AS user_id,
  u.email AS user_email,
  u.name AS user_name,
  u.age AS user_age,
  c.id AS course_id,
  c.slug AS course_slug,
  c.title AS course_title,
  c.price AS course_price
FROM orders AS o
LEFT JOIN users AS u ON o.user_id = u.id
LEFT JOIN courses AS c ON o.course_id = c.id
ORDER BY o.created_at DESC
LIMIT ? OFFSET ?;