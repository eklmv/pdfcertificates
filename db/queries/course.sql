-- name: CreateCourse :one
INSERT INTO course (data)
VALUES (coalesce(sqlc.narg(data), '{}'::jsonb))
RETURNING *;

-- name: GetCourse :one
SELECT * FROM course
WHERE course_id = $1
LIMIT 1;

-- name: ListCourses :many
SELECT * FROM course
ORDER BY course_id
LIMIT $1 OFFSET $2;

-- name: ListCoursesLen :one
SELECT count(*) FROM course;

-- name: UpdateCourse :one
UPDATE course
SET data = coalesce(sqlc.narg(data), '{}'::jsonb)
WHERE course_id = $1
RETURNING *;

-- name: DeleteCourse :one
DELETE FROM course
WHERE course_id = $1
RETURNING *;
