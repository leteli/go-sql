-- name: GetUserByCourseMemberID :one
SELECT
  cm.id AS course_member_id,
  u.id AS user_id,
  u.email AS user_email,
  u.name AS user_name,
  u.age AS user_age
FROM course_members AS cm
LEFT JOIN users AS u ON cm.user_id = u.id
WHERE cm.id = ?;

-- name: GetCourseWithMembers :one
WITH members AS (
  SELECT
    cm.course_id,
    cm.user_id,
    cm.joined_at
  FROM course_members AS cm
  WHERE cm.course_id = @id
)
SELECT
  c.id AS course_id,
  c.slug AS course_slug,
  c.title AS course_title,
  c.price AS course_price,
  COALESCE(
    json_group_array(
      json_object(
        'id', u.id,
        'email', u.email,
        'name', u.name,
        'age', u.age
      )
    ) FILTER (WHERE u.id IS NOT NULL),
    json('[]')
  ) AS members
FROM courses AS c
LEFT JOIN members AS m ON c.id = m.course_id
LEFT JOIN users AS u ON u.id = m.user_id
WHERE c.id = @id;