-- name: CreateCourseMember :exec
INSERT INTO course_members
(user_id, course_id)
VALUES (?, ?);