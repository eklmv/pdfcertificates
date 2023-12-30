-- name: CreateCourse :one
INSERT INTO course (data)
VALUES ($1)
RETURNING *;

-- name: GetCourse :one
SELECT * FROM course
WHERE course_id = $1
LIMIT 1;

-- name: ListCourses :many
SELECT * FROM course
ORDER BY course_id
LIMIT $1 OFFSET $2;

-- name: UpdateCourse :one
UPDATE course
SET data = $2
WHERE course_id = $1
RETURNING *;

-- name: DeleteCourse :one
DELETE FROM course
WHERE course_id = $1
RETURNING *;
